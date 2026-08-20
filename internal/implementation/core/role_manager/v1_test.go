package role_manager

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
	roles   []*input_itf.RoleEntity
	batches int
}

func (s *fakeStore) Upsert(role *input_itf.RoleEntity) error {
	s.roles = append(s.roles, role)
	return nil
}

func (s *fakeStore) UpsertMany(roles []*input_itf.RoleEntity) error {
	s.batches++
	s.roles = append(s.roles, roles...)

	return nil
}

func (s *fakeStore) List() ([]*input_itf.RoleEntity, error) {
	return s.roles, nil
}

func (s *fakeStore) Find(id uuid.UUID) (*input_itf.RoleEntity, error) {
	for _, role := range s.roles {
		if role.ID == id {
			return role, nil
		}
	}

	return nil, nil
}

func (s *fakeStore) Remove(id uuid.UUID) error {
	kept := []*input_itf.RoleEntity{}

	for _, role := range s.roles {
		if role.ID != id {
			kept = append(kept, role)
		}
	}

	s.roles = kept

	return nil
}

type fakeArchive struct {
	doc     *input_itf.RoleExport
	written *input_itf.RoleExport
	path    string
	readErr error
}

func (a *fakeArchive) Write(path string, doc *input_itf.RoleExport) error {
	a.path = path
	a.written = doc

	return nil
}

func (a *fakeArchive) Read(_ string) (*input_itf.RoleExport, error) {
	if a.readErr != nil {
		return nil, a.readErr
	}

	return a.doc, nil
}

func storedRole(id uuid.UUID, name string) *input_itf.RoleEntity {
	now := helpers.NewUTC()

	return &input_itf.RoleEntity{
		ID:     id,
		Name:   name,
		Effort: enums.EffortStandard,
		Inputs: map[string]*input_itf.RoleInputEntity{
			"scope": {Description: "What to read", Required: true, Type: enums.TextInput.String()},
		},
		Instructions: map[string]string{"base": "You review code."},
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func importedRecord(id uuid.UUID, name string) *input_itf.RoleRecord {
	return &input_itf.RoleRecord{
		ID:     id,
		Name:   name,
		Effort: enums.EffortStandard,
		Inputs: map[string]*input_itf.RoleInputRecord{
			"scope": {Description: "What to read", Required: true, Type: enums.TextInput.String()},
		},
		Instructions: map[string]string{"base": "You review code."},
	}
}

func manager(t *testing.T, store *fakeStore, archive *fakeArchive) core_itf.RoleManager {
	t.Helper()

	built, err := InitV1(store, archive)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	return built
}

func TestExportWritesEveryPickedRole(t *testing.T) {
	first, second := uuid.New(), uuid.New()
	store := &fakeStore{roles: []*input_itf.RoleEntity{
		storedRole(first, "Code reviewer"),
		storedRole(second, "Test writer"),
		storedRole(uuid.New(), "Left behind"),
	}}
	archive := &fakeArchive{}

	count, err := manager(t, store, archive).Export([]uuid.UUID{first, second}, "/tmp/roles.json")
	if err != nil {
		t.Fatalf("export: %v", err)
	}

	if count != 2 {
		t.Fatalf("expected 2 exported, got %d", count)
	}

	if archive.path != "/tmp/roles.json" {
		t.Fatalf("wrote to the wrong path: %s", archive.path)
	}

	if archive.written.Version != input_itf.ArchiveVersion || archive.written.ExportedAt.IsZero() {
		t.Fatalf("envelope is not stamped: %+v", archive.written)
	}

	if len(archive.written.Roles) != 2 {
		t.Fatalf("expected 2 records, got %d", len(archive.written.Roles))
	}

	if archive.written.Roles[0].ID != first || archive.written.Roles[0].Name != "Code reviewer" {
		t.Fatalf("first record is wrong: %+v", archive.written.Roles[0])
	}

	if archive.written.Roles[0].Inputs["scope"].Type != enums.TextInput.String() {
		t.Fatalf("inputs did not make it into the record: %+v", archive.written.Roles[0].Inputs)
	}
}

func TestExportWritesARoleOnceEvenIfItIsPickedTwice(t *testing.T) {
	id := uuid.New()
	store := &fakeStore{roles: []*input_itf.RoleEntity{storedRole(id, "Code reviewer")}}
	archive := &fakeArchive{}

	count, err := manager(t, store, archive).Export([]uuid.UUID{id, id}, "/tmp/roles.json")
	if err != nil {
		t.Fatalf("export: %v", err)
	}

	if count != 1 || len(archive.written.Roles) != 1 {
		t.Fatalf("expected one record, got %d", len(archive.written.Roles))
	}
}

func TestExportRefusesAnEmptyPickOrAnUnknownRole(t *testing.T) {
	store := &fakeStore{}
	archive := &fakeArchive{}
	roles := manager(t, store, archive)

	if _, err := roles.Export(nil, "/tmp/roles.json"); err == nil {
		t.Fatal("expected an empty pick to be refused")
	}

	if _, err := roles.Export([]uuid.UUID{uuid.New()}, "/tmp/roles.json"); err == nil {
		t.Fatal("expected an unknown role to be refused")
	}

	if archive.written != nil {
		t.Fatal("nothing should have been written")
	}
}

func TestImportSavesEveryRecordInOneBatch(t *testing.T) {
	store := &fakeStore{roles: []*input_itf.RoleEntity{
		storedRole(uuid.New(), "Already here"),
	}}
	archive := &fakeArchive{doc: &input_itf.RoleExport{
		Version:    input_itf.ArchiveVersion,
		ExportedAt: helpers.NewUTC(),
		Roles: []*input_itf.RoleRecord{
			importedRecord(uuid.New(), "Code reviewer"),
			importedRecord(uuid.New(), "Test writer"),
		},
	}}

	count, err := manager(t, store, archive).Import("/tmp/roles.json")
	if err != nil {
		t.Fatalf("import: %v", err)
	}

	if count != 2 {
		t.Fatalf("expected 2 imported, got %d", count)
	}

	if store.batches != 1 {
		t.Fatalf("expected one all-or-nothing write, got %d", store.batches)
	}

	if len(store.roles) != 3 {
		t.Fatalf("expected 3 roles on disk, got %d", len(store.roles))
	}

	saved := store.roles[len(store.roles)-1]
	if saved.CreatedAt.IsZero() || saved.UpdatedAt.IsZero() {
		t.Fatalf("imported role is not stamped: %+v", saved)
	}

	if saved.ID != archive.doc.Roles[1].ID {
		t.Fatalf("imported role id = %v, want the id it was exported with %v",
			saved.ID, archive.doc.Roles[1].ID)
	}
}

func TestImportIsBlockedByAConflict(t *testing.T) {
	takenID := uuid.New()

	cases := []struct {
		name    string
		stored  []*input_itf.RoleEntity
		records []*input_itf.RoleRecord
		want    string
	}{
		{
			name:    "a name already here",
			stored:  []*input_itf.RoleEntity{storedRole(uuid.New(), "Code reviewer")},
			records: []*input_itf.RoleRecord{importedRecord(uuid.New(), "Code reviewer")},
			want:    `a role named "Code reviewer" is already here`,
		},
		{
			name:    "the same name in a different case",
			stored:  []*input_itf.RoleEntity{storedRole(uuid.New(), "Code reviewer")},
			records: []*input_itf.RoleRecord{importedRecord(uuid.New(), "code REVIEWER")},
			want:    "is already here",
		},
		{
			name:    "an id already here under another name",
			stored:  []*input_itf.RoleEntity{storedRole(takenID, "Something else")},
			records: []*input_itf.RoleRecord{importedRecord(takenID, "Code reviewer")},
			want:    "carries the id of a role already here",
		},
		{
			name:   "a re-import of the very same file",
			stored: []*input_itf.RoleEntity{storedRole(takenID, "Code reviewer")},
			records: []*input_itf.RoleRecord{
				importedRecord(takenID, "Code reviewer"),
			},
			want: "is already here",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			store := &fakeStore{roles: testCase.stored}
			archive := &fakeArchive{doc: &input_itf.RoleExport{
				Version:    input_itf.ArchiveVersion,
				ExportedAt: helpers.NewUTC(),
				Roles:      testCase.records,
			}}

			_, err := manager(t, store, archive).Import("/tmp/roles.json")
			if err == nil {
				t.Fatal("expected the import to be blocked")
			}

			if !strings.Contains(err.Error(), testCase.want) {
				t.Fatalf("error does not say why: %v", err)
			}

			if !strings.Contains(err.Error(), string(enums.ErrRoleConflict)) {
				t.Fatalf("error is not typed for the frontend: %v", err)
			}

			if store.batches != 0 || len(store.roles) != len(testCase.stored) {
				t.Fatal("a blocked import must write nothing")
			}
		})
	}
}

func TestImportIsBlockedWhenTheFileRepeatsItself(t *testing.T) {
	repeated := uuid.New()

	cases := []struct {
		name    string
		records []*input_itf.RoleRecord
		want    string
	}{
		{
			name: "two records share a name",
			records: []*input_itf.RoleRecord{
				importedRecord(uuid.New(), "Code reviewer"),
				importedRecord(uuid.New(), "Code reviewer"),
			},
			want: `the file lists "Code reviewer" twice`,
		},
		{
			name: "two records share an id",
			records: []*input_itf.RoleRecord{
				importedRecord(repeated, "Code reviewer"),
				importedRecord(repeated, "Test writer"),
			},
			want: "the file lists the id",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			store := &fakeStore{}
			archive := &fakeArchive{doc: &input_itf.RoleExport{
				Version:    input_itf.ArchiveVersion,
				ExportedAt: helpers.NewUTC(),
				Roles:      testCase.records,
			}}

			_, err := manager(t, store, archive).Import("/tmp/roles.json")
			if err == nil {
				t.Fatal("expected the import to be blocked")
			}

			if !strings.Contains(err.Error(), testCase.want) {
				t.Fatalf("error does not say why: %v", err)
			}

			if len(store.roles) != 0 {
				t.Fatal("a blocked import must write nothing")
			}
		})
	}
}

func TestImportSurfacesWhatTheArchiveRefused(t *testing.T) {
	store := &fakeStore{}
	archive := &fakeArchive{readErr: custom_error.TypedCritical(
		enums.ErrRoleFileInvalid, "roles.json is not a role file")}

	_, err := manager(t, store, archive).Import("/tmp/roles.json")
	if err == nil || !strings.Contains(err.Error(), "is not a role file") {
		t.Fatalf("expected the archive's own error, got %v", err)
	}

	if len(store.roles) != 0 {
		t.Fatal("a refused file must write nothing")
	}
}
