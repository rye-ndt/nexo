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

func TestUsageSumsEveryTokenBucket(t *testing.T) {
	proc := &openCodeProc{session: "ses_main", ctxWindow: ctxWindow}

	proc.trackUsage([]byte(`{"properties":{"info":{"sessionID":"ses_main","tokens":` +
		`{"input":10,"output":20,"reasoning":300,"cache":{"read":4000,"write":50000}}}}}`))

	usage := proc.snapshotUsage()
	if usage.Used != 54330 {
		t.Fatalf("used = %d, want 54330", usage.Used)
	}
	if usage.Total != ctxWindow {
		t.Fatalf("total = %d, want %d", usage.Total, ctxWindow)
	}
}

func TestChildSessionUsageLeavesTheNodeReadingAlone(t *testing.T) {
	proc := &openCodeProc{session: "ses_main", ctxWindow: ctxWindow}

	proc.trackUsage([]byte(messageUpdated("ses_main", 120_000)))
	proc.trackUsage([]byte(messageUpdated("ses_child", 3_000)))

	if used := proc.snapshotUsage().Used; used != 120_000 {
		t.Fatalf("used = %d, want this session's 120000", used)
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

func turn(id string, output int) string {
	return usageLine("ses_main", id, 0, output, 0, 0, 0)
}

func turnTokens(id string, in, out, cacheRead, cacheWrite int) string {
	return usageLine("ses_main", id, in, out, 0, cacheRead, cacheWrite)
}

func childTurnTokens(id string, in, out, cacheRead, cacheWrite int) string {
	return usageLine("ses_child", id, in, out, 0, cacheRead, cacheWrite)
}

// The prompt is re-read whole on every turn, so a billed total that counted the input
// side would grow with the conversation rather than with what the agent wrote.
func TestBilledCountsOutputAndReasoningOnly(t *testing.T) {
	proc := &openCodeProc{session: "ses_main", ctxWindow: ctxWindow}

	proc.trackUsage([]byte(`{"properties":{"info":{"id":"msg_01","sessionID":"ses_main","tokens":` +
		`{"input":10,"output":20,"reasoning":300,"cache":{"read":4000,"write":50000}}}}}`))

	usage := proc.snapshotUsage()
	if usage.Billed != 320 {
		t.Fatalf("billed = %d, want the output and reasoning tokens (320)", usage.Billed)
	}
	if usage.Used != 54330 {
		t.Fatalf("used = %d, want every bucket (54330)", usage.Used)
	}
}

func TestBilledAddsUpEveryTurn(t *testing.T) {
	proc := &openCodeProc{session: "ses_main", ctxWindow: ctxWindow}

	proc.trackUsage([]byte(turn("msg_01", 100)))
	proc.trackUsage([]byte(turn("msg_02", 150)))

	usage := proc.snapshotUsage()
	if usage.Billed != 250 {
		t.Fatalf("billed = %d, want 250", usage.Billed)
	}
	if usage.Used != 150 {
		t.Fatalf("used = %d, want the latest turn's 150", usage.Used)
	}
}

// The stream re-reports a message as it grows, so counting each event would bill
// the whole turn once per update.
func TestBilledCountsARestreamedTurnOnce(t *testing.T) {
	proc := &openCodeProc{session: "ses_main", ctxWindow: ctxWindow}

	proc.trackUsage([]byte(turn("msg_01", 100)))
	proc.trackUsage([]byte(turn("msg_01", 180)))

	if billed := proc.snapshotUsage().Billed; billed != 180 {
		t.Fatalf("billed = %d, want 180", billed)
	}
}

func TestBilledCountsARestreamedTurnOnceAcrossAnUnnamedEvent(t *testing.T) {
	proc := &openCodeProc{session: "ses_main", ctxWindow: ctxWindow}

	proc.trackUsage([]byte(turn("msg_01", 100)))
	proc.trackUsage([]byte(turn("", 50)))
	proc.trackUsage([]byte(turn("msg_01", 180)))

	if billed := proc.snapshotUsage().Billed; billed != 230 {
		t.Fatalf("billed = %d, want the named turn billed once plus the unnamed 50 (230)", billed)
	}
}

func TestBilledLeavesChildSessionsOut(t *testing.T) {
	proc := &openCodeProc{session: "ses_main", ctxWindow: ctxWindow}

	proc.trackUsage([]byte(turn("msg_01", 100)))
	proc.trackUsage([]byte(messageUpdated("ses_child", 3_000)))

	if billed := proc.snapshotUsage().Billed; billed != 100 {
		t.Fatalf("billed = %d, want this session's 100", billed)
	}
}

// A cache write is billed near the full input rate and a cache read at a tenth of
// it, so the two have to land in different counters to price a run at all.
func TestInputTakesCacheWritesAndCachedTakesCacheReads(t *testing.T) {
	proc := &openCodeProc{session: "ses_main", ctxWindow: ctxWindow}

	proc.trackUsage([]byte(turnTokens("msg_01", 10, 20, 300, 4000)))

	usage := proc.snapshotUsage()
	if usage.Input != 4010 {
		t.Fatalf("input = %d, want the fresh input plus the cache write (4010)", usage.Input)
	}
	if usage.Cached != 300 {
		t.Fatalf("cached = %d, want the cache read (300)", usage.Cached)
	}
	if usage.Used != 4330 {
		t.Fatalf("used = %d, want every bucket (4330)", usage.Used)
	}
}

func TestInputAndCachedAddUpEveryTurn(t *testing.T) {
	proc := &openCodeProc{session: "ses_main", ctxWindow: ctxWindow}

	proc.trackUsage([]byte(turnTokens("msg_01", 10, 5, 100, 1000)))
	proc.trackUsage([]byte(turnTokens("msg_02", 20, 5, 200, 2000)))

	usage := proc.snapshotUsage()
	if usage.Input != 3030 {
		t.Fatalf("input = %d, want 3030", usage.Input)
	}
	if usage.Cached != 300 {
		t.Fatalf("cached = %d, want 300", usage.Cached)
	}
}

func TestInputAndCachedCountARestreamedTurnOnce(t *testing.T) {
	proc := &openCodeProc{session: "ses_main", ctxWindow: ctxWindow}

	proc.trackUsage([]byte(turnTokens("msg_01", 10, 5, 100, 1000)))
	proc.trackUsage([]byte(turnTokens("msg_01", 10, 9, 100, 1200)))

	usage := proc.snapshotUsage()
	if usage.Input != 1210 {
		t.Fatalf("input = %d, want 1210", usage.Input)
	}
	if usage.Cached != 100 {
		t.Fatalf("cached = %d, want 100", usage.Cached)
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

func TestInputAndCachedLeaveChildSessionsOut(t *testing.T) {
	proc := &openCodeProc{session: "ses_main", ctxWindow: ctxWindow}

	proc.trackUsage([]byte(turnTokens("msg_01", 10, 5, 100, 1000)))
	proc.trackUsage([]byte(childTurnTokens("msg_child", 900, 900, 900, 900)))

	usage := proc.snapshotUsage()
	if usage.Input != 1010 {
		t.Fatalf("input = %d, want this session's 1010", usage.Input)
	}
	if usage.Cached != 100 {
		t.Fatalf("cached = %d, want this session's 100", usage.Cached)
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
