package mcp_proxy

import (
	"encoding/json"
	"testing"

	core_itf "hexago/internal/interface/core"

	"github.com/google/uuid"
)

type recordingBroker struct {
	seen *core_itf.ApprovalRequest
}

func (b *recordingBroker) TrackAutopilot(core_itf.AutopilotReader) {}

func (b *recordingBroker) Request(req *core_itf.ApprovalRequest) (*core_itf.ApprovalAnswer, error) {
	b.seen = req

	return &core_itf.ApprovalAnswer{Approved: true, OptionIDs: []string{req.Options[0].ID}}, nil
}

func (b *recordingBroker) Pending() []*core_itf.ApprovalRequest  { return nil }
func (b *recordingBroker) Answer(*core_itf.ApprovalAnswer) error { return nil }
func (b *recordingBroker) Awaiting(uuid.UUID) bool               { return false }
func (b *recordingBroker) Stop()                                 {}

func newApprovalProxy(t *testing.T, broker core_itf.ApprovalBroker) *v1 {
	t.Helper()

	proxy := newTestProxy(t, newFakeMCPStore(), &fakeHttpCli{})
	proxy.approvalBroker = broker

	return proxy
}

func approvalCall(t *testing.T, recommended string) json.RawMessage {
	t.Helper()

	raw, err := json.Marshal(map[string]any{
		"question":              "where do we store it?",
		"recommended_option_id": recommended,
		"options": []map[string]any{
			{"id": "sqlite", "label": "Store it in SQLite"},
			{"id": "files", "label": "Store it on disk"},
		},
	})
	if err != nil {
		t.Fatalf("encode arguments: %v", err)
	}

	return raw
}

func TestApprovalCallFlagsTheOptionTheAgentRecommended(t *testing.T) {
	broker := &recordingBroker{}
	proxy := newApprovalProxy(t, broker)

	result := proxy.callApproval(approvalCall(t, "files"), uuid.New())
	if result.IsError {
		t.Fatalf("a call naming one of its own options was refused: %+v", result)
	}

	if broker.seen == nil {
		t.Fatal("the request never reached the broker")
	}

	for _, option := range broker.seen.Options {
		if want := option.ID == "files"; option.Recommended != want {
			t.Fatalf("option %q recommended = %v, want %v", option.ID, option.Recommended, want)
		}
	}
}

func TestApprovalCallNamingAnOptionThatIsNotOfferedIsRefused(t *testing.T) {
	broker := &recordingBroker{}
	proxy := newApprovalProxy(t, broker)

	for name, recommended := range map[string]string{"unknown option": "redis", "no option": ""} {
		t.Run(name, func(t *testing.T) {
			result := proxy.callApproval(approvalCall(t, recommended), uuid.New())

			if !result.IsError {
				t.Fatalf("the call was accepted, want a refusal: %+v", result)
			}
		})
	}
}
