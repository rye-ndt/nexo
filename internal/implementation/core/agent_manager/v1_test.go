package agent_manager

import (
	"testing"
	"time"

	"hexago/internal/helpers/enums"
	core_itf "hexago/internal/interface/core"
	input_itf "hexago/internal/interface/input"
	output_itf "hexago/internal/interface/output"

	"github.com/google/uuid"
)

type stubHarness struct{}

func (h *stubHarness) Auth() (string, error)                         { return "", nil }
func (h *stubHarness) SubmitAuthCode(string) error                   { return nil }
func (h *stubHarness) Logout() error                                 { return nil }
func (h *stubHarness) Status() (*input_itf.AgentStatus, error)       { return &input_itf.AgentStatus{}, nil }
func (h *stubHarness) SupportedModels() []enums.ModelName            { return []enums.ModelName{enums.Sonnet} }
func (h *stubHarness) Install(func(input_itf.InstallProgress)) error { return nil }
func (h *stubHarness) Uninstall() error                              { return nil }
func (h *stubHarness) Shutdown()                                     {}
func (h *stubHarness) Support(enums.ModelName) bool                  { return true }
func (h *stubHarness) Send(string, string) error                     { return nil }
func (h *stubHarness) Listen(string) (<-chan string, error)          { return nil, nil }
func (h *stubHarness) Alive(string) (time.Time, error)               { return time.Now(), nil }
func (h *stubHarness) Usage(string) (*input_itf.ContextUsage, error) { return nil, nil }
func (h *stubHarness) Activity(string) ([]input_itf.Activity, error) { return nil, nil }
func (h *stubHarness) Kill(string) error                             { return nil }

func (h *stubHarness) Spawn(
	enums.ModelName,
	enums.ThinkingLevel,
	[]string,
	string,
) (uuid.UUID, error) {
	return uuid.Must(uuid.NewV7()), nil
}

type stubLimit struct {
	limit int
}

func (c *stubLimit) MaxRunningAgents() int         { return c.limit }
func (c *stubLimit) SetMaxRunningAgents(int) error { return nil }
func (c *stubLimit) AgentDefaults() map[enums.Effort]*output_itf.AgentDefault {
	return nil
}

func (c *stubLimit) AgentDefault(enums.Effort) (*output_itf.AgentDefault, error) {
	return nil, nil
}

func (c *stubLimit) SetAgentDefault(enums.Effort, *output_itf.AgentDefault) error { return nil }
func (c *stubLimit) ModelPrice(enums.ModelName) *output_itf.TokenPrices           { return nil }
func (c *stubLimit) SetModelPrices(enums.ModelName, *output_itf.TokenPrices) error {
	return nil
}
func (c *stubLimit) Onboarded() bool                  { return true }
func (c *stubLimit) CompleteOnboarding() error        { return nil }
func (c *stubLimit) Language() enums.Language         { return enums.English }
func (c *stubLimit) SetLanguage(enums.Language) error { return nil }
func (c *stubLimit) Autopilot() bool                  { return false }
func (c *stubLimit) SetAutopilot(bool) error          { return nil }

type stubApprovals struct{}

func (stubApprovals) Awaiting(uuid.UUID) bool { return false }

func newManager(t *testing.T, limit int) core_itf.AgentManager {
	t.Helper()

	manager, err := InitV1(
		&input_itf.AgentManagerConfig{
			FreezeTimeout:        time.Minute,
			ConnectivityCacheTTL: time.Second,
		},
		nil,
		map[enums.AgentHarness]input_itf.AgentHarness{enums.ClaudeCode: &stubHarness{}},
		stubApprovals{},
		&stubLimit{limit: limit},
	)
	if err != nil {
		t.Fatalf("init agent manager: %v", err)
	}

	return manager
}

func request(offFleet bool) *core_itf.AgentRequest {
	return &core_itf.AgentRequest{
		Name:          enums.Sonnet,
		ThinkingLevel: enums.MedThinking,
		OffFleet:      offFleet,
	}
}

func TestFleetStopsAtTheLimitAndFreesTheSlotOnKill(t *testing.T) {
	manager := newManager(t, 2)

	first, err := manager.RequestInstance(request(false))
	if err != nil {
		t.Fatalf("first agent refused: %v", err)
	}

	if _, err := manager.RequestInstance(request(false)); err != nil {
		t.Fatalf("second agent refused: %v", err)
	}

	if _, err := manager.RequestInstance(request(false)); err == nil {
		t.Fatal("a third agent started while the limit was two")
	}

	if err := manager.Kill(first.ID); err != nil {
		t.Fatalf("kill: %v", err)
	}

	if _, err := manager.RequestInstance(request(false)); err != nil {
		t.Fatalf("the freed slot was not reused: %v", err)
	}
}

func TestKilledTwiceDoesNotHandOutASlotTwice(t *testing.T) {
	manager := newManager(t, 1)

	agent, err := manager.RequestInstance(request(false))
	if err != nil {
		t.Fatalf("first agent refused: %v", err)
	}

	for range 3 {
		if err := manager.Kill(agent.ID); err != nil {
			t.Fatalf("kill: %v", err)
		}
	}

	if _, err := manager.RequestInstance(request(false)); err != nil {
		t.Fatalf("the freed slot was not reused: %v", err)
	}

	if _, err := manager.RequestInstance(request(false)); err == nil {
		t.Fatal("repeated kills handed out more slots than the limit allows")
	}
}

func TestOffFleetAgentIgnoresTheLimitAndTakesNoSlot(t *testing.T) {
	manager := newManager(t, 1)

	if _, err := manager.RequestInstance(request(false)); err != nil {
		t.Fatalf("fleet agent refused: %v", err)
	}

	if _, err := manager.RequestInstance(request(true)); err != nil {
		t.Fatalf("off-fleet agent refused while the fleet was full: %v", err)
	}

	helper, err := manager.RequestInstance(request(true))
	if err != nil {
		t.Fatalf("second off-fleet agent refused: %v", err)
	}

	if err := manager.Kill(helper.ID); err != nil {
		t.Fatalf("kill: %v", err)
	}

	if _, err := manager.RequestInstance(request(false)); err == nil {
		t.Fatal("an off-fleet agent freed a fleet slot it never took")
	}
}
