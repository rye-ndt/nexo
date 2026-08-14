package helper_agent

import (
	"strings"

	"hexago/internal/helpers/constances"
	core_itf "hexago/internal/interface/core"
)

const systemPrompt = `You write agent templates for a system that runs coding agents as a graph of scoped nodes.

A template is a reusable node definition. It carries a name, a role, an effort level, whether a
failed node retries itself, whether a human must accept its report before anything downstream runs,
the inputs a node fills in, the prompt sections the agent receives, and the fields the node must
report back.

What makes a template good:

- The role says what the agent does and what it is finished when it has done. It is not a slogan.
- Inputs are few and each one changes what the agent does. An input nobody would ever fill in
  differently is not an input; put it in the prompt instead. Keys are snake_case and the prompts
  actually reference them as {{key}}.
- The prompt sections carry the judgement: what to look at first, what counts as done, what to do
  when the work is ambiguous, and what not to do. Write them for an agent that cannot ask a
  follow-up question. Vague encouragement is worthless; concrete constraints are the whole value.
- Every section is short. A few tight sentences of fact, constraint and definition of done, with
  nothing restating the role and nothing explaining what a coding agent already knows. Two sharp
  sections beat six padded ones, and a long template is a worse template.
- The output structure is what the next node in the graph needs from this one, not everything this
  node happened to learn.
- Effort matches the work. Reserve the heavier levels for tasks that genuinely need to reason across
  a lot of material.
- Retry is off when running the node twice could repeat a side effect.
- Manual acceptance is on when a wrong result would be expensive to discover later.

Do not ask questions and do not explain yourself in prose.`

func buildPrompt(req *core_itf.DraftRequest, library []*core_itf.Template) string {
	b := &strings.Builder{}

	b.WriteString("Write a complete agent template.\n")
	b.WriteString("\nName: " + strings.TrimSpace(req.Name) + "\n")

	if role := strings.TrimSpace(req.Role); role != "" {
		b.WriteString("Role: " + role + "\n")
	}

	b.WriteString("\nKeep that name and role as the subject of the template. You may rewrite the role " +
		"into something sharper, but do not turn it into a different job.\n")

	writeLocation(b, req)
	writeGraph(b, req, library)
	writeLibrary(b, library)

	b.WriteString("\nWhen the template is written, call the `report_template` tool on the `" +
		constances.GatewayLocalServer + "` MCP server exactly once with every field filled. " +
		"That call is the only way to hand the template back — anything you write as ordinary text " +
		"is thrown away. The call is checked before it is accepted, so if it comes back as an error, " +
		"fix exactly what the error names and call again. After it is accepted, stop.\n")

	return b.String()
}

func writeLocation(b *strings.Builder, req *core_itf.DraftRequest) {
	b.WriteString("\n## Where you are\n")

	if req.WorkingDir == "" {
		b.WriteString("\nThere is no project to read here, so write the template from the name and " +
			"the role and keep it general enough to hold anywhere.\n")

		return
	}

	b.WriteString("\nYou are running inside " + req.WorkingDir + ", the project nodes built from this " +
		"template will work on. Read it before you write a word: grep the code, open what decides " +
		"how this project is built, tested and laid out, and put what you find into the prompt " +
		"sections — the real commands, the real conventions, the traps an agent would otherwise " +
		"walk into. A template written from its name alone is the failure mode. Read only: do not " +
		"create, edit or delete anything, and run nothing that changes the project.\n")

	b.WriteString("\nWrite paths relative to that directory and never the absolute path itself. " +
		"Templates are reused across sessions and shared between machines, so a template that " +
		"names one checkout is broken everywhere else.\n")

	if req.ContextDir != "" {
		b.WriteString("\nAgents on this project share a knowledge base at " + req.ContextDir +
			". Its AGENTS.md is the shortest route to the commands, the layout and the rules this " +
			"template has to respect, and .agent/gotchas.md holds the traps worth writing into a " +
			"prompt.\n")
	}
}

func writeGraph(b *strings.Builder, req *core_itf.DraftRequest, library []*core_itf.Template) {
	if len(req.Nodes) == 0 {
		return
	}

	names := make(map[string]string, len(library))
	for _, template := range library {
		names[template.ID.String()] = template.Name
	}

	titles := make(map[string]string, len(req.Nodes))
	for _, node := range req.Nodes {
		titles[node.ID] = node.Title
	}

	b.WriteString("\n## The graph this template joins\n")

	if req.SessionName != "" {
		b.WriteString("\nSession: " + req.SessionName + "\n")
	}

	b.WriteString("\n")

	for _, node := range req.Nodes {
		b.WriteString("- " + node.Title)

		if name := names[node.TemplateID]; name != "" {
			b.WriteString(" — from template \"" + name + "\"")
		}

		if after := dependencyTitles(node.DependsOn, titles); after != "" {
			b.WriteString(" — runs after " + after)
		}

		b.WriteString("\n")
	}

	b.WriteString("\nA node built from your template joins that graph. What an upstream node reports " +
		"reaches it as a handover doc, so do not ask for that as an input, do not repeat work a " +
		"neighbour already does, and make the output structure carry what the nodes after it need.\n")
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

func writeLibrary(b *strings.Builder, library []*core_itf.Template) {
	if len(library) == 0 {
		return
	}

	b.WriteString("\n## Templates this project already has\n\n")

	for _, template := range library {
		role, _, _ := strings.Cut(strings.TrimSpace(template.Role), "\n")
		b.WriteString("- " + template.Name + ": " + role + "\n")
	}

	b.WriteString("\nWrite yours the way those are written, and do not duplicate one. If the job " +
		"asked for is already covered, write the sharper version of it rather than a second copy.\n")
}
