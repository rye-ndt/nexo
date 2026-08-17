package mcp_proxy_helpers

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"hexago/internal/helpers/constances"
	"hexago/internal/helpers/custom_error"
	"hexago/internal/helpers/enums"
	input_itf "hexago/internal/interface/input"
)

const (
	chromeProbeInterval = 250 * time.Millisecond
	chromeProfileMode   = 0o700
	chromeSeedFileMode  = 0o600
	chromePortFile      = "NexoDevToolsEndpoint"
)

var chromeSkippedDirs = map[string]any{
	"Cache":             struct{}{},
	"Code Cache":        struct{}{},
	"GPUCache":          struct{}{},
	"ShaderCache":       struct{}{},
	"GrShaderCache":     struct{}{},
	"DawnCache":         struct{}{},
	"DawnGraphiteCache": struct{}{},
	"DawnWebGPUCache":   struct{}{},
	"Service Worker":    struct{}{},
	"Extension State":   struct{}{},
	"blob_storage":      struct{}{},
	"Crashpad":          struct{}{},
}

type chromeVersion struct {
	WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
}

type ChromeLauncher struct {
	locker        sync.Mutex
	httpCli       input_itf.HttpCli
	debugPort     int
	profileDir    string
	launchTimeout time.Duration
	callTimeout   time.Duration
	endpoint      string
	browserWS     string
	proc          *exec.Cmd
}

func NewChromeLauncher(cfg *input_itf.MCPChromeConfig, dataDir string, httpCli input_itf.HttpCli) *ChromeLauncher {
	return &ChromeLauncher{
		httpCli:       httpCli,
		debugPort:     cfg.DebugPort,
		profileDir:    filepath.Join(dataDir, cfg.ProfileDir),
		launchTimeout: cfg.LaunchTimeout,
		callTimeout:   cfg.CallTimeout,
	}
}

func (l *ChromeLauncher) Endpoint() string {
	l.locker.Lock()
	defer l.locker.Unlock()

	return l.endpoint
}

func (l *ChromeLauncher) Ensure() error {
	l.locker.Lock()
	defer l.locker.Unlock()

	if l.connect() == nil {
		return nil
	}

	binary := chromeBinary()
	if binary == "" {
		return custom_error.TypedCritical(enums.ErrChromeNotFound, "cannot find Google Chrome on this machine")
	}

	if _, err := os.Stat(l.profileDir); err != nil {
		if err := seedChromeProfile(defaultChromeProfile(), l.profileDir); err != nil {
			return err
		}
	}

	port, err := freeLoopbackPort(l.debugPort)
	if err != nil {
		return err
	}

	if err := l.forgetPort(); err != nil {
		return err
	}

	cmd := exec.Command(binary,
		"--remote-debugging-address="+constances.GlobalLocalHost,
		"--remote-debugging-port="+strconv.Itoa(port),
		"--user-data-dir="+l.profileDir,
		"--no-first-run",
		"--no-default-browser-check",
	)

	if err := cmd.Start(); err != nil {
		return custom_error.TypedCritical(enums.ErrChromeLaunchFailed, "cannot start Chrome: %v", err)
	}

	l.proc = cmd

	exited := make(chan struct{})

	go func() {
		_ = cmd.Wait()
		close(exited)
	}()

	deadline := time.Now().Add(l.launchTimeout)

	for time.Now().Before(deadline) {
		if l.attach(port) == nil {
			return l.rememberPort(port)
		}

		select {
		case <-exited:
			l.proc = nil

			return custom_error.TypedCritical(
				enums.ErrChromeLaunchFailed,
				"Chrome quit before it opened its debugging port; another Chrome may already be running on the profile at %s",
				l.profileDir,
			)
		case <-time.After(chromeProbeInterval):
		}
	}

	l.kill()

	return custom_error.TypedCritical(
		enums.ErrChromeLaunchFailed,
		"Chrome did not start listening on %s:%d within %s", constances.GlobalLocalHost, port, l.launchTimeout,
	)
}

func (l *ChromeLauncher) Stop() error {
	l.locker.Lock()
	defer l.locker.Unlock()

	defer func() {
		l.endpoint = ""
		l.browserWS = ""
		_ = l.forgetPort()
	}()

	if l.proc != nil && l.proc.Process != nil {
		proc := l.proc.Process
		l.proc = nil

		if err := proc.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
			return custom_error.TypedCritical(enums.ErrChromeLaunchFailed, "cannot stop Chrome: %v", err)
		}

		return nil
	}

	return l.closeAdopted()
}

func (l *ChromeLauncher) connect() error {
	port, browserWS, err := l.rememberedPort()
	if err != nil {
		return err
	}

	if err := l.attach(port); err != nil {
		return err
	}

	if l.browserWS != browserWS {
		l.endpoint = ""
		l.browserWS = ""

		return custom_error.TypedCritical(
			enums.ErrChromeNotConnected,
			"port %d is answered by another browser, not the Chrome that Nexo left running", port,
		)
	}

	return nil
}

func (l *ChromeLauncher) attach(port int) error {
	endpoint := fmt.Sprintf("http://%s:%d", constances.GlobalLocalHost, port)

	version := &chromeVersion{}
	if err := l.httpCli.GetJSON(endpoint+"/json/version", version); err != nil {
		return custom_error.TypedCritical(enums.ErrChromeNotConnected, "nothing answers on %s: %v", endpoint, err)
	}

	if version.WebSocketDebuggerURL == "" {
		return custom_error.TypedCritical(
			enums.ErrChromeNotConnected,
			"%s is answered by something that does not speak the devtools protocol", endpoint,
		)
	}

	l.endpoint = endpoint
	l.browserWS = version.WebSocketDebuggerURL

	return nil
}

func (l *ChromeLauncher) rememberedPort() (int, string, error) {
	raw, err := os.ReadFile(filepath.Join(l.profileDir, chromePortFile))
	if err != nil {
		return 0, "", custom_error.TypedCritical(enums.ErrChromeNotConnected, "Nexo has no Chrome running: %v", err)
	}

	lines := strings.SplitN(strings.TrimSpace(string(raw)), "\n", 2)
	if len(lines) < 2 {
		return 0, "", custom_error.TypedCritical(enums.ErrChromeNotConnected, "the remembered chrome endpoint is incomplete")
	}

	port, err := strconv.Atoi(strings.TrimSpace(lines[0]))
	if err != nil || port <= 0 {
		return 0, "", custom_error.TypedCritical(enums.ErrChromeNotConnected, "the remembered chrome port %q is unreadable", lines[0])
	}

	browserWS := strings.TrimSpace(lines[1])
	if browserWS == "" {
		return 0, "", custom_error.TypedCritical(enums.ErrChromeNotConnected, "the remembered chrome endpoint has no browser websocket")
	}

	return port, browserWS, nil
}

func (l *ChromeLauncher) rememberPort(port int) error {
	body := strconv.Itoa(port) + "\n" + l.browserWS + "\n"

	if err := os.WriteFile(filepath.Join(l.profileDir, chromePortFile), []byte(body), chromeSeedFileMode); err != nil {
		return custom_error.TypedCritical(enums.ErrChromeLaunchFailed, "cannot record the chrome debugging endpoint: %v", err)
	}

	return nil
}

func (l *ChromeLauncher) forgetPort() error {
	err := os.Remove(filepath.Join(l.profileDir, chromePortFile))
	if err == nil || errors.Is(err, os.ErrNotExist) {
		return nil
	}

	return custom_error.TypedCritical(enums.ErrChromeLaunchFailed, "cannot clear the stale chrome debugging endpoint: %v", err)
}

func (l *ChromeLauncher) closeAdopted() error {
	if l.browserWS == "" && l.connect() != nil {
		return nil
	}

	return CDPQuitBrowser(l.browserWS, l.callTimeout)
}

func (l *ChromeLauncher) kill() {
	if l.proc == nil || l.proc.Process == nil {
		l.proc = nil

		return
	}

	_ = l.proc.Process.Kill()
	l.proc = nil
}

func freeLoopbackPort(preferred int) (int, error) {
	for _, port := range []int{preferred, 0} {
		listener, err := net.Listen("tcp4", net.JoinHostPort(constances.GlobalLocalHost, strconv.Itoa(port)))
		if err != nil {
			continue
		}

		bound := listener.Addr().(*net.TCPAddr).Port

		if err := listener.Close(); err != nil {
			return 0, custom_error.TypedCritical(enums.ErrChromeLaunchFailed, "cannot release the reserved chrome port: %v", err)
		}

		return bound, nil
	}

	return 0, custom_error.TypedCritical(enums.ErrChromeLaunchFailed, "cannot reserve a local port for Chrome")
}

func chromeBinary() string {
	home, _ := os.UserHomeDir()

	switch runtime.GOOS {
	case "darwin":
		return firstExistingFile([]string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			filepath.Join(home, "Applications", "Google Chrome.app", "Contents", "MacOS", "Google Chrome"),
			"/Applications/Google Chrome Canary.app/Contents/MacOS/Google Chrome Canary",
			filepath.Join(home, "Applications", "Google Chrome Canary.app", "Contents", "MacOS", "Google Chrome Canary"),
		})
	case "windows":
		return firstExistingFile([]string{
			filepath.Join(os.Getenv("ProgramFiles"), "Google", "Chrome", "Application", "chrome.exe"),
			filepath.Join(os.Getenv("ProgramFiles(x86)"), "Google", "Chrome", "Application", "chrome.exe"),
			filepath.Join(os.Getenv("LOCALAPPDATA"), "Google", "Chrome", "Application", "chrome.exe"),
		})
	default:
		for _, name := range []string{"google-chrome", "google-chrome-stable", "chromium", "chromium-browser"} {
			if path, err := exec.LookPath(name); err == nil {
				return path
			}
		}

		return ""
	}
}

func firstExistingFile(paths []string) string {
	for _, path := range paths {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path
		}
	}

	return ""
}

func defaultChromeProfile() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "Google", "Chrome")
	case "windows":
		return filepath.Join(os.Getenv("LOCALAPPDATA"), "Google", "Chrome", "User Data")
	default:
		return filepath.Join(home, ".config", "google-chrome")
	}
}

func seedChromeProfile(source, target string) error {
	if source == "" {
		return nil
	}

	if _, err := os.Stat(source); err != nil {
		return nil
	}

	if err := os.MkdirAll(target, chromeProfileMode); err != nil {
		return custom_error.TypedCritical(enums.ErrChromeLaunchFailed, "cannot create the chrome profile %q: %v", target, err)
	}

	if err := copyProfileFile(filepath.Join(source, "Local State"), filepath.Join(target, "Local State")); err != nil {
		return err
	}

	return copyProfileDir(filepath.Join(source, "Default"), filepath.Join(target, "Default"))
}

func copyProfileDir(source, target string) error {
	entries, err := os.ReadDir(source)
	if err != nil {
		return nil
	}

	if err := os.MkdirAll(target, chromeProfileMode); err != nil {
		return custom_error.TypedCritical(enums.ErrChromeLaunchFailed, "cannot create the chrome profile %q: %v", target, err)
	}

	for _, entry := range entries {
		name := entry.Name()

		if entry.IsDir() {
			if _, skipped := chromeSkippedDirs[name]; skipped {
				continue
			}

			if err := copyProfileDir(filepath.Join(source, name), filepath.Join(target, name)); err != nil {
				return err
			}

			continue
		}

		if strings.HasPrefix(name, "Singleton") || strings.HasSuffix(name, ".lock") {
			continue
		}

		if err := copyProfileFile(filepath.Join(source, name), filepath.Join(target, name)); err != nil {
			return err
		}
	}

	return nil
}

func copyProfileFile(source, target string) error {
	in, err := os.Open(source)
	if err != nil {
		return nil
	}

	defer in.Close()

	out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, chromeSeedFileMode)
	if err != nil {
		return custom_error.TypedCritical(enums.ErrChromeLaunchFailed, "cannot write the chrome profile file %q: %v", target, err)
	}

	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return custom_error.TypedCritical(enums.ErrChromeLaunchFailed, "cannot copy the chrome profile file %q: %v", target, err)
	}

	return nil
}
