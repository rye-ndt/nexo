package coordinator

import (
	"maps"
	"slices"
	"strings"

	"hexago/internal/helpers/constances"
	core_itf "hexago/internal/interface/core"
)

const gatedReviewInstruction = "\nA human operator reviews your report before anything downstream runs, " +
	"so nothing continues until they confirm it. Write the `outcome` field as a complete, " +
	"self-contained briefing for that operator: what you did and why, the plan you followed, " +
	"the decisions you took and the alternatives you rejected, the risks you see, and anything " +
	"that could surprise them. Assume they have not read the code or watched you work. " +
	"Also fill `approved_decisions`, `rejected_decisions`, `must_avoid`, `nuances` and " +
	"`known_gaps` wherever they apply.\n"

const approvalInstruction = "\nIf you reach a fork you should not settle alone, call the `request_approval` tool on the `" +
	constances.GatewayLocalServer +
	"` MCP server: a choice that would be expensive to undo, a decision the rest of the work has to be " +
	"built on, or a permission this task does not grant you. Ask the question plainly, lay out the " +
	"options you see with what each one costs, and the call will wait for a human, so you can sit on " +
	"it. If nobody answers in time it comes back as an error, and then you report what you were " +
	"blocked on rather than choosing for them. Asking is cheap and a wrong guess made quietly is " +
	"not, so raise it rather than picking one and hoping.\n"

const reportInstruction = "\nWhen you are finished, call the `report_task` tool on the `" +
	constances.GatewayLocalServer +
	"` MCP server exactly once, with status completed or failed and a complete, honest handover doc. " +
	"After the tool returns, stop.\n"

const tldrInstruction = "\nOne field of that doc is not for the next agent. `tldr` is read by a person who has " +
	"not seen this codebase and may not know what this task was for. Write exactly one sentence " +
	"saying what you did and how you did it, in plain words a non-programmer would follow: no file " +
	"paths, no function or type names, no jargon, no abbreviations from this project. It has to make " +
	"sense on its own, to someone who reads nothing else. If you failed, say in that one sentence " +
	"what you were trying to do and what stopped you.\n"

func buildPrompt(spec *core_itf.TaskSpec, status *core_itf.SessionStatus) string {
	b := &strings.Builder{}

	b.WriteString("# Task: " + spec.Name + "\n")

	if guidance := strings.TrimSpace(spec.ExtraGuidance); guidance != "" {
		b.WriteString("\n" + guidance + "\n")
	}

	b.WriteString("\nYou are already running inside " + status.WorkingDirPath + "; all work happens there.\n")

	writeHandovers(b, spec, status)

	if spec.ManualAcceptRequired {
		b.WriteString(gatedReviewInstruction)
	}

	b.WriteString(approvalInstruction)
	b.WriteString(reportInstruction)
	b.WriteString(tldrInstruction)

	return b.String()
}

func writeHandovers(b *strings.Builder, spec *core_itf.TaskSpec, status *core_itf.SessionStatus) {
	for _, dep := range spec.DependsOn {
		report, found := status.Tasks[dep]
		if !found {
			continue
		}

		for _, doc := range report.HandoverDocs {
			writeDoc(b, doc)
		}
	}
}

func writeDoc(b *strings.Builder, doc *core_itf.HandoverDoc) {
	b.WriteString("\n## Handover from \"" + doc.Task + "\"\n")
	b.WriteString("\nOutcome: " + doc.Outcome + "\n")

	sections := []struct {
		title string
		items map[string]string
	}{
		{"Blockers", doc.Blockers},
		{"Approved decisions", doc.ApprovedDecisions},
		{"Rejected decisions", doc.RejectedDecisions},
		{"Current behaviors", doc.CurrentBehaviors},
		{"Changed behaviors", doc.ChangedBehaviors},
		{"Must avoid", doc.MustAvoid},
		{"Nuances", doc.Nuances},
		{"Known gaps", doc.KnownGaps},
	}

	for _, section := range sections {
		if len(section.items) == 0 {
			continue
		}

		b.WriteString("\n### " + section.title + "\n")

		for _, key := range slices.Sorted(maps.Keys(section.items)) {
			b.WriteString("- " + key + ": " + section.items[key] + "\n")
		}
	}
}
