package coordinator_test

import (
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"hexago/internal/helpers/enums"
	"hexago/internal/implementation/core/coordinator"
	"hexago/internal/implementation/core/session_manager"
	"hexago/internal/implementation/input/storage"
	"hexago/internal/implementation/input/workspace_history"
	"hexago/internal/implementation/output/message_queue"
	core_itf "hexago/internal/interface/core"
	input_itf "hexago/internal/interface/input"

	"github.com/google/uuid"
)

type recordingAgents struct {
	mu      sync.Mutex
	prompts map[string]string
}

func (a *recordingAgents) SupportedAgents() (map[enums.AgentHarness][]enums.ModelName, error) {
	return nil, nil
}

func (a *recordingAgents) Admin(enums.AgentHarness) (input_itf.AgentAdmin, error) { return nil, nil }

func (a *recordingAgents) RequestInstance(*core_itf.AgentRequest) (*core_itf.Agent, error) {
	return &core_itf.Agent{ID: uuid.New(), HealthStatus: enums.Healthy}, nil
}

func (a *recordingAgents) Send(_ uuid.UUID, message string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	name, _, _ := strings.Cut(message, "\n")
	a.prompts[strings.TrimSpace(strings.TrimPrefix(name, "# Task:"))] = message

	return nil
}

func (a *recordingAgents) promptFor(name string) string {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.prompts[name]
}

func (a *recordingAgents) Listen(uuid.UUID) (<-chan string, error) { return nil, nil }

func (a *recordingAgents) ContextUsage(uuid.UUID) (*input_itf.ContextUsage, error) {
	return nil, nil
}

func (a *recordingAgents) Activity(uuid.UUID) ([]input_itf.Activity, error) { return nil, nil }

func (a *recordingAgents) Kill(uuid.UUID) error { return nil }

func (a *recordingAgents) HeartBeat(uuid.UUID) error { return nil }

func restartConfig() *input_itf.SessionConfig {
	return &input_itf.SessionConfig{
		HeartbeatTimeout:       time.Hour,
		HeartbeatScanInterval:  time.Hour,
		AgentHeartbeatInterval: time.Hour,
	}
}

func managerOn(t *testing.T, store input_itf.Storage) core_itf.SessionManager {
	t.Helper()

	sessions, err := session_manager.InitV1(restartConfig(), store.TaskStore(), message_queue.InitV1())
	if err != nil {
		t.Fatalf("init session manager: %v", err)
	}

	return sessions
}

func waitFor(t *testing.T, what string, done func() bool) {
	t.Helper()

	deadline := time.Now().Add(5 * time.Second)

	for time.Now().Before(deadline) {
		if done() {
			return
		}

		time.Sleep(10 * time.Millisecond)
	}

	t.Fatalf("timed out waiting for %s", what)
}

// The run is stopped the way a power cut stops it: the manager holding it is dropped
// without a pause, and a second one is built over the same database.
func TestAnInterruptedRunComesBackPausedAndFinishesOnResume(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "harness.db")

	store, err := storage.New(dbPath)
	if err != nil {
		t.Fatalf("init storage: %v", err)
	}

	before := managerOn(t, store)

	sessionID, err := before.NewSession(&core_itf.InitSession{WorkingDirPath: t.TempDir()})
	if err != nil {
		t.Fatalf("new session: %v", err)
	}

	first, err := before.AddTask(sessionID, &core_itf.AddTask{
		Name:       "map the ports",
		AgentSpecs: &core_itf.AgentRequest{Name: enums.Sonnet},
	})
	if err != nil {
		t.Fatalf("add first task: %v", err)
	}

	second, err := before.AddTask(sessionID, &core_itf.AddTask{
		Name:       "wire it up",
		DependsOn:  []uuid.UUID{first},
		AgentSpecs: &core_itf.AgentRequest{Name: enums.Sonnet},
	})
	if err != nil {
		t.Fatalf("add second task: %v", err)
	}

	finished := uuid.New()
	if err := before.Assign(first, finished); err != nil {
		t.Fatalf("assign first task: %v", err)
	}

	err = before.Report(finished, enums.TaskCompleted, []*core_itf.HandoverDoc{{
		TLDR:     "listed the ports the coordinator calls",
		Outcome:  "seven ports, each with the call that reaches it",
		Nuances:  map[string]string{"scope": "stayed inside interface/core"},
		Blockers: map[string]string{},
	}})
	if err != nil {
		t.Fatalf("report first task: %v", err)
	}

	interrupted := uuid.New()
	if err := before.Assign(second, interrupted); err != nil {
		t.Fatalf("assign second task: %v", err)
	}

	before.Stop()

	after := managerOn(t, store)
	t.Cleanup(after.Stop)

	if err := after.Restore(); err != nil {
		t.Fatalf("restore: %v", err)
	}

	status, err := after.Status(sessionID)
	if err != nil {
		t.Fatalf("status after restore: %v", err)
	}

	if status.Status != enums.SessionPaused {
		t.Fatalf("restored session is %s, want %s", status.Status, enums.SessionPaused)
	}

	if got := status.Tasks[first].Status; got != enums.TaskCompleted {
		t.Fatalf("finished task came back %s, want %s", got, enums.TaskCompleted)
	}

	if got := status.Tasks[second].Status; got != enums.TaskNotTaken {
		t.Fatalf("interrupted task came back %s, want %s", got, enums.TaskNotTaken)
	}

	docs := status.Tasks[first].HandoverDocs
	if len(docs) != 1 || docs[0].Outcome != "seven ports, each with the call that reaches it" {
		t.Fatalf("handover doc did not survive the restart: %+v", docs)
	}

	ready, err := after.ReadyTasks(sessionID)
	if err != nil {
		t.Fatalf("ready tasks while paused: %v", err)
	}

	if len(ready) != 0 {
		t.Fatalf("a paused session offered %d tasks, want none", len(ready))
	}

	history, err := workspace_history.InitV1(filepath.Join(t.TempDir(), "sessions"))
	if err != nil {
		t.Fatalf("init workspace history: %v", err)
	}

	agents := &recordingAgents{prompts: map[string]string{}}

	coord, err := coordinator.InitV1(restartConfig(), after, agents, history, silentLogger{})
	if err != nil {
		t.Fatalf("init coordinator: %v", err)
	}

	t.Cleanup(coord.Stop)

	if err := coord.Run(sessionID); err != nil {
		t.Fatalf("resume the restored session: %v", err)
	}

	waitFor(t, "the parked task to be prompted again", func() bool {
		return agents.promptFor("wire it up") != ""
	})

	prompt := agents.promptFor("wire it up")
	if !strings.Contains(prompt, "seven ports, each with the call that reaches it") {
		t.Fatalf("the resumed prompt lost the upstream handover doc:\n%s", prompt)
	}

	if !strings.Contains(prompt, "stayed inside interface/core") {
		t.Fatalf("the resumed prompt lost the upstream nuances:\n%s", prompt)
	}

	resumed, err := after.Status(sessionID)
	if err != nil {
		t.Fatalf("status after resume: %v", err)
	}

	if resumed.Status != enums.SessionProcessing {
		t.Fatalf("resumed session is %s, want %s", resumed.Status, enums.SessionProcessing)
	}
}
