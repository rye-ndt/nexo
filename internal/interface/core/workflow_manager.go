package core_itf

import (
	"time"

	"hexago/internal/helpers/enums"
	input_itf "hexago/internal/interface/input"

	"github.com/google/uuid"
)

type AddStep struct {
	Name            string
	Effort          enums.Effort
	AutoRetry       bool
	PauseForReview  bool
	ExtraGuidance   string
	OutputStructure string
	DependsOn       []uuid.UUID
	AgentSpecs      *AgentRequest
}

type StepSpec struct {
	StepID         uuid.UUID
	Name           string
	PauseForReview bool
	ExtraGuidance  string
	DependsOn      []uuid.UUID
	AgentSpecs     *AgentRequest
}

type Handoff struct {
	Step              string
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

// ContextUsage is the window of the attempt being looked at, which is what the step's
// ring draws. Spent is every token the step has ever cost, every attempt of it summed,
// which is what it is billed for — the two differ as soon as a step is retried.
type StepResult struct {
	StepID       uuid.UUID
	Name         string
	AgentID      uuid.UUID
	Effort       enums.Effort
	Status       enums.StepStatus
	Handoffs     []*Handoff
	ContextUsage *input_itf.ContextUsage
	Spent        *input_itf.ContextUsage
	Activity     []input_itf.Activity
}

type WorkflowProgress struct {
	WorkflowID uuid.UUID
	StepID     uuid.UUID
	AgentID    uuid.UUID
	Event      enums.WorkflowEvent
}

// TokensBilled, TokensInput and TokensCached are the sum of what every step in Steps
// has spent, so the workflow total and its steps can never tell different stories.
type WorkflowStatus struct {
	ID             uuid.UUID
	Status         enums.WorkflowStatus
	ProjectDirPath string
	Steps          map[uuid.UUID]*StepResult
	TokensBilled   int
	TokensInput    int
	TokensCached   int
	StartedAt      time.Time
	CompletedAt    time.Time
}

type InitWorkflow struct {
	ProjectDirPath string
}

type StepReporter interface {
	Report(agentID uuid.UUID, status enums.StepStatus, docs []*Handoff) error
}

// LiveAgentReader reads what an agent has spent of its window, and what it is doing, while it
// is still alive. AgentManager satisfies it; the workflow manager takes it separately because
// the two are wired in a cycle — the agent manager is built on top of the MCP gateway that
// reports here.
type LiveAgentReader interface {
	ContextUsage(agentID uuid.UUID) (*input_itf.ContextUsage, error)
	Activity(agentID uuid.UUID) ([]input_itf.Activity, error)
}

type WorkflowManager interface {
	StepReporter
	TrackLiveAgents(reader LiveAgentReader)
	NewWorkflow(p *InitWorkflow) (uuid.UUID, error)
	AddStep(workflow uuid.UUID, step *AddStep) (uuid.UUID, error)
	ReadySteps(workflow uuid.UUID) ([]*StepSpec, error)
	Assign(stepID, agentID uuid.UUID) error
	Execute(workflow uuid.UUID) (<-chan *WorkflowProgress, error)
	RetryStep(stepID uuid.UUID) error
	RewindTo(stepID uuid.UUID) error
	AnswerReview(stepID uuid.UUID, accepted bool) error
	Cancel(workflow uuid.UUID) ([]uuid.UUID, error)
	Pause(workflow uuid.UUID) ([]uuid.UUID, error)
	Restore() error
	Status(id uuid.UUID) (*WorkflowStatus, error)
	HeartBeat(agentID uuid.UUID) error
	Stop()
}
