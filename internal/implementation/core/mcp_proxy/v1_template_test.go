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

var errRefused = errors.New("`description` is empty. Say what an agent built from this role does.")

type fakeDeliverer struct {
	mu        sync.Mutex
	drafting  uuid.UUID
	delivered []*core_itf.Role
	err       error
}

func (d *fakeDeliverer) Blocked() string { return "" }

func (d *fakeDeliverer) Draft(_ *core_itf.DraftRequest) (*core_itf.Role, error) {
	return nil, nil
}

func (d *fakeDeliverer) Drafting(agentID uuid.UUID) bool {
	return agentID != uuid.Nil && d.drafting == agentID
}

func (d *fakeDeliverer) Deliver(_ uuid.UUID, role *core_itf.Role) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.err != nil {
		return d.err
	}

	d.delivered = append(d.delivered, role)

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
		"name":        "Code reviewer",
		"description": "Reads a diff and reports the defects it can prove.",
		"effort":      "deep",
		"instructions": []map[string]any{
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

// The two report tools are mutually exclusive per agent: a workflow agent must not
// be able to reach for report_role, and the agent drafting a role must not
// be able to close a step it was never given.
func TestRoleToolIsOfferedOnlyToTheDraftingAgent(t *testing.T) {
	proxy := newTestProxy(t, newFakeMCPStore(), &fakeHttpCli{})
	proxy.reporter = &fakeReporter{}

	drafting := uuid.New()
	proxy.TrackRoleHelper(&fakeDeliverer{drafting: drafting})

	names := func(agentID uuid.UUID) []string {
		found := []string{}
		for _, tool := range proxy.localTools(agentID) {
			found = append(found, tool.name)
		}

		return found
	}

	workflow := names(uuid.New())

	if slices.Contains(workflow, draftTool) {
		t.Error("a workflow agent is offered the role tool")
	}

	if !slices.Contains(workflow, reportTool) {
		t.Error("a workflow agent lost the report tool")
	}

	drafter := names(drafting)

	if !slices.Contains(drafter, draftTool) {
		t.Error("the drafting agent is not offered the role tool")
	}

	if slices.Contains(drafter, reportTool) {
		t.Error("the drafting agent is offered the step report tool")
	}
}

func TestRoleToolRequiresTheFieldsARoleCannotDoWithout(t *testing.T) {
	required, ok := draftToolSchema["required"].([]string)
	if !ok {
		t.Fatal("the role tool has no required list")
	}

	for _, field := range []string{"name", "description", "effort", "instructions"} {
		if !slices.Contains(required, field) {
			t.Errorf("the role tool does not require %q", field)
		}
	}
}

func TestRoleToolFoldsArraysIntoTheMapsARoleHolds(t *testing.T) {
	proxy := newTestProxy(t, newFakeMCPStore(), &fakeHttpCli{})
	deliverer := &fakeDeliverer{}
	proxy.TrackRoleHelper(deliverer)

	arguments := draftArgs(t, map[string]any{
		"inputs": []map[string]any{
			{"key": "diff_path", "description": "Where the diff is", "type": "text", "required": true},
		},
		"instructions": []map[string]any{
			{"key": "base", "value": "Review {{diff_path}}."},
			{"key": "limits", "value": "Report only what you can prove."},
		},
	})

	if result := proxy.callDraft(arguments, uuid.New()); result.IsError {
		t.Fatalf("role call failed: %v", result.Content)
	}

	if deliverer.count() != 1 {
		t.Fatalf("delivered %d roles, want 1", deliverer.count())
	}

	role := deliverer.delivered[0]

	if len(role.Instructions) != 2 {
		t.Errorf("kept %d instructions, want 2", len(role.Instructions))
	}

	if role.Instructions["limits"] != "Report only what you can prove." {
		t.Errorf("the instruction keyed by %q did not survive", "limits")
	}

	input := role.Inputs["diff_path"]
	if input == nil {
		t.Fatal("the input keyed by diff_path did not survive")
	}

	if !input.Required || input.Type != "text" {
		t.Errorf("input = %+v, want a required text input", input)
	}
}

func TestRoleToolRejectsDuplicateKeys(t *testing.T) {
	proxy := newTestProxy(t, newFakeMCPStore(), &fakeHttpCli{})
	deliverer := &fakeDeliverer{}
	proxy.TrackRoleHelper(deliverer)

	arguments := draftArgs(t, map[string]any{
		"inputs": []map[string]any{
			{"key": "target", "description": "first", "type": "text"},
			{"key": "target", "description": "second", "type": "text"},
		},
	})

	result := proxy.callDraft(arguments, uuid.New())

	if !result.IsError {
		t.Error("two inputs sharing a key were accepted")
	}

	if deliverer.count() != 0 {
		t.Error("a role with duplicate input keys reached the helper")
	}
}

// A refusal from the helper has to reach the agent as a tool error, because that
// error is the only thing telling it what to fix before calling again.
func TestRoleToolPassesTheHelpersRefusalBackToTheAgent(t *testing.T) {
	proxy := newTestProxy(t, newFakeMCPStore(), &fakeHttpCli{})
	proxy.TrackRoleHelper(&fakeDeliverer{err: errRefused})

	result := proxy.callDraft(draftArgs(t, nil), uuid.New())

	if !result.IsError {
		t.Fatal("a refused role was reported as accepted")
	}

	if len(result.Content) == 0 || result.Content[0].Text != errRefused.Error() {
		t.Errorf("result = %v, want the helper's reason", result.Content)
	}
}

func TestRoleToolRejectsAnUnidentifiedAgent(t *testing.T) {
	proxy := newTestProxy(t, newFakeMCPStore(), &fakeHttpCli{})
	deliverer := &fakeDeliverer{}
	proxy.TrackRoleHelper(deliverer)

	if result := proxy.callDraft(draftArgs(t, nil), uuid.Nil); !result.IsError {
		t.Error("a role from an unidentifiable agent was accepted")
	}

	if deliverer.count() != 0 {
		t.Error("a role from an unidentifiable agent reached the helper")
	}
}
