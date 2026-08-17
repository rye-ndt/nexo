package open_code

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
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

const openCodeName = "open-code"
const openCodeLabel = "open code"

const (
	permissionAllow = "allow"
	sseDataPrefix   = "data: "
)

var openCodePlatforms = map[string]string{
	enums.Mac.String():     enums.Mac.String(),
	enums.Linux.String():   enums.Linux.String(),
	enums.Windows.String(): enums.Windows.String(),
}

var openCodePermissions = map[string]string{
	"edit":     permissionAllow,
	"bash":     permissionAllow,
	"webfetch": permissionAllow,
}

type githubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Digest             string `json:"digest"`
	} `json:"assets"`
}

type openCodeProc struct {
	cmd       *exec.Cmd
	out       chan string
	port      int
	session   string
	done      chan struct{}
	exited    chan struct{}
	lastOut   atomic.Int64
	ctxWindow int
	usageMu   sync.Mutex
	usage     input_itf.ContextUsage
	billed    int
	msgID     string
	msgUsed   int
}

type openCodeUsage struct {
	Properties struct {
		Info struct {
			ID        string `json:"id"`
			SessionID string `json:"sessionID"`
			Tokens    struct {
				Input     int `json:"input"`
				Output    int `json:"output"`
				Reasoning int `json:"reasoning"`
				Cache     struct {
					Read  int `json:"read"`
					Write int `json:"write"`
				} `json:"cache"`
			} `json:"tokens"`
		} `json:"info"`
	} `json:"properties"`
}

func (p *openCodeProc) trackUsage(line []byte) {
	event := &openCodeUsage{}
	if err := json.Unmarshal(line, event); err != nil {
		return
	}

	// The /event stream carries every session on this port, child sessions of the
	// Task tool included. Only this proc's own session is this node's context.
	if id := event.Properties.Info.SessionID; id != "" && id != p.session {
		return
	}

	reported := event.Properties.Info.Tokens

	used := reported.Input +
		reported.Output +
		reported.Reasoning +
		reported.Cache.Read +
		reported.Cache.Write

	if used == 0 {
		return
	}

	p.usageMu.Lock()
	// The stream re-reports a message as it grows, so its tokens only join the running
	// total once the next message starts. An event with no id cannot be matched that
	// way, so it counts as a turn of its own — and leaves the named message in flight
	// alone, which would otherwise be billed again when it streams on.
	if id := event.Properties.Info.ID; id == "" {
		p.billed += used
	} else {
		if id != p.msgID {
			p.billed += p.msgUsed
			p.msgID = id
			p.msgUsed = 0
		}

		if used > p.msgUsed {
			p.msgUsed = used
		}
	}

	p.usage = input_itf.ContextUsage{
		Total:  p.ctxWindow,
		Used:   used,
		Billed: p.billed + p.msgUsed,
	}
	p.usageMu.Unlock()
}

func (p *openCodeProc) snapshotUsage() *input_itf.ContextUsage {
	p.usageMu.Lock()
	defer p.usageMu.Unlock()

	snapshot := p.usage
	snapshot.Total = p.ctxWindow

	return &snapshot
}

type openCodeCfg struct {
	Name          string            `mapstructure:"name" validate:"required"`
	BinName       string            `mapstructure:"bin_name" validate:"required"`
	ReleaseBase   string            `mapstructure:"release_base" validate:"required,http_url"`
	LoginTimeout  time.Duration     `mapstructure:"login_timeout" validate:"gt=0"`
	MaxInstance   int               `mapstructure:"max_instance" validate:"required,gt=0"`
	EnabledModels []enums.ModelName `mapstructure:"enabled_models" validate:"required,dive,model_name"`
}

type openCode struct {
	dir           string
	binPath       string
	dataDir       string
	authPath      string
	workspacesDir string

	agents       *harness_helper.Registry[*openCodeProc]
	authMu       sync.Mutex
	installMu    sync.Mutex
	cfg          *openCodeCfg
	httpCli      input_itf.HttpCli
	storage      input_itf.HarnessStorage
	agentCfg     map[string]any
	systemPrompt []byte
	env          []string
	sessionCli   *http.Client
	ctxWindow    int
}

func New(
	cfg input_itf.Config,
	httpCli input_itf.HttpCli,
	store input_itf.HarnessStorage,
	mcpGateway *core_itf.MCPGateway,
	raw map[string]any,
) (input_itf.AgentHarness, error) {
	openCfg, err := harness_helper.DecodeCfg[openCodeCfg](raw)
	if err != nil {
		return nil, err
	}

	base, err := os.UserConfigDir()
	if err != nil {
		return nil, custom_error.Critical("locate user config dir: %v", err)
	}

	dir := filepath.Join(base, cfg.Read().App.DataDir, "harness", openCodeName)
	dataDir := filepath.Join(dir, "config")

	sandboxEnv, err := harness_helper.SandboxEnv(
		filepath.Join(base, cfg.Read().App.DataDir, "tools"),
		"XDG_DATA_HOME=",
		"OPENCODE_",
	)
	if err != nil {
		return nil, err
	}

	agentCfg := map[string]any{
		"permission": openCodePermissions,
	}

	if block := openCodeMCPCfg(mcpGateway); block != nil {
		agentCfg["mcp"] = block
	}

	return &openCode{
		dir:           dir,
		binPath:       harness_helper.BinPath(dir, openCfg.BinName),
		dataDir:       dataDir,
		authPath:      filepath.Join(dataDir, "opencode", "auth.json"),
		workspacesDir: filepath.Join(dir, "workspaces"),

		agents:       harness_helper.NewRegistry[*openCodeProc](openCodeLabel, openCfg.MaxInstance),
		cfg:          openCfg,
		httpCli:      httpCli,
		storage:      store,
		agentCfg:     agentCfg,
		systemPrompt: prompts.System(),
		env:          append(sandboxEnv, "XDG_DATA_HOME="+dataDir),
		sessionCli:   &http.Client{Timeout: 2 * time.Second},
		ctxWindow:    cfg.Read().AgentManager.AllowedAgentContextWindow,
	}, nil
}

func (o *openCode) SupportedModels() []enums.ModelName {
	return slices.Clone(o.cfg.EnabledModels)
}

func (o *openCode) Support(name enums.ModelName) bool {
	return slices.Contains(o.cfg.EnabledModels, name)
}

func (o *openCode) Install(onProgress func(input_itf.InstallProgress)) error {
	o.installMu.Lock()
	defer o.installMu.Unlock()

	if err := harness_helper.Install(&harness_helper.InstallSpec{
		Name:    openCodeName,
		Label:   openCodeLabel,
		BinPath: o.binPath,
		Store:   o.storage,
		HttpCli: o.httpCli,
		Resolve: o.release,
		Place: func(downloaded, dest string) error {
			return extractBinary(downloaded, filepath.Base(dest), dest)
		},
	}, onProgress); err != nil {
		return err
	}

	o.agents.MarkInstalled()

	return nil
}

func (o *openCode) release() (*harness_helper.Release, error) {
	platform, err := harness_helper.Platform(openCodePlatforms)
	if err != nil {
		return nil, err
	}

	latest := &githubRelease{}

	if err := o.httpCli.GetJSON(o.cfg.ReleaseBase+"/releases/latest", latest); err != nil {
		return nil, custom_error.Critical("resolve latest release: %v", err)
	}

	want := o.cfg.BinName + "-" + platform + ".zip"

	for _, asset := range latest.Assets {
		if asset.Name != want {
			continue
		}

		return &harness_helper.Release{
			Version:  strings.TrimPrefix(latest.TagName, "v"),
			Platform: platform,
			URL:      asset.BrowserDownloadURL,
			Checksum: strings.TrimPrefix(asset.Digest, "sha256:"),
		}, nil
	}

	return nil, custom_error.Critical("no open code build for platform %s", platform)
}

func (o *openCode) Auth() (string, error) {
	if !o.authMu.TryLock() {
		return "", custom_error.Critical("auth is already in progress")
	}
	defer o.authMu.Unlock()

	if _, err := os.Stat(o.authPath); err == nil {
		return "", nil
	}

	if _, err := os.Stat(o.binPath); err != nil {
		return "", custom_error.Critical("open code is not installed, run Install first")
	}

	if runtime.GOOS != enums.Mac.String() {
		return "", custom_error.Critical("interactive login is only implemented for macOS")
	}

	if err := os.MkdirAll(o.dataDir, 0o755); err != nil {
		return "", custom_error.Critical("create login data dir: %v", err)
	}

	scriptPath := filepath.Join(o.dir, "login.sh")

	sh := fmt.Sprintf("#!/bin/sh\nexport XDG_DATA_HOME='%s'\nexec '%s' auth login\n",
		o.dataDir, o.binPath)
	if err := os.WriteFile(scriptPath, []byte(sh), 0o700); err != nil {
		return "", custom_error.Critical("write login script: %v", err)
	}

	if err := exec.Command("osascript",
		"-e", `tell application "Terminal" to activate`,
		"-e", fmt.Sprintf(`tell application "Terminal" to do script "sh '%s'"`, scriptPath),
	).Run(); err != nil {
		return "", custom_error.Critical("open Terminal for login: %v", err)
	}

	deadline := time.Now().Add(o.cfg.LoginTimeout)
	for time.Now().Before(deadline) {
		time.Sleep(2 * time.Second)
		if _, err := os.Stat(o.authPath); err == nil {
			return "", nil
		}
	}

	return "", custom_error.Critical("login timed out after %s", o.cfg.LoginTimeout)
}

func (o *openCode) SubmitAuthCode(code string) error {
	return custom_error.Critical("open code login does not use an auth code")
}

func (o *openCode) Status() (*input_itf.AgentStatus, error) {
	return harness_helper.Status(openCodeName, o.cfg.Name, o.authPath, o.storage, o.agents.Count())
}

func (o *openCode) Spawn(
	name enums.ModelName,
	thinkingLevel enums.ThinkingLevel,
	systemPrompts []string,
	workdir string,
) (uuid.UUID, error) {
	if workdir != "" {
		return uuid.Nil, custom_error.Critical("open code does not support a custom working dir yet")
	}

	if !o.Support(name) {
		return uuid.Nil, custom_error.Critical("open code does not support model %s", name)
	}

	if _, err := os.Stat(o.binPath); err != nil {
		return uuid.Nil, custom_error.Critical("open code is not installed, run Install first")
	}

	if _, err := os.Stat(o.authPath); err != nil {
		return uuid.Nil, custom_error.Critical("not authenticated, run Auth first")
	}

	if err := o.agents.Reserve(); err != nil {
		return uuid.Nil, err
	}

	uid, err := uuid.NewV7()
	if err != nil {
		return uuid.Nil, err
	}

	id := uid.String()

	unwind := &harness_helper.Unwind{}
	defer unwind.Run()

	workspace := filepath.Join(o.workspacesDir, id)
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		return uuid.Nil, custom_error.Critical("create agent workspace: %v", err)
	}

	unwind.Push(func() { os.RemoveAll(workspace) })

	agentCfg := maps.Clone(o.agentCfg)
	agentCfg["model"] = string(name)

	rawCfg, err := json.Marshal(agentCfg)
	if err != nil {
		return uuid.Nil, custom_error.Critical("build open code config: %v", err)
	}

	rawCfg = bytes.ReplaceAll(
		rawCfg,
		[]byte(constances.GatewayAgentPlaceholder),
		[]byte(id),
	)

	if err := harness_helper.WriteNewFile(filepath.Join(workspace, "opencode.json"), rawCfg, 0o600); err != nil {
		return uuid.Nil, custom_error.Critical("write opencode config: %v", err)
	}

	if err := os.WriteFile(filepath.Join(workspace, "AGENTS.md"), agentsFile(o.systemPrompt, systemPrompts), 0o644); err != nil {
		return uuid.Nil, custom_error.Critical("write agent prompt file: %v", err)
	}

	port, err := freePort(constances.GlobalLocalHost)
	if err != nil {
		return uuid.Nil, err
	}

	logFile, err := os.Create(filepath.Join(workspace, "serve.log"))
	if err != nil {
		return uuid.Nil, custom_error.Critical("open agent serve log: %v", err)
	}

	unwind.Push(func() { logFile.Close() })

	cmd := exec.Command(
		o.binPath,
		"serve",
		"--port", strconv.Itoa(port),
		"--hostname", constances.GlobalLocalHost,
	)

	cmd.Dir = workspace
	cmd.Env = o.env
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	harness_helper.SetProcAttrs(cmd)

	if err := cmd.Start(); err != nil {
		return uuid.Nil, custom_error.Critical("start open code: %v", err)
	}

	unwind.Push(func() {
		harness_helper.KillProc(cmd)
		cmd.Wait()
	})

	session, err := o.createSession(port)
	if err != nil {
		return uuid.Nil, err
	}

	out := make(chan string, 64)
	done := make(chan struct{})
	exited := make(chan struct{})

	proc := &openCodeProc{cmd: cmd, out: out, port: port, session: session, done: done, exited: exited, ctxWindow: o.ctxWindow}
	proc.lastOut.Store(helpers.NewUTCUnix())

	if err := o.agents.Admit(id, proc); err != nil {
		return uuid.Nil, err
	}

	unwind.Done()

	streamClosed := make(chan struct{})

	go streamEvents(o.baseURL(port), out, done, streamClosed, proc)

	go func() {
		select {
		case <-done:
		case <-streamClosed:
		}
		harness_helper.KillProc(cmd)
		cmd.Wait()
		logFile.Close()
		os.RemoveAll(workspace)
		o.agents.Forget(id)
		close(exited)
	}()

	return uid, nil
}

func (o *openCode) baseURL(port int) string {
	return "http://" + net.JoinHostPort(constances.GlobalLocalHost, strconv.Itoa(port))
}

func (o *openCode) createSession(port int) (string, error) {
	url := o.baseURL(port) + "/session"

	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		res, err := o.sessionCli.Post(url, "application/json", strings.NewReader("{}"))
		if err != nil {
			time.Sleep(200 * time.Millisecond)
			continue
		}
		if res.StatusCode != http.StatusOK {
			res.Body.Close()
			time.Sleep(200 * time.Millisecond)
			continue
		}

		var session struct {
			ID string `json:"id"`
		}
		err = json.NewDecoder(res.Body).Decode(&session)
		res.Body.Close()
		if err != nil {
			return "", custom_error.Critical("decode session: %v", err)
		}
		return session.ID, nil
	}

	return "", custom_error.Critical("open code server did not become ready on port %d", port)
}

func streamEvents(
	baseURL string,
	out chan string,
	done <-chan struct{},
	closed chan<- struct{},
	proc *openCodeProc,
) {
	defer close(closed)
	defer close(out)

	res, err := http.Get(baseURL + "/event")
	if err != nil {
		return
	}
	defer res.Body.Close()

	sc := bufio.NewScanner(res.Body)
	sc.Buffer(make([]byte, 64*1024), 8*1024*1024)
	for sc.Scan() {
		proc.lastOut.Store(helpers.NewUTCUnix())

		data, ok := strings.CutPrefix(sc.Text(), sseDataPrefix)
		if !ok || data == "" {
			continue
		}

		proc.trackUsage([]byte(data))

		select {
		case out <- data:
		case <-done:
			return
		default:
		}
	}
}

func (o *openCode) Send(id string, message string) error {
	a, err := o.agents.Get(id)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(map[string]any{
		"parts": []map[string]string{
			{"type": "text", "text": message},
		},
	})
	if err != nil {
		return custom_error.Critical("encode message for agent %s: %v", id, err)
	}

	url := fmt.Sprintf("%s/session/%s/message", o.baseURL(a.port), a.session)
	res, err := http.Post(url, "application/json", bytes.NewReader(payload))
	if err != nil {
		return custom_error.Critical("send to agent %s: %v", id, err)
	}
	defer res.Body.Close()
	io.Copy(io.Discard, res.Body)

	if res.StatusCode != http.StatusOK {
		return custom_error.Critical("send to agent %s: %s", id, res.Status)
	}
	return nil
}

func (o *openCode) Alive(id string) (time.Time, error) {
	a, err := o.agents.Get(id)
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

func (o *openCode) Usage(id string) (*input_itf.ContextUsage, error) {
	a, err := o.agents.Get(id)
	if err != nil {
		return nil, err
	}

	return a.snapshotUsage(), nil
}

func (o *openCode) Activity(id string) ([]input_itf.Activity, error) {
	return []input_itf.Activity{}, nil
}

func (o *openCode) Listen(id string) (<-chan string, error) {
	a, err := o.agents.Get(id)
	if err != nil {
		return nil, err
	}

	return a.out, nil
}

func (o *openCode) stopAll() {
	procs := o.agents.Drain()

	for _, a := range procs {
		close(a.done)
	}

	for _, a := range procs {
		select {
		case <-a.exited:
		case <-time.After(5 * time.Second):
		}
	}
}

func (o *openCode) Shutdown() {
	o.stopAll()
}

func (o *openCode) Uninstall() error {
	o.agents.MarkUninstalled()

	o.stopAll()

	if err := os.RemoveAll(o.dir); err != nil {
		return custom_error.Critical("remove install dir: %v", err)
	}

	return nil
}

func (o *openCode) Kill(id string) error {
	a, err := o.agents.Take(id)
	if err != nil {
		return err
	}

	close(a.done)

	return nil
}

func agentsFile(base []byte, extra []string) []byte {
	parts := []string{strings.TrimSpace(string(base))}

	for _, prompt := range extra {
		if trimmed := strings.TrimSpace(prompt); trimmed != "" {
			parts = append(parts, trimmed)
		}
	}

	return []byte(strings.Join(slices.DeleteFunc(parts, func(s string) bool { return s == "" }), "\n\n"))
}

func freePort(host string) (int, error) {
	ln, err := net.Listen("tcp", net.JoinHostPort(host, "0"))
	if err != nil {
		return 0, custom_error.Critical("reserve a local port: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()
	return port, nil
}

func openCodeMCPCfg(gateway *core_itf.MCPGateway) map[string]any {
	if gateway == nil || len(gateway.Servers) == 0 {
		return nil
	}

	block := map[string]any{}

	for _, server := range gateway.Servers {
		block[server.Name] = map[string]any{
			"type":    "remote",
			"url":     harness_helper.GatewayURL(gateway, server),
			"enabled": true,
			"headers": harness_helper.GatewayHeaders(gateway, server),
		}
	}

	return block
}

func extractBinary(archive, member, dest string) error {
	r, err := zip.OpenReader(archive)
	if err != nil {
		return custom_error.Critical("open archive: %v", err)
	}
	defer r.Close()

	for _, f := range r.File {
		if filepath.Base(f.Name) == member {
			return writeZipEntry(f, dest)
		}
	}

	return custom_error.Critical("%s not found in archive", member)
}

func writeZipEntry(f *zip.File, dest string) error {
	rc, err := f.Open()
	if err != nil {
		return custom_error.Critical("read %s from archive: %v", f.Name, err)
	}
	defer rc.Close()

	tmp := dest + ".tmp"

	out, err := os.Create(tmp)
	if err != nil {
		return custom_error.Critical("create %s: %v", tmp, err)
	}

	_, err = io.Copy(out, rc)

	if closeErr := out.Close(); err == nil {
		err = closeErr
	}

	if err == nil {
		err = os.Rename(tmp, dest)
	}

	if err != nil {
		os.Remove(tmp)

		return custom_error.Critical("extract %s: %v", f.Name, err)
	}

	return nil
}
