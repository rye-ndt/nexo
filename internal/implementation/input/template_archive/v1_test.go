package template_archive

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"hexago/internal/helpers"
	"hexago/internal/helpers/enums"
	input_itf "hexago/internal/interface/input"

	"github.com/google/uuid"
)

func validDoc() *input_itf.TemplateExport {
	return &input_itf.TemplateExport{
		Version:    input_itf.ArchiveVersion,
		ExportedAt: helpers.NewUTC(),
		Templates: []*input_itf.TemplateRecord{{
			ID:        uuid.MustParse("0192f3a1-0001-7000-8000-000000000001"),
			Name:      "Code reviewer",
			Role:      "Reads a diff.",
			TaskLevel: enums.HeavyTask,
			Retryable: true,
			Params: map[string]*input_itf.TemplateParamRecord{
				"scope": {Description: "What to read", Required: true, Type: enums.TextParam.String()},
			},
			SystemPrompts:   map[string]string{"base": "You review code."},
			OutputStructure: "verdict: ship | fix first",
		}},
	}
}

func writeFile(t *testing.T, body string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "templates.json")

	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("cannot seed the file: %v", err)
	}

	return path
}

func TestWriteThenReadKeepsEveryField(t *testing.T) {
	archive := InitV1()
	path := filepath.Join(t.TempDir(), "out", "templates.json")

	if err := archive.Write(path, validDoc()); err != nil {
		t.Fatalf("write: %v", err)
	}

	read, err := archive.Read(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	if len(read.Templates) != 1 {
		t.Fatalf("expected 1 template, got %d", len(read.Templates))
	}

	first := read.Templates[0]
	want := validDoc().Templates[0]

	if first.ID != want.ID || first.Name != want.Name || first.TaskLevel != want.TaskLevel {
		t.Fatalf("template did not survive the round trip: %+v", first)
	}

	if first.Params["scope"].Type != enums.TextParam.String() || !first.Params["scope"].Required {
		t.Fatalf("param did not survive the round trip: %+v", first.Params["scope"])
	}

	if first.SystemPrompts["base"] != want.SystemPrompts["base"] {
		t.Fatalf("prompt did not survive the round trip: %+v", first.SystemPrompts)
	}
}

func TestWriteRefusesAnEmptyExport(t *testing.T) {
	archive := InitV1()

	if err := archive.Write(filepath.Join(t.TempDir(), "out.json"), &input_itf.TemplateExport{}); err == nil {
		t.Fatal("expected an empty export to be refused")
	}
}

func TestReadRejectsInvalidFiles(t *testing.T) {
	cases := []struct {
		name string
		body string
		want string
	}{
		{
			name: "not json at all",
			body: "this is not json",
			want: "is not a template file",
		},
		{
			name: "a newer version",
			body: `{"version":2,"exported_at":"2026-08-13T09:00:00Z","templates":[]}`,
			want: "version 2 template file",
		},
		{
			name: "no templates",
			body: `{"version":1,"exported_at":"2026-08-13T09:00:00Z","templates":[]}`,
			want: "templates is missing",
		},
		{
			name: "a record with no name",
			body: `{"version":1,"exported_at":"2026-08-13T09:00:00Z","templates":[
				{"id":"0192f3a1-0001-7000-8000-000000000001","task_level":"heavy_task",
				 "system_prompts":{"base":"hi"}}]}`,
			want: "templates[0].name is missing",
		},
		{
			name: "a record with no prompts",
			body: `{"version":1,"exported_at":"2026-08-13T09:00:00Z","templates":[
				{"id":"0192f3a1-0001-7000-8000-000000000001","name":"Reviewer",
				 "task_level":"heavy_task","system_prompts":{}}]}`,
			want: "templates[0].system_prompts is missing",
		},
		{
			name: "an unknown task level",
			body: `{"version":1,"exported_at":"2026-08-13T09:00:00Z","templates":[
				{"id":"0192f3a1-0001-7000-8000-000000000001","name":"Reviewer",
				 "task_level":"weekly_task","system_prompts":{"base":"hi"}}]}`,
			want: "templates[0].task_level is not one of the values",
		},
		{
			name: "an unknown param type",
			body: `{"version":1,"exported_at":"2026-08-13T09:00:00Z","templates":[
				{"id":"0192f3a1-0001-7000-8000-000000000001","name":"Reviewer",
				 "task_level":"heavy_task","system_prompts":{"base":"hi"},
				 "params":{"scope":{"description":"d","required":true,"type":"colour"}}}]}`,
			want: "type is not one of the values",
		},
	}

	archive := InitV1()

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := archive.Read(writeFile(t, testCase.body))
			if err == nil {
				t.Fatal("expected the file to be rejected")
			}

			if !strings.Contains(err.Error(), testCase.want) {
				t.Fatalf("error does not say why: %v", err)
			}

			if !strings.Contains(err.Error(), string(enums.ErrTemplateFileInvalid)) {
				t.Fatalf("error is not typed for the frontend: %v", err)
			}
		})
	}
}

func TestReadReportsAMissingFile(t *testing.T) {
	archive := InitV1()

	if _, err := archive.Read(filepath.Join(t.TempDir(), "nothing.json")); err == nil {
		t.Fatal("expected a missing file to be reported")
	}
}
