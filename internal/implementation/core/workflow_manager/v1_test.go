package workflow_manager

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
	mu        sync.Mutex
	workflows []*input_itf.WorkflowEntity
	steps     []*input_itf.StepEntity
	reports   []*input_itf.StepResultEntity
}

func (s *fakeStore) SaveStepHistory(
	workflows []*input_itf.WorkflowEntity,
	steps []*input_itf.StepEntity,
	reports []*input_itf.StepResultEntity,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, workflow := range workflows {
		clone := *workflow
		s.workflows = append(s.workflows, &clone)
	}

	for _, step := range steps {
		clone := *step
		s.steps = append(s.steps, &clone)
	}

	for _, report := range reports {
		clone := *report
		s.reports = append(s.reports, &clone)
	}

	return nil
}

func (s *fakeStore) LoadStepHistory() ([]*input_itf.WorkflowSnapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	order := []uuid.UUID{}
	byWorkflow := map[uuid.UUID]*input_itf.WorkflowSnapshot{}

	for _, workflow := range s.workflows {
		snapshot, seen := byWorkflow[workflow.ID]
		if !seen {
			snapshot = &input_itf.WorkflowSnapshot{}
			byWorkflow[workflow.ID] = snapshot
			order = append(order, workflow.ID)
		}

		clone := *workflow
		snapshot.Workflow = &clone
	}

	kept := map[uuid.UUID]*input_itf.StepEntity{}
	ownerOfStep := map[uuid.UUID]uuid.UUID{}

	for _, step := range s.steps {
		snapshot, found := byWorkflow[step.WorkflowID]
		if !found {
			continue
		}

		clone := *step

		if previous, seen := kept[step.ID]; seen {
			*previous = clone
			continue
		}

		kept[step.ID] = &clone
		ownerOfStep[step.ID] = step.WorkflowID
		snapshot.Steps = append(snapshot.Steps, &clone)
	}

	for _, report := range s.reports {
		owner, found := ownerOfStep[report.StepID]
		if !found {
			continue
		}

		clone := *report
		byWorkflow[owner].Reports = append(byWorkflow[owner].Reports, &clone)
	}

	loaded := make([]*input_itf.WorkflowSnapshot, 0, len(order))
	for _, id := range order {
		loaded = append(loaded, byWorkflow[id])
	}

	return loaded, nil
}

func (s *fakeStore) lastStatusFor(stepID uuid.UUID) enums.StepStatus {
	s.mu.Lock()
	defer s.mu.Unlock()

	status := enums.StepStatus("")

	for _, step := range s.steps {
		if step.ID == stepID {
			status = step.Status
		}
	}

	return status
}

func (s *fakeStore) lastReportFor(stepID uuid.UUID) *input_itf.StepResultEntity {
	s.mu.Lock()
	defer s.mu.Unlock()

	var report *input_itf.StepResultEntity

	for _, r := range s.reports {
		if r.StepID == stepID {
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

func newManager(t *testing.T) (core_itf.WorkflowManager, *fakeStore) {
	t.Helper()

	return newManagerWith(t, fakeMQ{})
}

func newManagerWith(t *testing.T, mq output_itf.MessageQ) (core_itf.WorkflowManager, *fakeStore) {
	t.Helper()

	store := &fakeStore{}

	return managerOver(t, store, mq), store
}

func managerOver(t *testing.T, store *fakeStore, mq output_itf.MessageQ) core_itf.WorkflowManager {
	t.Helper()

	manager, err := InitV1(&input_itf.WorkflowConfig{
		HeartbeatTimeout:       30 * time.Minute,
		HeartbeatScanInterval:  time.Minute,
		AgentHeartbeatInterval: time.Minute,
	}, store, mq)
	if err != nil {
		t.Fatalf("init workflow manager: %v", err)
	}

	t.Cleanup(manager.Stop)

	return manager
}

func restoredManager(t *testing.T, store *fakeStore) core_itf.WorkflowManager {
	t.Helper()

	manager := managerOver(t, store, fakeMQ{})

	if err := manager.Restore(); err != nil {
		t.Fatalf("restore: %v", err)
	}

	return manager
}

func addStep(t *testing.T, manager core_itf.WorkflowManager, workflow uuid.UUID, name string, gated bool, deps ...uuid.UUID) uuid.UUID {
	t.Helper()

	stepID, err := manager.AddStep(workflow, &core_itf.AddStep{
		Name:           name,
		PauseForReview: gated,
		DependsOn:      deps,
		AgentSpecs:     &core_itf.AgentRequest{Name: enums.Sonnet},
	})
	if err != nil {
		t.Fatalf("add step %s: %v", name, err)
	}

	return stepID
}

func statusOf(t *testing.T, manager core_itf.WorkflowManager, workflow, stepID uuid.UUID) enums.StepStatus {
	t.Helper()

	status, err := manager.Status(workflow)
	if err != nil {
		t.Fatalf("workflow status: %v", err)
	}

	report, found := status.Steps[stepID]
	if !found {
		t.Fatalf("step %v missing from workflow status", stepID)
	}

	return report.Status
}

func readyIDs(t *testing.T, manager core_itf.WorkflowManager, workflow uuid.UUID) map[uuid.UUID]bool {
	t.Helper()

	specs, err := manager.ReadySteps(workflow)
	if err != nil {
		t.Fatalf("ready steps: %v", err)
	}

	ids := map[uuid.UUID]bool{}
	for _, spec := range specs {
		ids[spec.StepID] = true
	}

	return ids
}

func newWorkflow(t *testing.T, manager core_itf.WorkflowManager) uuid.UUID {
	t.Helper()

	workflow, err := manager.NewWorkflow(&core_itf.InitWorkflow{ProjectDirPath: t.TempDir()})
	if err != nil {
		t.Fatalf("new workflow: %v", err)
	}

	return workflow
}

func reportDone(t *testing.T, manager core_itf.WorkflowManager, stepID, agentID uuid.UUID) {
	t.Helper()

	if err := manager.Assign(stepID, agentID); err != nil {
		t.Fatalf("assign: %v", err)
	}

	if err := manager.Report(agentID, enums.StepCompleted, []*core_itf.Handoff{{
		Outcome: "briefing",
	}}); err != nil {
		t.Fatalf("report: %v", err)
	}
}

func TestGatedStepHoldsDownstreamUntilAccepted(t *testing.T) {
	manager, store := newManager(t)
	workflow := newWorkflow(t, manager)

	gate := addStep(t, manager, workflow, "plan", true)
	downstream := addStep(t, manager, workflow, "implement", false, gate)

	agentID := uuid.New()
	reportDone(t, manager, gate, agentID)

	if got := statusOf(t, manager, workflow, gate); got != enums.StepAwaitingReview {
		t.Fatalf("gated step status = %s, want %s", got, enums.StepAwaitingReview)
	}

	if got := store.lastStatusFor(gate); got != enums.StepAwaitingReview {
		t.Fatalf("stored status = %s, want %s", got, enums.StepAwaitingReview)
	}

	if ready := readyIDs(t, manager, workflow); ready[downstream] {
		t.Fatal("downstream step became ready while the gate was open")
	}

	if err := manager.AnswerReview(gate, true); err != nil {
		t.Fatalf("accept: %v", err)
	}

	if got := statusOf(t, manager, workflow, gate); got != enums.StepCompleted {
		t.Fatalf("accepted step status = %s, want %s", got, enums.StepCompleted)
	}

	if ready := readyIDs(t, manager, workflow); !ready[downstream] {
		t.Fatal("downstream step did not become ready after acceptance")
	}
}

func TestRejectingGatedStepFailsItAndKeepsDownstreamBlocked(t *testing.T) {
	manager, _ := newManager(t)
	workflow := newWorkflow(t, manager)

	gate := addStep(t, manager, workflow, "plan", true)
	downstream := addStep(t, manager, workflow, "implement", false, gate)

	reportDone(t, manager, gate, uuid.New())

	if err := manager.AnswerReview(gate, false); err != nil {
		t.Fatalf("reject: %v", err)
	}

	if got := statusOf(t, manager, workflow, gate); got != enums.StepFailed {
		t.Fatalf("rejected step status = %s, want %s", got, enums.StepFailed)
	}

	if ready := readyIDs(t, manager, workflow); ready[downstream] {
		t.Fatal("downstream step became ready after a rejection")
	}
}

func TestGatedFailureIsNotHeld(t *testing.T) {
	manager, _ := newManager(t)
	workflow := newWorkflow(t, manager)

	step := addStep(t, manager, workflow, "plan", true)
	agentID := uuid.New()

	if err := manager.Assign(step, agentID); err != nil {
		t.Fatalf("assign: %v", err)
	}

	if err := manager.Report(agentID, enums.StepFailed, []*core_itf.Handoff{{
		Outcome: "blocked",
	}}); err != nil {
		t.Fatalf("report: %v", err)
	}

	if got := statusOf(t, manager, workflow, step); got != enums.StepFailed {
		t.Fatalf("failed gated step status = %s, want %s", got, enums.StepFailed)
	}
}

func TestReportNamesTheHandoffAfterTheStepNotTheAgent(t *testing.T) {
	manager, store := newManager(t)
	workflow := newWorkflow(t, manager)

	step := addStep(t, manager, workflow, "wire the proxy", false)
	agentID := uuid.New()

	if err := manager.Assign(step, agentID); err != nil {
		t.Fatalf("assign: %v", err)
	}

	doc := &core_itf.Handoff{Step: "some name the agent made up", Outcome: "done"}
	if err := manager.Report(agentID, enums.StepCompleted, []*core_itf.Handoff{doc}); err != nil {
		t.Fatalf("report: %v", err)
	}

	if doc.Step != "wire the proxy" {
		t.Fatalf("handoff step = %q, want the step name %q", doc.Step, "wire the proxy")
	}

	status, err := manager.Status(workflow)
	if err != nil {
		t.Fatalf("workflow status: %v", err)
	}

	kept := status.Steps[step].Handoffs
	if len(kept) != 1 {
		t.Fatalf("kept %d handoffs, want 1", len(kept))
	}

	if kept[0].Step != "wire the proxy" {
		t.Fatalf("kept handoff step = %q, want the step name", kept[0].Step)
	}

	record := store.lastReportFor(step)
	if record == nil {
		t.Fatal("the report never reached the database")
	}

	if len(record.Handoffs) != 1 || record.Handoffs[0].Step != "wire the proxy" {
		t.Fatalf("stored handoffs = %+v, want one named after the step", record.Handoffs)
	}
}

func TestAnswerReviewRejectsStepThatIsNotAwaiting(t *testing.T) {
	manager, _ := newManager(t)
	workflow := newWorkflow(t, manager)

	step := addStep(t, manager, workflow, "implement", false)

	if err := manager.AnswerReview(step, true); err == nil {
		t.Fatal("accepting a step that is not awaiting acceptance should fail")
	}

	if err := manager.AnswerReview(uuid.New(), true); err == nil {
		t.Fatal("accepting an unknown step should fail")
	}
}

func TestGatedCompletionDoesNotCountAsRetry(t *testing.T) {
	manager, _ := newManager(t)
	workflow := newWorkflow(t, manager)

	step := addStep(t, manager, workflow, "plan", true)
	reportDone(t, manager, step, uuid.New())

	specs, err := manager.ReadySteps(workflow)
	if err != nil {
		t.Fatalf("ready steps: %v", err)
	}

	for _, spec := range specs {
		if spec.StepID == step {
			t.Fatal("a gated step must not be schedulable while awaiting acceptance")
		}
	}

	if err := manager.AnswerReview(step, true); err != nil {
		t.Fatalf("accept: %v", err)
	}

	status, err := manager.Status(workflow)
	if err != nil {
		t.Fatalf("workflow status: %v", err)
	}

	if status.Status != enums.WorkflowCompleted {
		t.Fatalf("workflow status = %s, want %s", status.Status, enums.WorkflowCompleted)
	}
}

func TestRewindToResetsEverythingDownstream(t *testing.T) {
	manager, _ := newManager(t)
	workflow := newWorkflow(t, manager)

	a := addStep(t, manager, workflow, "a", false)
	b := addStep(t, manager, workflow, "b", false, a)
	c := addStep(t, manager, workflow, "c", false, b)
	d := addStep(t, manager, workflow, "d", false, c)

	for _, stepID := range []uuid.UUID{a, b, c, d} {
		reportDone(t, manager, stepID, uuid.New())
	}

	status, err := manager.Status(workflow)
	if err != nil {
		t.Fatalf("workflow status: %v", err)
	}

	if status.Status != enums.WorkflowCompleted {
		t.Fatalf("workflow status before rewind = %s, want %s", status.Status, enums.WorkflowCompleted)
	}

	if err := manager.RewindTo(b); err != nil {
		t.Fatalf("rewind: %v", err)
	}

	status, err = manager.Status(workflow)
	if err != nil {
		t.Fatalf("workflow status: %v", err)
	}

	if status.Status == enums.WorkflowCompleted {
		t.Fatal("workflow is still completed after a rewind")
	}

	for _, stepID := range []uuid.UUID{a, b} {
		report := status.Steps[stepID]

		if report.Status != enums.StepCompleted {
			t.Fatalf("step %v status = %s, want %s", stepID, report.Status, enums.StepCompleted)
		}

		if len(report.Handoffs) == 0 {
			t.Fatalf("step %v lost its handoffs", stepID)
		}
	}

	for _, stepID := range []uuid.UUID{c, d} {
		report := status.Steps[stepID]

		if report.Status != enums.StepNotTaken {
			t.Fatalf("step %v status = %s, want %s", stepID, report.Status, enums.StepNotTaken)
		}

		if len(report.Handoffs) != 0 {
			t.Fatalf("step %v kept %d handoffs after a rewind", stepID, len(report.Handoffs))
		}
	}

	if ready := readyIDs(t, manager, workflow); !ready[c] || ready[d] {
		t.Fatal("after a rewind only the first rewound step should be ready")
	}

	if err := manager.RewindTo(uuid.New()); err == nil {
		t.Fatal("rewinding to an unknown step should fail")
	}
}

func TestCancelClearsAnOpenGate(t *testing.T) {
	manager, _ := newManager(t)
	workflow := newWorkflow(t, manager)

	step := addStep(t, manager, workflow, "plan", true)
	reportDone(t, manager, step, uuid.New())

	if _, err := manager.Cancel(workflow); err != nil {
		t.Fatalf("cancel: %v", err)
	}

	if got := statusOf(t, manager, workflow, step); got != enums.StepCancelled {
		t.Fatalf("gated step status after cancel = %s, want %s", got, enums.StepCancelled)
	}

	status, err := manager.Status(workflow)
	if err != nil {
		t.Fatalf("workflow status: %v", err)
	}

	if status.Status != enums.WorkflowCompleted {
		t.Fatalf("cancelled workflow did not drain, status = %s", status.Status)
	}
}

func TestCancelledWorkflowRunsAgainFromTheRewindPoint(t *testing.T) {
	manager, _ := newManager(t)
	workflow := newWorkflow(t, manager)

	a := addStep(t, manager, workflow, "a", false)
	b := addStep(t, manager, workflow, "b", false, a)
	c := addStep(t, manager, workflow, "c", false, b)

	for _, stepID := range []uuid.UUID{a, b, c} {
		reportDone(t, manager, stepID, uuid.New())
	}

	if _, err := manager.Cancel(workflow); err != nil {
		t.Fatalf("cancel: %v", err)
	}

	if err := manager.RewindTo(b); err != nil {
		t.Fatalf("rewind: %v", err)
	}

	if ready := readyIDs(t, manager, workflow); !ready[c] {
		t.Fatal("a rewound step stayed unschedulable after the workflow was cancelled")
	}

	if err := manager.Assign(c, uuid.New()); err != nil {
		t.Fatalf("assign after a cancel and rewind: %v", err)
	}

	status, err := manager.Status(workflow)
	if err != nil {
		t.Fatalf("workflow status: %v", err)
	}

	if status.Status != enums.WorkflowProcessing {
		t.Fatalf("workflow status while the rewound step runs = %s, want %s", status.Status, enums.WorkflowProcessing)
	}
}

func TestCancelledWorkflowTakesARetriedStepAgain(t *testing.T) {
	manager, _ := newManager(t)
	workflow := newWorkflow(t, manager)

	step := addStep(t, manager, workflow, "implement", false)

	if _, err := manager.Cancel(workflow); err != nil {
		t.Fatalf("cancel: %v", err)
	}

	if err := manager.RetryStep(step); err != nil {
		t.Fatalf("retry: %v", err)
	}

	if ready := readyIDs(t, manager, workflow); !ready[step] {
		t.Fatal("a retried step stayed unschedulable after the workflow was cancelled")
	}

	if err := manager.Assign(step, uuid.New()); err != nil {
		t.Fatalf("assign after a cancel and retry: %v", err)
	}
}

func TestCancelledWorkflowTakesWorkAgainOnTheNextRun(t *testing.T) {
	manager, _ := newManager(t)
	workflow := newWorkflow(t, manager)

	step, err := manager.AddStep(workflow, &core_itf.AddStep{
		Name:       "implement",
		AutoRetry:  true,
		AgentSpecs: &core_itf.AgentRequest{Name: enums.Sonnet},
	})
	if err != nil {
		t.Fatalf("add step: %v", err)
	}

	if err := manager.Assign(step, uuid.New()); err != nil {
		t.Fatalf("assign: %v", err)
	}

	if _, err := manager.Cancel(workflow); err != nil {
		t.Fatalf("cancel: %v", err)
	}

	if ready := readyIDs(t, manager, workflow); ready[step] {
		t.Fatal("a cancelled workflow still offers its steps")
	}

	if err := manager.Assign(step, uuid.New()); err == nil {
		t.Fatal("a cancelled workflow accepted an assignment")
	}

	if _, err := manager.Execute(workflow); err != nil {
		t.Fatalf("execute after cancel: %v", err)
	}

	if ready := readyIDs(t, manager, workflow); !ready[step] {
		t.Fatal("running the workflow again did not reopen the pool")
	}

	if err := manager.Assign(step, uuid.New()); err != nil {
		t.Fatalf("assign after a cancel and a new run: %v", err)
	}

	status, err := manager.Status(workflow)
	if err != nil {
		t.Fatalf("workflow status: %v", err)
	}

	if status.Status != enums.WorkflowProcessing {
		t.Fatalf("workflow status while the resumed step runs = %s, want %s", status.Status, enums.WorkflowProcessing)
	}
}

func TestPauseParksOnlyTheRunningStep(t *testing.T) {
	manager, store := newManager(t)
	workflow := newWorkflow(t, manager)

	done := addStep(t, manager, workflow, "done", false)
	gate := addStep(t, manager, workflow, "plan", true)
	running := addStep(t, manager, workflow, "implement", false)
	idle := addStep(t, manager, workflow, "document", false)

	reportDone(t, manager, done, uuid.New())
	reportDone(t, manager, gate, uuid.New())

	agentID := uuid.New()
	if err := manager.Assign(running, agentID); err != nil {
		t.Fatalf("assign: %v", err)
	}

	killed, err := manager.Pause(workflow)
	if err != nil {
		t.Fatalf("pause: %v", err)
	}

	if len(killed) != 1 || killed[0] != agentID {
		t.Fatalf("pause returned %v, want only the agent holding the running step", killed)
	}

	if got := statusOf(t, manager, workflow, running); got != enums.StepNotTaken {
		t.Fatalf("the running step after pause = %s, want %s", got, enums.StepNotTaken)
	}

	if got := store.lastStatusFor(running); got != enums.StepNotTaken {
		t.Fatalf("the stored running step after pause = %s, want %s", got, enums.StepNotTaken)
	}

	if got := statusOf(t, manager, workflow, gate); got != enums.StepAwaitingReview {
		t.Fatalf("the gated step after pause = %s, want %s", got, enums.StepAwaitingReview)
	}

	if got := statusOf(t, manager, workflow, done); got != enums.StepCompleted {
		t.Fatalf("the completed step after pause = %s, want %s", got, enums.StepCompleted)
	}

	if got := statusOf(t, manager, workflow, idle); got != enums.StepNotTaken {
		t.Fatalf("the untouched step after pause = %s, want %s", got, enums.StepNotTaken)
	}

	if ready := readyIDs(t, manager, workflow); len(ready) != 0 {
		t.Fatalf("a paused workflow still offers %d steps", len(ready))
	}
}

func TestPausedWorkflowTakesTheParkedStepAgainOnTheNextRun(t *testing.T) {
	manager, _ := newManager(t)
	workflow := newWorkflow(t, manager)

	running := addStep(t, manager, workflow, "implement", false)

	if err := manager.Assign(running, uuid.New()); err != nil {
		t.Fatalf("assign: %v", err)
	}

	if _, err := manager.Pause(workflow); err != nil {
		t.Fatalf("pause: %v", err)
	}

	if status := workflowStatusOf(t, manager, workflow); status.Status != enums.WorkflowPaused {
		t.Fatalf("workflow status while paused = %s, want %s", status.Status, enums.WorkflowPaused)
	}

	if err := manager.Assign(running, uuid.New()); err == nil {
		t.Fatal("a paused workflow accepted an assignment")
	}

	if _, err := manager.Execute(workflow); err != nil {
		t.Fatalf("execute after pause: %v", err)
	}

	if ready := readyIDs(t, manager, workflow); !ready[running] {
		t.Fatal("resuming the workflow did not offer the parked step again")
	}

	if status := workflowStatusOf(t, manager, workflow); status.Status != enums.WorkflowProcessing {
		t.Fatalf("workflow status after resume = %s, want %s", status.Status, enums.WorkflowProcessing)
	}

	if err := manager.Assign(running, uuid.New()); err != nil {
		t.Fatalf("assign after a pause and a new run: %v", err)
	}
}

func TestPauseKeepsTheTokensTheKilledAgentSpent(t *testing.T) {
	manager, live := trackedManager(t)
	workflow := newWorkflow(t, manager)
	stepID := addStep(t, manager, workflow, "implement", false)

	agentID := uuid.New()
	live.spend(agentID, 850)

	if err := manager.Assign(stepID, agentID); err != nil {
		t.Fatalf("assign: %v", err)
	}

	if _, err := manager.Pause(workflow); err != nil {
		t.Fatalf("pause: %v", err)
	}

	if got := tokensBilled(t, manager, workflow); got != 850 {
		t.Fatalf("tokens billed after pause = %d, want the 850 the killed agent spent", got)
	}
}

func TestRestoreBringsBackAnInterruptedWorkflowAsPaused(t *testing.T) {
	manager, store := newManager(t)
	live := &fakeLive{usage: map[uuid.UUID]input_itf.ContextUsage{}}
	manager.TrackLiveAgents(live)

	workflow := newWorkflow(t, manager)
	done := addStep(t, manager, workflow, "plan", false)
	running := addStep(t, manager, workflow, "implement", false)

	finished := uuid.New()
	live.spend(finished, 900)
	reportDone(t, manager, done, finished)

	if err := manager.Assign(running, uuid.New()); err != nil {
		t.Fatalf("assign: %v", err)
	}

	restored := restoredManager(t, store)
	status := workflowStatusOf(t, restored, workflow)

	if status.Status != enums.WorkflowPaused {
		t.Fatalf("restored workflow status = %s, want %s", status.Status, enums.WorkflowPaused)
	}

	if len(status.Steps) != 2 {
		t.Fatalf("restored workflow has %d steps, want 2", len(status.Steps))
	}

	if got := status.Steps[done].Status; got != enums.StepCompleted {
		t.Fatalf("restored completed step = %s, want %s", got, enums.StepCompleted)
	}

	docs := status.Steps[done].Handoffs
	if len(docs) != 1 || docs[0].Outcome != "briefing" || docs[0].Step != "plan" {
		t.Fatalf("restored handoffs = %+v, want the one the finished step wrote", docs)
	}

	if got := spentOn(t, status, done).Billed; got != 900 {
		t.Fatalf("restored step spent %d, want the 900 its attempt cost", got)
	}

	if got := status.Steps[running].Status; got != enums.StepNotTaken {
		t.Fatalf("restored running step = %s, want %s", got, enums.StepNotTaken)
	}

	if got := store.lastStatusFor(running); got != enums.StepNotTaken {
		t.Fatalf("the stored running step after restore = %s, want %s", got, enums.StepNotTaken)
	}

	if ready := readyIDs(t, restored, workflow); len(ready) != 0 {
		t.Fatalf("a restored workflow offers %d steps before it runs again", len(ready))
	}

	if _, err := restored.Execute(workflow); err != nil {
		t.Fatalf("execute after restore: %v", err)
	}

	if ready := readyIDs(t, restored, workflow); !ready[running] {
		t.Fatal("resuming a restored workflow did not offer the interrupted step again")
	}
}

func TestRestoreKeepsAWorkflowThatNeverRanRunnable(t *testing.T) {
	manager, store := newManager(t)
	workflow := newWorkflow(t, manager)
	step := addStep(t, manager, workflow, "implement", false)

	restored := restoredManager(t, store)

	if status := workflowStatusOf(t, restored, workflow); status.Status != enums.WorkflowInit {
		t.Fatalf("restored workflow status = %s, want %s", status.Status, enums.WorkflowInit)
	}

	if ready := readyIDs(t, restored, workflow); !ready[step] {
		t.Fatal("a restored workflow that never ran does not offer its work")
	}
}

func TestRestoreTwiceChangesNothing(t *testing.T) {
	manager, store := newManager(t)
	workflow := newWorkflow(t, manager)
	done := addStep(t, manager, workflow, "plan", false)
	running := addStep(t, manager, workflow, "implement", false)

	reportDone(t, manager, done, uuid.New())

	if err := manager.Assign(running, uuid.New()); err != nil {
		t.Fatalf("assign: %v", err)
	}

	restored := restoredManager(t, store)

	if err := restored.Restore(); err != nil {
		t.Fatalf("second restore: %v", err)
	}

	status := workflowStatusOf(t, restored, workflow)

	if status.Status != enums.WorkflowPaused {
		t.Fatalf("twice restored workflow status = %s, want %s", status.Status, enums.WorkflowPaused)
	}

	if len(status.Steps) != 2 {
		t.Fatalf("twice restored workflow has %d steps, want 2", len(status.Steps))
	}

	if got := len(status.Steps[done].Handoffs); got != 1 {
		t.Fatalf("twice restored step kept %d handoffs, want 1", got)
	}

	if got := status.Steps[running].Status; got != enums.StepNotTaken {
		t.Fatalf("twice restored running step = %s, want %s", got, enums.StepNotTaken)
	}
}

func TestExecuteAgainReplacesTheProgressStream(t *testing.T) {
	manager, _ := newManagerWith(t, message_queue.InitV1())
	workflow := newWorkflow(t, manager)

	retired, err := manager.Execute(workflow)
	if err != nil {
		t.Fatalf("first execute: %v", err)
	}

	current, err := manager.Execute(workflow)
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

	step := addStep(t, manager, workflow, "implement", false)

	select {
	case event := <-current:
		if event.StepID != step {
			t.Fatalf("progress reported step %v, want %v", event.StepID, step)
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

func trackedManager(t *testing.T) (core_itf.WorkflowManager, *fakeLive) {
	t.Helper()

	manager, _ := newManager(t)
	live := &fakeLive{usage: map[uuid.UUID]input_itf.ContextUsage{}}
	manager.TrackLiveAgents(live)

	return manager, live
}

func workflowStatusOf(t *testing.T, manager core_itf.WorkflowManager, workflow uuid.UUID) *core_itf.WorkflowStatus {
	t.Helper()

	status, err := manager.Status(workflow)
	if err != nil {
		t.Fatalf("workflow status: %v", err)
	}

	return status
}

func tokensBilled(t *testing.T, manager core_itf.WorkflowManager, workflow uuid.UUID) int {
	t.Helper()

	return workflowStatusOf(t, manager, workflow).TokensBilled
}

func spentOn(t *testing.T, status *core_itf.WorkflowStatus, stepID uuid.UUID) *input_itf.ContextUsage {
	t.Helper()

	report, found := status.Steps[stepID]
	if !found {
		t.Fatalf("step %v missing from workflow status", stepID)
	}

	if report.Spent == nil {
		t.Fatalf("step %v has no spend of its own", stepID)
	}

	return report.Spent
}

// The point of per-step accumulation: a node's bill and the workflow's bill can never
// tell different stories.
func assertTotalsSumTheSteps(t *testing.T, status *core_itf.WorkflowStatus) {
	t.Helper()

	billed, input, cached := 0, 0, 0

	for stepID := range status.Steps {
		spent := spentOn(t, status, stepID)

		billed += spent.Billed
		input += spent.Input
		cached += spent.Cached
	}

	if status.TokensBilled != billed || status.TokensInput != input || status.TokensCached != cached {
		t.Fatalf("workflow totals (billed %d, input %d, cached %d) do not sum the steps (%d, %d, %d)",
			status.TokensBilled, status.TokensInput, status.TokensCached, billed, input, cached)
	}
}

func TestRetriedStepKeepsEveryAttemptOnItsOwnBill(t *testing.T) {
	manager, live := trackedManager(t)
	workflow := newWorkflow(t, manager)

	retried := addStep(t, manager, workflow, "implement", false)
	other := addStep(t, manager, workflow, "document", false)

	first := uuid.New()
	live.spend(first, 1_000)

	if err := manager.Assign(retried, first); err != nil {
		t.Fatalf("assign the first attempt: %v", err)
	}

	if err := manager.Report(first, enums.StepFailed, []*core_itf.Handoff{{
		Outcome: "ran out of room",
	}}); err != nil {
		t.Fatalf("report the first attempt: %v", err)
	}

	second := uuid.New()
	live.spend(second, 400)

	if err := manager.Assign(retried, second); err != nil {
		t.Fatalf("assign the retry: %v", err)
	}

	if got := tokensBilled(t, manager, workflow); got != 1_400 {
		t.Fatalf("tokens billed while the retry runs = %d, want the failed attempt plus the live one (1400)", got)
	}

	if err := manager.Report(second, enums.StepCompleted, []*core_itf.Handoff{{
		Outcome: "done",
	}}); err != nil {
		t.Fatalf("report the retry: %v", err)
	}

	third := uuid.New()
	live.spend(third, 250)
	reportDone(t, manager, other, third)

	status := workflowStatusOf(t, manager, workflow)

	if got := spentOn(t, status, retried).Billed; got != 1_400 {
		t.Fatalf("the retried step spent %d, want both of its attempts (1400)", got)
	}

	if got := spentOn(t, status, other).Billed; got != 250 {
		t.Fatalf("the untouched step spent %d, want only its own 250", got)
	}

	if status.TokensBilled != 1_650 {
		t.Fatalf("workflow billed %d, want 1650", status.TokensBilled)
	}

	assertTotalsSumTheSteps(t, status)
}

func TestInputAndCachedTokensAreBilledPerStep(t *testing.T) {
	manager, live := trackedManager(t)
	workflow := newWorkflow(t, manager)

	first := addStep(t, manager, workflow, "implement", false)
	second := addStep(t, manager, workflow, "document", false)

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

	status := workflowStatusOf(t, manager, workflow)

	if spent := spentOn(t, status, first); spent.Input != 4_000 || spent.Cached != 120_000 {
		t.Fatalf("the first step spent %d input and %d cached, want 4000 and 120000", spent.Input, spent.Cached)
	}

	if spent := spentOn(t, status, second); spent.Input != 700 || spent.Cached != 9_000 {
		t.Fatalf("the second step spent %d input and %d cached, want 700 and 9000", spent.Input, spent.Cached)
	}

	if status.TokensInput != 4_700 || status.TokensCached != 129_000 {
		t.Fatalf("workflow input %d and cached %d, want 4700 and 129000", status.TokensInput, status.TokensCached)
	}

	assertTotalsSumTheSteps(t, status)
}

func TestCancelBillsTheKilledAgentToTheStepItWasWorking(t *testing.T) {
	manager, live := trackedManager(t)
	workflow := newWorkflow(t, manager)

	killed := addStep(t, manager, workflow, "implement", false)
	idle := addStep(t, manager, workflow, "document", false)

	agentID := uuid.New()
	live.spendUsage(agentID, input_itf.ContextUsage{
		Total: 200_000, Used: 40_000, Billed: 900, Input: 2_000, Cached: 50_000,
	})

	if err := manager.Assign(killed, agentID); err != nil {
		t.Fatalf("assign: %v", err)
	}

	if _, err := manager.Cancel(workflow); err != nil {
		t.Fatalf("cancel: %v", err)
	}

	status := workflowStatusOf(t, manager, workflow)

	spent := spentOn(t, status, killed)
	if spent.Billed != 900 || spent.Input != 2_000 || spent.Cached != 50_000 {
		t.Fatalf("the cancelled step spent %+v, want what the killed agent had used", spent)
	}

	if got := spentOn(t, status, idle).Billed; got != 0 {
		t.Fatalf("a step no agent ever took spent %d, want 0", got)
	}

	assertTotalsSumTheSteps(t, status)
}

func TestDroppingASilentAgentBillsItsStep(t *testing.T) {
	manager, live := trackedManager(t)
	workflow := newWorkflow(t, manager)

	dropped := addStep(t, manager, workflow, "implement", false)
	idle := addStep(t, manager, workflow, "document", false)

	agentID := uuid.New()
	live.spendUsage(agentID, input_itf.ContextUsage{
		Total: 200_000, Used: 20_000, Billed: 700, Input: 1_500, Cached: 30_000,
	})

	if err := manager.Assign(dropped, agentID); err != nil {
		t.Fatalf("assign: %v", err)
	}

	if err := manager.(*v1).dropStep(agentID, time.Now().Unix()+1); err != nil {
		t.Fatalf("drop the silent agent: %v", err)
	}

	if got := statusOf(t, manager, workflow, dropped); got != enums.StepCancelled {
		t.Fatalf("step after the drop = %s, want cancelled", got)
	}

	status := workflowStatusOf(t, manager, workflow)

	spent := spentOn(t, status, dropped)
	if spent.Billed != 700 || spent.Input != 1_500 || spent.Cached != 30_000 {
		t.Fatalf("the dropped step spent %+v, want what the silent agent had used", spent)
	}

	if got := spentOn(t, status, idle).Billed; got != 0 {
		t.Fatalf("a step no agent ever took spent %d, want 0", got)
	}

	assertTotalsSumTheSteps(t, status)
}

func TestARunningAgentIsBilledToItsStepExactlyOnce(t *testing.T) {
	manager, live := trackedManager(t)
	workflow := newWorkflow(t, manager)

	stepID := addStep(t, manager, workflow, "implement", false)

	agentID := uuid.New()
	live.spendUsage(agentID, input_itf.ContextUsage{
		Total: 200_000, Used: 15_000, Billed: 500, Input: 1_200, Cached: 20_000,
	})

	if err := manager.Assign(stepID, agentID); err != nil {
		t.Fatalf("assign: %v", err)
	}

	running := workflowStatusOf(t, manager, workflow)

	spent := spentOn(t, running, stepID)
	if spent.Billed != 500 || spent.Input != 1_200 || spent.Cached != 20_000 {
		t.Fatalf("the running step spent %+v, want the agent's live window", spent)
	}

	assertTotalsSumTheSteps(t, running)

	if err := manager.Report(agentID, enums.StepCompleted, []*core_itf.Handoff{{
		Outcome: "done",
	}}); err != nil {
		t.Fatalf("report: %v", err)
	}

	reported := workflowStatusOf(t, manager, workflow)

	after := spentOn(t, reported, stepID)
	if after.Billed != 500 || after.Input != 1_200 || after.Cached != 20_000 {
		t.Fatalf("the reported step spent %+v, want the same window counted once", after)
	}

	assertTotalsSumTheSteps(t, reported)
}

func TestEffortSurvivesFromAddStepToStatus(t *testing.T) {
	manager, _ := newManager(t)
	workflow := newWorkflow(t, manager)

	stepID, err := manager.AddStep(workflow, &core_itf.AddStep{
		Name:       "implement",
		Effort:     enums.EffortDeep,
		AgentSpecs: &core_itf.AgentRequest{Name: enums.Sonnet},
	})
	if err != nil {
		t.Fatalf("add step: %v", err)
	}

	status := workflowStatusOf(t, manager, workflow)

	if got := status.Steps[stepID].Effort; got != enums.EffortDeep {
		t.Fatalf("step level = %q, want %q", got, enums.EffortDeep)
	}
}

func TestWorkflowStatusReportsTheRunWindow(t *testing.T) {
	manager, _ := newManager(t)
	workflow := newWorkflow(t, manager)
	stepID := addStep(t, manager, workflow, "implement", false)

	before, err := manager.Status(workflow)
	if err != nil {
		t.Fatalf("workflow status before the run: %v", err)
	}

	if !before.StartedAt.IsZero() {
		t.Fatalf("a workflow with nothing assigned started at %v, want the zero time", before.StartedAt)
	}

	agentID := uuid.New()
	reportDone(t, manager, stepID, agentID)

	after, err := manager.Status(workflow)
	if err != nil {
		t.Fatalf("workflow status after the run: %v", err)
	}

	if after.StartedAt.IsZero() {
		t.Fatal("the finished workflow has no start time to measure from")
	}

	if after.CompletedAt.IsZero() {
		t.Fatal("the finished workflow has no completion time to measure to")
	}

	if after.CompletedAt.Before(after.StartedAt) {
		t.Fatalf("the run ended at %v, before it started at %v", after.CompletedAt, after.StartedAt)
	}
}
