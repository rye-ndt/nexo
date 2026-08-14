package mcp_proxy

import (
	"encoding/json"
	"errors"
	"slices"
	"sync"
	"testing"

	core_itf "hexago/internal/interface/core"

	"github.com/google/uuid"
)

var errRefused = errors.New("`role` is empty. Say what an agent built from this template does.")

type fakeDeliverer struct {
	mu        sync.Mutex
	drafting  uuid.UUID
	delivered []*core_itf.Template
	err       error
}

func (d *fakeDeliverer) Blocked() string { return "" }

func (d *fakeDeliverer) Draft(_ *core_itf.DraftRequest) (*core_itf.Template, error) {
	return nil, nil
}

func (d *fakeDeliverer) Drafting(agentID uuid.UUID) bool {
	return agentID != uuid.Nil && d.drafting == agentID
}

func (d *fakeDeliverer) Deliver(_ uuid.UUID, template *core_itf.Template) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.err != nil {
		return d.err
	}

	d.delivered = append(d.delivered, template)

	return nil
}

func (d *fakeDeliverer) count() int {
	d.mu.Lock()
	defer d.mu.Unlock()

	return len(d.delivered)
}

func draftArgs(t *testing.T, overrides map[string]any) json.RawMessage {
	t.Helper()

	args := map[string]any{
		"name":       "Code reviewer",
		"role":       "Reads a diff and reports the defects it can prove.",
		"task_level": "heavy_task",
		"system_prompts": []map[string]any{
			{"key": "base", "value": "Review the diff."},
		},
	}

	for key, value := range overrides {
		args[key] = value
	}

	raw, err := json.Marshal(args)
	if err != nil {
		t.Fatalf("encode arguments: %v", err)
	}

	return raw
}

// The two report tools are mutually exclusive per agent: a session agent must not
// be able to reach for report_template, and the agent drafting a template must not
// be able to close a task it was never given.
func TestTemplateToolIsOfferedOnlyToTheDraftingAgent(t *testing.T) {
	proxy := newTestProxy(t, newFakeMCPStore(), &fakeHttpCli{})
	proxy.reporter = &fakeReporter{}

	drafting := uuid.New()
	proxy.TrackTemplateHelper(&fakeDeliverer{drafting: drafting})

	names := func(agentID uuid.UUID) []string {
		found := []string{}
		for _, tool := range proxy.localTools(agentID) {
			found = append(found, tool.name)
		}

		return found
	}

	session := names(uuid.New())

	if slices.Contains(session, draftTool) {
		t.Error("a session agent is offered the template tool")
	}

	if !slices.Contains(session, reportTool) {
		t.Error("a session agent lost the report tool")
	}

	drafter := names(drafting)

	if !slices.Contains(drafter, draftTool) {
		t.Error("the drafting agent is not offered the template tool")
	}

	if slices.Contains(drafter, reportTool) {
		t.Error("the drafting agent is offered the task report tool")
	}
}

func TestTemplateToolRequiresTheFieldsATemplateCannotDoWithout(t *testing.T) {
	required, ok := draftToolSchema["required"].([]string)
	if !ok {
		t.Fatal("the template tool has no required list")
	}

	for _, field := range []string{"name", "role", "task_level", "system_prompts"} {
		if !slices.Contains(required, field) {
			t.Errorf("the template tool does not require %q", field)
		}
	}
}

func TestTemplateToolFoldsArraysIntoTheMapsATemplateHolds(t *testing.T) {
	proxy := newTestProxy(t, newFakeMCPStore(), &fakeHttpCli{})
	deliverer := &fakeDeliverer{}
	proxy.TrackTemplateHelper(deliverer)

	arguments := draftArgs(t, map[string]any{
		"params": []map[string]any{
			{"key": "diff_path", "description": "Where the diff is", "type": "text", "required": true},
		},
		"system_prompts": []map[string]any{
			{"key": "base", "value": "Review {{diff_path}}."},
			{"key": "limits", "value": "Report only what you can prove."},
		},
	})

	if result := proxy.callDraft(arguments, uuid.New()); result.IsError {
		t.Fatalf("template call failed: %v", result.Content)
	}

	if deliverer.count() != 1 {
		t.Fatalf("delivered %d templates, want 1", deliverer.count())
	}

	template := deliverer.delivered[0]

	if len(template.SystemPrompts) != 2 {
		t.Errorf("kept %d system prompts, want 2", len(template.SystemPrompts))
	}

	if template.SystemPrompts["limits"] != "Report only what you can prove." {
		t.Errorf("system prompt keyed by %q did not survive", "limits")
	}

	param := template.Params["diff_path"]
	if param == nil {
		t.Fatal("the param keyed by diff_path did not survive")
	}

	if !param.Required || param.Type != "text" {
		t.Errorf("param = %+v, want a required text param", param)
	}
}

func TestTemplateToolRejectsDuplicateKeys(t *testing.T) {
	proxy := newTestProxy(t, newFakeMCPStore(), &fakeHttpCli{})
	deliverer := &fakeDeliverer{}
	proxy.TrackTemplateHelper(deliverer)

	arguments := draftArgs(t, map[string]any{
		"params": []map[string]any{
			{"key": "target", "description": "first", "type": "text"},
			{"key": "target", "description": "second", "type": "text"},
		},
	})

	result := proxy.callDraft(arguments, uuid.New())

	if !result.IsError {
		t.Error("two params sharing a key were accepted")
	}

	if deliverer.count() != 0 {
		t.Error("a template with duplicate param keys reached the helper")
	}
}

// A refusal from the helper has to reach the agent as a tool error, because that
// error is the only thing telling it what to fix before calling again.
func TestTemplateToolPassesTheHelpersRefusalBackToTheAgent(t *testing.T) {
	proxy := newTestProxy(t, newFakeMCPStore(), &fakeHttpCli{})
	proxy.TrackTemplateHelper(&fakeDeliverer{err: errRefused})

	result := proxy.callDraft(draftArgs(t, nil), uuid.New())

	if !result.IsError {
		t.Fatal("a refused template was reported as accepted")
	}

	if len(result.Content) == 0 || result.Content[0].Text != errRefused.Error() {
		t.Errorf("result = %v, want the helper's reason", result.Content)
	}
}

func TestTemplateToolRejectsAnUnidentifiedAgent(t *testing.T) {
	proxy := newTestProxy(t, newFakeMCPStore(), &fakeHttpCli{})
	deliverer := &fakeDeliverer{}
	proxy.TrackTemplateHelper(deliverer)

	if result := proxy.callDraft(draftArgs(t, nil), uuid.Nil); !result.IsError {
		t.Error("a template from an unidentifiable agent was accepted")
	}

	if deliverer.count() != 0 {
		t.Error("a template from an unidentifiable agent reached the helper")
	}
}
