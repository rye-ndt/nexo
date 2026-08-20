package role_archive

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

func validDoc() *input_itf.RoleExport {
	return &input_itf.RoleExport{
		Version:    input_itf.ArchiveVersion,
		ExportedAt: helpers.NewUTC(),
		Roles: []*input_itf.RoleRecord{{
			ID:          uuid.MustParse("0192f3a1-0001-7000-8000-000000000001"),
			Name:        "Code reviewer",
			Description: "Reads a diff.",
			Effort:      enums.EffortDeep,
			Retryable:   true,
			Inputs: map[string]*input_itf.RoleInputRecord{
				"scope": {Description: "What to read", Required: true, Type: enums.TextInput.String()},
			},
			Instructions:    map[string]string{"base": "You review code."},
			OutputStructure: "verdict: ship | fix first",
		}},
	}
}

func writeFile(t *testing.T, body string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "roles.json")

	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("cannot seed the file: %v", err)
	}

	return path
}

func TestWriteThenReadKeepsEveryField(t *testing.T) {
	archive := InitV1()
	path := filepath.Join(t.TempDir(), "out", "roles.json")

	if err := archive.Write(path, validDoc()); err != nil {
		t.Fatalf("write: %v", err)
	}

	read, err := archive.Read(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	if len(read.Roles) != 1 {
		t.Fatalf("expected 1 role, got %d", len(read.Roles))
	}

	first := read.Roles[0]
	want := validDoc().Roles[0]

	if first.ID != want.ID || first.Name != want.Name || first.Effort != want.Effort {
		t.Fatalf("role did not survive the round trip: %+v", first)
	}

	if first.Inputs["scope"].Type != enums.TextInput.String() || !first.Inputs["scope"].Required {
		t.Fatalf("input did not survive the round trip: %+v", first.Inputs["scope"])
	}

	if first.Instructions["base"] != want.Instructions["base"] {
		t.Fatalf("prompt did not survive the round trip: %+v", first.Instructions)
	}
}

func TestWriteRefusesAnEmptyExport(t *testing.T) {
	archive := InitV1()

	if err := archive.Write(filepath.Join(t.TempDir(), "out.json"), &input_itf.RoleExport{}); err == nil {
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
			want: "is not a role file",
		},
		{
			name: "a newer version",
			body: `{"version":3,"exported_at":"2026-08-13T09:00:00Z","roles":[]}`,
			want: "version 3 role file",
		},
		{
			name: "no roles",
			body: `{"version": 2,"exported_at":"2026-08-13T09:00:00Z","roles":[]}`,
			want: "roles is missing",
		},
		{
			name: "a record with no name",
			body: `{"version": 2,"exported_at":"2026-08-13T09:00:00Z","roles":[
				{"id":"0192f3a1-0001-7000-8000-000000000001","effort":"deep",
				 "instructions":{"base":"hi"}}]}`,
			want: "roles[0].name is missing",
		},
		{
			name: "a record with no prompts",
			body: `{"version": 2,"exported_at":"2026-08-13T09:00:00Z","roles":[
				{"id":"0192f3a1-0001-7000-8000-000000000001","name":"Reviewer",
				 "effort":"deep","instructions":{}}]}`,
			want: "roles[0].instructions is missing",
		},
		{
			name: "an unknown effort",
			body: `{"version": 2,"exported_at":"2026-08-13T09:00:00Z","roles":[
				{"id":"0192f3a1-0001-7000-8000-000000000001","name":"Reviewer",
				 "effort":"weekly","instructions":{"base":"hi"}}]}`,
			want: "roles[0].effort is not one of the values",
		},
		{
			name: "an unknown input type",
			body: `{"version": 2,"exported_at":"2026-08-13T09:00:00Z","roles":[
				{"id":"0192f3a1-0001-7000-8000-000000000001","name":"Reviewer",
				 "effort":"deep","instructions":{"base":"hi"},
				 "inputs":{"scope":{"description":"d","required":true,"type":"colour"}}}]}`,
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

			if !strings.Contains(err.Error(), string(enums.ErrRoleFileInvalid)) {
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
