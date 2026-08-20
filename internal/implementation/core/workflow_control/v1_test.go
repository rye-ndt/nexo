package workflow_control

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

type fakeWorkflows struct {
	core_itf.WorkflowManager

	workflowID uuid.UUID
	init       *core_itf.InitWorkflow
	added      []*core_itf.AddStep
	stepIDs    []uuid.UUID
	statuses   map[uuid.UUID]*core_itf.WorkflowStatus
	answered   map[uuid.UUID]bool
}

func newFakeWorkflows() *fakeWorkflows {
	return &fakeWorkflows{
		workflowID: uuid.New(),
		statuses:   map[uuid.UUID]*core_itf.WorkflowStatus{},
		answered:   map[uuid.UUID]bool{},
	}
}

func (f *fakeWorkflows) NewWorkflow(p *core_itf.InitWorkflow) (uuid.UUID, error) {
	f.init = p
	return f.workflowID, nil
}

func (f *fakeWorkflows) AddStep(workflow uuid.UUID, step *core_itf.AddStep) (uuid.UUID, error) {
	f.added = append(f.added, step)

	id := uuid.New()
	f.stepIDs = append(f.stepIDs, id)

	return id, nil
}

func (f *fakeWorkflows) Status(id uuid.UUID) (*core_itf.WorkflowStatus, error) {
	status, found := f.statuses[id]
	if !found {
		return nil, errors.New("workflow not found")
	}

	return status, nil
}

func (f *fakeWorkflows) AnswerReview(stepID uuid.UUID, accepted bool) error {
	f.answered[stepID] = accepted
	return nil
}

func (f *fakeWorkflows) addedNames() []string {
	names := make([]string, 0, len(f.added))
	for _, step := range f.added {
		names = append(names, step.Name)
	}

	return names
}

type fakeCoordinator struct {
	core_itf.Coordinator

	ran    []uuid.UUID
	paused []uuid.UUID
}

func (f *fakeCoordinator) Run(workflow uuid.UUID) error {
	f.ran = append(f.ran, workflow)
	return nil
}

func (f *fakeCoordinator) Pause(workflow uuid.UUID) error {
	f.paused = append(f.paused, workflow)
	return nil
}

type fakeRoles struct {
	core_itf.RoleManager

	byID map[uuid.UUID]*core_itf.Role
}

func (f *fakeRoles) Get(id uuid.UUID) (*core_itf.Role, error) {
	role, found := f.byID[id]
	if !found {
		return nil, errors.New("role not found")
	}

	return role, nil
}

func (f *fakeRoles) List() ([]*core_itf.Role, error) {
	roles := make([]*core_itf.Role, 0, len(f.byID))
	for _, role := range f.byID {
		roles = append(roles, role)
	}

	return roles, nil
}

type fakeUserConfig struct {
	output_itf.UserConfig

	autopilot bool
}

func (f *fakeUserConfig) Autopilot() bool {
	return f.autopilot
}

func (f *fakeUserConfig) AgentDefault(level enums.Effort) (*output_itf.AgentDefault, error) {
	return &output_itf.AgentDefault{Model: "sonnet", ThinkingLevel: "medium"}, nil
}

type fakeStorage struct {
	input_itf.StepStorage

	snapshots []*input_itf.WorkflowSnapshot
	loadErr   error
	loads     int
}

func (f *fakeStorage) LoadStepHistory() ([]*input_itf.WorkflowSnapshot, error) {
	f.loads++

	return f.snapshots, f.loadErr
}

type harness struct {
	control     core_itf.WorkflowControl
	workflows   *fakeWorkflows
	coordinator *fakeCoordinator
	roles       *fakeRoles
	userConfig  *fakeUserConfig
	storage     *fakeStorage
}

func newHarness(t *testing.T, cfg *input_itf.ControlConfig) *harness {
	t.Helper()

	h := &harness{
		workflows:   newFakeWorkflows(),
		coordinator: &fakeCoordinator{},
		roles:       &fakeRoles{byID: map[uuid.UUID]*core_itf.Role{}},
		userConfig:  &fakeUserConfig{},
		storage:     &fakeStorage{},
	}

	control, err := InitV1(cfg, h.workflows, h.coordinator, h.roles, h.userConfig, h.storage)
	if err != nil {
		t.Fatalf("init workflow control: %v", err)
	}

	h.control = control

	return h
}

func anyWorkspaceConfig() *input_itf.ControlConfig {
	return &input_itf.ControlConfig{
		Enabled:             true,
		EndpointFile:        "control.json",
		AllowAnyWorkspace:   true,
		MaxStepsPerWorkflow: 20,
		MaxWorkflowsListed:  10,
	}
}

func step(clientID, name string) *core_itf.ControlStepSpec {
	return &core_itf.ControlStepSpec{ClientID: clientID, Name: name, Prompt: name}
}

func TestWorkspaceOutsideAllowedRootsIsRejected(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()

	cfg := anyWorkspaceConfig()
	cfg.AllowAnyWorkspace = false
	cfg.AllowedRoots = []string{root}

	h := newHarness(t, cfg)

	_, err := h.control.CreateWorkflow(&core_itf.ControlWorkflowSpec{
		ProjectDirPath: outside,
		Steps:          []*core_itf.ControlStepSpec{step("a", "first")},
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

	if _, err := h.control.CreateWorkflow(&core_itf.ControlWorkflowSpec{
		ProjectDirPath: root + "/nested/deeper",
		Steps:          []*core_itf.ControlStepSpec{step("a", "first")},
	}); err != nil {
		t.Fatalf("a working dir inside an allowed root was rejected: %v", err)
	}
}

func TestEmptyAllowedRootsRejectsEveryWorkspace(t *testing.T) {
	cfg := anyWorkspaceConfig()
	cfg.AllowAnyWorkspace = false

	h := newHarness(t, cfg)

	_, err := h.control.CreateWorkflow(&core_itf.ControlWorkflowSpec{
		ProjectDirPath: t.TempDir(),
		Steps:          []*core_itf.ControlStepSpec{step("a", "first")},
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

	if _, err := h.control.CreateWorkflow(&core_itf.ControlWorkflowSpec{
		ProjectDirPath: "/somewhere/else",
		Steps:          []*core_itf.ControlStepSpec{step("a", "first")},
	}); err != nil {
		t.Fatalf("allow any workspace still rejected a path: %v", err)
	}
}

func TestDepsDeclaredOutOfOrderResolve(t *testing.T) {
	h := newHarness(t, anyWorkspaceConfig())

	last := step("c", "last")
	last.DependsOn = []string{"b"}

	middle := step("b", "middle")
	middle.DependsOn = []string{"a"}

	ref, err := h.control.CreateWorkflow(&core_itf.ControlWorkflowSpec{
		ProjectDirPath: "/work",
		Steps:          []*core_itf.ControlStepSpec{last, middle, step("a", "first")},
	})
	if err != nil {
		t.Fatalf("out of order dependencies did not resolve: %v", err)
	}

	names := h.workflows.addedNames()
	want := []string{"first", "middle", "last"}

	for i, name := range want {
		if names[i] != name {
			t.Fatalf("steps were added as %v, want %v", names, want)
		}
	}

	if len(ref.StepIDs) != 3 {
		t.Fatalf("workflow ref holds %d step ids, want 3", len(ref.StepIDs))
	}

	if got := len(h.workflows.added[2].DependsOn); got != 1 {
		t.Fatalf("last step has %d dependencies, want 1", got)
	}

	if h.workflows.added[2].DependsOn[0] != h.workflows.stepIDs[1] {
		t.Fatal("last step does not depend on the middle step")
	}
}

func TestCyclicDepsAreRejected(t *testing.T) {
	h := newHarness(t, anyWorkspaceConfig())

	first := step("a", "first")
	first.DependsOn = []string{"b"}

	second := step("b", "second")
	second.DependsOn = []string{"a"}

	_, err := h.control.CreateWorkflow(&core_itf.ControlWorkflowSpec{
		ProjectDirPath: "/work",
		Steps:          []*core_itf.ControlStepSpec{first, second},
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

	only := step("a", "first")
	only.DependsOn = []string{"ghost"}

	_, err := h.control.CreateWorkflow(&core_itf.ControlWorkflowSpec{
		ProjectDirPath: "/work",
		Steps:          []*core_itf.ControlStepSpec{only},
	})
	if err == nil {
		t.Fatal("a dependency on an unknown client id was accepted")
	}
}

func TestDuplicateClientIDIsRejected(t *testing.T) {
	h := newHarness(t, anyWorkspaceConfig())

	_, err := h.control.CreateWorkflow(&core_itf.ControlWorkflowSpec{
		ProjectDirPath: "/work",
		Steps:          []*core_itf.ControlStepSpec{step("a", "first"), step("a", "second")},
	})
	if err == nil {
		t.Fatal("two steps sharing a client id were accepted")
	}
}

func TestTooManyStepsAreRejected(t *testing.T) {
	cfg := anyWorkspaceConfig()
	cfg.MaxStepsPerWorkflow = 1

	h := newHarness(t, cfg)

	_, err := h.control.CreateWorkflow(&core_itf.ControlWorkflowSpec{
		ProjectDirPath: "/work",
		Steps:          []*core_itf.ControlStepSpec{step("a", "first"), step("b", "second")},
	})
	if err == nil {
		t.Fatal("a workflow over the step limit was accepted")
	}
}

func reviewRole() *core_itf.Role {
	return &core_itf.Role{
		ID:              uuid.New(),
		Name:            "review",
		Description:     "reviewer",
		Effort:          enums.EffortDeep,
		PauseForReview:  true,
		OutputStructure: "a list of findings",
		Instructions: map[string]string{
			"2_then":  "then read the diff",
			"1_first": "you are a reviewer",
			"3_last":  "report once",
		},
		Inputs: map[string]*core_itf.RoleInputs{
			"ticket": {Type: string(enums.TextInput), Required: true},
			"depth":  {Type: string(enums.SelectInput), Default: "shallow", Options: []string{"shallow", "deep"}},
		},
	}
}

func withRole(t *testing.T, h *harness) *core_itf.Role {
	t.Helper()

	role := reviewRole()
	h.roles.byID[role.ID] = role

	return role
}

func TestRoleStepTakesPromptsInKeyOrder(t *testing.T) {
	h := newHarness(t, anyWorkspaceConfig())
	role := withRole(t, h)

	backed := step("a", "review")
	backed.RoleID = role.ID
	backed.Inputs = map[string]string{"ticket": "NEX-1"}

	if _, err := h.control.CreateWorkflow(&core_itf.ControlWorkflowSpec{
		ProjectDirPath: "/work",
		Steps:          []*core_itf.ControlStepSpec{backed},
	}); err != nil {
		t.Fatalf("create workflow: %v", err)
	}

	added := h.workflows.added[0]
	want := []string{"you are a reviewer", "then read the diff", "report once"}

	for i, prompt := range want {
		if added.AgentSpecs.Instructions[i] != prompt {
			t.Fatalf("system prompts are %v, want %v", added.AgentSpecs.Instructions, want)
		}
	}

	if added.Effort != enums.EffortDeep {
		t.Fatalf("step level is %s, want %s", added.Effort, enums.EffortDeep)
	}

	if added.OutputStructure != role.OutputStructure {
		t.Fatalf("output structure is %q, want %q", added.OutputStructure, role.OutputStructure)
	}

	if !added.PauseForReview {
		t.Fatal("the role asks for a manual accept and the step did not get one")
	}
}

func TestRoleDefaultFillsAMissingInput(t *testing.T) {
	h := newHarness(t, anyWorkspaceConfig())
	role := withRole(t, h)

	backed := step("a", "review")
	backed.RoleID = role.ID
	backed.Prompt = "review {{ticket}}"
	backed.Inputs = map[string]string{"ticket": "NEX-1"}

	if _, err := h.control.CreateWorkflow(&core_itf.ControlWorkflowSpec{
		ProjectDirPath: "/work",
		Steps:          []*core_itf.ControlStepSpec{backed},
	}); err != nil {
		t.Fatalf("create workflow: %v", err)
	}

	want := "review NEX-1\n\nInputs:\n- depth: shallow"
	if got := h.workflows.added[0].ExtraGuidance; got != want {
		t.Fatalf("prompt is %q, want %q", got, want)
	}
}

func TestRoleMissingRequiredInputIsRejected(t *testing.T) {
	h := newHarness(t, anyWorkspaceConfig())
	role := withRole(t, h)

	backed := step("a", "review")
	backed.RoleID = role.ID

	_, err := h.control.CreateWorkflow(&core_itf.ControlWorkflowSpec{
		ProjectDirPath: "/work",
		Steps:          []*core_itf.ControlStepSpec{backed},
	})
	if err == nil {
		t.Fatal("a step missing a required input was accepted")
	}

	if !strings.Contains(err.Error(), "ticket") {
		t.Fatalf("error does not name the missing input: %v", err)
	}
}

func TestRoleUnknownInputIsRejected(t *testing.T) {
	h := newHarness(t, anyWorkspaceConfig())
	role := withRole(t, h)

	backed := step("a", "review")
	backed.RoleID = role.ID
	backed.Inputs = map[string]string{"ticket": "NEX-1", "colour": "blue"}

	_, err := h.control.CreateWorkflow(&core_itf.ControlWorkflowSpec{
		ProjectDirPath: "/work",
		Steps:          []*core_itf.ControlStepSpec{backed},
	})
	if err == nil {
		t.Fatal("a step setting an undeclared input was accepted")
	}

	if !strings.Contains(err.Error(), "colour") {
		t.Fatalf("error does not name the unknown input: %v", err)
	}
}

func TestRoleSelectInputOutsideOptionsIsRejected(t *testing.T) {
	h := newHarness(t, anyWorkspaceConfig())
	role := withRole(t, h)

	backed := step("a", "review")
	backed.RoleID = role.ID
	backed.Inputs = map[string]string{"ticket": "NEX-1", "depth": "exhaustive"}

	if _, err := h.control.CreateWorkflow(&core_itf.ControlWorkflowSpec{
		ProjectDirPath: "/work",
		Steps:          []*core_itf.ControlStepSpec{backed},
	}); err == nil {
		t.Fatal("a select value outside the options was accepted")
	}
}

func TestFreeformStepDefaultsToDailyLevel(t *testing.T) {
	h := newHarness(t, anyWorkspaceConfig())

	if _, err := h.control.CreateWorkflow(&core_itf.ControlWorkflowSpec{
		ProjectDirPath: "/work",
		Steps:          []*core_itf.ControlStepSpec{step("a", "first")},
	}); err != nil {
		t.Fatalf("create workflow: %v", err)
	}

	if got := h.workflows.added[0].Effort; got != enums.EffortStandard {
		t.Fatalf("step level is %s, want %s", got, enums.EffortStandard)
	}
}

func TestFreeformStepRejectsAnUnknownLevel(t *testing.T) {
	h := newHarness(t, anyWorkspaceConfig())

	loose := step("a", "first")
	loose.Effort = "titanic_step"

	if _, err := h.control.CreateWorkflow(&core_itf.ControlWorkflowSpec{
		ProjectDirPath: "/work",
		Steps:          []*core_itf.ControlStepSpec{loose},
	}); err == nil {
		t.Fatal("an unknown step level was accepted")
	}
}

func TestPromptRendersReferencesAndAppendsTheRest(t *testing.T) {
	h := newHarness(t, anyWorkspaceConfig())

	loose := step("a", "first")
	loose.Prompt = "fix {{ticket}} on {{ branch }}"
	loose.Inputs = map[string]string{
		"ticket": "NEX-9",
		"branch": "main",
		"owner":  "rye",
		"note":   "",
	}

	if _, err := h.control.CreateWorkflow(&core_itf.ControlWorkflowSpec{
		ProjectDirPath: "/work",
		Steps:          []*core_itf.ControlStepSpec{loose},
	}); err != nil {
		t.Fatalf("create workflow: %v", err)
	}

	want := "fix NEX-9 on main\n\nInputs:\n- owner: rye"
	if got := h.workflows.added[0].ExtraGuidance; got != want {
		t.Fatalf("prompt is %q, want %q", got, want)
	}
}

func TestPromptWithoutInputsIsUnchanged(t *testing.T) {
	h := newHarness(t, anyWorkspaceConfig())

	loose := step("a", "first")
	loose.Prompt = "just do the thing"

	if _, err := h.control.CreateWorkflow(&core_itf.ControlWorkflowSpec{
		ProjectDirPath: "/work",
		Steps:          []*core_itf.ControlStepSpec{loose},
	}); err != nil {
		t.Fatalf("create workflow: %v", err)
	}

	if got := h.workflows.added[0].ExtraGuidance; got != loose.Prompt {
		t.Fatalf("prompt is %q, want %q", got, loose.Prompt)
	}
}

func TestAutostartFallsBackToTheConfig(t *testing.T) {
	cfg := anyWorkspaceConfig()
	cfg.AutostartDefault = true

	h := newHarness(t, cfg)

	ref, err := h.control.CreateWorkflow(&core_itf.ControlWorkflowSpec{
		ProjectDirPath: "/work",
		Steps:          []*core_itf.ControlStepSpec{step("a", "first")},
	})
	if err != nil {
		t.Fatalf("create workflow: %v", err)
	}

	if !ref.Started {
		t.Fatal("the config default asks for an autostart and the workflow did not start")
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

	ref, err := h.control.CreateWorkflow(&core_itf.ControlWorkflowSpec{
		ProjectDirPath: "/work",
		Autostart:      &off,
		Steps:          []*core_itf.ControlStepSpec{step("a", "first")},
	})
	if err != nil {
		t.Fatalf("create workflow: %v", err)
	}

	if ref.Started || len(h.coordinator.ran) != 0 {
		t.Fatal("autostart false still started the workflow")
	}
}

func TestAutostartTrueRunsTheCoordinatorOnce(t *testing.T) {
	h := newHarness(t, anyWorkspaceConfig())
	on := true

	ref, err := h.control.CreateWorkflow(&core_itf.ControlWorkflowSpec{
		ProjectDirPath: "/work",
		Autostart:      &on,
		Steps:          []*core_itf.ControlStepSpec{step("a", "first"), step("b", "second")},
	})
	if err != nil {
		t.Fatalf("create workflow: %v", err)
	}

	if !ref.Started {
		t.Fatal("autostart true did not mark the workflow started")
	}

	if len(h.coordinator.ran) != 1 || h.coordinator.ran[0] != h.workflows.workflowID {
		t.Fatalf("the coordinator ran %v, want one run of %v", h.coordinator.ran, h.workflows.workflowID)
	}
}

func TestAutopilotDropsTheManualAccept(t *testing.T) {
	h := newHarness(t, anyWorkspaceConfig())
	h.userConfig.autopilot = true
	role := withRole(t, h)

	backed := step("a", "review")
	backed.RoleID = role.ID
	backed.Inputs = map[string]string{"ticket": "NEX-1"}

	gated := step("b", "gated")
	gated.PauseForReview = true

	if _, err := h.control.CreateWorkflow(&core_itf.ControlWorkflowSpec{
		ProjectDirPath: "/work",
		Steps:          []*core_itf.ControlStepSpec{backed, gated},
	}); err != nil {
		t.Fatalf("create workflow: %v", err)
	}

	for _, added := range h.workflows.added {
		if added.PauseForReview {
			t.Fatalf("step %s kept its manual accept under autopilot", added.Name)
		}
	}
}

func TestWorkflowStateComesStraightFromTheManager(t *testing.T) {
	h := newHarness(t, anyWorkspaceConfig())

	workflowID := uuid.MustParse("00000000-0000-7000-8000-000000000001")
	want := &core_itf.WorkflowStatus{
		ID:             workflowID,
		Status:         enums.WorkflowProcessing,
		ProjectDirPath: "/work",
		Steps:          map[uuid.UUID]*core_itf.StepResult{},
	}
	h.workflows.statuses[workflowID] = want

	state, err := h.control.WorkflowState(workflowID)
	if err != nil {
		t.Fatalf("workflow state: %v", err)
	}

	if state != want {
		t.Fatalf("state is %p, want manager value %p", state, want)
	}

	if h.storage.loads != 0 {
		t.Fatalf("workflow_status read the step history %d times, want none", h.storage.loads)
	}
}

func TestListWorkflowsClampsTheLimitAndOrdersNewestFirst(t *testing.T) {
	cfg := anyWorkspaceConfig()
	cfg.MaxWorkflowsListed = 2

	h := newHarness(t, cfg)

	base := time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC)

	for i, status := range []enums.WorkflowStatus{enums.WorkflowInit, enums.WorkflowPaused, enums.WorkflowCompleted} {
		id := uuid.New()

		h.storage.snapshots = append(h.storage.snapshots, &input_itf.WorkflowSnapshot{
			Workflow: &input_itf.WorkflowEntity{
				ID:        id,
				TotalStep: i + 1,
				CreatedAt: base.Add(time.Duration(i) * time.Hour),
			},
		})

		h.workflows.statuses[id] = &core_itf.WorkflowStatus{
			ID:             id,
			Status:         status,
			ProjectDirPath: "/work",
			Steps:          make(map[uuid.UUID]*core_itf.StepResult, i+1),
		}
		for step := 0; step <= i; step++ {
			h.workflows.statuses[id].Steps[uuid.New()] = &core_itf.StepResult{}
		}
	}

	for _, limit := range []int{0, -3, 9} {
		listed, err := h.control.ListWorkflows(limit)
		if err != nil {
			t.Fatalf("list workflows with limit %d: %v", limit, err)
		}

		if len(listed) != cfg.MaxWorkflowsListed {
			t.Fatalf("limit %d listed %d workflows, want %d", limit, len(listed), cfg.MaxWorkflowsListed)
		}

		if listed[0].Status != enums.WorkflowCompleted || listed[1].Status != enums.WorkflowPaused {
			t.Fatalf("workflows are not newest first: %s then %s", listed[0].Status, listed[1].Status)
		}

		if len(listed[0].Steps) != 3 || len(listed[1].Steps) != 2 {
			t.Fatalf("step totals are %d and %d, want 3 and 2", len(listed[0].Steps), len(listed[1].Steps))
		}
	}
}

func TestListWorkflowsSkipsAWorkflowTheManagerDoesNotKnow(t *testing.T) {
	h := newHarness(t, anyWorkspaceConfig())

	known := uuid.New()

	h.storage.snapshots = []*input_itf.WorkflowSnapshot{
		{Workflow: &input_itf.WorkflowEntity{ID: known, CreatedAt: time.Now()}},
		{Workflow: &input_itf.WorkflowEntity{ID: uuid.New(), CreatedAt: time.Now().Add(-time.Hour)}},
	}

	h.workflows.statuses[known] = &core_itf.WorkflowStatus{ID: known, Status: enums.WorkflowInit}

	listed, err := h.control.ListWorkflows(0)
	if err != nil {
		t.Fatalf("list workflows: %v", err)
	}

	if len(listed) != 1 || listed[0].ID != known {
		t.Fatalf("listing is %+v, want only the workflow the manager knows", listed)
	}
}

func TestListRolesComesStraightFromTheManager(t *testing.T) {
	h := newHarness(t, anyWorkspaceConfig())
	role := withRole(t, h)

	listed, err := h.control.ListRoles()
	if err != nil {
		t.Fatalf("list roles: %v", err)
	}

	if len(listed) != 1 {
		t.Fatalf("listed %d roles, want 1", len(listed))
	}

	if listed[0] != role {
		t.Fatalf("role is %p, want manager value %p", listed[0], role)
	}
}

func TestLifecycleCallsForwardIDs(t *testing.T) {
	h := newHarness(t, anyWorkspaceConfig())
	workflowID := uuid.New()

	if err := h.control.StartWorkflow(workflowID); err != nil {
		t.Fatalf("start workflow: %v", err)
	}

	if err := h.control.PauseWorkflow(workflowID); err != nil {
		t.Fatalf("pause workflow: %v", err)
	}

	if len(h.coordinator.ran) != 1 || h.coordinator.ran[0] != workflowID {
		t.Fatalf("the coordinator ran %v, want one run of %v", h.coordinator.ran, workflowID)
	}

	if len(h.coordinator.paused) != 1 || h.coordinator.paused[0] != workflowID {
		t.Fatalf("the coordinator paused %v, want %v", h.coordinator.paused, workflowID)
	}
}

func TestAnswerReviewReachesTheWorkflowManager(t *testing.T) {
	h := newHarness(t, anyWorkspaceConfig())
	stepID := uuid.New()

	if err := h.control.AnswerReview(stepID, true); err != nil {
		t.Fatalf("answer acceptance: %v", err)
	}

	if accepted, found := h.workflows.answered[stepID]; !found || !accepted {
		t.Fatalf("the workflow manager saw %v for %v", h.workflows.answered, stepID)
	}
}
