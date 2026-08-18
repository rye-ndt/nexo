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
	listTemplatesTool    = "list_templates"
	createSessionTool    = "create_session"
	startSessionTool     = "start_session"
	pauseSessionTool     = "pause_session"
	cancelSessionTool    = "cancel_session"
	sessionStatusTool    = "session_status"
	listSessionsTool     = "list_sessions"
	answerAcceptanceTool = "answer_acceptance"
)

const controlUnavailable = "this app is not ready to take control calls yet"

const listTemplatesToolDescription = `List the task templates this app already has.
A template fixes what an agent does: its role, how much effort a node built from it deserves, whether a
human must read the report before anything downstream runs, and the params a node fills in before it runs.
Read this before create_session: giving a task a template_id and its params is better than writing the
whole role out again in a raw prompt.`

const createSessionToolDescription = `Create a session: a graph of scoped agent tasks over one working directory.
Every task is one node with its own prompt; a node runs once every node it depends on has finished, and what a
finished node learned is carried into its dependents' prompts.
You name each task with your own client_id, and depends_on refers to those client_ids from this same call —
they are local to the call, not ids the app already knows. The answer maps every client_id to the real task id.
With autostart false the graph is only staged and nothing runs until start_session; leave autostart out to let
the app's own setting decide.`

const startSessionToolDescription = `Start running a session's graph, and the way to resume one too:
a session that was paused, or that was interrupted when the app closed, picks up from the tasks that never
finished rather than rerunning the ones that did.
The call returns as soon as the run is under way, so poll session_status to follow it.`

const pauseSessionToolDescription = `Pause a running session. The tasks in flight are stopped and no further task is
handed out, while everything already finished is kept. Call start_session to pick the run back up where it stopped.`

const cancelSessionToolDescription = `Cancel a session for good. Running tasks are killed and nothing else is handed
out; unlike pause, this cannot be resumed. Whatever the agents already wrote into the working directory stays on disk.`

const sessionStatusToolDescription = `Read where a session stands: its own status, every task with its status and the
one-sentence tldr its agent left, the tokens billed so far, and when it started and finished.
This is how you follow a run — poll it rather than treating a start_session call as the work being done.
A task sitting in awaiting_accept is waiting for you to call answer_acceptance.`

const listSessionsToolDescription = `List the most recent sessions, newest first, with their status and working
directory. Use it to find the session id the other tools need. limit caps how many come back; leave it out for the
app's default.`

const answerAcceptanceToolDescription = `Answer the accept gate on a task that finished and is waiting on a human.
Accepting marks the task completed and releases everything downstream of it; rejecting marks it failed, so its
dependents stay blocked. Only a task whose status is awaiting_accept has a gate to answer.`

type controlSessionArgs struct {
	SessionID string `json:"session_id"`
}

type listSessionsArgs struct {
	Limit int `json:"limit"`
}

type answerAcceptanceArgs struct {
	TaskID   string `json:"task_id"`
	Accepted bool   `json:"accepted"`
}

type controlTaskArgs struct {
	ClientID             string            `json:"client_id"`
	Name                 string            `json:"name"`
	Prompt               string            `json:"prompt"`
	TemplateID           string            `json:"template_id"`
	Params               map[string]string `json:"params"`
	TaskLevel            string            `json:"task_level"`
	SystemPrompts        []string          `json:"system_prompts"`
	OutputStructure      string            `json:"output_structure"`
	DependsOn            []string          `json:"depends_on"`
	AutoRetry            bool              `json:"auto_retry"`
	ManualAcceptRequired bool              `json:"manual_accept_required"`
}

type createSessionArgs struct {
	WorkingDirPath string             `json:"working_dir_path"`
	ContextDirPath string             `json:"context_dir_path"`
	Autostart      *bool              `json:"autostart"`
	Tasks          []*controlTaskArgs `json:"tasks"`
}

type templateParamPayload struct {
	Key         string   `json:"key"`
	Description string   `json:"description,omitempty"`
	Type        string   `json:"type,omitempty"`
	Default     string   `json:"default,omitempty"`
	Required    bool     `json:"required"`
	Options     []string `json:"options,omitempty"`
}

type templatePayload struct {
	ID                   string                  `json:"id"`
	Name                 string                  `json:"name"`
	Role                 string                  `json:"role"`
	TaskLevel            string                  `json:"task_level"`
	ManualAcceptRequired bool                    `json:"manual_accept_required"`
	Params               []*templateParamPayload `json:"params"`
}

type sessionRefPayload struct {
	SessionID string            `json:"session_id"`
	TaskIDs   map[string]string `json:"task_ids"`
	Started   bool              `json:"started"`
}

type taskStatePayload struct {
	TaskID    string `json:"task_id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	TaskLevel string `json:"task_level,omitempty"`
	TLDR      string `json:"tldr,omitempty"`
	Outcome   string `json:"outcome,omitempty"`
}

type sessionStatePayload struct {
	SessionID      string              `json:"session_id"`
	Status         string              `json:"status"`
	WorkingDirPath string              `json:"working_dir_path"`
	ContextDirPath string              `json:"context_dir_path,omitempty"`
	Tasks          []*taskStatePayload `json:"tasks"`
	TokensBilled   int                 `json:"tokens_billed"`
	TokensInput    int                 `json:"tokens_input"`
	TokensCached   int                 `json:"tokens_cached"`
	StartedAt      string              `json:"started_at,omitempty"`
	CompletedAt    string              `json:"completed_at,omitempty"`
}

type sessionSummaryPayload struct {
	SessionID      string `json:"session_id"`
	Status         string `json:"status"`
	WorkingDirPath string `json:"working_dir_path"`
	TotalTask      int    `json:"total_task"`
	StartedAt      string `json:"started_at,omitempty"`
	CompletedAt    string `json:"completed_at,omitempty"`
}

type sessionAckPayload struct {
	SessionID string `json:"session_id"`
	Result    string `json:"result"`
}

type acceptanceAckPayload struct {
	TaskID   string `json:"task_id"`
	Accepted bool   `json:"accepted"`
	Result   string `json:"result"`
}

func (s *v1) TrackSessionControl(control core_itf.SessionControl) {
	s.locker.Lock()
	defer s.locker.Unlock()

	s.sessionControl = control
}

func (s *v1) control() core_itf.SessionControl {
	s.locker.RLock()
	defer s.locker.RUnlock()

	return s.sessionControl
}

func (s *v1) serveControl(w http.ResponseWriter, r *http.Request) {
	serveRPC(w, r, constances.ControlLocalServer, s.controlTools())
}

var listTemplatesToolSchema = objectSchema(map[string]any{})

var createSessionToolSchema = objectSchema(map[string]any{
	"working_dir_path": stringProp("Absolute path to the directory the agents work in. Every node in the session shares it."),
	"context_dir_path": stringProp("Optional absolute path holding read-only material the agents may consult."),
	"autostart": map[string]any{
		"type":        "boolean",
		"description": "True to run the graph immediately, false to only stage it and wait for start_session. Leave it out to follow the app's own setting.",
	},
	"tasks": map[string]any{
		"type":        "array",
		"minItems":    1,
		"description": "The nodes of the graph, each one a task scoped tightly enough that a single agent can finish it.",
		"items": objectSchema(map[string]any{
			"client_id": stringProp("Your own short name for this node, unique within this call. depends_on refers to it."),
			"name":      stringProp("Two or three words naming what this node does, read by a person."),
			"prompt":    stringProp("What this node's agent must do. Say the goal and what done looks like, not the steps."),
			"template_id": stringProp("Optional id from list_templates. The template supplies the role, effort and " +
				"output structure, and the prompt narrows them to this node."),
			"params": map[string]any{
				"type":                 "object",
				"additionalProperties": map[string]any{"type": "string"},
				"description":          "Values for the template's params, keyed by the param key.",
			},
			"task_level": map[string]any{
				"type":        "string",
				"enum":        helpers.Labels(enums.TaskLevels()),
				"description": "How much effort this node deserves. Leave it out to keep the template's level.",
			},
			"system_prompts": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "Extra prompt sections for this node, on top of whatever the template already says.",
			},
			"output_structure": stringProp("The fields this node must report, one per line as `name: what it holds`. " +
				"Leave it empty to report in its own words or to keep the template's structure."),
			"depends_on": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "The client_ids of the nodes that must finish first. Their handover docs become this node's context.",
			},
			"auto_retry": map[string]any{
				"type":        "boolean",
				"description": "Whether a failed run of this node is retried automatically. Keep it false when a retry could repeat a side effect.",
			},
			"manual_accept_required": map[string]any{
				"type":        "boolean",
				"description": "Whether a human must accept this node's report before anything downstream runs.",
			},
		}, "client_id", "name", "prompt"),
	},
}, "working_dir_path", "tasks")

var sessionIDToolSchema = objectSchema(map[string]any{
	"session_id": stringProp("The id returned by create_session or listed by list_sessions."),
}, "session_id")

var listSessionsToolSchema = objectSchema(map[string]any{
	"limit": map[string]any{
		"type":        "integer",
		"minimum":     1,
		"description": "How many sessions to return, newest first. Leave it out for the app's default.",
	},
})

var answerAcceptanceToolSchema = objectSchema(map[string]any{
	"task_id": stringProp("The id of the task waiting in awaiting_accept, as reported by session_status."),
	"accepted": map[string]any{
		"type":        "boolean",
		"description": "True to accept the report and release the nodes downstream, false to mark the task failed.",
	},
}, "task_id", "accepted")

func (s *v1) controlTools() []*rpcTool {
	return []*rpcTool{
		{
			name:        listTemplatesTool,
			description: listTemplatesToolDescription,
			input:       listTemplatesToolSchema,
			call:        s.callListTemplates,
		},
		{
			name:        createSessionTool,
			description: createSessionToolDescription,
			input:       createSessionToolSchema,
			call:        s.callCreateSession,
		},
		{
			name:        startSessionTool,
			description: startSessionToolDescription,
			input:       sessionIDToolSchema,
			call:        s.callStartSession,
		},
		{
			name:        pauseSessionTool,
			description: pauseSessionToolDescription,
			input:       sessionIDToolSchema,
			call:        s.callPauseSession,
		},
		{
			name:        cancelSessionTool,
			description: cancelSessionToolDescription,
			input:       sessionIDToolSchema,
			call:        s.callCancelSession,
		},
		{
			name:        sessionStatusTool,
			description: sessionStatusToolDescription,
			input:       sessionIDToolSchema,
			call:        s.callSessionStatus,
		},
		{
			name:        listSessionsTool,
			description: listSessionsToolDescription,
			input:       listSessionsToolSchema,
			call:        s.callListSessions,
		},
		{
			name:        answerAcceptanceTool,
			description: answerAcceptanceToolDescription,
			input:       answerAcceptanceToolSchema,
			call:        s.callAnswerAcceptance,
		},
	}
}

func (s *v1) callListTemplates(_ json.RawMessage, _ uuid.UUID) *toolResult {
	control := s.control()
	if control == nil {
		return errorResult(controlUnavailable)
	}

	templates, err := control.ListTemplates()
	if err != nil {
		return errorResult(err.Error())
	}

	payload := make([]*templatePayload, 0, len(templates))

	for _, template := range templates {
		if template == nil {
			continue
		}

		item := &templatePayload{
			ID:                   template.ID.String(),
			Name:                 template.Name,
			Role:                 template.Role,
			TaskLevel:            template.TaskLevel.String(),
			ManualAcceptRequired: template.ManualAcceptRequired,
			Params:               make([]*templateParamPayload, 0, len(template.Params)),
		}

		for _, key := range slices.Sorted(maps.Keys(template.Params)) {
			param := template.Params[key]
			if param == nil {
				continue
			}

			item.Params = append(item.Params, &templateParamPayload{
				Key:         key,
				Description: param.Description,
				Type:        param.Type,
				Default:     param.Default,
				Required:    param.Required,
				Options:     param.Options,
			})
		}

		payload = append(payload, item)
	}

	return controlResult(payload)
}

func (s *v1) callCreateSession(arguments json.RawMessage, _ uuid.UUID) *toolResult {
	control := s.control()
	if control == nil {
		return errorResult(controlUnavailable)
	}

	args := createSessionArgs{}
	if err := parseControlArgs(arguments, &args); err != nil {
		return errorResult("cannot parse tool arguments: " + err.Error())
	}

	spec := &core_itf.ControlSessionSpec{
		WorkingDirPath: args.WorkingDirPath,
		ContextDirPath: args.ContextDirPath,
		Autostart:      args.Autostart,
		Tasks:          make([]*core_itf.ControlTaskSpec, 0, len(args.Tasks)),
	}

	for _, task := range args.Tasks {
		if task == nil {
			return errorResult("one of the tasks is empty")
		}

		templateID := uuid.Nil
		if task.TemplateID != "" {
			parsed, err := parseControlID("template", task.TemplateID)
			if err != nil {
				return errorResult(err.Error())
			}
			templateID = parsed
		}

		spec.Tasks = append(spec.Tasks, &core_itf.ControlTaskSpec{
			ClientID:             task.ClientID,
			Name:                 task.Name,
			Prompt:               task.Prompt,
			TemplateID:           templateID,
			Params:               task.Params,
			TaskLevel:            task.TaskLevel,
			SystemPrompts:        task.SystemPrompts,
			OutputStructure:      task.OutputStructure,
			DependsOn:            task.DependsOn,
			AutoRetry:            task.AutoRetry,
			ManualAcceptRequired: task.ManualAcceptRequired,
		})
	}

	ref, err := control.CreateSession(spec)
	if err != nil {
		return errorResult(err.Error())
	}

	return controlResult(&sessionRefPayload{
		SessionID: ref.SessionID.String(),
		TaskIDs:   stringIDs(ref.TaskIDs),
		Started:   ref.Started,
	})
}

func (s *v1) callStartSession(arguments json.RawMessage, _ uuid.UUID) *toolResult {
	return s.callSessionAction(arguments, "started", func(control core_itf.SessionControl, sessionID uuid.UUID) error {
		return control.StartSession(sessionID)
	})
}

func (s *v1) callPauseSession(arguments json.RawMessage, _ uuid.UUID) *toolResult {
	return s.callSessionAction(arguments, "paused", func(control core_itf.SessionControl, sessionID uuid.UUID) error {
		return control.PauseSession(sessionID)
	})
}

func (s *v1) callCancelSession(arguments json.RawMessage, _ uuid.UUID) *toolResult {
	return s.callSessionAction(arguments, "cancelled", func(control core_itf.SessionControl, sessionID uuid.UUID) error {
		return control.CancelSession(sessionID)
	})
}

func (s *v1) callSessionAction(
	arguments json.RawMessage,
	result string,
	action func(control core_itf.SessionControl, sessionID uuid.UUID) error,
) *toolResult {
	control := s.control()
	if control == nil {
		return errorResult(controlUnavailable)
	}

	args := controlSessionArgs{}
	if err := parseControlArgs(arguments, &args); err != nil {
		return errorResult("cannot parse tool arguments: " + err.Error())
	}

	sessionID, err := parseControlID("session", args.SessionID)
	if err != nil {
		return errorResult(err.Error())
	}

	if err := action(control, sessionID); err != nil {
		return errorResult(err.Error())
	}

	return controlResult(&sessionAckPayload{SessionID: args.SessionID, Result: result})
}

func (s *v1) callSessionStatus(arguments json.RawMessage, _ uuid.UUID) *toolResult {
	control := s.control()
	if control == nil {
		return errorResult(controlUnavailable)
	}

	args := controlSessionArgs{}
	if err := parseControlArgs(arguments, &args); err != nil {
		return errorResult("cannot parse tool arguments: " + err.Error())
	}

	sessionID, err := parseControlID("session", args.SessionID)
	if err != nil {
		return errorResult(err.Error())
	}

	state, err := control.SessionState(sessionID)
	if err != nil {
		return errorResult(err.Error())
	}

	payload := &sessionStatePayload{
		SessionID:      state.ID.String(),
		Status:         string(state.Status),
		WorkingDirPath: state.WorkingDirPath,
		ContextDirPath: state.ContextDirPath,
		Tasks:          make([]*taskStatePayload, 0, len(state.Tasks)),
		TokensBilled:   state.TokensBilled,
		TokensInput:    state.TokensInput,
		TokensCached:   state.TokensCached,
		StartedAt:      momentText(state.StartedAt),
		CompletedAt:    momentText(state.CompletedAt),
	}

	for taskID, report := range state.Tasks {
		payload.Tasks = append(payload.Tasks, taskState(taskID, report))
	}
	sort.Slice(payload.Tasks, func(i, j int) bool { return payload.Tasks[i].TaskID < payload.Tasks[j].TaskID })

	return controlResult(payload)
}

func (s *v1) callListSessions(arguments json.RawMessage, _ uuid.UUID) *toolResult {
	control := s.control()
	if control == nil {
		return errorResult(controlUnavailable)
	}

	args := listSessionsArgs{}
	if err := parseControlArgs(arguments, &args); err != nil {
		return errorResult("cannot parse tool arguments: " + err.Error())
	}

	sessions, err := control.ListSessions(args.Limit)
	if err != nil {
		return errorResult(err.Error())
	}

	payload := make([]*sessionSummaryPayload, 0, len(sessions))

	for _, session := range sessions {
		if session == nil {
			continue
		}

		payload = append(payload, &sessionSummaryPayload{
			SessionID:      session.ID.String(),
			Status:         string(session.Status),
			WorkingDirPath: session.WorkingDirPath,
			TotalTask:      len(session.Tasks),
			StartedAt:      momentText(session.StartedAt),
			CompletedAt:    momentText(session.CompletedAt),
		})
	}

	return controlResult(payload)
}

func (s *v1) callAnswerAcceptance(arguments json.RawMessage, _ uuid.UUID) *toolResult {
	control := s.control()
	if control == nil {
		return errorResult(controlUnavailable)
	}

	args := answerAcceptanceArgs{}
	if err := parseControlArgs(arguments, &args); err != nil {
		return errorResult("cannot parse tool arguments: " + err.Error())
	}

	taskID, err := parseControlID("task", args.TaskID)
	if err != nil {
		return errorResult(err.Error())
	}

	if err := control.AnswerAcceptance(taskID, args.Accepted); err != nil {
		return errorResult(err.Error())
	}

	result := "rejected"
	if args.Accepted {
		result = "accepted"
	}

	return controlResult(&acceptanceAckPayload{
		TaskID:   args.TaskID,
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

func taskState(taskID uuid.UUID, report *core_itf.TaskReport) *taskStatePayload {
	payload := &taskStatePayload{TaskID: taskID.String()}
	if report == nil {
		return payload
	}

	payload.Name = report.Name
	payload.Status = string(report.Status)
	payload.TaskLevel = report.TaskLevel.String()

	if last := len(report.HandoverDocs) - 1; last >= 0 && report.HandoverDocs[last] != nil {
		payload.TLDR = report.HandoverDocs[last].TLDR
		payload.Outcome = report.HandoverDocs[last].Outcome
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
