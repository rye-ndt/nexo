package input_itf

import (
	"hexago/internal/helpers/enums"
	"time"

	"github.com/google/uuid"
)

type WorkflowEntity struct {
	ID             uuid.UUID
	ProjectDirPath string
	StartedAt      time.Time
	CompletedAt    time.Time
	TotalStep      int
	TotalRetry     int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type HandoffEntity struct {
	Step              string            `json:"step_name"`
	TLDR              string            `json:"tldr"`
	Outcome           string            `json:"outcome"`
	Blockers          map[string]string `json:"blockers"`
	ApprovedDecisions map[string]string `json:"approved_decisions"`
	RejectedDecisions map[string]string `json:"rejected_decisions"`
	CurrentBehaviors  map[string]string `json:"current_behaviors"`
	ChangedBehaviors  map[string]string `json:"changed_behaviors"`
	MustAvoid         map[string]string `json:"must_avoid"`
	Nuances           map[string]string `json:"nuances"`
	KnownGaps         map[string]string `json:"known_gaps"`
}

type StepResultEntity struct {
	ID            uuid.UUID
	StepID        uuid.UUID
	AgentID       uuid.UUID
	AttemptStatus enums.StepStatus
	Handoffs      []*HandoffEntity
	ContextUsage  *ContextUsage
	StartedAt     time.Time
	CompletedAt   time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type StepEntity struct {
	ID               uuid.UUID
	WorkflowID       uuid.UUID
	Name             string
	Effort           enums.Effort
	PreferredModel   enums.ModelName
	ThinkingLevel    enums.ThinkingLevel
	Instructions     []string
	AutoRetry        bool
	PauseForReview   bool
	ExtraGuidance    string
	RetryCount       int
	Status           enums.StepStatus
	DependsOnStepIDs uuid.UUIDs
	LastReportID     uuid.UUID
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type WorkflowSnapshot struct {
	Workflow *WorkflowEntity
	Steps    []*StepEntity
	Reports  []*StepResultEntity
}

type StepStorage interface {
	SaveStepHistory(
		workflows []*WorkflowEntity,
		steps []*StepEntity,
		reports []*StepResultEntity,
	) error
	LoadStepHistory() ([]*WorkflowSnapshot, error)
}
