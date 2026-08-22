package mcp_proxy

import (
	"encoding/json"
	"maps"
	"net/http"
	"slices"
	"sort"
	"time"

	"hexago/internal/helpers"
	"hexago/internal/helpers/constances"
	"hexago/internal/helpers/custom_error"
	"hexago/internal/helpers/enums"
	core_itf "hexago/internal/interface/core"

	"github.com/google/uuid"
)

const (
	listRolesTool        = "list_roles"
	createWorkflowTool   = "create_workflow"
	startWorkflowTool    = "start_workflow"
	pauseWorkflowTool    = "pause_workflow"
	cancelWorkflowTool   = "cancel_workflow"
	workflowStatusTool   = "workflow_status"
	listWorkflowsTool    = "list_workflows"
	answerAcceptanceTool = "answer_review"
)

const controlUnavailable = "this app is not ready to take control calls yet"

const listRolesToolDescription = `List the roles this app already has.
A role fixes what an agent does: its description, how much effort a step built from it deserves, whether
a human must review the result before anything downstream runs, and the inputs a step fills in before it runs.
Read this before create_workflow: giving a step a role_id and its inputs is better than writing the
whole job out again in a raw prompt.`

const createWorkflowToolDescription = `Create a workflow: a graph of scoped agent steps over one project folder.
Every step has its own prompt; a step runs once every step it depends on has finished, and what a
finished step learned is carried into its dependents' prompts.
You name each step with your own client_id, and depends_on refers to those client_ids from this same call —
they are local to the call, not ids the app already knows. The answer maps every client_id to the real step id.
With autostart false the graph is only staged and nothing runs until start_workflow; leave autostart out to let
the app's own setting decide.`

const startWorkflowToolDescription = `Start running a workflow's graph, and the way to resume one too:
a workflow that was paused, or that was interrupted when the app closed, picks up from the steps that never
finished rather than rerunning the ones that did.
The call returns as soon as the run is under way, so poll workflow_status to follow it.`

const pauseWorkflowToolDescription = `Pause a running workflow. The steps in flight are stopped and no further step is
handed out, while everything already finished is kept. Call start_workflow to pick the run back up where it stopped.`

const cancelWorkflowToolDescription = `Cancel a workflow for good. Running steps are killed and nothing else is handed
out; unlike pause, this cannot be resumed. Whatever the agents already wrote into the project folder stays on disk.`

const workflowStatusToolDescription = `Read where a workflow stands: its own status, every step with its status and the
one-sentence tldr its agent left, the tokens billed so far, and when it started and finished.
This is how you follow a run — poll it rather than treating a start_workflow call as the work being done.
A step sitting in awaiting_review is waiting for you to call answer_review.`

const listWorkflowsToolDescription = `List the most recent workflows, newest first, with their status and project
folder. Use it to find the workflow id the other tools need. limit caps how many come back; leave it out for the
app's default.`

const answerAcceptanceToolDescription = `Answer the review on a step that finished and is waiting on a human.
Accepting marks the step completed and releases everything downstream of it; rejecting marks it failed, so its
dependents stay blocked. Only a step whose status is awaiting_review is waiting to be reviewed.`

type controlWorkflowArgs struct {
	WorkflowID string `json:"workflow_id"`
}

type listWorkflowsArgs struct {
	Limit int `json:"limit"`
}

type answerReviewArgs struct {
	StepID   string `json:"step_id"`
	Accepted bool   `json:"accepted"`
}

type controlStepArgs struct {
	ClientID        string            `json:"client_id"`
	Name            string            `json:"name"`
	Prompt          string            `json:"prompt"`
	RoleID          string            `json:"role_id"`
	Inputs          map[string]string `json:"inputs"`
	Effort          string            `json:"effort"`
	Instructions    []string          `json:"instructions"`
	OutputStructure string            `json:"output_structure"`
	DependsOn       []string          `json:"depends_on"`
	AutoRetry       bool              `json:"auto_retry"`
	PauseForReview  bool              `json:"pause_for_review"`
}

type createWorkflowArgs struct {
	ProjectDirPath string             `json:"project_dir_path"`
	Autostart      *bool              `json:"autostart"`
	Steps          []*controlStepArgs `json:"steps"`
}

type roleInputPayload struct {
	Key         string   `json:"key"`
	Description string   `json:"description,omitempty"`
	Type        string   `json:"type,omitempty"`
	Default     string   `json:"default,omitempty"`
	Required    bool     `json:"required"`
	Options     []string `json:"options,omitempty"`
}

type rolePayload struct {
	ID             string              `json:"id"`
	Name           string              `json:"name"`
	Description    string              `json:"description"`
	Effort         string              `json:"effort"`
	PauseForReview bool                `json:"pause_for_review"`
	Inputs         []*roleInputPayload `json:"inputs"`
}

type workflowRefPayload struct {
	WorkflowID string            `json:"workflow_id"`
	StepIDs    map[string]string `json:"step_ids"`
	Started    bool              `json:"started"`
}

type stepStatePayload struct {
	StepID  string `json:"step_id"`
	Name    string `json:"name"`
	Status  string `json:"status"`
	Effort  string `json:"effort,omitempty"`
	TLDR    string `json:"tldr,omitempty"`
	Outcome string `json:"outcome,omitempty"`
}

type workflowStatePayload struct {
	WorkflowID     string              `json:"workflow_id"`
	Status         string              `json:"status"`
	ProjectDirPath string              `json:"project_dir_path"`
	Steps          []*stepStatePayload `json:"steps"`
	TokensBilled   int                 `json:"tokens_billed"`
	TokensInput    int                 `json:"tokens_input"`
	TokensCached   int                 `json:"tokens_cached"`
	StartedAt      string              `json:"started_at,omitempty"`
	CompletedAt    string              `json:"completed_at,omitempty"`
}

type workflowSummaryPayload struct {
	WorkflowID     string `json:"workflow_id"`
	Status         string `json:"status"`
	ProjectDirPath string `json:"project_dir_path"`
	TotalStep      int    `json:"total_step"`
	StartedAt      string `json:"started_at,omitempty"`
	CompletedAt    string `json:"completed_at,omitempty"`
}

type workflowAckPayload struct {
	WorkflowID string `json:"workflow_id"`
	Result     string `json:"result"`
}

type acceptanceAckPayload struct {
	StepID   string `json:"step_id"`
	Accepted bool   `json:"accepted"`
	Result   string `json:"result"`
}

func (s *v1) TrackWorkflowControl(ports *core_itf.ControlPorts) {
	s.locker.Lock()
	defer s.locker.Unlock()

	s.controlPorts = ports
}

func (s *v1) control() *core_itf.ControlPorts {
	s.locker.RLock()
	defer s.locker.RUnlock()

	return s.controlPorts
}

func (s *v1) serveControl(w http.ResponseWriter, r *http.Request) {
	serveRPC(w, r, constances.ControlLocalServer, s.controlTools())
}

var listRolesToolSchema = objectSchema(map[string]any{})

var createWorkflowToolSchema = objectSchema(map[string]any{
	"project_dir_path": stringProp("Absolute path to the project folder the agents work in. Every step in the workflow shares it."),
	"autostart": map[string]any{
		"type":        "boolean",
		"description": "True to run the graph immediately, false to only stage it and wait for start_workflow. Leave it out to follow the app's own setting.",
	},
	"steps": map[string]any{
		"type":        "array",
		"minItems":    1,
		"description": "The steps of the graph, each one scoped tightly enough that a single agent can finish it.",
		"items": objectSchema(map[string]any{
			"client_id": stringProp("Your own short name for this step, unique within this call. depends_on refers to it."),
			"name":      stringProp("Two or three words naming what this step does, read by a person."),
			"prompt":    stringProp("What this step's agent must do. Say the goal and what done looks like, not how to get there."),
			"role_id": stringProp("Optional id from list_roles. The role supplies the instructions, effort and " +
				"output structure, and the prompt narrows them to this step."),
			"inputs": map[string]any{
				"type":                 "object",
				"additionalProperties": map[string]any{"type": "string"},
				"description":          "Values for the role's inputs, keyed by the input key.",
			},
			"effort": map[string]any{
				"type":        "string",
				"enum":        helpers.Labels(enums.Efforts()),
				"description": "How much effort this step deserves. Leave it out to keep the role's effort.",
			},
			"instructions": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "Extra instruction sections for this step, on top of whatever the role already says.",
			},
			"output_structure": stringProp("The fields this step must report, one per line as `name: what it holds`. " +
				"Leave it empty to report in its own words or to keep the role's structure."),
			"depends_on": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "The client_ids of the steps that must finish first. Their handoffs become this step's context.",
			},
			"auto_retry": map[string]any{
				"type":        "boolean",
				"description": "Whether a failed run of this step is retried automatically. Keep it false when a retry could repeat a side effect.",
			},
			"pause_for_review": map[string]any{
				"type":        "boolean",
				"description": "Whether a human must review this step's result before anything downstream runs.",
			},
		}, "client_id", "name", "prompt"),
	},
}, "project_dir_path", "steps")

var workflowIDToolSchema = objectSchema(map[string]any{
	"workflow_id": stringProp("The id returned by create_workflow or listed by list_workflows."),
}, "workflow_id")

var listWorkflowsToolSchema = objectSchema(map[string]any{
	"limit": map[string]any{
		"type":        "integer",
		"minimum":     1,
		"description": "How many workflows to return, newest first. Leave it out for the app's default.",
	},
})

var answerAcceptanceToolSchema = objectSchema(map[string]any{
	"step_id": stringProp("The id of the step waiting in awaiting_review, as reported by workflow_status."),
	"accepted": map[string]any{
		"type":        "boolean",
		"description": "True to accept the result and release the steps downstream, false to mark the step failed.",
	},
}, "step_id", "accepted")

func (s *v1) controlTools() []*rpcTool {
	return []*rpcTool{
		{
			name:        listRolesTool,
			description: listRolesToolDescription,
			input:       listRolesToolSchema,
			call:        s.callListRoles,
		},
		{
			name:        createWorkflowTool,
			description: createWorkflowToolDescription,
			input:       createWorkflowToolSchema,
			call:        s.callCreateWorkflow,
		},
		{
			name:        startWorkflowTool,
			description: startWorkflowToolDescription,
			input:       workflowIDToolSchema,
			call:        s.callStartWorkflow,
		},
		{
			name:        pauseWorkflowTool,
			description: pauseWorkflowToolDescription,
			input:       workflowIDToolSchema,
			call:        s.callPauseWorkflow,
		},
		{
			name:        cancelWorkflowTool,
			description: cancelWorkflowToolDescription,
			input:       workflowIDToolSchema,
			call:        s.callCancelWorkflow,
		},
		{
			name:        workflowStatusTool,
			description: workflowStatusToolDescription,
			input:       workflowIDToolSchema,
			call:        s.callWorkflowStatus,
		},
		{
			name:        listWorkflowsTool,
			description: listWorkflowsToolDescription,
			input:       listWorkflowsToolSchema,
			call:        s.callListWorkflows,
		},
		{
			name:        answerAcceptanceTool,
			description: answerAcceptanceToolDescription,
			input:       answerAcceptanceToolSchema,
			call:        s.callAnswerAcceptance,
		},
	}
}

func (s *v1) callListRoles(_ json.RawMessage, _ uuid.UUID) *toolResult {
	control := s.control()
	if control == nil {
		return errorResult(controlUnavailable)
	}

	roles, err := control.Roles.List()
	if err != nil {
		return errorResult(err.Error())
	}

	payload := make([]*rolePayload, 0, len(roles))

	for _, role := range roles {
		if role == nil {
			continue
		}

		item := &rolePayload{
			ID:             role.ID.String(),
			Name:           role.Name,
			Description:    role.Description,
			Effort:         role.Effort.String(),
			PauseForReview: role.PauseForReview,
			Inputs:         make([]*roleInputPayload, 0, len(role.Inputs)),
		}

		for _, key := range slices.Sorted(maps.Keys(role.Inputs)) {
			input := role.Inputs[key]
			if input == nil {
				continue
			}

			item.Inputs = append(item.Inputs, &roleInputPayload{
				Key:         key,
				Description: input.Description,
				Type:        input.Type,
				Default:     input.Default,
				Required:    input.Required,
				Options:     input.Options,
			})
		}

		payload = append(payload, item)
	}

	return controlResult(payload)
}

func (s *v1) callCreateWorkflow(arguments json.RawMessage, _ uuid.UUID) *toolResult {
	control := s.control()
	if control == nil {
		return errorResult(controlUnavailable)
	}

	args := createWorkflowArgs{}
	if err := parseControlArgs(arguments, &args); err != nil {
		return errorResult("cannot parse tool arguments: " + err.Error())
	}

	spec := &core_itf.ControlWorkflowSpec{
		ProjectDirPath: args.ProjectDirPath,
		Autostart:      args.Autostart,
		Steps:          make([]*core_itf.ControlStepSpec, 0, len(args.Steps)),
	}

	for _, step := range args.Steps {
		if step == nil {
			return errorResult("one of the steps is empty")
		}

		roleID := uuid.Nil
		if step.RoleID != "" {
			parsed, err := parseControlID("role", step.RoleID)
			if err != nil {
				return errorResult(err.Error())
			}
			roleID = parsed
		}

		spec.Steps = append(spec.Steps, &core_itf.ControlStepSpec{
			ClientID:        step.ClientID,
			Name:            step.Name,
			Prompt:          step.Prompt,
			RoleID:          roleID,
			Inputs:          step.Inputs,
			Effort:          step.Effort,
			Instructions:    step.Instructions,
			OutputStructure: step.OutputStructure,
			DependsOn:       step.DependsOn,
			AutoRetry:       step.AutoRetry,
			PauseForReview:  step.PauseForReview,
		})
	}

	ref, err := control.Control.CreateWorkflow(spec)
	if err != nil {
		return errorResult(err.Error())
	}

	return controlResult(&workflowRefPayload{
		WorkflowID: ref.WorkflowID.String(),
		StepIDs:    stringIDs(ref.StepIDs),
		Started:    ref.Started,
	})
}

func (s *v1) callStartWorkflow(arguments json.RawMessage, _ uuid.UUID) *toolResult {
	return s.callWorkflowAction(arguments, "started", func(control *core_itf.ControlPorts, workflowID uuid.UUID) error {
		return control.Coordinator.Run(workflowID)
	})
}

func (s *v1) callPauseWorkflow(arguments json.RawMessage, _ uuid.UUID) *toolResult {
	return s.callWorkflowAction(arguments, "paused", func(control *core_itf.ControlPorts, workflowID uuid.UUID) error {
		return control.Coordinator.Pause(workflowID)
	})
}

func (s *v1) callCancelWorkflow(arguments json.RawMessage, _ uuid.UUID) *toolResult {
	return s.callWorkflowAction(arguments, "cancelled", func(control *core_itf.ControlPorts, workflowID uuid.UUID) error {
		return control.Coordinator.Cancel(workflowID)
	})
}

func (s *v1) callWorkflowAction(
	arguments json.RawMessage,
	result string,
	action func(control *core_itf.ControlPorts, workflowID uuid.UUID) error,
) *toolResult {
	control := s.control()
	if control == nil {
		return errorResult(controlUnavailable)
	}

	args := controlWorkflowArgs{}
	if err := parseControlArgs(arguments, &args); err != nil {
		return errorResult("cannot parse tool arguments: " + err.Error())
	}

	workflowID, err := parseControlID("workflow", args.WorkflowID)
	if err != nil {
		return errorResult(err.Error())
	}

	if err := action(control, workflowID); err != nil {
		return errorResult(err.Error())
	}

	return controlResult(&workflowAckPayload{WorkflowID: args.WorkflowID, Result: result})
}

func (s *v1) callWorkflowStatus(arguments json.RawMessage, _ uuid.UUID) *toolResult {
	control := s.control()
	if control == nil {
		return errorResult(controlUnavailable)
	}

	args := controlWorkflowArgs{}
	if err := parseControlArgs(arguments, &args); err != nil {
		return errorResult("cannot parse tool arguments: " + err.Error())
	}

	workflowID, err := parseControlID("workflow", args.WorkflowID)
	if err != nil {
		return errorResult(err.Error())
	}

	state, err := control.Workflows.Status(workflowID)
	if err != nil {
		return errorResult(err.Error())
	}

	payload := &workflowStatePayload{
		WorkflowID:     state.ID.String(),
		Status:         string(state.Status),
		ProjectDirPath: state.ProjectDirPath,
		Steps:          make([]*stepStatePayload, 0, len(state.Steps)),
		TokensBilled:   state.TokensBilled,
		TokensInput:    state.TokensInput,
		TokensCached:   state.TokensCached,
		StartedAt:      momentText(state.StartedAt),
		CompletedAt:    momentText(state.CompletedAt),
	}

	for stepID, report := range state.Steps {
		payload.Steps = append(payload.Steps, stepState(stepID, report))
	}
	sort.Slice(payload.Steps, func(i, j int) bool { return payload.Steps[i].StepID < payload.Steps[j].StepID })

	return controlResult(payload)
}

func (s *v1) callListWorkflows(arguments json.RawMessage, _ uuid.UUID) *toolResult {
	control := s.control()
	if control == nil {
		return errorResult(controlUnavailable)
	}

	args := listWorkflowsArgs{}
	if err := parseControlArgs(arguments, &args); err != nil {
		return errorResult("cannot parse tool arguments: " + err.Error())
	}

	workflows, err := control.Control.ListWorkflows(args.Limit)
	if err != nil {
		return errorResult(err.Error())
	}

	payload := make([]*workflowSummaryPayload, 0, len(workflows))

	for _, workflow := range workflows {
		if workflow == nil {
			continue
		}

		payload = append(payload, &workflowSummaryPayload{
			WorkflowID:     workflow.ID.String(),
			Status:         string(workflow.Status),
			ProjectDirPath: workflow.ProjectDirPath,
			TotalStep:      len(workflow.Steps),
			StartedAt:      momentText(workflow.StartedAt),
			CompletedAt:    momentText(workflow.CompletedAt),
		})
	}

	return controlResult(payload)
}

func (s *v1) callAnswerAcceptance(arguments json.RawMessage, _ uuid.UUID) *toolResult {
	control := s.control()
	if control == nil {
		return errorResult(controlUnavailable)
	}

	args := answerReviewArgs{}
	if err := parseControlArgs(arguments, &args); err != nil {
		return errorResult("cannot parse tool arguments: " + err.Error())
	}

	stepID, err := parseControlID("step", args.StepID)
	if err != nil {
		return errorResult(err.Error())
	}

	if err := control.Workflows.AnswerReview(stepID, args.Accepted); err != nil {
		return errorResult(err.Error())
	}

	result := "rejected"
	if args.Accepted {
		result = "accepted"
	}

	return controlResult(&acceptanceAckPayload{
		StepID:   args.StepID,
		Accepted: args.Accepted,
		Result:   result,
	})
}

func parseControlID(kind, value string) (uuid.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, custom_error.Critical("invalid %s id %s: %v", kind, value, err)
	}

	return id, nil
}

func stringIDs(ids map[string]uuid.UUID) map[string]string {
	result := make(map[string]string, len(ids))
	for clientID, id := range ids {
		result[clientID] = id.String()
	}

	return result
}

func stepState(stepID uuid.UUID, report *core_itf.StepResult) *stepStatePayload {
	payload := &stepStatePayload{StepID: stepID.String()}
	if report == nil {
		return payload
	}

	payload.Name = report.Name
	payload.Status = string(report.Status)
	payload.Effort = report.Effort.String()

	if last := len(report.Handoffs) - 1; last >= 0 && report.Handoffs[last] != nil {
		payload.TLDR = report.Handoffs[last].TLDR
		payload.Outcome = report.Handoffs[last].Outcome
	}

	return payload
}

func momentText(at time.Time) string {
	if at.IsZero() {
		return ""
	}

	return at.Format(time.RFC3339)
}

func parseControlArgs(arguments json.RawMessage, target any) error {
	if len(arguments) == 0 {
		return nil
	}

	return json.Unmarshal(arguments, target)
}

func controlResult(payload any) *toolResult {
	raw, err := json.Marshal(payload)
	if err != nil {
		return errorResult("cannot encode the answer: " + err.Error())
	}

	return textResult(string(raw))
}
