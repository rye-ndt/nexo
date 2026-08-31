package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"hexago/internal/helpers/enums"
	input_itf "hexago/internal/interface/input"

	"github.com/google/uuid"
)

func TestRoleRoundTripsPauseForReview(t *testing.T) {
	path := filepath.Join(t.TempDir(), "harness.db")

	store, err := New(path)
	if err != nil {
		t.Fatalf("open storage: %v", err)
	}

	roles := store.RoleStore()

	gated := &input_itf.RoleEntity{
		ID:             uuid.New(),
		Name:           "planner",
		Effort:         enums.EffortDeep,
		PauseForReview: true,
		Instructions:   map[string]string{"base": "plan carefully"},
	}

	plain := &input_itf.RoleEntity{
		ID:           uuid.New(),
		Name:         "implementer",
		Effort:       enums.EffortStandard,
		Instructions: map[string]string{"base": "write code"},
	}

	for _, role := range []*input_itf.RoleEntity{gated, plain} {
		if err := roles.Upsert(role); err != nil {
			t.Fatalf("upsert %s: %v", role.Name, err)
		}
	}

	found, err := roles.Find(gated.ID)
	if err != nil {
		t.Fatalf("find gated role: %v", err)
	}

	if !found.PauseForReview {
		t.Fatal("gated role lost pause_for_review on the round trip")
	}

	found, err = roles.Find(plain.ID)
	if err != nil {
		t.Fatalf("find plain role: %v", err)
	}

	if found.PauseForReview {
		t.Fatal("plain role gained pause_for_review on the round trip")
	}

	gated.PauseForReview = false
	if err := roles.Upsert(gated); err != nil {
		t.Fatalf("re-upsert gated role: %v", err)
	}

	found, err = roles.Find(gated.ID)
	if err != nil {
		t.Fatalf("re-find gated role: %v", err)
	}

	if found.PauseForReview {
		t.Fatal("clearing pause_for_review did not persist")
	}

	listed, err := roles.List()
	if err != nil {
		t.Fatalf("list roles: %v", err)
	}

	own := 0
	for _, role := range listed {
		if role.ID == gated.ID || role.ID == plain.ID {
			own++
		}
	}

	if own != 2 {
		t.Fatalf("listed %d of the two roles this test wrote, want 2", own)
	}
}

func TestRolesListNewestUpdatedFirst(t *testing.T) {
	path := filepath.Join(t.TempDir(), "harness.db")

	store, err := New(path)
	if err != nil {
		t.Fatalf("open storage: %v", err)
	}

	roles := store.RoleStore()
	base := time.Date(2031, 4, 5, 6, 7, 8, 0, time.UTC)

	written := []*input_itf.RoleEntity{
		{Name: "oldest", UpdatedAt: base},
		{Name: "beta", UpdatedAt: base.Add(time.Hour)},
		{Name: "alpha", UpdatedAt: base.Add(time.Hour)},
		{Name: "newest", UpdatedAt: base.Add(2 * time.Hour)},
	}

	ids := map[uuid.UUID]bool{}

	for _, role := range written {
		role.ID = uuid.New()
		role.Effort = enums.EffortStandard
		role.Instructions = map[string]string{"base": "do the work"}
		ids[role.ID] = true

		if err := roles.Upsert(role); err != nil {
			t.Fatalf("upsert role %s: %v", role.Name, err)
		}
	}

	listed, err := roles.List()
	if err != nil {
		t.Fatalf("list roles: %v", err)
	}

	order := []string{}

	for _, role := range listed {
		if ids[role.ID] {
			order = append(order, role.Name)
		}
	}

	want := []string{"newest", "alpha", "beta", "oldest"}
	if !reflect.DeepEqual(order, want) {
		t.Fatalf("listed roles as %v, want %v", order, want)
	}
}

func TestMigrationsAreIdempotentAcrossReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "harness.db")

	first, err := New(path)
	if err != nil {
		t.Fatalf("open storage: %v", err)
	}

	role := &input_itf.RoleEntity{
		ID:             uuid.New(),
		Name:           "planner",
		Effort:         enums.EffortDeep,
		PauseForReview: true,
		Instructions:   map[string]string{"base": "plan carefully"},
	}

	if err := first.RoleStore().Upsert(role); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	second, err := New(path)
	if err != nil {
		t.Fatalf("reopen storage: %v", err)
	}

	found, err := second.RoleStore().Find(role.ID)
	if err != nil {
		t.Fatalf("find after reopen: %v", err)
	}

	if !found.PauseForReview {
		t.Fatal("pause_for_review did not survive a reopen")
	}
}

func TestMCPCredentialsRoundTripAccountAndDelete(t *testing.T) {
	path := filepath.Join(t.TempDir(), "harness.db")

	store, err := New(path)
	if err != nil {
		t.Fatalf("open storage: %v", err)
	}

	mcps := store.MCPStore()

	entity := &input_itf.MCPEntity{
		Name:              "atlassian",
		EncryptedOAuthKey: "cipher",
		Account:           "rye@nexo.dev",
		ExpiredAt:         time.Now().Add(time.Hour),
	}

	if err := mcps.UpsertCredentials(entity); err != nil {
		t.Fatalf("upsert credentials: %v", err)
	}

	found, err := mcps.GetCredentials(entity.Name)
	if err != nil {
		t.Fatalf("get credentials: %v", err)
	}

	if found.Account != entity.Account {
		t.Fatalf("account = %q, want %q", found.Account, entity.Account)
	}

	authenticated, err := mcps.ListAuthenticated()
	if err != nil {
		t.Fatalf("list authenticated: %v", err)
	}

	if len(authenticated) != 1 || authenticated[0].Account != entity.Account {
		t.Fatalf("list authenticated = %+v, want one row carrying the account", authenticated)
	}

	if err := mcps.DeleteCredentials(entity.Name); err != nil {
		t.Fatalf("delete credentials: %v", err)
	}

	gone, err := mcps.GetCredentials(entity.Name)
	if err != nil {
		t.Fatalf("get credentials after delete: %v", err)
	}

	if gone != nil {
		t.Fatalf("credentials survived the delete: %+v", gone)
	}

	if err := mcps.DeleteCredentials(entity.Name); err != nil {
		t.Fatalf("second delete should be a no-op, got: %v", err)
	}
}

// Step writes arrive concurrently from the coordinator, the heartbeat watcher and
// every agent's report. On the default rollback journal with no busy timeout the
// second writer of an overlapping pair fails immediately with SQLITE_BUSY, which
// makes WorkflowManager roll back a report that already happened.
func TestOverlappingStepWritesDoNotHitSQLiteBusy(t *testing.T) {
	// A space, matching the real macOS "Application Support" data dir.
	dir := filepath.Join(t.TempDir(), "Application Support")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("make data dir: %v", err)
	}

	store, err := New(filepath.Join(dir, "harness.db"))
	if err != nil {
		t.Fatalf("open storage: %v", err)
	}

	steps := store.StepStore()
	workflow := &input_itf.WorkflowEntity{ID: uuid.New(), CreatedAt: time.Now()}

	const writers, each = 8, 50

	errs := make(chan error, writers*each)
	start := make(chan struct{})

	var wg sync.WaitGroup

	for i := 0; i < writers; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			step := &input_itf.StepEntity{
				ID:         uuid.New(),
				WorkflowID: workflow.ID,
				Status:     enums.StepProcessing,
			}

			<-start

			for j := 0; j < each; j++ {
				if err := steps.SaveStepHistory(
					[]*input_itf.WorkflowEntity{workflow},
					[]*input_itf.StepEntity{step},
					nil,
				); err != nil {
					errs <- err
				}
			}
		}()
	}

	close(start)
	wg.Wait()
	close(errs)

	failed := 0
	for err := range errs {
		if failed == 0 {
			t.Errorf("concurrent write failed: %v", err)
		}

		failed++
	}

	if failed > 0 {
		t.Fatalf("%d of %d concurrent writes failed", failed, writers*each)
	}
}

func TestLoadStepHistoryRoundTripsWorkflowsStepsAndReports(t *testing.T) {
	store, err := New(filepath.Join(t.TempDir(), "harness.db"))
	if err != nil {
		t.Fatalf("open storage: %v", err)
	}

	steps := store.StepStore()
	base := time.Now().UTC().Truncate(time.Second)

	first := &input_itf.WorkflowEntity{
		ID:             uuid.New(),
		ProjectDirPath: "/repo/nexo",
		StartedAt:      base,
		CompletedAt:    base.Add(10 * time.Minute),
		TotalStep:      2,
		TotalRetry:     1,
		CreatedAt:      base,
		UpdatedAt:      base.Add(10 * time.Minute),
	}

	second := &input_itf.WorkflowEntity{
		ID:             uuid.New(),
		ProjectDirPath: "/repo/other",
		StartedAt:      base.Add(time.Hour),
		TotalStep:      1,
		CreatedAt:      base.Add(time.Hour),
		UpdatedAt:      base.Add(time.Hour),
	}

	planner := &input_itf.StepEntity{
		ID:             uuid.New(),
		WorkflowID:     first.ID,
		Name:           "plan the slice",
		Effort:         enums.EffortDeep,
		PreferredModel: enums.Opus,
		ThinkingLevel:  enums.HighThinking,
		Instructions:   []string{"be terse", "no new comments"},
		AutoRetry:      true,
		PauseForReview: true,
		ExtraGuidance:  "read CLAUDE.md first",
		RetryCount:     2,
		Status:         enums.StepAwaitingReview,
		CreatedAt:      base,
		UpdatedAt:      base.Add(time.Minute),
	}

	implementer := &input_itf.StepEntity{
		ID:               uuid.New(),
		WorkflowID:       first.ID,
		Name:             "write the code",
		Effort:           enums.EffortStandard,
		PreferredModel:   enums.Sonnet,
		ThinkingLevel:    enums.MedThinking,
		Status:           enums.StepNotTaken,
		DependsOnStepIDs: uuid.UUIDs{planner.ID},
		CreatedAt:        base.Add(time.Second),
		UpdatedAt:        base.Add(time.Second),
	}

	stray := &input_itf.StepEntity{
		ID:         uuid.New(),
		WorkflowID: second.ID,
		Name:       "unrelated",
		Status:     enums.StepCompleted,
		CreatedAt:  base.Add(time.Hour),
		UpdatedAt:  base.Add(time.Hour),
	}

	failed := &input_itf.StepResultEntity{
		ID:            uuid.New(),
		StepID:        planner.ID,
		AgentID:       uuid.New(),
		AttemptStatus: enums.StepFailed,
		Handoffs: []*input_itf.HandoffEntity{{
			Step:     "plan the slice",
			TLDR:     "ran out of context",
			Blockers: map[string]string{"context": "window filled"},
		}},
		ContextUsage: &input_itf.ContextUsage{Total: 200000, Used: 199000, Billed: 4000, Input: 90000, Cached: 105000},
		StartedAt:    base,
		CompletedAt:  base.Add(30 * time.Second),
		CreatedAt:    base.Add(30 * time.Second),
		UpdatedAt:    base.Add(30 * time.Second),
	}

	accepted := &input_itf.StepResultEntity{
		ID:            uuid.New(),
		StepID:        planner.ID,
		AgentID:       uuid.New(),
		AttemptStatus: enums.StepAwaitingReview,
		Handoffs: []*input_itf.HandoffEntity{{
			Step:              "plan the slice",
			TLDR:              "plan is three queries",
			Outcome:           "ready for review",
			ApprovedDecisions: map[string]string{"reports": "oldest first"},
			ChangedBehaviors:  map[string]string{"load": "stitches in memory"},
			KnownGaps:         map[string]string{"drafts": "untouched"},
		}},
		ContextUsage: &input_itf.ContextUsage{Total: 200000, Used: 51000, Billed: 8000, Input: 30000, Cached: 12000},
		StartedAt:    base.Add(time.Minute),
		CompletedAt:  base.Add(2 * time.Minute),
		CreatedAt:    base.Add(2 * time.Minute),
		UpdatedAt:    base.Add(2 * time.Minute),
	}

	planner.LastReportID = accepted.ID

	if err := steps.SaveStepHistory(
		[]*input_itf.WorkflowEntity{first, second},
		[]*input_itf.StepEntity{planner, implementer, stray},
		[]*input_itf.StepResultEntity{failed, accepted},
	); err != nil {
		t.Fatalf("save step history: %v", err)
	}

	snapshots, err := steps.LoadStepHistory()
	if err != nil {
		t.Fatalf("load step history: %v", err)
	}

	if len(snapshots) != 2 {
		t.Fatalf("loaded %d snapshots, want 2", len(snapshots))
	}

	if snapshots[0].Workflow.ID != first.ID || snapshots[1].Workflow.ID != second.ID {
		t.Fatalf("snapshots came back out of created_at order: %s then %s",
			snapshots[0].Workflow.ID, snapshots[1].Workflow.ID)
	}

	gotWorkflow := snapshots[0].Workflow

	if gotWorkflow.ProjectDirPath != first.ProjectDirPath {
		t.Fatalf("workflow path = %q, want %q", gotWorkflow.ProjectDirPath, first.ProjectDirPath)
	}

	if gotWorkflow.TotalStep != first.TotalStep || gotWorkflow.TotalRetry != first.TotalRetry {
		t.Fatalf("workflow totals = %d / %d, want %d / %d",
			gotWorkflow.TotalStep, gotWorkflow.TotalRetry, first.TotalStep, first.TotalRetry)
	}

	if !gotWorkflow.StartedAt.Equal(first.StartedAt) || !gotWorkflow.CompletedAt.Equal(first.CompletedAt) {
		t.Fatalf("workflow window = %s..%s, want %s..%s",
			gotWorkflow.StartedAt, gotWorkflow.CompletedAt, first.StartedAt, first.CompletedAt)
	}

	if !gotWorkflow.CreatedAt.Equal(first.CreatedAt) || !gotWorkflow.UpdatedAt.Equal(first.UpdatedAt) {
		t.Fatalf("workflow timestamps = %s / %s, want %s / %s",
			gotWorkflow.CreatedAt, gotWorkflow.UpdatedAt, first.CreatedAt, first.UpdatedAt)
	}

	if len(snapshots[0].Steps) != 2 {
		t.Fatalf("first workflow loaded %d steps, want 2", len(snapshots[0].Steps))
	}

	if len(snapshots[1].Steps) != 1 || snapshots[1].Steps[0].ID != stray.ID {
		t.Fatalf("second workflow steps = %+v, want only %s", snapshots[1].Steps, stray.ID)
	}

	if len(snapshots[1].Reports) != 0 {
		t.Fatalf("second workflow gained %d reports", len(snapshots[1].Reports))
	}

	gotStep := snapshots[0].Steps[0]

	if gotStep.ID != planner.ID {
		t.Fatalf("steps came back out of created_at order: first is %s, want %s", gotStep.ID, planner.ID)
	}

	if gotStep.WorkflowID != planner.WorkflowID || gotStep.Name != planner.Name {
		t.Fatalf("step identity = %s / %q, want %s / %q",
			gotStep.WorkflowID, gotStep.Name, planner.WorkflowID, planner.Name)
	}

	if gotStep.Effort != planner.Effort || gotStep.PreferredModel != planner.PreferredModel ||
		gotStep.ThinkingLevel != planner.ThinkingLevel {
		t.Fatalf("step enums = %s / %s / %s, want %s / %s / %s",
			gotStep.Effort, gotStep.PreferredModel, gotStep.ThinkingLevel,
			planner.Effort, planner.PreferredModel, planner.ThinkingLevel)
	}

	if !reflect.DeepEqual(gotStep.Instructions, planner.Instructions) {
		t.Fatalf("step system prompts = %#v, want %#v", gotStep.Instructions, planner.Instructions)
	}

	if !gotStep.AutoRetry || !gotStep.PauseForReview {
		t.Fatalf("step flags = auto_retry %t / manual_accept %t, want both true",
			gotStep.AutoRetry, gotStep.PauseForReview)
	}

	if gotStep.ExtraGuidance != planner.ExtraGuidance || gotStep.RetryCount != planner.RetryCount {
		t.Fatalf("step guidance = %q / retries %d, want %q / %d",
			gotStep.ExtraGuidance, gotStep.RetryCount, planner.ExtraGuidance, planner.RetryCount)
	}

	if gotStep.Status != enums.StepAwaitingReview || gotStep.LastReportID != accepted.ID {
		t.Fatalf("step status = %s / last report %s, want %s / %s",
			gotStep.Status, gotStep.LastReportID, enums.StepAwaitingReview, accepted.ID)
	}

	if !gotStep.CreatedAt.Equal(planner.CreatedAt) || !gotStep.UpdatedAt.Equal(planner.UpdatedAt) {
		t.Fatalf("step timestamps = %s / %s, want %s / %s",
			gotStep.CreatedAt, gotStep.UpdatedAt, planner.CreatedAt, planner.UpdatedAt)
	}

	gotDependant := snapshots[0].Steps[1]

	if !reflect.DeepEqual(gotDependant.DependsOnStepIDs, uuid.UUIDs{planner.ID}) {
		t.Fatalf("depends on = %v, want %v", gotDependant.DependsOnStepIDs, uuid.UUIDs{planner.ID})
	}

	if gotDependant.Status != enums.StepNotTaken {
		t.Fatalf("dependant status = %s, want %s", gotDependant.Status, enums.StepNotTaken)
	}

	if len(snapshots[0].Reports) != 2 {
		t.Fatalf("first workflow loaded %d reports, want 2", len(snapshots[0].Reports))
	}

	gotFailed, gotAccepted := snapshots[0].Reports[0], snapshots[0].Reports[1]

	if gotFailed.ID != failed.ID || gotAccepted.ID != accepted.ID {
		t.Fatalf("reports came back as %s then %s, want %s then %s",
			gotFailed.ID, gotAccepted.ID, failed.ID, accepted.ID)
	}

	if gotAccepted.StepID != accepted.StepID || gotAccepted.AgentID != accepted.AgentID {
		t.Fatalf("report owners = %s / %s, want %s / %s",
			gotAccepted.StepID, gotAccepted.AgentID, accepted.StepID, accepted.AgentID)
	}

	if gotFailed.AttemptStatus != enums.StepFailed || gotAccepted.AttemptStatus != enums.StepAwaitingReview {
		t.Fatalf("attempt statuses = %s / %s, want %s / %s",
			gotFailed.AttemptStatus, gotAccepted.AttemptStatus, enums.StepFailed, enums.StepAwaitingReview)
	}

	if !reflect.DeepEqual(gotAccepted.Handoffs, accepted.Handoffs) {
		t.Fatalf("handoffs = %#v, want %#v", gotAccepted.Handoffs[0], accepted.Handoffs[0])
	}

	if !reflect.DeepEqual(gotAccepted.ContextUsage, accepted.ContextUsage) {
		t.Fatalf("context usage = %+v, want %+v", gotAccepted.ContextUsage, accepted.ContextUsage)
	}

	if !gotAccepted.StartedAt.Equal(accepted.StartedAt) || !gotAccepted.CompletedAt.Equal(accepted.CompletedAt) {
		t.Fatalf("report window = %s..%s, want %s..%s",
			gotAccepted.StartedAt, gotAccepted.CompletedAt, accepted.StartedAt, accepted.CompletedAt)
	}

	if !gotAccepted.CreatedAt.Equal(accepted.CreatedAt) || !gotAccepted.UpdatedAt.Equal(accepted.UpdatedAt) {
		t.Fatalf("report timestamps = %s / %s, want %s / %s",
			gotAccepted.CreatedAt, gotAccepted.UpdatedAt, accepted.CreatedAt, accepted.UpdatedAt)
	}
}

func TestLoadStepHistoryReturnsReportsOldestFirst(t *testing.T) {
	store, err := New(filepath.Join(t.TempDir(), "harness.db"))
	if err != nil {
		t.Fatalf("open storage: %v", err)
	}

	steps := store.StepStore()
	base := time.Now().UTC().Truncate(time.Second)

	workflow := &input_itf.WorkflowEntity{ID: uuid.New(), CreatedAt: base, UpdatedAt: base}
	step := &input_itf.StepEntity{
		ID:         uuid.New(),
		WorkflowID: workflow.ID,
		Status:     enums.StepCompleted,
		CreatedAt:  base,
		UpdatedAt:  base,
	}

	newest := &input_itf.StepResultEntity{
		ID:        uuid.New(),
		StepID:    step.ID,
		CreatedAt: base.Add(3 * time.Minute),
	}
	oldest := &input_itf.StepResultEntity{
		ID:        uuid.New(),
		StepID:    step.ID,
		CreatedAt: base.Add(time.Minute),
	}
	middle := &input_itf.StepResultEntity{
		ID:        uuid.New(),
		StepID:    step.ID,
		CreatedAt: base.Add(2 * time.Minute),
	}

	if err := steps.SaveStepHistory(
		[]*input_itf.WorkflowEntity{workflow},
		[]*input_itf.StepEntity{step},
		[]*input_itf.StepResultEntity{newest, oldest, middle},
	); err != nil {
		t.Fatalf("save step history: %v", err)
	}

	snapshots, err := steps.LoadStepHistory()
	if err != nil {
		t.Fatalf("load step history: %v", err)
	}

	if len(snapshots) != 1 {
		t.Fatalf("loaded %d snapshots, want 1", len(snapshots))
	}

	want := []uuid.UUID{oldest.ID, middle.ID, newest.ID}
	got := make([]uuid.UUID, 0, len(snapshots[0].Reports))

	for _, r := range snapshots[0].Reports {
		got = append(got, r.ID)
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("reports came back as %v, want oldest first %v", got, want)
	}
}

func TestLoadStepHistoryKeepsAMissingContextUsageNil(t *testing.T) {
	store, err := New(filepath.Join(t.TempDir(), "harness.db"))
	if err != nil {
		t.Fatalf("open storage: %v", err)
	}

	steps := store.StepStore()
	base := time.Now().UTC().Truncate(time.Second)

	workflow := &input_itf.WorkflowEntity{ID: uuid.New(), CreatedAt: base, UpdatedAt: base}
	step := &input_itf.StepEntity{
		ID:         uuid.New(),
		WorkflowID: workflow.ID,
		Status:     enums.StepFailed,
		CreatedAt:  base,
		UpdatedAt:  base,
	}

	report := &input_itf.StepResultEntity{
		ID:            uuid.New(),
		StepID:        step.ID,
		AttemptStatus: enums.StepFailed,
		CreatedAt:     base,
	}

	if err := steps.SaveStepHistory(
		[]*input_itf.WorkflowEntity{workflow},
		[]*input_itf.StepEntity{step},
		[]*input_itf.StepResultEntity{report},
	); err != nil {
		t.Fatalf("save step history: %v", err)
	}

	snapshots, err := steps.LoadStepHistory()
	if err != nil {
		t.Fatalf("load step history: %v", err)
	}

	if len(snapshots) != 1 || len(snapshots[0].Reports) != 1 {
		t.Fatalf("loaded %d snapshots, want 1 carrying 1 report", len(snapshots))
	}

	if snapshots[0].Reports[0].ContextUsage != nil {
		t.Fatalf("context usage came back as %+v, want nil", snapshots[0].Reports[0].ContextUsage)
	}
}

func TestLoadStepHistoryOnAnEmptyDatabase(t *testing.T) {
	store, err := New(filepath.Join(t.TempDir(), "harness.db"))
	if err != nil {
		t.Fatalf("open storage: %v", err)
	}

	snapshots, err := store.StepStore().LoadStepHistory()
	if err != nil {
		t.Fatalf("load step history: %v", err)
	}

	if len(snapshots) != 0 {
		t.Fatalf("loaded %d snapshots from an empty database, want 0", len(snapshots))
	}
}

func TestDSNKeepsAWindowsDriveOutOfTheURIAuthority(t *testing.T) {
	cases := []struct {
		name string
		path string
		want string
	}{
		{
			name: "a unix path is already rooted",
			path: "/home/rye/.local/share/nexo/harness.db",
			want: "file:///home/rye/.local/share/nexo/harness.db",
		},
		{
			name: "a windows path gains the third slash",
			path: "C:/Users/rye/AppData/Roaming/nexo/harness.db",
			want: "file:///C:/Users/rye/AppData/Roaming/nexo/harness.db",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := dsn(tc.path)

			base, query, found := strings.Cut(got, "?")
			if !found {
				t.Fatalf("dsn dropped the pragmas entirely: %s", got)
			}

			if base != tc.want {
				t.Fatalf("dsn(%q) = %q, want %q", tc.path, base, tc.want)
			}

			for _, pragma := range []string{"busy_timeout", "journal_mode"} {
				if !strings.Contains(query, pragma) {
					t.Fatalf("dsn(%q) dropped %s: %s", tc.path, pragma, query)
				}
			}
		})
	}
}

func TestAVersionPastTheListIsRebasedSoLaterMigrationsStillRun(t *testing.T) {
	path := filepath.Join(t.TempDir(), "harness.db")

	opened, err := New(path)
	if err != nil {
		t.Fatalf("open storage: %v", err)
	}

	store, ok := opened.(*litesql)
	if !ok {
		t.Fatalf("storage is %T, want *litesql", opened)
	}

	ahead := len(migrations) + 71

	if _, err := store.db.Exec(fmt.Sprintf(`PRAGMA user_version = %d`, ahead)); err != nil {
		t.Fatalf("stamp an old chain's version: %v", err)
	}

	version, err := alignedVersion(store.db)
	if err != nil {
		t.Fatalf("align version: %v", err)
	}

	if version != len(migrations) {
		t.Fatalf("aligned version = %d, want %d", version, len(migrations))
	}

	var stored int
	if err := store.db.QueryRow(`PRAGMA user_version`).Scan(&stored); err != nil {
		t.Fatalf("read back version: %v", err)
	}

	if stored != len(migrations) {
		t.Fatalf("stored user_version = %d, want %d: a migration added later would be skipped", stored, len(migrations))
	}
}

func TestAVersionInsideTheListIsLeftAlone(t *testing.T) {
	path := filepath.Join(t.TempDir(), "harness.db")

	opened, err := New(path)
	if err != nil {
		t.Fatalf("open storage: %v", err)
	}

	store, ok := opened.(*litesql)
	if !ok {
		t.Fatalf("storage is %T, want *litesql", opened)
	}

	if _, err := store.db.Exec(`PRAGMA user_version = 2`); err != nil {
		t.Fatalf("stamp a partial version: %v", err)
	}

	version, err := alignedVersion(store.db)
	if err != nil {
		t.Fatalf("align version: %v", err)
	}

	if version != 2 {
		t.Fatalf("aligned version = %d, want 2", version)
	}
}

func TestADatabaseFromBeforeTheCollapsedSchemaIsRefusedNotRebased(t *testing.T) {
	path := filepath.Join(t.TempDir(), "harness.db")

	old, err := sql.Open(driverName, dsn(path))
	if err != nil {
		t.Fatalf("open raw db: %v", err)
	}

	for _, statement := range []string{
		`CREATE TABLE sessions (id TEXT PRIMARY KEY, total_task INTEGER NOT NULL)`,
		`CREATE TABLE tasks (id TEXT PRIMARY KEY, session_id TEXT NOT NULL)`,
		`PRAGMA user_version = 33`,
	} {
		if _, err := old.Exec(statement); err != nil {
			t.Fatalf("build a v1.0.0 database: %v", err)
		}
	}

	if err := old.Close(); err != nil {
		t.Fatalf("close raw db: %v", err)
	}

	opened, openErr := New(path)
	if openErr == nil {
		t.Fatal("a pre-collapse database was accepted; every later query would fail on a missing table")
	}

	if opened != nil {
		t.Fatal("a refused database still handed back a store")
	}

	if !strings.Contains(openErr.Error(), "predates the collapsed schema") {
		t.Fatalf("error does not say what is wrong: %v", openErr)
	}

	reopened, err := sql.Open(driverName, dsn(path))
	if err != nil {
		t.Fatalf("reopen raw db: %v", err)
	}
	defer reopened.Close()

	var version int
	if err := reopened.QueryRow(`PRAGMA user_version`).Scan(&version); err != nil {
		t.Fatalf("read back version: %v", err)
	}

	if version != 33 {
		t.Fatalf("user_version = %d, want 33: rolling back to the previous build can no longer resume", version)
	}
}
