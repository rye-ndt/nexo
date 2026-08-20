package core_itf

import (
	"hexago/internal/helpers/enums"
	input_itf "hexago/internal/interface/input"

	"github.com/google/uuid"
)

type AgentRequest struct {
	Name          enums.ModelName
	ThinkingLevel enums.ThinkingLevel
	Instructions  []string
	ProjectDir    string
}

type Agent struct {
	ID           uuid.UUID
	HealthStatus enums.AgentInstanceStatus
}

// ApprovalWaitReader tells whether an agent is parked on a question a human has not
// answered yet. ApprovalBroker satisfies it; the agent manager takes only this method
// because a silent agent waiting on a person is not a frozen one.
type ApprovalWaitReader interface {
	Awaiting(agentID uuid.UUID) bool
}

type AgentManager interface {
	SupportedAgents() (map[enums.AgentHarness][]enums.ModelName, error)
	Admin(name enums.AgentHarness) (input_itf.AgentAdmin, error)
	RequestInstance(specs *AgentRequest) (*Agent, error)
	Send(agentID uuid.UUID, message string) error
	Listen(agentID uuid.UUID) (<-chan string, error)
	ContextUsage(agentID uuid.UUID) (*input_itf.ContextUsage, error)
	Activity(agentID uuid.UUID) ([]input_itf.Activity, error)
	Kill(agentID uuid.UUID) error
	HeartBeat(agentID uuid.UUID) error
}
