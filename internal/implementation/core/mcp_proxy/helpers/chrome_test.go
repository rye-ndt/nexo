package mcp_proxy_helpers

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	input_itf "hexago/internal/interface/input"
)

type stubHttpCli struct {
	version any
	err     error
	asked   []string
}

func (c *stubHttpCli) GetJSON(url string, v any) error {
	c.asked = append(c.asked, url)

	if c.err != nil {
		return c.err
	}

	raw, err := json.Marshal(c.version)
	if err != nil {
		return err
	}

	return json.Unmarshal(raw, v)
}

func (c *stubHttpCli) Reachable(string) error                        { return nil }
func (c *stubHttpCli) GetString(string) (string, error)              { return "", nil }
func (c *stubHttpCli) PostForm(string, map[string]string, any) error { return nil }
func (c *stubHttpCli) PostJSON(string, any, any) error               { return nil }
func (c *stubHttpCli) Download(string, string, *input_itf.DownloadParams) error {
	return nil
}

func (c *stubHttpCli) Stream(*input_itf.HttpRequest) (*input_itf.HttpResponse, error) {
	return nil, nil
}

func newTestLauncher(t *testing.T, httpCli input_itf.HttpCli) *ChromeLauncher {
	t.Helper()

	return NewChromeLauncher(
		&input_itf.MCPChromeConfig{
			DebugPort:     9222,
			ProfileDir:    "chrome",
			LaunchTimeout: time.Second,
			CallTimeout:   time.Second,
			MaxPageChars:  100,
		},
		t.TempDir(),
		httpCli,
	)
}

func publishPort(t *testing.T, launcher *ChromeLauncher, body string) {
	t.Helper()

	if err := os.MkdirAll(launcher.profileDir, chromeProfileMode); err != nil {
		t.Fatalf("create profile: %v", err)
	}

	path := filepath.Join(launcher.profileDir, chromePortFile)
	if err := os.WriteFile(path, []byte(body), chromeSeedFileMode); err != nil {
		t.Fatalf("publish port: %v", err)
	}
}

func TestFreeLoopbackPortSkipsAPortSomeoneElseHolds(t *testing.T) {
	squatter, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("squat a port: %v", err)
	}

	defer squatter.Close()

	taken := squatter.Addr().(*net.TCPAddr).Port

	port, err := freeLoopbackPort(taken)
	if err != nil {
		t.Fatalf("free port: %v", err)
	}

	if port == taken {
		t.Fatalf("handed out port %d while another program is listening on it", taken)
	}

	probe, err := net.Listen("tcp4", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
	if err != nil {
		t.Fatalf("port %d is not actually free: %v", port, err)
	}

	probe.Close()
}

func TestFreeLoopbackPortPrefersTheConfiguredOne(t *testing.T) {
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve a port: %v", err)
	}

	wanted := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	port, err := freeLoopbackPort(wanted)
	if err != nil {
		t.Fatalf("free port: %v", err)
	}

	if port != wanted {
		t.Fatalf("want the configured port %d, got %d", wanted, port)
	}
}

func TestConnectAdoptsTheChromeNexoLeftRunning(t *testing.T) {
	browserWS := "ws://127.0.0.1:41234/devtools/browser/nexo-one"
	httpCli := &stubHttpCli{version: map[string]string{"webSocketDebuggerUrl": browserWS}}

	launcher := newTestLauncher(t, httpCli)
	publishPort(t, launcher, "41234\n"+browserWS+"\n")

	if err := launcher.connect(); err != nil {
		t.Fatalf("connect: %v", err)
	}

	if launcher.endpoint != "http://127.0.0.1:41234" {
		t.Fatalf("want the remembered port in the endpoint, got %q", launcher.endpoint)
	}

	if httpCli.asked[0] != "http://127.0.0.1:41234/json/version" {
		t.Fatalf("probed %q instead of the remembered port", httpCli.asked[0])
	}
}

func TestConnectRefusesAStrangerAnsweringOnTheRememberedPort(t *testing.T) {
	httpCli := &stubHttpCli{version: map[string]string{
		"webSocketDebuggerUrl": "ws://127.0.0.1:9222/devtools/browser/somebody-else",
	}}

	launcher := newTestLauncher(t, httpCli)
	publishPort(t, launcher, "9222\nws://127.0.0.1:9222/devtools/browser/nexo-one\n")

	if err := launcher.connect(); err == nil {
		t.Fatal("connect accepted a browser that is not the one Nexo left running")
	}

	if launcher.endpoint != "" {
		t.Fatalf("a refused browser was still adopted as %q", launcher.endpoint)
	}
}

func TestAttachRefusesAnEndpointThatDoesNotSpeakDevTools(t *testing.T) {
	launcher := newTestLauncher(t, &stubHttpCli{version: map[string]string{"Browser": "Chrome/151.0.0.0"}})

	if err := launcher.attach(9222); err == nil {
		t.Fatal("attach accepted an answer with no browser websocket")
	}

	if launcher.endpoint != "" {
		t.Fatalf("a refused endpoint was still adopted as %q", launcher.endpoint)
	}
}

func TestRememberedPortRejectsAMarkerItCannotTrust(t *testing.T) {
	launcher := newTestLauncher(t, &stubHttpCli{})

	if _, _, err := launcher.rememberedPort(); err == nil {
		t.Fatal("a missing marker must not resolve to an endpoint")
	}

	publishPort(t, launcher, "9222\n")

	if _, _, err := launcher.rememberedPort(); err == nil {
		t.Fatal("a marker without a browser websocket must not resolve")
	}

	publishPort(t, launcher, "not-a-port\nws://127.0.0.1:9222/devtools/browser/nexo-one\n")

	if _, _, err := launcher.rememberedPort(); err == nil {
		t.Fatal("an unreadable port must not resolve")
	}
}

func TestRememberPortRoundTripsWhatEnsureAttachedTo(t *testing.T) {
	browserWS := "ws://127.0.0.1:41234/devtools/browser/nexo-one"

	launcher := newTestLauncher(t, &stubHttpCli{version: map[string]string{"webSocketDebuggerUrl": browserWS}})
	publishPort(t, launcher, "stale\n")

	if err := launcher.attach(41234); err != nil {
		t.Fatalf("attach: %v", err)
	}

	if err := launcher.rememberPort(41234); err != nil {
		t.Fatalf("remember port: %v", err)
	}

	port, remembered, err := launcher.rememberedPort()
	if err != nil {
		t.Fatalf("remembered port: %v", err)
	}

	if port != 41234 || remembered != browserWS {
		t.Fatalf("want %d and %q, got %d and %q", 41234, browserWS, port, remembered)
	}

	if err := launcher.forgetPort(); err != nil {
		t.Fatalf("forget port: %v", err)
	}

	if _, _, err := launcher.rememberedPort(); err == nil {
		t.Fatal("the stale marker survived")
	}

	if err := launcher.forgetPort(); err != nil {
		t.Fatalf("forgetting an absent marker must be a no-op, got %v", err)
	}
}
