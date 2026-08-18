package core_itf

import (
	"time"

	"hexago/internal/helpers/enums"
	input_itf "hexago/internal/interface/input"

	"github.com/google/uuid"
)

type AddTask struct {
	Name                 string
	TaskLevel            enums.TaskLevel
	AutoRetry            bool
	ManualAcceptRequired bool
	ExtraGuidance        string
	OutputStructure      string
	DependsOn            []uuid.UUID
	AgentSpecs           *AgentRequest
}

type TaskSpec struct {
	TaskID               uuid.UUID
	Name                 string
	ManualAcceptRequired bool
	ExtraGuidance        string
	DependsOn            []uuid.UUID
	AgentSpecs           *AgentRequest
}

type HandoverDoc struct {
	Task              string
	TLDR              string
	Outcome           string
	Blockers          map[string]string
	ApprovedDecisions map[string]string
	RejectedDecisions map[string]string
	CurrentBehaviors  map[string]string
	ChangedBehaviors  map[string]string
	MustAvoid         map[string]string
	Nuances           map[string]string
	KnownGaps         map[string]string
}

// ContextUsage is the window of the attempt being looked at, which is what the node's
// ring draws. Spent is every token the node has ever cost, every attempt of it summed,
// which is what it is billed for — the two differ as soon as a node is retried.
type TaskReport struct {
	TaskID       uuid.UUID
	Name         string
	AgentID      uuid.UUID
	TaskLevel    enums.TaskLevel
	Status       enums.TaskStatus
	HandoverDocs []*HandoverDoc
	ContextUsage *input_itf.ContextUsage
	Spent        *input_itf.ContextUsage
	Activity     []input_itf.Activity
}

type SessionProgress struct {
	SessionID uuid.UUID
	TaskID    uuid.UUID
	AgentID   uuid.UUID
	Event     enums.SessionEvent
}

// TokensBilled, TokensInput and TokensCached are the sum of what every task in Tasks
// has spent, so the session total and its nodes can never tell different stories.
type SessionStatus struct {
	ID             uuid.UUID
	Status         enums.SessionStatus
	WorkingDirPath string
	ContextDirPath string
	Tasks          map[uuid.UUID]*TaskReport
	TokensBilled   int
	TokensInput    int
	TokensCached   int
	StartedAt      time.Time
	CompletedAt    time.Time
}

type InitSession struct {
	WorkingDirPath string
	ContextDirPath string
}

type TaskReporter interface {
	Report(agentID uuid.UUID, status enums.TaskStatus, docs []*HandoverDoc) error
}

// LiveAgentReader reads what an agent has spent of its window, and what it is doing, while it
// is still alive. AgentManager satisfies it; the session manager takes it separately because
// the two are wired in a cycle — the agent manager is built on top of the MCP gateway that
// reports here.
type LiveAgentReader interface {
	ContextUsage(agentID uuid.UUID) (*input_itf.ContextUsage, error)
	Activity(agentID uuid.UUID) ([]input_itf.Activity, error)
}

type SessionManager interface {
	TaskReporter
	TrackLiveAgents(reader LiveAgentReader)
	NewSession(p *InitSession) (uuid.UUID, error)
	AddTask(session uuid.UUID, task *AddTask) (uuid.UUID, error)
	ReadyTasks(session uuid.UUID) ([]*TaskSpec, error)
	Assign(taskID, agentID uuid.UUID) error
	Execute(session uuid.UUID) (<-chan *SessionProgress, error)
	RetryTask(taskID uuid.UUID) error
	RewindTo(taskID uuid.UUID) error
	AnswerAcceptance(taskID uuid.UUID, accepted bool) error
	Cancel(session uuid.UUID) ([]uuid.UUID, error)
	Pause(session uuid.UUID) ([]uuid.UUID, error)
	Restore() error
	Status(id uuid.UUID) (*SessionStatus, error)
	HeartBeat(agentID uuid.UUID) error
	Stop()
}
