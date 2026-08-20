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
	locker    sync.Mutex
	cfg       *input_itf.WorkflowConfig
	workflows core_itf.WorkflowManager
	agents    core_itf.AgentManager
	history   input_itf.WorkspaceHistory
	logger    output_itf.Logger
	running   map[uuid.UUID]chan struct{}
	watchers  map[uuid.UUID]chan struct{}
	stop      chan struct{}
	stopOnce  sync.Once
}

type workspace struct {
	dir      string
	excludes []string
}

func InitV1(
	cfg *input_itf.WorkflowConfig,
	workflows core_itf.WorkflowManager,
	agents core_itf.AgentManager,
	history input_itf.WorkspaceHistory,
	logger output_itf.Logger,
) (core_itf.Coordinator, error) {
	if err := helpers.ValidateStruct(cfg); err != nil {
		return nil, custom_error.Critical("invalid coordinator config: %v", err)
	}

	return &v1{
		cfg:       cfg,
		workflows: workflows,
		agents:    agents,
		history:   history,
		logger:    logger,
		running:   map[uuid.UUID]chan struct{}{},
		watchers:  map[uuid.UUID]chan struct{}{},
		stop:      make(chan struct{}),
	}, nil
}

func (c *v1) Run(workflow uuid.UUID) error {
	halt, err := c.startRunning(workflow)
	if err != nil {
		return err
	}

	progress, err := c.workflows.Execute(workflow)
	if err != nil {
		c.forget(workflow, halt)
		return err
	}

	go c.runWorkflow(workflow, halt, progress)

	return nil
}

func (c *v1) Cancel(workflow uuid.UUID) error {
	return c.stopRun(workflow, c.workflows.Cancel)
}

func (c *v1) Pause(workflow uuid.UUID) error {
	return c.stopRun(workflow, c.workflows.Pause)
}

func (c *v1) stopRun(workflow uuid.UUID, halt func(uuid.UUID) ([]uuid.UUID, error)) error {
	c.halt(workflow)

	agentIDs, err := halt(workflow)

	for _, agentID := range agentIDs {
		c.releaseAgent(agentID)
	}

	return err
}

func (c *v1) RevertTo(workflow, stepID uuid.UUID) error {
	status, err := c.workflows.Status(workflow)
	if err != nil {
		return err
	}

	if err := c.Cancel(workflow); err != nil {
		return err
	}

	if err := c.history.RestoreTo(workflow, stepID, status.ProjectDirPath); err != nil {
		return err
	}

	return c.workflows.RewindTo(stepID)
}

func (c *v1) Stop() {
	c.stopOnce.Do(func() { close(c.stop) })
}

func (c *v1) startRunning(workflow uuid.UUID) (chan struct{}, error) {
	c.locker.Lock()
	defer c.locker.Unlock()

	if _, found := c.running[workflow]; found {
		return nil, custom_error.Critical("workflow %v is already running", workflow)
	}

	halt := make(chan struct{})
	c.running[workflow] = halt

	return halt, nil
}

func (c *v1) halt(workflow uuid.UUID) {
	c.locker.Lock()
	running := c.running[workflow]
	delete(c.running, workflow)
	c.locker.Unlock()

	if running != nil {
		close(running)
	}
}

// Only forgets the run the caller started: a relaunched workflow must not be dropped
// by the loop it replaced.
func (c *v1) forget(workflow uuid.UUID, halt <-chan struct{}) {
	c.locker.Lock()
	defer c.locker.Unlock()

	if c.running[workflow] == halt {
		delete(c.running, workflow)
	}
}

func (c *v1) baseline(workflow uuid.UUID) *workspace {
	status, err := c.workflows.Status(workflow)
	if err != nil {
		c.logger.Error("workspace snapshot", "workflow", workflow, "err", err)

		return nil
	}

	space := &workspace{dir: status.ProjectDirPath, excludes: contextExcludes(status.ProjectDirPath)}
	c.snapshot(workflow, uuid.Nil, space)

	return space
}

func (c *v1) snapshot(workflow, stepID uuid.UUID, space *workspace) {
	if space == nil {
		return
	}

	if err := c.history.Commit(workflow, stepID, space.dir, space.excludes); err != nil {
		c.logger.Error("workspace snapshot", "workflow", workflow, "step", stepID, "err", err)
	}
}

func contextExcludes(projectDir string) []string {
	base, err := filepath.Rel(projectDir, helpers.KnowledgeDir(projectDir))
	if err != nil {
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

func (c *v1) runWorkflow(
	workflow uuid.UUID,
	halt <-chan struct{},
	progress <-chan *core_itf.WorkflowProgress,
) {
	defer c.forget(workflow, halt)

	space := c.baseline(workflow)

	ticker := time.NewTicker(scheduleInterval)
	defer ticker.Stop()

	c.schedule(workflow)

	for {
		select {
		case <-c.stop:
			return
		case <-halt:
			return
		case <-ticker.C:
			c.schedule(workflow)
		case event, open := <-progress:
			if !open {
				return
			}

			switch event.Event {
			case enums.WorkflowStepResulted, enums.WorkflowStepDropped:
				c.releaseAgent(event.AgentID)
				c.snapshot(workflow, event.StepID, space)
			}

			c.schedule(workflow)
		}
	}
}

func (c *v1) schedule(workflow uuid.UUID) {
	specs, err := c.workflows.ReadySteps(workflow)
	if err != nil || len(specs) == 0 {
		return
	}

	status, err := c.workflows.Status(workflow)
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

func (c *v1) start(spec *core_itf.StepSpec, status *core_itf.WorkflowStatus) (outOfCapacity bool) {
	agent, err := c.agents.RequestInstance(&core_itf.AgentRequest{
		Name:          spec.AgentSpecs.Name,
		ThinkingLevel: spec.AgentSpecs.ThinkingLevel,
		Instructions:  spec.AgentSpecs.Instructions,
		ProjectDir:    status.ProjectDirPath,
	})
	if err != nil {
		return true
	}

	if err := c.workflows.Assign(spec.StepID, agent.ID); err != nil {
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
				_ = c.workflows.HeartBeat(agentID)

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
	_ = c.workflows.Report(agentID, enums.StepFailed, []*core_itf.Handoff{{
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
