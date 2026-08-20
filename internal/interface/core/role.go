package core_itf

import (
	"hexago/internal/helpers/enums"

	"github.com/google/uuid"
)

type RoleInputs struct {
	Description string
	Required    bool
	Type        string
	Default     string
	Options     []string
}

type Role struct {
	ID              uuid.UUID
	Name            string
	Description     string
	Effort          enums.Effort
	Retryable       bool
	PauseForReview  bool
	Inputs          map[string]*RoleInputs
	Instructions    map[string]string
	OutputStructure string
}

type RoleManager interface {
	Upsert(role *Role) (uuid.UUID, error)
	List() ([]*Role, error)
	Remove(id uuid.UUID) error
	Get(id uuid.UUID) (*Role, error)
	Export(ids []uuid.UUID, path string) (int, error)
	Import(path string) (int, error)
}
