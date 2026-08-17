package mcp_proxy

import (
	"encoding/json"
	"testing"

	"hexago/internal/helpers"
	"hexago/internal/helpers/constances"
	input_itf "hexago/internal/interface/input"

	"github.com/google/uuid"
)

func TestChromeEnabledFollowsTheCredentialsRow(t *testing.T) {
	db := newFakeMCPStore()
	proxy := &v1{db: db}

	if proxy.chromeEnabled() {
		t.Fatal("chrome should be disabled without a credentials row")
	}

	if err := db.UpsertCredentials(&input_itf.MCPEntity{
		Name:              constances.ChromeLocalServer,
		EncryptedOAuthKey: "cipher",
		ExpiredAt:         helpers.NewUTC().Add(chromeEnabledTTL),
	}); err != nil {
		t.Fatalf("seed row: %v", err)
	}

	if !proxy.chromeEnabled() {
		t.Fatal("chrome should be enabled once the row exists")
	}

	if err := db.DeleteCredentials(constances.ChromeLocalServer); err != nil {
		t.Fatalf("delete row: %v", err)
	}

	if proxy.chromeEnabled() {
		t.Fatal("chrome should be disabled again once the row is gone")
	}
}

func TestChromeToolRefusesBeforeChromeIsEnabled(t *testing.T) {
	proxy := &v1{db: newFakeMCPStore()}

	called := false
	tool := proxy.chromeBrowserTool("chrome_list_tabs", "", objectSchema(map[string]any{}), func(*chromeArgs) *toolResult {
		called = true

		return textResult("ok")
	})

	result := tool.call(json.RawMessage(`{}`), uuid.Nil)

	if called {
		t.Fatal("the call ran while chrome was disabled")
	}

	if !result.IsError || result.Content[0].Text != chromeNotConnected {
		t.Fatalf("want the not-connected message, got %+v", result)
	}
}

func TestChromeKeyEventCarriesEnterText(t *testing.T) {
	event, bad := chromeKeyEvent("Enter")
	if bad != nil {
		t.Fatalf("enter: %s", bad.Content[0].Text)
	}

	if event["text"] != "\r" || event["windowsVirtualKeyCode"] != 13 {
		t.Fatalf("enter must carry a carriage return and its key code, got %+v", event)
	}

	printable, bad := chromeKeyEvent("a")
	if bad != nil {
		t.Fatalf("printable: %s", bad.Content[0].Text)
	}

	if printable["text"] != "a" {
		t.Fatalf("a printable key must be sent as text, got %+v", printable)
	}

	if _, bad := chromeKeyEvent("PageDownish"); bad == nil {
		t.Fatal("an unknown key must be reported back to the agent")
	}
}

func TestJSStringEscapesASelectorWithQuotes(t *testing.T) {
	selector := `a[title="one\two"]`

	decoded := ""
	if err := json.Unmarshal([]byte(jsString(selector)), &decoded); err != nil {
		t.Fatalf("the escaped selector is not a json literal: %v", err)
	}

	if decoded != selector {
		t.Fatalf("want %q, got %q", selector, decoded)
	}
}
