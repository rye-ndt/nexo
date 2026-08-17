package mcp_proxy_helpers

import (
	"encoding/json"
	"net/http"
	neturl "net/url"
	"strings"
	"time"

	"hexago/internal/helpers/custom_error"
	"hexago/internal/helpers/enums"
	input_itf "hexago/internal/interface/input"

	"github.com/gorilla/websocket"
)

const (
	cdpPageType     = "page"
	cdpDevToolsPage = "devtools://"
)

type CDPTarget struct {
	ID                   string `json:"id"`
	Type                 string `json:"type"`
	Title                string `json:"title"`
	URL                  string `json:"url"`
	WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
}

type CDPSession struct {
	conn    *websocket.Conn
	nextID  int
	timeout time.Duration
}

type cdpError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type cdpResponse struct {
	ID     int             `json:"id"`
	Result json.RawMessage `json:"result"`
	Error  *cdpError       `json:"error"`
}

type cdpRemoteObject struct {
	Type        string          `json:"type"`
	Value       json.RawMessage `json:"value"`
	Description string          `json:"description"`
}

type cdpExceptionDetails struct {
	Text      string          `json:"text"`
	Exception cdpRemoteObject `json:"exception"`
}

type cdpEvaluation struct {
	Result           cdpRemoteObject      `json:"result"`
	ExceptionDetails *cdpExceptionDetails `json:"exceptionDetails"`
}

func CDPTargets(httpCli input_itf.HttpCli, endpoint string) ([]*CDPTarget, error) {
	targets := []*CDPTarget{}
	if err := httpCli.GetJSON(endpoint+"/json/list", &targets); err != nil {
		return nil, custom_error.TypedCritical(enums.ErrChromeNotConnected, "cannot list the chrome tabs: %v", err)
	}

	pages := make([]*CDPTarget, 0, len(targets))

	for _, target := range targets {
		if target.Type != cdpPageType || target.WebSocketDebuggerURL == "" {
			continue
		}

		if strings.HasPrefix(target.URL, cdpDevToolsPage) {
			continue
		}

		pages = append(pages, target)
	}

	return pages, nil
}

func CDPNewTarget(httpCli input_itf.HttpCli, endpoint, url string) (*CDPTarget, error) {
	res, err := httpCli.Stream(&input_itf.HttpRequest{
		Method: http.MethodPut,
		URL:    endpoint + "/json/new?" + neturl.QueryEscape(url),
	})
	if err != nil {
		return nil, custom_error.TypedCritical(enums.ErrChromeNotConnected, "cannot open a chrome tab at %q: %v", url, err)
	}

	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		return nil, custom_error.TypedCritical(enums.ErrMcpRequestFailed, "chrome refused to open %q with status %d", url, res.StatusCode)
	}

	target := &CDPTarget{}
	if err := json.NewDecoder(res.Body).Decode(target); err != nil {
		return nil, custom_error.TypedCritical(enums.ErrMcpRequestFailed, "cannot read the chrome tab opened at %q: %v", url, err)
	}

	return target, nil
}

func CDPCloseTarget(httpCli input_itf.HttpCli, endpoint, targetID string) error {
	res, err := httpCli.Stream(&input_itf.HttpRequest{
		Method: http.MethodGet,
		URL:    endpoint + "/json/close/" + targetID,
	})
	if err != nil {
		return custom_error.TypedCritical(enums.ErrChromeNotConnected, "cannot close the chrome tab %q: %v", targetID, err)
	}

	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		return custom_error.TypedCritical(enums.ErrMcpRequestFailed, "chrome refused to close the tab %q with status %d", targetID, res.StatusCode)
	}

	return nil
}

func DialCDP(target *CDPTarget, timeout time.Duration) (*CDPSession, error) {
	dialer := &websocket.Dialer{HandshakeTimeout: timeout}

	conn, _, err := dialer.Dial(target.WebSocketDebuggerURL, nil)
	if err != nil {
		return nil, custom_error.TypedCritical(enums.ErrChromeNotConnected, "cannot attach to the chrome tab %q: %v", target.ID, err)
	}

	return &CDPSession{conn: conn, nextID: 1, timeout: timeout}, nil
}

func (s *CDPSession) Call(method string, params map[string]any) (json.RawMessage, error) {
	id := s.nextID
	s.nextID++

	command := map[string]any{"id": id, "method": method}
	if params != nil {
		command["params"] = params
	}

	deadline := time.Now().Add(s.timeout)

	if err := s.conn.SetWriteDeadline(deadline); err != nil {
		return nil, custom_error.TypedCritical(enums.ErrChromeNotConnected, "cannot send %q to chrome: %v", method, err)
	}

	if err := s.conn.WriteJSON(command); err != nil {
		return nil, custom_error.TypedCritical(enums.ErrChromeNotConnected, "cannot send %q to chrome: %v", method, err)
	}

	for {
		if err := s.conn.SetReadDeadline(deadline); err != nil {
			return nil, custom_error.TypedCritical(enums.ErrChromeNotConnected, "cannot wait for the answer to %q: %v", method, err)
		}

		answer := &cdpResponse{}
		if err := s.conn.ReadJSON(answer); err != nil {
			return nil, custom_error.TypedCritical(enums.ErrMcpRequestFailed, "chrome did not answer %q: %v", method, err)
		}

		if answer.ID != id {
			continue
		}

		if answer.Error != nil {
			return nil, custom_error.TypedCritical(enums.ErrMcpRequestFailed, "chrome rejected %q: %s", method, answer.Error.Message)
		}

		return answer.Result, nil
	}
}

func (s *CDPSession) Eval(expression string) (json.RawMessage, error) {
	raw, err := s.Call("Runtime.evaluate", map[string]any{
		"expression":    expression,
		"returnByValue": true,
		"awaitPromise":  true,
	})
	if err != nil {
		return nil, err
	}

	evaluation := &cdpEvaluation{}
	if err := json.Unmarshal(raw, evaluation); err != nil {
		return nil, custom_error.TypedCritical(enums.ErrMcpRequestFailed, "cannot read what the page returned: %v", err)
	}

	if evaluation.ExceptionDetails != nil {
		return nil, custom_error.TypedCritical(enums.ErrMcpRequestFailed, "the page threw: %s", evaluation.ExceptionDetails.message())
	}

	if len(evaluation.Result.Value) > 0 {
		return evaluation.Result.Value, nil
	}

	return json.Marshal(evaluation.Result.Description)
}

func (s *CDPSession) Close() error {
	return s.conn.Close()
}

func (d *cdpExceptionDetails) message() string {
	if d.Exception.Description != "" {
		return d.Exception.Description
	}

	if len(d.Exception.Value) > 0 {
		return string(d.Exception.Value)
	}

	return d.Text
}
