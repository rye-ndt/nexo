package input_itf

import (
	"encoding/json"
	"time"
)

type WorkflowExport struct {
	Version    int             `json:"version"`
	ExportedAt time.Time       `json:"exported_at"`
	Workflow   json.RawMessage `json:"workflow"`
}

type WorkflowArchive interface {
	Write(path string, doc *WorkflowExport) error
	Read(path string) (*WorkflowExport, error)
}
