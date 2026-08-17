package input_itf

import (
	"hexago/internal/helpers/enums"
	"time"

	"github.com/google/uuid"
)

type SessionEntity struct {
	ID             uuid.UUID
	WorkingDirPath string
	ContextDirPath string
	StartedAt      time.Time
	CompletedAt    time.Time
	TotalTask      int
	TotalRetry     int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type HandoverDocEntity struct {
	Task              string            `json:"task_name"`
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

type TaskReportEntity struct {
	ID            uuid.UUID
	TaskID        uuid.UUID
	AgentID       uuid.UUID
	AttemptStatus enums.TaskStatus
	HandoverDocs  []*HandoverDocEntity
	ContextUsage  *ContextUsage
	StartedAt     time.Time
	CompletedAt   time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type TaskEntity struct {
	ID                   uuid.UUID
	SessionID            uuid.UUID
	Name                 string
	TaskLevel            enums.TaskLevel
	PreferredModel       enums.ModelName
	ThinkingLevel        enums.ThinkingLevel
	SystemPrompts        []string
	AutoRetry            bool
	ManualAcceptRequired bool
	ExtraGuidance        string
	RetryCount           int
	Status               enums.TaskStatus
	DependsOnTaskIDs     uuid.UUIDs
	LastReportID         uuid.UUID
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type TaskStorage interface {
	SaveTaskHistory(
		sessions []*SessionEntity,
		tasks []*TaskEntity,
		reports []*TaskReportEntity,
	) error
}
