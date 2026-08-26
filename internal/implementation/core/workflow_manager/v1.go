package workflow_manager

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"hexago/internal/helpers"
	"hexago/internal/helpers/custom_error"
	"hexago/internal/helpers/enums"
	core_itf "hexago/internal/interface/core"
	input_itf "hexago/internal/interface/input"

	"github.com/google/uuid"
)

const progressBufferSize = 32

const maxAutoRetry = 3

// errNoChange makes editStep return a nil change rather than an error.
var errNoChange = errors.New("step needs no change")

type AgentHandle struct {
	AgentID       uuid.UUID
	StepID        uuid.UUID
	AssignedAt    time.Time
	LastHeartBeat int64
}

// spentByStep carries only the three cumulative counters of ContextUsage. Total and
// Used are readings of one attempt's window, so summing them over attempts says
// nothing.
type workflowMetadata struct {
	info            *input_itf.WorkflowEntity
	stepIDToStep    map[uuid.UUID]*input_itf.StepEntity
	agentIDToHandle map[uuid.UUID]*AgentHandle
	stepIDToReport  map[uuid.UUID]*core_itf.StepResult
	spentByStep     map[uuid.UUID]input_itf.ContextUsage
	halt            enums.WorkflowHalt
	run             chan *core_itf.WorkflowProgress
}

type v1 struct {
	locker    sync.Mutex
	cfg       *input_itf.WorkflowConfig
	db        input_itf.StepStorage
	live      core_itf.LiveAgentReader
	workflows map[uuid.UUID]*workflowMetadata
	stop      chan struct{}
}

func InitV1(cfg *input_itf.WorkflowConfig, db input_itf.StepStorage) (core_itf.WorkflowManager, error) {
	if err := helpers.ValidateStruct(cfg); err != nil {
		return nil, custom_error.Critical("invalid workflow manager config: %v", err)
	}

	s := &v1{
		cfg:       cfg,
		db:        db,
		workflows: map[uuid.UUID]*workflowMetadata{},
		stop:      make(chan struct{}),
	}

	go s.watchHeartbeats()

	return s, nil
}

func (s *v1) Stop() {
	close(s.stop)

	s.locker.Lock()
	defer s.locker.Unlock()

	for _, workflow := range s.workflows {
		if workflow.run == nil {
			continue
		}

		close(workflow.run)
		workflow.run = nil
	}
}

func (s *v1) TrackLiveAgents(reader core_itf.LiveAgentReader) {
	s.locker.Lock()
	defer s.locker.Unlock()

	s.live = reader
}

type change struct {
	workflow *workflowMetadata
	info     *input_itf.WorkflowEntity
	step     *input_itf.StepEntity
	report   *input_itf.StepResultEntity
	progress *core_itf.WorkflowProgress
	rollback func()
}

func (s *v1) editStep(
	stepID, agentID uuid.UUID,
	kind enums.WorkflowEvent,
	edit func(workflow *workflowMetadata, step *input_itf.StepEntity) (undo func(), err error),
) (*change, error) {
	s.locker.Lock()
	defer s.locker.Unlock()

	workflow, step, found := s.findStep(stepID)
	if !found {
		return nil, custom_error.Critical("step %v not found", stepID)
	}

	prevStep := *step
	prevInfo := *workflow.info

	undo, err := edit(workflow, step)
	switch {
	case errors.Is(err, errNoChange):
		return nil, nil
	case err != nil:
		return nil, err
	}

	now := helpers.NewUTC()
	step.UpdatedAt = now
	workflow.info.UpdatedAt = now

	stepSnapshot := *step
	infoSnapshot := *workflow.info

	return &change{
		workflow: workflow,
		info:     &infoSnapshot,
		step:     &stepSnapshot,
		progress: &core_itf.WorkflowProgress{
			WorkflowID: infoSnapshot.ID,
			StepID:     stepID,
			AgentID:    agentID,
			Event:      kind,
		},
		rollback: func() {
			*step = prevStep
			*workflow.info = prevInfo

			if undo != nil {
				undo()
			}
		},
	}, nil
}

func (s *v1) persist(c *change, what string) error {
	if err := s.save(c); err != nil {
		s.locker.Lock()
		c.rollback()
		s.locker.Unlock()

		return custom_error.Critical("cannot save %s: %v", what, err)
	}

	return nil
}

// save writes every change in one transaction. Changes staged together share a
// workflow snapshot, so the workflow is deduplicated on the way in.
func (s *v1) save(changes ...*change) error {
	workflows := make([]*input_itf.WorkflowEntity, 0, 1)
	steps := make([]*input_itf.StepEntity, 0, len(changes))
	reports := make([]*input_itf.StepResultEntity, 0, len(changes))
	seen := map[uuid.UUID]bool{}

	for _, c := range changes {
		if c.info != nil && !seen[c.info.ID] {
			seen[c.info.ID] = true
			workflows = append(workflows, c.info)
		}

		if c.step != nil {
			steps = append(steps, c.step)
		}

		if c.report != nil {
			reports = append(reports, c.report)
		}
	}

	return s.db.SaveStepHistory(workflows, steps, reports)
}

func (s *v1) commit(c *change, what string) error {
	if err := s.persist(c, what); err != nil {
		return err
	}

	return s.publish(c.progress)
}

func (s *v1) settle(c *change, what string) error {
	return firstErr(s.commit(c, what), s.drainIfDone(c.workflow))
}

func firstErr(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *v1) publish(progress *core_itf.WorkflowProgress) error {
	s.locker.Lock()
	defer s.locker.Unlock()

	workflow, found := s.workflows[progress.WorkflowID]
	if !found || workflow.run == nil {
		return nil
	}

	select {
	case workflow.run <- progress:
		return nil
	default:
		return custom_error.Bypass("event %s of workflow %v dropped, channel is full", progress.Event, progress.WorkflowID)
	}
}

func (s *v1) NewWorkflow(p *core_itf.InitWorkflow) (uuid.UUID, error) {
	projectDir, err := existingDir("working dir", p.ProjectDirPath)
	if err != nil {
		return uuid.Nil, err
	}

	if err := initContext(helpers.KnowledgeDir(projectDir)); err != nil {
		return uuid.Nil, err
	}

	uid, err := uuid.NewV7()
	if err != nil {
		return uuid.Nil, err
	}

	now := helpers.NewUTC()

	info := &input_itf.WorkflowEntity{
		ID:             uid,
		ProjectDirPath: projectDir,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.db.SaveStepHistory([]*input_itf.WorkflowEntity{info}, nil, nil); err != nil {
		return uuid.Nil, custom_error.Critical("cannot save workflow info: %v", err)
	}

	s.locker.Lock()
	s.workflows[uid] = &workflowMetadata{
		info:            info,
		stepIDToStep:    map[uuid.UUID]*input_itf.StepEntity{},
		agentIDToHandle: map[uuid.UUID]*AgentHandle{},
		stepIDToReport:  map[uuid.UUID]*core_itf.StepResult{},
		spentByStep:     map[uuid.UUID]input_itf.ContextUsage{},
	}
	s.locker.Unlock()

	return uid, nil
}

func (s *v1) AddStep(workflowID uuid.UUID, step *core_itf.AddStep) (uuid.UUID, error) {
	if step.AgentSpecs == nil {
		return uuid.Nil, custom_error.Critical("step %s is missing its agent specs", step.Name)
	}

	uid, err := uuid.NewV7()
	if err != nil {
		return uuid.Nil, err
	}

	now := helpers.NewUTC()

	t := &input_itf.StepEntity{
		ID:               uid,
		WorkflowID:       workflowID,
		Name:             step.Name,
		Effort:           step.Effort,
		PreferredModel:   step.AgentSpecs.Name,
		ThinkingLevel:    step.AgentSpecs.ThinkingLevel,
		Instructions:     withOutputStructure(step.AgentSpecs.Instructions, step.OutputStructure),
		AutoRetry:        step.AutoRetry,
		PauseForReview:   step.PauseForReview,
		ExtraGuidance:    step.ExtraGuidance,
		DependsOnStepIDs: step.DependsOn,
		Status:           enums.StepNotTaken,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	c, err := s.addStep(workflowID, t, step.DependsOn)
	if err != nil {
		return uuid.Nil, err
	}

	return uid, s.commit(c, "new step")
}

func (s *v1) addStep(workflowID uuid.UUID, t *input_itf.StepEntity, dependsOn []uuid.UUID) (*change, error) {
	s.locker.Lock()
	defer s.locker.Unlock()

	workflow, found := s.workflows[workflowID]
	if !found {
		return nil, custom_error.Critical("workflow %v not found", workflowID)
	}

	for _, dep := range dependsOn {
		if _, found := workflow.stepIDToStep[dep]; !found {
			return nil, custom_error.Critical("step %s depends on unknown step %v", t.Name, dep)
		}
	}

	t.Instructions = withContextProtocol(t.Instructions, helpers.KnowledgeDir(workflow.info.ProjectDirPath))

	workflow.stepIDToStep[t.ID] = t

	prevInfo := *workflow.info
	workflow.info.TotalStep += 1
	workflow.info.CompletedAt = time.Time{}
	workflow.info.UpdatedAt = helpers.NewUTC()
	infoSnapshot := *workflow.info

	return &change{
		workflow: workflow,
		info:     &infoSnapshot,
		step:     t,
		progress: &core_itf.WorkflowProgress{
			WorkflowID: workflowID,
			StepID:     t.ID,
			Event:      enums.WorkflowStepCreated,
		},
		rollback: func() {
			delete(workflow.stepIDToStep, t.ID)
			*workflow.info = prevInfo
		},
	}, nil
}

func (s *v1) Assign(stepID, agentID uuid.UUID) error {
	c, err := s.editStep(stepID, agentID, enums.WorkflowStepStatusChanged,
		func(workflow *workflowMetadata, step *input_itf.StepEntity) (func(), error) {
			if workflow.halt != enums.HaltNone {
				return nil, custom_error.Critical("workflow %v is %s", workflow.info.ID, workflow.halt)
			}

			if !step.Status.Takeable() {
				return nil, custom_error.Critical("step %v is %s and cannot be assigned", stepID, step.Status)
			}

			now := helpers.NewUTC()

			step.Status = enums.StepProcessing
			workflow.info.CompletedAt = time.Time{}

			workflow.agentIDToHandle[agentID] = &AgentHandle{
				AgentID:       agentID,
				StepID:        stepID,
				AssignedAt:    now,
				LastHeartBeat: helpers.NewUTCUnix(),
			}

			if workflow.info.StartedAt.IsZero() {
				workflow.info.StartedAt = now
			}

			return func() { delete(workflow.agentIDToHandle, agentID) }, nil
		})
	if err != nil {
		return err
	}

	return s.commit(c, "step assignment")
}

// Everything the revert point does not vouch for goes back into the pool, not only what
// hangs off it: a revert cancels the whole run first, so a parallel branch that never
// finished would otherwise sit at cancelled forever with nothing to take it.
func (s *v1) RewindTo(stepID uuid.UUID) error {
	workflow, undone, err := s.rewoundBy(stepID)
	if err != nil {
		return err
	}

	commitErrs := make([]error, 0, len(undone))

	for _, dependent := range undone {
		c, err := s.editStep(dependent, uuid.Nil, enums.WorkflowStepStatusChanged,
			func(workflow *workflowMetadata, step *input_itf.StepEntity) (func(), error) {
				kept, hadReport := workflow.stepIDToReport[step.ID]

				step.Status = enums.StepNotTaken
				delete(workflow.stepIDToReport, step.ID)
				workflow.info.CompletedAt = time.Time{}

				return func() {
					if hadReport {
						workflow.stepIDToReport[step.ID] = kept
					}
				}, nil
			})
		if err != nil {
			return err
		}

		commitErrs = append(commitErrs, s.commit(c, "step rewind"))
	}

	s.resume(workflow)

	return firstErr(commitErrs...)
}

// The revert point keeps its own result, and so does any step outside its shadow that
// already reached one. Every other step — the ones downstream of it, and the ones a
// cancel parked mid-flight elsewhere in the graph — is undone.
func (s *v1) rewoundBy(stepID uuid.UUID) (*workflowMetadata, []uuid.UUID, error) {
	s.locker.Lock()
	defer s.locker.Unlock()

	workflow, _, found := s.findStep(stepID)
	if !found {
		return nil, nil, custom_error.Critical("step %v not found", stepID)
	}

	downstream := map[uuid.UUID]bool{stepID: true}

	for grew := true; grew; {
		grew = false

		for id, step := range workflow.stepIDToStep {
			if downstream[id] {
				continue
			}

			for _, dep := range step.DependsOnStepIDs {
				if downstream[dep] {
					downstream[id] = true
					grew = true

					break
				}
			}
		}
	}

	ids := make([]uuid.UUID, 0, len(workflow.stepIDToStep))

	for id, step := range workflow.stepIDToStep {
		if id == stepID {
			continue
		}

		if !downstream[id] && step.Status.Reached() {
			continue
		}

		ids = append(ids, id)
	}

	sort.Slice(ids, func(i, j int) bool { return ids[i].String() < ids[j].String() })

	return workflow, ids, nil
}

func (s *v1) AnswerReview(stepID uuid.UUID, accepted bool) error {
	c, err := s.editStep(stepID, uuid.Nil, enums.WorkflowStepStatusChanged,
		func(_ *workflowMetadata, step *input_itf.StepEntity) (func(), error) {
			if step.Status != enums.StepAwaitingReview {
				return nil, custom_error.Critical("step %v is %s and is not awaiting acceptance", stepID, step.Status)
			}

			step.Status = enums.StepFailed
			if accepted {
				step.Status = enums.StepCompleted
			}

			return nil, nil
		})
	if err != nil {
		return err
	}

	return s.settle(c, "step acceptance")
}

func (s *v1) Report(agentID uuid.UUID, status enums.StepStatus, docs []*core_itf.Handoff) error {
	if status != enums.StepCompleted && status != enums.StepFailed {
		return custom_error.Critical("agent %v reported unsupported status %s", agentID, status)
	}

	if len(docs) == 0 {
		return custom_error.Critical("report from agent %v is missing a handoff", agentID)
	}

	for _, doc := range docs {
		if doc == nil {
			return custom_error.Critical("report from agent %v has an empty handoff", agentID)
		}
	}

	handle, found := s.handleOf(agentID)
	if !found {
		return custom_error.Critical("agent %v is not assigned to any step", agentID)
	}

	reportID, err := uuid.NewV7()
	if err != nil {
		return custom_error.Critical("cannot create uuid: %v", err)
	}

	report := &core_itf.StepResult{
		StepID:       handle.StepID,
		Status:       status,
		Handoffs:     docs,
		ContextUsage: s.readContextUsage(agentID),
	}

	c, err := s.editStep(handle.StepID, agentID, enums.WorkflowStepResulted,
		func(workflow *workflowMetadata, step *input_itf.StepEntity) (func(), error) {
			if step.Status != enums.StepProcessing {
				return nil, custom_error.Critical("step %v is %s and cannot be reported", step.ID, step.Status)
			}

			for _, doc := range docs {
				doc.Step = step.Name
			}

			step.Status = status
			if status == enums.StepCompleted && step.PauseForReview {
				step.Status = enums.StepAwaitingReview
			}

			step.LastReportID = reportID

			if !status.Removable() {
				step.RetryCount += 1
			}

			if status == enums.StepFailed {
				workflow.info.TotalRetry += 1
			}

			return nil, nil
		})
	if err != nil {
		return err
	}

	c.report = &input_itf.StepResultEntity{
		ID:            reportID,
		StepID:        handle.StepID,
		AgentID:       agentID,
		AttemptStatus: status,
		Handoffs:      handoffEntities(docs),
		ContextUsage:  report.ContextUsage,
		StartedAt:     handle.AssignedAt,
		CompletedAt:   helpers.NewUTC(),
		CreatedAt:     helpers.NewUTC(),
		UpdatedAt:     helpers.NewUTC(),
	}

	if err := s.persist(c, "step report"); err != nil {
		return err
	}

	s.closeHandle(c.workflow, agentID, report)

	return firstErr(s.publish(c.progress), s.drainIfDone(c.workflow))
}

// Runs after the database write, so a failed write leaves the agent still holding its step.
func (s *v1) closeHandle(workflow *workflowMetadata, agentID uuid.UUID, report *core_itf.StepResult) {
	s.locker.Lock()
	defer s.locker.Unlock()

	delete(workflow.agentIDToHandle, agentID)
	workflow.keepReport(report)
}

func handoffEntities(docs []*core_itf.Handoff) []*input_itf.HandoffEntity {
	entities := make([]*input_itf.HandoffEntity, 0, len(docs))

	for _, doc := range docs {
		entities = append(entities, (*input_itf.HandoffEntity)(doc))
	}

	return entities
}

func handoffs(entities []*input_itf.HandoffEntity) []*core_itf.Handoff {
	docs := make([]*core_itf.Handoff, 0, len(entities))

	for _, entity := range entities {
		docs = append(docs, (*core_itf.Handoff)(entity))
	}

	return docs
}

func (s *v1) Cancel(workflowID uuid.UUID) ([]uuid.UUID, error) {
	return s.haltWorkflow(workflowID, enums.HaltCancelled)
}

func (s *v1) Pause(workflowID uuid.UUID) ([]uuid.UUID, error) {
	return s.haltWorkflow(workflowID, enums.HaltPaused)
}

func (s *v1) haltWorkflow(workflowID uuid.UUID, halt enums.WorkflowHalt) ([]uuid.UUID, error) {
	changes, retired, err := s.stageHalt(workflowID, halt)
	if err != nil {
		return nil, err
	}

	agentIDs := make([]uuid.UUID, 0, len(retired))
	for _, handle := range retired {
		agentIDs = append(agentIDs, handle.AgentID)
	}

	if err := s.save(changes...); err != nil {
		s.locker.Lock()
		for _, c := range changes {
			c.rollback()
		}
		s.locker.Unlock()

		return nil, custom_error.Critical("cannot save workflow %s: %v", halt, err)
	}

	workflow, found := s.workflow(workflowID)
	if found {
		s.retireTokens(workflow, retired)
	}

	publishErrs := make([]error, 0, len(changes)+1)

	for _, c := range changes {
		publishErrs = append(publishErrs, s.publish(c.progress))
	}

	if found {
		publishErrs = append(publishErrs, s.drainIfDone(workflow))
	}

	return agentIDs, firstErr(publishErrs...)
}

// A halted agent never reports, so what it already spent is read before the drop is
// published: the coordinator answers that event by killing the agent, and a killed
// agent has no usage left to read. The handles come from stageHalt because the
// workflow no longer holds them by the time the tokens are credited.
func (s *v1) retireTokens(workflow *workflowMetadata, retired []*AgentHandle) {
	for _, handle := range retired {
		s.creditTokens(workflow, handle.StepID, s.readContextUsage(handle.AgentID))
	}
}

func (s *v1) creditTokens(workflow *workflowMetadata, stepID uuid.UUID, usage *input_itf.ContextUsage) {
	if usage == nil {
		return
	}

	s.locker.Lock()
	workflow.creditSpent(stepID, usage)
	s.locker.Unlock()
}

func (s *v1) stageHalt(workflowID uuid.UUID, halt enums.WorkflowHalt) ([]*change, []*AgentHandle, error) {
	s.locker.Lock()
	defer s.locker.Unlock()

	workflow, found := s.workflows[workflowID]
	if !found {
		return nil, nil, custom_error.Critical("workflow %v not found to mark as %s", workflowID, halt)
	}

	infoSnapshot := *workflow.info
	changes := []*change{}

	for stepID, step := range workflow.stepIDToStep {
		parked, parks := halt.Park(step.Status)
		if !parks {
			continue
		}

		kind := enums.WorkflowStepStatusChanged
		agentID := uuid.Nil
		handle, taken := workflow.findHandle(stepID)

		if taken {
			agentID = handle.AgentID
			kind = enums.WorkflowStepDropped
		}

		prevStep := *step
		step.Status = parked
		step.UpdatedAt = helpers.NewUTC()
		stepSnapshot := *step

		changes = append(changes, &change{
			workflow: workflow,
			info:     &infoSnapshot,
			step:     &stepSnapshot,
			progress: &core_itf.WorkflowProgress{
				WorkflowID: workflowID,
				StepID:     stepID,
				AgentID:    agentID,
				Event:      kind,
			},
			rollback: func() {
				*step = prevStep

				if taken {
					workflow.agentIDToHandle[agentID] = handle
				}
			},
		})
	}

	retired := make([]*AgentHandle, 0, len(workflow.agentIDToHandle))
	for _, handle := range workflow.agentIDToHandle {
		clone := *handle
		retired = append(retired, &clone)
	}

	sort.Slice(retired, func(i, j int) bool {
		return retired[i].AgentID.String() < retired[j].AgentID.String()
	})

	workflow.agentIDToHandle = map[uuid.UUID]*AgentHandle{}
	workflow.halt = halt

	return changes, retired, nil
}

func (s *v1) Restore() error {
	snapshots, err := s.db.LoadStepHistory()
	if err != nil {
		return custom_error.Critical("cannot load step history: %v", err)
	}

	interrupted := []uuid.UUID{}

	s.locker.Lock()
	for _, snapshot := range snapshots {
		workflow := restored(snapshot)
		s.workflows[workflow.info.ID] = workflow

		if !workflow.info.StartedAt.IsZero() && workflow.info.CompletedAt.IsZero() {
			interrupted = append(interrupted, workflow.info.ID)
		}
	}
	s.locker.Unlock()

	for _, workflowID := range interrupted {
		if _, err := s.Pause(workflowID); err != nil {
			return err
		}
	}

	return nil
}

func restored(snapshot *input_itf.WorkflowSnapshot) *workflowMetadata {
	workflow := &workflowMetadata{
		info:            snapshot.Workflow,
		stepIDToStep:    make(map[uuid.UUID]*input_itf.StepEntity, len(snapshot.Steps)),
		agentIDToHandle: map[uuid.UUID]*AgentHandle{},
		stepIDToReport:  map[uuid.UUID]*core_itf.StepResult{},
		spentByStep:     map[uuid.UUID]input_itf.ContextUsage{},
	}

	for _, step := range snapshot.Steps {
		workflow.stepIDToStep[step.ID] = step
	}

	for _, report := range snapshot.Reports {
		workflow.keepReport(&core_itf.StepResult{
			StepID:       report.StepID,
			Status:       report.AttemptStatus,
			Handoffs:     handoffs(report.Handoffs),
			ContextUsage: report.ContextUsage,
		})
	}

	return workflow
}

func (s *v1) resume(workflow *workflowMetadata) {
	s.locker.Lock()
	defer s.locker.Unlock()

	workflow.halt = enums.HaltNone
}

func (s *v1) drainIfDone(workflow *workflowMetadata) error {
	c, drained := s.stageDrain(workflow)
	if !drained {
		return nil
	}

	return s.commit(c, "workflow drained")
}

func (s *v1) stageDrain(workflow *workflowMetadata) (*change, bool) {
	s.locker.Lock()
	defer s.locker.Unlock()

	if !workflow.info.CompletedAt.IsZero() || len(workflow.stepIDToStep) == 0 {
		return nil, false
	}

	for _, step := range workflow.stepIDToStep {
		if !step.Status.Removable() {
			return nil, false
		}
	}

	prevInfo := *workflow.info
	now := helpers.NewUTC()
	workflow.info.CompletedAt = now
	workflow.info.UpdatedAt = now
	infoSnapshot := *workflow.info

	return &change{
		workflow: workflow,
		info:     &infoSnapshot,
		progress: &core_itf.WorkflowProgress{
			WorkflowID: infoSnapshot.ID,
			Event:      enums.WorkflowDrained,
		},
		rollback: func() { *workflow.info = prevInfo },
	}, true
}

func (s *v1) ReadySteps(workflowID uuid.UUID) ([]*core_itf.StepSpec, error) {
	s.locker.Lock()

	workflow, found := s.workflows[workflowID]
	if !found {
		s.locker.Unlock()
		return nil, custom_error.Critical("workflow %v not found", workflowID)
	}

	specs := []*core_itf.StepSpec{}

	if workflow.halt == enums.HaltNone {
		for stepID, step := range workflow.stepIDToStep {
			if _, taken := workflow.findHandle(stepID); taken {
				continue
			}

			if !workflow.isReady(step) {
				continue
			}

			specs = append(specs, stepSpec(stepID, step))
		}
	}

	s.locker.Unlock()

	sort.Slice(specs, func(i, j int) bool {
		return specs[i].StepID.String() < specs[j].StepID.String()
	})

	return specs, nil
}

func stepSpec(stepID uuid.UUID, t *input_itf.StepEntity) *core_itf.StepSpec {
	return &core_itf.StepSpec{
		StepID:         stepID,
		Name:           t.Name,
		PauseForReview: t.PauseForReview,
		ExtraGuidance:  t.ExtraGuidance,
		DependsOn:      t.DependsOnStepIDs,
		AgentSpecs: &core_itf.AgentRequest{
			Name:          t.PreferredModel,
			ThinkingLevel: t.ThinkingLevel,
			Instructions:  t.Instructions,
		},
	}
}

func (s *v1) Status(id uuid.UUID) (*core_itf.WorkflowStatus, error) {
	s.locker.Lock()

	workflow, found := s.workflows[id]
	if !found {
		s.locker.Unlock()
		return nil, custom_error.Critical("workflow %v not found", id)
	}

	stepToAgent := map[uuid.UUID]uuid.UUID{}
	for agentID, handle := range workflow.agentIDToHandle {
		stepToAgent[handle.StepID] = agentID
	}

	steps := map[uuid.UUID]*core_itf.StepResult{}

	for stepID, step := range workflow.stepIDToStep {
		spent := workflow.spentByStep[stepID]

		report := &core_itf.StepResult{
			StepID:   stepID,
			Name:     step.Name,
			Effort:   step.Effort,
			Model:    step.PreferredModel,
			Status:   step.Status,
			Handoffs: []*core_itf.Handoff{},
			Spent:    &spent,
		}

		if reported, found := workflow.stepIDToReport[stepID]; found {
			report.Handoffs = append(report.Handoffs, reported.Handoffs...)
			report.ContextUsage = reported.ContextUsage
		}

		steps[stepID] = report
	}

	status := &core_itf.WorkflowStatus{
		ID:             id,
		Status:         workflowStatus(workflow),
		ProjectDirPath: workflow.info.ProjectDirPath,
		Steps:          steps,
		StartedAt:      workflow.info.StartedAt,
		CompletedAt:    workflow.info.CompletedAt,
	}

	s.locker.Unlock()

	// Outside the lock: reading usage takes the same lock to reach the reader.
	// An agent still working has not reported, so its window is what the step has
	// spent on top of the attempts already accumulated.
	for stepID, agentID := range stepToAgent {
		step, found := status.Steps[stepID]
		if !found {
			continue
		}

		step.AgentID = agentID

		if usage := s.readContextUsage(agentID); usage != nil {
			step.ContextUsage = usage

			spent := creditedWith(*step.Spent, usage)
			step.Spent = &spent
		}

		step.Activity = s.readActivity(agentID)
	}

	for _, step := range status.Steps {
		status.TokensBilled += step.Spent.Billed
		status.TokensInput += step.Spent.Input
		status.TokensCached += step.Spent.Cached
	}

	return status, nil
}

func (s *v1) Execute(workflowID uuid.UUID) (<-chan *core_itf.WorkflowProgress, error) {
	workflow, found := s.workflow(workflowID)
	if !found {
		return nil, custom_error.Critical("workflow %v not found", workflowID)
	}

	s.retireRun(workflow)
	s.resume(workflow)

	running := s.beginRun(workflow)
	if running == nil {
		return nil, custom_error.Critical("workflow manager is stopped")
	}

	return running, nil
}

func (s *v1) retireRun(workflow *workflowMetadata) {
	s.locker.Lock()
	running := workflow.run
	workflow.run = nil
	s.locker.Unlock()

	if running == nil {
		return
	}

	close(running)
}

func (s *v1) beginRun(workflow *workflowMetadata) <-chan *core_itf.WorkflowProgress {
	s.locker.Lock()
	defer s.locker.Unlock()

	select {
	case <-s.stop:
		return nil
	default:
	}

	workflow.run = make(chan *core_itf.WorkflowProgress, progressBufferSize)

	return workflow.run
}

func (s *v1) HeartBeat(agentID uuid.UUID) error {
	s.locker.Lock()
	defer s.locker.Unlock()

	handle, found := s.findAgent(agentID)
	if !found {
		return custom_error.Critical("agent %v is not assigned to any step", agentID)
	}

	handle.LastHeartBeat = helpers.NewUTCUnix()

	return nil
}

func (s *v1) watchHeartbeats() {
	ticker := time.NewTicker(s.cfg.HeartbeatScanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stop:
			return
		case <-ticker.C:
			deadline := helpers.NewUTCUnix() - int64(s.cfg.HeartbeatTimeout.Seconds())

			for _, agentID := range s.staleAgents(deadline) {
				_ = s.dropStep(agentID, deadline)
			}
		}
	}
}

func (s *v1) staleAgents(deadline int64) []uuid.UUID {
	s.locker.Lock()
	defer s.locker.Unlock()

	stale := []uuid.UUID{}

	for _, workflow := range s.workflows {
		for agentID, handle := range workflow.agentIDToHandle {
			if handle.LastHeartBeat <= deadline {
				stale = append(stale, agentID)
			}
		}
	}

	return stale
}

// The agent may have reported since it was found stale, so the deadline is checked
// again under the lock.
func (s *v1) dropStep(agentID uuid.UUID, deadline int64) error {
	handle, alive := s.handleOf(agentID)
	if !alive {
		return nil
	}

	spent := s.readContextUsage(agentID)

	c, err := s.editStep(handle.StepID, agentID, enums.WorkflowStepDropped,
		func(workflow *workflowMetadata, step *input_itf.StepEntity) (func(), error) {
			current, found := workflow.agentIDToHandle[agentID]
			if !found || current.LastHeartBeat > deadline || step.Status != enums.StepProcessing {
				return nil, errNoChange
			}

			step.Status = enums.StepCancelled
			delete(workflow.agentIDToHandle, agentID)

			return func() { workflow.agentIDToHandle[agentID] = current }, nil
		})
	if err != nil {
		return custom_error.Critical("cannot drop step of stale agent %v: %v", agentID, err)
	}

	if c == nil {
		return nil
	}

	if err := s.persist(c, "step drop"); err != nil {
		return err
	}

	s.creditTokens(c.workflow, handle.StepID, spent)

	return firstErr(s.publish(c.progress), s.drainIfDone(c.workflow))
}

func (s *v1) workflow(workflowID uuid.UUID) (*workflowMetadata, bool) {
	s.locker.Lock()
	defer s.locker.Unlock()

	workflow, found := s.workflows[workflowID]

	return workflow, found
}

func (s *v1) handleOf(agentID uuid.UUID) (*AgentHandle, bool) {
	s.locker.Lock()
	defer s.locker.Unlock()

	handle, found := s.findAgent(agentID)
	if !found {
		return nil, false
	}

	clone := *handle

	return &clone, true
}

// findAgent and findStep read shared state; the caller holds the lock.
func (s *v1) findAgent(agentID uuid.UUID) (*AgentHandle, bool) {
	for _, workflow := range s.workflows {
		if handle, found := workflow.agentIDToHandle[agentID]; found {
			return handle, true
		}
	}

	return nil, false
}

func (s *v1) findStep(stepID uuid.UUID) (*workflowMetadata, *input_itf.StepEntity, bool) {
	for _, workflow := range s.workflows {
		if step, found := workflow.stepIDToStep[stepID]; found {
			return workflow, step, true
		}
	}

	return nil, nil, false
}

func (s *v1) liveReader() core_itf.LiveAgentReader {
	s.locker.Lock()
	defer s.locker.Unlock()

	return s.live
}

// The reading has to happen while the agent is still alive: the coordinator kills it
// as soon as it sees the report event, and a dead agent has no window left to read.
func (s *v1) readContextUsage(agentID uuid.UUID) *input_itf.ContextUsage {
	reader := s.liveReader()
	if reader == nil {
		return nil
	}

	usage, err := reader.ContextUsage(agentID)
	if err != nil {
		return nil
	}

	return usage
}

func (s *v1) readActivity(agentID uuid.UUID) []input_itf.Activity {
	reader := s.liveReader()
	if reader == nil {
		return nil
	}

	activity, err := reader.Activity(agentID)
	if err != nil {
		return nil
	}

	return activity
}

func (m *workflowMetadata) isReady(t *input_itf.StepEntity) bool {
	for _, dep := range t.DependsOnStepIDs {
		depStep, found := m.stepIDToStep[dep]
		if !found || depStep.Status != enums.StepCompleted {
			return false
		}
	}

	if t.Status == enums.StepNotTaken {
		return true
	}

	return t.AutoRetry && t.Status.Retryable() && t.RetryCount < maxAutoRetry
}

func (m *workflowMetadata) findHandle(stepID uuid.UUID) (*AgentHandle, bool) {
	for _, handle := range m.agentIDToHandle {
		if handle.StepID == stepID {
			return handle, true
		}
	}

	return nil, false
}

func (m *workflowMetadata) keepReport(report *core_itf.StepResult) {
	kept, found := m.stepIDToReport[report.StepID]
	if !found {
		kept = &core_itf.StepResult{
			StepID:   report.StepID,
			Handoffs: []*core_itf.Handoff{},
		}

		m.stepIDToReport[report.StepID] = kept
	}

	kept.Status = report.Status
	kept.Handoffs = append(kept.Handoffs, report.Handoffs...)

	if report.ContextUsage != nil {
		kept.ContextUsage = report.ContextUsage
		m.creditSpent(report.StepID, report.ContextUsage)
	}
}

func (m *workflowMetadata) creditSpent(stepID uuid.UUID, usage *input_itf.ContextUsage) {
	m.spentByStep[stepID] = creditedWith(m.spentByStep[stepID], usage)
}

func creditedWith(spent input_itf.ContextUsage, usage *input_itf.ContextUsage) input_itf.ContextUsage {
	if usage == nil {
		return spent
	}

	spent.Billed += usage.Billed
	spent.Input += usage.Input
	spent.Cached += usage.Cached

	return spent
}

func workflowStatus(workflow *workflowMetadata) enums.WorkflowStatus {
	switch {
	case !workflow.info.CompletedAt.IsZero():
		return enums.WorkflowCompleted
	case workflow.halt == enums.HaltPaused:
		return enums.WorkflowPaused
	case !workflow.info.StartedAt.IsZero():
		return enums.WorkflowProcessing
	default:
		return enums.WorkflowInit
	}
}

func existingDir(name, path string) (string, error) {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return "", custom_error.Critical("workflow %s is empty", name)
	}

	cleaned, err := absPath(name, trimmed)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(cleaned)
	switch {
	case os.IsNotExist(err):
		return "", custom_error.Critical("workflow %s %q does not exist", name, cleaned)
	case err != nil:
		return "", dirError(name, cleaned, "accessed", err)
	case !info.IsDir():
		return "", custom_error.Critical("workflow %s %q is not a directory", name, cleaned)
	}

	dir, err := os.Open(cleaned)
	if err != nil {
		return "", dirError(name, cleaned, "opened", err)
	}
	defer dir.Close()

	if _, err := dir.ReadDir(1); err != nil && !errors.Is(err, io.EOF) {
		return "", dirError(name, cleaned, "read", err)
	}

	return cleaned, nil
}

func absPath(name, path string) (string, error) {
	if !filepath.IsAbs(path) {
		return "", custom_error.Critical("workflow %s %q must be an absolute path", name, path)
	}

	return filepath.Clean(path), nil
}

func dirError(name, path, verb string, err error) error {
	if os.IsPermission(err) {
		return custom_error.Critical("workflow %s %q cannot be %s: permission denied", name, path, verb)
	}

	return custom_error.Critical("workflow %s %q cannot be %s: %v", name, path, verb, err)
}
