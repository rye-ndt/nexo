package template_archive

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-playground/validator/v10"

	"hexago/internal/helpers"
	"hexago/internal/helpers/custom_error"
	"hexago/internal/helpers/enums"
	input_itf "hexago/internal/interface/input"
)

const filePerm = 0o644

type v1 struct{}

func InitV1() input_itf.TemplateArchive {
	return &v1{}
}

func (a *v1) Write(path string, doc *input_itf.TemplateExport) error {
	if strings.TrimSpace(path) == "" {
		return custom_error.Critical("no file path to write the templates to")
	}

	if doc == nil || len(doc.Templates) == 0 {
		return custom_error.Critical("there is nothing to export")
	}

	encoded, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return custom_error.Critical("cannot write the template file: %v", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return custom_error.Critical("cannot create %s: %v", filepath.Dir(path), err)
	}

	if err := os.WriteFile(path, append(encoded, '\n'), filePerm); err != nil {
		return custom_error.Critical("cannot write %s: %v", path, err)
	}

	return nil
}

func (a *v1) Read(path string) (*input_itf.TemplateExport, error) {
	if strings.TrimSpace(path) == "" {
		return nil, custom_error.Critical("no file path to read the templates from")
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, custom_error.Critical("cannot read %s: %v", path, err)
	}

	doc := &input_itf.TemplateExport{}

	if err := json.Unmarshal(raw, doc); err != nil {
		return nil, custom_error.TypedCritical(enums.ErrTemplateFileInvalid,
			"%s is not a template file: %v", quoted(path), err)
	}

	if doc.Version != input_itf.ArchiveVersion {
		return nil, custom_error.TypedCritical(enums.ErrTemplateFileInvalid,
			"%s is a version %d template file, and this app reads version %d",
			quoted(path), doc.Version, input_itf.ArchiveVersion)
	}

	if err := helpers.ValidateStruct(doc); err != nil {
		return nil, custom_error.TypedCritical(enums.ErrTemplateFileInvalid,
			"%s does not hold valid templates: %s", quoted(path), issues(err))
	}

	return doc, nil
}

func quoted(path string) string {
	return "\u201c" + filepath.Base(path) + "\u201d"
}

func issues(err error) string {
	var failures validator.ValidationErrors

	if !errors.As(err, &failures) {
		return err.Error()
	}

	lines := make([]string, 0, len(failures))

	for _, failure := range failures {
		field := strings.TrimPrefix(failure.Namespace(), "TemplateExport.")

		switch failure.Tag() {
		case "required", "gt":
			lines = append(lines, field+" is missing")
		default:
			lines = append(lines, field+" is not one of the values this app knows")
		}
	}

	return strings.Join(lines, ", ")
}
