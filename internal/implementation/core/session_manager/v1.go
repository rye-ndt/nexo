package session_manager

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
	output_itf "hexago/internal/interface/output"

	"github.com/google/uuid"
)

const progressBufferSize = 32

const maxAutoRetry = 3

// errNoChange makes editTask return a nil change rather than an error.
var errNoChange = errors.New("task needs no change")

type AgentHandle struct {
	AgentID       uuid.UUID
	TaskID        uuid.UUID
	AssignedAt    time.Time
	LastHeartBeat int64
}

type sessionMetadata struct {
	info            *input_itf.SessionEntity
	taskIDToTask    map[uuid.UUID]*input_itf.TaskEntity
	agentIDToHandle map[uuid.UUID]*AgentHandle
	taskIDToReport  map[uuid.UUID]*core_itf.TaskReport
	tokensBilled    int
	cancelled       bool
	run             chan struct{}
}

type v1 struct {
	locker   sync.Mutex
	cfg      *input_itf.SessionConfig
	db       input_itf.TaskStorage
	mq       output_itf.MessageQ
	live     core_itf.LiveAgentReader
	sessions map[uuid.UUID]*sessionMetadata
	stop     chan struct{}
}

func InitV1(cfg *input_itf.SessionConfig, db input_itf.TaskStorage, mq output_itf.MessageQ) (core_itf.SessionManager, error) {
	if err := helpers.ValidateStruct(cfg); err != nil {
		return nil, custom_error.Critical("invalid session manager config: %v", err)
	}

	s := &v1{
		cfg:      cfg,
		db:       db,
		mq:       mq,
		sessions: map[uuid.UUID]*sessionMetadata{},
		stop:     make(chan struct{}),
	}

	go s.watchHeartbeats()

	return s, nil
}

func (s *v1) Stop() {
	close(s.stop)
}

func (s *v1) TrackLiveAgents(reader core_itf.LiveAgentReader) {
	s.locker.Lock()
	defer s.locker.Unlock()

	s.live = reader
}

type change struct {
	session  *sessionMetadata
	info     *input_itf.SessionEntity
	task     *input_itf.TaskEntity
	report   *input_itf.TaskReportEntity
	progress *core_itf.SessionProgress
	rollback func()
}

func (s *v1) editTask(
	taskID, agentID uuid.UUID,
	kind enums.SessionEvent,
	edit func(session *sessionMetadata, task *input_itf.TaskEntity) (undo func(), err error),
) (*change, error) {
	s.locker.Lock()
	defer s.locker.Unlock()

	session, task, found := s.findTask(taskID)
	if !found {
		return nil, custom_error.Critical("task %v not found", taskID)
	}

	prevTask := *task
	prevInfo := *session.info

	undo, err := edit(session, task)
	switch {
	case errors.Is(err, errNoChange):
		return nil, nil
	case err != nil:
		return nil, err
	}

	now := helpers.NewUTC()
	task.UpdatedAt = now
	session.info.UpdatedAt = now

	taskSnapshot := *task
	infoSnapshot := *session.info

	return &change{
		session: session,
		info:    &infoSnapshot,
		task:    &taskSnapshot,
		progress: &core_itf.SessionProgress{
			SessionID: infoSnapshot.ID,
			TaskID:    taskID,
			AgentID:   agentID,
			Event:     kind,
		},
		rollback: func() {
			*task = prevTask
			*session.info = prevInfo

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
// session snapshot, so the session is deduplicated on the way in.
func (s *v1) save(changes ...*change) error {
	sessions := make([]*input_itf.SessionEntity, 0, 1)
	tasks := make([]*input_itf.TaskEntity, 0, len(changes))
	reports := make([]*input_itf.TaskReportEntity, 0, len(changes))
	seen := map[uuid.UUID]bool{}

	for _, c := range changes {
		if c.info != nil && !seen[c.info.ID] {
			seen[c.info.ID] = true
			sessions = append(sessions, c.info)
		}

		if c.task != nil {
			tasks = append(tasks, c.task)
		}

		if c.report != nil {
			reports = append(reports, c.report)
		}
	}

	return s.db.SaveTaskHistory(sessions, tasks, reports)
}

func (s *v1) commit(c *change, what string) error {
	if err := s.persist(c, what); err != nil {
		return err
	}

	return s.publish(c.progress)
}

func (s *v1) settle(c *change, what string) error {
	return firstErr(s.commit(c, what), s.drainIfDone(c.session))
}

func firstErr(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *v1) publish(progress *core_itf.SessionProgress) error {
	return s.mq.Emit(progress.SessionID, progress.Event, progress)
}

func (s *v1) NewSession(p *core_itf.InitSession) (uuid.UUID, error) {
	workingDir, err := existingDir("working dir", p.WorkingDirPath)
	if err != nil {
		return uuid.Nil, err
	}

	contextDir, err := makeDir("context dir", p.ContextDirPath, workingDir)
	if err != nil {
		return uuid.Nil, err
	}

	if err := initContext(contextDir); err != nil {
		return uuid.Nil, err
	}

	uid, err := uuid.NewV7()
	if err != nil {
		return uuid.Nil, err
	}

	now := helpers.NewUTC()

	info := &input_itf.SessionEntity{
		ID:             uid,
		WorkingDirPath: workingDir,
		ContextDirPath: contextDir,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.db.SaveTaskHistory([]*input_itf.SessionEntity{info}, nil, nil); err != nil {
		return uuid.Nil, custom_error.Critical("cannot save session info: %v", err)
	}

	s.locker.Lock()
	s.sessions[uid] = &sessionMetadata{
		info:            info,
		taskIDToTask:    map[uuid.UUID]*input_itf.TaskEntity{},
		agentIDToHandle: map[uuid.UUID]*AgentHandle{},
		taskIDToReport:  map[uuid.UUID]*core_itf.TaskReport{},
	}
	s.locker.Unlock()

	return uid, nil
}

func (s *v1) AddTask(sessionID uuid.UUID, task *core_itf.AddTask) (uuid.UUID, error) {
	if task.AgentSpecs == nil {
		return uuid.Nil, custom_error.Critical("task %s is missing its agent specs", task.Name)
	}

	uid, err := uuid.NewV7()
	if err != nil {
		return uuid.Nil, err
	}

	now := helpers.NewUTC()

	t := &input_itf.TaskEntity{
		ID:                   uid,
		SessionID:            sessionID,
		Name:                 task.Name,
		PreferredModel:       task.AgentSpecs.Name,
		ThinkingLevel:        task.AgentSpecs.ThinkingLevel,
		SystemPrompts:        withOutputStructure(task.AgentSpecs.SystemPrompts, task.OutputStructure),
		AutoRetry:            task.AutoRetry,
		ManualAcceptRequired: task.ManualAcceptRequired,
		ExtraGuidance:        task.ExtraGuidance,
		DependsOnTaskIDs:     task.DependsOn,
		Status:               enums.TaskNotTaken,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	c, err := s.addTask(sessionID, t, task.DependsOn)
	if err != nil {
		return uuid.Nil, err
	}

	return uid, s.commit(c, "new task")
}

func (s *v1) addTask(sessionID uuid.UUID, t *input_itf.TaskEntity, dependsOn []uuid.UUID) (*change, error) {
	s.locker.Lock()
	defer s.locker.Unlock()

	session, found := s.sessions[sessionID]
	if !found {
		return nil, custom_error.Critical("session %v not found", sessionID)
	}

	for _, dep := range dependsOn {
		if _, found := session.taskIDToTask[dep]; !found {
			return nil, custom_error.Critical("task %s depends on unknown task %v", t.Name, dep)
		}
	}

	t.SystemPrompts = withContextProtocol(t.SystemPrompts, session.info.ContextDirPath)

	session.taskIDToTask[t.ID] = t

	prevInfo := *session.info
	session.info.TotalTask += 1
	session.info.CompletedAt = time.Time{}
	session.info.UpdatedAt = helpers.NewUTC()
	infoSnapshot := *session.info

	return &change{
		session: session,
		info:    &infoSnapshot,
		task:    t,
		progress: &core_itf.SessionProgress{
			SessionID: sessionID,
			TaskID:    t.ID,
			Event:     enums.SessionTaskCreated,
		},
		rollback: func() {
			delete(session.taskIDToTask, t.ID)
			*session.info = prevInfo
		},
	}, nil
}

func (s *v1) Assign(taskID, agentID uuid.UUID) error {
	c, err := s.editTask(taskID, agentID, enums.SessionTaskStatusChanged,
		func(session *sessionMetadata, task *input_itf.TaskEntity) (func(), error) {
			if session.cancelled {
				return nil, custom_error.Critical("session %v is cancelled", session.info.ID)
			}

			if !task.Status.Takeable() {
				return nil, custom_error.Critical("task %v is %s and cannot be assigned", taskID, task.Status)
			}

			now := helpers.NewUTC()

			task.Status = enums.TaskProcessing
			session.info.CompletedAt = time.Time{}

			session.agentIDToHandle[agentID] = &AgentHandle{
				AgentID:       agentID,
				TaskID:        taskID,
				AssignedAt:    now,
				LastHeartBeat: helpers.NewUTCUnix(),
			}

			if session.info.StartedAt.IsZero() {
				session.info.StartedAt = now
			}

			return func() { delete(session.agentIDToHandle, agentID) }, nil
		})
	if err != nil {
		return err
	}

	return s.commit(c, "task assignment")
}

func (s *v1) RetryTask(taskID uuid.UUID) error {
	c, err := s.editTask(taskID, uuid.Nil, enums.SessionTaskStatusChanged,
		func(session *sessionMetadata, task *input_itf.TaskEntity) (func(), error) {
			if !task.Status.Takeable() {
				return nil, custom_error.Critical("task %v is %s and cannot be retried", taskID, task.Status)
			}

			task.Status = enums.TaskNotTaken
			session.info.CompletedAt = time.Time{}

			return nil, nil
		})
	if err != nil {
		return err
	}

	if err := s.commit(c, "task retry"); err != nil {
		return err
	}

	s.resume(c.session)

	return nil
}

func (s *v1) RewindTo(taskID uuid.UUID) error {
	session, dependents, err := s.dependentsOf(taskID)
	if err != nil {
		return err
	}

	commitErrs := make([]error, 0, len(dependents))

	for _, dependent := range dependents {
		c, err := s.editTask(dependent, uuid.Nil, enums.SessionTaskStatusChanged,
			func(session *sessionMetadata, task *input_itf.TaskEntity) (func(), error) {
				kept, hadReport := session.taskIDToReport[task.ID]

				task.Status = enums.TaskNotTaken
				delete(session.taskIDToReport, task.ID)
				session.info.CompletedAt = time.Time{}

				return func() {
					if hadReport {
						session.taskIDToReport[task.ID] = kept
					}
				}, nil
			})
		if err != nil {
			return err
		}

		commitErrs = append(commitErrs, s.commit(c, "task rewind"))
	}

	s.resume(session)

	return firstErr(commitErrs...)
}

func (s *v1) dependentsOf(taskID uuid.UUID) (*sessionMetadata, []uuid.UUID, error) {
	s.locker.Lock()
	defer s.locker.Unlock()

	session, _, found := s.findTask(taskID)
	if !found {
		return nil, nil, custom_error.Critical("task %v not found", taskID)
	}

	downstream := map[uuid.UUID]bool{taskID: true}

	for grew := true; grew; {
		grew = false

		for id, task := range session.taskIDToTask {
			if downstream[id] {
				continue
			}

			for _, dep := range task.DependsOnTaskIDs {
				if downstream[dep] {
					downstream[id] = true
					grew = true

					break
				}
			}
		}
	}

	delete(downstream, taskID)

	ids := make([]uuid.UUID, 0, len(downstream))
	for id := range downstream {
		ids = append(ids, id)
	}

	sort.Slice(ids, func(i, j int) bool { return ids[i].String() < ids[j].String() })

	return session, ids, nil
}

func (s *v1) AnswerAcceptance(taskID uuid.UUID, accepted bool) error {
	c, err := s.editTask(taskID, uuid.Nil, enums.SessionTaskStatusChanged,
		func(_ *sessionMetadata, task *input_itf.TaskEntity) (func(), error) {
			if task.Status != enums.TaskAwaitingAccept {
				return nil, custom_error.Critical("task %v is %s and is not awaiting acceptance", taskID, task.Status)
			}

			task.Status = enums.TaskFailed
			if accepted {
				task.Status = enums.TaskCompleted
			}

			return nil, nil
		})
	if err != nil {
		return err
	}

	return s.settle(c, "task acceptance")
}

func (s *v1) Report(agentID uuid.UUID, status enums.TaskStatus, docs []*core_itf.HandoverDoc) error {
	if status != enums.TaskCompleted && status != enums.TaskFailed {
		return custom_error.Critical("agent %v reported unsupported status %s", agentID, status)
	}

	if len(docs) == 0 {
		return custom_error.Critical("report from agent %v is missing a handover doc", agentID)
	}

	for _, doc := range docs {
		if doc == nil {
			return custom_error.Critical("report from agent %v has an empty handover doc", agentID)
		}
	}

	handle, found := s.handleOf(agentID)
	if !found {
		return custom_error.Critical("agent %v is not assigned to any task", agentID)
	}

	reportID, err := uuid.NewV7()
	if err != nil {
		return custom_error.Critical("cannot create uuid: %v", err)
	}

	report := &core_itf.TaskReport{
		TaskID:       handle.TaskID,
		Status:       status,
		HandoverDocs: docs,
		ContextUsage: s.readContextUsage(agentID),
	}

	c, err := s.editTask(handle.TaskID, agentID, enums.SessionTaskReported,
		func(session *sessionMetadata, task *input_itf.TaskEntity) (func(), error) {
			if task.Status != enums.TaskProcessing {
				return nil, custom_error.Critical("task %v is %s and cannot be reported", task.ID, task.Status)
			}

			for _, doc := range docs {
				doc.Task = task.Name
			}

			task.Status = status
			if status == enums.TaskCompleted && task.ManualAcceptRequired {
				task.Status = enums.TaskAwaitingAccept
			}

			task.LastReportID = reportID

			if !status.Removable() {
				task.RetryCount += 1
			}

			if status == enums.TaskFailed {
				session.info.TotalRetry += 1
			}

			return nil, nil
		})
	if err != nil {
		return err
	}

	c.report = &input_itf.TaskReportEntity{
		ID:            reportID,
		TaskID:        handle.TaskID,
		AgentID:       agentID,
		AttemptStatus: status,
		HandoverDocs:  handoverDocEntities(docs),
		ContextUsage:  report.ContextUsage,
		StartedAt:     handle.AssignedAt,
		CompletedAt:   helpers.NewUTC(),
		CreatedAt:     helpers.NewUTC(),
		UpdatedAt:     helpers.NewUTC(),
	}

	if err := s.persist(c, "task report"); err != nil {
		return err
	}

	s.closeHandle(c.session, agentID, report)

	return firstErr(s.publish(c.progress), s.drainIfDone(c.session))
}

// Runs after the database write, so a failed write leaves the agent still holding its task.
func (s *v1) closeHandle(session *sessionMetadata, agentID uuid.UUID, report *core_itf.TaskReport) {
	s.locker.Lock()
	defer s.locker.Unlock()

	delete(session.agentIDToHandle, agentID)
	session.keepReport(report)
}

func handoverDocEntities(docs []*core_itf.HandoverDoc) []*input_itf.HandoverDocEntity {
	entities := make([]*input_itf.HandoverDocEntity, 0, len(docs))

	for _, doc := range docs {
		entities = append(entities, &input_itf.HandoverDocEntity{
			Task:              doc.Task,
			TLDR:              doc.TLDR,
			Outcome:           doc.Outcome,
			Blockers:          doc.Blockers,
			ApprovedDecisions: doc.ApprovedDecisions,
			RejectedDecisions: doc.RejectedDecisions,
			CurrentBehaviors:  doc.CurrentBehaviors,
			ChangedBehaviors:  doc.ChangedBehaviors,
			MustAvoid:         doc.MustAvoid,
			Nuances:           doc.Nuances,
			KnownGaps:         doc.KnownGaps,
		})
	}

	return entities
}

func (s *v1) Cancel(sessionID uuid.UUID) ([]uuid.UUID, error) {
	changes, agentIDs, err := s.cancelSession(sessionID)
	if err != nil {
		return nil, err
	}

	if err := s.save(changes...); err != nil {
		s.locker.Lock()
		for _, c := range changes {
			c.rollback()
		}
		s.locker.Unlock()

		return nil, custom_error.Critical("cannot save task cancel: %v", err)
	}

	session, found := s.session(sessionID)
	if found {
		s.retireTokens(session, agentIDs)
	}

	publishErrs := make([]error, 0, len(changes)+1)

	for _, c := range changes {
		publishErrs = append(publishErrs, s.publish(c.progress))
	}

	if found {
		publishErrs = append(publishErrs, s.drainIfDone(session))
	}

	return agentIDs, firstErr(publishErrs...)
}

// A cancelled agent never reports, so what it already spent is read before the drop is
// published: the coordinator answers that event by killing the agent, and a killed
// agent has no usage left to read.
func (s *v1) retireTokens(session *sessionMetadata, agentIDs []uuid.UUID) {
	for _, agentID := range agentIDs {
		s.creditTokens(session, s.readContextUsage(agentID))
	}
}

func (s *v1) creditTokens(session *sessionMetadata, usage *input_itf.ContextUsage) {
	if usage == nil {
		return
	}

	s.locker.Lock()
	session.tokensBilled += usage.Billed
	s.locker.Unlock()
}

func (s *v1) cancelSession(sessionID uuid.UUID) ([]*change, []uuid.UUID, error) {
	s.locker.Lock()
	defer s.locker.Unlock()

	session, found := s.sessions[sessionID]
	if !found {
		return nil, nil, custom_error.Critical("session %v not found to cancel", sessionID)
	}

	infoSnapshot := *session.info
	changes := []*change{}

	for taskID, task := range session.taskIDToTask {
		if !task.Status.Cancellable() {
			continue
		}

		kind := enums.SessionTaskStatusChanged
		agentID := uuid.Nil
		handle, taken := session.findHandle(taskID)

		if taken {
			agentID = handle.AgentID
			kind = enums.SessionTaskDropped
		}

		prevTask := *task
		task.Status = enums.TaskCancelled
		task.UpdatedAt = helpers.NewUTC()
		taskSnapshot := *task

		changes = append(changes, &change{
			session: session,
			info:    &infoSnapshot,
			task:    &taskSnapshot,
			progress: &core_itf.SessionProgress{
				SessionID: sessionID,
				TaskID:    taskID,
				AgentID:   agentID,
				Event:     kind,
			},
			rollback: func() {
				*task = prevTask

				if taken {
					session.agentIDToHandle[agentID] = handle
				}
			},
		})
	}

	agentIDs := make([]uuid.UUID, 0, len(session.agentIDToHandle))
	for agentID := range session.agentIDToHandle {
		agentIDs = append(agentIDs, agentID)
	}

	sort.Slice(agentIDs, func(i, j int) bool { return agentIDs[i].String() < agentIDs[j].String() })

	session.agentIDToHandle = map[uuid.UUID]*AgentHandle{}
	session.cancelled = true

	return changes, agentIDs, nil
}

func (s *v1) resume(session *sessionMetadata) {
	s.locker.Lock()
	defer s.locker.Unlock()

	session.cancelled = false
}

func (s *v1) drainIfDone(session *sessionMetadata) error {
	c, drained := s.stageDrain(session)
	if !drained {
		return nil
	}

	return s.commit(c, "session drained")
}

func (s *v1) stageDrain(session *sessionMetadata) (*change, bool) {
	s.locker.Lock()
	defer s.locker.Unlock()

	if !session.info.CompletedAt.IsZero() || len(session.taskIDToTask) == 0 {
		return nil, false
	}

	for _, task := range session.taskIDToTask {
		if !task.Status.Removable() {
			return nil, false
		}
	}

	prevInfo := *session.info
	now := helpers.NewUTC()
	session.info.CompletedAt = now
	session.info.UpdatedAt = now
	infoSnapshot := *session.info

	return &change{
		session: session,
		info:    &infoSnapshot,
		progress: &core_itf.SessionProgress{
			SessionID: infoSnapshot.ID,
			Event:     enums.SessionDrained,
		},
		rollback: func() { *session.info = prevInfo },
	}, true
}

func (s *v1) ReadyTasks(sessionID uuid.UUID) ([]*core_itf.TaskSpec, error) {
	s.locker.Lock()

	session, found := s.sessions[sessionID]
	if !found {
		s.locker.Unlock()
		return nil, custom_error.Critical("session %v not found", sessionID)
	}

	specs := []*core_itf.TaskSpec{}

	if !session.cancelled {
		for taskID, task := range session.taskIDToTask {
			if _, taken := session.findHandle(taskID); taken {
				continue
			}

			if !session.isReady(task) {
				continue
			}

			specs = append(specs, taskSpec(taskID, task))
		}
	}

	s.locker.Unlock()

	sort.Slice(specs, func(i, j int) bool {
		return specs[i].TaskID.String() < specs[j].TaskID.String()
	})

	return specs, nil
}

func taskSpec(taskID uuid.UUID, t *input_itf.TaskEntity) *core_itf.TaskSpec {
	return &core_itf.TaskSpec{
		TaskID:               taskID,
		Name:                 t.Name,
		ManualAcceptRequired: t.ManualAcceptRequired,
		ExtraGuidance:        t.ExtraGuidance,
		DependsOn:            t.DependsOnTaskIDs,
		AgentSpecs: &core_itf.AgentRequest{
			Name:          t.PreferredModel,
			ThinkingLevel: t.ThinkingLevel,
			SystemPrompts: t.SystemPrompts,
		},
	}
}

func (s *v1) Status(id uuid.UUID) (*core_itf.SessionStatus, error) {
	s.locker.Lock()

	session, found := s.sessions[id]
	if !found {
		s.locker.Unlock()
		return nil, custom_error.Critical("session %v not found", id)
	}

	taskToAgent := map[uuid.UUID]uuid.UUID{}
	for agentID, handle := range session.agentIDToHandle {
		taskToAgent[handle.TaskID] = agentID
	}

	tasks := map[uuid.UUID]*core_itf.TaskReport{}

	for taskID, task := range session.taskIDToTask {
		report := &core_itf.TaskReport{
			TaskID:       taskID,
			Status:       task.Status,
			HandoverDocs: []*core_itf.HandoverDoc{},
		}

		if reported, found := session.taskIDToReport[taskID]; found {
			report.HandoverDocs = append(report.HandoverDocs, reported.HandoverDocs...)
			report.ContextUsage = reported.ContextUsage
		}

		tasks[taskID] = report
	}

	status := &core_itf.SessionStatus{
		ID:             id,
		Status:         sessionStatus(session.info),
		WorkingDirPath: session.info.WorkingDirPath,
		ContextDirPath: session.info.ContextDirPath,
		Tasks:          tasks,
		TokensBilled:   session.tokensBilled,
		StartedAt:      session.info.StartedAt,
		CompletedAt:    session.info.CompletedAt,
	}

	s.locker.Unlock()

	// Outside the lock: reading usage takes the same lock to reach the reader.
	for taskID, agentID := range taskToAgent {
		task, found := status.Tasks[taskID]
		if !found {
			continue
		}

		task.AgentID = agentID

		if usage := s.readContextUsage(agentID); usage != nil {
			task.ContextUsage = usage
			status.TokensBilled += usage.Billed
		}

		task.Activity = s.readActivity(agentID)
	}

	return status, nil
}

func (s *v1) Execute(sessionID uuid.UUID) (<-chan *core_itf.SessionProgress, error) {
	session, found := s.session(sessionID)
	if !found {
		return nil, custom_error.Critical("session %v not found", sessionID)
	}

	s.retireRun(sessionID, session)

	events := enums.SessionEvents()
	streams := make([]<-chan any, 0, len(events))

	for i, event := range events {
		stream, err := s.mq.Subscribe(sessionID, event)
		if err != nil {
			for _, subscribed := range events[:i] {
				s.mq.Unsubscribe(sessionID, subscribed)
			}

			return nil, custom_error.Critical("cannot watch session %v: %v", sessionID, err)
		}

		streams = append(streams, stream)
	}

	s.resume(session)

	running := s.beginRun(session)

	out := make(chan *core_itf.SessionProgress, progressBufferSize)

	wg := sync.WaitGroup{}

	for _, stream := range streams {
		wg.Add(1)

		go func(stream <-chan any) {
			defer wg.Done()

			s.forward(stream, out, running)
		}(stream)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out, nil
}

func (s *v1) retireRun(sessionID uuid.UUID, session *sessionMetadata) {
	s.locker.Lock()
	running := session.run
	session.run = nil
	s.locker.Unlock()

	if running == nil {
		return
	}

	close(running)

	for _, event := range enums.SessionEvents() {
		s.mq.Unsubscribe(sessionID, event)
	}
}

func (s *v1) beginRun(session *sessionMetadata) <-chan struct{} {
	s.locker.Lock()
	defer s.locker.Unlock()

	session.run = make(chan struct{})

	return session.run
}

func (s *v1) forward(stream <-chan any, out chan<- *core_itf.SessionProgress, running <-chan struct{}) {
	for {
		select {
		case <-s.stop:
			return
		case <-running:
			return
		case data, open := <-stream:
			if !open {
				return
			}

			progress, ok := data.(*core_itf.SessionProgress)
			if !ok {
				continue
			}

			select {
			case out <- progress:
			case <-running:
				return
			case <-s.stop:
				return
			}
		}
	}
}

func (s *v1) HeartBeat(agentID uuid.UUID) error {
	s.locker.Lock()
	defer s.locker.Unlock()

	handle, found := s.findAgent(agentID)
	if !found {
		return custom_error.Critical("agent %v is not assigned to any task", agentID)
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
				_ = s.dropTask(agentID, deadline)
			}
		}
	}
}

func (s *v1) staleAgents(deadline int64) []uuid.UUID {
	s.locker.Lock()
	defer s.locker.Unlock()

	stale := []uuid.UUID{}

	for _, session := range s.sessions {
		for agentID, handle := range session.agentIDToHandle {
			if handle.LastHeartBeat <= deadline {
				stale = append(stale, agentID)
			}
		}
	}

	return stale
}

// The agent may have reported since it was found stale, so the deadline is checked
// again under the lock.
func (s *v1) dropTask(agentID uuid.UUID, deadline int64) error {
	handle, alive := s.handleOf(agentID)
	if !alive {
		return nil
	}

	spent := s.readContextUsage(agentID)

	c, err := s.editTask(handle.TaskID, agentID, enums.SessionTaskDropped,
		func(session *sessionMetadata, task *input_itf.TaskEntity) (func(), error) {
			current, found := session.agentIDToHandle[agentID]
			if !found || current.LastHeartBeat > deadline || task.Status != enums.TaskProcessing {
				return nil, errNoChange
			}

			task.Status = enums.TaskCancelled
			delete(session.agentIDToHandle, agentID)

			return func() { session.agentIDToHandle[agentID] = current }, nil
		})
	if err != nil {
		return custom_error.Critical("cannot drop task of stale agent %v: %v", agentID, err)
	}

	if c == nil {
		return nil
	}

	if err := s.persist(c, "task drop"); err != nil {
		return err
	}

	s.creditTokens(c.session, spent)

	return firstErr(s.publish(c.progress), s.drainIfDone(c.session))
}

func (s *v1) session(sessionID uuid.UUID) (*sessionMetadata, bool) {
	s.locker.Lock()
	defer s.locker.Unlock()

	session, found := s.sessions[sessionID]

	return session, found
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

// findAgent and findTask read shared state; the caller holds the lock.
func (s *v1) findAgent(agentID uuid.UUID) (*AgentHandle, bool) {
	for _, session := range s.sessions {
		if handle, found := session.agentIDToHandle[agentID]; found {
			return handle, true
		}
	}

	return nil, false
}

func (s *v1) findTask(taskID uuid.UUID) (*sessionMetadata, *input_itf.TaskEntity, bool) {
	for _, session := range s.sessions {
		if task, found := session.taskIDToTask[taskID]; found {
			return session, task, true
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

func (m *sessionMetadata) isReady(t *input_itf.TaskEntity) bool {
	for _, dep := range t.DependsOnTaskIDs {
		depTask, found := m.taskIDToTask[dep]
		if !found || depTask.Status != enums.TaskCompleted {
			return false
		}
	}

	if t.Status == enums.TaskNotTaken {
		return true
	}

	return t.AutoRetry && t.Status.Retryable() && t.RetryCount < maxAutoRetry
}

func (m *sessionMetadata) findHandle(taskID uuid.UUID) (*AgentHandle, bool) {
	for _, handle := range m.agentIDToHandle {
		if handle.TaskID == taskID {
			return handle, true
		}
	}

	return nil, false
}

func (m *sessionMetadata) keepReport(report *core_itf.TaskReport) {
	kept, found := m.taskIDToReport[report.TaskID]
	if !found {
		kept = &core_itf.TaskReport{
			TaskID:       report.TaskID,
			HandoverDocs: []*core_itf.HandoverDoc{},
		}

		m.taskIDToReport[report.TaskID] = kept
	}

	kept.Status = report.Status
	kept.HandoverDocs = append(kept.HandoverDocs, report.HandoverDocs...)

	if report.ContextUsage != nil {
		kept.ContextUsage = report.ContextUsage
		m.tokensBilled += report.ContextUsage.Billed
	}
}

func sessionStatus(info *input_itf.SessionEntity) enums.SessionStatus {
	switch {
	case !info.CompletedAt.IsZero():
		return enums.SessionCompleted
	case !info.StartedAt.IsZero():
		return enums.SessionProcessing
	default:
		return enums.SessionInit
	}
}

func makeDir(name, path, fallback string) (string, error) {
	trimmed := strings.TrimSpace(path)

	if trimmed == "" {
		return fallback, nil
	}

	cleaned, err := absPath(name, trimmed)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(cleaned, 0o755); err != nil {
		return "", dirError(name, cleaned, "created", err)
	}

	return existingDir(name, cleaned)
}

func existingDir(name, path string) (string, error) {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return "", custom_error.Critical("session %s is empty", name)
	}

	cleaned, err := absPath(name, trimmed)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(cleaned)
	switch {
	case os.IsNotExist(err):
		return "", custom_error.Critical("session %s %q does not exist", name, cleaned)
	case err != nil:
		return "", dirError(name, cleaned, "accessed", err)
	case !info.IsDir():
		return "", custom_error.Critical("session %s %q is not a directory", name, cleaned)
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
		return "", custom_error.Critical("session %s %q must be an absolute path", name, path)
	}

	return filepath.Clean(path), nil
}

func dirError(name, path, verb string, err error) error {
	if os.IsPermission(err) {
		return custom_error.Critical("session %s %q cannot be %s: permission denied", name, path, verb)
	}

	return custom_error.Critical("session %s %q cannot be %s: %v", name, path, verb, err)
}
