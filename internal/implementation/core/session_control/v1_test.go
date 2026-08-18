package session_control

import (
	"errors"
	"strings"
	"testing"
	"time"

	"hexago/internal/helpers/enums"
	core_itf "hexago/internal/interface/core"
	input_itf "hexago/internal/interface/input"
	output_itf "hexago/internal/interface/output"

	"github.com/google/uuid"
)

type fakeSessions struct {
	core_itf.SessionManager

	sessionID uuid.UUID
	init      *core_itf.InitSession
	added     []*core_itf.AddTask
	taskIDs   []uuid.UUID
	statuses  map[uuid.UUID]*core_itf.SessionStatus
	answered  map[uuid.UUID]bool
}

func newFakeSessions() *fakeSessions {
	return &fakeSessions{
		sessionID: uuid.New(),
		statuses:  map[uuid.UUID]*core_itf.SessionStatus{},
		answered:  map[uuid.UUID]bool{},
	}
}

func (f *fakeSessions) NewSession(p *core_itf.InitSession) (uuid.UUID, error) {
	f.init = p
	return f.sessionID, nil
}

func (f *fakeSessions) AddTask(session uuid.UUID, task *core_itf.AddTask) (uuid.UUID, error) {
	f.added = append(f.added, task)

	id := uuid.New()
	f.taskIDs = append(f.taskIDs, id)

	return id, nil
}

func (f *fakeSessions) Status(id uuid.UUID) (*core_itf.SessionStatus, error) {
	status, found := f.statuses[id]
	if !found {
		return nil, errors.New("session not found")
	}

	return status, nil
}

func (f *fakeSessions) AnswerAcceptance(taskID uuid.UUID, accepted bool) error {
	f.answered[taskID] = accepted
	return nil
}

func (f *fakeSessions) addedNames() []string {
	names := make([]string, 0, len(f.added))
	for _, task := range f.added {
		names = append(names, task.Name)
	}

	return names
}

type fakeCoordinator struct {
	core_itf.Coordinator

	ran    []uuid.UUID
	paused []uuid.UUID
}

func (f *fakeCoordinator) Run(session uuid.UUID) error {
	f.ran = append(f.ran, session)
	return nil
}

func (f *fakeCoordinator) Pause(session uuid.UUID) error {
	f.paused = append(f.paused, session)
	return nil
}

type fakeTemplates struct {
	core_itf.AgentTemplateManager

	byID map[uuid.UUID]*core_itf.Template
}

func (f *fakeTemplates) Get(id uuid.UUID) (*core_itf.Template, error) {
	template, found := f.byID[id]
	if !found {
		return nil, errors.New("template not found")
	}

	return template, nil
}

func (f *fakeTemplates) List() ([]*core_itf.Template, error) {
	templates := make([]*core_itf.Template, 0, len(f.byID))
	for _, template := range f.byID {
		templates = append(templates, template)
	}

	return templates, nil
}

type fakeUserConfig struct {
	output_itf.UserConfig

	autopilot bool
}

func (f *fakeUserConfig) Autopilot() bool {
	return f.autopilot
}

func (f *fakeUserConfig) AgentDefault(level enums.TaskLevel) (*output_itf.AgentDefault, error) {
	return &output_itf.AgentDefault{Model: "sonnet", ThinkingLevel: "medium"}, nil
}

type fakeStorage struct {
	input_itf.TaskStorage

	snapshots []*input_itf.SessionSnapshot
	loadErr   error
	loads     int
}

func (f *fakeStorage) LoadTaskHistory() ([]*input_itf.SessionSnapshot, error) {
	f.loads++

	return f.snapshots, f.loadErr
}

type harness struct {
	control     core_itf.SessionControl
	sessions    *fakeSessions
	coordinator *fakeCoordinator
	templates   *fakeTemplates
	userConfig  *fakeUserConfig
	storage     *fakeStorage
}

func newHarness(t *testing.T, cfg *input_itf.ControlConfig) *harness {
	t.Helper()

	h := &harness{
		sessions:    newFakeSessions(),
		coordinator: &fakeCoordinator{},
		templates:   &fakeTemplates{byID: map[uuid.UUID]*core_itf.Template{}},
		userConfig:  &fakeUserConfig{},
		storage:     &fakeStorage{},
	}

	control, err := InitV1(cfg, h.sessions, h.coordinator, h.templates, h.userConfig, h.storage)
	if err != nil {
		t.Fatalf("init session control: %v", err)
	}

	h.control = control

	return h
}

func anyWorkspaceConfig() *input_itf.ControlConfig {
	return &input_itf.ControlConfig{
		Enabled:            true,
		EndpointFile:       "control.json",
		AllowAnyWorkspace:  true,
		MaxTasksPerSession: 20,
		MaxSessionsListed:  10,
	}
}

func task(clientID, name string) *core_itf.ControlTaskSpec {
	return &core_itf.ControlTaskSpec{ClientID: clientID, Name: name, Prompt: name}
}

func TestWorkspaceOutsideAllowedRootsIsRejected(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()

	cfg := anyWorkspaceConfig()
	cfg.AllowAnyWorkspace = false
	cfg.AllowedRoots = []string{root}

	h := newHarness(t, cfg)

	_, err := h.control.CreateSession(&core_itf.ControlSessionSpec{
		WorkingDirPath: outside,
		Tasks:          []*core_itf.ControlTaskSpec{task("a", "first")},
	})
	if err == nil {
		t.Fatal("a working dir outside every allowed root was accepted")
	}

	if !strings.Contains(err.Error(), "allowed_roots") {
		t.Fatalf("error does not name the config field to change: %v", err)
	}
}

func TestWorkspaceInsideAllowedRootIsAccepted(t *testing.T) {
	root := t.TempDir()

	cfg := anyWorkspaceConfig()
	cfg.AllowAnyWorkspace = false
	cfg.AllowedRoots = []string{root}

	h := newHarness(t, cfg)

	if _, err := h.control.CreateSession(&core_itf.ControlSessionSpec{
		WorkingDirPath: root + "/nested/deeper",
		Tasks:          []*core_itf.ControlTaskSpec{task("a", "first")},
	}); err != nil {
		t.Fatalf("a working dir inside an allowed root was rejected: %v", err)
	}
}

func TestEmptyAllowedRootsRejectsEveryWorkspace(t *testing.T) {
	cfg := anyWorkspaceConfig()
	cfg.AllowAnyWorkspace = false

	h := newHarness(t, cfg)

	_, err := h.control.CreateSession(&core_itf.ControlSessionSpec{
		WorkingDirPath: t.TempDir(),
		Tasks:          []*core_itf.ControlTaskSpec{task("a", "first")},
	})
	if err == nil {
		t.Fatal("an empty allowed roots list accepted a workspace")
	}

	if !strings.Contains(err.Error(), "allow_any_workspace") {
		t.Fatalf("error does not name the config field to change: %v", err)
	}
}

func TestAllowAnyWorkspaceSkipsTheGuard(t *testing.T) {
	h := newHarness(t, anyWorkspaceConfig())

	if _, err := h.control.CreateSession(&core_itf.ControlSessionSpec{
		WorkingDirPath: "/somewhere/else",
		Tasks:          []*core_itf.ControlTaskSpec{task("a", "first")},
	}); err != nil {
		t.Fatalf("allow any workspace still rejected a path: %v", err)
	}
}

func TestDepsDeclaredOutOfOrderResolve(t *testing.T) {
	h := newHarness(t, anyWorkspaceConfig())

	last := task("c", "last")
	last.DependsOn = []string{"b"}

	middle := task("b", "middle")
	middle.DependsOn = []string{"a"}

	ref, err := h.control.CreateSession(&core_itf.ControlSessionSpec{
		WorkingDirPath: "/work",
		Tasks:          []*core_itf.ControlTaskSpec{last, middle, task("a", "first")},
	})
	if err != nil {
		t.Fatalf("out of order dependencies did not resolve: %v", err)
	}

	names := h.sessions.addedNames()
	want := []string{"first", "middle", "last"}

	for i, name := range want {
		if names[i] != name {
			t.Fatalf("tasks were added as %v, want %v", names, want)
		}
	}

	if len(ref.TaskIDs) != 3 {
		t.Fatalf("session ref holds %d task ids, want 3", len(ref.TaskIDs))
	}

	if got := len(h.sessions.added[2].DependsOn); got != 1 {
		t.Fatalf("last task has %d dependencies, want 1", got)
	}

	if h.sessions.added[2].DependsOn[0] != h.sessions.taskIDs[1] {
		t.Fatal("last task does not depend on the middle task")
	}
}

func TestCyclicDepsAreRejected(t *testing.T) {
	h := newHarness(t, anyWorkspaceConfig())

	first := task("a", "first")
	first.DependsOn = []string{"b"}

	second := task("b", "second")
	second.DependsOn = []string{"a"}

	_, err := h.control.CreateSession(&core_itf.ControlSessionSpec{
		WorkingDirPath: "/work",
		Tasks:          []*core_itf.ControlTaskSpec{first, second},
	})
	if err == nil {
		t.Fatal("a cyclic graph was accepted")
	}

	if !strings.Contains(err.Error(), "cyclic") {
		t.Fatalf("error does not explain the cycle: %v", err)
	}
}

func TestUnknownDepIsRejected(t *testing.T) {
	h := newHarness(t, anyWorkspaceConfig())

	only := task("a", "first")
	only.DependsOn = []string{"ghost"}

	_, err := h.control.CreateSession(&core_itf.ControlSessionSpec{
		WorkingDirPath: "/work",
		Tasks:          []*core_itf.ControlTaskSpec{only},
	})
	if err == nil {
		t.Fatal("a dependency on an unknown client id was accepted")
	}
}

func TestDuplicateClientIDIsRejected(t *testing.T) {
	h := newHarness(t, anyWorkspaceConfig())

	_, err := h.control.CreateSession(&core_itf.ControlSessionSpec{
		WorkingDirPath: "/work",
		Tasks:          []*core_itf.ControlTaskSpec{task("a", "first"), task("a", "second")},
	})
	if err == nil {
		t.Fatal("two tasks sharing a client id were accepted")
	}
}

func TestTooManyTasksAreRejected(t *testing.T) {
	cfg := anyWorkspaceConfig()
	cfg.MaxTasksPerSession = 1

	h := newHarness(t, cfg)

	_, err := h.control.CreateSession(&core_itf.ControlSessionSpec{
		WorkingDirPath: "/work",
		Tasks:          []*core_itf.ControlTaskSpec{task("a", "first"), task("b", "second")},
	})
	if err == nil {
		t.Fatal("a session over the task limit was accepted")
	}
}

func reviewTemplate() *core_itf.Template {
	return &core_itf.Template{
		ID:                   uuid.New(),
		Name:                 "review",
		Role:                 "reviewer",
		TaskLevel:            enums.HeavyTask,
		ManualAcceptRequired: true,
		OutputStructure:      "a list of findings",
		SystemPrompts: map[string]string{
			"2_then":  "then read the diff",
			"1_first": "you are a reviewer",
			"3_last":  "report once",
		},
		Params: map[string]*core_itf.TemplateParams{
			"ticket": {Type: string(enums.TextParam), Required: true},
			"depth":  {Type: string(enums.SelectParam), Default: "shallow", Options: []string{"shallow", "deep"}},
		},
	}
}

func withTemplate(t *testing.T, h *harness) *core_itf.Template {
	t.Helper()

	template := reviewTemplate()
	h.templates.byID[template.ID] = template

	return template
}

func TestTemplateTaskTakesPromptsInKeyOrder(t *testing.T) {
	h := newHarness(t, anyWorkspaceConfig())
	template := withTemplate(t, h)

	backed := task("a", "review")
	backed.TemplateID = template.ID
	backed.Params = map[string]string{"ticket": "NEX-1"}

	if _, err := h.control.CreateSession(&core_itf.ControlSessionSpec{
		WorkingDirPath: "/work",
		Tasks:          []*core_itf.ControlTaskSpec{backed},
	}); err != nil {
		t.Fatalf("create session: %v", err)
	}

	added := h.sessions.added[0]
	want := []string{"you are a reviewer", "then read the diff", "report once"}

	for i, prompt := range want {
		if added.AgentSpecs.SystemPrompts[i] != prompt {
			t.Fatalf("system prompts are %v, want %v", added.AgentSpecs.SystemPrompts, want)
		}
	}

	if added.TaskLevel != enums.HeavyTask {
		t.Fatalf("task level is %s, want %s", added.TaskLevel, enums.HeavyTask)
	}

	if added.OutputStructure != template.OutputStructure {
		t.Fatalf("output structure is %q, want %q", added.OutputStructure, template.OutputStructure)
	}

	if !added.ManualAcceptRequired {
		t.Fatal("the template asks for a manual accept and the task did not get one")
	}
}

func TestTemplateDefaultFillsAMissingParam(t *testing.T) {
	h := newHarness(t, anyWorkspaceConfig())
	template := withTemplate(t, h)

	backed := task("a", "review")
	backed.TemplateID = template.ID
	backed.Prompt = "review {{ticket}}"
	backed.Params = map[string]string{"ticket": "NEX-1"}

	if _, err := h.control.CreateSession(&core_itf.ControlSessionSpec{
		WorkingDirPath: "/work",
		Tasks:          []*core_itf.ControlTaskSpec{backed},
	}); err != nil {
		t.Fatalf("create session: %v", err)
	}

	want := "review NEX-1\n\nInputs:\n- depth: shallow"
	if got := h.sessions.added[0].ExtraGuidance; got != want {
		t.Fatalf("prompt is %q, want %q", got, want)
	}
}

func TestTemplateMissingRequiredParamIsRejected(t *testing.T) {
	h := newHarness(t, anyWorkspaceConfig())
	template := withTemplate(t, h)

	backed := task("a", "review")
	backed.TemplateID = template.ID

	_, err := h.control.CreateSession(&core_itf.ControlSessionSpec{
		WorkingDirPath: "/work",
		Tasks:          []*core_itf.ControlTaskSpec{backed},
	})
	if err == nil {
		t.Fatal("a task missing a required param was accepted")
	}

	if !strings.Contains(err.Error(), "ticket") {
		t.Fatalf("error does not name the missing param: %v", err)
	}
}

func TestTemplateUnknownParamIsRejected(t *testing.T) {
	h := newHarness(t, anyWorkspaceConfig())
	template := withTemplate(t, h)

	backed := task("a", "review")
	backed.TemplateID = template.ID
	backed.Params = map[string]string{"ticket": "NEX-1", "colour": "blue"}

	_, err := h.control.CreateSession(&core_itf.ControlSessionSpec{
		WorkingDirPath: "/work",
		Tasks:          []*core_itf.ControlTaskSpec{backed},
	})
	if err == nil {
		t.Fatal("a task setting an undeclared param was accepted")
	}

	if !strings.Contains(err.Error(), "colour") {
		t.Fatalf("error does not name the unknown param: %v", err)
	}
}

func TestTemplateSelectParamOutsideOptionsIsRejected(t *testing.T) {
	h := newHarness(t, anyWorkspaceConfig())
	template := withTemplate(t, h)

	backed := task("a", "review")
	backed.TemplateID = template.ID
	backed.Params = map[string]string{"ticket": "NEX-1", "depth": "exhaustive"}

	if _, err := h.control.CreateSession(&core_itf.ControlSessionSpec{
		WorkingDirPath: "/work",
		Tasks:          []*core_itf.ControlTaskSpec{backed},
	}); err == nil {
		t.Fatal("a select value outside the options was accepted")
	}
}

func TestFreeformTaskDefaultsToDailyLevel(t *testing.T) {
	h := newHarness(t, anyWorkspaceConfig())

	if _, err := h.control.CreateSession(&core_itf.ControlSessionSpec{
		WorkingDirPath: "/work",
		Tasks:          []*core_itf.ControlTaskSpec{task("a", "first")},
	}); err != nil {
		t.Fatalf("create session: %v", err)
	}

	if got := h.sessions.added[0].TaskLevel; got != enums.DailyTask {
		t.Fatalf("task level is %s, want %s", got, enums.DailyTask)
	}
}

func TestFreeformTaskRejectsAnUnknownLevel(t *testing.T) {
	h := newHarness(t, anyWorkspaceConfig())

	loose := task("a", "first")
	loose.TaskLevel = "titanic_task"

	if _, err := h.control.CreateSession(&core_itf.ControlSessionSpec{
		WorkingDirPath: "/work",
		Tasks:          []*core_itf.ControlTaskSpec{loose},
	}); err == nil {
		t.Fatal("an unknown task level was accepted")
	}
}

func TestPromptRendersReferencesAndAppendsTheRest(t *testing.T) {
	h := newHarness(t, anyWorkspaceConfig())

	loose := task("a", "first")
	loose.Prompt = "fix {{ticket}} on {{ branch }}"
	loose.Params = map[string]string{
		"ticket": "NEX-9",
		"branch": "main",
		"owner":  "rye",
		"note":   "",
	}

	if _, err := h.control.CreateSession(&core_itf.ControlSessionSpec{
		WorkingDirPath: "/work",
		Tasks:          []*core_itf.ControlTaskSpec{loose},
	}); err != nil {
		t.Fatalf("create session: %v", err)
	}

	want := "fix NEX-9 on main\n\nInputs:\n- owner: rye"
	if got := h.sessions.added[0].ExtraGuidance; got != want {
		t.Fatalf("prompt is %q, want %q", got, want)
	}
}

func TestPromptWithoutParamsIsUnchanged(t *testing.T) {
	h := newHarness(t, anyWorkspaceConfig())

	loose := task("a", "first")
	loose.Prompt = "just do the thing"

	if _, err := h.control.CreateSession(&core_itf.ControlSessionSpec{
		WorkingDirPath: "/work",
		Tasks:          []*core_itf.ControlTaskSpec{loose},
	}); err != nil {
		t.Fatalf("create session: %v", err)
	}

	if got := h.sessions.added[0].ExtraGuidance; got != loose.Prompt {
		t.Fatalf("prompt is %q, want %q", got, loose.Prompt)
	}
}

func TestAutostartFallsBackToTheConfig(t *testing.T) {
	cfg := anyWorkspaceConfig()
	cfg.AutostartDefault = true

	h := newHarness(t, cfg)

	ref, err := h.control.CreateSession(&core_itf.ControlSessionSpec{
		WorkingDirPath: "/work",
		Tasks:          []*core_itf.ControlTaskSpec{task("a", "first")},
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	if !ref.Started {
		t.Fatal("the config default asks for an autostart and the session did not start")
	}

	if len(h.coordinator.ran) != 1 {
		t.Fatalf("the coordinator ran %d times, want 1", len(h.coordinator.ran))
	}
}

func TestAutostartFalseOverridesTheConfig(t *testing.T) {
	cfg := anyWorkspaceConfig()
	cfg.AutostartDefault = true

	h := newHarness(t, cfg)
	off := false

	ref, err := h.control.CreateSession(&core_itf.ControlSessionSpec{
		WorkingDirPath: "/work",
		Autostart:      &off,
		Tasks:          []*core_itf.ControlTaskSpec{task("a", "first")},
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	if ref.Started || len(h.coordinator.ran) != 0 {
		t.Fatal("autostart false still started the session")
	}
}

func TestAutostartTrueRunsTheCoordinatorOnce(t *testing.T) {
	h := newHarness(t, anyWorkspaceConfig())
	on := true

	ref, err := h.control.CreateSession(&core_itf.ControlSessionSpec{
		WorkingDirPath: "/work",
		Autostart:      &on,
		Tasks:          []*core_itf.ControlTaskSpec{task("a", "first"), task("b", "second")},
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	if !ref.Started {
		t.Fatal("autostart true did not mark the session started")
	}

	if len(h.coordinator.ran) != 1 || h.coordinator.ran[0] != h.sessions.sessionID {
		t.Fatalf("the coordinator ran %v, want one run of %v", h.coordinator.ran, h.sessions.sessionID)
	}
}

func TestAutopilotDropsTheManualAccept(t *testing.T) {
	h := newHarness(t, anyWorkspaceConfig())
	h.userConfig.autopilot = true
	template := withTemplate(t, h)

	backed := task("a", "review")
	backed.TemplateID = template.ID
	backed.Params = map[string]string{"ticket": "NEX-1"}

	gated := task("b", "gated")
	gated.ManualAcceptRequired = true

	if _, err := h.control.CreateSession(&core_itf.ControlSessionSpec{
		WorkingDirPath: "/work",
		Tasks:          []*core_itf.ControlTaskSpec{backed, gated},
	}); err != nil {
		t.Fatalf("create session: %v", err)
	}

	for _, added := range h.sessions.added {
		if added.ManualAcceptRequired {
			t.Fatalf("task %s kept its manual accept under autopilot", added.Name)
		}
	}
}

func TestSessionStateComesStraightFromTheManager(t *testing.T) {
	h := newHarness(t, anyWorkspaceConfig())

	sessionID := uuid.MustParse("00000000-0000-7000-8000-000000000001")
	want := &core_itf.SessionStatus{
		ID:             sessionID,
		Status:         enums.SessionProcessing,
		WorkingDirPath: "/work",
		Tasks:          map[uuid.UUID]*core_itf.TaskReport{},
	}
	h.sessions.statuses[sessionID] = want

	state, err := h.control.SessionState(sessionID)
	if err != nil {
		t.Fatalf("session state: %v", err)
	}

	if state != want {
		t.Fatalf("state is %p, want manager value %p", state, want)
	}

	if h.storage.loads != 0 {
		t.Fatalf("session_status read the task history %d times, want none", h.storage.loads)
	}
}

func TestListSessionsClampsTheLimitAndOrdersNewestFirst(t *testing.T) {
	cfg := anyWorkspaceConfig()
	cfg.MaxSessionsListed = 2

	h := newHarness(t, cfg)

	base := time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC)

	for i, status := range []enums.SessionStatus{enums.SessionInit, enums.SessionPaused, enums.SessionCompleted} {
		id := uuid.New()

		h.storage.snapshots = append(h.storage.snapshots, &input_itf.SessionSnapshot{
			Session: &input_itf.SessionEntity{
				ID:        id,
				TotalTask: i + 1,
				CreatedAt: base.Add(time.Duration(i) * time.Hour),
			},
		})

		h.sessions.statuses[id] = &core_itf.SessionStatus{
			ID:             id,
			Status:         status,
			WorkingDirPath: "/work",
			Tasks:          make(map[uuid.UUID]*core_itf.TaskReport, i+1),
		}
		for task := 0; task <= i; task++ {
			h.sessions.statuses[id].Tasks[uuid.New()] = &core_itf.TaskReport{}
		}
	}

	for _, limit := range []int{0, -3, 9} {
		listed, err := h.control.ListSessions(limit)
		if err != nil {
			t.Fatalf("list sessions with limit %d: %v", limit, err)
		}

		if len(listed) != cfg.MaxSessionsListed {
			t.Fatalf("limit %d listed %d sessions, want %d", limit, len(listed), cfg.MaxSessionsListed)
		}

		if listed[0].Status != enums.SessionCompleted || listed[1].Status != enums.SessionPaused {
			t.Fatalf("sessions are not newest first: %s then %s", listed[0].Status, listed[1].Status)
		}

		if len(listed[0].Tasks) != 3 || len(listed[1].Tasks) != 2 {
			t.Fatalf("task totals are %d and %d, want 3 and 2", len(listed[0].Tasks), len(listed[1].Tasks))
		}
	}
}

func TestListSessionsSkipsASessionTheManagerDoesNotKnow(t *testing.T) {
	h := newHarness(t, anyWorkspaceConfig())

	known := uuid.New()

	h.storage.snapshots = []*input_itf.SessionSnapshot{
		{Session: &input_itf.SessionEntity{ID: known, CreatedAt: time.Now()}},
		{Session: &input_itf.SessionEntity{ID: uuid.New(), CreatedAt: time.Now().Add(-time.Hour)}},
	}

	h.sessions.statuses[known] = &core_itf.SessionStatus{ID: known, Status: enums.SessionInit}

	listed, err := h.control.ListSessions(0)
	if err != nil {
		t.Fatalf("list sessions: %v", err)
	}

	if len(listed) != 1 || listed[0].ID != known {
		t.Fatalf("listing is %+v, want only the session the manager knows", listed)
	}
}

func TestListTemplatesComesStraightFromTheManager(t *testing.T) {
	h := newHarness(t, anyWorkspaceConfig())
	template := withTemplate(t, h)

	listed, err := h.control.ListTemplates()
	if err != nil {
		t.Fatalf("list templates: %v", err)
	}

	if len(listed) != 1 {
		t.Fatalf("listed %d templates, want 1", len(listed))
	}

	if listed[0] != template {
		t.Fatalf("template is %p, want manager value %p", listed[0], template)
	}
}

func TestLifecycleCallsForwardIDs(t *testing.T) {
	h := newHarness(t, anyWorkspaceConfig())
	sessionID := uuid.New()

	if err := h.control.StartSession(sessionID); err != nil {
		t.Fatalf("start session: %v", err)
	}

	if err := h.control.PauseSession(sessionID); err != nil {
		t.Fatalf("pause session: %v", err)
	}

	if len(h.coordinator.ran) != 1 || h.coordinator.ran[0] != sessionID {
		t.Fatalf("the coordinator ran %v, want one run of %v", h.coordinator.ran, sessionID)
	}

	if len(h.coordinator.paused) != 1 || h.coordinator.paused[0] != sessionID {
		t.Fatalf("the coordinator paused %v, want %v", h.coordinator.paused, sessionID)
	}
}

func TestAnswerAcceptanceReachesTheSessionManager(t *testing.T) {
	h := newHarness(t, anyWorkspaceConfig())
	taskID := uuid.New()

	if err := h.control.AnswerAcceptance(taskID, true); err != nil {
		t.Fatalf("answer acceptance: %v", err)
	}

	if accepted, found := h.sessions.answered[taskID]; !found || !accepted {
		t.Fatalf("the session manager saw %v for %v", h.sessions.answered, taskID)
	}
}
