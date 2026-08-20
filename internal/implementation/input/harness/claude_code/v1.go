package claude_code

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode"
	"unicode/utf8"

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

const harnessName = "claude-code"
const harnessLabel = "claude code"
const maxActivity = 40

var claudePlatforms = map[string]string{
	enums.Mac.String():     "darwin",
	enums.Linux.String():   "linux",
	enums.Windows.String(): "win32",
}

const (
	eventAssistant = "assistant"
	blockText      = "text"
	blockToolUse   = "tool_use"
)

const (
	toolRead      = "Read"
	toolEdit      = "Edit"
	toolWrite     = "Write"
	toolGlob      = "Glob"
	toolGrep      = "Grep"
	toolBash      = "Bash"
	toolWebFetch  = "WebFetch"
	toolWebSearch = "WebSearch"
	toolAskUser   = "AskUserQuestion"
)

var authURLRe = regexp.MustCompile(`https://[A-Za-z0-9._~:/?#&=%+-]+`)

// A bullet marker only counts with a space after it, so a diff body line such as
// "+func narration()" keeps its marker and gets skipped instead of reading as prose.
var markdownNoise = regexp.MustCompile("^(?:[>#]+\\s*|[-*+]\\s+|\\d+[.)]\\s+)")

var markdownEmphasis = strings.NewReplacer("**", "", "__", "", "`", "")

var baseTools = []string{
	toolRead, toolEdit, toolWrite, toolGlob, toolGrep, toolBash, toolWebFetch, toolWebSearch,
}

func allowedTools(gateway *core_itf.MCPGateway) string {
	if gateway == nil {
		return strings.Join(baseTools, ",")
	}

	tools := make([]string, 0, len(baseTools)+len(gateway.Servers))
	tools = append(tools, baseTools...)

	for _, server := range gateway.Servers {
		tools = append(tools, "mcp__"+server.Name)
	}

	return strings.Join(tools, ",")
}

type authSession struct {
	cmd    *exec.Cmd
	tty    *os.File
	mu     sync.Mutex
	out    []byte
	killed bool
	done   chan struct{}
}

func (s *authSession) read() {
	buf := make([]byte, 4096)
	for {
		n, err := s.tty.Read(buf)
		if n > 0 {
			s.mu.Lock()
			s.out = append(s.out, buf[:n]...)
			if len(s.out) > 128*1024 {
				s.out = append(s.out[:0:0], s.out[len(s.out)-64*1024:]...)
			}
			s.mu.Unlock()
		}
		if err != nil {
			break
		}
	}
	s.cmd.Wait()
	close(s.done)
}

func (s *authSession) snapshot() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return string(s.out)
}

func (s *authSession) wasKilled() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.killed
}

func (s *authSession) close() {
	select {
	case <-s.done:
	default:
		s.mu.Lock()
		s.killed = true
		s.mu.Unlock()
	}
	s.cmd.Process.Kill()
	s.tty.Close()
	select {
	case <-s.done:
	case <-time.After(5 * time.Second):
	}
}

type claudeManifest struct {
	Platforms map[string]struct {
		Binary   string `json:"binary"`
		Checksum string `json:"checksum"`
	} `json:"platforms"`
}

type agentProc struct {
	cmd         *exec.Cmd
	stdin       io.WriteCloser
	stdinMu     sync.Mutex
	stdout      io.ReadCloser
	out         chan string
	done        chan struct{}
	exited      chan struct{}
	lastOut     atomic.Int64
	ctxWindow   int
	usageMu     sync.Mutex
	usage       input_itf.ContextUsage
	meter       harness_helper.TurnMeter
	activityMu  sync.Mutex
	activity    []input_itf.Activity
	activitySeq int
}

type claudeBlock struct {
	Type  string `json:"type"`
	Text  string `json:"text"`
	Name  string `json:"name"`
	Input struct {
		FilePath string `json:"file_path"`
		Command  string `json:"command"`
		Pattern  string `json:"pattern"`
		URL      string `json:"url"`
	} `json:"input"`
}

type claudeEvent struct {
	Type            string `json:"type"`
	ParentToolUseID string `json:"parent_tool_use_id"`
	Message         struct {
		ID      string        `json:"id"`
		Content []claudeBlock `json:"content"`
		Usage   struct {
			InputTokens              int `json:"input_tokens"`
			OutputTokens             int `json:"output_tokens"`
			CacheReadInputTokens     int `json:"cache_read_input_tokens"`
			CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
		} `json:"usage"`
	} `json:"message"`
}

func (a *agentProc) track(line []byte) {
	event := &claudeEvent{}
	if err := json.Unmarshal(line, event); err != nil || event.Type != eventAssistant {
		return
	}

	a.trackUsage(event)
	a.trackActivity(event)
}

// A subagent spawned by the Task tool streams its own assistant events with
// parent_tool_use_id set. Their usage belongs to the subagent's context, not to
// this node's, so counting them makes the ring collapse and rebound mid-run.
func (a *agentProc) trackUsage(event *claudeEvent) {
	if event.ParentToolUseID != "" {
		return
	}

	reported := event.Message.Usage

	used := reported.InputTokens +
		reported.OutputTokens +
		reported.CacheReadInputTokens +
		reported.CacheCreationInputTokens

	if used == 0 {
		return
	}

	a.usageMu.Lock()
	total := a.meter.Observe(event.Message.ID, harness_helper.TokenCounts{
		Billed: reported.OutputTokens,
		Input:  reported.InputTokens + reported.CacheCreationInputTokens,
		Cached: reported.CacheReadInputTokens,
	})

	a.usage = input_itf.ContextUsage{
		Total:  a.ctxWindow,
		Used:   used,
		Billed: total.Billed,
		Input:  total.Input,
		Cached: total.Cached,
	}
	a.usageMu.Unlock()
}

func (a *agentProc) snapshotUsage() *input_itf.ContextUsage {
	a.usageMu.Lock()
	defer a.usageMu.Unlock()

	snapshot := a.usage
	snapshot.Total = a.ctxWindow

	return &snapshot
}

func (a *agentProc) trackActivity(event *claudeEvent) {
	for _, block := range event.Message.Content {
		text := digest(&block)
		if text == "" {
			continue
		}

		a.activityMu.Lock()
		a.activitySeq++
		a.activity = append(a.activity, input_itf.Activity{
			Seq:  a.activitySeq,
			At:   helpers.NewUTC(),
			Text: text,
		})
		if len(a.activity) > maxActivity {
			a.activity = append(a.activity[:0:0], a.activity[len(a.activity)-maxActivity:]...)
		}
		a.activityMu.Unlock()
	}
}

func (a *agentProc) snapshotActivity() []input_itf.Activity {
	a.activityMu.Lock()
	defer a.activityMu.Unlock()

	return slices.Clone(a.activity)
}

func digest(block *claudeBlock) string {
	switch block.Type {
	case blockText:
		return narration(block.Text)
	case blockToolUse:
		return toolPhrase(block)
	default:
		return ""
	}
}

func toolPhrase(block *claudeBlock) string {
	switch block.Name {
	case toolRead:
		return "Reading " + shortPath(block.Input.FilePath)
	case toolEdit, toolWrite:
		return "Editing " + shortPath(block.Input.FilePath)
	case toolBash:
		return "Running " + clip(firstSentence(block.Input.Command), 80)
	case toolGrep, toolGlob:
		return "Searching for " + clip(block.Input.Pattern, 80)
	case toolWebFetch:
		return "Fetching " + clip(block.Input.URL, 80)
	default:
		if block.Name == "" {
			return ""
		}
		return "Using " + block.Name
	}
}

// The activity feed shows one truncated line per message, so an agent that opens
// with a heading, a bullet or a code fence is unreadable there whatever the system
// prompt asked for. Take the first line that reads as a sentence and drop the rest.
func narration(text string) string {
	fenced := false

	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "```") {
			fenced = !fenced
			continue
		}

		if fenced {
			continue
		}

		trimmed = unmark(trimmed)
		if trimmed == "" {
			continue
		}

		if first, _ := utf8.DecodeRuneInString(trimmed); !unicode.IsLetter(first) {
			continue
		}

		return clip(firstSentence(trimmed), 140)
	}

	return ""
}

func unmark(line string) string {
	for {
		next := strings.TrimSpace(markdownNoise.ReplaceAllString(line, ""))
		if next == line {
			break
		}

		line = next
	}

	return strings.TrimSpace(markdownEmphasis.Replace(line))
}

func firstSentence(text string) string {
	line, _, _ := strings.Cut(strings.TrimSpace(text), "\n")
	for i := range len(line) {
		if line[i] != '.' {
			continue
		}
		if i+1 == len(line) || line[i+1] == ' ' {
			return line[:i+1]
		}
	}
	return line
}

func clip(text string, limit int) string {
	trimmed := strings.TrimSpace(text)
	runes := []rune(trimmed)
	if len(runes) <= limit {
		return trimmed
	}
	return strings.TrimSpace(string(runes[:limit])) + "…"
}

func shortPath(path string) string {
	cleaned := strings.Trim(filepath.ToSlash(path), "/")
	if cleaned == "" {
		return path
	}

	parts := strings.Split(cleaned, "/")
	if len(parts) <= 2 {
		return cleaned
	}

	return strings.Join(parts[len(parts)-2:], "/")
}

type claudeCodeCfg struct {
	Name          string            `mapstructure:"name" validate:"required"`
	BinName       string            `mapstructure:"bin_name" validate:"required"`
	ReleaseBase   string            `mapstructure:"release_base" validate:"required,http_url"`
	LoginTimeout  time.Duration     `mapstructure:"login_timeout" validate:"gt=0"`
	TokenRegex    string            `mapstructure:"token_regex" validate:"required"`
	AnsiRegex     string            `mapstructure:"ansi_regex" validate:"required"`
	MaxInstance   int               `mapstructure:"max_instance" validate:"required,gt=0"`
	EnabledModels []enums.ModelName `mapstructure:"enabled_models" validate:"required,dive,model_name"`
}

type claudeCode struct {
	dir           string
	binPath       string
	configDir     string
	tokenPath     string
	workspacesDir string

	agents    *harness_helper.Registry[*agentProc]
	authMu    sync.Mutex
	installMu sync.Mutex
	auth      *authSession
	loginMu   sync.Mutex
	cfg       *claudeCodeCfg
	httpCli   input_itf.HttpCli
	storage   input_itf.HarnessStorage
	tokenRe   *regexp.Regexp
	ansiRe    *regexp.Regexp
	spawnArgs []string
	mcpCfg    []byte
	baseEnv   []string
	ctxWindow int
}

func New(
	cfg input_itf.Config,
	httpCli input_itf.HttpCli,
	store input_itf.HarnessStorage,
	mcpGateway *core_itf.MCPGateway,
	raw map[string]any,
) (input_itf.AgentHarness, error) {
	claudeCfg, err := harness_helper.DecodeCfg[claudeCodeCfg](raw)
	if err != nil {
		return nil, err
	}

	base, err := os.UserConfigDir()
	if err != nil {
		return nil, custom_error.Critical("locate user config dir: %v", err)
	}

	tokenRe, err := regexp.Compile(claudeCfg.TokenRegex)
	if err != nil {
		return nil, custom_error.Critical("compile token_regex: %v", err)
	}

	ansiRe, err := regexp.Compile(claudeCfg.AnsiRegex)
	if err != nil {
		return nil, custom_error.Critical("compile ansi_regex: %v", err)
	}

	mcpCfg, err := claudeMCPConfig(mcpGateway)
	if err != nil {
		return nil, err
	}

	dir := filepath.Join(base, cfg.Read().App.DataDir, "harness", harnessName)
	configDir := filepath.Join(dir, "config")

	sandboxEnv, err := harness_helper.SandboxEnv(filepath.Join(base, cfg.Read().App.DataDir, "tools"))
	if err != nil {
		return nil, err
	}

	return &claudeCode{
		dir:           dir,
		binPath:       harness_helper.BinPath(dir, claudeCfg.BinName),
		configDir:     configDir,
		tokenPath:     filepath.Join(dir, "credentials"),
		workspacesDir: filepath.Join(dir, "workspaces"),
		agents:        harness_helper.NewRegistry[*agentProc](harnessLabel, claudeCfg.MaxInstance),
		cfg:           claudeCfg,
		httpCli:       httpCli,
		storage:       store,
		tokenRe:       tokenRe,
		ansiRe:        ansiRe,
		mcpCfg:        mcpCfg,
		baseEnv:       append(sandboxEnv, "CLAUDE_CONFIG_DIR="+configDir),
		ctxWindow:     cfg.Read().AgentManager.AllowedAgentContextWindow,
		spawnArgs: []string{"-p",
			"--input-format", "stream-json",
			"--output-format", "stream-json",
			"--verbose",
			"--permission-mode", "dontAsk",
			"--allowedTools", allowedTools(mcpGateway),
			"--disallowedTools", toolAskUser,
			"--append-system-prompt", string(prompts.System()),
		},
	}, nil
}

func (c *claudeCode) SupportedModels() []enums.ModelName {
	return slices.Clone(c.cfg.EnabledModels)
}

func (c *claudeCode) Support(name enums.ModelName) bool {
	return slices.Contains(c.cfg.EnabledModels, name)
}

func (c *claudeCode) Install(onProgress func(input_itf.InstallProgress)) error {
	c.installMu.Lock()
	defer c.installMu.Unlock()

	if err := harness_helper.Install(&harness_helper.InstallSpec{
		Name:    harnessName,
		Label:   harnessLabel,
		BinPath: c.binPath,
		Store:   c.storage,
		HttpCli: c.httpCli,
		Resolve: c.release,
	}, onProgress); err != nil {
		return err
	}

	c.agents.MarkInstalled()

	return nil
}

func (c *claudeCode) release() (*harness_helper.Release, error) {
	platform, err := harness_helper.Platform(claudePlatforms)
	if err != nil {
		return nil, err
	}

	version, err := c.httpCli.GetString(c.cfg.ReleaseBase + "/stable")
	if err != nil {
		return nil, custom_error.Critical("resolve stable version: %v", err)
	}

	manifest := &claudeManifest{}

	if err := c.httpCli.GetJSON(c.cfg.ReleaseBase+"/"+version+"/manifest.json", manifest); err != nil {
		return nil, custom_error.Critical("fetch manifest: %v", err)
	}

	entry, found := manifest.Platforms[platform]
	if !found {
		return nil, custom_error.Critical("no claude code build for platform %s", platform)
	}

	return &harness_helper.Release{
		Version:  version,
		Platform: platform,
		URL:      c.cfg.ReleaseBase + "/" + version + "/" + platform + "/" + entry.Binary,
		Checksum: entry.Checksum,
	}, nil
}

func (c *claudeCode) Auth() (string, error) {
	if !c.loginMu.TryLock() {
		return "", custom_error.Critical("auth is already in progress")
	}
	defer c.loginMu.Unlock()

	if _, err := os.Stat(c.tokenPath); err == nil {
		return "", nil
	}

	if _, err := os.Stat(c.binPath); err != nil {
		return "", custom_error.Critical("claude code is not installed, run Install first")
	}

	if err := os.MkdirAll(c.configDir, 0o755); err != nil {
		return "", custom_error.Critical("create login config dir: %v", err)
	}

	cmd := exec.Command(c.binPath, "setup-token")
	cmd.Env = append(slices.Clone(c.baseEnv), "TERM=xterm-256color")

	tty, err := harness_helper.StartPty(cmd, 500, 50)
	if err != nil {
		return "", custom_error.Critical("start login session: %v", err)
	}

	s := &authSession{cmd: cmd, tty: tty, done: make(chan struct{})}
	go s.read()

	c.authMu.Lock()
	old := c.auth
	c.auth = s
	c.authMu.Unlock()
	if old != nil {
		old.close()
	}

	go func() {
		select {
		case <-s.done:
		case <-time.After(c.cfg.LoginTimeout):
		}
		c.dropAuth(s)
	}()

	go func() {
		tok, err := c.waitFor(s, c.tokenRe, c.cfg.LoginTimeout)
		if err != nil {
			return
		}
		if err := c.storeToken(s, tok); err != nil {
			return
		}
		c.dropAuth(s)
	}()

	url, err := c.waitFor(s, authURLRe, 30*time.Second)
	if err != nil {
		c.dropAuth(s)
		return "", custom_error.Critical("wait for login url: %v", err)
	}

	return url, nil
}

func (c *claudeCode) SubmitAuthCode(code string) error {
	if _, err := os.Stat(c.tokenPath); err == nil {
		return nil
	}

	c.authMu.Lock()
	s := c.auth
	c.authMu.Unlock()
	if s == nil {
		return custom_error.Critical("no login in progress, run Auth first")
	}

	if _, err := s.tty.Write([]byte(strings.TrimSpace(code) + "\r")); err != nil {
		return custom_error.Critical("write auth code: %v", err)
	}

	tok, err := c.waitFor(s, c.tokenRe, 30*time.Second)
	if err != nil {
		return custom_error.Critical("confirm login: %v", err)
	}

	if err := c.storeToken(s, tok); err != nil {
		return err
	}

	c.dropAuth(s)

	return nil
}

func (c *claudeCode) storeToken(s *authSession, tok string) error {
	c.authMu.Lock()
	defer c.authMu.Unlock()

	if c.auth != s {
		return custom_error.Critical("login was cancelled")
	}

	if err := os.WriteFile(c.tokenPath, []byte(tok), 0o600); err != nil {
		return custom_error.Critical("store login token: %v", err)
	}

	return nil
}

func (c *claudeCode) Logout() error {
	c.authMu.Lock()
	s := c.auth
	c.auth = nil
	err := harness_helper.Logout(harnessLabel, c.tokenPath)
	c.authMu.Unlock()

	if s != nil {
		s.close()
	}

	c.stopAll()

	return err
}

func (c *claudeCode) dropAuth(s *authSession) {
	c.authMu.Lock()
	if c.auth == s {
		c.auth = nil
	}
	c.authMu.Unlock()
	s.close()
}

func (c *claudeCode) waitFor(s *authSession, re *regexp.Regexp, timeout time.Duration) (string, error) {
	match := func(settled bool) (string, bool) {
		clean := c.ansiRe.ReplaceAllString(strings.ReplaceAll(s.snapshot(), "\r", ""), "")

		loc := re.FindStringIndex(clean)
		if loc == nil || (!settled && loc[1] >= len(clean)) {
			return "", false
		}

		return clean[loc[0]:loc[1]], true
	}

	deadline := time.Now().Add(timeout)

	for {
		if found, ok := match(false); ok {
			return found, nil
		}

		select {
		case <-s.done:
			if s.wasKilled() {
				return "", custom_error.Critical("login session was cancelled")
			}

			if found, ok := match(true); ok {
				return found, nil
			}

			return "", custom_error.Critical("login process exited unexpectedly")
		default:
		}

		if time.Now().After(deadline) {
			return "", custom_error.Critical("timed out after %s", timeout)
		}

		time.Sleep(200 * time.Millisecond)
	}
}

func (c *claudeCode) Status() (*input_itf.AgentStatus, error) {
	return harness_helper.Status(harnessName, c.cfg.Name, c.tokenPath, c.storage, c.agents.Count())
}

func (c *claudeCode) Spawn(
	name enums.ModelName,
	thinkingLevel enums.ThinkingLevel,
	systemPrompts []string,
	workdir string,
) (uuid.UUID, error) {
	if !c.Support(name) {
		return uuid.Nil, custom_error.Critical("claude code does not support model %s", name)
	}

	if _, err := os.Stat(c.binPath); err != nil {
		return uuid.Nil, custom_error.Critical("claude code is not installed, run Install first")
	}

	token, err := os.ReadFile(c.tokenPath)
	if err != nil {
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

	args := append(slices.Clone(c.spawnArgs), "--model", string(name))

	for _, prompt := range systemPrompts {
		if trimmed := strings.TrimSpace(prompt); trimmed != "" {
			args = append(args, "--append-system-prompt", trimmed)
		}
	}

	if c.mcpCfg != nil {
		mcpPath := filepath.Join(workspace, "mcp.json")

		agentCfg := bytes.ReplaceAll(
			c.mcpCfg,
			[]byte(constances.GatewayAgentPlaceholder),
			[]byte(id),
		)

		if err := harness_helper.WriteNewFile(mcpPath, agentCfg, 0o600); err != nil {
			return uuid.Nil, custom_error.Critical("write mcp config: %v", err)
		}

		args = append(args, "--mcp-config", mcpPath, "--strict-mcp-config")
	}

	cmd := exec.Command(c.binPath, args...)
	cmd.Dir = workspace
	if workdir != "" {
		cmd.Dir = workdir
	}
	cmd.Env = append(slices.Clone(c.baseEnv),
		"CLAUDE_CODE_OAUTH_TOKEN="+strings.TrimSpace(string(token)),
	)

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

	stderr, err := os.Create(filepath.Join(workspace, "stderr.log"))
	if err != nil {
		return uuid.Nil, custom_error.Critical("open agent stderr log: %v", err)
	}

	unwind.Push(func() { stderr.Close() })

	cmd.Stderr = stderr
	harness_helper.SetProcAttrs(cmd)

	if err := cmd.Start(); err != nil {
		return uuid.Nil, custom_error.Critical("start claude code: %v", err)
	}

	unwind.Push(func() {
		harness_helper.KillProc(cmd)
		cmd.Wait()
	})

	out := make(chan string, 64)
	done := make(chan struct{})
	exited := make(chan struct{})

	proc := &agentProc{
		cmd:       cmd,
		stdin:     stdin,
		stdout:    stdout,
		out:       out,
		done:      done,
		exited:    exited,
		ctxWindow: c.ctxWindow,
	}

	proc.lastOut.Store(helpers.NewUTCUnix())

	if err := c.agents.Admit(id, proc); err != nil {
		return uuid.Nil, err
	}

	unwind.Done()

	go func() {
		sc := bufio.NewScanner(stdout)
		sc.Buffer(make([]byte, 64*1024), 8*1024*1024)
		for sc.Scan() {
			proc.lastOut.Store(helpers.NewUTCUnix())
			proc.track(sc.Bytes())

			select {
			case out <- sc.Text():
			case <-done:
			default:
			}
		}
		close(out)
		harness_helper.KillProc(cmd)
		cmd.Wait()
		stderr.Close()
		os.RemoveAll(workspace)
		c.agents.Forget(id)
		close(exited)
	}()

	return uid, nil
}

func (c *claudeCode) Send(id string, message string) error {
	a, err := c.agents.Get(id)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(map[string]any{
		"type": "user",
		"message": map[string]any{
			"role": "user",
			"content": []map[string]string{
				{"type": "text", "text": message},
			},
		},
	})
	if err != nil {
		return custom_error.Critical("encode message for agent %s: %v", id, err)
	}

	a.stdinMu.Lock()
	_, err = a.stdin.Write(append(payload, '\n'))
	a.stdinMu.Unlock()
	if err != nil {
		return custom_error.Critical("write to agent %s: %v", id, err)
	}

	return nil
}

func (c *claudeCode) Alive(id string) (time.Time, error) {
	a, err := c.agents.Get(id)
	if err != nil {
		return time.Time{}, err
	}

	select {
	case <-a.exited:
		return time.Time{}, custom_error.Critical("agent %s has exited", id)
	default:
	}

	return time.Unix(a.lastOut.Load(), 0).UTC(), nil
}

func (c *claudeCode) Usage(id string) (*input_itf.ContextUsage, error) {
	a, err := c.agents.Get(id)
	if err != nil {
		return nil, err
	}

	return a.snapshotUsage(), nil
}

func (c *claudeCode) Activity(id string) ([]input_itf.Activity, error) {
	a, err := c.agents.Get(id)
	if err != nil {
		return nil, err
	}

	return a.snapshotActivity(), nil
}

func (c *claudeCode) Listen(id string) (<-chan string, error) {
	a, err := c.agents.Get(id)
	if err != nil {
		return nil, err
	}

	return a.out, nil
}

func (c *claudeCode) stopAll() {
	c.authMu.Lock()
	s := c.auth
	c.auth = nil
	c.authMu.Unlock()
	if s != nil {
		s.close()
	}

	procs := c.agents.Drain()

	for _, a := range procs {
		a.stop()
	}

	for _, a := range procs {
		select {
		case <-a.exited:
		case <-time.After(5 * time.Second):
		}
	}
}

func (c *claudeCode) Shutdown() {
	c.stopAll()
}

func (c *claudeCode) Uninstall() error {
	c.agents.MarkUninstalled()

	c.stopAll()

	if err := os.RemoveAll(c.dir); err != nil {
		return custom_error.Critical("remove install dir: %v", err)
	}

	return nil
}

func (c *claudeCode) Kill(id string) error {
	a, err := c.agents.Take(id)
	if err != nil {
		return err
	}

	return a.stop()
}

func (a *agentProc) stop() error {
	close(a.done)
	a.stdin.Close()

	return harness_helper.SignalProc(a.cmd)
}

func claudeMCPConfig(gateway *core_itf.MCPGateway) ([]byte, error) {
	if gateway == nil || len(gateway.Servers) == 0 {
		return nil, nil
	}

	servers := map[string]any{}

	for _, server := range gateway.Servers {
		servers[server.Name] = map[string]any{
			"type":    "http",
			"url":     harness_helper.GatewayURL(gateway, server),
			"headers": harness_helper.GatewayHeaders(gateway, server),
		}
	}

	raw, err := json.Marshal(map[string]any{"mcpServers": servers})
	if err != nil {
		return nil, custom_error.Critical("cannot build claude code mcp config: %v", err)
	}

	return raw, nil
}
