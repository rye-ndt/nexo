package input_itf

import (
	"hexago/internal/helpers/enums"
	"time"

	"github.com/google/uuid"
)

const ArchiveVersion = 2

type RoleInputRecord struct {
	Description string   `json:"description" mapstructure:"description"`
	Required    bool     `json:"required" mapstructure:"required"`
	Type        string   `json:"type" mapstructure:"type" validate:"required,input_type"`
	Default     string   `json:"default" mapstructure:"default"`
	Options     []string `json:"options" mapstructure:"options"`
}

type RoleRecord struct {
	ID              uuid.UUID                   `json:"id" mapstructure:"id" validate:"required"`
	Name            string                      `json:"name" mapstructure:"name" validate:"required"`
	Description     string                      `json:"description" mapstructure:"description"`
	Effort          enums.Effort                `json:"effort" mapstructure:"effort" validate:"required,effort"`
	Retryable       bool                        `json:"retryable" mapstructure:"retryable"`
	PauseForReview  bool                        `json:"pause_for_review" mapstructure:"pause_for_review"`
	Inputs          map[string]*RoleInputRecord `json:"inputs" mapstructure:"inputs" validate:"dive,keys,required,endkeys,required"`
	Instructions    map[string]string           `json:"instructions" mapstructure:"instructions" validate:"required,gt=0,dive,keys,required,endkeys,required"`
	OutputStructure string                      `json:"output_structure" mapstructure:"output_structure"`
}

type RoleExport struct {
	Version    int           `json:"version" mapstructure:"version" validate:"required,eq=2"`
	ExportedAt time.Time     `json:"exported_at" mapstructure:"exported_at" validate:"required"`
	Roles      []*RoleRecord `json:"roles" mapstructure:"roles" validate:"required,gt=0,dive,required"`
}

type RoleArchive interface {
	Write(path string, doc *RoleExport) error
	Read(path string) (*RoleExport, error)
}
