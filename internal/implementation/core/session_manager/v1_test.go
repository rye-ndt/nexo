package session_manager

import (
	"errors"
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
	mu       sync.Mutex
	sessions []*input_itf.SessionEntity
	tasks    []*input_itf.TaskEntity
	reports  []*input_itf.TaskReportEntity
}

func (s *fakeStore) SaveTaskHistory(
	sessions []*input_itf.SessionEntity,
	tasks []*input_itf.TaskEntity,
	reports []*input_itf.TaskReportEntity,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, session := range sessions {
		clone := *session
		s.sessions = append(s.sessions, &clone)
	}

	for _, task := range tasks {
		clone := *task
		s.tasks = append(s.tasks, &clone)
	}

	for _, report := range reports {
		clone := *report
		s.reports = append(s.reports, &clone)
	}

	return nil
}

func (s *fakeStore) LoadTaskHistory() ([]*input_itf.SessionSnapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	order := []uuid.UUID{}
	bySession := map[uuid.UUID]*input_itf.SessionSnapshot{}

	for _, session := range s.sessions {
		snapshot, seen := bySession[session.ID]
		if !seen {
			snapshot = &input_itf.SessionSnapshot{}
			bySession[session.ID] = snapshot
			order = append(order, session.ID)
		}

		clone := *session
		snapshot.Session = &clone
	}

	kept := map[uuid.UUID]*input_itf.TaskEntity{}
	ownerOfTask := map[uuid.UUID]uuid.UUID{}

	for _, task := range s.tasks {
		snapshot, found := bySession[task.SessionID]
		if !found {
			continue
		}

		clone := *task

		if previous, seen := kept[task.ID]; seen {
			*previous = clone
			continue
		}

		kept[task.ID] = &clone
		ownerOfTask[task.ID] = task.SessionID
		snapshot.Tasks = append(snapshot.Tasks, &clone)
	}

	for _, report := range s.reports {
		owner, found := ownerOfTask[report.TaskID]
		if !found {
			continue
		}

		clone := *report
		bySession[owner].Reports = append(bySession[owner].Reports, &clone)
	}

	loaded := make([]*input_itf.SessionSnapshot, 0, len(order))
	for _, id := range order {
		loaded = append(loaded, bySession[id])
	}

	return loaded, nil
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

	return managerOver(t, store, mq), store
}

func managerOver(t *testing.T, store *fakeStore, mq output_itf.MessageQ) core_itf.SessionManager {
	t.Helper()

	manager, err := InitV1(&input_itf.SessionConfig{
		HeartbeatTimeout:       30 * time.Minute,
		HeartbeatScanInterval:  time.Minute,
		AgentHeartbeatInterval: time.Minute,
	}, store, mq)
	if err != nil {
		t.Fatalf("init session manager: %v", err)
	}

	t.Cleanup(manager.Stop)

	return manager
}

func restoredManager(t *testing.T, store *fakeStore) core_itf.SessionManager {
	t.Helper()

	manager := managerOver(t, store, fakeMQ{})

	if err := manager.Restore(); err != nil {
		t.Fatalf("restore: %v", err)
	}

	return manager
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

func TestPauseParksOnlyTheRunningTask(t *testing.T) {
	manager, store := newManager(t)
	session := newSession(t, manager)

	done := addTask(t, manager, session, "done", false)
	gate := addTask(t, manager, session, "plan", true)
	running := addTask(t, manager, session, "implement", false)
	idle := addTask(t, manager, session, "document", false)

	reportDone(t, manager, done, uuid.New())
	reportDone(t, manager, gate, uuid.New())

	agentID := uuid.New()
	if err := manager.Assign(running, agentID); err != nil {
		t.Fatalf("assign: %v", err)
	}

	killed, err := manager.Pause(session)
	if err != nil {
		t.Fatalf("pause: %v", err)
	}

	if len(killed) != 1 || killed[0] != agentID {
		t.Fatalf("pause returned %v, want only the agent holding the running task", killed)
	}

	if got := statusOf(t, manager, session, running); got != enums.TaskNotTaken {
		t.Fatalf("the running task after pause = %s, want %s", got, enums.TaskNotTaken)
	}

	if got := store.lastStatusFor(running); got != enums.TaskNotTaken {
		t.Fatalf("the stored running task after pause = %s, want %s", got, enums.TaskNotTaken)
	}

	if got := statusOf(t, manager, session, gate); got != enums.TaskAwaitingAccept {
		t.Fatalf("the gated task after pause = %s, want %s", got, enums.TaskAwaitingAccept)
	}

	if got := statusOf(t, manager, session, done); got != enums.TaskCompleted {
		t.Fatalf("the completed task after pause = %s, want %s", got, enums.TaskCompleted)
	}

	if got := statusOf(t, manager, session, idle); got != enums.TaskNotTaken {
		t.Fatalf("the untouched task after pause = %s, want %s", got, enums.TaskNotTaken)
	}

	if ready := readyIDs(t, manager, session); len(ready) != 0 {
		t.Fatalf("a paused session still offers %d tasks", len(ready))
	}
}

func TestPausedSessionTakesTheParkedTaskAgainOnTheNextRun(t *testing.T) {
	manager, _ := newManager(t)
	session := newSession(t, manager)

	running := addTask(t, manager, session, "implement", false)

	if err := manager.Assign(running, uuid.New()); err != nil {
		t.Fatalf("assign: %v", err)
	}

	if _, err := manager.Pause(session); err != nil {
		t.Fatalf("pause: %v", err)
	}

	if status := sessionStatusOf(t, manager, session); status.Status != enums.SessionPaused {
		t.Fatalf("session status while paused = %s, want %s", status.Status, enums.SessionPaused)
	}

	if _, err := manager.Execute(session); err != nil {
		t.Fatalf("execute after pause: %v", err)
	}

	if ready := readyIDs(t, manager, session); !ready[running] {
		t.Fatal("resuming the session did not offer the parked task again")
	}

	if status := sessionStatusOf(t, manager, session); status.Status != enums.SessionProcessing {
		t.Fatalf("session status after resume = %s, want %s", status.Status, enums.SessionProcessing)
	}

	if err := manager.Assign(running, uuid.New()); err != nil {
		t.Fatalf("assign after a pause and a new run: %v", err)
	}
}

func TestPausedSessionRefusesAnAssignment(t *testing.T) {
	manager, _ := newManager(t)
	session := newSession(t, manager)

	task := addTask(t, manager, session, "implement", false)

	if err := manager.Assign(task, uuid.New()); err != nil {
		t.Fatalf("assign: %v", err)
	}

	if _, err := manager.Pause(session); err != nil {
		t.Fatalf("pause: %v", err)
	}

	if err := manager.Assign(task, uuid.New()); err == nil {
		t.Fatal("a paused session accepted an assignment")
	}
}

func TestPauseKeepsTheTokensTheKilledAgentSpent(t *testing.T) {
	manager, live := trackedManager(t)
	session := newSession(t, manager)
	taskID := addTask(t, manager, session, "implement", false)

	agentID := uuid.New()
	live.spend(agentID, 850)

	if err := manager.Assign(taskID, agentID); err != nil {
		t.Fatalf("assign: %v", err)
	}

	if _, err := manager.Pause(session); err != nil {
		t.Fatalf("pause: %v", err)
	}

	if got := tokensBilled(t, manager, session); got != 850 {
		t.Fatalf("tokens billed after pause = %d, want the 850 the killed agent spent", got)
	}
}

func TestRestoreBringsBackAnInterruptedSessionAsPaused(t *testing.T) {
	manager, store := newManager(t)
	live := &fakeLive{usage: map[uuid.UUID]input_itf.ContextUsage{}}
	manager.TrackLiveAgents(live)

	session := newSession(t, manager)
	done := addTask(t, manager, session, "plan", false)
	running := addTask(t, manager, session, "implement", false)

	finished := uuid.New()
	live.spend(finished, 900)
	reportDone(t, manager, done, finished)

	if err := manager.Assign(running, uuid.New()); err != nil {
		t.Fatalf("assign: %v", err)
	}

	restored := restoredManager(t, store)
	status := sessionStatusOf(t, restored, session)

	if status.Status != enums.SessionPaused {
		t.Fatalf("restored session status = %s, want %s", status.Status, enums.SessionPaused)
	}

	if len(status.Tasks) != 2 {
		t.Fatalf("restored session has %d tasks, want 2", len(status.Tasks))
	}

	if got := status.Tasks[done].Status; got != enums.TaskCompleted {
		t.Fatalf("restored completed task = %s, want %s", got, enums.TaskCompleted)
	}

	docs := status.Tasks[done].HandoverDocs
	if len(docs) != 1 || docs[0].Outcome != "briefing" || docs[0].Task != "plan" {
		t.Fatalf("restored handover docs = %+v, want the one the finished task wrote", docs)
	}

	if got := spentOn(t, status, done).Billed; got != 900 {
		t.Fatalf("restored task spent %d, want the 900 its attempt cost", got)
	}

	if got := status.Tasks[running].Status; got != enums.TaskNotTaken {
		t.Fatalf("restored running task = %s, want %s", got, enums.TaskNotTaken)
	}

	if got := store.lastStatusFor(running); got != enums.TaskNotTaken {
		t.Fatalf("the stored running task after restore = %s, want %s", got, enums.TaskNotTaken)
	}

	if ready := readyIDs(t, restored, session); len(ready) != 0 {
		t.Fatalf("a restored session offers %d tasks before it runs again", len(ready))
	}

	if _, err := restored.Execute(session); err != nil {
		t.Fatalf("execute after restore: %v", err)
	}

	if ready := readyIDs(t, restored, session); !ready[running] {
		t.Fatal("resuming a restored session did not offer the interrupted task again")
	}
}

func TestRestoreKeepsASessionThatNeverRanRunnable(t *testing.T) {
	manager, store := newManager(t)
	session := newSession(t, manager)
	task := addTask(t, manager, session, "implement", false)

	restored := restoredManager(t, store)

	if status := sessionStatusOf(t, restored, session); status.Status != enums.SessionInit {
		t.Fatalf("restored session status = %s, want %s", status.Status, enums.SessionInit)
	}

	if ready := readyIDs(t, restored, session); !ready[task] {
		t.Fatal("a restored session that never ran does not offer its work")
	}
}

func TestRestoreTwiceChangesNothing(t *testing.T) {
	manager, store := newManager(t)
	session := newSession(t, manager)
	done := addTask(t, manager, session, "plan", false)
	running := addTask(t, manager, session, "implement", false)

	reportDone(t, manager, done, uuid.New())

	if err := manager.Assign(running, uuid.New()); err != nil {
		t.Fatalf("assign: %v", err)
	}

	restored := restoredManager(t, store)

	if err := restored.Restore(); err != nil {
		t.Fatalf("second restore: %v", err)
	}

	status := sessionStatusOf(t, restored, session)

	if status.Status != enums.SessionPaused {
		t.Fatalf("twice restored session status = %s, want %s", status.Status, enums.SessionPaused)
	}

	if len(status.Tasks) != 2 {
		t.Fatalf("twice restored session has %d tasks, want 2", len(status.Tasks))
	}

	if got := len(status.Tasks[done].HandoverDocs); got != 1 {
		t.Fatalf("twice restored task kept %d handover docs, want 1", got)
	}

	if got := status.Tasks[running].Status; got != enums.TaskNotTaken {
		t.Fatalf("twice restored running task = %s, want %s", got, enums.TaskNotTaken)
	}
}

func TestRestoredSessionFeedsTheHandoverDocDownstream(t *testing.T) {
	manager, store := newManager(t)
	session := newSession(t, manager)

	upstream := addTask(t, manager, session, "plan", false)
	downstream := addTask(t, manager, session, "implement", false, upstream)

	reportDone(t, manager, upstream, uuid.New())

	restored := restoredManager(t, store)
	status := sessionStatusOf(t, restored, session)

	docs := status.Tasks[upstream].HandoverDocs
	if len(docs) != 1 || docs[0].Outcome != "briefing" {
		t.Fatalf("the completed dependency handed over %+v, want its briefing", docs)
	}

	if _, err := restored.Execute(session); err != nil {
		t.Fatalf("execute after restore: %v", err)
	}

	if ready := readyIDs(t, restored, session); !ready[downstream] {
		t.Fatal("the downstream task of a restored dependency never became ready")
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

type fakeLive struct {
	mu    sync.Mutex
	usage map[uuid.UUID]input_itf.ContextUsage
}

func (l *fakeLive) spend(agentID uuid.UUID, billed int) {
	l.spendUsage(agentID, input_itf.ContextUsage{Total: 200_000, Used: billed, Billed: billed})
}

func (l *fakeLive) spendUsage(agentID uuid.UUID, usage input_itf.ContextUsage) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.usage[agentID] = usage
}

func (l *fakeLive) ContextUsage(agentID uuid.UUID) (*input_itf.ContextUsage, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	usage, found := l.usage[agentID]
	if !found {
		return nil, errors.New("no such agent")
	}

	return &usage, nil
}

func (l *fakeLive) Activity(uuid.UUID) ([]input_itf.Activity, error) {
	return nil, nil
}

func trackedManager(t *testing.T) (core_itf.SessionManager, *fakeLive) {
	t.Helper()

	manager, _ := newManager(t)
	live := &fakeLive{usage: map[uuid.UUID]input_itf.ContextUsage{}}
	manager.TrackLiveAgents(live)

	return manager, live
}

func sessionStatusOf(t *testing.T, manager core_itf.SessionManager, session uuid.UUID) *core_itf.SessionStatus {
	t.Helper()

	status, err := manager.Status(session)
	if err != nil {
		t.Fatalf("session status: %v", err)
	}

	return status
}

func tokensBilled(t *testing.T, manager core_itf.SessionManager, session uuid.UUID) int {
	t.Helper()

	return sessionStatusOf(t, manager, session).TokensBilled
}

func spentOn(t *testing.T, status *core_itf.SessionStatus, taskID uuid.UUID) *input_itf.ContextUsage {
	t.Helper()

	report, found := status.Tasks[taskID]
	if !found {
		t.Fatalf("task %v missing from session status", taskID)
	}

	if report.Spent == nil {
		t.Fatalf("task %v has no spend of its own", taskID)
	}

	return report.Spent
}

// The point of per-task accumulation: a node's bill and the session's bill can never
// tell different stories.
func assertTotalsSumTheTasks(t *testing.T, status *core_itf.SessionStatus) {
	t.Helper()

	billed, input, cached := 0, 0, 0

	for taskID := range status.Tasks {
		spent := spentOn(t, status, taskID)

		billed += spent.Billed
		input += spent.Input
		cached += spent.Cached
	}

	if status.TokensBilled != billed || status.TokensInput != input || status.TokensCached != cached {
		t.Fatalf("session totals (billed %d, input %d, cached %d) do not sum the tasks (%d, %d, %d)",
			status.TokensBilled, status.TokensInput, status.TokensCached, billed, input, cached)
	}
}

func TestSessionTokensKeepTheAttemptsThatAlreadyEnded(t *testing.T) {
	manager, live := trackedManager(t)
	session := newSession(t, manager)
	taskID := addTask(t, manager, session, "implement", false)

	first := uuid.New()
	live.spend(first, 1_000)

	if err := manager.Assign(taskID, first); err != nil {
		t.Fatalf("assign the first attempt: %v", err)
	}

	if err := manager.Report(first, enums.TaskFailed, []*core_itf.HandoverDoc{{
		Outcome: "ran out of room",
	}}); err != nil {
		t.Fatalf("report the first attempt: %v", err)
	}

	second := uuid.New()
	live.spend(second, 400)

	if err := manager.Assign(taskID, second); err != nil {
		t.Fatalf("assign the retry: %v", err)
	}

	if got := tokensBilled(t, manager, session); got != 1_400 {
		t.Fatalf("tokens billed while the retry runs = %d, want the failed attempt plus the live one (1400)", got)
	}

	if err := manager.Report(second, enums.TaskCompleted, []*core_itf.HandoverDoc{{
		Outcome: "done",
	}}); err != nil {
		t.Fatalf("report the retry: %v", err)
	}

	if got := tokensBilled(t, manager, session); got != 1_400 {
		t.Fatalf("tokens billed after the retry reported = %d, want 1400", got)
	}
}

func TestCancelKeepsTheTokensTheKilledAgentsSpent(t *testing.T) {
	manager, live := trackedManager(t)
	session := newSession(t, manager)
	taskID := addTask(t, manager, session, "implement", false)

	agentID := uuid.New()
	live.spend(agentID, 900)

	if err := manager.Assign(taskID, agentID); err != nil {
		t.Fatalf("assign: %v", err)
	}

	if _, err := manager.Cancel(session); err != nil {
		t.Fatalf("cancel: %v", err)
	}

	if got := tokensBilled(t, manager, session); got != 900 {
		t.Fatalf("tokens billed after cancel = %d, want the 900 the killed agent spent", got)
	}
}

func TestDroppingASilentAgentKeepsTheTokensItSpent(t *testing.T) {
	manager, live := trackedManager(t)
	session := newSession(t, manager)
	taskID := addTask(t, manager, session, "implement", false)

	agentID := uuid.New()
	live.spend(agentID, 700)

	if err := manager.Assign(taskID, agentID); err != nil {
		t.Fatalf("assign: %v", err)
	}

	if err := manager.(*v1).dropTask(agentID, time.Now().Unix()+1); err != nil {
		t.Fatalf("drop the silent agent: %v", err)
	}

	if got := statusOf(t, manager, session, taskID); got != enums.TaskCancelled {
		t.Fatalf("task after the drop = %s, want cancelled", got)
	}

	if got := tokensBilled(t, manager, session); got != 700 {
		t.Fatalf("tokens billed after the drop = %d, want the 700 the silent agent spent", got)
	}
}

func TestRetriedTaskKeepsEveryAttemptOnItsOwnBill(t *testing.T) {
	manager, live := trackedManager(t)
	session := newSession(t, manager)

	retried := addTask(t, manager, session, "implement", false)
	other := addTask(t, manager, session, "document", false)

	first := uuid.New()
	live.spend(first, 1_000)

	if err := manager.Assign(retried, first); err != nil {
		t.Fatalf("assign the first attempt: %v", err)
	}

	if err := manager.Report(first, enums.TaskFailed, []*core_itf.HandoverDoc{{
		Outcome: "ran out of room",
	}}); err != nil {
		t.Fatalf("report the first attempt: %v", err)
	}

	second := uuid.New()
	live.spend(second, 400)

	if err := manager.Assign(retried, second); err != nil {
		t.Fatalf("assign the retry: %v", err)
	}

	if err := manager.Report(second, enums.TaskCompleted, []*core_itf.HandoverDoc{{
		Outcome: "done",
	}}); err != nil {
		t.Fatalf("report the retry: %v", err)
	}

	third := uuid.New()
	live.spend(third, 250)
	reportDone(t, manager, other, third)

	status := sessionStatusOf(t, manager, session)

	if got := spentOn(t, status, retried).Billed; got != 1_400 {
		t.Fatalf("the retried task spent %d, want both of its attempts (1400)", got)
	}

	if got := spentOn(t, status, other).Billed; got != 250 {
		t.Fatalf("the untouched task spent %d, want only its own 250", got)
	}

	if status.TokensBilled != 1_650 {
		t.Fatalf("session billed %d, want 1650", status.TokensBilled)
	}

	assertTotalsSumTheTasks(t, status)
}

func TestInputAndCachedTokensAreBilledPerTask(t *testing.T) {
	manager, live := trackedManager(t)
	session := newSession(t, manager)

	first := addTask(t, manager, session, "implement", false)
	second := addTask(t, manager, session, "document", false)

	firstAgent := uuid.New()
	live.spendUsage(firstAgent, input_itf.ContextUsage{
		Total: 200_000, Used: 30_000, Billed: 900, Input: 4_000, Cached: 120_000,
	})

	secondAgent := uuid.New()
	live.spendUsage(secondAgent, input_itf.ContextUsage{
		Total: 200_000, Used: 10_000, Billed: 100, Input: 700, Cached: 9_000,
	})

	reportDone(t, manager, first, firstAgent)
	reportDone(t, manager, second, secondAgent)

	status := sessionStatusOf(t, manager, session)

	if spent := spentOn(t, status, first); spent.Input != 4_000 || spent.Cached != 120_000 {
		t.Fatalf("the first task spent %d input and %d cached, want 4000 and 120000", spent.Input, spent.Cached)
	}

	if spent := spentOn(t, status, second); spent.Input != 700 || spent.Cached != 9_000 {
		t.Fatalf("the second task spent %d input and %d cached, want 700 and 9000", spent.Input, spent.Cached)
	}

	if status.TokensInput != 4_700 || status.TokensCached != 129_000 {
		t.Fatalf("session input %d and cached %d, want 4700 and 129000", status.TokensInput, status.TokensCached)
	}

	assertTotalsSumTheTasks(t, status)
}

func TestCancelBillsTheKilledAgentToTheTaskItWasWorking(t *testing.T) {
	manager, live := trackedManager(t)
	session := newSession(t, manager)

	killed := addTask(t, manager, session, "implement", false)
	idle := addTask(t, manager, session, "document", false)

	agentID := uuid.New()
	live.spendUsage(agentID, input_itf.ContextUsage{
		Total: 200_000, Used: 40_000, Billed: 900, Input: 2_000, Cached: 50_000,
	})

	if err := manager.Assign(killed, agentID); err != nil {
		t.Fatalf("assign: %v", err)
	}

	if _, err := manager.Cancel(session); err != nil {
		t.Fatalf("cancel: %v", err)
	}

	status := sessionStatusOf(t, manager, session)

	spent := spentOn(t, status, killed)
	if spent.Billed != 900 || spent.Input != 2_000 || spent.Cached != 50_000 {
		t.Fatalf("the cancelled task spent %+v, want what the killed agent had used", spent)
	}

	if got := spentOn(t, status, idle).Billed; got != 0 {
		t.Fatalf("a task no agent ever took spent %d, want 0", got)
	}

	assertTotalsSumTheTasks(t, status)
}

func TestDroppingASilentAgentBillsItsTask(t *testing.T) {
	manager, live := trackedManager(t)
	session := newSession(t, manager)

	dropped := addTask(t, manager, session, "implement", false)
	idle := addTask(t, manager, session, "document", false)

	agentID := uuid.New()
	live.spendUsage(agentID, input_itf.ContextUsage{
		Total: 200_000, Used: 20_000, Billed: 700, Input: 1_500, Cached: 30_000,
	})

	if err := manager.Assign(dropped, agentID); err != nil {
		t.Fatalf("assign: %v", err)
	}

	if err := manager.(*v1).dropTask(agentID, time.Now().Unix()+1); err != nil {
		t.Fatalf("drop the silent agent: %v", err)
	}

	status := sessionStatusOf(t, manager, session)

	spent := spentOn(t, status, dropped)
	if spent.Billed != 700 || spent.Input != 1_500 || spent.Cached != 30_000 {
		t.Fatalf("the dropped task spent %+v, want what the silent agent had used", spent)
	}

	if got := spentOn(t, status, idle).Billed; got != 0 {
		t.Fatalf("a task no agent ever took spent %d, want 0", got)
	}

	assertTotalsSumTheTasks(t, status)
}

func TestARunningAgentIsBilledToItsTaskExactlyOnce(t *testing.T) {
	manager, live := trackedManager(t)
	session := newSession(t, manager)

	taskID := addTask(t, manager, session, "implement", false)

	agentID := uuid.New()
	live.spendUsage(agentID, input_itf.ContextUsage{
		Total: 200_000, Used: 15_000, Billed: 500, Input: 1_200, Cached: 20_000,
	})

	if err := manager.Assign(taskID, agentID); err != nil {
		t.Fatalf("assign: %v", err)
	}

	running := sessionStatusOf(t, manager, session)

	spent := spentOn(t, running, taskID)
	if spent.Billed != 500 || spent.Input != 1_200 || spent.Cached != 20_000 {
		t.Fatalf("the running task spent %+v, want the agent's live window", spent)
	}

	assertTotalsSumTheTasks(t, running)

	if err := manager.Report(agentID, enums.TaskCompleted, []*core_itf.HandoverDoc{{
		Outcome: "done",
	}}); err != nil {
		t.Fatalf("report: %v", err)
	}

	reported := sessionStatusOf(t, manager, session)

	after := spentOn(t, reported, taskID)
	if after.Billed != 500 || after.Input != 1_200 || after.Cached != 20_000 {
		t.Fatalf("the reported task spent %+v, want the same window counted once", after)
	}

	assertTotalsSumTheTasks(t, reported)
}

func TestTaskLevelSurvivesFromAddTaskToStatus(t *testing.T) {
	manager, _ := newManager(t)
	session := newSession(t, manager)

	taskID, err := manager.AddTask(session, &core_itf.AddTask{
		Name:       "implement",
		TaskLevel:  enums.HeavyTask,
		AgentSpecs: &core_itf.AgentRequest{Name: enums.Sonnet},
	})
	if err != nil {
		t.Fatalf("add task: %v", err)
	}

	status := sessionStatusOf(t, manager, session)

	if got := status.Tasks[taskID].TaskLevel; got != enums.HeavyTask {
		t.Fatalf("task level = %q, want %q", got, enums.HeavyTask)
	}
}

func TestSessionStatusReportsTheRunWindow(t *testing.T) {
	manager, _ := newManager(t)
	session := newSession(t, manager)
	taskID := addTask(t, manager, session, "implement", false)

	before, err := manager.Status(session)
	if err != nil {
		t.Fatalf("session status before the run: %v", err)
	}

	if !before.StartedAt.IsZero() {
		t.Fatalf("a session with nothing assigned started at %v, want the zero time", before.StartedAt)
	}

	agentID := uuid.New()
	reportDone(t, manager, taskID, agentID)

	after, err := manager.Status(session)
	if err != nil {
		t.Fatalf("session status after the run: %v", err)
	}

	if after.StartedAt.IsZero() {
		t.Fatal("the finished session has no start time to measure from")
	}

	if after.CompletedAt.IsZero() {
		t.Fatal("the finished session has no completion time to measure to")
	}

	if after.CompletedAt.Before(after.StartedAt) {
		t.Fatalf("the run ended at %v, before it started at %v", after.CompletedAt, after.StartedAt)
	}
}
