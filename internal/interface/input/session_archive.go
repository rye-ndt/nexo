package input_itf

import (
	"encoding/json"
	"time"
)

type SessionExport struct {
	Version    int             `json:"version"`
	ExportedAt time.Time       `json:"exported_at"`
	Session    json.RawMessage `json:"session"`
}

type SessionArchive interface {
	Write(path string, doc *SessionExport) error
	Read(path string) (*SessionExport, error)
}
