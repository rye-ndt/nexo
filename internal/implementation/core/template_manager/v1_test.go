package template_manager

import (
	"strings"
	"testing"

	"hexago/internal/helpers"
	"hexago/internal/helpers/custom_error"
	"hexago/internal/helpers/enums"
	core_itf "hexago/internal/interface/core"
	input_itf "hexago/internal/interface/input"

	"github.com/google/uuid"
)

type fakeStore struct {
	templates []*input_itf.TemplateEntity
	batches   int
}

func (s *fakeStore) Upsert(template *input_itf.TemplateEntity) error {
	s.templates = append(s.templates, template)
	return nil
}

func (s *fakeStore) UpsertMany(templates []*input_itf.TemplateEntity) error {
	s.batches++
	s.templates = append(s.templates, templates...)

	return nil
}

func (s *fakeStore) List() ([]*input_itf.TemplateEntity, error) {
	return s.templates, nil
}

func (s *fakeStore) Find(id uuid.UUID) (*input_itf.TemplateEntity, error) {
	for _, template := range s.templates {
		if template.ID == id {
			return template, nil
		}
	}

	return nil, nil
}

func (s *fakeStore) Remove(id uuid.UUID) error {
	kept := []*input_itf.TemplateEntity{}

	for _, template := range s.templates {
		if template.ID != id {
			kept = append(kept, template)
		}
	}

	s.templates = kept

	return nil
}

type fakeArchive struct {
	doc     *input_itf.TemplateExport
	written *input_itf.TemplateExport
	path    string
	readErr error
}

func (a *fakeArchive) Write(path string, doc *input_itf.TemplateExport) error {
	a.path = path
	a.written = doc

	return nil
}

func (a *fakeArchive) Read(_ string) (*input_itf.TemplateExport, error) {
	if a.readErr != nil {
		return nil, a.readErr
	}

	return a.doc, nil
}

func storedTemplate(id uuid.UUID, name string) *input_itf.TemplateEntity {
	now := helpers.NewUTC()

	return &input_itf.TemplateEntity{
		ID:        id,
		Name:      name,
		TaskLevel: enums.DailyTask,
		Params: map[string]*input_itf.TemplateParamEntity{
			"scope": {Description: "What to read", Required: true, Type: enums.TextParam.String()},
		},
		SystemPrompts: map[string]string{"base": "You review code."},
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func importedRecord(id uuid.UUID, name string) *input_itf.TemplateRecord {
	return &input_itf.TemplateRecord{
		ID:        id,
		Name:      name,
		TaskLevel: enums.DailyTask,
		Params: map[string]*input_itf.TemplateParamRecord{
			"scope": {Description: "What to read", Required: true, Type: enums.TextParam.String()},
		},
		SystemPrompts: map[string]string{"base": "You review code."},
	}
}

func manager(t *testing.T, store *fakeStore, archive *fakeArchive) core_itf.AgentTemplateManager {
	t.Helper()

	built, err := InitV1(store, archive)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	return built
}

func TestExportWritesEveryPickedTemplate(t *testing.T) {
	first, second := uuid.New(), uuid.New()
	store := &fakeStore{templates: []*input_itf.TemplateEntity{
		storedTemplate(first, "Code reviewer"),
		storedTemplate(second, "Test writer"),
		storedTemplate(uuid.New(), "Left behind"),
	}}
	archive := &fakeArchive{}

	count, err := manager(t, store, archive).Export([]uuid.UUID{first, second}, "/tmp/templates.json")
	if err != nil {
		t.Fatalf("export: %v", err)
	}

	if count != 2 {
		t.Fatalf("expected 2 exported, got %d", count)
	}

	if archive.path != "/tmp/templates.json" {
		t.Fatalf("wrote to the wrong path: %s", archive.path)
	}

	if archive.written.Version != input_itf.ArchiveVersion || archive.written.ExportedAt.IsZero() {
		t.Fatalf("envelope is not stamped: %+v", archive.written)
	}

	if len(archive.written.Templates) != 2 {
		t.Fatalf("expected 2 records, got %d", len(archive.written.Templates))
	}

	if archive.written.Templates[0].ID != first || archive.written.Templates[0].Name != "Code reviewer" {
		t.Fatalf("first record is wrong: %+v", archive.written.Templates[0])
	}

	if archive.written.Templates[0].Params["scope"].Type != enums.TextParam.String() {
		t.Fatalf("params did not make it into the record: %+v", archive.written.Templates[0].Params)
	}
}

func TestExportWritesATemplateOnceEvenIfItIsPickedTwice(t *testing.T) {
	id := uuid.New()
	store := &fakeStore{templates: []*input_itf.TemplateEntity{storedTemplate(id, "Code reviewer")}}
	archive := &fakeArchive{}

	count, err := manager(t, store, archive).Export([]uuid.UUID{id, id}, "/tmp/templates.json")
	if err != nil {
		t.Fatalf("export: %v", err)
	}

	if count != 1 || len(archive.written.Templates) != 1 {
		t.Fatalf("expected one record, got %d", len(archive.written.Templates))
	}
}

func TestExportRefusesAnEmptyPickOrAnUnknownTemplate(t *testing.T) {
	store := &fakeStore{}
	archive := &fakeArchive{}
	templates := manager(t, store, archive)

	if _, err := templates.Export(nil, "/tmp/templates.json"); err == nil {
		t.Fatal("expected an empty pick to be refused")
	}

	if _, err := templates.Export([]uuid.UUID{uuid.New()}, "/tmp/templates.json"); err == nil {
		t.Fatal("expected an unknown template to be refused")
	}

	if archive.written != nil {
		t.Fatal("nothing should have been written")
	}
}

func TestImportSavesEveryRecordInOneBatch(t *testing.T) {
	store := &fakeStore{templates: []*input_itf.TemplateEntity{
		storedTemplate(uuid.New(), "Already here"),
	}}
	archive := &fakeArchive{doc: &input_itf.TemplateExport{
		Version:    input_itf.ArchiveVersion,
		ExportedAt: helpers.NewUTC(),
		Templates: []*input_itf.TemplateRecord{
			importedRecord(uuid.New(), "Code reviewer"),
			importedRecord(uuid.New(), "Test writer"),
		},
	}}

	count, err := manager(t, store, archive).Import("/tmp/templates.json")
	if err != nil {
		t.Fatalf("import: %v", err)
	}

	if count != 2 {
		t.Fatalf("expected 2 imported, got %d", count)
	}

	if store.batches != 1 {
		t.Fatalf("expected one all-or-nothing write, got %d", store.batches)
	}

	if len(store.templates) != 3 {
		t.Fatalf("expected 3 templates on disk, got %d", len(store.templates))
	}

	saved := store.templates[len(store.templates)-1]
	if saved.CreatedAt.IsZero() || saved.UpdatedAt.IsZero() {
		t.Fatalf("imported template is not stamped: %+v", saved)
	}

	if saved.ID != archive.doc.Templates[1].ID {
		t.Fatalf("imported template id = %v, want the id it was exported with %v",
			saved.ID, archive.doc.Templates[1].ID)
	}
}

func TestImportIsBlockedByAConflict(t *testing.T) {
	takenID := uuid.New()

	cases := []struct {
		name    string
		stored  []*input_itf.TemplateEntity
		records []*input_itf.TemplateRecord
		want    string
	}{
		{
			name:    "a name already here",
			stored:  []*input_itf.TemplateEntity{storedTemplate(uuid.New(), "Code reviewer")},
			records: []*input_itf.TemplateRecord{importedRecord(uuid.New(), "Code reviewer")},
			want:    `a template named "Code reviewer" is already here`,
		},
		{
			name:    "the same name in a different case",
			stored:  []*input_itf.TemplateEntity{storedTemplate(uuid.New(), "Code reviewer")},
			records: []*input_itf.TemplateRecord{importedRecord(uuid.New(), "code REVIEWER")},
			want:    "is already here",
		},
		{
			name:    "an id already here under another name",
			stored:  []*input_itf.TemplateEntity{storedTemplate(takenID, "Something else")},
			records: []*input_itf.TemplateRecord{importedRecord(takenID, "Code reviewer")},
			want:    "carries the id of a template already here",
		},
		{
			name:   "a re-import of the very same file",
			stored: []*input_itf.TemplateEntity{storedTemplate(takenID, "Code reviewer")},
			records: []*input_itf.TemplateRecord{
				importedRecord(takenID, "Code reviewer"),
			},
			want: "is already here",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			store := &fakeStore{templates: testCase.stored}
			archive := &fakeArchive{doc: &input_itf.TemplateExport{
				Version:    input_itf.ArchiveVersion,
				ExportedAt: helpers.NewUTC(),
				Templates:  testCase.records,
			}}

			_, err := manager(t, store, archive).Import("/tmp/templates.json")
			if err == nil {
				t.Fatal("expected the import to be blocked")
			}

			if !strings.Contains(err.Error(), testCase.want) {
				t.Fatalf("error does not say why: %v", err)
			}

			if !strings.Contains(err.Error(), string(enums.ErrTemplateConflict)) {
				t.Fatalf("error is not typed for the frontend: %v", err)
			}

			if store.batches != 0 || len(store.templates) != len(testCase.stored) {
				t.Fatal("a blocked import must write nothing")
			}
		})
	}
}

func TestImportIsBlockedWhenTheFileRepeatsItself(t *testing.T) {
	repeated := uuid.New()

	cases := []struct {
		name    string
		records []*input_itf.TemplateRecord
		want    string
	}{
		{
			name: "two records share a name",
			records: []*input_itf.TemplateRecord{
				importedRecord(uuid.New(), "Code reviewer"),
				importedRecord(uuid.New(), "Code reviewer"),
			},
			want: `the file lists "Code reviewer" twice`,
		},
		{
			name: "two records share an id",
			records: []*input_itf.TemplateRecord{
				importedRecord(repeated, "Code reviewer"),
				importedRecord(repeated, "Test writer"),
			},
			want: "the file lists the id",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			store := &fakeStore{}
			archive := &fakeArchive{doc: &input_itf.TemplateExport{
				Version:    input_itf.ArchiveVersion,
				ExportedAt: helpers.NewUTC(),
				Templates:  testCase.records,
			}}

			_, err := manager(t, store, archive).Import("/tmp/templates.json")
			if err == nil {
				t.Fatal("expected the import to be blocked")
			}

			if !strings.Contains(err.Error(), testCase.want) {
				t.Fatalf("error does not say why: %v", err)
			}

			if len(store.templates) != 0 {
				t.Fatal("a blocked import must write nothing")
			}
		})
	}
}

func TestImportSurfacesWhatTheArchiveRefused(t *testing.T) {
	store := &fakeStore{}
	archive := &fakeArchive{readErr: custom_error.TypedCritical(
		enums.ErrTemplateFileInvalid, "templates.json is not a template file")}

	_, err := manager(t, store, archive).Import("/tmp/templates.json")
	if err == nil || !strings.Contains(err.Error(), "is not a template file") {
		t.Fatalf("expected the archive's own error, got %v", err)
	}

	if len(store.templates) != 0 {
		t.Fatal("a refused file must write nothing")
	}
}
