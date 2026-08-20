package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"hexago/internal/helpers"
	"hexago/internal/helpers/enums"
	input_itf "hexago/internal/interface/input"
)

const (
	explorerRoleID    = "01988000-0001-7000-8000-000000000001"
	plannerRoleID     = "01988000-0002-7000-8000-000000000002"
	implementerRoleID = "01988000-0003-7000-8000-000000000003"
	reviewerRoleID    = "01988000-0004-7000-8000-000000000004"
	testerRoleID      = "01988000-0005-7000-8000-000000000005"

	starterWorkflowID  = "01988001-0000-7000-8000-000000000001"
	starterStepOneID   = "01988002-0001-7000-8000-000000000001"
	starterStepTwoID   = "01988002-0002-7000-8000-000000000002"
	starterStepThreeID = "01988002-0003-7000-8000-000000000003"
)

type seededPoint struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type seededSpend struct {
	Input  int `json:"input"`
	Cached int `json:"cached"`
	Output int `json:"output"`
}

type seededRemoteIDs struct {
	WorkflowID string            `json:"workflowId"`
	StepIDs    map[string]string `json:"stepIds"`
}

type seededStepSpec struct {
	Effort          string   `json:"effort"`
	Instructions    []string `json:"instructions"`
	OutputStructure string   `json:"outputStructure"`
	PauseForReview  bool     `json:"pauseForReview"`
}

type seededStep struct {
	ID        string           `json:"id"`
	Title     string           `json:"title"`
	Prompt    string           `json:"prompt"`
	State     string           `json:"state"`
	Position  seededPoint      `json:"position"`
	DependsOn []string         `json:"dependsOn"`
	RoleID    string           `json:"roleId"`
	Spec      *seededStepSpec  `json:"spec"`
	Values    *json.RawMessage `json:"values"`
	AgentID   string           `json:"agentId"`
	Run       *json.RawMessage `json:"run"`
	Report    *json.RawMessage `json:"report"`
}

type seededWorkflow struct {
	ID         string           `json:"id"`
	Name       string           `json:"name"`
	CreatedAt  string           `json:"createdAt"`
	RailRank   *int             `json:"railRank"`
	Locked     bool             `json:"locked"`
	Started    bool             `json:"started"`
	Cancelled  bool             `json:"cancelled"`
	Paused     bool             `json:"paused"`
	ProjectDir string           `json:"projectDir"`
	Steps      []seededStep     `json:"steps"`
	Remote     *seededRemoteIDs `json:"remote"`
	Spent      *seededSpend     `json:"spent"`
	CostUsd    *float64         `json:"costUsd"`
	Priced     *bool            `json:"priced"`
	StartedAt  string           `json:"startedAt"`
	FinishedAt string           `json:"finishedAt"`
}

func openSeeded(t *testing.T) input_itf.Storage {
	t.Helper()

	store, err := New(filepath.Join(t.TempDir(), "harness.db"))
	if err != nil {
		t.Fatalf("open storage: %v", err)
	}

	return store
}

func seededRolesByID(t *testing.T, store input_itf.Storage) map[string]*input_itf.RoleEntity {
	t.Helper()

	listed, err := store.RoleStore().List()
	if err != nil {
		t.Fatalf("list roles: %v", err)
	}

	byID := make(map[string]*input_itf.RoleEntity, len(listed))
	for _, role := range listed {
		byID[role.ID.String()] = role
	}

	if len(byID) != len(listed) {
		t.Fatalf("listed %d roles but only %d distinct ids", len(listed), len(byID))
	}

	return byID
}

func seededStarterWorkflow(t *testing.T, store input_itf.Storage) seededWorkflow {
	t.Helper()

	drafts, err := store.DraftStore().List()
	if err != nil {
		t.Fatalf("list drafts: %v", err)
	}

	if len(drafts) != 1 {
		t.Fatalf("a fresh database holds %d workflow drafts, want 1", len(drafts))
	}

	decoder := json.NewDecoder(strings.NewReader(drafts[0].Doc))
	decoder.DisallowUnknownFields()

	doc := seededWorkflow{}
	if err := decoder.Decode(&doc); err != nil {
		t.Fatalf("the starter draft does not decode into the frontend Workflow shape: %v", err)
	}

	if drafts[0].ID.String() != doc.ID {
		t.Fatalf("draft row id %s does not match the doc's id %s", drafts[0].ID, doc.ID)
	}

	return doc
}

func TestFreshDatabaseSeedsTheFiveDefaultRoles(t *testing.T) {
	cases := []struct {
		name           string
		id             string
		roleName       string
		effort         enums.Effort
		pauseForReview bool
		inputs         int
	}{
		{name: "explorer", id: explorerRoleID, roleName: "Explorer", effort: enums.EffortQuick},
		{name: "planner", id: plannerRoleID, roleName: "Planner", effort: enums.EffortStandard},
		{name: "implementer", id: implementerRoleID, roleName: "Implementer", effort: enums.EffortDeep, inputs: 1},
		{name: "reviewer", id: reviewerRoleID, roleName: "Reviewer", effort: enums.EffortDeep, pauseForReview: true},
		{name: "tester", id: testerRoleID, roleName: "Tester", effort: enums.EffortStandard},
	}

	byID := seededRolesByID(t, openSeeded(t))

	if len(byID) != len(cases) {
		t.Fatalf("a fresh database holds %d roles, want %d", len(byID), len(cases))
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			role, ok := byID[tc.id]
			if !ok {
				t.Fatalf("no seeded role carries id %s", tc.id)
			}

			if role.Name != tc.roleName {
				t.Fatalf("role %s is named %q, want %q", tc.id, role.Name, tc.roleName)
			}

			if role.Effort != tc.effort {
				t.Fatalf("role %q has effort %q, want %q", role.Name, role.Effort, tc.effort)
			}

			if role.PauseForReview != tc.pauseForReview {
				t.Fatalf("role %q has pause_for_review %t, want %t",
					role.Name, role.PauseForReview, tc.pauseForReview)
			}

			if len(role.Instructions) == 0 {
				t.Fatalf("role %q ships without a single instruction block", role.Name)
			}

			for key, block := range role.Instructions {
				if key == "" || strings.TrimSpace(block) == "" {
					t.Fatalf("role %q has an empty instruction block under key %q", role.Name, key)
				}
			}

			if len(role.Inputs) != tc.inputs {
				t.Fatalf("role %q declares %d inputs, want %d", role.Name, len(role.Inputs), tc.inputs)
			}
		})
	}
}

func TestSeededImplementerDeclaresTheChangeInput(t *testing.T) {
	byID := seededRolesByID(t, openSeeded(t))

	implementer, ok := byID[implementerRoleID]
	if !ok {
		t.Fatalf("no seeded role carries id %s", implementerRoleID)
	}

	change, ok := implementer.Inputs["change"]
	if !ok {
		t.Fatalf("Implementer's inputs are %v, want one keyed \"change\"", reflect.ValueOf(implementer.Inputs).MapKeys())
	}

	if change == nil {
		t.Fatal("Implementer's \"change\" input decoded as null")
	}

	if change.Type != enums.TextareaInput.String() {
		t.Fatalf("the \"change\" input has type %q, want %q", change.Type, enums.TextareaInput)
	}

	if !change.Required {
		t.Fatal("the \"change\" input is optional, want required")
	}
}

func TestSeededRolesPassTheValidationTheAppApplies(t *testing.T) {
	byID := seededRolesByID(t, openSeeded(t))

	for id, role := range byID {
		t.Run(role.Name, func(t *testing.T) {
			if err := helpers.ValidateStruct(role); err != nil {
				t.Fatalf("seeded role %q (%s) would be refused by the app: %v", role.Name, id, err)
			}

			if !role.Effort.Valid() {
				t.Fatalf("seeded role %q has effort %q, which is not an Effort", role.Name, role.Effort)
			}

			for key, input := range role.Inputs {
				if !enums.InputType(input.Type).Valid() {
					t.Fatalf("input %q of role %q has type %q, which is not an InputType",
						key, role.Name, input.Type)
				}
			}
		})
	}
}

func TestFreshDatabaseSeedsTheStarterWorkflowDraft(t *testing.T) {
	doc := seededStarterWorkflow(t, openSeeded(t))

	if doc.ID != starterWorkflowID {
		t.Fatalf("starter workflow id = %s, want %s", doc.ID, starterWorkflowID)
	}

	if doc.Name != "Start here" {
		t.Fatalf("starter workflow name = %q, want %q", doc.Name, "Start here")
	}

	if doc.ProjectDir != "" {
		t.Fatalf("starter workflow projectDir = %q, want it empty", doc.ProjectDir)
	}

	if _, err := time.Parse(time.RFC3339, doc.CreatedAt); err != nil {
		t.Fatalf("starter workflow createdAt %q is not RFC3339: %v", doc.CreatedAt, err)
	}

	cases := []struct {
		name      string
		id        string
		roleID    string
		dependsOn []string
	}{
		{name: "explore", id: starterStepOneID, roleID: explorerRoleID, dependsOn: []string{}},
		{name: "plan", id: starterStepTwoID, roleID: plannerRoleID, dependsOn: []string{starterStepOneID}},
		{name: "review", id: starterStepThreeID, roleID: reviewerRoleID, dependsOn: []string{starterStepTwoID}},
	}

	if len(doc.Steps) != len(cases) {
		t.Fatalf("starter workflow holds %d steps, want %d", len(doc.Steps), len(cases))
	}

	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			step := doc.Steps[i]

			if step.ID != tc.id {
				t.Fatalf("step %d id = %s, want %s", i, step.ID, tc.id)
			}

			if step.RoleID != tc.roleID {
				t.Fatalf("step %s roleId = %s, want %s", step.ID, step.RoleID, tc.roleID)
			}

			if !reflect.DeepEqual(step.DependsOn, tc.dependsOn) {
				t.Fatalf("step %s dependsOn = %v, want %v", step.ID, step.DependsOn, tc.dependsOn)
			}

			if step.State != "idle" {
				t.Fatalf("step %s state = %q, want %q", step.ID, step.State, "idle")
			}
		})
	}
}

func TestEverySeededStepStartsFromASeededRole(t *testing.T) {
	store := openSeeded(t)

	byID := seededRolesByID(t, store)
	doc := seededStarterWorkflow(t, store)

	for _, step := range doc.Steps {
		if step.RoleID == "" {
			t.Fatalf("step %s starts from no role at all", step.ID)
		}

		if _, ok := byID[step.RoleID]; !ok {
			t.Fatalf("step %s starts from role %s, which the role seed never inserted",
				step.ID, step.RoleID)
		}
	}
}

func TestStarterWorkflowIsSkippedWhenADraftAlreadyExists(t *testing.T) {
	path := filepath.Join(t.TempDir(), "harness.db")

	cut := len(migrations)
	for i, statement := range migrations {
		if statement == seedRoles {
			cut = i
			break
		}
	}

	if cut == len(migrations) {
		t.Fatal("the role seed is not in the migration list")
	}

	staged, err := sql.Open(driverName, dsn(path))
	if err != nil {
		t.Fatalf("open staged database: %v", err)
	}

	for i := 0; i < cut; i++ {
		if _, err := staged.Exec(migrations[i]); err != nil {
			staged.Close()
			t.Fatalf("staged migration %d: %v", i+1, err)
		}
	}

	// user_version is how migrate() knows where to resume, so stamping it here
	// makes the reopen below run exactly the two seed entries and nothing else.
	if _, err := staged.Exec(fmt.Sprintf(`PRAGMA user_version = %d`, cut)); err != nil {
		staged.Close()
		t.Fatalf("stamp user_version: %v", err)
	}

	const existingID = "01988009-0000-7000-8000-000000000009"

	if _, err := staged.Exec(`INSERT INTO workflow_drafts (id, doc, updated_at) VALUES (?, ?, ?)`,
		existingID, `{"id":"`+existingID+`","name":"Mine","steps":[]}`,
		formatTime(time.Now()),
	); err != nil {
		staged.Close()
		t.Fatalf("insert the pre-existing draft: %v", err)
	}

	if err := staged.Close(); err != nil {
		t.Fatalf("close staged database: %v", err)
	}

	store, err := New(path)
	if err != nil {
		t.Fatalf("reopen storage: %v", err)
	}

	drafts, err := store.DraftStore().List()
	if err != nil {
		t.Fatalf("list drafts: %v", err)
	}

	if len(drafts) != 1 || drafts[0].ID.String() != existingID {
		ids := make([]string, 0, len(drafts))
		for _, draft := range drafts {
			ids = append(ids, draft.ID.String())
		}

		t.Fatalf("drafts after the seed ran = %v, want only the pre-existing %s", ids, existingID)
	}

	roles, err := store.RoleStore().List()
	if err != nil {
		t.Fatalf("list roles: %v", err)
	}

	if len(roles) != 5 {
		t.Fatalf("the role seed inserted %d roles alongside the skipped workflow, want 5", len(roles))
	}
}

func TestReopeningDoesNotDuplicateTheSeededRolesOrDraft(t *testing.T) {
	path := filepath.Join(t.TempDir(), "harness.db")

	for _, pass := range []string{"first open", "reopen"} {
		store, err := New(path)
		if err != nil {
			t.Fatalf("%s: open storage: %v", pass, err)
		}

		roles, err := store.RoleStore().List()
		if err != nil {
			t.Fatalf("%s: list roles: %v", pass, err)
		}

		if len(roles) != 5 {
			t.Fatalf("%s: holds %d roles, want 5", pass, len(roles))
		}

		drafts, err := store.DraftStore().List()
		if err != nil {
			t.Fatalf("%s: list drafts: %v", pass, err)
		}

		if len(drafts) != 1 {
			t.Fatalf("%s: holds %d workflow drafts, want 1", pass, len(drafts))
		}
	}
}

func TestAnExistingLibraryIsLeftAlone(t *testing.T) {
	path := filepath.Join(t.TempDir(), "harness.db")

	cut := len(migrations)
	for i, statement := range migrations {
		if statement == seedRoles {
			cut = i
			break
		}
	}

	if cut == len(migrations) {
		t.Fatal("the role seed is not in the migration list")
	}

	staged, err := sql.Open(driverName, dsn(path))
	if err != nil {
		t.Fatalf("open staged database: %v", err)
	}

	for i := 0; i < cut; i++ {
		if _, err := staged.Exec(migrations[i]); err != nil {
			staged.Close()
			t.Fatalf("staged migration %d: %v", i+1, err)
		}
	}

	if _, err := staged.Exec(fmt.Sprintf(`PRAGMA user_version = %d`, cut)); err != nil {
		staged.Close()
		t.Fatalf("stamp user_version: %v", err)
	}

	const mine = "01988009-0000-7000-8000-000000000009"

	if _, err := staged.Exec(`INSERT INTO roles (`+roleColumns+`)
		VALUES (?, ?, ?, ?, 0, 0, '{}', json_object('base', 'b'), 'o', ?, ?)`,
		mine, "Mine", "A role this install already had", "quick",
		formatTime(time.Now()), formatTime(time.Now()),
	); err != nil {
		staged.Close()
		t.Fatalf("insert the pre-existing role: %v", err)
	}

	if err := staged.Close(); err != nil {
		t.Fatalf("close staged database: %v", err)
	}

	store, err := New(path)
	if err != nil {
		t.Fatalf("reopen storage: %v", err)
	}

	roles, err := store.RoleStore().List()
	if err != nil {
		t.Fatalf("list roles: %v", err)
	}

	if len(roles) != 1 || roles[0].ID.String() != mine {
		names := make([]string, 0, len(roles))
		for _, role := range roles {
			names = append(names, role.Name)
		}

		t.Fatalf("roles after the seed ran = %v, want only the pre-existing one", names)
	}
}
