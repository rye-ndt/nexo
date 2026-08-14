package helper_agent

import (
	"strings"
	"testing"

	core_itf "hexago/internal/interface/core"

	"github.com/google/uuid"
)

func TestPromptCarriesTheProjectTheGraphAndTheLibrary(t *testing.T) {
	explorer := uuid.New()

	prompt := buildPrompt(&core_itf.DraftRequest{
		Name:        "Migration planner",
		Role:        "Plans the migration.",
		SessionName: "Port the storage layer",
		WorkingDir:  "/Users/rye/code/thing",
		ContextDir:  "/Users/rye/code/thing/.nexo",
		Nodes: []*core_itf.DraftGraphNode{
			{ID: "a", Title: "Map the call sites", TemplateID: explorer.String()},
			{ID: "b", Title: "Rewrite the adapters", DependsOn: []string{"a", "gone"}},
		},
	}, []*core_itf.Template{
		{ID: explorer, Name: "Explorer", Role: "Maps the call sites.\nSecond line."},
	})

	for _, want := range []string{
		"/Users/rye/code/thing",
		"/Users/rye/code/thing/.nexo",
		"Port the storage layer",
		`Map the call sites — from template "Explorer"`,
		"Rewrite the adapters — runs after Map the call sites",
		"- Explorer: Maps the call sites.",
		"report_template",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("the prompt never says %q:\n%s", want, prompt)
		}
	}

	if strings.Contains(prompt, "Second line.") {
		t.Fatalf("a multi-line role should be cut to its first line:\n%s", prompt)
	}
}

func TestPromptSaysThereIsNothingToReadWithoutAProject(t *testing.T) {
	prompt := buildPrompt(&core_itf.DraftRequest{Name: "Migration planner"}, nil)

	for _, unwanted := range []string{"Read it before you write a word", "The graph this template joins"} {
		if strings.Contains(prompt, unwanted) {
			t.Fatalf("there is no project and no graph, yet the prompt says %q:\n%s", unwanted, prompt)
		}
	}

	if !strings.Contains(prompt, "There is no project to read here") {
		t.Fatalf("a template drafted from Settings should say so:\n%s", prompt)
	}
}
