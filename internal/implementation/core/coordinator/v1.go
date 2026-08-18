package coordinator

import (
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"hexago/internal/helpers"
	"hexago/internal/helpers/custom_error"
	"hexago/internal/helpers/enums"
	"hexago/internal/helpers/prompts"
	core_itf "hexago/internal/interface/core"
	input_itf "hexago/internal/interface/input"
	output_itf "hexago/internal/interface/output"

	"github.com/google/uuid"
)

const scheduleInterval = 15 * time.Second

type v1 struct {
	locker   sync.Mutex
	cfg      *input_itf.SessionConfig
	sessions core_itf.SessionManager
	agents   core_itf.AgentManager
	history  input_itf.WorkspaceHistory
	logger   output_itf.Logger
	running  map[uuid.UUID]chan struct{}
	watchers map[uuid.UUID]chan struct{}
	stop     chan struct{}
	stopOnce sync.Once
}

type workspace struct {
	dir      string
	excludes []string
}

func InitV1(
	cfg *input_itf.SessionConfig,
	sessions core_itf.SessionManager,
	agents core_itf.AgentManager,
	history input_itf.WorkspaceHistory,
	logger output_itf.Logger,
) (core_itf.Coordinator, error) {
	if err := helpers.ValidateStruct(cfg); err != nil {
		return nil, custom_error.Critical("invalid coordinator config: %v", err)
	}

	return &v1{
		cfg:      cfg,
		sessions: sessions,
		agents:   agents,
		history:  history,
		logger:   logger,
		running:  map[uuid.UUID]chan struct{}{},
		watchers: map[uuid.UUID]chan struct{}{},
		stop:     make(chan struct{}),
	}, nil
}

func (c *v1) Run(session uuid.UUID) error {
	halt, err := c.startRunning(session)
	if err != nil {
		return err
	}

	progress, err := c.sessions.Execute(session)
	if err != nil {
		c.forget(session, halt)
		return err
	}

	go c.runSession(session, halt, progress)

	return nil
}

func (c *v1) Cancel(session uuid.UUID) error {
	return c.stopRun(session, c.sessions.Cancel)
}

func (c *v1) Pause(session uuid.UUID) error {
	return c.stopRun(session, c.sessions.Pause)
}

func (c *v1) stopRun(session uuid.UUID, halt func(uuid.UUID) ([]uuid.UUID, error)) error {
	c.halt(session)

	agentIDs, err := halt(session)

	for _, agentID := range agentIDs {
		c.releaseAgent(agentID)
	}

	return err
}

func (c *v1) RevertTo(session, taskID uuid.UUID) error {
	status, err := c.sessions.Status(session)
	if err != nil {
		return err
	}

	if err := c.Cancel(session); err != nil {
		return err
	}

	if err := c.history.RestoreTo(session, taskID, status.WorkingDirPath); err != nil {
		return err
	}

	return c.sessions.RewindTo(taskID)
}

func (c *v1) Stop() {
	c.stopOnce.Do(func() { close(c.stop) })
}

func (c *v1) startRunning(session uuid.UUID) (chan struct{}, error) {
	c.locker.Lock()
	defer c.locker.Unlock()

	if _, found := c.running[session]; found {
		return nil, custom_error.Critical("session %v is already running", session)
	}

	halt := make(chan struct{})
	c.running[session] = halt

	return halt, nil
}

func (c *v1) halt(session uuid.UUID) {
	c.locker.Lock()
	running := c.running[session]
	delete(c.running, session)
	c.locker.Unlock()

	if running != nil {
		close(running)
	}
}

// Only forgets the run the caller started: a relaunched session must not be dropped
// by the loop it replaced.
func (c *v1) forget(session uuid.UUID, halt <-chan struct{}) {
	c.locker.Lock()
	defer c.locker.Unlock()

	if c.running[session] == halt {
		delete(c.running, session)
	}
}

func (c *v1) baseline(session uuid.UUID) *workspace {
	status, err := c.sessions.Status(session)
	if err != nil {
		c.logger.Error("workspace snapshot", "session", session, "err", err)

		return nil
	}

	space := &workspace{dir: status.WorkingDirPath, excludes: contextExcludes(status)}
	c.snapshot(session, uuid.Nil, space)

	return space
}

func (c *v1) snapshot(session, taskID uuid.UUID, space *workspace) {
	if space == nil {
		return
	}

	if err := c.history.Commit(session, taskID, space.dir, space.excludes); err != nil {
		c.logger.Error("workspace snapshot", "session", session, "task", taskID, "err", err)
	}
}

func contextExcludes(status *core_itf.SessionStatus) []string {
	base, err := filepath.Rel(status.WorkingDirPath, status.ContextDirPath)
	if err != nil || base == ".." || strings.HasPrefix(base, ".."+string(filepath.Separator)) {
		return nil
	}

	skeleton := prompts.ContextSkeleton()
	patterns := make([]string, 0, len(skeleton))
	seen := make(map[string]struct{}, len(skeleton))

	for path := range skeleton {
		top, _, _ := strings.Cut(filepath.ToSlash(path), "/")
		if top == "" {
			continue
		}

		pattern := "/" + filepath.ToSlash(filepath.Join(base, top))
		if _, found := seen[pattern]; found {
			continue
		}

		seen[pattern] = struct{}{}
		patterns = append(patterns, pattern)
	}

	sort.Strings(patterns)

	return patterns
}

func (c *v1) runSession(
	session uuid.UUID,
	halt <-chan struct{},
	progress <-chan *core_itf.SessionProgress,
) {
	defer c.forget(session, halt)

	space := c.baseline(session)

	ticker := time.NewTicker(scheduleInterval)
	defer ticker.Stop()

	c.schedule(session)

	for {
		select {
		case <-c.stop:
			return
		case <-halt:
			return
		case <-ticker.C:
			c.schedule(session)
		case event, open := <-progress:
			if !open {
				return
			}

			switch event.Event {
			case enums.SessionTaskReported, enums.SessionTaskDropped:
				c.releaseAgent(event.AgentID)
				c.snapshot(session, event.TaskID, space)
			}

			c.schedule(session)
		}
	}
}

func (c *v1) schedule(session uuid.UUID) {
	specs, err := c.sessions.ReadyTasks(session)
	if err != nil || len(specs) == 0 {
		return
	}

	status, err := c.sessions.Status(session)
	if err != nil {
		return
	}

	for _, spec := range specs {
		if spec.AgentSpecs == nil {
			continue
		}

		if stop := c.start(spec, status); stop {
			return
		}
	}
}

func (c *v1) start(spec *core_itf.TaskSpec, status *core_itf.SessionStatus) (outOfCapacity bool) {
	agent, err := c.agents.RequestInstance(&core_itf.AgentRequest{
		Name:          spec.AgentSpecs.Name,
		ThinkingLevel: spec.AgentSpecs.ThinkingLevel,
		SystemPrompts: spec.AgentSpecs.SystemPrompts,
		WorkingDir:    status.WorkingDirPath,
	})
	if err != nil {
		return true
	}

	if err := c.sessions.Assign(spec.TaskID, agent.ID); err != nil {
		_ = c.agents.Kill(agent.ID)
		return false
	}

	if err := c.agents.Send(agent.ID, buildPrompt(spec, status)); err != nil {
		c.reportDeath(agent.ID, "The assistant working on this step stopped running before it could start, so nothing was done.", err)
		return false
	}

	c.watch(agent.ID)

	return false
}

func (c *v1) watch(agentID uuid.UUID) {
	watcher := make(chan struct{})

	c.locker.Lock()
	c.watchers[agentID] = watcher
	c.locker.Unlock()

	go func() {
		ticker := time.NewTicker(c.cfg.AgentHeartbeatInterval)
		defer ticker.Stop()

		for {
			select {
			case <-c.stop:
				return
			case <-watcher:
				return
			case <-ticker.C:
				_ = c.sessions.HeartBeat(agentID)

				if err := c.agents.HeartBeat(agentID); err != nil {
					c.forgetWatcher(agentID)
					c.reportDeath(agentID, "The assistant working on this step stopped responding partway through, so the work is unfinished.", err)

					return
				}
			}
		}
	}()
}

func (c *v1) reportDeath(agentID uuid.UUID, tldr string, cause error) {
	_ = c.sessions.Report(agentID, enums.TaskFailed, []*core_itf.HandoverDoc{{
		TLDR:    tldr,
		Outcome: "agent died before reporting: " + cause.Error(),
	}})

	_ = c.agents.Kill(agentID)
}

func (c *v1) releaseAgent(agentID uuid.UUID) {
	if agentID == uuid.Nil {
		return
	}

	if watcher := c.forgetWatcher(agentID); watcher != nil {
		close(watcher)
	}

	_ = c.agents.Kill(agentID)
}

func (c *v1) forgetWatcher(agentID uuid.UUID) chan struct{} {
	c.locker.Lock()
	defer c.locker.Unlock()

	watcher := c.watchers[agentID]
	delete(c.watchers, agentID)

	return watcher
}
