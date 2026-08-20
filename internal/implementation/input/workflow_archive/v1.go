package workflow_archive

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

func InitV1() input_itf.WorkflowArchive {
	return &v1{}
}

func (a *v1) Write(path string, doc *input_itf.WorkflowExport) error {
	if strings.TrimSpace(path) == "" {
		return custom_error.Critical("no file path to write the workflow to")
	}

	if doc == nil || len(doc.Workflow) == 0 {
		return custom_error.Critical("there is nothing to export")
	}

	if !json.Valid(doc.Workflow) {
		return custom_error.Critical("the workflow could not be written as JSON")
	}

	encoded, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return custom_error.Critical("cannot write the workflow file: %v", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return custom_error.Critical("cannot create %s: %v", filepath.Dir(path), err)
	}

	if err := os.WriteFile(path, append(encoded, '\n'), filePerm); err != nil {
		return custom_error.Critical("cannot write %s: %v", path, err)
	}

	return nil
}

func (a *v1) Read(path string) (*input_itf.WorkflowExport, error) {
	if strings.TrimSpace(path) == "" {
		return nil, custom_error.Critical("no file path to read the workflow from")
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, custom_error.Critical("cannot read %s: %v", path, err)
	}

	doc := &input_itf.WorkflowExport{}

	if err := json.Unmarshal(raw, doc); err != nil {
		return nil, custom_error.TypedCritical(enums.ErrWorkflowFileInvalid,
			"%s is not a workflow file: %v", quoted(path), err)
	}

	if doc.Version != input_itf.ArchiveVersion {
		return nil, custom_error.TypedCritical(enums.ErrWorkflowFileInvalid,
			"%s is a version %d workflow file, and this app reads version %d",
			quoted(path), doc.Version, input_itf.ArchiveVersion)
	}

	if len(doc.Workflow) == 0 {
		return nil, custom_error.TypedCritical(enums.ErrWorkflowFileInvalid,
			"%s holds no workflow", quoted(path))
	}

	if !json.Valid(doc.Workflow) {
		return nil, custom_error.TypedCritical(enums.ErrWorkflowFileInvalid,
			"%s does not hold a readable workflow", quoted(path))
	}

	return doc, nil
}

func quoted(path string) string {
	return "“" + filepath.Base(path) + "”"
}
