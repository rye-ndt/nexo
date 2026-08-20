package input_itf

import (
	"hexago/internal/helpers/enums"
	"time"

	"github.com/google/uuid"
)

type RoleInputEntity struct {
	Description string   `json:"description"`
	Required    bool     `json:"required"`
	Type        string   `json:"type,omitempty"`
	Default     string   `json:"default,omitempty"`
	Options     []string `json:"options,omitempty"`
}

type RoleEntity struct {
	ID              uuid.UUID
	Name            string `validate:"required"`
	Description     string
	Effort          enums.Effort `validate:"required,effort"`
	Retryable       bool
	PauseForReview  bool
	Inputs          map[string]*RoleInputEntity `validate:"dive,keys,required,endkeys,required"`
	Instructions    map[string]string           `validate:"required,gt=0,dive,keys,required,endkeys,required"`
	OutputStructure string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type RoleStorage interface {
	Upsert(role *RoleEntity) error
	UpsertMany(roles []*RoleEntity) error
	List() ([]*RoleEntity, error)
	Find(id uuid.UUID) (*RoleEntity, error)
	Remove(id uuid.UUID) error
}
