package storage

import (
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

func TestTemplateRoundTripsManualAcceptRequired(t *testing.T) {
	path := filepath.Join(t.TempDir(), "harness.db")

	store, err := New(path)
	if err != nil {
		t.Fatalf("open storage: %v", err)
	}

	templates := store.TemplateStore()

	gated := &input_itf.TemplateEntity{
		ID:                   uuid.New(),
		Name:                 "planner",
		TaskLevel:            enums.HeavyTask,
		ManualAcceptRequired: true,
		SystemPrompts:        map[string]string{"base": "plan carefully"},
	}

	plain := &input_itf.TemplateEntity{
		ID:            uuid.New(),
		Name:          "implementer",
		TaskLevel:     enums.DailyTask,
		SystemPrompts: map[string]string{"base": "write code"},
	}

	for _, template := range []*input_itf.TemplateEntity{gated, plain} {
		if err := templates.Upsert(template); err != nil {
			t.Fatalf("upsert %s: %v", template.Name, err)
		}
	}

	found, err := templates.Find(gated.ID)
	if err != nil {
		t.Fatalf("find gated template: %v", err)
	}

	if !found.ManualAcceptRequired {
		t.Fatal("gated template lost manual_accept_required on the round trip")
	}

	found, err = templates.Find(plain.ID)
	if err != nil {
		t.Fatalf("find plain template: %v", err)
	}

	if found.ManualAcceptRequired {
		t.Fatal("plain template gained manual_accept_required on the round trip")
	}

	gated.ManualAcceptRequired = false
	if err := templates.Upsert(gated); err != nil {
		t.Fatalf("re-upsert gated template: %v", err)
	}

	found, err = templates.Find(gated.ID)
	if err != nil {
		t.Fatalf("re-find gated template: %v", err)
	}

	if found.ManualAcceptRequired {
		t.Fatal("clearing manual_accept_required did not persist")
	}

	listed, err := templates.List()
	if err != nil {
		t.Fatalf("list templates: %v", err)
	}

	if len(listed) != 2 {
		t.Fatalf("listed %d templates, want 2", len(listed))
	}
}

func TestMigrationsAreIdempotentAcrossReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "harness.db")

	first, err := New(path)
	if err != nil {
		t.Fatalf("open storage: %v", err)
	}

	template := &input_itf.TemplateEntity{
		ID:                   uuid.New(),
		Name:                 "planner",
		TaskLevel:            enums.HeavyTask,
		ManualAcceptRequired: true,
		SystemPrompts:        map[string]string{"base": "plan carefully"},
	}

	if err := first.TemplateStore().Upsert(template); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	second, err := New(path)
	if err != nil {
		t.Fatalf("reopen storage: %v", err)
	}

	found, err := second.TemplateStore().Find(template.ID)
	if err != nil {
		t.Fatalf("find after reopen: %v", err)
	}

	if !found.ManualAcceptRequired {
		t.Fatal("manual_accept_required did not survive a reopen")
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

// Task writes arrive concurrently from the coordinator, the heartbeat watcher and
// every agent's report. On the default rollback journal with no busy timeout the
// second writer of an overlapping pair fails immediately with SQLITE_BUSY, which
// makes SessionManager roll back a report that already happened.
func TestOverlappingTaskWritesDoNotHitSQLiteBusy(t *testing.T) {
	// A space, matching the real macOS "Application Support" data dir.
	dir := filepath.Join(t.TempDir(), "Application Support")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("make data dir: %v", err)
	}

	store, err := New(filepath.Join(dir, "harness.db"))
	if err != nil {
		t.Fatalf("open storage: %v", err)
	}

	tasks := store.TaskStore()
	session := &input_itf.SessionEntity{ID: uuid.New(), CreatedAt: time.Now()}

	const writers, each = 8, 50

	errs := make(chan error, writers*each)
	start := make(chan struct{})

	var wg sync.WaitGroup

	for i := 0; i < writers; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			task := &input_itf.TaskEntity{
				ID:        uuid.New(),
				SessionID: session.ID,
				Status:    enums.TaskProcessing,
			}

			<-start

			for j := 0; j < each; j++ {
				if err := tasks.SaveTaskHistory(
					[]*input_itf.SessionEntity{session},
					[]*input_itf.TaskEntity{task},
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

func TestLoadTaskHistoryRoundTripsSessionsTasksAndReports(t *testing.T) {
	store, err := New(filepath.Join(t.TempDir(), "harness.db"))
	if err != nil {
		t.Fatalf("open storage: %v", err)
	}

	tasks := store.TaskStore()
	base := time.Now().UTC().Truncate(time.Second)

	first := &input_itf.SessionEntity{
		ID:             uuid.New(),
		WorkingDirPath: "/repo/nexo",
		ContextDirPath: "/repo/nexo/.nexo",
		StartedAt:      base,
		CompletedAt:    base.Add(10 * time.Minute),
		TotalTask:      2,
		TotalRetry:     1,
		CreatedAt:      base,
		UpdatedAt:      base.Add(10 * time.Minute),
	}

	second := &input_itf.SessionEntity{
		ID:             uuid.New(),
		WorkingDirPath: "/repo/other",
		StartedAt:      base.Add(time.Hour),
		TotalTask:      1,
		CreatedAt:      base.Add(time.Hour),
		UpdatedAt:      base.Add(time.Hour),
	}

	planner := &input_itf.TaskEntity{
		ID:                   uuid.New(),
		SessionID:            first.ID,
		Name:                 "plan the slice",
		TaskLevel:            enums.HeavyTask,
		PreferredModel:       enums.Opus,
		ThinkingLevel:        enums.HighThinking,
		SystemPrompts:        []string{"be terse", "no new comments"},
		AutoRetry:            true,
		ManualAcceptRequired: true,
		ExtraGuidance:        "read CLAUDE.md first",
		RetryCount:           2,
		Status:               enums.TaskAwaitingAccept,
		CreatedAt:            base,
		UpdatedAt:            base.Add(time.Minute),
	}

	implementer := &input_itf.TaskEntity{
		ID:               uuid.New(),
		SessionID:        first.ID,
		Name:             "write the code",
		TaskLevel:        enums.DailyTask,
		PreferredModel:   enums.Sonnet,
		ThinkingLevel:    enums.MedThinking,
		Status:           enums.TaskNotTaken,
		DependsOnTaskIDs: uuid.UUIDs{planner.ID},
		CreatedAt:        base.Add(time.Second),
		UpdatedAt:        base.Add(time.Second),
	}

	stray := &input_itf.TaskEntity{
		ID:        uuid.New(),
		SessionID: second.ID,
		Name:      "unrelated",
		Status:    enums.TaskCompleted,
		CreatedAt: base.Add(time.Hour),
		UpdatedAt: base.Add(time.Hour),
	}

	failed := &input_itf.TaskReportEntity{
		ID:            uuid.New(),
		TaskID:        planner.ID,
		AgentID:       uuid.New(),
		AttemptStatus: enums.TaskFailed,
		HandoverDocs: []*input_itf.HandoverDocEntity{{
			Task:     "plan the slice",
			TLDR:     "ran out of context",
			Blockers: map[string]string{"context": "window filled"},
		}},
		ContextUsage: &input_itf.ContextUsage{Total: 200000, Used: 199000, Billed: 4000, Input: 90000, Cached: 105000},
		StartedAt:    base,
		CompletedAt:  base.Add(30 * time.Second),
		CreatedAt:    base.Add(30 * time.Second),
		UpdatedAt:    base.Add(30 * time.Second),
	}

	accepted := &input_itf.TaskReportEntity{
		ID:            uuid.New(),
		TaskID:        planner.ID,
		AgentID:       uuid.New(),
		AttemptStatus: enums.TaskAwaitingAccept,
		HandoverDocs: []*input_itf.HandoverDocEntity{{
			Task:              "plan the slice",
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

	if err := tasks.SaveTaskHistory(
		[]*input_itf.SessionEntity{first, second},
		[]*input_itf.TaskEntity{planner, implementer, stray},
		[]*input_itf.TaskReportEntity{failed, accepted},
	); err != nil {
		t.Fatalf("save task history: %v", err)
	}

	snapshots, err := tasks.LoadTaskHistory()
	if err != nil {
		t.Fatalf("load task history: %v", err)
	}

	if len(snapshots) != 2 {
		t.Fatalf("loaded %d snapshots, want 2", len(snapshots))
	}

	if snapshots[0].Session.ID != first.ID || snapshots[1].Session.ID != second.ID {
		t.Fatalf("snapshots came back out of created_at order: %s then %s",
			snapshots[0].Session.ID, snapshots[1].Session.ID)
	}

	gotSession := snapshots[0].Session

	if gotSession.WorkingDirPath != first.WorkingDirPath || gotSession.ContextDirPath != first.ContextDirPath {
		t.Fatalf("session paths = %q / %q, want %q / %q",
			gotSession.WorkingDirPath, gotSession.ContextDirPath,
			first.WorkingDirPath, first.ContextDirPath)
	}

	if gotSession.TotalTask != first.TotalTask || gotSession.TotalRetry != first.TotalRetry {
		t.Fatalf("session totals = %d / %d, want %d / %d",
			gotSession.TotalTask, gotSession.TotalRetry, first.TotalTask, first.TotalRetry)
	}

	if !gotSession.StartedAt.Equal(first.StartedAt) || !gotSession.CompletedAt.Equal(first.CompletedAt) {
		t.Fatalf("session window = %s..%s, want %s..%s",
			gotSession.StartedAt, gotSession.CompletedAt, first.StartedAt, first.CompletedAt)
	}

	if !gotSession.CreatedAt.Equal(first.CreatedAt) || !gotSession.UpdatedAt.Equal(first.UpdatedAt) {
		t.Fatalf("session timestamps = %s / %s, want %s / %s",
			gotSession.CreatedAt, gotSession.UpdatedAt, first.CreatedAt, first.UpdatedAt)
	}

	if len(snapshots[0].Tasks) != 2 {
		t.Fatalf("first session loaded %d tasks, want 2", len(snapshots[0].Tasks))
	}

	if len(snapshots[1].Tasks) != 1 || snapshots[1].Tasks[0].ID != stray.ID {
		t.Fatalf("second session tasks = %+v, want only %s", snapshots[1].Tasks, stray.ID)
	}

	if len(snapshots[1].Reports) != 0 {
		t.Fatalf("second session gained %d reports", len(snapshots[1].Reports))
	}

	gotTask := snapshots[0].Tasks[0]

	if gotTask.ID != planner.ID {
		t.Fatalf("tasks came back out of created_at order: first is %s, want %s", gotTask.ID, planner.ID)
	}

	if gotTask.SessionID != planner.SessionID || gotTask.Name != planner.Name {
		t.Fatalf("task identity = %s / %q, want %s / %q",
			gotTask.SessionID, gotTask.Name, planner.SessionID, planner.Name)
	}

	if gotTask.TaskLevel != planner.TaskLevel || gotTask.PreferredModel != planner.PreferredModel ||
		gotTask.ThinkingLevel != planner.ThinkingLevel {
		t.Fatalf("task enums = %s / %s / %s, want %s / %s / %s",
			gotTask.TaskLevel, gotTask.PreferredModel, gotTask.ThinkingLevel,
			planner.TaskLevel, planner.PreferredModel, planner.ThinkingLevel)
	}

	if !reflect.DeepEqual(gotTask.SystemPrompts, planner.SystemPrompts) {
		t.Fatalf("task system prompts = %#v, want %#v", gotTask.SystemPrompts, planner.SystemPrompts)
	}

	if !gotTask.AutoRetry || !gotTask.ManualAcceptRequired {
		t.Fatalf("task flags = auto_retry %t / manual_accept %t, want both true",
			gotTask.AutoRetry, gotTask.ManualAcceptRequired)
	}

	if gotTask.ExtraGuidance != planner.ExtraGuidance || gotTask.RetryCount != planner.RetryCount {
		t.Fatalf("task guidance = %q / retries %d, want %q / %d",
			gotTask.ExtraGuidance, gotTask.RetryCount, planner.ExtraGuidance, planner.RetryCount)
	}

	if gotTask.Status != enums.TaskAwaitingAccept || gotTask.LastReportID != accepted.ID {
		t.Fatalf("task status = %s / last report %s, want %s / %s",
			gotTask.Status, gotTask.LastReportID, enums.TaskAwaitingAccept, accepted.ID)
	}

	if !gotTask.CreatedAt.Equal(planner.CreatedAt) || !gotTask.UpdatedAt.Equal(planner.UpdatedAt) {
		t.Fatalf("task timestamps = %s / %s, want %s / %s",
			gotTask.CreatedAt, gotTask.UpdatedAt, planner.CreatedAt, planner.UpdatedAt)
	}

	gotDependant := snapshots[0].Tasks[1]

	if !reflect.DeepEqual(gotDependant.DependsOnTaskIDs, uuid.UUIDs{planner.ID}) {
		t.Fatalf("depends on = %v, want %v", gotDependant.DependsOnTaskIDs, uuid.UUIDs{planner.ID})
	}

	if gotDependant.Status != enums.TaskNotTaken {
		t.Fatalf("dependant status = %s, want %s", gotDependant.Status, enums.TaskNotTaken)
	}

	if len(snapshots[0].Reports) != 2 {
		t.Fatalf("first session loaded %d reports, want 2", len(snapshots[0].Reports))
	}

	gotFailed, gotAccepted := snapshots[0].Reports[0], snapshots[0].Reports[1]

	if gotFailed.ID != failed.ID || gotAccepted.ID != accepted.ID {
		t.Fatalf("reports came back as %s then %s, want %s then %s",
			gotFailed.ID, gotAccepted.ID, failed.ID, accepted.ID)
	}

	if gotAccepted.TaskID != accepted.TaskID || gotAccepted.AgentID != accepted.AgentID {
		t.Fatalf("report owners = %s / %s, want %s / %s",
			gotAccepted.TaskID, gotAccepted.AgentID, accepted.TaskID, accepted.AgentID)
	}

	if gotFailed.AttemptStatus != enums.TaskFailed || gotAccepted.AttemptStatus != enums.TaskAwaitingAccept {
		t.Fatalf("attempt statuses = %s / %s, want %s / %s",
			gotFailed.AttemptStatus, gotAccepted.AttemptStatus, enums.TaskFailed, enums.TaskAwaitingAccept)
	}

	if !reflect.DeepEqual(gotAccepted.HandoverDocs, accepted.HandoverDocs) {
		t.Fatalf("handover docs = %#v, want %#v", gotAccepted.HandoverDocs[0], accepted.HandoverDocs[0])
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

func TestLoadTaskHistoryReturnsReportsOldestFirst(t *testing.T) {
	store, err := New(filepath.Join(t.TempDir(), "harness.db"))
	if err != nil {
		t.Fatalf("open storage: %v", err)
	}

	tasks := store.TaskStore()
	base := time.Now().UTC().Truncate(time.Second)

	session := &input_itf.SessionEntity{ID: uuid.New(), CreatedAt: base, UpdatedAt: base}
	task := &input_itf.TaskEntity{
		ID:        uuid.New(),
		SessionID: session.ID,
		Status:    enums.TaskCompleted,
		CreatedAt: base,
		UpdatedAt: base,
	}

	newest := &input_itf.TaskReportEntity{
		ID:        uuid.New(),
		TaskID:    task.ID,
		CreatedAt: base.Add(3 * time.Minute),
	}
	oldest := &input_itf.TaskReportEntity{
		ID:        uuid.New(),
		TaskID:    task.ID,
		CreatedAt: base.Add(time.Minute),
	}
	middle := &input_itf.TaskReportEntity{
		ID:        uuid.New(),
		TaskID:    task.ID,
		CreatedAt: base.Add(2 * time.Minute),
	}

	if err := tasks.SaveTaskHistory(
		[]*input_itf.SessionEntity{session},
		[]*input_itf.TaskEntity{task},
		[]*input_itf.TaskReportEntity{newest, oldest, middle},
	); err != nil {
		t.Fatalf("save task history: %v", err)
	}

	snapshots, err := tasks.LoadTaskHistory()
	if err != nil {
		t.Fatalf("load task history: %v", err)
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

func TestLoadTaskHistoryKeepsAMissingContextUsageNil(t *testing.T) {
	store, err := New(filepath.Join(t.TempDir(), "harness.db"))
	if err != nil {
		t.Fatalf("open storage: %v", err)
	}

	tasks := store.TaskStore()
	base := time.Now().UTC().Truncate(time.Second)

	session := &input_itf.SessionEntity{ID: uuid.New(), CreatedAt: base, UpdatedAt: base}
	task := &input_itf.TaskEntity{
		ID:        uuid.New(),
		SessionID: session.ID,
		Status:    enums.TaskFailed,
		CreatedAt: base,
		UpdatedAt: base,
	}

	report := &input_itf.TaskReportEntity{
		ID:            uuid.New(),
		TaskID:        task.ID,
		AttemptStatus: enums.TaskFailed,
		CreatedAt:     base,
	}

	if err := tasks.SaveTaskHistory(
		[]*input_itf.SessionEntity{session},
		[]*input_itf.TaskEntity{task},
		[]*input_itf.TaskReportEntity{report},
	); err != nil {
		t.Fatalf("save task history: %v", err)
	}

	snapshots, err := tasks.LoadTaskHistory()
	if err != nil {
		t.Fatalf("load task history: %v", err)
	}

	if len(snapshots) != 1 || len(snapshots[0].Reports) != 1 {
		t.Fatalf("loaded %d snapshots, want 1 carrying 1 report", len(snapshots))
	}

	if snapshots[0].Reports[0].ContextUsage != nil {
		t.Fatalf("context usage came back as %+v, want nil", snapshots[0].Reports[0].ContextUsage)
	}
}

func TestLoadTaskHistoryOnAnEmptyDatabase(t *testing.T) {
	store, err := New(filepath.Join(t.TempDir(), "harness.db"))
	if err != nil {
		t.Fatalf("open storage: %v", err)
	}

	snapshots, err := store.TaskStore().LoadTaskHistory()
	if err != nil {
		t.Fatalf("load task history: %v", err)
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
