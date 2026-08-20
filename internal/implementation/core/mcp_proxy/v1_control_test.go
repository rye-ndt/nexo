package mcp_proxy

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"

	"hexago/internal/helpers/constances"
	"hexago/internal/helpers/enums"
	core_itf "hexago/internal/interface/core"

	"github.com/google/uuid"
)

var errPortRefused = errors.New("workflow 7f3 is already running")

var (
	workflowOne = uuid.MustParse("00000000-0000-7000-8000-000000000001")
	roleOne     = uuid.MustParse("00000000-0000-7000-8000-000000000002")
	stepExplore = uuid.MustParse("00000000-0000-7000-8000-000000000003")
	stepFix     = uuid.MustParse("00000000-0000-7000-8000-000000000004")
	stepReview  = uuid.MustParse("00000000-0000-7000-8000-000000000005")
)

type fakeWorkflowControl struct {
	mu        sync.Mutex
	spec      *core_itf.ControlWorkflowSpec
	started   []uuid.UUID
	paused    []uuid.UUID
	cancelled []uuid.UUID
	answered  map[uuid.UUID]bool
	limit     int
	err       error
}

func newFakeWorkflowControl() *fakeWorkflowControl {
	return &fakeWorkflowControl{answered: map[uuid.UUID]bool{}}
}

func (c *fakeWorkflowControl) CreateWorkflow(spec *core_itf.ControlWorkflowSpec) (*core_itf.ControlWorkflowRef, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.err != nil {
		return nil, c.err
	}

	c.spec = spec

	ids := map[string]uuid.UUID{}
	for _, step := range spec.Steps {
		ids[step.ClientID] = map[string]uuid.UUID{"explore": stepExplore, "fix": stepFix}[step.ClientID]
	}

	return &core_itf.ControlWorkflowRef{WorkflowID: workflowOne, StepIDs: ids}, nil
}

func (c *fakeWorkflowControl) StartWorkflow(workflowID uuid.UUID) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.err != nil {
		return c.err
	}

	c.started = append(c.started, workflowID)

	return nil
}

func (c *fakeWorkflowControl) PauseWorkflow(workflowID uuid.UUID) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.err != nil {
		return c.err
	}

	c.paused = append(c.paused, workflowID)

	return nil
}

func (c *fakeWorkflowControl) CancelWorkflow(workflowID uuid.UUID) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.err != nil {
		return c.err
	}

	c.cancelled = append(c.cancelled, workflowID)

	return nil
}

func (c *fakeWorkflowControl) WorkflowState(workflowID uuid.UUID) (*core_itf.WorkflowStatus, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.err != nil {
		return nil, c.err
	}

	return &core_itf.WorkflowStatus{
		ID:             workflowID,
		Status:         enums.WorkflowProcessing,
		ProjectDirPath: "/tmp/work",
		TokensBilled:   120,
		Steps: map[uuid.UUID]*core_itf.StepResult{
			stepReview: {
				StepID: stepReview,
				Name:   "Review diff",
				Status: enums.StepAwaitingReview,
				Handoffs: []*core_itf.Handoff{
					{TLDR: "read the diff"},
				},
			},
		},
	}, nil
}

func (c *fakeWorkflowControl) ListWorkflows(limit int) ([]*core_itf.WorkflowStatus, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.err != nil {
		return nil, c.err
	}

	c.limit = limit

	return []*core_itf.WorkflowStatus{
		{
			ID:             workflowOne,
			Status:         enums.WorkflowCompleted,
			ProjectDirPath: "/tmp/work",
			Steps: map[uuid.UUID]*core_itf.StepResult{
				stepExplore: {},
				stepFix:     {},
			},
		},
	}, nil
}

func (c *fakeWorkflowControl) ListRoles() ([]*core_itf.Role, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.err != nil {
		return nil, c.err
	}

	return []*core_itf.Role{
		{
			ID:          roleOne,
			Name:        "Code reviewer",
			Description: "Reads a diff and reports the defects it can prove.",
			Effort:      enums.EffortDeep,
			Inputs: map[string]*core_itf.RoleInputs{
				"diff_path": {Description: "Where the diff is", Type: "text", Required: true},
			},
		},
	}, nil
}

func (c *fakeWorkflowControl) AnswerReview(stepID uuid.UUID, accepted bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.err != nil {
		return c.err
	}

	c.answered[stepID] = accepted

	return nil
}

func (c *fakeWorkflowControl) createdSpec() *core_itf.ControlWorkflowSpec {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.spec
}

type rpcAnswer struct {
	Result struct {
		Tools []struct {
			Name string `json:"name"`
		} `json:"tools"`
		Content []toolContent `json:"content"`
		IsError bool          `json:"isError"`
	} `json:"result"`
	Error *rpcError `json:"error"`
}

func servedControl(t *testing.T, proxy *v1) (*httptest.Server, *core_itf.MCPGateway) {
	t.Helper()

	gateway, err := proxy.Serve()
	if err != nil {
		t.Fatalf("serve gateway: %v", err)
	}

	t.Cleanup(func() { proxy.Close() })

	srv := httptest.NewServer(proxy.gatewayHttpServer.Handler)
	t.Cleanup(srv.Close)

	return srv, gateway
}

func postControl(t *testing.T, srv *httptest.Server, header, token string, body map[string]any) *http.Response {
	t.Helper()

	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("encode request: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, srv.URL+controlPath, bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	req.Header.Set(header, token)

	res, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("call control endpoint: %v", err)
	}

	t.Cleanup(func() { res.Body.Close() })

	return res
}

func callControl(t *testing.T, srv *httptest.Server, token, method string, params any) *rpcAnswer {
	t.Helper()

	res := postControl(t, srv, constances.ControlTokenHeader, token, map[string]any{
		"jsonrpc": jsonRPCVersion,
		"id":      1,
		"method":  method,
		"params":  params,
	})

	if res.StatusCode != http.StatusOK {
		t.Fatalf("control endpoint returned %d, want 200", res.StatusCode)
	}

	answer := &rpcAnswer{}
	if err := json.NewDecoder(res.Body).Decode(answer); err != nil {
		t.Fatalf("decode answer: %v", err)
	}

	if answer.Error != nil {
		t.Fatalf("control call failed at the rpc level: %+v", answer.Error)
	}

	return answer
}

func TestControlServerExposesEveryControlTool(t *testing.T) {
	proxy := newTestProxy(t, newFakeMCPStore(), &fakeHttpCli{})
	proxy.TrackWorkflowControl(newFakeWorkflowControl())

	srv, _ := servedControl(t, proxy)

	answer := callControl(t, srv, proxy.controlToken, methodToolsList, nil)

	got := make([]string, 0, len(answer.Result.Tools))
	for _, tool := range answer.Result.Tools {
		got = append(got, tool.Name)
	}

	want := []string{
		listRolesTool,
		createWorkflowTool,
		startWorkflowTool,
		pauseWorkflowTool,
		cancelWorkflowTool,
		workflowStatusTool,
		listWorkflowsTool,
		answerAcceptanceTool,
	}

	slices.Sort(got)
	slices.Sort(want)

	if !slices.Equal(got, want) {
		t.Fatalf("control tools = %v, want %v", got, want)
	}
}

func TestControlEndpointRejectsTheHarnessGatewayToken(t *testing.T) {
	proxy := newTestProxy(t, newFakeMCPStore(), &fakeHttpCli{})
	proxy.TrackWorkflowControl(newFakeWorkflowControl())

	srv, gateway := servedControl(t, proxy)

	body := map[string]any{"jsonrpc": jsonRPCVersion, "id": 1, "method": methodToolsList}

	res := postControl(t, srv, gatewayTokenHeader, gateway.Token, body)
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("the harness gateway token reached the control endpoint with %d, want 403", res.StatusCode)
	}

	res = postControl(t, srv, constances.ControlTokenHeader, gateway.Token, body)
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("the gateway token in the control header returned %d, want 403", res.StatusCode)
	}

	res = postControl(t, srv, constances.ControlTokenHeader, proxy.controlToken, body)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("the control token returned %d, want 200", res.StatusCode)
	}
}

func TestControlToolsAreNotOfferedOnTheHarnessGateway(t *testing.T) {
	proxy := newTestProxy(t, newFakeMCPStore(), &fakeHttpCli{})
	proxy.TrackWorkflowControl(newFakeWorkflowControl())

	_, gateway := servedControl(t, proxy)

	for _, server := range gateway.Servers {
		if server.Name == constances.ControlLocalServer {
			t.Fatal("the control server is advertised to the agents the app spawns")
		}
	}
}

func TestCreateWorkflowHandsTheParsedSpecToTheControlPort(t *testing.T) {
	proxy := newTestProxy(t, newFakeMCPStore(), &fakeHttpCli{})
	control := newFakeWorkflowControl()
	proxy.TrackWorkflowControl(control)

	srv, _ := servedControl(t, proxy)

	answer := callControl(t, srv, proxy.controlToken, methodToolsCall, map[string]any{
		"name": createWorkflowTool,
		"arguments": map[string]any{
			"project_dir_path": "/tmp/work",
			"steps": []map[string]any{
				{
					"client_id": "explore",
					"name":      "Explore",
					"prompt":    "Find where workflows are persisted.",
					"effort":    "standard",
				},
				{
					"client_id":        "fix",
					"name":             "Fix",
					"prompt":           "Fix what explore found.",
					"role_id":          roleOne.String(),
					"inputs":           map[string]string{"diff_path": "/tmp/diff"},
					"depends_on":       []string{"explore"},
					"auto_retry":       true,
					"pause_for_review": true,
				},
			},
		},
	})

	if answer.Result.IsError {
		t.Fatalf("create_workflow failed: %v", answer.Result.Content)
	}

	spec := control.createdSpec()
	if spec == nil {
		t.Fatal("create_workflow never reached the control port")
	}

	if spec.ProjectDirPath != "/tmp/work" {
		t.Errorf("spec project dir = %q, want the one the caller sent", spec.ProjectDirPath)
	}

	if spec.Autostart != nil {
		t.Errorf("autostart = %v, want nil so the app's own setting decides", *spec.Autostart)
	}

	if len(spec.Steps) != 2 {
		t.Fatalf("spec carries %d steps, want 2", len(spec.Steps))
	}

	if spec.Steps[0].ClientID != "explore" || spec.Steps[0].Effort != "standard" {
		t.Errorf("first step = %+v, want the explore node", spec.Steps[0])
	}

	fix := spec.Steps[1]

	if !slices.Equal(fix.DependsOn, []string{"explore"}) {
		t.Errorf("depends_on = %v, want the client id of the step it follows", fix.DependsOn)
	}

	if fix.RoleID != roleOne || fix.Inputs["diff_path"] != "/tmp/diff" {
		t.Errorf("second step = %+v, want the role and its inputs", fix)
	}

	if !fix.AutoRetry || !fix.PauseForReview {
		t.Errorf("second step flags = %v and %v, want both true", fix.AutoRetry, fix.PauseForReview)
	}

	payload := struct {
		WorkflowID string            `json:"workflow_id"`
		StepIDs    map[string]string `json:"step_ids"`
	}{}

	if len(answer.Result.Content) == 0 {
		t.Fatal("create_workflow answered with no content")
	}

	if err := json.Unmarshal([]byte(answer.Result.Content[0].Text), &payload); err != nil {
		t.Fatalf("decode create_workflow answer: %v", err)
	}

	if payload.WorkflowID != workflowOne.String() || payload.StepIDs["explore"] != stepExplore.String() {
		t.Errorf("answer = %+v, want the workflow id and a real step id per client id", payload)
	}
}

func TestCreateWorkflowIgnoresARetiredContextDirArgument(t *testing.T) {
	proxy := newTestProxy(t, newFakeMCPStore(), &fakeHttpCli{})
	control := newFakeWorkflowControl()
	proxy.TrackWorkflowControl(control)

	srv, _ := servedControl(t, proxy)

	answer := callControl(t, srv, proxy.controlToken, methodToolsCall, map[string]any{
		"name": createWorkflowTool,
		"arguments": map[string]any{
			"project_dir_path": "/tmp/work",
			"context_dir_path": "/tmp/context",
			"steps": []map[string]any{
				{"client_id": "only", "name": "Only", "prompt": "Do the thing."},
			},
		},
	})

	if answer.Result.IsError {
		t.Fatalf("create_workflow refused an argument it should ignore: %v", answer.Result.Content)
	}

	if control.createdSpec() == nil {
		t.Fatal("create_workflow never reached the control port")
	}
}

func TestCreateWorkflowKeepsAnExplicitAutostartFalse(t *testing.T) {
	proxy := newTestProxy(t, newFakeMCPStore(), &fakeHttpCli{})
	control := newFakeWorkflowControl()
	proxy.TrackWorkflowControl(control)

	arguments, err := json.Marshal(map[string]any{
		"project_dir_path": "/tmp/work",
		"autostart":        false,
		"steps": []map[string]any{
			{"client_id": "only", "name": "Only", "prompt": "Do the thing."},
		},
	})
	if err != nil {
		t.Fatalf("encode arguments: %v", err)
	}

	if result := proxy.callCreateWorkflow(arguments, uuid.Nil); result.IsError {
		t.Fatalf("create_workflow failed: %v", result.Content)
	}

	spec := control.createdSpec()

	if spec.Autostart == nil {
		t.Fatal("an explicit autostart false was flattened into nil")
	}

	if *spec.Autostart {
		t.Error("autostart = true, want the false the caller sent")
	}
}

func TestControlToolsFailWhenNoWorkflowControlIsTracked(t *testing.T) {
	proxy := newTestProxy(t, newFakeMCPStore(), &fakeHttpCli{})

	for _, tool := range proxy.controlTools() {
		result := tool.call(json.RawMessage(`{}`), uuid.Nil)

		if !result.IsError {
			t.Errorf("%s answered as a success while the app tracks no workflow control", tool.name)
			continue
		}

		if len(result.Content) == 0 || result.Content[0].Text != controlUnavailable {
			t.Errorf("%s answered %v, want %q", tool.name, result.Content, controlUnavailable)
		}
	}
}

func TestControlToolsSurfaceThePortsErrorAsAToolError(t *testing.T) {
	proxy := newTestProxy(t, newFakeMCPStore(), &fakeHttpCli{})
	proxy.TrackWorkflowControl(&fakeWorkflowControl{err: errPortRefused, answered: map[uuid.UUID]bool{}})

	srv, _ := servedControl(t, proxy)

	answer := callControl(t, srv, proxy.controlToken, methodToolsCall, map[string]any{
		"name":      startWorkflowTool,
		"arguments": map[string]any{"workflow_id": workflowOne.String()},
	})

	if !answer.Result.IsError {
		t.Fatal("a refusal from the control port was reported as a success")
	}

	if len(answer.Result.Content) == 0 || answer.Result.Content[0].Text != errPortRefused.Error() {
		t.Errorf("answer = %v, want the port's reason", answer.Result.Content)
	}

	for _, tool := range proxy.controlTools() {
		result := tool.call(json.RawMessage(`{}`), uuid.Nil)

		if !result.IsError {
			t.Errorf("%s answered as a success while the control port refuses everything", tool.name)
		}
	}
}

func TestControlToolsReadTheWorkflowAndTheAcceptGate(t *testing.T) {
	proxy := newTestProxy(t, newFakeMCPStore(), &fakeHttpCli{})
	control := newFakeWorkflowControl()
	proxy.TrackWorkflowControl(control)

	srv, _ := servedControl(t, proxy)

	answer := callControl(t, srv, proxy.controlToken, methodToolsCall, map[string]any{
		"name":      workflowStatusTool,
		"arguments": map[string]any{"workflow_id": workflowOne.String()},
	})

	if answer.Result.IsError {
		t.Fatalf("workflow_status failed: %v", answer.Result.Content)
	}

	state := struct {
		WorkflowID string `json:"workflow_id"`
		Status     string `json:"status"`
		Steps      []struct {
			StepID string `json:"step_id"`
			Status string `json:"status"`
			TLDR   string `json:"tldr"`
		} `json:"steps"`
		TokensBilled int `json:"tokens_billed"`
	}{}

	if err := json.Unmarshal([]byte(answer.Result.Content[0].Text), &state); err != nil {
		t.Fatalf("decode workflow_status answer: %v", err)
	}

	if state.WorkflowID != workflowOne.String() || state.Status != "processing" || state.TokensBilled != 120 {
		t.Errorf("state = %+v, want the workflow the port reported", state)
	}

	if len(state.Steps) != 1 || state.Steps[0].StepID != stepReview.String() || state.Steps[0].TLDR != "read the diff" {
		t.Fatalf("steps = %+v, want the one step the port reported", state.Steps)
	}

	answer = callControl(t, srv, proxy.controlToken, methodToolsCall, map[string]any{
		"name":      answerAcceptanceTool,
		"arguments": map[string]any{"step_id": stepReview.String(), "accepted": true},
	})

	if answer.Result.IsError {
		t.Fatalf("answer_review failed: %v", answer.Result.Content)
	}

	control.mu.Lock()
	accepted, answered := control.answered[stepReview]
	control.mu.Unlock()

	if !answered || !accepted {
		t.Errorf("the accept gate on step-review was answered %v, %v, want an acceptance", accepted, answered)
	}
}

func TestListWorkflowsTakesNoArgumentsAtAll(t *testing.T) {
	proxy := newTestProxy(t, newFakeMCPStore(), &fakeHttpCli{})
	control := newFakeWorkflowControl()
	proxy.TrackWorkflowControl(control)

	srv, _ := servedControl(t, proxy)

	answer := callControl(t, srv, proxy.controlToken, methodToolsCall, map[string]any{"name": listWorkflowsTool})

	if answer.Result.IsError {
		t.Fatalf("list_workflows failed without arguments: %v", answer.Result.Content)
	}

	workflows := []struct {
		WorkflowID string `json:"workflow_id"`
		TotalStep  int    `json:"total_step"`
	}{}

	if err := json.Unmarshal([]byte(answer.Result.Content[0].Text), &workflows); err != nil {
		t.Fatalf("decode list_workflows answer: %v", err)
	}

	if len(workflows) != 1 || workflows[0].WorkflowID != workflowOne.String() || workflows[0].TotalStep != 2 {
		t.Errorf("workflows = %+v, want the one summary the port reported", workflows)
	}

	control.mu.Lock()
	limit := control.limit
	control.mu.Unlock()

	if limit != 0 {
		t.Errorf("limit = %d, want 0 so the app picks its own default", limit)
	}
}

func TestControlToolsRejectMalformedArguments(t *testing.T) {
	proxy := newTestProxy(t, newFakeMCPStore(), &fakeHttpCli{})
	proxy.TrackWorkflowControl(newFakeWorkflowControl())

	result := proxy.callCreateWorkflow(json.RawMessage(`{"steps": "not an array"}`), uuid.Nil)

	if !result.IsError {
		t.Fatal("a malformed create_workflow payload was accepted")
	}

	if len(result.Content) == 0 || !bytes.HasPrefix([]byte(result.Content[0].Text), []byte("cannot parse tool arguments: ")) {
		t.Errorf("answer = %v, want a parse failure", result.Content)
	}

	result = proxy.callStartWorkflow(json.RawMessage(`{"workflow_id": "not-a-uuid"}`), uuid.Nil)
	if !result.IsError || len(result.Content) == 0 || !strings.Contains(result.Content[0].Text, "invalid workflow id") {
		t.Errorf("answer = %v, want an invalid workflow id failure", result.Content)
	}
}

func TestAuthenticatingStripsBothTokensBeforeAnythingIsForwarded(t *testing.T) {
	proxy := newTestProxy(t, newFakeMCPStore(), &fakeHttpCli{})

	gateway, err := proxy.Serve()
	if err != nil {
		t.Fatalf("serve: %v", err)
	}
	defer proxy.Close()

	seen := http.Header{}

	guarded := proxy.authenticated(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		seen = r.Header.Clone()
	}))

	request := httptest.NewRequest(http.MethodPost, constances.GatewayMCPPath+"atlassian", nil)
	request.Header.Set(gatewayTokenHeader, gateway.Token)
	request.Header.Set(constances.ControlTokenHeader, proxy.controlToken)

	guarded.ServeHTTP(httptest.NewRecorder(), request)

	if got := seen.Get(gatewayTokenHeader); got != "" {
		t.Fatalf("the gateway token survived into the forwarded request: %q", got)
	}

	if got := seen.Get(constances.ControlTokenHeader); got != "" {
		t.Fatalf("the control token would be forwarded to a remote mcp server: %q", got)
	}
}
