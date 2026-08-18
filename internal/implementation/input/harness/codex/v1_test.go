package codex

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"testing"

	"hexago/internal/helpers/constances"
	core_itf "hexago/internal/interface/core"
)

func TestSendSteersAnActiveTurn(t *testing.T) {
	reader, writer := io.Pipe()
	defer reader.Close()
	defer writer.Close()

	proc := newAgentProc(nil, writer, nil, 1000)
	proc.threadID = "thread-id"
	proc.turnID = "turn-id"

	checked := make(chan error, 1)
	go func() {
		line, err := bufio.NewReader(reader).ReadBytes('\n')
		if err != nil {
			checked <- err
			return
		}

		var request struct {
			ID     int            `json:"id"`
			Method string         `json:"method"`
			Params map[string]any `json:"params"`
		}
		if err := json.Unmarshal(line, &request); err != nil {
			checked <- err
			return
		}
		if request.Method != "turn/steer" {
			checked <- fmt.Errorf("method = %q, want turn/steer", request.Method)
			return
		}
		if request.Params["expectedTurnId"] != "turn-id" {
			checked <- fmt.Errorf("expectedTurnId = %v, want turn-id", request.Params["expectedTurnId"])
			return
		}

		response, err := json.Marshal(map[string]any{"id": request.ID, "result": map[string]string{"turnId": "turn-id"}})
		if err == nil {
			proc.track(response)
		}
		checked <- err
	}()

	if err := proc.send("continue"); err != nil {
		t.Fatal(err)
	}
	if err := <-checked; err != nil {
		t.Fatal(err)
	}
}

func TestCodexMCPConfigScopesGatewayRequestsToAgent(t *testing.T) {
	gateway := &core_itf.MCPGateway{
		BaseURL:     "http://127.0.0.1:1234",
		Token:       "gateway-token",
		TokenHeader: "X-Gateway-Token",
		Servers:     []core_itf.MCPGatewayServer{{Name: "harness"}},
	}

	config := codexMCPConfig(gateway, "agent-id")
	server := config["harness"].(map[string]any)
	headers := server["http_headers"].(map[string]string)

	if got := server["url"]; got != "http://127.0.0.1:1234/mcp/harness" {
		t.Fatalf("unexpected gateway url: %v", got)
	}
	if got := headers[constances.GatewayAgentHeader]; got != "agent-id" {
		t.Fatalf("unexpected agent header: %q", got)
	}
	if got := headers["X-Gateway-Token"]; got != "gateway-token" {
		t.Fatalf("unexpected gateway token: %q", got)
	}
}

func TestTrackUsageSeparatesCachedInput(t *testing.T) {
	proc := &agentProc{threadID: "thread", ctxWindow: 1000}
	raw, err := json.Marshal(map[string]any{
		"threadId": "thread",
		"tokenUsage": map[string]any{
			"last": map[string]int{
				"inputTokens": 100, "cachedInputTokens": 40, "outputTokens": 20,
				"reasoningOutputTokens": 5, "totalTokens": 125,
			},
			"total": map[string]int{
				"inputTokens": 500, "cachedInputTokens": 300, "outputTokens": 80,
				"reasoningOutputTokens": 20, "totalTokens": 600,
			},
			"modelContextWindow": 2000,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	proc.trackUsage(raw)
	usage := proc.snapshotUsage()

	if usage.Total != 2000 || usage.Used != 125 || usage.Input != 200 || usage.Cached != 300 || usage.Billed != 100 {
		t.Fatalf("unexpected usage: %+v", usage)
	}
}

func TestActivityTextSummarizesCodexItems(t *testing.T) {
	change := codexItem{Type: "fileChange"}
	change.Changes = append(change.Changes, struct {
		Path string `json:"path"`
	}{Path: "/workspace/internal/service.go"})

	tests := []struct {
		item codexItem
		want string
	}{
		{item: codexItem{Type: "commandExecution", Command: "go test ./..."}, want: "Running go test ./..."},
		{item: change, want: "Editing internal/service.go"},
		{item: codexItem{Type: "mcpToolCall", Server: "harness", Tool: "report_task"}, want: "Using harness.report_task"},
		{item: codexItem{Type: "webSearch", Query: "Codex app server"}, want: "Searching web for Codex app server"},
	}

	for _, test := range tests {
		if got := activityText(&test.item); got != test.want {
			t.Errorf("activityText(%s) = %q, want %q", test.item.Type, got, test.want)
		}
	}
}
