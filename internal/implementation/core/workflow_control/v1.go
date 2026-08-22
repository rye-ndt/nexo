package workflow_control

import (
	"maps"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"

	"hexago/internal/helpers"
	"hexago/internal/helpers/custom_error"
	"hexago/internal/helpers/enums"
	core_itf "hexago/internal/interface/core"
	input_itf "hexago/internal/interface/input"
	output_itf "hexago/internal/interface/output"

	"github.com/google/uuid"
)

var promptReference = regexp.MustCompile(`\{\{\s*([\w.-]+)\s*\}\}`)

type v1 struct {
	cfg         input_itf.ControlConfig
	workflows   core_itf.WorkflowManager
	coordinator core_itf.Coordinator
	roles       core_itf.RoleManager
	userConfig  output_itf.UserConfig
	db          input_itf.StepStorage
}

func InitV1(
	cfg *input_itf.ControlConfig,
	workflows core_itf.WorkflowManager,
	coordinator core_itf.Coordinator,
	roles core_itf.RoleManager,
	userConfig output_itf.UserConfig,
	db input_itf.StepStorage,
) (core_itf.WorkflowControl, error) {
	if err := helpers.ValidateStruct(cfg); err != nil {
		return nil, custom_error.Critical("invalid workflow control config: %v", err)
	}

	return &v1{
		cfg:         *cfg,
		workflows:   workflows,
		coordinator: coordinator,
		roles:       roles,
		userConfig:  userConfig,
		db:          db,
	}, nil
}

func (s *v1) CreateWorkflow(spec *core_itf.ControlWorkflowSpec) (*core_itf.ControlWorkflowRef, error) {
	if spec == nil || len(spec.Steps) == 0 {
		return nil, custom_error.Critical("workflow spec has no steps")
	}

	if len(spec.Steps) > s.cfg.MaxStepsPerWorkflow {
		return nil, custom_error.Critical(
			"workflow asks for %d steps, mcp_servers.control.max_steps_per_workflow allows %d",
			len(spec.Steps), s.cfg.MaxStepsPerWorkflow,
		)
	}

	if err := s.checkWorkspace(spec.ProjectDirPath); err != nil {
		return nil, err
	}

	if err := checkClientIDs(spec.Steps); err != nil {
		return nil, err
	}

	workflowID, err := s.workflows.NewWorkflow(&core_itf.InitWorkflow{
		ProjectDirPath: spec.ProjectDirPath,
	})
	if err != nil {
		return nil, err
	}

	clientToStep, err := s.addSteps(workflowID, spec.Steps)
	if err != nil {
		return nil, custom_error.Critical(
			"workflow %s was created but could not be filled in, so cancel_workflow is what clears it: %v",
			workflowID, err,
		)
	}

	ref := &core_itf.ControlWorkflowRef{
		WorkflowID: workflowID,
		StepIDs:    clientToStep,
	}

	if !autostartWanted(spec.Autostart, s.cfg.AutostartDefault) {
		return ref, nil
	}

	if err := s.coordinator.Run(workflowID); err != nil {
		return nil, custom_error.Critical(
			"workflow %s was created but did not start, so start_workflow is what picks it up: %v",
			workflowID, err,
		)
	}

	ref.Started = true

	return ref, nil
}

func (s *v1) ListWorkflows(limit int) ([]*core_itf.WorkflowStatus, error) {
	snapshots, err := s.db.LoadStepHistory()
	if err != nil {
		return nil, custom_error.Critical("cannot load workflow history: %v", err)
	}

	if limit <= 0 || limit > s.cfg.MaxWorkflowsListed {
		limit = s.cfg.MaxWorkflowsListed
	}

	workflows := make([]*input_itf.WorkflowEntity, 0, len(snapshots))

	for _, snapshot := range snapshots {
		if snapshot == nil || snapshot.Workflow == nil {
			continue
		}

		workflows = append(workflows, snapshot.Workflow)
	}

	sort.Slice(workflows, func(i, j int) bool {
		return workflows[i].CreatedAt.After(workflows[j].CreatedAt)
	})

	statuses := make([]*core_itf.WorkflowStatus, 0, min(limit, len(workflows)))

	for _, workflow := range workflows {
		if len(statuses) == limit {
			break
		}

		status, err := s.workflows.Status(workflow.ID)
		if err != nil {
			continue
		}

		statuses = append(statuses, status)
	}

	return statuses, nil
}

func (s *v1) addSteps(workflowID uuid.UUID, steps []*core_itf.ControlStepSpec) (map[string]uuid.UUID, error) {
	autopilot := s.userConfig.Autopilot()
	clientToStep := make(map[string]uuid.UUID, len(steps))
	remaining := slices.Clone(steps)

	for len(remaining) > 0 {
		next := make([]*core_itf.ControlStepSpec, 0, len(remaining))

		for _, step := range remaining {
			deps, resolved := resolveDeps(step.DependsOn, clientToStep)
			if !resolved {
				next = append(next, step)
				continue
			}

			stepID, err := s.addStep(workflowID, step, deps, autopilot)
			if err != nil {
				return nil, err
			}

			clientToStep[step.ClientID] = stepID
		}

		if len(next) == len(remaining) {
			return nil, custom_error.Critical(
				"cannot resolve the dependencies of %s: the graph is cyclic or names a client id that is not in the workflow",
				strings.Join(clientIDs(next), ", "),
			)
		}

		remaining = next
	}

	return clientToStep, nil
}

func (s *v1) addStep(
	workflowID uuid.UUID,
	step *core_itf.ControlStepSpec,
	deps []uuid.UUID,
	autopilot bool,
) (uuid.UUID, error) {
	added, err := s.resolveStep(step)
	if err != nil {
		return uuid.Nil, err
	}

	agentDefault, err := s.userConfig.AgentDefault(added.Effort)
	if err != nil {
		return uuid.Nil, err
	}

	added.DependsOn = deps
	added.PauseForReview = added.PauseForReview && !autopilot
	added.AgentSpecs.Name = agentDefault.Model
	added.AgentSpecs.ThinkingLevel = agentDefault.ThinkingLevel

	return s.workflows.AddStep(workflowID, added)
}

func (s *v1) resolveStep(step *core_itf.ControlStepSpec) (*core_itf.AddStep, error) {
	if step.RoleID == uuid.Nil {
		return freeformStep(step)
	}

	return s.roleStep(step)
}

func freeformStep(step *core_itf.ControlStepSpec) (*core_itf.AddStep, error) {
	level, err := effort(step.ClientID, step.Effort, enums.EffortStandard)
	if err != nil {
		return nil, err
	}

	return &core_itf.AddStep{
		Name:            step.Name,
		Effort:          level,
		ExtraGuidance:   renderPrompt(step.Prompt, step.Inputs),
		OutputStructure: step.OutputStructure,
		AutoRetry:       step.AutoRetry,
		PauseForReview:  step.PauseForReview,
		AgentSpecs:      &core_itf.AgentRequest{Instructions: step.Instructions},
	}, nil
}

func (s *v1) roleStep(step *core_itf.ControlStepSpec) (*core_itf.AddStep, error) {
	role, err := s.roles.Get(step.RoleID)
	if err != nil {
		return nil, err
	}

	level, err := effort(step.ClientID, step.Effort, role.Effort)
	if err != nil {
		return nil, err
	}

	values, err := inputValues(step, role)
	if err != nil {
		return nil, err
	}

	added := &core_itf.AddStep{
		Name:            step.Name,
		Effort:          level,
		ExtraGuidance:   renderPrompt(step.Prompt, values),
		OutputStructure: role.OutputStructure,
		AutoRetry:       step.AutoRetry,
		PauseForReview:  role.PauseForReview || step.PauseForReview,
		AgentSpecs:      &core_itf.AgentRequest{Instructions: orderedPrompts(role.Instructions)},
	}

	if len(step.Instructions) > 0 {
		added.AgentSpecs.Instructions = step.Instructions
	}

	if step.OutputStructure != "" {
		added.OutputStructure = step.OutputStructure
	}

	return added, nil
}

func inputValues(step *core_itf.ControlStepSpec, role *core_itf.Role) (map[string]string, error) {
	values := make(map[string]string, len(role.Inputs))

	for key, value := range step.Inputs {
		input, known := role.Inputs[key]
		if !known {
			return nil, custom_error.Critical(
				"step %s sets input %s, which role %s does not declare",
				step.ClientID, key, role.Name,
			)
		}

		if value == "" {
			continue
		}

		if err := checkOptions(key, input, value); err != nil {
			return nil, err
		}

		values[key] = value
	}

	for key, input := range role.Inputs {
		if _, given := values[key]; given {
			continue
		}

		if input.Default != "" {
			values[key] = input.Default
			continue
		}

		if input.Required {
			return nil, custom_error.Critical("step %s is missing required input %s", step.ClientID, key)
		}
	}

	return values, nil
}

func checkOptions(key string, input *core_itf.RoleInputs, value string) error {
	switch enums.InputType(input.Type) {
	case enums.SelectInput:
		return checkChoice(key, input.Options, value)
	case enums.MultiInput:
		for _, choice := range strings.Split(value, ",") {
			if err := checkChoice(key, input.Options, strings.TrimSpace(choice)); err != nil {
				return err
			}
		}
	}

	return nil
}

func checkChoice(key string, options []string, value string) error {
	if slices.Contains(options, value) {
		return nil
	}

	return custom_error.Critical(
		"input %s does not accept %s, the options are %s",
		key, value, strings.Join(options, ", "),
	)
}

func effort(clientID, wanted string, fallback enums.Effort) (enums.Effort, error) {
	if wanted == "" {
		if fallback.Valid() {
			return fallback, nil
		}

		return enums.EffortStandard, nil
	}

	level := enums.Effort(wanted)
	if !level.Valid() {
		return "", custom_error.Critical("step %s asks for an unknown step level %s", clientID, wanted)
	}

	return level, nil
}

func renderPrompt(prompt string, inputs map[string]string) string {
	if len(inputs) == 0 {
		return prompt
	}

	referenced := map[string]bool{}
	for _, match := range promptReference.FindAllStringSubmatch(prompt, -1) {
		referenced[match[1]] = true
	}

	filled := promptReference.ReplaceAllStringFunc(prompt, func(reference string) string {
		value := inputs[promptReference.FindStringSubmatch(reference)[1]]
		if value == "" {
			return reference
		}

		return value
	})

	rest := make([]string, 0, len(inputs))

	for key, value := range inputs {
		if referenced[key] || value == "" {
			continue
		}

		rest = append(rest, key)
	}

	if len(rest) == 0 {
		return filled
	}

	sort.Strings(rest)

	lines := make([]string, 0, len(rest))
	for _, key := range rest {
		lines = append(lines, "- "+key+": "+inputs[key])
	}

	return filled + "\n\nInputs:\n" + strings.Join(lines, "\n")
}

func orderedPrompts(prompts map[string]string) []string {
	ordered := make([]string, 0, len(prompts))
	for _, key := range slices.Sorted(maps.Keys(prompts)) {
		ordered = append(ordered, prompts[key])
	}

	return ordered
}

func (s *v1) checkWorkspace(path string) error {
	if s.cfg.AllowAnyWorkspace {
		return nil
	}

	if len(s.cfg.AllowedRoots) == 0 {
		return custom_error.Critical(
			"no workspace is allowed: set mcp_servers.control.allowed_roots, or mcp_servers.control.allow_any_workspace, in config.yaml",
		)
	}

	target, err := resolvePath(path)
	if err != nil {
		return err
	}

	for _, root := range s.cfg.AllowedRoots {
		resolved, err := resolvePath(root)
		if err != nil {
			continue
		}

		if within(resolved, target) {
			return nil
		}
	}

	return custom_error.Critical(
		"working dir %s is outside every mcp_servers.control.allowed_roots entry in config.yaml", path,
	)
}

func resolvePath(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", custom_error.Critical("cannot resolve path %s: %v", path, err)
	}

	missing := ""

	for {
		real, err := filepath.EvalSymlinks(abs)
		if err == nil {
			return filepath.Join(real, missing), nil
		}

		parent := filepath.Dir(abs)
		if parent == abs {
			return filepath.Join(abs, missing), nil
		}

		missing = filepath.Join(filepath.Base(abs), missing)
		abs = parent
	}
}

func within(root, target string) bool {
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return false
	}

	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func checkClientIDs(steps []*core_itf.ControlStepSpec) error {
	seen := make(map[string]bool, len(steps))

	for _, step := range steps {
		if step == nil {
			return custom_error.Critical("workflow spec holds an empty step")
		}

		if seen[step.ClientID] {
			return custom_error.Critical("client id %s is used by more than one step", step.ClientID)
		}

		seen[step.ClientID] = true
	}

	return nil
}

func resolveDeps(dependsOn []string, clientToStep map[string]uuid.UUID) ([]uuid.UUID, bool) {
	deps := make([]uuid.UUID, 0, len(dependsOn))

	for _, clientID := range dependsOn {
		stepID, found := clientToStep[clientID]
		if !found {
			return nil, false
		}

		deps = append(deps, stepID)
	}

	return deps, true
}

func clientIDs(steps []*core_itf.ControlStepSpec) []string {
	ids := make([]string, 0, len(steps))
	for _, step := range steps {
		ids = append(ids, step.ClientID)
	}

	return ids
}

func autostartWanted(wanted *bool, fallback bool) bool {
	if wanted == nil {
		return fallback
	}

	return *wanted
}
