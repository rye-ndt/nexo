package mcp_proxy

import (
	"encoding/json"
	"net/http"

	"hexago/internal/helpers"
	"hexago/internal/helpers/constances"
	"hexago/internal/helpers/enums"
	core_itf "hexago/internal/interface/core"

	"github.com/google/uuid"
)

const (
	approvalTool = "request_approval"
	reportTool   = "report_task"
	draftTool    = "report_template"
)

const approvalToolDescription = `Ask the human operator to approve a decision or grant a permission, then wait for their answer.
Call this instead of asking the user directly: you are running non-interactively and have no other way to reach them.
Use it when a choice would be expensive to undo, when you need a decision locked in before continuing, or when you need
permission the current task does not already grant. The call blocks until the operator answers.
The operator may approve, which returns the option they picked, or reject, which returns approved=false and no option.`

const reportToolDescription = `Report the assigned task as finished. Call this exactly once, when the task is done.
Use status completed when the goal is met, failed when you are blocked and cannot meet it.
The handover doc you submit here is the only context the next agent gets, so rejected decisions
and things to avoid matter as much as the outcome itself.
The tldr is the exception: it is read by a person, not by an agent, so write it for someone
who has never seen this project.`

const draftToolDescription = `Submit the finished template. Call this exactly once, when every field is filled.
This call is the only way to hand the template back: text you write in the conversation is discarded.
The call is checked before it is accepted, so a template that is incomplete or inconsistent comes back
as an error describing what is wrong. Fix what it names and call again.`

const (
	rejectedWithGuidance = "The operator rejected this and is sending their guidance as a separate message. " +
		"Wait for that message before continuing; do not act on the rejected approach."
	rejectedNoGuidance = "The operator rejected this and gave no guidance. " +
		"Do not retry the same approach; stop and report what is blocked."
	reportReceived = "Report received. This task is closed — stop now and take no further actions."
	draftReceived  = "Template accepted. You are done — stop now and take no further actions."
)

type approvalArgs struct {
	Kind        string `json:"kind"`
	Question    string `json:"question"`
	Detail      string `json:"detail"`
	MultiSelect bool   `json:"multi_select"`
	Options     []struct {
		ID          string `json:"id"`
		Label       string `json:"label"`
		Description string `json:"description"`
	} `json:"options"`
}

type reportArgs struct {
	Status            string            `json:"status"`
	TLDR              string            `json:"tldr"`
	Outcome           string            `json:"outcome"`
	Blockers          map[string]string `json:"blockers"`
	ApprovedDecisions map[string]string `json:"approved_decisions"`
	RejectedDecisions map[string]string `json:"rejected_decisions"`
	CurrentBehaviors  map[string]string `json:"current_behaviors"`
	ChangedBehaviors  map[string]string `json:"changed_behaviors"`
	MustAvoid         map[string]string `json:"must_avoid"`
	Nuances           map[string]string `json:"nuances"`
	KnownGaps         map[string]string `json:"known_gaps"`
}

type templateParamArgs struct {
	Key         string   `json:"key"`
	Description string   `json:"description"`
	Type        string   `json:"type"`
	Required    bool     `json:"required"`
	Default     string   `json:"default"`
	Options     []string `json:"options"`
}

type templatePromptArgs struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Params and prompts arrive as arrays rather than the maps Template holds, so the
// tool schema can state what a key and a value are. They are folded into maps here.
type templateArgs struct {
	Name                 string                `json:"name"`
	Role                 string                `json:"role"`
	TaskLevel            string                `json:"task_level"`
	Retryable            bool                  `json:"retryable"`
	ManualAcceptRequired bool                  `json:"manual_accept_required"`
	Params               []*templateParamArgs  `json:"params"`
	SystemPrompts        []*templatePromptArgs `json:"system_prompts"`
	OutputStructure      string                `json:"output_structure"`
}

func (s *v1) serveLocal(w http.ResponseWriter, r *http.Request) {
	serveRPC(w, r, constances.GatewayLocalServer, s.localTools(agentFromHeader(r)))
}

var approvalToolSchema = objectSchema(map[string]any{
	"question": stringProp("The decision to be made, phrased as a question the operator can answer."),
	"kind": map[string]any{
		"type":        "string",
		"enum":        []string{enums.ApproveDecision.String(), enums.ApprovePermission.String()},
		"description": "decision to lock a choice in, permission to be allowed to do something.",
	},
	"detail": stringProp("Context the operator needs: what you found, what each option costs."),
	"options": map[string]any{
		"type":        "array",
		"minItems":    1,
		"description": "The choices available. Put your recommendation first.",
		"items": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id":          map[string]any{"type": "string"},
				"label":       map[string]any{"type": "string"},
				"description": map[string]any{"type": "string"},
			},
			"required": []string{"id", "label"},
		},
	},
	"multi_select": map[string]any{
		"type":        "boolean",
		"description": "Whether the operator may pick more than one option.",
	},
}, "question", "options")

var reportToolSchema = objectSchema(map[string]any{
	"status": map[string]any{
		"type":        "string",
		"enum":        []string{string(enums.TaskCompleted), string(enums.TaskFailed)},
		"description": "completed when the goal is met, failed when you are blocked.",
	},
	"tldr": stringProp("Exactly one sentence, written for a person who has not read the code and knows " +
		"nothing about this project: what you did and how you did it. Use plain words, " +
		"no file paths, no identifiers, no jargon. It must make sense on its own."),
	"outcome":            stringProp("What was achieved, or why the task failed."),
	"blockers":           handoverSection("What is blocking further progress, keyed by a short name."),
	"approved_decisions": handoverSection("Decisions the operator approved, keyed by a short name."),
	"rejected_decisions": handoverSection("Decisions the operator rejected, keyed by a short name."),
	"current_behaviors":  handoverSection("How the system behaves now, keyed by a short name."),
	"changed_behaviors":  handoverSection("Behaviors this task changed, keyed by a short name."),
	"must_avoid":         handoverSection("Approaches the next agent must not take, keyed by a short name."),
	"nuances":            handoverSection("Subtleties the next agent needs, keyed by a short name."),
	"known_gaps":         handoverSection("Work knowingly left undone, keyed by a short name."),
}, "status", "tldr", "outcome")

func (s *v1) localTools(agentID uuid.UUID) []*rpcTool {
	if drafter := s.drafter(); drafter != nil && drafter.Drafting(agentID) {
		return []*rpcTool{{
			name:        draftTool,
			description: draftToolDescription,
			input:       draftToolSchema,
			call:        s.callDraft,
		}}
	}

	tools := make([]*rpcTool, 0, 2)

	tools = append(tools, &rpcTool{
		name:        approvalTool,
		description: approvalToolDescription,
		input:       approvalToolSchema,
		call:        s.callApproval,
	})

	if s.reporter != nil {
		tools = append(tools, &rpcTool{
			name:        reportTool,
			description: reportToolDescription,
			input:       reportToolSchema,
			call:        s.callReport,
		})
	}

	return tools
}

var draftToolSchema = objectSchema(map[string]any{
	"name": stringProp("The template's name: two or three words naming the job it does."),
	"role": stringProp("One or two sentences saying what an agent built from this template does."),
	"task_level": map[string]any{
		"type":        "string",
		"enum":        helpers.Labels(enums.TaskLevels()),
		"description": "How much effort a node built from this template deserves.",
	},
	"retryable": map[string]any{
		"type":        "boolean",
		"description": "Whether a failed node should be retried automatically. False when a retry could repeat a side effect.",
	},
	"manual_accept_required": map[string]any{
		"type":        "boolean",
		"description": "Whether a human must read the report before anything downstream runs.",
	},
	"params": map[string]any{
		"type":        "array",
		"description": "The inputs a node fills in before it runs. Keep them few and each one load-bearing.",
		"items": objectSchema(map[string]any{
			"key":         stringProp("snake_case identifier the prompts reference."),
			"description": stringProp("What the person filling this in should type."),
			"type": map[string]any{
				"type": "string",
				"enum": helpers.Labels(enums.ParamTypes()),
			},
			"required": map[string]any{"type": "boolean"},
			"default":  stringProp("Optional starting value. Leave empty when there is no sane default."),
			"options": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "The choices, required when type is select and empty otherwise.",
			},
		}, "key", "description", "type"),
	},
	"system_prompts": map[string]any{
		"type":     "array",
		"minItems": 1,
		"description": "The prompt sections given to the agent. They are stored by key and read back " +
			"in alphabetical order, so name them so that order reads correctly. " +
			"Reference a param as {{key}}.",
		"items": objectSchema(map[string]any{
			"key":   stringProp("snake_case name for this section, such as base or constraints."),
			"value": stringProp("The prompt text itself."),
		}, "key", "value"),
	},
	"output_structure": stringProp(
		"The fields every node must report, one per line as `name: what it holds`, or empty when " +
			"the node should report in its own words. A name uses only letters, numbers and " +
			"underscores and cannot start with a number. A name with nothing after the colon opens " +
			"a group whose fields are indented two more spaces. A group whose first line starts " +
			"with `- ` is a list: it describes one element, holds exactly one item, and its fields " +
			"line up two spaces past the dash. Indent with spaces, never tabs.",
	),
}, "name", "role", "task_level", "system_prompts")

func (s *v1) callDraft(arguments json.RawMessage, agentID uuid.UUID) *toolResult {
	if agentID == uuid.Nil {
		return errorResult("cannot identify the calling agent")
	}

	drafter := s.drafter()
	if drafter == nil {
		return errorResult("no template is being drafted right now")
	}

	args := templateArgs{}
	if err := json.Unmarshal(arguments, &args); err != nil {
		return errorResult("cannot parse tool arguments: " + err.Error())
	}

	template := &core_itf.Template{
		Name:                 args.Name,
		Role:                 args.Role,
		TaskLevel:            enums.TaskLevel(args.TaskLevel),
		Retryable:            args.Retryable,
		ManualAcceptRequired: args.ManualAcceptRequired,
		OutputStructure:      args.OutputStructure,
		Params:               make(map[string]*core_itf.TemplateParams, len(args.Params)),
		SystemPrompts:        make(map[string]string, len(args.SystemPrompts)),
	}

	for _, param := range args.Params {
		if param == nil {
			return errorResult("one of the params is empty")
		}

		if _, taken := template.Params[param.Key]; taken {
			return errorResult("two params share the key " + param.Key)
		}

		template.Params[param.Key] = &core_itf.TemplateParams{
			Description: param.Description,
			Required:    param.Required,
			Type:        param.Type,
			Default:     param.Default,
			Options:     param.Options,
		}
	}

	for _, prompt := range args.SystemPrompts {
		if prompt == nil {
			return errorResult("one of the system prompts is empty")
		}

		if _, taken := template.SystemPrompts[prompt.Key]; taken {
			return errorResult("two system prompts share the key " + prompt.Key)
		}

		template.SystemPrompts[prompt.Key] = prompt.Value
	}

	if err := drafter.Deliver(agentID, template); err != nil {
		return errorResult(err.Error())
	}

	return textResult(draftReceived)
}

func handoverSection(description string) map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": map[string]any{"type": "string"},
		"description":          description,
	}
}

func (s *v1) callApproval(arguments json.RawMessage, agentID uuid.UUID) *toolResult {
	if agentID == uuid.Nil {
		return errorResult("cannot identify the calling agent")
	}

	args := approvalArgs{}
	if err := json.Unmarshal(arguments, &args); err != nil {
		return errorResult("cannot parse tool arguments: " + err.Error())
	}

	request := &core_itf.ApprovalRequest{
		AgentID:     agentID,
		Kind:        enums.ApprovalKind(args.Kind),
		Question:    args.Question,
		Detail:      args.Detail,
		MultiSelect: args.MultiSelect,
		Options:     make([]*core_itf.ApprovalOption, 0, len(args.Options)),
	}

	for _, option := range args.Options {
		request.Options = append(request.Options, &core_itf.ApprovalOption{
			ID:          option.ID,
			Label:       option.Label,
			Description: option.Description,
		})
	}

	answer, err := s.approvalBroker.Request(request)
	if err != nil {
		return errorResult(err.Error())
	}

	return answerResult(request, answer)
}

func (s *v1) callReport(arguments json.RawMessage, agentID uuid.UUID) *toolResult {
	if agentID == uuid.Nil {
		return errorResult("cannot identify the calling agent")
	}

	args := reportArgs{}
	if err := json.Unmarshal(arguments, &args); err != nil {
		return errorResult("cannot parse tool arguments: " + err.Error())
	}

	docs := []*core_itf.HandoverDoc{{
		TLDR:              args.TLDR,
		Outcome:           args.Outcome,
		Blockers:          args.Blockers,
		ApprovedDecisions: args.ApprovedDecisions,
		RejectedDecisions: args.RejectedDecisions,
		CurrentBehaviors:  args.CurrentBehaviors,
		ChangedBehaviors:  args.ChangedBehaviors,
		MustAvoid:         args.MustAvoid,
		Nuances:           args.Nuances,
		KnownGaps:         args.KnownGaps,
	}}

	if err := s.reporter.Report(agentID, enums.TaskStatus(args.Status), docs); err != nil {
		return errorResult(err.Error())
	}

	return textResult(reportReceived)
}

func answerResult(request *core_itf.ApprovalRequest, answer *core_itf.ApprovalAnswer) *toolResult {
	type selected struct {
		ID    string `json:"id"`
		Label string `json:"label"`
	}

	picked := []selected{}

	for _, id := range answer.OptionIDs {
		for _, option := range request.Options {
			if option.ID == id {
				picked = append(picked, selected{ID: option.ID, Label: option.Label})
				break
			}
		}
	}

	payload := map[string]any{
		"approved": answer.Approved,
		"selected": picked,
	}

	switch {
	case answer.Approved:
		payload["guidance"] = answer.Guidance
	case answer.Guidance != "":
		payload["note"] = rejectedWithGuidance
	default:
		payload["note"] = rejectedNoGuidance
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return errorResult("cannot encode the operator's answer: " + err.Error())
	}

	return textResult(string(raw))
}
