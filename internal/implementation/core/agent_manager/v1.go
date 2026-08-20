package agent_manager

import (
	"sync"
	"time"

	"hexago/internal/helpers"
	"hexago/internal/helpers/custom_error"
	"hexago/internal/helpers/enums"
	core_itf "hexago/internal/interface/core"
	input_itf "hexago/internal/interface/input"

	"github.com/google/uuid"
)

type connectivity struct {
	CheckedAt time.Time
	Err       error
}

type instance struct {
	agent   *core_itf.Agent
	harness input_itf.AgentHarness
}

type agentManagerV1 struct {
	locker         sync.Mutex
	cfg            *input_itf.AgentManagerConfig
	httpCli        input_itf.HttpCli
	harnesses      map[enums.AgentHarness]input_itf.AgentHarness
	instances      map[uuid.UUID]*instance
	approvalBroker core_itf.ApprovalWaitReader
	online         connectivity
}

func InitV1(
	cfg *input_itf.AgentManagerConfig,
	httpCli input_itf.HttpCli,
	harnesses map[enums.AgentHarness]input_itf.AgentHarness,
	approvalBroker core_itf.ApprovalWaitReader,
) (core_itf.AgentManager, error) {
	if cfg == nil {
		return nil, custom_error.Critical("agent manager config not found")
	}

	return &agentManagerV1{
		cfg:            cfg,
		httpCli:        httpCli,
		harnesses:      harnesses,
		instances:      map[uuid.UUID]*instance{},
		approvalBroker: approvalBroker,
	}, nil
}

func (m *agentManagerV1) SupportedAgents() (map[enums.AgentHarness][]enums.ModelName, error) {
	supported := map[enums.AgentHarness][]enums.ModelName{}

	for name, harness := range m.harnesses {
		supported[name] = harness.SupportedModels()
	}

	return supported, nil
}

func (m *agentManagerV1) Admin(name enums.AgentHarness) (input_itf.AgentAdmin, error) {
	harness, found := m.harnesses[name]
	if !found {
		return nil, custom_error.Critical("no configured harness named %s", name)
	}

	return harness, nil
}

func (m *agentManagerV1) RequestInstance(specs *core_itf.AgentRequest) (*core_itf.Agent, error) {
	harness, found := m.harnessFor(specs.Name)
	if !found {
		return nil, custom_error.Critical("no harness client support model %s", specs.Name)
	}

	agentID, err := harness.Spawn(specs.Name, specs.ThinkingLevel, specs.Instructions, specs.ProjectDir)
	if err != nil {
		return nil, err
	}

	agent := &core_itf.Agent{
		ID:           agentID,
		HealthStatus: enums.Healthy,
	}

	clone := *agent

	m.locker.Lock()
	m.instances[agentID] = &instance{agent: agent, harness: harness}
	m.locker.Unlock()

	return &clone, nil
}

func (m *agentManagerV1) Send(agentID uuid.UUID, message string) error {
	harness, err := m.harnessOf(agentID)
	if err != nil {
		return err
	}

	return harness.Send(agentID.String(), message)
}

func (m *agentManagerV1) Listen(agentID uuid.UUID) (<-chan string, error) {
	harness, err := m.harnessOf(agentID)
	if err != nil {
		return nil, err
	}

	return harness.Listen(agentID.String())
}

func (m *agentManagerV1) ContextUsage(agentID uuid.UUID) (*input_itf.ContextUsage, error) {
	harness, err := m.harnessOf(agentID)
	if err != nil {
		return nil, err
	}

	return harness.Usage(agentID.String())
}

func (m *agentManagerV1) Activity(agentID uuid.UUID) ([]input_itf.Activity, error) {
	harness, err := m.harnessOf(agentID)
	if err != nil {
		return nil, err
	}

	return harness.Activity(agentID.String())
}

func (m *agentManagerV1) Kill(agentID uuid.UUID) error {
	harness, err := m.harnessOf(agentID)
	if err != nil {
		return err
	}

	killErr := harness.Kill(agentID.String())

	m.setHealth(agentID, enums.Terminated)

	return killErr
}

// An agent waiting on a human is not frozen, so that check comes before the timeout.
func (m *agentManagerV1) HeartBeat(agentID uuid.UUID) error {
	if agentID == uuid.Nil {
		return custom_error.Critical("heartbeat needs an agent id")
	}

	harness, err := m.liveHarnessOf(agentID)
	if err != nil {
		return err
	}

	lastOut, err := harness.Alive(agentID.String())
	if err != nil {
		m.setHealth(agentID, enums.Terminated)

		return custom_error.Critical("workflow of agent %v is gone: %v", agentID, err)
	}

	if m.approvalBroker.Awaiting(agentID) {
		m.setHealth(agentID, enums.AwaitingHuman)

		return nil
	}

	frozenFor := time.Since(lastOut)
	if frozenFor <= m.cfg.FreezeTimeout {
		m.setHealth(agentID, enums.Healthy)

		return nil
	}

	m.setHealth(agentID, enums.NotResponding)

	if err := m.checkConnectivity(); err != nil {
		return custom_error.Critical(
			"workflow of agent %v is silent for %v and the network is unreachable: %v", agentID, frozenFor, err,
		)
	}

	return custom_error.Critical("workflow of agent %v is frozen, silent for %v", agentID, frozenFor)
}

func (m *agentManagerV1) checkConnectivity() error {
	m.locker.Lock()
	fresh := time.Since(m.online.CheckedAt) <= m.cfg.ConnectivityCacheTTL
	cached := m.online.Err
	m.locker.Unlock()

	if fresh {
		return cached
	}

	err := m.httpCli.Reachable(m.cfg.ConnectivityProbeURL)

	m.locker.Lock()
	m.online = connectivity{CheckedAt: helpers.NewUTC(), Err: err}
	m.locker.Unlock()

	return err
}

func (m *agentManagerV1) harnessOf(agentID uuid.UUID) (input_itf.AgentHarness, error) {
	m.locker.Lock()
	defer m.locker.Unlock()

	live, found := m.instances[agentID]
	if !found {
		return nil, custom_error.Critical("agent %v not found", agentID)
	}

	return live.harness, nil
}

func (m *agentManagerV1) liveHarnessOf(agentID uuid.UUID) (input_itf.AgentHarness, error) {
	m.locker.Lock()
	defer m.locker.Unlock()

	live, found := m.instances[agentID]
	if !found {
		return nil, custom_error.Critical("agent %v not found", agentID)
	}

	if live.agent.HealthStatus == enums.Terminated {
		return nil, custom_error.Critical("agent %v is already terminated", agentID)
	}

	return live.harness, nil
}

func (m *agentManagerV1) setHealth(agentID uuid.UUID, status enums.AgentInstanceStatus) {
	m.locker.Lock()
	defer m.locker.Unlock()

	live, found := m.instances[agentID]
	if !found {
		return
	}

	live.agent.HealthStatus = status
}

func (m *agentManagerV1) harnessFor(model enums.ModelName) (input_itf.AgentHarness, bool) {
	for _, harness := range m.harnesses {
		if harness.Support(model) {
			return harness, true
		}
	}

	return nil, false
}
