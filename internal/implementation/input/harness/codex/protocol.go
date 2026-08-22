package codex

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"hexago/internal/helpers"
	"hexago/internal/helpers/custom_error"
	"hexago/internal/implementation/input/harness/harness_helper"
	input_itf "hexago/internal/interface/input"
)

const maxActivity = 40

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type rpcEnvelope struct {
	ID     json.RawMessage `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
	Result json.RawMessage `json:"result"`
	Error  *rpcError       `json:"error"`
}

type rpcResponse struct {
	result json.RawMessage
	err    *rpcError
}

type tokenBreakdown struct {
	InputTokens           int `json:"inputTokens"`
	CachedInputTokens     int `json:"cachedInputTokens"`
	OutputTokens          int `json:"outputTokens"`
	ReasoningOutputTokens int `json:"reasoningOutputTokens"`
	TotalTokens           int `json:"totalTokens"`
}

type tokenUsage struct {
	Last               tokenBreakdown `json:"last"`
	Total              tokenBreakdown `json:"total"`
	ModelContextWindow *int           `json:"modelContextWindow"`
}

type codexItem struct {
	Type    string   `json:"type"`
	Text    string   `json:"text"`
	Summary []string `json:"summary"`
	Command string   `json:"command"`
	Server  string   `json:"server"`
	Tool    string   `json:"tool"`
	Query   string   `json:"query"`
	Path    string   `json:"path"`
	Prompt  string   `json:"prompt"`
	Changes []struct {
		Path string `json:"path"`
	} `json:"changes"`
}

type agentProc struct {
	cmd       *exec.Cmd
	stdin     io.WriteCloser
	stdout    io.ReadCloser
	done      chan struct{}
	exited    chan struct{}
	stopOnce  sync.Once
	lastOut   atomic.Int64
	ctxWindow int
	threadID  string
	effort    string

	writeMu sync.Mutex
	rpcMu   sync.Mutex
	nextID  int64
	pending map[int64]chan rpcResponse

	sendMu  sync.Mutex
	stateMu sync.Mutex
	turnID  string

	usageMu sync.Mutex
	usage   input_itf.ContextUsage

	activityMu  sync.Mutex
	activity    []input_itf.Activity
	activitySeq int
}

func newAgentProc(cmd *exec.Cmd, stdin io.WriteCloser, stdout io.ReadCloser, ctxWindow int) *agentProc {
	return &agentProc{
		cmd:       cmd,
		stdin:     stdin,
		stdout:    stdout,
		done:      make(chan struct{}),
		exited:    make(chan struct{}),
		ctxWindow: ctxWindow,
		pending:   map[int64]chan rpcResponse{},
	}
}

func (p *agentProc) initialize() error {
	if _, err := p.call("initialize", map[string]any{
		"clientInfo": map[string]string{"name": "nexo", "version": "1"},
	}); err != nil {
		return err
	}

	return p.notify("initialized", nil)
}

func (p *agentProc) startThread(params map[string]any) (string, error) {
	raw, err := p.call("thread/start", params)
	if err != nil {
		return "", err
	}

	var response struct {
		Thread struct {
			ID string `json:"id"`
		} `json:"thread"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		return "", custom_error.Critical("decode thread/start response: %v", err)
	}
	if response.Thread.ID == "" {
		return "", custom_error.Critical("thread/start returned an empty thread id")
	}
	p.stateMu.Lock()
	p.threadID = response.Thread.ID
	p.stateMu.Unlock()

	return response.Thread.ID, nil
}

func (p *agentProc) send(message string) error {
	p.sendMu.Lock()
	defer p.sendMu.Unlock()

	input := []map[string]string{{"type": "text", "text": message}}
	threadID, effort := p.identity()
	active := p.activeTurn()

	if active != "" {
		_, err := p.call("turn/steer", map[string]any{
			"threadId":       threadID,
			"expectedTurnId": active,
			"input":          input,
		})
		if err == nil {
			return nil
		}
		if p.activeTurn() != "" {
			return err
		}
	}

	raw, err := p.call("turn/start", map[string]any{
		"threadId": threadID,
		"input":    input,
		"effort":   effort,
	})
	if err != nil {
		return err
	}

	var response struct {
		Turn struct {
			ID string `json:"id"`
		} `json:"turn"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		return custom_error.Critical("decode turn/start response: %v", err)
	}
	if response.Turn.ID == "" {
		return custom_error.Critical("turn/start returned an empty turn id")
	}

	p.stateMu.Lock()
	p.turnID = response.Turn.ID
	p.stateMu.Unlock()

	return nil
}

func (p *agentProc) activeTurn() string {
	p.stateMu.Lock()
	defer p.stateMu.Unlock()

	return p.turnID
}

func (p *agentProc) identity() (string, string) {
	p.stateMu.Lock()
	defer p.stateMu.Unlock()

	return p.threadID, p.effort
}

func (p *agentProc) thread() string {
	p.stateMu.Lock()
	defer p.stateMu.Unlock()

	return p.threadID
}

func (p *agentProc) call(method string, params any) (json.RawMessage, error) {
	p.rpcMu.Lock()
	p.nextID++
	id := p.nextID
	response := make(chan rpcResponse, 1)
	p.pending[id] = response
	p.rpcMu.Unlock()

	defer func() {
		p.rpcMu.Lock()
		delete(p.pending, id)
		p.rpcMu.Unlock()
	}()

	if err := p.write(map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"method":  method,
		"params":  params,
	}); err != nil {
		return nil, err
	}

	timer := time.NewTimer(rpcTimeout)
	defer timer.Stop()

	select {
	case reply := <-response:
		if reply.err != nil {
			return nil, custom_error.Critical("%s failed (%d): %s", method, reply.err.Code, reply.err.Message)
		}
		return reply.result, nil
	case <-p.exited:
		return nil, custom_error.Critical("codex app server exited during %s", method)
	case <-timer.C:
		return nil, custom_error.Critical("codex app server timed out during %s", method)
	}
}

func (p *agentProc) notify(method string, params any) error {
	message := map[string]any{"jsonrpc": "2.0", "method": method}
	if params != nil {
		message["params"] = params
	}

	return p.write(message)
}

func (p *agentProc) write(message any) error {
	raw, err := json.Marshal(message)
	if err != nil {
		return custom_error.Critical("encode codex app-server message: %v", err)
	}

	p.writeMu.Lock()
	_, err = p.stdin.Write(append(raw, '\n'))
	p.writeMu.Unlock()
	if err != nil {
		return custom_error.Critical("write codex app-server message: %v", err)
	}

	return nil
}

func (p *agentProc) read() {
	scanner := bufio.NewScanner(p.stdout)
	scanner.Buffer(make([]byte, 64*1024), 8*1024*1024)

	for scanner.Scan() {
		p.lastOut.Store(helpers.NewUTCUnix())
		line := scanner.Bytes()
		p.track(line)
	}
}

func (p *agentProc) track(line []byte) {
	event := &rpcEnvelope{}
	if err := json.Unmarshal(line, event); err != nil {
		return
	}

	if len(event.ID) != 0 && event.Method == "" {
		var id int64
		if err := json.Unmarshal(event.ID, &id); err == nil {
			p.rpcMu.Lock()
			pending := p.pending[id]
			p.rpcMu.Unlock()
			if pending != nil {
				pending <- rpcResponse{result: event.Result, err: event.Error}
			}
		}
		return
	}

	if len(event.ID) != 0 && event.Method != "" {
		p.handleServerRequest(event)
		return
	}

	switch event.Method {
	case "thread/tokenUsage/updated":
		p.trackUsage(event.Params)
	case "item/completed":
		p.trackActivity(event.Params)
	case "turn/completed":
		p.trackTurnCompleted(event.Params)
	}
}

func (p *agentProc) handleServerRequest(event *rpcEnvelope) {
	var id any
	if err := json.Unmarshal(event.ID, &id); err != nil {
		return
	}

	var result any
	switch event.Method {
	case "item/commandExecution/requestApproval", "item/fileChange/requestApproval":
		result = map[string]string{"decision": "accept"}
	case "currentTime/read":
		result = map[string]int64{"currentTimeAt": time.Now().Unix()}
	default:
		p.write(map[string]any{
			"jsonrpc": "2.0",
			"id":      id,
			"error":   map[string]any{"code": -32601, "message": "request is not supported by nexo"},
		})
		return
	}

	p.write(map[string]any{"jsonrpc": "2.0", "id": id, "result": result})
}

func (p *agentProc) trackTurnCompleted(raw json.RawMessage) {
	var event struct {
		ThreadID string `json:"threadId"`
		Turn     struct {
			ID string `json:"id"`
		} `json:"turn"`
	}
	if err := json.Unmarshal(raw, &event); err != nil || event.ThreadID != p.thread() {
		return
	}

	p.stateMu.Lock()
	if p.turnID == event.Turn.ID {
		p.turnID = ""
	}
	p.stateMu.Unlock()
}

func (p *agentProc) trackUsage(raw json.RawMessage) {
	var event struct {
		ThreadID   string     `json:"threadId"`
		TokenUsage tokenUsage `json:"tokenUsage"`
	}
	if err := json.Unmarshal(raw, &event); err != nil || event.ThreadID != p.thread() {
		return
	}

	total := event.TokenUsage.Total
	input := total.InputTokens - total.CachedInputTokens
	if input < 0 {
		input = 0
	}

	p.usageMu.Lock()
	window := p.ctxWindow
	if event.TokenUsage.ModelContextWindow != nil && *event.TokenUsage.ModelContextWindow > 0 {
		window = *event.TokenUsage.ModelContextWindow
	}
	p.usage = input_itf.ContextUsage{
		Total:  window,
		Used:   event.TokenUsage.Last.TotalTokens,
		Billed: total.OutputTokens + total.ReasoningOutputTokens,
		Input:  input,
		Cached: total.CachedInputTokens,
	}
	p.usageMu.Unlock()
}

func (p *agentProc) snapshotUsage() *input_itf.ContextUsage {
	p.usageMu.Lock()
	defer p.usageMu.Unlock()

	snapshot := p.usage
	if snapshot.Total == 0 {
		snapshot.Total = p.ctxWindow
	}

	return &snapshot
}

func (p *agentProc) trackActivity(raw json.RawMessage) {
	var event struct {
		ThreadID string    `json:"threadId"`
		Item     codexItem `json:"item"`
	}
	if err := json.Unmarshal(raw, &event); err != nil || event.ThreadID != p.thread() {
		return
	}

	text := activityText(&event.Item)
	if text == "" {
		return
	}

	p.activityMu.Lock()
	p.activitySeq++
	p.activity = append(p.activity, input_itf.Activity{
		Seq:  p.activitySeq,
		At:   helpers.NewUTC(),
		Text: text,
	})
	if len(p.activity) > maxActivity {
		p.activity = append([]input_itf.Activity(nil), p.activity[len(p.activity)-maxActivity:]...)
	}
	p.activityMu.Unlock()
}

func (p *agentProc) snapshotActivity() []input_itf.Activity {
	p.activityMu.Lock()
	defer p.activityMu.Unlock()

	return append([]input_itf.Activity(nil), p.activity...)
}

func (p *agentProc) stop() error {
	var err error
	p.stopOnce.Do(func() {
		close(p.done)
		p.stdin.Close()
		err = harness_helper.SignalProc(p.cmd)
	})

	return err
}

func activityText(item *codexItem) string {
	switch item.Type {
	case "agentMessage":
		return clip(firstLine(item.Text), 180)
	case "plan":
		return "Planning: " + clip(firstLine(item.Text), 160)
	case "reasoning":
		if len(item.Summary) > 0 {
			return clip(firstLine(item.Summary[0]), 180)
		}
	case "commandExecution":
		return "Running " + clip(item.Command, 160)
	case "fileChange":
		if len(item.Changes) > 0 {
			return "Editing " + shortPath(item.Changes[0].Path)
		}
	case "mcpToolCall":
		return "Using " + strings.Trim(item.Server+"."+item.Tool, ".")
	case "dynamicToolCall":
		return "Using " + item.Tool
	case "collabAgentToolCall":
		return "Delegating: " + clip(firstLine(item.Prompt), 150)
	case "webSearch":
		return "Searching web for " + clip(item.Query, 140)
	case "imageView":
		return "Viewing " + shortPath(item.Path)
	}

	return ""
}

func firstLine(text string) string {
	line, _, _ := strings.Cut(strings.TrimSpace(text), "\n")
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

	return fmt.Sprintf("%s/%s", parts[len(parts)-2], parts[len(parts)-1])
}
