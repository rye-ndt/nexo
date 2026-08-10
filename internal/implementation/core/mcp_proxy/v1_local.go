package mcp_proxy

import (
	"encoding/json"
	"net/http"

	"hexago/internal/helpers/constances"
	"hexago/internal/helpers/enums"
	core_itf "hexago/internal/interface/core"

	"github.com/google/uuid"
)

const (
	approvalTool = "request_approval"
	reportTool   = "report_task"
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

const (
	rejectedWithGuidance = "The operator rejected this and is sending their guidance as a separate message. " +
		"Wait for that message before continuing; do not act on the rejected approach."
	rejectedNoGuidance = "The operator rejected this and gave no guidance. " +
		"Do not retry the same approach; stop and report what is blocked."
	reportReceived = "Report received. This task is closed — stop now and take no further actions."
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

func (s *v1) serveLocal(w http.ResponseWriter, r *http.Request) {
	serveRPC(w, r, constances.GatewayLocalServer, s.localTools())
}

func (s *v1) localTools() []*rpcTool {
	tools := []*rpcTool{{
		name:        approvalTool,
		description: approvalToolDescription,
		input: objectSchema(map[string]any{
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
		}, "question", "options"),
		call: s.callApproval,
	}}

	if s.reporter != nil {
		tools = append(tools, &rpcTool{
			name:        reportTool,
			description: reportToolDescription,
			input: objectSchema(map[string]any{
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
			}, "status", "tldr", "outcome"),
			call: s.callReport,
		})
	}

	return tools
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
