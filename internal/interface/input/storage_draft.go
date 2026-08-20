package input_itf

import (
	"time"

	"github.com/google/uuid"
)

type WorkflowDraftEntity struct {
	ID        uuid.UUID
	Doc       string
	UpdatedAt time.Time
}

type DraftStorage interface {
	List() ([]*WorkflowDraftEntity, error)
	Save(draft *WorkflowDraftEntity) error
	Delete(id uuid.UUID) error
}
