package session_archive

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"hexago/internal/helpers/custom_error"
	"hexago/internal/helpers/enums"
	input_itf "hexago/internal/interface/input"
)

const filePerm = 0o644

type v1 struct{}

func InitV1() input_itf.SessionArchive {
	return &v1{}
}

func (a *v1) Write(path string, doc *input_itf.SessionExport) error {
	if strings.TrimSpace(path) == "" {
		return custom_error.Critical("no file path to write the session to")
	}

	if doc == nil || len(doc.Session) == 0 {
		return custom_error.Critical("there is nothing to export")
	}

	if !json.Valid(doc.Session) {
		return custom_error.Critical("the session could not be written as JSON")
	}

	encoded, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return custom_error.Critical("cannot write the session file: %v", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return custom_error.Critical("cannot create %s: %v", filepath.Dir(path), err)
	}

	if err := os.WriteFile(path, append(encoded, '\n'), filePerm); err != nil {
		return custom_error.Critical("cannot write %s: %v", path, err)
	}

	return nil
}

func (a *v1) Read(path string) (*input_itf.SessionExport, error) {
	if strings.TrimSpace(path) == "" {
		return nil, custom_error.Critical("no file path to read the session from")
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, custom_error.Critical("cannot read %s: %v", path, err)
	}

	doc := &input_itf.SessionExport{}

	if err := json.Unmarshal(raw, doc); err != nil {
		return nil, custom_error.TypedCritical(enums.ErrSessionFileInvalid,
			"%s is not a session file: %v", quoted(path), err)
	}

	if doc.Version != input_itf.ArchiveVersion {
		return nil, custom_error.TypedCritical(enums.ErrSessionFileInvalid,
			"%s is a version %d session file, and this app reads version %d",
			quoted(path), doc.Version, input_itf.ArchiveVersion)
	}

	if len(doc.Session) == 0 {
		return nil, custom_error.TypedCritical(enums.ErrSessionFileInvalid,
			"%s holds no session", quoted(path))
	}

	if !json.Valid(doc.Session) {
		return nil, custom_error.TypedCritical(enums.ErrSessionFileInvalid,
			"%s does not hold a readable session", quoted(path))
	}

	return doc, nil
}

func quoted(path string) string {
	return "“" + filepath.Base(path) + "”"
}
