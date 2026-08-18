package open_code

import (
	"fmt"
	"testing"
)

const ctxWindow = 200_000

func usageLine(sessionID, id string, in, out, reasoning, cacheRead, cacheWrite int) string {
	return fmt.Sprintf(
		`{"type":"message.updated","properties":{"info":{"id":%q,"sessionID":%q,"tokens":`+
			`{"input":%d,"output":%d,"reasoning":%d,"cache":{"read":%d,"write":%d}}}}}`,
		id, sessionID, in, out, reasoning, cacheRead, cacheWrite,
	)
}

func messageUpdated(sessionID string, output int) string {
	return usageLine(sessionID, "", 0, output, 0, 0, 0)
}

func turnTokens(id string, in, out, cacheRead, cacheWrite int) string {
	return usageLine("ses_main", id, in, out, 0, cacheRead, cacheWrite)
}

func childTurnTokens(id string, in, out, cacheRead, cacheWrite int) string {
	return usageLine("ses_child", id, in, out, 0, cacheRead, cacheWrite)
}

// Every counter is read off the same event: Used is the whole window, Billed is what
// the agent wrote and thought, and a cache write is billed near the full input rate
// while a cache read is a tenth of it, so the two must land in different counters.
func TestOneTurnFillsEveryCounter(t *testing.T) {
	proc := &openCodeProc{session: "ses_main", ctxWindow: ctxWindow}

	proc.trackUsage([]byte(usageLine("ses_main", "msg_01", 10, 20, 300, 4000, 50_000)))

	usage := proc.snapshotUsage()
	if usage.Used != 54330 {
		t.Fatalf("used = %d, want every bucket (54330)", usage.Used)
	}
	if usage.Total != ctxWindow {
		t.Fatalf("total = %d, want %d", usage.Total, ctxWindow)
	}
	if usage.Billed != 320 {
		t.Fatalf("billed = %d, want the output and reasoning tokens (320)", usage.Billed)
	}
	if usage.Input != 50_010 {
		t.Fatalf("input = %d, want the fresh input plus the cache write (50010)", usage.Input)
	}
	if usage.Cached != 4000 {
		t.Fatalf("cached = %d, want the cache read (4000)", usage.Cached)
	}
}

func TestEveryCounterAddsUpAcrossTurns(t *testing.T) {
	proc := &openCodeProc{session: "ses_main", ctxWindow: ctxWindow}

	proc.trackUsage([]byte(turnTokens("msg_01", 10, 100, 100, 1000)))
	proc.trackUsage([]byte(turnTokens("msg_02", 20, 150, 200, 2000)))

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

// A child session runs on a context of its own, so its turns belong to neither the
// node's reading nor its bill.
func TestChildSessionTurnsAreLeftOutOfEveryCounter(t *testing.T) {
	proc := &openCodeProc{session: "ses_main", ctxWindow: ctxWindow}

	proc.trackUsage([]byte(turnTokens("msg_01", 10, 100, 100, 1000)))
	proc.trackUsage([]byte(childTurnTokens("msg_child", 900, 900, 900, 900)))

	usage := proc.snapshotUsage()
	if usage.Billed != 100 {
		t.Fatalf("billed = %d, want this session's 100", usage.Billed)
	}
	if usage.Input != 1010 {
		t.Fatalf("input = %d, want this session's 1010", usage.Input)
	}
	if usage.Cached != 100 {
		t.Fatalf("cached = %d, want this session's 100", usage.Cached)
	}
	if usage.Used != 1210 {
		t.Fatalf("used = %d, want this session's window (1210)", usage.Used)
	}
}

func TestUsageFollowsTheLatestOwnSessionTurn(t *testing.T) {
	proc := &openCodeProc{session: "ses_main", ctxWindow: ctxWindow}

	proc.trackUsage([]byte(messageUpdated("ses_main", 120_000)))
	proc.trackUsage([]byte(messageUpdated("ses_child", 3_000)))
	proc.trackUsage([]byte(messageUpdated("ses_main", 140_000)))

	if used := proc.snapshotUsage().Used; used != 140_000 {
		t.Fatalf("used = %d, want 140000", used)
	}
}

func TestEveryCounterRollsOverOnTheSameEvent(t *testing.T) {
	proc := &openCodeProc{session: "ses_main", ctxWindow: ctxWindow}

	proc.trackUsage([]byte(turnTokens("msg_01", 10, 5, 100, 1000)))
	proc.trackUsage([]byte(turnTokens("", 7, 3, 50, 0)))
	proc.trackUsage([]byte(turnTokens("msg_01", 10, 8, 100, 1200)))

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

// An event without a sessionID cannot be attributed, so it is still counted
// rather than silently leaving the ring empty.
func TestUnattributedEventStillCounts(t *testing.T) {
	proc := &openCodeProc{session: "ses_main", ctxWindow: ctxWindow}

	proc.trackUsage([]byte(`{"properties":{"info":{"tokens":{"input":9000}}}}`))

	if used := proc.snapshotUsage().Used; used != 9000 {
		t.Fatalf("used = %d, want 9000", used)
	}
}
