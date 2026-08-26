package manual_approval_broker

import (
	"testing"
	"time"

	"hexago/internal/helpers/enums"
	core_itf "hexago/internal/interface/core"
	input_itf "hexago/internal/interface/input"

	"github.com/google/uuid"
)

const testTimeout = 2 * time.Second

type raised struct {
	answer *core_itf.ApprovalAnswer
	err    error
}

func newBroker(t *testing.T) core_itf.ApprovalBroker {
	t.Helper()

	broker, err := InitV1(&input_itf.ApprovalBrokerConfig{Timeout: testTimeout})
	if err != nil {
		t.Fatalf("init approval broker: %v", err)
	}

	t.Cleanup(broker.Stop)

	return broker
}

func newRequest(agentID uuid.UUID, question string, multiSelect bool) *core_itf.ApprovalRequest {
	return &core_itf.ApprovalRequest{
		AgentID:     agentID,
		Kind:        enums.ApproveDecision,
		Question:    question,
		Detail:      "the caller has to pick before it can continue",
		MultiSelect: multiSelect,
		Options: []*core_itf.ApprovalOption{
			{ID: "sqlite", Label: "Store it in SQLite", Recommended: true},
			{ID: "files", Label: "Store it on disk"},
		},
	}
}

func request(t *testing.T, broker core_itf.ApprovalBroker, req *core_itf.ApprovalRequest) <-chan *raised {
	t.Helper()

	done := make(chan *raised, 1)

	go func() {
		answer, err := broker.Request(req)
		done <- &raised{answer: answer, err: err}
	}()

	return done
}

func waitPending(t *testing.T, broker core_itf.ApprovalBroker, want int) []*core_itf.ApprovalRequest {
	t.Helper()

	deadline := time.Now().Add(testTimeout)

	for {
		pending := broker.Pending()
		if len(pending) == want {
			return pending
		}

		if time.Now().After(deadline) {
			t.Fatalf("pending approvals = %d, want %d", len(pending), want)
		}

		time.Sleep(time.Millisecond)
	}
}

func waitRaised(t *testing.T, done <-chan *raised) *raised {
	t.Helper()

	select {
	case got := <-done:
		return got
	case <-time.After(testTimeout):
		t.Fatal("the blocked caller was never released")
		return nil
	}
}

func expectStillBlocked(t *testing.T, broker core_itf.ApprovalBroker, done <-chan *raised, err error) {
	t.Helper()

	if err == nil {
		t.Fatal("the answer was accepted, want a refusal")
	}

	select {
	case got := <-done:
		t.Fatalf("the caller was released by a refused answer: %+v", got)
	default:
	}

	waitPending(t, broker, 1)
}

func TestApproveWithOneOptionReleasesTheCaller(t *testing.T) {
	broker := newBroker(t)
	done := request(t, broker, newRequest(uuid.New(), "where do we store it?", false))

	pending := waitPending(t, broker, 1)

	if err := broker.Answer(&core_itf.ApprovalAnswer{
		RequestID: pending[0].ID,
		Approved:  true,
		OptionIDs: []string{"sqlite"},
		Guidance:  "cheapest to maintain",
	}); err != nil {
		t.Fatalf("answer: %v", err)
	}

	got := waitRaised(t, done)
	if got.err != nil {
		t.Fatalf("request: %v", got.err)
	}

	if !got.answer.Approved || len(got.answer.OptionIDs) != 1 || got.answer.OptionIDs[0] != "sqlite" {
		t.Fatalf("answer = %+v, want approved with option sqlite", got.answer)
	}

	waitPending(t, broker, 0)
}

func TestRejectWithNoOptionReleasesTheCaller(t *testing.T) {
	broker := newBroker(t)
	done := request(t, broker, newRequest(uuid.New(), "where do we store it?", false))

	pending := waitPending(t, broker, 1)

	if err := broker.Answer(&core_itf.ApprovalAnswer{
		RequestID: pending[0].ID,
		Approved:  false,
		Guidance:  "neither, keep it in memory",
	}); err != nil {
		t.Fatalf("answer: %v", err)
	}

	got := waitRaised(t, done)
	if got.err != nil {
		t.Fatalf("request: %v", got.err)
	}

	if got.answer.Approved {
		t.Fatalf("answer = %+v, want a rejection", got.answer)
	}

	if got.answer.Guidance != "neither, keep it in memory" {
		t.Fatalf("guidance = %q, want it carried to the caller", got.answer.Guidance)
	}
}

// Every refusal leaves the agent blocked on its question rather than releasing it with
// an answer the user did not give.
func TestAnAnswerTheRequestCannotTakeIsRefused(t *testing.T) {
	cases := map[string]struct {
		answer core_itf.ApprovalAnswer
	}{
		"a rejection carrying an option":     {answer: core_itf.ApprovalAnswer{Approved: false, OptionIDs: []string{"sqlite"}}},
		"several options on a single select": {answer: core_itf.ApprovalAnswer{Approved: true, OptionIDs: []string{"sqlite", "files"}}},
		"an option nobody offered":           {answer: core_itf.ApprovalAnswer{Approved: true, OptionIDs: []string{"redis"}}},
		"an approval picking nothing":        {answer: core_itf.ApprovalAnswer{Approved: true}},
	}

	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			broker := newBroker(t)
			done := request(t, broker, newRequest(uuid.New(), "where do we store it?", false))

			pending := waitPending(t, broker, 1)

			answer := testCase.answer
			answer.RequestID = pending[0].ID

			expectStillBlocked(t, broker, done, broker.Answer(&answer))
		})
	}
}

func TestSeveralOptionsOnAMultiSelectAreAccepted(t *testing.T) {
	broker := newBroker(t)
	done := request(t, broker, newRequest(uuid.New(), "where do we store it?", true))

	pending := waitPending(t, broker, 1)

	if err := broker.Answer(&core_itf.ApprovalAnswer{
		RequestID: pending[0].ID,
		Approved:  true,
		OptionIDs: []string{"sqlite", "files"},
	}); err != nil {
		t.Fatalf("answer: %v", err)
	}

	if got := waitRaised(t, done); got.err != nil {
		t.Fatalf("request: %v", got.err)
	}
}

func TestAnswerToAnUnknownRequestIsRefused(t *testing.T) {
	broker := newBroker(t)

	if err := broker.Answer(&core_itf.ApprovalAnswer{
		RequestID: uuid.New(),
		Approved:  false,
	}); err == nil {
		t.Fatal("an answer to a request nobody is waiting on was accepted")
	}
}

func TestPendingIsOrderedByRequestTime(t *testing.T) {
	broker := newBroker(t)
	questions := []string{"first fork", "second fork", "third fork"}

	for i, question := range questions {
		request(t, broker, newRequest(uuid.New(), question, false))
		waitPending(t, broker, i+1)
	}

	pending := waitPending(t, broker, len(questions))

	for i, want := range questions {
		if pending[i].Question != want {
			t.Fatalf("pending[%d] = %q, want %q", i, pending[i].Question, want)
		}

		if i > 0 && pending[i].RequestedAt.Before(pending[i-1].RequestedAt) {
			t.Fatalf("pending[%d] was requested before pending[%d]", i, i-1)
		}
	}
}

func TestAwaitingOnlyMatchesTheOwningAgent(t *testing.T) {
	broker := newBroker(t)
	agentID := uuid.New()

	request(t, broker, newRequest(agentID, "where do we store it?", false))
	waitPending(t, broker, 1)

	if !broker.Awaiting(agentID) {
		t.Fatal("the agent blocked on its own approval is not reported as awaiting")
	}

	if broker.Awaiting(uuid.Nil) {
		t.Fatal("the nil agent is reported as awaiting")
	}

	if broker.Awaiting(uuid.New()) {
		t.Fatal("an unrelated agent is reported as awaiting")
	}
}

func TestStopReleasesABlockedCaller(t *testing.T) {
	broker := newBroker(t)
	done := request(t, broker, newRequest(uuid.New(), "where do we store it?", false))

	waitPending(t, broker, 1)
	broker.Stop()

	got := waitRaised(t, done)
	if got.err == nil {
		t.Fatalf("request returned %+v on shutdown, want an error", got.answer)
	}
}

type flying bool

func (f flying) Autopilot() bool { return bool(f) }

func TestAutopilotAnswersWithTheRecommendedOptionWithoutRaisingIt(t *testing.T) {
	broker := newBroker(t)
	broker.TrackAutopilot(flying(true))

	answer, err := broker.Request(newRequest(uuid.New(), "where do we store it?", false))
	if err != nil {
		t.Fatalf("request under autopilot: %v", err)
	}

	if !answer.Approved || len(answer.OptionIDs) != 1 || answer.OptionIDs[0] != "sqlite" {
		t.Fatalf("answer = %+v, want the recommended option approved", answer)
	}

	if pending := broker.Pending(); len(pending) != 0 {
		t.Fatalf("autopilot left %d requests on the operator's screen", len(pending))
	}
}

func TestWithoutAutopilotTheRequestStillWaitsForAPerson(t *testing.T) {
	broker := newBroker(t)
	broker.TrackAutopilot(flying(false))

	done := request(t, broker, newRequest(uuid.New(), "where do we store it?", false))
	pending := waitPending(t, broker, 1)

	if err := broker.Answer(&core_itf.ApprovalAnswer{
		RequestID: pending[0].ID,
		Approved:  true,
		OptionIDs: []string{"files"},
	}); err != nil {
		t.Fatalf("answer: %v", err)
	}

	got := waitRaised(t, done)
	if got.err != nil {
		t.Fatalf("request: %v", got.err)
	}

	if got.answer.OptionIDs[0] != "files" {
		t.Fatalf("answer = %+v, want the option the operator picked rather than the recommendation", got.answer)
	}
}

// The recommendation is what autopilot answers with, so a request that names none or
// names several has no single answer to give and is refused at the door.
func TestARequestMustRecommendExactlyOneOption(t *testing.T) {
	cases := map[string][]*core_itf.ApprovalOption{
		"no recommendation": {
			{ID: "sqlite", Label: "Store it in SQLite"},
			{ID: "files", Label: "Store it on disk"},
		},
		"two recommendations": {
			{ID: "sqlite", Label: "Store it in SQLite", Recommended: true},
			{ID: "files", Label: "Store it on disk", Recommended: true},
		},
	}

	for name, options := range cases {
		t.Run(name, func(t *testing.T) {
			broker := newBroker(t)

			req := newRequest(uuid.New(), "where do we store it?", false)
			req.Options = options

			if _, err := broker.Request(req); err == nil {
				t.Fatal("the request was accepted, want a refusal")
			}
		})
	}
}
