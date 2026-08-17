package mcp_proxy_helpers

import (
	"errors"
	"fmt"
	"io"
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

type ChromeLauncher struct {
	locker        sync.Mutex
	httpCli       input_itf.HttpCli
	endpoint      string
	debugPort     int
	profileDir    string
	launchTimeout time.Duration
	proc          *exec.Cmd
}

func NewChromeLauncher(cfg *input_itf.MCPChromeConfig, dataDir string, httpCli input_itf.HttpCli) *ChromeLauncher {
	return &ChromeLauncher{
		httpCli:       httpCli,
		endpoint:      fmt.Sprintf("http://%s:%d", constances.GlobalLocalHost, cfg.DebugPort),
		debugPort:     cfg.DebugPort,
		profileDir:    filepath.Join(dataDir, cfg.ProfileDir),
		launchTimeout: cfg.LaunchTimeout,
	}
}

func (l *ChromeLauncher) Endpoint() string {
	return l.endpoint
}

func (l *ChromeLauncher) running() bool {
	return l.httpCli.GetJSON(l.endpoint+"/json/version", &map[string]any{}) == nil
}

func (l *ChromeLauncher) Ensure() error {
	l.locker.Lock()
	defer l.locker.Unlock()

	if l.running() {
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

	cmd := exec.Command(binary,
		"--remote-debugging-port="+strconv.Itoa(l.debugPort),
		"--user-data-dir="+l.profileDir,
		"--no-first-run",
		"--no-default-browser-check",
	)

	if err := cmd.Start(); err != nil {
		return custom_error.TypedCritical(enums.ErrChromeLaunchFailed, "cannot start Chrome: %v", err)
	}

	go func() {
		_ = cmd.Wait()
	}()

	l.proc = cmd

	deadline := time.Now().Add(l.launchTimeout)
	for time.Now().Before(deadline) {
		if l.running() {
			return nil
		}

		time.Sleep(chromeProbeInterval)
	}

	return custom_error.TypedCritical(
		enums.ErrChromeLaunchFailed,
		"Chrome did not start listening on %s within %s", l.endpoint, l.launchTimeout,
	)
}

func (l *ChromeLauncher) Stop() error {
	l.locker.Lock()
	defer l.locker.Unlock()

	if l.proc == nil || l.proc.Process == nil {
		return nil
	}

	proc := l.proc.Process
	l.proc = nil

	if err := proc.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
		return custom_error.TypedCritical(enums.ErrChromeLaunchFailed, "cannot stop Chrome: %v", err)
	}

	return nil
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
