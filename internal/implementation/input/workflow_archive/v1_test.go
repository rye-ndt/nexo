package workflow_archive

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"hexago/internal/helpers"
	"hexago/internal/helpers/enums"
	input_itf "hexago/internal/interface/input"
)

const workflowBody = `{"id":"0198e3a1-0000-7000-8000-000000000001","name":"Coordinator port",` +
	`"steps":[{"id":"step-1","title":"Map the ports","dependsOn":[],` +
	`"spec":{"effort":"deep","instructions":["You review code."]}}]}`

func validDoc() *input_itf.WorkflowExport {
	return &input_itf.WorkflowExport{
		Version:    input_itf.ArchiveVersion,
		ExportedAt: helpers.NewUTC(),
		Workflow:   json.RawMessage(workflowBody),
	}
}

func writeFile(t *testing.T, body string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "workflow.json")

	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("cannot seed the file: %v", err)
	}

	return path
}

func TestWriteThenReadKeepsTheWorkflowBody(t *testing.T) {
	archive := InitV1()
	path := filepath.Join(t.TempDir(), "out", "workflow.json")

	if err := archive.Write(path, validDoc()); err != nil {
		t.Fatalf("write: %v", err)
	}

	read, err := archive.Read(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	if read.Version != input_itf.ArchiveVersion {
		t.Fatalf("expected version %d, got %d", input_itf.ArchiveVersion, read.Version)
	}

	var compacted bytes.Buffer

	if err := json.Compact(&compacted, read.Workflow); err != nil {
		t.Fatalf("the workflow did not come back as JSON: %v", err)
	}

	if compacted.String() != workflowBody {
		t.Fatalf("workflow did not survive the round trip: %s", read.Workflow)
	}
}

func TestWriteRefusesAnEmptyExport(t *testing.T) {
	archive := InitV1()
	path := filepath.Join(t.TempDir(), "out.json")

	if err := archive.Write(path, nil); err == nil {
		t.Fatal("expected a nil export to be refused")
	}

	if err := archive.Write(path, &input_itf.WorkflowExport{Version: input_itf.ArchiveVersion}); err == nil {
		t.Fatal("expected an empty export to be refused")
	}
}

func TestWriteRefusesABodyThatIsNotJSON(t *testing.T) {
	archive := InitV1()

	doc := validDoc()
	doc.Workflow = json.RawMessage("{ this was typed by hand")

	if err := archive.Write(filepath.Join(t.TempDir(), "out.json"), doc); err == nil {
		t.Fatal("expected a body that is not JSON to be refused")
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
			want: "is not a workflow file",
		},
		{
			name: "a newer version",
			body: `{"version":3,"exported_at":"2026-08-13T09:00:00Z","workflow":{"name":"Port"}}`,
			want: "version 3 workflow file",
		},
		{
			name: "no workflow",
			body: `{"version":2,"exported_at":"2026-08-13T09:00:00Z"}`,
			want: "holds no workflow",
		},
		{
			name: "a workflow that is not json",
			body: `{"version":2,"exported_at":"2026-08-13T09:00:00Z","workflow":{"name":}}`,
			want: "is not a workflow file",
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

			if !strings.Contains(err.Error(), string(enums.ErrWorkflowFileInvalid)) {
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
