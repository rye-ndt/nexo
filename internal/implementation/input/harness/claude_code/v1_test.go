package claude_code

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	core_itf "hexago/internal/interface/core"
)

const ctxWindow = 200_000

func assistantLine(parentToolUseID, id string, in, out, cacheRead, cacheWrite int) string {
	return fmt.Sprintf(
		`{"type":"assistant","parent_tool_use_id":%s,"message":{"id":%q,"usage":`+
			`{"input_tokens":%d,"output_tokens":%d,`+
			`"cache_read_input_tokens":%d,"cache_creation_input_tokens":%d}}}`,
		parentToolUseID, id, in, out, cacheRead, cacheWrite,
	)
}

func mainAssistant(used int) string {
	return assistantLine(`null`, "", 0, used, 0, 0)
}

func subAssistant(used int) string {
	return assistantLine(`"toolu_01"`, "", 0, used, 0, 0)
}

func turnTokens(id string, in, out, cacheRead, cacheWrite int) string {
	return assistantLine(`null`, id, in, out, cacheRead, cacheWrite)
}

func subTurnTokens(id string, in, out, cacheRead, cacheWrite int) string {
	return assistantLine(`"toolu_01"`, id, in, out, cacheRead, cacheWrite)
}

// Every counter is read off the same event: Used is the whole window, Billed is what
// the agent wrote, and a cache write is billed near the full input rate while a cache
// read is a tenth of it, so the two must land in different counters.
func TestOneTurnFillsEveryCounter(t *testing.T) {
	proc := &agentProc{ctxWindow: ctxWindow}

	proc.track([]byte(turnTokens("msg_01", 10, 20, 300, 4000)))

	usage := proc.snapshotUsage()
	if usage.Used != 4330 {
		t.Fatalf("used = %d, want every bucket (4330)", usage.Used)
	}
	if usage.Total != ctxWindow {
		t.Fatalf("total = %d, want %d", usage.Total, ctxWindow)
	}
	if usage.Billed != 20 {
		t.Fatalf("billed = %d, want the 20 output tokens", usage.Billed)
	}
	if usage.Input != 4010 {
		t.Fatalf("input = %d, want the fresh input plus the cache write (4010)", usage.Input)
	}
	if usage.Cached != 300 {
		t.Fatalf("cached = %d, want the cache read (300)", usage.Cached)
	}
}

func TestEveryCounterAddsUpAcrossTurns(t *testing.T) {
	proc := &agentProc{ctxWindow: ctxWindow}

	proc.track([]byte(turnTokens("msg_01", 10, 100, 100, 1000)))
	proc.track([]byte(turnTokens("msg_02", 20, 150, 200, 2000)))

	usage := proc.snapshotUsage()
	if usage.Billed != 250 {
		t.Fatalf("billed = %d, want 250", usage.Billed)
	}
	if usage.Input != 3030 {
		t.Fatalf("input = %d, want 3030", usage.Input)
	}
	if usage.Cached != 300 {
		t.Fatalf("cached = %d, want 300", usage.Cached)
	}
	if usage.Used != 2370 {
		t.Fatalf("used = %d, want the latest turn's window (2370)", usage.Used)
	}
}

// A subagent runs on a context of its own, so its turns belong to neither the node's
// reading nor its bill.
func TestSubagentTurnsAreLeftOutOfEveryCounter(t *testing.T) {
	proc := &agentProc{ctxWindow: ctxWindow}

	proc.track([]byte(turnTokens("msg_01", 10, 100, 100, 1000)))
	proc.track([]byte(subTurnTokens("msg_sub", 900, 900, 900, 900)))

	usage := proc.snapshotUsage()
	if usage.Billed != 100 {
		t.Fatalf("billed = %d, want the main agent's 100", usage.Billed)
	}
	if usage.Input != 1010 {
		t.Fatalf("input = %d, want the main agent's 1010", usage.Input)
	}
	if usage.Cached != 100 {
		t.Fatalf("cached = %d, want the main agent's 100", usage.Cached)
	}
	if usage.Used != 1210 {
		t.Fatalf("used = %d, want the main agent's window (1210)", usage.Used)
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

func TestEveryCounterRollsOverOnTheSameEvent(t *testing.T) {
	proc := &agentProc{ctxWindow: ctxWindow}

	proc.track([]byte(turnTokens("msg_01", 10, 5, 100, 1000)))
	proc.track([]byte(turnTokens("", 7, 3, 50, 0)))
	proc.track([]byte(turnTokens("msg_01", 10, 8, 100, 1200)))

	usage := proc.snapshotUsage()
	if usage.Billed != 11 {
		t.Fatalf("billed = %d, want the unnamed 3 plus the named turn's 8 (11)", usage.Billed)
	}
	if usage.Input != 1217 {
		t.Fatalf("input = %d, want the unnamed 7 plus the named turn's 1210 (1217)", usage.Input)
	}
	if usage.Cached != 150 {
		t.Fatalf("cached = %d, want the unnamed 50 plus the named turn's 100 (150)", usage.Cached)
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
