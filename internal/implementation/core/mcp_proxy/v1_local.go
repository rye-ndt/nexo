package mcp_proxy

import (
	"encoding/json"
	"net/http"
	"slices"

	"hexago/internal/helpers"
	"hexago/internal/helpers/constances"
	"hexago/internal/helpers/enums"
	core_itf "hexago/internal/interface/core"

	"github.com/google/uuid"
)

const (
	approvalTool = "request_approval"
	reportTool   = "report_step"
	draftTool    = "report_role"
)

const approvalToolDescription = `Ask the human operator to approve a decision or grant a permission, then wait for their answer.
Call this instead of asking the user directly: you are running non-interactively and have no other way to reach them.
Use it when a choice would be expensive to undo, when you need a decision locked in before continuing, or when you need
permission the current step does not already grant. The call blocks until the operator answers.
The operator may approve, which returns the option they picked, or reject, which returns approved=false and no option.
Always name the one option you recommend: it is shown first and it is the answer taken on your behalf when the operator
has handed the run over to autopilot.`

const reportToolDescription = `Report the assigned step as finished. Call this exactly once, when the step is done.
Use status completed when the goal is met, failed when you are blocked and cannot meet it.
The handoff you submit here is the only context the next agent gets, so rejected decisions
and things to avoid matter as much as the outcome itself.
The tldr is the exception: it is read by a person, not by an agent, so write it for someone
who has never seen this project.`

const draftToolDescription = `Submit the finished role. Call this exactly once, when every field is filled.
This call is the only way to hand the role back: text you write in the conversation is discarded.
The call is checked before it is accepted, so a role that is incomplete or inconsistent comes back
as an error describing what is wrong. Fix what it names and call again.`

const (
	rejectedWithGuidance = "The operator rejected this and is sending their guidance as a separate message. " +
		"Wait for that message before continuing; do not act on the rejected approach."
	rejectedNoGuidance = "The operator rejected this and gave no guidance. " +
		"Do not retry the same approach; stop and report what is blocked."
	reportReceived = "Report received. This step is closed — stop now and take no further actions."
	draftReceived  = "Role accepted. You are done — stop now and take no further actions."
)

type approvalArgs struct {
	Kind              string `json:"kind"`
	Question          string `json:"question"`
	Detail            string `json:"detail"`
	MultiSelect       bool   `json:"multi_select"`
	RecommendedOption string `json:"recommended_option_id"`
	Options           []struct {
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

type roleInputArgs struct {
	Key         string   `json:"key"`
	Description string   `json:"description"`
	Type        string   `json:"type"`
	Required    bool     `json:"required"`
	Default     string   `json:"default"`
	Options     []string `json:"options"`
}

type roleInstructionArgs struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Inputs and instructions arrive as arrays rather than the maps Role holds, so the
// tool schema can state what a key and a value are. They are folded into maps here.
type roleArgs struct {
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Effort          string                 `json:"effort"`
	Retryable       bool                   `json:"retryable"`
	PauseForReview  bool                   `json:"pause_for_review"`
	Inputs          []*roleInputArgs       `json:"inputs"`
	Instructions    []*roleInstructionArgs `json:"instructions"`
	OutputStructure string                 `json:"output_structure"`
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
	"recommended_option_id": stringProp(
		"The id of the one option you recommend. It is shown first, and under autopilot it is " +
			"the answer taken on your behalf, so recommend the option you would pick yourself."),
	"options": map[string]any{
		"type":        "array",
		"minItems":    1,
		"description": "The choices available.",
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
}, "question", "options", "recommended_option_id")

var reportToolSchema = objectSchema(map[string]any{
	"status": map[string]any{
		"type":        "string",
		"enum":        []string{string(enums.StepCompleted), string(enums.StepFailed)},
		"description": "completed when the goal is met, failed when you are blocked.",
	},
	"tldr": stringProp("Exactly one sentence, written for a person who has not read the code and knows " +
		"nothing about this project: what you did and how you did it. Use plain words, " +
		"no file paths, no identifiers, no jargon. It must make sense on its own."),
	"outcome":            stringProp("What was achieved, or why the step failed."),
	"blockers":           handoffSection("What is blocking further progress, keyed by a short name."),
	"approved_decisions": handoffSection("Decisions the operator approved, keyed by a short name."),
	"rejected_decisions": handoffSection("Decisions the operator rejected, keyed by a short name."),
	"current_behaviors":  handoffSection("How the system behaves now, keyed by a short name."),
	"changed_behaviors":  handoffSection("Behaviors this step changed, keyed by a short name."),
	"must_avoid":         handoffSection("Approaches the next agent must not take, keyed by a short name."),
	"nuances":            handoffSection("Subtleties the next agent needs, keyed by a short name."),
	"known_gaps":         handoffSection("Work knowingly left undone, keyed by a short name."),
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
	"name":        stringProp("The role's name: two or three words naming the job it does."),
	"description": stringProp("One or two sentences saying what an agent built from this role does."),
	"effort": map[string]any{
		"type":        "string",
		"enum":        helpers.Labels(enums.Efforts()),
		"description": "How much effort a step built from this role deserves.",
	},
	"retryable": map[string]any{
		"type":        "boolean",
		"description": "Whether a failed step should be retried automatically. False when a retry could repeat a side effect.",
	},
	"pause_for_review": map[string]any{
		"type":        "boolean",
		"description": "Whether a human must review the result before anything downstream runs.",
	},
	"inputs": map[string]any{
		"type":        "array",
		"description": "The inputs a step fills in before it runs. Keep them few and each one load-bearing.",
		"items": objectSchema(map[string]any{
			"key":         stringProp("snake_case identifier the instructions reference."),
			"description": stringProp("What the person filling this in should type."),
			"type": map[string]any{
				"type": "string",
				"enum": helpers.Labels(enums.InputTypes()),
			},
			"required": map[string]any{"type": "boolean"},
			"default":  stringProp("Optional starting value, a comma-separated list when type is multiselect. Leave empty when there is no sane default."),
			"options": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "The choices, required when type is select or multiselect and empty otherwise.",
			},
		}, "key", "description", "type"),
	},
	"instructions": map[string]any{
		"type":     "array",
		"minItems": 1,
		"description": "The instruction sections given to the agent. They are stored by key and read " +
			"back in alphabetical order, so name them so that order reads correctly. " +
			"Reference an input as {{key}}.",
		"items": objectSchema(map[string]any{
			"key":   stringProp("snake_case name for this section, such as base or constraints."),
			"value": stringProp("The instruction text itself."),
		}, "key", "value"),
	},
	"output_structure": stringProp(
		"The fields every step must report, one per line as `name: what it holds`, or empty when " +
			"the step should report in its own words. A name uses only letters, numbers and " +
			"underscores and cannot start with a number. A name with nothing after the colon opens " +
			"a group whose fields are indented two more spaces. A group whose first line starts " +
			"with `- ` is a list: it describes one element, holds exactly one item, and its fields " +
			"line up two spaces past the dash. Indent with spaces, never tabs.",
	),
}, "name", "description", "effort", "instructions")

func (s *v1) callDraft(arguments json.RawMessage, agentID uuid.UUID) *toolResult {
	if agentID == uuid.Nil {
		return errorResult("cannot identify the calling agent")
	}

	drafter := s.drafter()
	if drafter == nil {
		return errorResult("no role is being drafted right now")
	}

	args := roleArgs{}
	if err := json.Unmarshal(arguments, &args); err != nil {
		return errorResult("cannot parse tool arguments: " + err.Error())
	}

	role := &core_itf.Role{
		Name:            args.Name,
		Description:     args.Description,
		Effort:          enums.Effort(args.Effort),
		Retryable:       args.Retryable,
		PauseForReview:  args.PauseForReview,
		OutputStructure: args.OutputStructure,
		Inputs:          make(map[string]*core_itf.RoleInputs, len(args.Inputs)),
		Instructions:    make(map[string]string, len(args.Instructions)),
	}

	for _, input := range args.Inputs {
		if input == nil {
			return errorResult("one of the inputs is empty")
		}

		if _, taken := role.Inputs[input.Key]; taken {
			return errorResult("two inputs share the key " + input.Key)
		}

		role.Inputs[input.Key] = &core_itf.RoleInputs{
			Description: input.Description,
			Required:    input.Required,
			Type:        input.Type,
			Default:     input.Default,
			Options:     input.Options,
		}
	}

	for _, instruction := range args.Instructions {
		if instruction == nil {
			return errorResult("one of the instructions is empty")
		}

		if _, taken := role.Instructions[instruction.Key]; taken {
			return errorResult("two instructions share the key " + instruction.Key)
		}

		role.Instructions[instruction.Key] = instruction.Value
	}

	if err := drafter.Deliver(agentID, role); err != nil {
		return errorResult(err.Error())
	}

	return textResult(draftReceived)
}

func handoffSection(description string) map[string]any {
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
			Recommended: option.ID == args.RecommendedOption,
		})
	}

	if !slices.ContainsFunc(request.Options, func(option *core_itf.ApprovalOption) bool {
		return option.Recommended
	}) {
		return errorResult("recommended_option_id must be the id of one of the options you offered")
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

	docs := []*core_itf.Handoff{{
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

	if err := s.reporter.Report(agentID, enums.StepStatus(args.Status), docs); err != nil {
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
