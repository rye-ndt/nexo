package helper_agent

import (
	"strings"

	"hexago/internal/helpers"
	"hexago/internal/helpers/constances"
	core_itf "hexago/internal/interface/core"
)

const systemPrompt = `You write agent roles for a system that runs coding agents as a graph of scoped steps.

A role is a reusable step definition. It carries a name, a description, an effort level, whether a
failed step retries itself, whether a human must review the result before anything downstream runs,
the inputs a step fills in, the instructions the agent receives, and the fields the step must
report back.

What makes a role good:

- The description says what the agent does and what it is finished when it has done. It is not a slogan.
- Inputs are few and each one changes what the agent does. An input nobody would ever fill in
  differently is not an input; put it in the instructions instead. Keys are snake_case and the
  instructions actually reference them as {{key}}.
- The instructions carry the judgement: what to look at first, what counts as done, what to do
  when the work is ambiguous, and what not to do. Write them for an agent that cannot ask a
  follow-up question. Vague encouragement is worthless; concrete constraints are the whole value.
- Every section is short. A few tight sentences of fact, constraint and definition of done, with
  nothing restating the description and nothing explaining what a coding agent already knows. Two
  sharp sections beat six padded ones, and a long role is a worse role.
- The output structure is what the next step in the graph needs from this one, not everything this
  step happened to learn.
- Effort matches the work. Reserve the heavier levels for steps that genuinely need to reason across
  a lot of material.
- Retry is off when running the step twice could repeat a side effect.
- Pause for review is on when a wrong result would be expensive to discover later.

Do not ask questions and do not explain yourself in prose.`

func buildPrompt(req *core_itf.DraftRequest, library []*core_itf.Role) string {
	b := &strings.Builder{}

	b.WriteString("Write a complete agent role.\n")
	b.WriteString("\nName: " + strings.TrimSpace(req.Name) + "\n")

	if description := strings.TrimSpace(req.Description); description != "" {
		b.WriteString("Description: " + description + "\n")
	}

	b.WriteString("\nKeep that name and description as the subject of the role. You may rewrite the " +
		"description into something sharper, but do not turn it into a different job.\n")

	writeLocation(b, req)
	writeGraph(b, req, library)
	writeLibrary(b, library)

	b.WriteString("\nWhen the role is written, call the `report_role` tool on the `" +
		constances.GatewayLocalServer + "` MCP server exactly once with every field filled. " +
		"That call is the only way to hand the role back — anything you write as ordinary text " +
		"is thrown away. The call is checked before it is accepted, so if it comes back as an error, " +
		"fix exactly what the error names and call again. After it is accepted, stop.\n")

	return b.String()
}

func writeLocation(b *strings.Builder, req *core_itf.DraftRequest) {
	b.WriteString("\n## Where you are\n")

	if req.ProjectDir == "" {
		b.WriteString("\nThere is no project to read here, so write the role from the name and " +
			"the description and keep it general enough to hold anywhere.\n")

		return
	}

	b.WriteString("\nYou are running inside " + req.ProjectDir + ", the project folder that steps built " +
		"from this role will work on. Read it before you write a word: grep the code, open what decides " +
		"how this project is built, tested and laid out, and put what you find into the " +
		"instructions — the real commands, the real conventions, the traps an agent would otherwise " +
		"walk into. A role written from its name alone is the failure mode. Read only: do not " +
		"create, edit or delete anything, and run nothing that changes the project.\n")

	b.WriteString("\nWrite paths relative to that folder and never the absolute path itself. " +
		"Roles are reused across workflows and shared between machines, so a role that " +
		"names one checkout is broken everywhere else.\n")

	b.WriteString("\nAgents on this project share a knowledge base at " + helpers.KnowledgeDir(req.ProjectDir) +
		". Its AGENTS.md is the shortest route to the commands, the layout and the rules this " +
		"role has to respect, and .agent/gotchas.md holds the traps worth writing into the " +
		"instructions.\n")
}

func writeGraph(b *strings.Builder, req *core_itf.DraftRequest, library []*core_itf.Role) {
	if len(req.Steps) == 0 {
		return
	}

	names := make(map[string]string, len(library))
	for _, role := range library {
		names[role.ID.String()] = role.Name
	}

	titles := make(map[string]string, len(req.Steps))
	for _, step := range req.Steps {
		titles[step.ID] = step.Title
	}

	b.WriteString("\n## The graph this role joins\n")

	if req.WorkflowName != "" {
		b.WriteString("\nWorkflow: " + req.WorkflowName + "\n")
	}

	b.WriteString("\n")

	for _, step := range req.Steps {
		b.WriteString("- " + step.Title)

		if name := names[step.RoleID]; name != "" {
			b.WriteString(" — from role \"" + name + "\"")
		}

		if after := dependencyTitles(step.DependsOn, titles); after != "" {
			b.WriteString(" — runs after " + after)
		}

		b.WriteString("\n")
	}

	b.WriteString("\nA step built from your role joins that graph. What an upstream step reports " +
		"reaches it as a handoff, so do not ask for that as an input, do not repeat work a " +
		"neighbour already does, and make the output structure carry what the steps after it need.\n")
}

func dependencyTitles(dependsOn []string, titles map[string]string) string {
	names := make([]string, 0, len(dependsOn))

	for _, id := range dependsOn {
		if title := titles[id]; title != "" {
			names = append(names, title)
		}
	}

	return strings.Join(names, ", ")
}

func writeLibrary(b *strings.Builder, library []*core_itf.Role) {
	if len(library) == 0 {
		return
	}

	b.WriteString("\n## Roles this project already has\n\n")

	for _, role := range library {
		summary, _, _ := strings.Cut(strings.TrimSpace(role.Description), "\n")
		b.WriteString("- " + role.Name + ": " + summary + "\n")
	}

	b.WriteString("\nWrite yours the way those are written, and do not duplicate one. If the job " +
		"asked for is already covered, write the sharper version of it rather than a second copy.\n")
}
