package open_code

import (
	"fmt"
	"testing"
)

const ctxWindow = 200_000

func messageUpdated(sessionID string, input int) string {
	return fmt.Sprintf(
		`{"type":"message.updated","properties":{"info":{"sessionID":%q,"tokens":`+
			`{"input":%d,"output":0,"reasoning":0,"cache":{"read":0,"write":0}}}}}`,
		sessionID, input,
	)
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

func turn(id string, input int) string {
	return fmt.Sprintf(
		`{"type":"message.updated","properties":{"info":{"id":%q,"sessionID":"ses_main",`+
			`"tokens":{"input":%d,"output":0,"reasoning":0,"cache":{"read":0,"write":0}}}}}`,
		id, input,
	)
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

// An event without a sessionID cannot be attributed, so it is still counted
// rather than silently leaving the ring empty.
func TestUnattributedEventStillCounts(t *testing.T) {
	proc := &openCodeProc{session: "ses_main", ctxWindow: ctxWindow}

	proc.trackUsage([]byte(`{"properties":{"info":{"tokens":{"input":9000}}}}`))

	if used := proc.snapshotUsage().Used; used != 9000 {
		t.Fatalf("used = %d, want 9000", used)
	}
}
