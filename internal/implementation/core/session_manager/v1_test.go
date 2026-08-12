package session_manager

import (
	"sync"
	"testing"
	"time"

	"hexago/internal/helpers/enums"
	"hexago/internal/implementation/output/message_queue"
	core_itf "hexago/internal/interface/core"
	input_itf "hexago/internal/interface/input"
	output_itf "hexago/internal/interface/output"

	"github.com/google/uuid"
)

type fakeStore struct {
	mu      sync.Mutex
	tasks   []*input_itf.TaskEntity
	reports []*input_itf.TaskReportEntity
}

func (s *fakeStore) SaveTaskHistory(
	_ []*input_itf.SessionEntity,
	tasks []*input_itf.TaskEntity,
	reports []*input_itf.TaskReportEntity,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tasks = append(s.tasks, tasks...)
	s.reports = append(s.reports, reports...)

	return nil
}

func (s *fakeStore) lastStatusFor(taskID uuid.UUID) enums.TaskStatus {
	s.mu.Lock()
	defer s.mu.Unlock()

	status := enums.TaskStatus("")

	for _, task := range s.tasks {
		if task.ID == taskID {
			status = task.Status
		}
	}

	return status
}

func (s *fakeStore) lastReportFor(taskID uuid.UUID) *input_itf.TaskReportEntity {
	s.mu.Lock()
	defer s.mu.Unlock()

	var report *input_itf.TaskReportEntity

	for _, r := range s.reports {
		if r.TaskID == taskID {
			report = r
		}
	}

	return report
}

type fakeMQ struct{}

func (fakeMQ) Emit(uuid.UUID, output_itf.MQEvent, any) error { return nil }

func (fakeMQ) Subscribe(uuid.UUID, output_itf.MQEvent) (<-chan any, error) {
	return make(chan any), nil
}

func (fakeMQ) Unsubscribe(uuid.UUID, output_itf.MQEvent) error {
	return nil
}

func newManager(t *testing.T) (core_itf.SessionManager, *fakeStore) {
	t.Helper()

	return newManagerWith(t, fakeMQ{})
}

func newManagerWith(t *testing.T, mq output_itf.MessageQ) (core_itf.SessionManager, *fakeStore) {
	t.Helper()

	store := &fakeStore{}

	manager, err := InitV1(&input_itf.SessionConfig{
		HeartbeatTimeout:       30 * time.Minute,
		HeartbeatScanInterval:  time.Minute,
		AgentHeartbeatInterval: time.Minute,
	}, store, mq)
	if err != nil {
		t.Fatalf("init session manager: %v", err)
	}

	t.Cleanup(manager.Stop)

	return manager, store
}

func addTask(t *testing.T, manager core_itf.SessionManager, session uuid.UUID, name string, gated bool, deps ...uuid.UUID) uuid.UUID {
	t.Helper()

	taskID, err := manager.AddTask(session, &core_itf.AddTask{
		Name:                 name,
		ManualAcceptRequired: gated,
		DependsOn:            deps,
		AgentSpecs:           &core_itf.AgentRequest{Name: enums.Sonnet},
	})
	if err != nil {
		t.Fatalf("add task %s: %v", name, err)
	}

	return taskID
}

func statusOf(t *testing.T, manager core_itf.SessionManager, session, taskID uuid.UUID) enums.TaskStatus {
	t.Helper()

	status, err := manager.Status(session)
	if err != nil {
		t.Fatalf("session status: %v", err)
	}

	report, found := status.Tasks[taskID]
	if !found {
		t.Fatalf("task %v missing from session status", taskID)
	}

	return report.Status
}

func readyIDs(t *testing.T, manager core_itf.SessionManager, session uuid.UUID) map[uuid.UUID]bool {
	t.Helper()

	specs, err := manager.ReadyTasks(session)
	if err != nil {
		t.Fatalf("ready tasks: %v", err)
	}

	ids := map[uuid.UUID]bool{}
	for _, spec := range specs {
		ids[spec.TaskID] = true
	}

	return ids
}

func newSession(t *testing.T, manager core_itf.SessionManager) uuid.UUID {
	t.Helper()

	session, err := manager.NewSession(&core_itf.InitSession{WorkingDirPath: t.TempDir()})
	if err != nil {
		t.Fatalf("new session: %v", err)
	}

	return session
}

func reportDone(t *testing.T, manager core_itf.SessionManager, taskID, agentID uuid.UUID) {
	t.Helper()

	if err := manager.Assign(taskID, agentID); err != nil {
		t.Fatalf("assign: %v", err)
	}

	if err := manager.Report(agentID, enums.TaskCompleted, []*core_itf.HandoverDoc{{
		Outcome: "briefing",
	}}); err != nil {
		t.Fatalf("report: %v", err)
	}
}

func TestGatedTaskHoldsDownstreamUntilAccepted(t *testing.T) {
	manager, store := newManager(t)
	session := newSession(t, manager)

	gate := addTask(t, manager, session, "plan", true)
	downstream := addTask(t, manager, session, "implement", false, gate)

	agentID := uuid.New()
	reportDone(t, manager, gate, agentID)

	if got := statusOf(t, manager, session, gate); got != enums.TaskAwaitingAccept {
		t.Fatalf("gated task status = %s, want %s", got, enums.TaskAwaitingAccept)
	}

	if got := store.lastStatusFor(gate); got != enums.TaskAwaitingAccept {
		t.Fatalf("stored status = %s, want %s", got, enums.TaskAwaitingAccept)
	}

	if ready := readyIDs(t, manager, session); ready[downstream] {
		t.Fatal("downstream task became ready while the gate was open")
	}

	if err := manager.AnswerAcceptance(gate, true); err != nil {
		t.Fatalf("accept: %v", err)
	}

	if got := statusOf(t, manager, session, gate); got != enums.TaskCompleted {
		t.Fatalf("accepted task status = %s, want %s", got, enums.TaskCompleted)
	}

	if ready := readyIDs(t, manager, session); !ready[downstream] {
		t.Fatal("downstream task did not become ready after acceptance")
	}
}

func TestRejectingGatedTaskFailsItAndKeepsDownstreamBlocked(t *testing.T) {
	manager, _ := newManager(t)
	session := newSession(t, manager)

	gate := addTask(t, manager, session, "plan", true)
	downstream := addTask(t, manager, session, "implement", false, gate)

	reportDone(t, manager, gate, uuid.New())

	if err := manager.AnswerAcceptance(gate, false); err != nil {
		t.Fatalf("reject: %v", err)
	}

	if got := statusOf(t, manager, session, gate); got != enums.TaskFailed {
		t.Fatalf("rejected task status = %s, want %s", got, enums.TaskFailed)
	}

	if ready := readyIDs(t, manager, session); ready[downstream] {
		t.Fatal("downstream task became ready after a rejection")
	}
}

func TestUngatedTaskCompletesWithoutAGate(t *testing.T) {
	manager, _ := newManager(t)
	session := newSession(t, manager)

	task := addTask(t, manager, session, "implement", false)
	reportDone(t, manager, task, uuid.New())

	if got := statusOf(t, manager, session, task); got != enums.TaskCompleted {
		t.Fatalf("ungated task status = %s, want %s", got, enums.TaskCompleted)
	}
}

func TestGatedFailureIsNotHeld(t *testing.T) {
	manager, _ := newManager(t)
	session := newSession(t, manager)

	task := addTask(t, manager, session, "plan", true)
	agentID := uuid.New()

	if err := manager.Assign(task, agentID); err != nil {
		t.Fatalf("assign: %v", err)
	}

	if err := manager.Report(agentID, enums.TaskFailed, []*core_itf.HandoverDoc{{
		Outcome: "blocked",
	}}); err != nil {
		t.Fatalf("report: %v", err)
	}

	if got := statusOf(t, manager, session, task); got != enums.TaskFailed {
		t.Fatalf("failed gated task status = %s, want %s", got, enums.TaskFailed)
	}
}

func TestReportNamesTheHandoverAfterTheTaskNotTheAgent(t *testing.T) {
	manager, store := newManager(t)
	session := newSession(t, manager)

	task := addTask(t, manager, session, "wire the proxy", false)
	agentID := uuid.New()

	if err := manager.Assign(task, agentID); err != nil {
		t.Fatalf("assign: %v", err)
	}

	doc := &core_itf.HandoverDoc{Task: "some name the agent made up", Outcome: "done"}
	if err := manager.Report(agentID, enums.TaskCompleted, []*core_itf.HandoverDoc{doc}); err != nil {
		t.Fatalf("report: %v", err)
	}

	if doc.Task != "wire the proxy" {
		t.Fatalf("handover doc task = %q, want the task name %q", doc.Task, "wire the proxy")
	}

	status, err := manager.Status(session)
	if err != nil {
		t.Fatalf("session status: %v", err)
	}

	kept := status.Tasks[task].HandoverDocs
	if len(kept) != 1 {
		t.Fatalf("kept %d handover docs, want 1", len(kept))
	}

	if kept[0].Task != "wire the proxy" {
		t.Fatalf("kept handover doc task = %q, want the task name", kept[0].Task)
	}

	record := store.lastReportFor(task)
	if record == nil {
		t.Fatal("the report never reached the database")
	}

	if len(record.HandoverDocs) != 1 || record.HandoverDocs[0].Task != "wire the proxy" {
		t.Fatalf("stored handover docs = %+v, want one named after the task", record.HandoverDocs)
	}
}

func TestAnswerAcceptanceRejectsTaskThatIsNotAwaiting(t *testing.T) {
	manager, _ := newManager(t)
	session := newSession(t, manager)

	task := addTask(t, manager, session, "implement", false)

	if err := manager.AnswerAcceptance(task, true); err == nil {
		t.Fatal("accepting a task that is not awaiting acceptance should fail")
	}

	if err := manager.AnswerAcceptance(uuid.New(), true); err == nil {
		t.Fatal("accepting an unknown task should fail")
	}
}

func TestGatedCompletionDoesNotCountAsRetry(t *testing.T) {
	manager, _ := newManager(t)
	session := newSession(t, manager)

	task := addTask(t, manager, session, "plan", true)
	reportDone(t, manager, task, uuid.New())

	specs, err := manager.ReadyTasks(session)
	if err != nil {
		t.Fatalf("ready tasks: %v", err)
	}

	for _, spec := range specs {
		if spec.TaskID == task {
			t.Fatal("a gated task must not be schedulable while awaiting acceptance")
		}
	}

	if err := manager.AnswerAcceptance(task, true); err != nil {
		t.Fatalf("accept: %v", err)
	}

	status, err := manager.Status(session)
	if err != nil {
		t.Fatalf("session status: %v", err)
	}

	if status.Status != enums.SessionCompleted {
		t.Fatalf("session status = %s, want %s", status.Status, enums.SessionCompleted)
	}
}

func TestRewindToResetsEverythingDownstream(t *testing.T) {
	manager, _ := newManager(t)
	session := newSession(t, manager)

	a := addTask(t, manager, session, "a", false)
	b := addTask(t, manager, session, "b", false, a)
	c := addTask(t, manager, session, "c", false, b)
	d := addTask(t, manager, session, "d", false, c)

	for _, taskID := range []uuid.UUID{a, b, c, d} {
		reportDone(t, manager, taskID, uuid.New())
	}

	status, err := manager.Status(session)
	if err != nil {
		t.Fatalf("session status: %v", err)
	}

	if status.Status != enums.SessionCompleted {
		t.Fatalf("session status before rewind = %s, want %s", status.Status, enums.SessionCompleted)
	}

	if err := manager.RewindTo(b); err != nil {
		t.Fatalf("rewind: %v", err)
	}

	status, err = manager.Status(session)
	if err != nil {
		t.Fatalf("session status: %v", err)
	}

	if status.Status == enums.SessionCompleted {
		t.Fatal("session is still completed after a rewind")
	}

	for _, taskID := range []uuid.UUID{a, b} {
		report := status.Tasks[taskID]

		if report.Status != enums.TaskCompleted {
			t.Fatalf("task %v status = %s, want %s", taskID, report.Status, enums.TaskCompleted)
		}

		if len(report.HandoverDocs) == 0 {
			t.Fatalf("task %v lost its handover docs", taskID)
		}
	}

	for _, taskID := range []uuid.UUID{c, d} {
		report := status.Tasks[taskID]

		if report.Status != enums.TaskNotTaken {
			t.Fatalf("task %v status = %s, want %s", taskID, report.Status, enums.TaskNotTaken)
		}

		if len(report.HandoverDocs) != 0 {
			t.Fatalf("task %v kept %d handover docs after a rewind", taskID, len(report.HandoverDocs))
		}
	}

	if ready := readyIDs(t, manager, session); !ready[c] || ready[d] {
		t.Fatal("after a rewind only the first rewound task should be ready")
	}

	if err := manager.RewindTo(uuid.New()); err == nil {
		t.Fatal("rewinding to an unknown task should fail")
	}
}

func TestCancelClearsAnOpenGate(t *testing.T) {
	manager, _ := newManager(t)
	session := newSession(t, manager)

	task := addTask(t, manager, session, "plan", true)
	reportDone(t, manager, task, uuid.New())

	if _, err := manager.Cancel(session); err != nil {
		t.Fatalf("cancel: %v", err)
	}

	if got := statusOf(t, manager, session, task); got != enums.TaskCancelled {
		t.Fatalf("gated task status after cancel = %s, want %s", got, enums.TaskCancelled)
	}

	status, err := manager.Status(session)
	if err != nil {
		t.Fatalf("session status: %v", err)
	}

	if status.Status != enums.SessionCompleted {
		t.Fatalf("cancelled session did not drain, status = %s", status.Status)
	}
}

func TestCancelledSessionRunsAgainFromTheRewindPoint(t *testing.T) {
	manager, _ := newManager(t)
	session := newSession(t, manager)

	a := addTask(t, manager, session, "a", false)
	b := addTask(t, manager, session, "b", false, a)
	c := addTask(t, manager, session, "c", false, b)

	for _, taskID := range []uuid.UUID{a, b, c} {
		reportDone(t, manager, taskID, uuid.New())
	}

	if _, err := manager.Cancel(session); err != nil {
		t.Fatalf("cancel: %v", err)
	}

	if err := manager.RewindTo(b); err != nil {
		t.Fatalf("rewind: %v", err)
	}

	if ready := readyIDs(t, manager, session); !ready[c] {
		t.Fatal("a rewound task stayed unschedulable after the session was cancelled")
	}

	if err := manager.Assign(c, uuid.New()); err != nil {
		t.Fatalf("assign after a cancel and rewind: %v", err)
	}

	status, err := manager.Status(session)
	if err != nil {
		t.Fatalf("session status: %v", err)
	}

	if status.Status != enums.SessionProcessing {
		t.Fatalf("session status while the rewound task runs = %s, want %s", status.Status, enums.SessionProcessing)
	}
}

func TestCancelledSessionTakesARetriedTaskAgain(t *testing.T) {
	manager, _ := newManager(t)
	session := newSession(t, manager)

	task := addTask(t, manager, session, "implement", false)

	if _, err := manager.Cancel(session); err != nil {
		t.Fatalf("cancel: %v", err)
	}

	if err := manager.RetryTask(task); err != nil {
		t.Fatalf("retry: %v", err)
	}

	if ready := readyIDs(t, manager, session); !ready[task] {
		t.Fatal("a retried task stayed unschedulable after the session was cancelled")
	}

	if err := manager.Assign(task, uuid.New()); err != nil {
		t.Fatalf("assign after a cancel and retry: %v", err)
	}
}

func TestCancelledSessionTakesWorkAgainOnTheNextRun(t *testing.T) {
	manager, _ := newManager(t)
	session := newSession(t, manager)

	task, err := manager.AddTask(session, &core_itf.AddTask{
		Name:       "implement",
		AutoRetry:  true,
		AgentSpecs: &core_itf.AgentRequest{Name: enums.Sonnet},
	})
	if err != nil {
		t.Fatalf("add task: %v", err)
	}

	if err := manager.Assign(task, uuid.New()); err != nil {
		t.Fatalf("assign: %v", err)
	}

	if _, err := manager.Cancel(session); err != nil {
		t.Fatalf("cancel: %v", err)
	}

	if ready := readyIDs(t, manager, session); ready[task] {
		t.Fatal("a cancelled session still offers its tasks")
	}

	if err := manager.Assign(task, uuid.New()); err == nil {
		t.Fatal("a cancelled session accepted an assignment")
	}

	if _, err := manager.Execute(session); err != nil {
		t.Fatalf("execute after cancel: %v", err)
	}

	if ready := readyIDs(t, manager, session); !ready[task] {
		t.Fatal("running the session again did not reopen the pool")
	}

	if err := manager.Assign(task, uuid.New()); err != nil {
		t.Fatalf("assign after a cancel and a new run: %v", err)
	}

	status, err := manager.Status(session)
	if err != nil {
		t.Fatalf("session status: %v", err)
	}

	if status.Status != enums.SessionProcessing {
		t.Fatalf("session status while the resumed task runs = %s, want %s", status.Status, enums.SessionProcessing)
	}
}

func TestExecuteAgainReplacesTheProgressStream(t *testing.T) {
	manager, _ := newManagerWith(t, message_queue.InitV1())
	session := newSession(t, manager)

	retired, err := manager.Execute(session)
	if err != nil {
		t.Fatalf("first execute: %v", err)
	}

	current, err := manager.Execute(session)
	if err != nil {
		t.Fatalf("second execute: %v", err)
	}

	select {
	case event, open := <-retired:
		if open {
			t.Fatalf("the retired progress stream is still delivering %s", event.Event)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the retired progress stream was never closed")
	}

	task := addTask(t, manager, session, "implement", false)

	select {
	case event := <-current:
		if event.TaskID != task {
			t.Fatalf("progress reported task %v, want %v", event.TaskID, task)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the current progress stream received nothing")
	}
}
