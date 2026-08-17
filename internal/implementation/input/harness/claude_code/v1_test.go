package claude_code

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	core_itf "hexago/internal/interface/core"
)

const ctxWindow = 200_000

func assistantLine(parentToolUseID string, used int) string {
	return fmt.Sprintf(
		`{"type":"assistant","parent_tool_use_id":%s,"message":{"usage":`+
			`{"input_tokens":%d,"output_tokens":0,`+
			`"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}}`,
		parentToolUseID, used,
	)
}

func mainAssistant(used int) string {
	return assistantLine(`null`, used)
}

func subAssistant(used int) string {
	return assistantLine(`"toolu_01"`, used)
}

func TestUsageSumsEveryTokenBucket(t *testing.T) {
	proc := &agentProc{ctxWindow: ctxWindow}

	proc.track([]byte(`{"type":"assistant","message":{"usage":` +
		`{"input_tokens":10,"output_tokens":20,` +
		`"cache_read_input_tokens":300,"cache_creation_input_tokens":4000}}}`))

	usage := proc.snapshotUsage()
	if usage.Used != 4330 {
		t.Fatalf("used = %d, want 4330", usage.Used)
	}
	if usage.Total != ctxWindow {
		t.Fatalf("total = %d, want %d", usage.Total, ctxWindow)
	}
}

func TestSubagentUsageLeavesTheNodeReadingAlone(t *testing.T) {
	proc := &agentProc{ctxWindow: ctxWindow}

	proc.track([]byte(mainAssistant(120_000)))
	proc.track([]byte(subAssistant(3_000)))

	if used := proc.snapshotUsage().Used; used != 120_000 {
		t.Fatalf("used = %d, want the main agent's 120000", used)
	}
}

func TestUsageFollowsTheLatestMainAgentTurn(t *testing.T) {
	proc := &agentProc{ctxWindow: ctxWindow}

	proc.track([]byte(mainAssistant(120_000)))
	proc.track([]byte(subAssistant(3_000)))
	proc.track([]byte(mainAssistant(140_000)))

	if used := proc.snapshotUsage().Used; used != 140_000 {
		t.Fatalf("used = %d, want 140000", used)
	}
}

func turn(id string, used int) string {
	return fmt.Sprintf(
		`{"type":"assistant","message":{"id":%q,"usage":`+
			`{"input_tokens":%d,"output_tokens":0,`+
			`"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}}`,
		id, used,
	)
}

func TestBilledAddsUpEveryTurn(t *testing.T) {
	proc := &agentProc{ctxWindow: ctxWindow}

	proc.track([]byte(turn("msg_01", 100)))
	proc.track([]byte(turn("msg_02", 150)))

	usage := proc.snapshotUsage()
	if usage.Billed != 250 {
		t.Fatalf("billed = %d, want 250", usage.Billed)
	}
	if usage.Used != 150 {
		t.Fatalf("used = %d, want the latest turn's 150", usage.Used)
	}
}

// The same message is reported again as it streams, so counting each event would
// bill the whole turn once per chunk.
func TestBilledCountsARestreamedTurnOnce(t *testing.T) {
	proc := &agentProc{ctxWindow: ctxWindow}

	proc.track([]byte(turn("msg_01", 100)))
	proc.track([]byte(turn("msg_01", 180)))

	if billed := proc.snapshotUsage().Billed; billed != 180 {
		t.Fatalf("billed = %d, want 180", billed)
	}
}

func TestBilledCountsARestreamedTurnOnceAcrossAnUnnamedEvent(t *testing.T) {
	proc := &agentProc{ctxWindow: ctxWindow}

	proc.track([]byte(turn("msg_01", 100)))
	proc.track([]byte(turn("", 50)))
	proc.track([]byte(turn("msg_01", 180)))

	if billed := proc.snapshotUsage().Billed; billed != 230 {
		t.Fatalf("billed = %d, want the named turn billed once plus the unnamed 50 (230)", billed)
	}
}

func TestBilledLeavesSubagentTurnsOut(t *testing.T) {
	proc := &agentProc{ctxWindow: ctxWindow}

	proc.track([]byte(turn("msg_01", 100)))
	proc.track([]byte(subAssistant(3_000)))

	if billed := proc.snapshotUsage().Billed; billed != 100 {
		t.Fatalf("billed = %d, want the main agent's 100", billed)
	}
}

// An event without a message id cannot be matched against the turn before it, so
// it counts as a turn of its own rather than replacing the running total.
func TestBilledCountsUnidentifiedTurns(t *testing.T) {
	proc := &agentProc{ctxWindow: ctxWindow}

	proc.track([]byte(mainAssistant(100)))
	proc.track([]byte(mainAssistant(150)))

	if billed := proc.snapshotUsage().Billed; billed != 250 {
		t.Fatalf("billed = %d, want 250", billed)
	}
}

func TestNonAssistantEventsAreIgnored(t *testing.T) {
	proc := &agentProc{ctxWindow: ctxWindow}

	proc.track([]byte(mainAssistant(120_000)))
	proc.track([]byte(`{"type":"result","message":{"usage":{"input_tokens":7}}}`))
	proc.track([]byte(`not json`))

	if used := proc.snapshotUsage().Used; used != 120_000 {
		t.Fatalf("used = %d, want 120000", used)
	}
}

func TestNarrationTakesTheOpeningSentence(t *testing.T) {
	got := narration("Wiring the retry gate into the session manager.\n\nHere is the plan:\n- step one")

	if want := "Wiring the retry gate into the session manager."; got != want {
		t.Fatalf("narration = %q, want %q", got, want)
	}
}

func TestNarrationStripsMarkdownAndCode(t *testing.T) {
	cases := map[string]struct {
		text string
		want string
	}{
		"heading":  {"## Reading the config\n\nrest", "Reading the config"},
		"bullet":   {"- Checking the viper loader", "Checking the viper loader"},
		"numbered": {"1. Checking the viper loader", "Checking the viper loader"},
		"bold":     {"**Checking the viper loader**", "Checking the viper loader"},
		"quote":    {"> Checking the viper loader", "Checking the viper loader"},
		"fence":    {"Checking the loader\n```go\nfunc main() {}\n```", "Checking the loader"},
		"json":     {`{"tool": "Read", "path": "/a/b"}`, ""},
		"diff":     {"@@ -1,4 +1,9 @@\n+func narration()", ""},
		"fenceTop": {"```json\n{\"a\":1}\n```", ""},
		"boldBullet": {
			"- **Checking** the viper loader",
			"Checking the viper loader",
		},
		"boldNumbered": {
			"1. **Checking** the viper loader",
			"Checking the viper loader",
		},
		"boldQuote": {
			"> **Checking** the viper loader",
			"Checking the viper loader",
		},
		"proseAfterFence": {
			"```go\nfunc main() {}\n```\nChecking the viper loader",
			"Checking the viper loader",
		},
		"skipsToProse": {
			"| col | col |\nChecking the viper loader",
			"Checking the viper loader",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := narration(tc.text); got != tc.want {
				t.Fatalf("narration = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestNarrationClipsALongLine(t *testing.T) {
	long := "Checking " + strings.Repeat("the viper loader ", 20)

	got := narration(long)
	if len([]rune(got)) != 141 {
		t.Fatalf("narration is %d runes, want 140 plus the ellipsis", len([]rune(got)))
	}
	if !strings.HasSuffix(got, "…") {
		t.Fatalf("narration = %q, want a trailing ellipsis", got)
	}
}

func TestToolUseStaysAutoNarrated(t *testing.T) {
	block := &claudeBlock{Type: blockToolUse, Name: toolRead}
	block.Input.FilePath = "/repo/internal/implementation/input/config/viper.go"

	if got, want := digest(block), "Reading config/viper.go"; got != want {
		t.Fatalf("digest = %q, want %q", got, want)
	}
}

func TestAllowedToolsCoversEveryGatewayServer(t *testing.T) {
	gateway := &core_itf.MCPGateway{Servers: []core_itf.MCPGatewayServer{
		{Name: "harness"},
		{Name: "figma"},
		{Name: "chrome-devtools"},
		{Name: "atlassian", AuthKeyName: "ATLASSIAN_OAUTH_SECRET"},
	}}

	allowed := strings.Split(allowedTools(gateway), ",")

	for _, want := range []string{
		toolRead, toolBash,
		"mcp__harness", "mcp__figma", "mcp__chrome-devtools", "mcp__atlassian",
	} {
		if !slices.Contains(allowed, want) {
			t.Fatalf("%q is missing from the allowlist: %v", want, allowed)
		}
	}

	if slices.Contains(allowed, toolAskUser) {
		t.Fatal("the ask-user tool must stay out of the allowlist")
	}
}

func TestAllowedToolsWithoutAGatewayKeepsTheBaseTools(t *testing.T) {
	allowed := strings.Split(allowedTools(nil), ",")

	if len(allowed) != len(baseTools) {
		t.Fatalf("want only the base tools, got %v", allowed)
	}
}
