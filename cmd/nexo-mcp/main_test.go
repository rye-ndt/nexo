package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func testOptions() *options {
	return &options{
		dataDir:      "nexo",
		endpointFile: "control.json",
		appName:      "Nexo",
		launch:       false,
		launchWait:   1,
	}
}

func writeEndpoint(t *testing.T, path, url string) {
	t.Helper()

	raw, err := json.Marshal(&endpoint{
		URL:         url,
		Token:       "a-token",
		TokenHeader: "X-Nexo-Control-Token",
		PID:         os.Getpid(),
	})
	if err != nil {
		t.Fatalf("marshal endpoint: %v", err)
	}

	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatalf("write endpoint: %v", err)
	}
}

func TestForwardsRequestAndWritesReply(t *testing.T) {
	var gotBody []byte
	var gotToken, gotType string

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		gotToken = r.Header.Get("X-Nexo-Control-Token")
		gotType = r.Header.Get("Content-Type")

		io.WriteString(w, `{"jsonrpc":"2.0","id":1,"result":{"tools":[]}}`)
	}))
	defer upstream.Close()

	path := filepath.Join(t.TempDir(), "control.json")
	writeEndpoint(t, path, upstream.URL)

	request := `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`
	out := &bytes.Buffer{}

	if err := run(strings.NewReader(request+"\n"), out, path, testOptions()); err != nil {
		t.Fatalf("run: %v", err)
	}

	if string(gotBody) != request {
		t.Fatalf("forwarded body = %s, want %s", gotBody, request)
	}

	if gotToken != "a-token" {
		t.Fatalf("token header = %q, want %q", gotToken, "a-token")
	}

	if gotType != "application/json" {
		t.Fatalf("content type = %q", gotType)
	}

	want := `{"jsonrpc":"2.0","id":1,"result":{"tools":[]}}` + "\n"
	if out.String() != want {
		t.Fatalf("stdout = %q, want %q", out.String(), want)
	}
}

func TestNotificationWritesNothing(t *testing.T) {
	calls := 0

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
	}))
	defer upstream.Close()

	path := filepath.Join(t.TempDir(), "control.json")
	writeEndpoint(t, path, upstream.URL)

	out := &bytes.Buffer{}
	in := strings.NewReader(`{"jsonrpc":"2.0","method":"notifications/initialized"}` + "\n")

	if err := run(in, out, path, testOptions()); err != nil {
		t.Fatalf("run: %v", err)
	}

	if out.Len() != 0 {
		t.Fatalf("stdout = %q, want nothing", out.String())
	}

	if calls != 0 {
		t.Fatalf("upstream calls = %d, want 0", calls)
	}
}

func TestMissingEndpointFileFailsMessageAndKeepsServing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "control.json")

	in := strings.NewReader(`{"jsonrpc":"2.0","id":"abc","method":"tools/list"}` + "\n" +
		`{"jsonrpc":"2.0","id":7,"method":"tools/list"}` + "\n")
	out := &bytes.Buffer{}

	if err := run(in, out, path, testOptions()); err != nil {
		t.Fatalf("run: %v", err)
	}

	lines := strings.Split(strings.TrimSuffix(out.String(), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("got %d replies, want 2: %q", len(lines), out.String())
	}

	first := &rpcErrorResponse{}
	if err := json.Unmarshal([]byte(lines[0]), first); err != nil {
		t.Fatalf("decode first reply: %v", err)
	}

	if string(first.ID) != `"abc"` {
		t.Fatalf("first id = %s, want \"abc\"", first.ID)
	}

	if first.JSONRPC != "2.0" || first.Error.Code != internalErrorCode {
		t.Fatalf("unexpected error envelope: %+v", first)
	}

	if !strings.Contains(first.Error.Message, "-launch") {
		t.Fatalf("error message does not mention -launch: %q", first.Error.Message)
	}

	second := &rpcErrorResponse{}
	if err := json.Unmarshal([]byte(lines[1]), second); err != nil {
		t.Fatalf("decode second reply: %v", err)
	}

	if string(second.ID) != "7" {
		t.Fatalf("second id = %s, want 7", second.ID)
	}
}

func TestRetriesAgainstRewrittenEndpoint(t *testing.T) {
	path := filepath.Join(t.TempDir(), "control.json")

	restarted := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"jsonrpc":"2.0","id":2,"result":"served"}`)
	}))
	defer restarted.Close()

	stale := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeEndpoint(t, path, restarted.URL)
		w.WriteHeader(http.StatusForbidden)
	}))
	defer stale.Close()

	writeEndpoint(t, path, stale.URL)

	out := &bytes.Buffer{}
	in := strings.NewReader(`{"jsonrpc":"2.0","id":2,"method":"tools/call"}` + "\n")

	if err := run(in, out, path, testOptions()); err != nil {
		t.Fatalf("run: %v", err)
	}

	want := `{"jsonrpc":"2.0","id":2,"result":"served"}` + "\n"
	if out.String() != want {
		t.Fatalf("stdout = %q, want %q", out.String(), want)
	}
}

func TestPayloadLargerThanScannerDefault(t *testing.T) {
	var got int

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		got = len(body)

		io.WriteString(w, `{"jsonrpc":"2.0","id":3,"result":"ok"}`)
	}))
	defer upstream.Close()

	path := filepath.Join(t.TempDir(), "control.json")
	writeEndpoint(t, path, upstream.URL)

	request := `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"guidance":"` +
		strings.Repeat("x", 200_000) + `"}}`

	out := &bytes.Buffer{}

	if err := run(strings.NewReader(request+"\n"), out, path, testOptions()); err != nil {
		t.Fatalf("run: %v", err)
	}

	if got != len(request) {
		t.Fatalf("forwarded %d bytes, want %d", got, len(request))
	}

	want := `{"jsonrpc":"2.0","id":3,"result":"ok"}` + "\n"
	if out.String() != want {
		t.Fatalf("stdout = %q, want %q", out.String(), want)
	}
}

func TestAStaleEndpointFromACrashedAppReadsAsNotRunning(t *testing.T) {
	served := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`{"jsonrpc":"2.0","id":3,"result":"served"}`))
	}))
	defer served.Close()

	path := filepath.Join(t.TempDir(), "control.json")

	raw, err := json.Marshal(&endpoint{
		URL:         served.URL,
		Token:       "a-token",
		TokenHeader: "X-Nexo-Control-Token",
		PID:         deadPID(t),
	})
	if err != nil {
		t.Fatalf("marshal endpoint: %v", err)
	}

	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatalf("write endpoint: %v", err)
	}

	out := &bytes.Buffer{}

	if err := run(strings.NewReader(`{"jsonrpc":"2.0","id":3,"method":"tools/list"}`+"\n"), out, path, testOptions()); err != nil {
		t.Fatalf("run: %v", err)
	}

	if !strings.Contains(out.String(), "not running") {
		t.Fatalf("a stale endpoint did not read as a closed app: %s", out.String())
	}
}

func deadPID(t *testing.T) int {
	t.Helper()

	spent := exec.Command("true")
	if err := spent.Run(); err != nil {
		t.Fatalf("run throwaway process: %v", err)
	}

	return spent.ProcessState.Pid()
}

func TestADroppedAnswerIsNotReplayedAgainstTheSameApp(t *testing.T) {
	calls := 0

	served := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++

		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer served.Close()

	path := filepath.Join(t.TempDir(), "control.json")
	writeEndpoint(t, path, served.URL)

	out := &bytes.Buffer{}

	create := `{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"create_workflow"}}`

	if err := run(strings.NewReader(create+"\n"), out, path, testOptions()); err != nil {
		t.Fatalf("run: %v", err)
	}

	if calls != 1 {
		t.Fatalf("the app was called %d times, want 1: create_workflow must not be replayed", calls)
	}

	if !strings.Contains(out.String(), "refused the call") {
		t.Fatalf("the failure was not reported back: %s", out.String())
	}
}
