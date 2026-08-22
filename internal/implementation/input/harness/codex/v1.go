package codex

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"hexago/internal/helpers"
	"hexago/internal/helpers/constances"
	"hexago/internal/helpers/custom_error"
	"hexago/internal/helpers/enums"
	"hexago/internal/helpers/prompts"
	"hexago/internal/implementation/input/harness/harness_helper"
	core_itf "hexago/internal/interface/core"
	input_itf "hexago/internal/interface/input"
)

const codexName = "codex"
const rpcTimeout = 30 * time.Second

type codexCfg struct {
	Name          string            `mapstructure:"name" validate:"required"`
	BinName       string            `mapstructure:"bin_name" validate:"required"`
	ReleaseBase   string            `mapstructure:"release_base" validate:"required,http_url"`
	LoginTimeout  time.Duration     `mapstructure:"login_timeout" validate:"gt=0"`
	MaxInstance   int               `mapstructure:"max_instance" validate:"required,gt=0"`
	EnabledModels []enums.ModelName `mapstructure:"enabled_models" validate:"required,dive,model_name"`
}

type codex struct {
	dir           string
	binPath       string
	configDir     string
	authPath      string
	workspacesDir string

	agents       *harness_helper.Registry[*agentProc]
	authMu       sync.Mutex
	installMu    sync.Mutex
	cfg          *codexCfg
	httpCli      input_itf.HttpCli
	storage      input_itf.HarnessStorage
	mcpGateway   *core_itf.MCPGateway
	systemPrompt []byte
	baseEnv      []string
	ctxWindow    int
}

func New(
	cfg input_itf.Config,
	httpCli input_itf.HttpCli,
	store input_itf.HarnessStorage,
	mcpGateway *core_itf.MCPGateway,
	raw map[string]any,
) (input_itf.AgentHarness, error) {
	codexCfg, err := harness_helper.DecodeCfg[codexCfg](raw)
	if err != nil {
		return nil, err
	}

	base, err := os.UserConfigDir()
	if err != nil {
		return nil, custom_error.Critical("locate user config dir: %v", err)
	}

	dir := filepath.Join(base, cfg.Read().App.DataDir, "harness", codexName)
	configDir := filepath.Join(dir, "config")

	sandboxEnv, err := harness_helper.SandboxEnv(
		filepath.Join(base, cfg.Read().App.DataDir, "tools"),
		"CODEX_HOME=",
		"OPENAI_API_KEY=",
		"CODEX_API_KEY=",
	)
	if err != nil {
		return nil, err
	}

	return &codex{
		dir:           dir,
		binPath:       harness_helper.BinPath(dir, codexCfg.BinName),
		configDir:     configDir,
		authPath:      filepath.Join(configDir, "auth.json"),
		workspacesDir: filepath.Join(dir, "workspaces"),
		agents:        harness_helper.NewRegistry[*agentProc](codexName, codexCfg.MaxInstance),
		cfg:           codexCfg,
		httpCli:       httpCli,
		storage:       store,
		mcpGateway:    mcpGateway,
		systemPrompt:  prompts.System(),
		baseEnv:       append(sandboxEnv, "CODEX_HOME="+configDir),
		ctxWindow:     cfg.Read().AgentManager.AllowedAgentContextWindow,
	}, nil
}

func (c *codex) SupportedModels() []enums.ModelName {
	return slices.Clone(c.cfg.EnabledModels)
}

func (c *codex) Support(name enums.ModelName) bool {
	return slices.Contains(c.cfg.EnabledModels, name)
}

func (c *codex) Install(onProgress func(input_itf.InstallProgress)) error {
	c.installMu.Lock()
	defer c.installMu.Unlock()

	if err := harness_helper.Install(&harness_helper.InstallSpec{
		Name:    codexName,
		Label:   codexName,
		BinPath: c.binPath,
		Store:   c.storage,
		HttpCli: c.httpCli,
		Resolve: c.release,
		Place:   extractBinary,
	}, onProgress); err != nil {
		return err
	}

	c.agents.MarkInstalled()

	return nil
}

func (c *codex) Auth() (string, error) {
	if !c.authMu.TryLock() {
		return "", custom_error.Critical("auth is already in progress")
	}
	defer c.authMu.Unlock()

	if _, err := os.Stat(c.binPath); err != nil {
		return "", custom_error.Critical("codex is not installed, run Install first")
	}

	if c.loggedIn() {
		return "", nil
	}

	if err := os.MkdirAll(c.configDir, 0o755); err != nil {
		return "", custom_error.Critical("create login config dir: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.cfg.LoginTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, c.binPath, "login")
	cmd.Env = slices.Clone(c.baseEnv)
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	harness_helper.SetProcAttrs(cmd)

	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return "", custom_error.Critical("login timed out after %s", c.cfg.LoginTimeout)
		}
		return "", custom_error.Critical("codex login failed: %v", err)
	}

	if !c.loggedIn() {
		return "", custom_error.Critical("codex login completed without an authenticated account")
	}

	return "", nil
}

func (c *codex) SubmitAuthCode(code string) error {
	return custom_error.Critical("codex login does not use an auth code")
}

func (c *codex) Logout() error {
	return harness_helper.Logout(codexName, c.authPath)
}

func (c *codex) Status() (*input_itf.AgentStatus, error) {
	status := &input_itf.AgentStatus{Name: c.cfg.Name, InstanceCount: c.agents.Count()}

	info, err := c.storage.Find(codexName)
	if err != nil {
		return nil, custom_error.Critical("find harness info: %v", err)
	}

	if info != nil {
		if _, err := os.Stat(info.Path); err == nil {
			status.Installed = true
			status.Version = info.Version
			status.LoggedIn = c.loggedIn()
		}
	}

	return status, nil
}

func (c *codex) loggedIn() bool {
	_, err := os.Stat(c.authPath)

	return err == nil
}

func (c *codex) Spawn(
	name enums.ModelName,
	thinkingLevel enums.ThinkingLevel,
	systemPrompts []string,
	workdir string,
) (uuid.UUID, error) {
	if !c.Support(name) {
		return uuid.Nil, custom_error.Critical("codex does not support model %s", name)
	}

	if _, err := os.Stat(c.binPath); err != nil {
		return uuid.Nil, custom_error.Critical("codex is not installed, run Install first")
	}

	if !c.loggedIn() {
		return uuid.Nil, custom_error.Critical("not authenticated, run Auth first")
	}

	if err := c.agents.Reserve(); err != nil {
		return uuid.Nil, err
	}

	uid, err := uuid.NewV7()
	if err != nil {
		return uuid.Nil, err
	}

	id := uid.String()
	unwind := &harness_helper.Unwind{}
	defer unwind.Run()

	workspace := filepath.Join(c.workspacesDir, id)
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		return uuid.Nil, custom_error.Critical("create agent workspace: %v", err)
	}
	unwind.Push(func() { os.RemoveAll(workspace) })

	stderr, err := os.Create(filepath.Join(workspace, "stderr.log"))
	if err != nil {
		return uuid.Nil, custom_error.Critical("open agent stderr log: %v", err)
	}
	unwind.Push(func() { stderr.Close() })

	cmd := exec.Command(c.binPath, "app-server", "--stdio")
	cmd.Dir = workspace
	if workdir != "" {
		cmd.Dir = workdir
	}
	cmd.Env = slices.Clone(c.baseEnv)
	cmd.Stderr = stderr
	harness_helper.SetProcAttrs(cmd)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return uuid.Nil, custom_error.Critical("open agent stdin: %v", err)
	}
	unwind.Push(func() { stdin.Close() })

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return uuid.Nil, custom_error.Critical("open agent stdout: %v", err)
	}
	unwind.Push(func() { stdout.Close() })

	if err := cmd.Start(); err != nil {
		return uuid.Nil, custom_error.Critical("start codex: %v", err)
	}
	unwind.Push(func() {
		harness_helper.KillProc(cmd)
		cmd.Wait()
	})

	proc := newAgentProc(cmd, stdin, stdout, c.ctxWindow)
	proc.lastOut.Store(helpers.NewUTCUnix())
	proc.effort = thinkingLevel.String()

	if err := c.agents.Admit(id, proc); err != nil {
		return uuid.Nil, err
	}

	unwind.Done()
	go c.run(id, workspace, stderr, proc)

	if err := proc.initialize(); err != nil {
		c.abortSpawn(id, proc)
		return uuid.Nil, custom_error.Critical("initialize codex agent: %v", err)
	}

	cwd := cmd.Dir
	params := map[string]any{
		"model":                 string(name),
		"cwd":                   cwd,
		"approvalPolicy":        "never",
		"sandbox":               "workspace-write",
		"ephemeral":             true,
		"developerInstructions": harness_helper.PromptText(c.systemPrompt, systemPrompts),
	}

	if mcp := codexMCPConfig(c.mcpGateway, id); mcp != nil {
		params["config"] = map[string]any{"mcp_servers": mcp}
	}

	if _, err := proc.startThread(params); err != nil {
		c.abortSpawn(id, proc)
		return uuid.Nil, custom_error.Critical("start codex thread: %v", err)
	}

	return uid, nil
}

func (c *codex) abortSpawn(id string, proc *agentProc) {
	c.agents.Take(id)
	proc.stop()
	select {
	case <-proc.exited:
	case <-time.After(5 * time.Second):
	}
}

func (c *codex) Send(id string, message string) error {
	proc, err := c.agents.Get(id)
	if err != nil {
		return err
	}

	if err := proc.send(message); err != nil {
		return custom_error.Critical("send to agent %s: %v", id, err)
	}

	return nil
}

func (c *codex) Alive(id string) (time.Time, error) {
	proc, err := c.agents.Get(id)
	if err != nil {
		return time.Time{}, err
	}

	select {
	case <-proc.exited:
		return time.Time{}, custom_error.Critical("agent %s has exited", id)
	default:
	}

	return time.Unix(proc.lastOut.Load(), 0).UTC(), nil
}

func (c *codex) Usage(id string) (*input_itf.ContextUsage, error) {
	proc, err := c.agents.Get(id)
	if err != nil {
		return nil, err
	}

	return proc.snapshotUsage(), nil
}

func (c *codex) Activity(id string) ([]input_itf.Activity, error) {
	proc, err := c.agents.Get(id)
	if err != nil {
		return nil, err
	}

	return proc.snapshotActivity(), nil
}

func (c *codex) Kill(id string) error {
	proc, err := c.agents.Take(id)
	if err != nil {
		return err
	}

	return proc.stop()
}

func (c *codex) stopAll() {
	procs := c.agents.Drain()

	for _, proc := range procs {
		proc.stop()
	}

	for _, proc := range procs {
		select {
		case <-proc.exited:
		case <-time.After(5 * time.Second):
		}
	}
}

func (c *codex) Shutdown() {
	c.stopAll()
}

func (c *codex) Uninstall() error {
	c.agents.MarkUninstalled()
	c.stopAll()

	if err := os.RemoveAll(c.dir); err != nil {
		return custom_error.Critical("remove install dir: %v", err)
	}

	return nil
}

func (c *codex) run(id, workspace string, stderr *os.File, proc *agentProc) {
	proc.read()
	harness_helper.KillProc(proc.cmd)
	proc.cmd.Wait()
	proc.stdin.Close()
	proc.stdout.Close()
	stderr.Close()
	os.RemoveAll(workspace)
	c.agents.Forget(id)
	close(proc.exited)
}

func codexMCPConfig(gateway *core_itf.MCPGateway, agentID string) map[string]any {
	if gateway == nil || len(gateway.Servers) == 0 {
		return nil
	}

	servers := map[string]any{}

	for _, server := range gateway.Servers {
		headers := harness_helper.GatewayHeaders(gateway, server)
		for name, value := range headers {
			headers[name] = strings.ReplaceAll(value, constances.GatewayAgentPlaceholder, agentID)
		}

		servers[server.Name] = map[string]any{
			"url":          harness_helper.GatewayURL(gateway, server),
			"http_headers": headers,
			"required":     true,
		}
	}

	return servers
}
