package mcp_proxy

import (
	"encoding/json"
	"net/http"

	"hexago/internal/helpers/constances"

	"github.com/google/uuid"
)

const (
	jsonRPCVersion         = "2.0"
	defaultProtocolVersion = "2025-06-18"
)

const (
	methodInitialize = "initialize"
	methodToolsList  = "tools/list"
	methodToolsCall  = "tools/call"
)

const (
	codeParseError     = -32700
	codeMethodNotFound = -32601
)

type rpcTool struct {
	name        string
	description string
	input       map[string]any
	call        func(arguments json.RawMessage, agentID uuid.UUID) *toolResult
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

const contentText = "text"

type toolContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type toolResult struct {
	Content []toolContent `json:"content"`
	IsError bool          `json:"isError"`
}

func serveRPC(w http.ResponseWriter, r *http.Request, serverName string, tools []*rpcTool) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	req := &rpcRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		writeRPC(w, &rpcResponse{
			JSONRPC: jsonRPCVersion,
			Error:   &rpcError{Code: codeParseError, Message: "cannot parse request"},
		})

		return
	}

	// A request without an id is a notification: it expects no answer.
	if len(req.ID) == 0 {
		w.WriteHeader(http.StatusAccepted)
		return
	}

	res := &rpcResponse{JSONRPC: jsonRPCVersion, ID: req.ID}

	switch req.Method {
	case methodInitialize:
		res.Result = initializeResult(req.Params, serverName)
	case methodToolsList:
		res.Result = map[string]any{"tools": toolSchemas(tools)}
	case methodToolsCall:
		res.Result = callTool(tools, req.Params, agentFromHeader(r))
	default:
		res.Error = &rpcError{Code: codeMethodNotFound, Message: "unknown method " + req.Method}
	}

	writeRPC(w, res)
}

func initializeResult(params json.RawMessage, serverName string) map[string]any {
	requested := struct {
		ProtocolVersion string `json:"protocolVersion"`
	}{}

	protocol := defaultProtocolVersion
	if err := json.Unmarshal(params, &requested); err == nil && requested.ProtocolVersion != "" {
		protocol = requested.ProtocolVersion
	}

	return map[string]any{
		"protocolVersion": protocol,
		"capabilities":    map[string]any{"tools": map[string]any{}},
		"serverInfo": map[string]any{
			"name":    serverName,
			"version": "1",
		},
	}
}

func toolSchemas(tools []*rpcTool) []any {
	schemas := make([]any, 0, len(tools))

	for _, tool := range tools {
		schemas = append(schemas, map[string]any{
			"name":        tool.name,
			"description": tool.description,
			"inputSchema": tool.input,
		})
	}

	return schemas
}

func callTool(tools []*rpcTool, params json.RawMessage, agentID uuid.UUID) *toolResult {
	call := struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}{}

	if err := json.Unmarshal(params, &call); err != nil {
		return errorResult("cannot parse tool arguments: " + err.Error())
	}

	for _, tool := range tools {
		if tool.name == call.Name {
			return tool.call(call.Arguments, agentID)
		}
	}

	return errorResult("unknown tool " + call.Name)
}

func objectSchema(properties map[string]any, required ...string) map[string]any {
	if required == nil {
		required = []string{}
	}

	return map[string]any{
		"type":       "object",
		"properties": properties,
		"required":   required,
	}
}

func stringProp(description string) map[string]any {
	return map[string]any{"type": "string", "description": description}
}

func errorResult(message string) *toolResult {
	return &toolResult{
		Content: []toolContent{{Type: contentText, Text: message}},
		IsError: true,
	}
}

func textResult(text string) *toolResult {
	return &toolResult{Content: []toolContent{{Type: contentText, Text: text}}}
}

func agentFromHeader(r *http.Request) uuid.UUID {
	parsed, err := uuid.Parse(r.Header.Get(constances.GatewayAgentHeader))
	if err != nil {
		return uuid.Nil
	}

	return parsed
}

func writeRPC(w http.ResponseWriter, res *rpcResponse) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
