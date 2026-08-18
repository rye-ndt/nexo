package session_control_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"hexago/internal/helpers/constances"
	"hexago/internal/helpers/enums"
	"hexago/internal/implementation/core/coordinator"
	"hexago/internal/implementation/core/manual_approval_broker"
	"hexago/internal/implementation/core/mcp_proxy"
	"hexago/internal/implementation/core/session_control"
	"hexago/internal/implementation/core/session_manager"
	"hexago/internal/implementation/core/template_manager"
	"hexago/internal/implementation/input/storage"
	"hexago/internal/implementation/input/template_archive"
	"hexago/internal/implementation/input/workspace_history"
	"hexago/internal/implementation/output/message_queue"
	core_itf "hexago/internal/interface/core"
	input_itf "hexago/internal/interface/input"
	output_itf "hexago/internal/interface/output"

	"github.com/google/uuid"
)

var shimBinary string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "nexo-shim")
	if err != nil {
		os.Exit(1)
	}

	binary := filepath.Join(dir, "nexo-mcp")

	if out, err := exec.Command("go", "build", "-o", binary, "hexago/cmd/nexo-mcp").CombinedOutput(); err != nil {
		os.Stderr.Write(out)
		os.RemoveAll(dir)
		os.Exit(1)
	}

	shimBinary = binary

	code := m.Run()

	os.RemoveAll(dir)
	os.Exit(code)
}

type spawnedAgent struct {
	id     uuid.UUID
	prompt string
}

type stubAgents struct {
	core_itf.AgentManager

	mu      sync.Mutex
	spawned []*spawnedAgent
}

func (a *stubAgents) RequestInstance(*core_itf.AgentRequest) (*core_itf.Agent, error) {
	agent := &core_itf.Agent{ID: uuid.New(), HealthStatus: enums.Healthy}

	a.mu.Lock()
	defer a.mu.Unlock()

	a.spawned = append(a.spawned, &spawnedAgent{id: agent.ID})

	return agent, nil
}

func (a *stubAgents) Send(agentID uuid.UUID, message string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	for _, agent := range a.spawned {
		if agent.id == agentID {
			agent.prompt = message
		}
	}

	return nil
}

func (a *stubAgents) Listen(uuid.UUID) (<-chan string, error) { return nil, nil }

func (a *stubAgents) ContextUsage(uuid.UUID) (*input_itf.ContextUsage, error) { return nil, nil }

func (a *stubAgents) Activity(uuid.UUID) ([]input_itf.Activity, error) { return nil, nil }

func (a *stubAgents) Kill(uuid.UUID) error { return nil }

func (a *stubAgents) HeartBeat(uuid.UUID) error { return nil }

func (a *stubAgents) promptFor(name string) (uuid.UUID, string, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()

	for _, agent := range a.spawned {
		if strings.Contains(agent.prompt, "# Task: "+name+"\n") {
			return agent.id, agent.prompt, true
		}
	}

	return uuid.Nil, "", false
}

type stubUserConfig struct {
	output_itf.UserConfig
}

func (c *stubUserConfig) AgentDefault(enums.TaskLevel) (*output_itf.AgentDefault, error) {
	return &output_itf.AgentDefault{Model: enums.Sonnet, ThinkingLevel: enums.LowThinking}, nil
}

func (c *stubUserConfig) Autopilot() bool { return false }

type stubLogger struct{}

func (l *stubLogger) Debug(string, ...any) {}
func (l *stubLogger) Info(string, ...any)  {}
func (l *stubLogger) Warn(string, ...any)  {}
func (l *stubLogger) Error(string, ...any) {}

type stubHttpCli struct {
	input_itf.HttpCli
}

type harness struct {
	baseURL      string
	gatewayToken string
	controlToken string
	agents       *stubAgents
	workingDir   string
}

func e2eControlConfig() *input_itf.ControlConfig {
	return &input_itf.ControlConfig{
		Enabled:            true,
		EndpointFile:       "control.json",
		AllowAnyWorkspace:  true,
		MaxTasksPerSession: 8,
		MaxSessionsListed:  10,
	}
}

func e2eMCPConfig() *input_itf.MCPServersConfig {
	return &input_itf.MCPServersConfig{
		EncodeKey:        "sk-test",
		Control:          e2eControlConfig(),
		SupportedServers: map[string]*input_itf.MCPServerConfig{},
		DefaultTokenTTL:  time.Hour,
		ShutdownGrace:    time.Second,
		Chrome: &input_itf.MCPChromeConfig{
			DebugPort:     9222,
			ProfileDir:    "chrome",
			LaunchTimeout: time.Second,
			CallTimeout:   time.Second,
			MaxPageChars:  100,
		},
	}
}

func e2eSessionConfig() *input_itf.SessionConfig {
	return &input_itf.SessionConfig{
		HeartbeatTimeout:       time.Hour,
		HeartbeatScanInterval:  time.Hour,
		AgentHeartbeatInterval: time.Hour,
	}
}

func newHarness(t *testing.T) *harness {
	t.Helper()

	dataDir := t.TempDir()

	store, err := storage.New(filepath.Join(dataDir, "harness.db"))
	if err != nil {
		t.Fatalf("init storage: %v", err)
	}

	sessions, err := session_manager.InitV1(e2eSessionConfig(), store.TaskStore(), message_queue.InitV1())
	if err != nil {
		t.Fatalf("init session manager: %v", err)
	}
	t.Cleanup(sessions.Stop)

	broker, err := manual_approval_broker.InitV1(&input_itf.ApprovalBrokerConfig{Timeout: time.Minute})
	if err != nil {
		t.Fatalf("init approval broker: %v", err)
	}
	t.Cleanup(broker.Stop)

	proxy, err := mcp_proxy.InitV1(
		e2eMCPConfig(),
		dataDir,
		store.MCPStore(),
		&stubHttpCli{},
		broker,
		sessions,
	)
	if err != nil {
		t.Fatalf("init mcp proxy: %v", err)
	}

	gateway, err := proxy.Serve()
	if err != nil {
		t.Fatalf("serve mcp proxy: %v", err)
	}
	t.Cleanup(func() { proxy.Close() })

	agents := &stubAgents{}
	sessions.TrackLiveAgents(agents)

	history, err := workspace_history.InitV1(filepath.Join(dataDir, "sessions"))
	if err != nil {
		t.Fatalf("init workspace history: %v", err)
	}

	runner, err := coordinator.InitV1(e2eSessionConfig(), sessions, agents, history, &stubLogger{})
	if err != nil {
		t.Fatalf("init coordinator: %v", err)
	}
	t.Cleanup(runner.Stop)

	templates, err := template_manager.InitV1(store.TemplateStore(), template_archive.InitV1())
	if err != nil {
		t.Fatalf("init template manager: %v", err)
	}

	control, err := session_control.InitV1(
		e2eControlConfig(),
		sessions,
		runner,
		templates,
		&stubUserConfig{},
		store.TaskStore(),
	)
	if err != nil {
		t.Fatalf("init session control: %v", err)
	}

	proxy.TrackSessionControl(control)

	return &harness{
		baseURL:      gateway.BaseURL,
		gatewayToken: gateway.Token,
		controlToken: readControlToken(t, filepath.Join(dataDir, "control.json"), gateway.BaseURL),
		agents:       agents,
		workingDir:   t.TempDir(),
	}
}

func readControlToken(t *testing.T, path, baseURL string) string {
	t.Helper()

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read control endpoint: %v", err)
	}

	published := struct {
		URL         string `json:"url"`
		Token       string `json:"token"`
		TokenHeader string `json:"token_header"`
		PID         int    `json:"pid"`
	}{}

	if err := json.Unmarshal(raw, &published); err != nil {
		t.Fatalf("parse control endpoint: %v", err)
	}

	if want := baseURL + constances.GatewayMCPPath + constances.ControlLocalServer; published.URL != want {
		t.Fatalf("control endpoint points at %s, want %s", published.URL, want)
	}

	if published.TokenHeader != constances.ControlTokenHeader {
		t.Fatalf("control endpoint names header %s, want %s", published.TokenHeader, constances.ControlTokenHeader)
	}

	if published.PID != os.Getpid() {
		t.Fatalf("control endpoint claims pid %d, want %d", published.PID, os.Getpid())
	}

	return published.Token
}

func (h *harness) callControl(t *testing.T, tool string, arguments map[string]any) map[string]any {
	t.Helper()

	return h.callTool(t, constances.ControlLocalServer, constances.ControlTokenHeader, h.controlToken, tool, arguments, "")
}

func (h *harness) callHarness(t *testing.T, agentID uuid.UUID, tool string, arguments map[string]any) map[string]any {
	t.Helper()

	return h.callTool(t, constances.GatewayLocalServer, "X-Harness-Gateway-Token", h.gatewayToken, tool, arguments, agentID.String())
}

func (h *harness) callTool(
	t *testing.T,
	server, tokenHeader, token, tool string,
	arguments map[string]any,
	agentID string,
) map[string]any {
	t.Helper()

	body, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params":  map[string]any{"name": tool, "arguments": arguments},
	})
	if err != nil {
		t.Fatalf("encode %s call: %v", tool, err)
	}

	url := h.baseURL + constances.GatewayMCPPath + server

	request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("build %s call: %v", tool, err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set(tokenHeader, token)

	if agentID != "" {
		request.Header.Set(constances.GatewayAgentHeader, agentID)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("send %s call: %v", tool, err)
	}
	defer response.Body.Close()

	raw, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read %s answer: %v", tool, err)
	}

	if response.StatusCode != http.StatusOK {
		t.Fatalf("%s answered %s: %s", tool, response.Status, raw)
	}

	answer := struct {
		Result struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
			IsError bool `json:"isError"`
		} `json:"result"`
	}{}

	if err := json.Unmarshal(raw, &answer); err != nil {
		t.Fatalf("parse %s answer: %v", tool, err)
	}

	if len(answer.Result.Content) == 0 {
		t.Fatalf("%s answered with no content: %s", tool, raw)
	}

	if answer.Result.IsError {
		t.Fatalf("%s failed: %s", tool, answer.Result.Content[0].Text)
	}

	payload := map[string]any{}
	if err := json.Unmarshal([]byte(answer.Result.Content[0].Text), &payload); err != nil {
		return map[string]any{"text": answer.Result.Content[0].Text}
	}

	return payload
}

func waitUntil(t *testing.T, what string, done func() bool) {
	t.Helper()

	deadline := time.Now().Add(10 * time.Second)

	for time.Now().Before(deadline) {
		if done() {
			return
		}

		time.Sleep(20 * time.Millisecond)
	}

	t.Fatalf("timed out waiting for %s", what)
}

func taskStateIn(state map[string]any, name string) map[string]any {
	tasks, ok := state["tasks"].([]any)
	if !ok {
		return nil
	}

	for _, task := range tasks {
		entry, ok := task.(map[string]any)
		if !ok {
			continue
		}

		if entry["name"] == name {
			return entry
		}
	}

	return nil
}

func TestAControlCallerRunsAGatedGraphEndToEnd(t *testing.T) {
	h := newHarness(t)

	created := h.callControl(t, "create_session", map[string]any{
		"working_dir_path": h.workingDir,
		"autostart":        true,
		"tasks": []map[string]any{
			{
				"client_id":              "survey",
				"name":                   "survey the ports",
				"prompt":                 "List every port in {{area}}.",
				"params":                 map[string]string{"area": "interface/core"},
				"manual_accept_required": true,
			},
			{
				"client_id":  "wire",
				"name":       "wire the port",
				"prompt":     "Wire what the survey found.",
				"depends_on": []string{"survey"},
			},
		},
	})

	if created["started"] != true {
		t.Fatalf("create_session did not start the run: %v", created)
	}

	sessionID, _ := created["session_id"].(string)
	if sessionID == "" {
		t.Fatalf("create_session returned no session id: %v", created)
	}

	taskIDs, _ := created["task_ids"].(map[string]any)
	if len(taskIDs) != 2 {
		t.Fatalf("create_session mapped %d client ids, want 2", len(taskIDs))
	}

	var surveyAgent uuid.UUID
	var surveyPrompt string

	waitUntil(t, "the survey task to reach an agent", func() bool {
		id, prompt, found := h.agents.promptFor("survey the ports")
		surveyAgent, surveyPrompt = id, prompt

		return found
	})

	if !strings.Contains(surveyPrompt, "List every port in interface/core.") {
		t.Fatalf("the param was not rendered into the prompt:\n%s", surveyPrompt)
	}

	if _, _, found := h.agents.promptFor("wire the port"); found {
		t.Fatal("the dependent task ran before its dependency finished")
	}

	h.callHarness(t, surveyAgent, "report_task", map[string]any{
		"status":  string(enums.TaskCompleted),
		"tldr":    "Listed the doors this program opens to the outside world.",
		"outcome": "seven ports, each with the call that reaches it",
		"nuances": map[string]string{"scope": "stayed inside interface/core"},
	})

	waitUntil(t, "the gate to open", func() bool {
		state := h.callControl(t, "session_status", map[string]any{"session_id": sessionID})
		task := taskStateIn(state, "survey the ports")

		return task != nil && task["status"] == string(enums.TaskAwaitingAccept)
	})

	if _, _, found := h.agents.promptFor("wire the port"); found {
		t.Fatal("the dependent task ran while the gate was still closed")
	}

	gated := h.callControl(t, "session_status", map[string]any{"session_id": sessionID})
	surveyTask := taskStateIn(gated, "survey the ports")

	if surveyTask["tldr"] != "Listed the doors this program opens to the outside world." {
		t.Fatalf("the tldr did not come back through session_status: %v", surveyTask)
	}

	h.callControl(t, "answer_acceptance", map[string]any{
		"task_id":  taskIDs["survey"],
		"accepted": true,
	})

	var wirePrompt string

	waitUntil(t, "the dependent task to be released", func() bool {
		_, prompt, found := h.agents.promptFor("wire the port")
		wirePrompt = prompt

		return found
	})

	if !strings.Contains(wirePrompt, "seven ports, each with the call that reaches it") {
		t.Fatalf("the handover doc did not reach the dependent task:\n%s", wirePrompt)
	}

	listed := h.callControl(t, "list_sessions", map[string]any{})

	text, isText := listed["text"].(string)
	if !isText || !strings.Contains(text, sessionID) {
		t.Fatalf("list_sessions did not carry the session back: %v", listed)
	}

	h.callControl(t, "cancel_session", map[string]any{"session_id": sessionID})

	waitUntil(t, "the session to report itself cancelled", func() bool {
		state := h.callControl(t, "session_status", map[string]any{"session_id": sessionID})
		task := taskStateIn(state, "wire the port")

		return task != nil && task["status"] == string(enums.TaskCancelled)
	})
}

func buildShim(t *testing.T) string {
	t.Helper()

	if shimBinary == "" {
		t.Skip("the shim was not built for this run")
	}

	return shimBinary
}

func TestTheStdioShimCarriesAControlCallIntoTheApp(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, "config"))

	configDir, err := os.UserConfigDir()
	if err != nil {
		t.Fatalf("user config dir: %v", err)
	}

	h := newHarness(t)
	shim := buildShim(t)

	published := filepath.Join(configDir, "nexo", "control.json")
	if err := os.MkdirAll(filepath.Dir(published), 0o700); err != nil {
		t.Fatalf("make endpoint dir: %v", err)
	}

	endpoint, err := json.Marshal(map[string]any{
		"url":          h.baseURL + constances.GatewayMCPPath + constances.ControlLocalServer,
		"token":        h.controlToken,
		"token_header": constances.ControlTokenHeader,
		"pid":          os.Getpid(),
	})
	if err != nil {
		t.Fatalf("encode endpoint: %v", err)
	}

	if err := os.WriteFile(published, endpoint, 0o600); err != nil {
		t.Fatalf("write endpoint: %v", err)
	}

	create, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "create_session",
			"arguments": map[string]any{
				"working_dir_path": h.workingDir,
				"tasks": []map[string]any{
					{"client_id": "only", "name": "read the readme", "prompt": "Read it."},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("encode create call: %v", err)
	}

	stdin := strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`,
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
		string(create),
	}, "\n") + "\n"

	run := exec.Command(shim)
	run.Stdin = strings.NewReader(stdin)

	out, err := run.Output()
	if err != nil {
		t.Fatalf("run shim: %v", err)
	}

	replies := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(replies) != 2 {
		t.Fatalf("shim wrote %d replies, want 2 (the notification must be silent):\n%s", len(replies), out)
	}

	listed := struct {
		Result struct {
			Tools []struct {
				Name string `json:"name"`
			} `json:"tools"`
		} `json:"result"`
	}{}

	if err := json.Unmarshal([]byte(replies[0]), &listed); err != nil {
		t.Fatalf("parse tools/list reply: %v", err)
	}

	if len(listed.Result.Tools) != 8 {
		t.Fatalf("shim carried back %d tools, want 8", len(listed.Result.Tools))
	}

	created := struct {
		Result struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
			IsError bool `json:"isError"`
		} `json:"result"`
	}{}

	if err := json.Unmarshal([]byte(replies[1]), &created); err != nil {
		t.Fatalf("parse create_session reply: %v", err)
	}

	if created.Result.IsError || len(created.Result.Content) == 0 {
		t.Fatalf("create_session through the shim failed: %s", replies[1])
	}

	if !strings.Contains(created.Result.Content[0].Text, `"session_id"`) {
		t.Fatalf("create_session through the shim returned no session: %s", created.Result.Content[0].Text)
	}
}

func TestTheStdioShimSaysTheAppIsClosedRatherThanHanging(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, "config"))

	shim := buildShim(t)

	run := exec.Command(shim)
	run.Stdin = strings.NewReader(`{"jsonrpc":"2.0","id":7,"method":"tools/list"}` + "\n")

	out, err := run.Output()
	if err != nil {
		t.Fatalf("run shim: %v", err)
	}

	failed := struct {
		ID    int `json:"id"`
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}{}

	if err := json.Unmarshal([]byte(strings.TrimSpace(string(out))), &failed); err != nil {
		t.Fatalf("parse failure reply: %v: %s", err, out)
	}

	if failed.ID != 7 {
		t.Fatalf("the failure reply lost the request id: %s", out)
	}

	if !strings.Contains(failed.Error.Message, "not running") {
		t.Fatalf("the failure reply does not say the app is closed: %s", failed.Error.Message)
	}
}
