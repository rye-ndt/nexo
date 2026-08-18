package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

var errEndpointStale = errors.New("the endpoint no longer answers")

const (
	internalErrorCode = -32603
	maxMessageSize    = 8 << 20
	requestTimeout    = 2 * time.Minute
	launchPollEvery   = 250 * time.Millisecond
)

type endpoint struct {
	URL         string `json:"url"`
	Token       string `json:"token"`
	TokenHeader string `json:"token_header"`
	PID         int    `json:"pid"`
}

type options struct {
	dataDir      string
	endpointFile string
	appName      string
	launch       bool
	launchWait   int
}

type rpcErrorBody struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type rpcErrorResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Error   rpcErrorBody    `json:"error"`
}

type shim struct {
	endpointFile string
	opts         *options
	client       *http.Client
	cached       *endpoint
}

func main() {
	opts := &options{}

	flag.StringVar(&opts.dataDir, "data-dir", "nexo", "app data dir name under the user config dir")
	flag.StringVar(&opts.endpointFile, "endpoint-file", "control.json", "endpoint file written by the app")
	flag.StringVar(&opts.appName, "app", "Nexo", "app to launch when it is not running")
	flag.BoolVar(&opts.launch, "launch", false, "launch the app when it is not running")
	flag.IntVar(&opts.launchWait, "launch-wait", 30, "seconds to wait for a launched app to serve")
	flag.Parse()

	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatalf("nexo-mcp: %v", err)
	}

	path := filepath.Join(configDir, opts.dataDir, opts.endpointFile)

	if err := run(os.Stdin, os.Stdout, path, opts); err != nil {
		log.Fatalf("nexo-mcp: %v", err)
	}
}

func run(in io.Reader, out io.Writer, endpointFile string, opts *options) error {
	s := &shim{
		endpointFile: endpointFile,
		opts:         opts,
		client:       &http.Client{Timeout: requestTimeout},
	}

	lines := bufio.NewScanner(in)
	lines.Buffer(make([]byte, 0, bufio.MaxScanTokenSize), maxMessageSize)

	writer := bufio.NewWriter(out)

	for lines.Scan() {
		request := bytes.TrimSpace(lines.Bytes())

		var envelope struct {
			ID json.RawMessage `json:"id"`
		}

		if err := json.Unmarshal(request, &envelope); err != nil || len(envelope.ID) == 0 {
			continue
		}

		reply, err := s.forward(request)
		if err != nil {
			reply = errorResponse(envelope.ID, err.Error())
		}

		if _, err := writer.Write(reply); err != nil {
			return err
		}

		if err := writer.WriteByte('\n'); err != nil {
			return err
		}

		if err := writer.Flush(); err != nil {
			return err
		}
	}

	return lines.Err()
}

func (s *shim) forward(request []byte) ([]byte, error) {
	current, err := s.resolve()
	if err != nil {
		return nil, err
	}

	reply, err := s.post(current, request)
	if err == nil {
		return reply, nil
	}

	if !errors.Is(err, errEndpointStale) {
		return nil, err
	}

	s.cached = nil

	refreshed, refreshErr := s.resolve()
	if refreshErr != nil {
		return nil, refreshErr
	}

	if refreshed.URL == current.URL && refreshed.Token == current.Token {
		return nil, err
	}

	return s.post(refreshed, request)
}

func (s *shim) resolve() (*endpoint, error) {
	if s.cached != nil {
		return s.cached, nil
	}

	current, err := readEndpoint(s.endpointFile)
	if err == nil && serving(current) {
		s.cached = current
		return current, nil
	}

	if !s.opts.launch {
		return nil, s.notRunning()
	}

	if err := exec.Command("open", "-a", s.opts.appName).Run(); err != nil {
		return nil, s.notRunning()
	}

	deadline := time.Now().Add(time.Duration(s.opts.launchWait) * time.Second)

	for time.Now().Before(deadline) {
		time.Sleep(launchPollEvery)

		if current, err := readEndpoint(s.endpointFile); err == nil && serving(current) {
			s.cached = current
			return current, nil
		}
	}

	return nil, fmt.Errorf("%s was launched but did not start serving within %d seconds, so there is nothing to control yet", s.opts.appName, s.opts.launchWait)
}

func (s *shim) notRunning() error {
	return fmt.Errorf("the %s app is not running, so there is nothing to control; open %s, or register this server with the -launch flag to have it started on demand", s.opts.appName, s.opts.appName)
}

func (s *shim) post(current *endpoint, request []byte) ([]byte, error) {
	call, err := http.NewRequest(http.MethodPost, current.URL, bytes.NewReader(request))
	if err != nil {
		return nil, fmt.Errorf("cannot address the %s app at %s", s.opts.appName, current.URL)
	}

	call.Header.Set("Content-Type", "application/json")

	if current.TokenHeader != "" {
		call.Header.Set(current.TokenHeader, current.Token)
	}

	response, err := s.client.Do(call)
	if err != nil {
		if errors.Is(err, syscall.ECONNREFUSED) {
			return nil, fmt.Errorf("cannot reach the %s app at %s: %w", s.opts.appName, current.URL, errEndpointStale)
		}

		return nil, fmt.Errorf("cannot reach the %s app at %s", s.opts.appName, current.URL)
	}
	defer response.Body.Close()

	reply, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read the answer from the %s app", s.opts.appName)
	}

	if response.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("the %s app refused the call with %s: %w", s.opts.appName, response.Status, errEndpointStale)
	}

	if response.StatusCode < 200 || response.StatusCode > 299 {
		return nil, fmt.Errorf("the %s app refused the call with %s", s.opts.appName, response.Status)
	}

	return bytes.TrimSpace(reply), nil
}

func serving(current *endpoint) bool {
	if current.PID <= 0 {
		return true
	}

	process, err := os.FindProcess(current.PID)
	if err != nil {
		return false
	}

	return process.Signal(syscall.Signal(0)) == nil
}

func readEndpoint(path string) (*endpoint, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	current := &endpoint{}
	if err := json.Unmarshal(raw, current); err != nil {
		return nil, err
	}

	if current.URL == "" {
		return nil, fmt.Errorf("no url in %s", path)
	}

	return current, nil
}

func errorResponse(id json.RawMessage, message string) []byte {
	reply, err := json.Marshal(&rpcErrorResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   rpcErrorBody{Code: internalErrorCode, Message: message},
	})
	if err != nil {
		return []byte(`{"jsonrpc":"2.0","id":null,"error":{"code":-32603,"message":"cannot encode the failure"}}`)
	}

	return reply
}
