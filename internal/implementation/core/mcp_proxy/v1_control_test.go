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

var errPortRefused = errors.New("session 7f3 is already running")

var (
	sessionOne  = uuid.MustParse("00000000-0000-7000-8000-000000000001")
	templateOne = uuid.MustParse("00000000-0000-7000-8000-000000000002")
	taskExplore = uuid.MustParse("00000000-0000-7000-8000-000000000003")
	taskFix     = uuid.MustParse("00000000-0000-7000-8000-000000000004")
	taskReview  = uuid.MustParse("00000000-0000-7000-8000-000000000005")
)

type fakeSessionControl struct {
	mu        sync.Mutex
	spec      *core_itf.ControlSessionSpec
	started   []uuid.UUID
	paused    []uuid.UUID
	cancelled []uuid.UUID
	answered  map[uuid.UUID]bool
	limit     int
	err       error
}

func newFakeSessionControl() *fakeSessionControl {
	return &fakeSessionControl{answered: map[uuid.UUID]bool{}}
}

func (c *fakeSessionControl) CreateSession(spec *core_itf.ControlSessionSpec) (*core_itf.ControlSessionRef, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.err != nil {
		return nil, c.err
	}

	c.spec = spec

	ids := map[string]uuid.UUID{}
	for _, task := range spec.Tasks {
		ids[task.ClientID] = map[string]uuid.UUID{"explore": taskExplore, "fix": taskFix}[task.ClientID]
	}

	return &core_itf.ControlSessionRef{SessionID: sessionOne, TaskIDs: ids}, nil
}

func (c *fakeSessionControl) StartSession(sessionID uuid.UUID) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.err != nil {
		return c.err
	}

	c.started = append(c.started, sessionID)

	return nil
}

func (c *fakeSessionControl) PauseSession(sessionID uuid.UUID) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.err != nil {
		return c.err
	}

	c.paused = append(c.paused, sessionID)

	return nil
}

func (c *fakeSessionControl) CancelSession(sessionID uuid.UUID) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.err != nil {
		return c.err
	}

	c.cancelled = append(c.cancelled, sessionID)

	return nil
}

func (c *fakeSessionControl) SessionState(sessionID uuid.UUID) (*core_itf.SessionStatus, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.err != nil {
		return nil, c.err
	}

	return &core_itf.SessionStatus{
		ID:             sessionID,
		Status:         enums.SessionProcessing,
		WorkingDirPath: "/tmp/work",
		TokensBilled:   120,
		Tasks: map[uuid.UUID]*core_itf.TaskReport{
			taskReview: {
				TaskID: taskReview,
				Name:   "Review diff",
				Status: enums.TaskAwaitingAccept,
				HandoverDocs: []*core_itf.HandoverDoc{
					{TLDR: "read the diff"},
				},
			},
		},
	}, nil
}

func (c *fakeSessionControl) ListSessions(limit int) ([]*core_itf.SessionStatus, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.err != nil {
		return nil, c.err
	}

	c.limit = limit

	return []*core_itf.SessionStatus{
		{
			ID:             sessionOne,
			Status:         enums.SessionCompleted,
			WorkingDirPath: "/tmp/work",
			Tasks: map[uuid.UUID]*core_itf.TaskReport{
				taskExplore: {},
				taskFix:     {},
			},
		},
	}, nil
}

func (c *fakeSessionControl) ListTemplates() ([]*core_itf.Template, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.err != nil {
		return nil, c.err
	}

	return []*core_itf.Template{
		{
			ID:        templateOne,
			Name:      "Code reviewer",
			Role:      "Reads a diff and reports the defects it can prove.",
			TaskLevel: enums.HeavyTask,
			Params: map[string]*core_itf.TemplateParams{
				"diff_path": {Description: "Where the diff is", Type: "text", Required: true},
			},
		},
	}, nil
}

func (c *fakeSessionControl) AnswerAcceptance(taskID uuid.UUID, accepted bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.err != nil {
		return c.err
	}

	c.answered[taskID] = accepted

	return nil
}

func (c *fakeSessionControl) createdSpec() *core_itf.ControlSessionSpec {
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
	proxy.TrackSessionControl(newFakeSessionControl())

	srv, _ := servedControl(t, proxy)

	answer := callControl(t, srv, proxy.controlToken, methodToolsList, nil)

	got := make([]string, 0, len(answer.Result.Tools))
	for _, tool := range answer.Result.Tools {
		got = append(got, tool.Name)
	}

	want := []string{
		listTemplatesTool,
		createSessionTool,
		startSessionTool,
		pauseSessionTool,
		cancelSessionTool,
		sessionStatusTool,
		listSessionsTool,
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
	proxy.TrackSessionControl(newFakeSessionControl())

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
	proxy.TrackSessionControl(newFakeSessionControl())

	_, gateway := servedControl(t, proxy)

	for _, server := range gateway.Servers {
		if server.Name == constances.ControlLocalServer {
			t.Fatal("the control server is advertised to the agents the app spawns")
		}
	}
}

func TestCreateSessionHandsTheParsedSpecToTheControlPort(t *testing.T) {
	proxy := newTestProxy(t, newFakeMCPStore(), &fakeHttpCli{})
	control := newFakeSessionControl()
	proxy.TrackSessionControl(control)

	srv, _ := servedControl(t, proxy)

	answer := callControl(t, srv, proxy.controlToken, methodToolsCall, map[string]any{
		"name": createSessionTool,
		"arguments": map[string]any{
			"working_dir_path": "/tmp/work",
			"context_dir_path": "/tmp/context",
			"tasks": []map[string]any{
				{
					"client_id":  "explore",
					"name":       "Explore",
					"prompt":     "Find where sessions are persisted.",
					"task_level": "daily_task",
				},
				{
					"client_id":              "fix",
					"name":                   "Fix",
					"prompt":                 "Fix what explore found.",
					"template_id":            templateOne.String(),
					"params":                 map[string]string{"diff_path": "/tmp/diff"},
					"depends_on":             []string{"explore"},
					"auto_retry":             true,
					"manual_accept_required": true,
				},
			},
		},
	})

	if answer.Result.IsError {
		t.Fatalf("create_session failed: %v", answer.Result.Content)
	}

	spec := control.createdSpec()
	if spec == nil {
		t.Fatal("create_session never reached the control port")
	}

	if spec.WorkingDirPath != "/tmp/work" || spec.ContextDirPath != "/tmp/context" {
		t.Errorf("spec dirs = %q and %q, want the ones the caller sent", spec.WorkingDirPath, spec.ContextDirPath)
	}

	if spec.Autostart != nil {
		t.Errorf("autostart = %v, want nil so the app's own setting decides", *spec.Autostart)
	}

	if len(spec.Tasks) != 2 {
		t.Fatalf("spec carries %d tasks, want 2", len(spec.Tasks))
	}

	if spec.Tasks[0].ClientID != "explore" || spec.Tasks[0].TaskLevel != "daily_task" {
		t.Errorf("first task = %+v, want the explore node", spec.Tasks[0])
	}

	fix := spec.Tasks[1]

	if !slices.Equal(fix.DependsOn, []string{"explore"}) {
		t.Errorf("depends_on = %v, want the client id of the node it follows", fix.DependsOn)
	}

	if fix.TemplateID != templateOne || fix.Params["diff_path"] != "/tmp/diff" {
		t.Errorf("second task = %+v, want the template and its params", fix)
	}

	if !fix.AutoRetry || !fix.ManualAcceptRequired {
		t.Errorf("second task flags = %v and %v, want both true", fix.AutoRetry, fix.ManualAcceptRequired)
	}

	payload := struct {
		SessionID string            `json:"session_id"`
		TaskIDs   map[string]string `json:"task_ids"`
	}{}

	if len(answer.Result.Content) == 0 {
		t.Fatal("create_session answered with no content")
	}

	if err := json.Unmarshal([]byte(answer.Result.Content[0].Text), &payload); err != nil {
		t.Fatalf("decode create_session answer: %v", err)
	}

	if payload.SessionID != sessionOne.String() || payload.TaskIDs["explore"] != taskExplore.String() {
		t.Errorf("answer = %+v, want the session id and a real task id per client id", payload)
	}
}

func TestCreateSessionKeepsAnExplicitAutostartFalse(t *testing.T) {
	proxy := newTestProxy(t, newFakeMCPStore(), &fakeHttpCli{})
	control := newFakeSessionControl()
	proxy.TrackSessionControl(control)

	arguments, err := json.Marshal(map[string]any{
		"working_dir_path": "/tmp/work",
		"autostart":        false,
		"tasks": []map[string]any{
			{"client_id": "only", "name": "Only", "prompt": "Do the thing."},
		},
	})
	if err != nil {
		t.Fatalf("encode arguments: %v", err)
	}

	if result := proxy.callCreateSession(arguments, uuid.Nil); result.IsError {
		t.Fatalf("create_session failed: %v", result.Content)
	}

	spec := control.createdSpec()

	if spec.Autostart == nil {
		t.Fatal("an explicit autostart false was flattened into nil")
	}

	if *spec.Autostart {
		t.Error("autostart = true, want the false the caller sent")
	}
}

func TestControlToolsFailWhenNoSessionControlIsTracked(t *testing.T) {
	proxy := newTestProxy(t, newFakeMCPStore(), &fakeHttpCli{})

	for _, tool := range proxy.controlTools() {
		result := tool.call(json.RawMessage(`{}`), uuid.Nil)

		if !result.IsError {
			t.Errorf("%s answered as a success while the app tracks no session control", tool.name)
			continue
		}

		if len(result.Content) == 0 || result.Content[0].Text != controlUnavailable {
			t.Errorf("%s answered %v, want %q", tool.name, result.Content, controlUnavailable)
		}
	}
}

func TestControlToolsSurfaceThePortsErrorAsAToolError(t *testing.T) {
	proxy := newTestProxy(t, newFakeMCPStore(), &fakeHttpCli{})
	proxy.TrackSessionControl(&fakeSessionControl{err: errPortRefused, answered: map[uuid.UUID]bool{}})

	srv, _ := servedControl(t, proxy)

	answer := callControl(t, srv, proxy.controlToken, methodToolsCall, map[string]any{
		"name":      startSessionTool,
		"arguments": map[string]any{"session_id": sessionOne.String()},
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

func TestControlToolsReadTheSessionAndTheAcceptGate(t *testing.T) {
	proxy := newTestProxy(t, newFakeMCPStore(), &fakeHttpCli{})
	control := newFakeSessionControl()
	proxy.TrackSessionControl(control)

	srv, _ := servedControl(t, proxy)

	answer := callControl(t, srv, proxy.controlToken, methodToolsCall, map[string]any{
		"name":      sessionStatusTool,
		"arguments": map[string]any{"session_id": sessionOne.String()},
	})

	if answer.Result.IsError {
		t.Fatalf("session_status failed: %v", answer.Result.Content)
	}

	state := struct {
		SessionID string `json:"session_id"`
		Status    string `json:"status"`
		Tasks     []struct {
			TaskID string `json:"task_id"`
			Status string `json:"status"`
			TLDR   string `json:"tldr"`
		} `json:"tasks"`
		TokensBilled int `json:"tokens_billed"`
	}{}

	if err := json.Unmarshal([]byte(answer.Result.Content[0].Text), &state); err != nil {
		t.Fatalf("decode session_status answer: %v", err)
	}

	if state.SessionID != sessionOne.String() || state.Status != "processing" || state.TokensBilled != 120 {
		t.Errorf("state = %+v, want the session the port reported", state)
	}

	if len(state.Tasks) != 1 || state.Tasks[0].TaskID != taskReview.String() || state.Tasks[0].TLDR != "read the diff" {
		t.Fatalf("tasks = %+v, want the one task the port reported", state.Tasks)
	}

	answer = callControl(t, srv, proxy.controlToken, methodToolsCall, map[string]any{
		"name":      answerAcceptanceTool,
		"arguments": map[string]any{"task_id": taskReview.String(), "accepted": true},
	})

	if answer.Result.IsError {
		t.Fatalf("answer_acceptance failed: %v", answer.Result.Content)
	}

	control.mu.Lock()
	accepted, answered := control.answered[taskReview]
	control.mu.Unlock()

	if !answered || !accepted {
		t.Errorf("the accept gate on task-review was answered %v, %v, want an acceptance", accepted, answered)
	}
}

func TestListSessionsTakesNoArgumentsAtAll(t *testing.T) {
	proxy := newTestProxy(t, newFakeMCPStore(), &fakeHttpCli{})
	control := newFakeSessionControl()
	proxy.TrackSessionControl(control)

	srv, _ := servedControl(t, proxy)

	answer := callControl(t, srv, proxy.controlToken, methodToolsCall, map[string]any{"name": listSessionsTool})

	if answer.Result.IsError {
		t.Fatalf("list_sessions failed without arguments: %v", answer.Result.Content)
	}

	sessions := []struct {
		SessionID string `json:"session_id"`
		TotalTask int    `json:"total_task"`
	}{}

	if err := json.Unmarshal([]byte(answer.Result.Content[0].Text), &sessions); err != nil {
		t.Fatalf("decode list_sessions answer: %v", err)
	}

	if len(sessions) != 1 || sessions[0].SessionID != sessionOne.String() || sessions[0].TotalTask != 2 {
		t.Errorf("sessions = %+v, want the one summary the port reported", sessions)
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
	proxy.TrackSessionControl(newFakeSessionControl())

	result := proxy.callCreateSession(json.RawMessage(`{"tasks": "not an array"}`), uuid.Nil)

	if !result.IsError {
		t.Fatal("a malformed create_session payload was accepted")
	}

	if len(result.Content) == 0 || !bytes.HasPrefix([]byte(result.Content[0].Text), []byte("cannot parse tool arguments: ")) {
		t.Errorf("answer = %v, want a parse failure", result.Content)
	}

	result = proxy.callStartSession(json.RawMessage(`{"session_id": "not-a-uuid"}`), uuid.Nil)
	if !result.IsError || len(result.Content) == 0 || !strings.Contains(result.Content[0].Text, "invalid session id") {
		t.Errorf("answer = %v, want an invalid session id failure", result.Content)
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
