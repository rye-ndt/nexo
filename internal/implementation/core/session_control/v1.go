package session_control

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
	sessions    core_itf.SessionManager
	coordinator core_itf.Coordinator
	templates   core_itf.AgentTemplateManager
	userConfig  output_itf.UserConfig
	db          input_itf.TaskStorage
}

func InitV1(
	cfg *input_itf.ControlConfig,
	sessions core_itf.SessionManager,
	coordinator core_itf.Coordinator,
	templates core_itf.AgentTemplateManager,
	userConfig output_itf.UserConfig,
	db input_itf.TaskStorage,
) (core_itf.SessionControl, error) {
	if err := helpers.ValidateStruct(cfg); err != nil {
		return nil, custom_error.Critical("invalid session control config: %v", err)
	}

	return &v1{
		cfg:         *cfg,
		sessions:    sessions,
		coordinator: coordinator,
		templates:   templates,
		userConfig:  userConfig,
		db:          db,
	}, nil
}

func (s *v1) CreateSession(spec *core_itf.ControlSessionSpec) (*core_itf.ControlSessionRef, error) {
	if spec == nil || len(spec.Tasks) == 0 {
		return nil, custom_error.Critical("session spec has no tasks")
	}

	if len(spec.Tasks) > s.cfg.MaxTasksPerSession {
		return nil, custom_error.Critical(
			"session asks for %d tasks, mcp_servers.control.max_tasks_per_session allows %d",
			len(spec.Tasks), s.cfg.MaxTasksPerSession,
		)
	}

	if err := s.checkWorkspace(spec.WorkingDirPath); err != nil {
		return nil, err
	}

	if strings.TrimSpace(spec.ContextDirPath) != "" {
		if err := s.checkWorkspace(spec.ContextDirPath); err != nil {
			return nil, err
		}
	}

	if err := checkClientIDs(spec.Tasks); err != nil {
		return nil, err
	}

	sessionID, err := s.sessions.NewSession(&core_itf.InitSession{
		WorkingDirPath: spec.WorkingDirPath,
		ContextDirPath: spec.ContextDirPath,
	})
	if err != nil {
		return nil, err
	}

	clientToTask, err := s.addTasks(sessionID, spec.Tasks)
	if err != nil {
		return nil, custom_error.Critical(
			"session %s was created but could not be filled in, so cancel_session is what clears it: %v",
			sessionID, err,
		)
	}

	ref := &core_itf.ControlSessionRef{
		SessionID: sessionID,
		TaskIDs:   clientToTask,
	}

	if !autostartWanted(spec.Autostart, s.cfg.AutostartDefault) {
		return ref, nil
	}

	if err := s.coordinator.Run(sessionID); err != nil {
		return nil, custom_error.Critical(
			"session %s was created but did not start, so start_session is what picks it up: %v",
			sessionID, err,
		)
	}

	ref.Started = true

	return ref, nil
}

func (s *v1) StartSession(sessionID uuid.UUID) error {
	return s.coordinator.Run(sessionID)
}

func (s *v1) PauseSession(sessionID uuid.UUID) error {
	return s.coordinator.Pause(sessionID)
}

func (s *v1) CancelSession(sessionID uuid.UUID) error {
	return s.coordinator.Cancel(sessionID)
}

func (s *v1) SessionState(sessionID uuid.UUID) (*core_itf.SessionStatus, error) {
	return s.sessions.Status(sessionID)
}

func (s *v1) ListSessions(limit int) ([]*core_itf.SessionStatus, error) {
	snapshots, err := s.db.LoadTaskHistory()
	if err != nil {
		return nil, custom_error.Critical("cannot load session history: %v", err)
	}

	if limit <= 0 || limit > s.cfg.MaxSessionsListed {
		limit = s.cfg.MaxSessionsListed
	}

	sessions := make([]*input_itf.SessionEntity, 0, len(snapshots))

	for _, snapshot := range snapshots {
		if snapshot == nil || snapshot.Session == nil {
			continue
		}

		sessions = append(sessions, snapshot.Session)
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].CreatedAt.After(sessions[j].CreatedAt)
	})

	statuses := make([]*core_itf.SessionStatus, 0, min(limit, len(sessions)))

	for _, session := range sessions {
		if len(statuses) == limit {
			break
		}

		status, err := s.sessions.Status(session.ID)
		if err != nil {
			continue
		}

		statuses = append(statuses, status)
	}

	return statuses, nil
}

func (s *v1) ListTemplates() ([]*core_itf.Template, error) {
	return s.templates.List()
}

func (s *v1) AnswerAcceptance(taskID uuid.UUID, accepted bool) error {
	return s.sessions.AnswerAcceptance(taskID, accepted)
}

func (s *v1) addTasks(sessionID uuid.UUID, tasks []*core_itf.ControlTaskSpec) (map[string]uuid.UUID, error) {
	autopilot := s.userConfig.Autopilot()
	clientToTask := make(map[string]uuid.UUID, len(tasks))
	remaining := slices.Clone(tasks)

	for len(remaining) > 0 {
		next := make([]*core_itf.ControlTaskSpec, 0, len(remaining))

		for _, task := range remaining {
			deps, resolved := resolveDeps(task.DependsOn, clientToTask)
			if !resolved {
				next = append(next, task)
				continue
			}

			taskID, err := s.addTask(sessionID, task, deps, autopilot)
			if err != nil {
				return nil, err
			}

			clientToTask[task.ClientID] = taskID
		}

		if len(next) == len(remaining) {
			return nil, custom_error.Critical(
				"cannot resolve the dependencies of %s: the graph is cyclic or names a client id that is not in the session",
				strings.Join(clientIDs(next), ", "),
			)
		}

		remaining = next
	}

	return clientToTask, nil
}

func (s *v1) addTask(
	sessionID uuid.UUID,
	task *core_itf.ControlTaskSpec,
	deps []uuid.UUID,
	autopilot bool,
) (uuid.UUID, error) {
	added, err := s.resolveTask(task)
	if err != nil {
		return uuid.Nil, err
	}

	agentDefault, err := s.userConfig.AgentDefault(added.TaskLevel)
	if err != nil {
		return uuid.Nil, err
	}

	added.DependsOn = deps
	added.ManualAcceptRequired = added.ManualAcceptRequired && !autopilot
	added.AgentSpecs.Name = agentDefault.Model
	added.AgentSpecs.ThinkingLevel = agentDefault.ThinkingLevel

	return s.sessions.AddTask(sessionID, added)
}

func (s *v1) resolveTask(task *core_itf.ControlTaskSpec) (*core_itf.AddTask, error) {
	if task.TemplateID == uuid.Nil {
		return freeformTask(task)
	}

	return s.templateTask(task)
}

func freeformTask(task *core_itf.ControlTaskSpec) (*core_itf.AddTask, error) {
	level, err := taskLevel(task.ClientID, task.TaskLevel, enums.DailyTask)
	if err != nil {
		return nil, err
	}

	return &core_itf.AddTask{
		Name:                 task.Name,
		TaskLevel:            level,
		ExtraGuidance:        renderPrompt(task.Prompt, task.Params),
		OutputStructure:      task.OutputStructure,
		AutoRetry:            task.AutoRetry,
		ManualAcceptRequired: task.ManualAcceptRequired,
		AgentSpecs:           &core_itf.AgentRequest{SystemPrompts: task.SystemPrompts},
	}, nil
}

func (s *v1) templateTask(task *core_itf.ControlTaskSpec) (*core_itf.AddTask, error) {
	template, err := s.templates.Get(task.TemplateID)
	if err != nil {
		return nil, err
	}

	level, err := taskLevel(task.ClientID, task.TaskLevel, template.TaskLevel)
	if err != nil {
		return nil, err
	}

	values, err := paramValues(task, template)
	if err != nil {
		return nil, err
	}

	added := &core_itf.AddTask{
		Name:                 task.Name,
		TaskLevel:            level,
		ExtraGuidance:        renderPrompt(task.Prompt, values),
		OutputStructure:      template.OutputStructure,
		AutoRetry:            task.AutoRetry,
		ManualAcceptRequired: template.ManualAcceptRequired || task.ManualAcceptRequired,
		AgentSpecs:           &core_itf.AgentRequest{SystemPrompts: orderedPrompts(template.SystemPrompts)},
	}

	if len(task.SystemPrompts) > 0 {
		added.AgentSpecs.SystemPrompts = task.SystemPrompts
	}

	if task.OutputStructure != "" {
		added.OutputStructure = task.OutputStructure
	}

	return added, nil
}

func paramValues(task *core_itf.ControlTaskSpec, template *core_itf.Template) (map[string]string, error) {
	values := make(map[string]string, len(template.Params))

	for key, value := range task.Params {
		param, known := template.Params[key]
		if !known {
			return nil, custom_error.Critical(
				"task %s sets param %s, which template %s does not declare",
				task.ClientID, key, template.Name,
			)
		}

		if value == "" {
			continue
		}

		if err := checkOptions(key, param, value); err != nil {
			return nil, err
		}

		values[key] = value
	}

	for key, param := range template.Params {
		if _, given := values[key]; given {
			continue
		}

		if param.Default != "" {
			values[key] = param.Default
			continue
		}

		if param.Required {
			return nil, custom_error.Critical("task %s is missing required param %s", task.ClientID, key)
		}
	}

	return values, nil
}

func checkOptions(key string, param *core_itf.TemplateParams, value string) error {
	switch enums.ParamType(param.Type) {
	case enums.SelectParam:
		return checkChoice(key, param.Options, value)
	case enums.MultiParam:
		for _, choice := range strings.Split(value, ",") {
			if err := checkChoice(key, param.Options, strings.TrimSpace(choice)); err != nil {
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
		"param %s does not accept %s, the options are %s",
		key, value, strings.Join(options, ", "),
	)
}

func taskLevel(clientID, wanted string, fallback enums.TaskLevel) (enums.TaskLevel, error) {
	if wanted == "" {
		if fallback.Valid() {
			return fallback, nil
		}

		return enums.DailyTask, nil
	}

	level := enums.TaskLevel(wanted)
	if !level.Valid() {
		return "", custom_error.Critical("task %s asks for an unknown task level %s", clientID, wanted)
	}

	return level, nil
}

func renderPrompt(prompt string, params map[string]string) string {
	if len(params) == 0 {
		return prompt
	}

	referenced := map[string]bool{}
	for _, match := range promptReference.FindAllStringSubmatch(prompt, -1) {
		referenced[match[1]] = true
	}

	filled := promptReference.ReplaceAllStringFunc(prompt, func(reference string) string {
		value := params[promptReference.FindStringSubmatch(reference)[1]]
		if value == "" {
			return reference
		}

		return value
	})

	rest := make([]string, 0, len(params))

	for key, value := range params {
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
		lines = append(lines, "- "+key+": "+params[key])
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

func checkClientIDs(tasks []*core_itf.ControlTaskSpec) error {
	seen := make(map[string]bool, len(tasks))

	for _, task := range tasks {
		if task == nil {
			return custom_error.Critical("session spec holds an empty task")
		}

		if seen[task.ClientID] {
			return custom_error.Critical("client id %s is used by more than one task", task.ClientID)
		}

		seen[task.ClientID] = true
	}

	return nil
}

func resolveDeps(dependsOn []string, clientToTask map[string]uuid.UUID) ([]uuid.UUID, bool) {
	deps := make([]uuid.UUID, 0, len(dependsOn))

	for _, clientID := range dependsOn {
		taskID, found := clientToTask[clientID]
		if !found {
			return nil, false
		}

		deps = append(deps, taskID)
	}

	return deps, true
}

func clientIDs(tasks []*core_itf.ControlTaskSpec) []string {
	ids := make([]string, 0, len(tasks))
	for _, task := range tasks {
		ids = append(ids, task.ClientID)
	}

	return ids
}

func autostartWanted(wanted *bool, fallback bool) bool {
	if wanted == nil {
		return fallback
	}

	return *wanted
}
