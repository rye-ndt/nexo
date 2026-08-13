package input_itf

import (
	"hexago/internal/helpers/enums"
	"time"

	"github.com/google/uuid"
)

const ArchiveVersion = 1

type TemplateParamRecord struct {
	Description string   `json:"description" mapstructure:"description"`
	Required    bool     `json:"required" mapstructure:"required"`
	Type        string   `json:"type" mapstructure:"type" validate:"required,param_type"`
	Default     string   `json:"default" mapstructure:"default"`
	Options     []string `json:"options" mapstructure:"options"`
}

type TemplateRecord struct {
	ID                   uuid.UUID                       `json:"id" mapstructure:"id" validate:"required"`
	Name                 string                          `json:"name" mapstructure:"name" validate:"required"`
	Role                 string                          `json:"role" mapstructure:"role"`
	TaskLevel            enums.TaskLevel                 `json:"task_level" mapstructure:"task_level" validate:"required,task_level"`
	Retryable            bool                            `json:"retryable" mapstructure:"retryable"`
	ManualAcceptRequired bool                            `json:"manual_accept_required" mapstructure:"manual_accept_required"`
	Params               map[string]*TemplateParamRecord `json:"params" mapstructure:"params" validate:"dive,keys,required,endkeys,required"`
	SystemPrompts        map[string]string               `json:"system_prompts" mapstructure:"system_prompts" validate:"required,gt=0,dive,keys,required,endkeys,required"`
	OutputStructure      string                          `json:"output_structure" mapstructure:"output_structure"`
}

type TemplateExport struct {
	Version    int               `json:"version" mapstructure:"version" validate:"required,eq=1"`
	ExportedAt time.Time         `json:"exported_at" mapstructure:"exported_at" validate:"required"`
	Templates  []*TemplateRecord `json:"templates" mapstructure:"templates" validate:"required,gt=0,dive,required"`
}

type TemplateArchive interface {
	Write(path string, doc *TemplateExport) error
	Read(path string) (*TemplateExport, error)
}
