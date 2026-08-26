package manual_approval_broker

import (
	"slices"
	"sync"
	"time"

	"hexago/internal/helpers"
	"hexago/internal/helpers/custom_error"
	"hexago/internal/helpers/enums"
	core_itf "hexago/internal/interface/core"
	input_itf "hexago/internal/interface/input"

	"github.com/google/uuid"
)

type pending struct {
	req    *core_itf.ApprovalRequest
	answer chan *core_itf.ApprovalAnswer
}

type v1 struct {
	locker    sync.Mutex
	cfg       *input_itf.ApprovalBrokerConfig
	waiting   map[uuid.UUID]*pending
	autopilot core_itf.AutopilotReader
	stop      chan struct{}
	stopped   bool
}

const autopilotGuidance = "Autopilot answered with the option you recommended. No person saw this question, " +
	"so treat it as your own call rather than an operator decision."

func InitV1(cfg *input_itf.ApprovalBrokerConfig) (core_itf.ApprovalBroker, error) {
	if err := helpers.ValidateStruct(cfg); err != nil {
		return nil, custom_error.Critical("invalid approval config: %v", err)
	}

	return &v1{
		cfg:     cfg,
		waiting: map[uuid.UUID]*pending{},
		stop:    make(chan struct{}),
	}, nil
}

func (b *v1) TrackAutopilot(reader core_itf.AutopilotReader) {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.autopilot = reader
}

func (b *v1) flying() bool {
	b.locker.Lock()
	defer b.locker.Unlock()

	return b.autopilot != nil && b.autopilot.Autopilot()
}

func (b *v1) Request(req *core_itf.ApprovalRequest) (*core_itf.ApprovalAnswer, error) {
	recommended, err := validRequest(req)
	if err != nil {
		return nil, err
	}

	uid, err := uuid.NewV7()
	if err != nil {
		return nil, custom_error.Critical("cannot create uuid: %v", err)
	}

	req.ID = uid
	req.RequestedAt = helpers.NewUTC()

	// Under autopilot nobody is watching, so the question is answered with the option the
	// agent itself recommended rather than parked until the request times out.
	if b.flying() {
		return &core_itf.ApprovalAnswer{
			RequestID: uid,
			Approved:  true,
			OptionIDs: []string{recommended.ID},
			Guidance:  autopilotGuidance,
		}, nil
	}

	waiter := &pending{
		req:    req,
		answer: make(chan *core_itf.ApprovalAnswer, 1),
	}

	b.locker.Lock()
	b.waiting[uid] = waiter
	b.locker.Unlock()

	defer b.forget(uid)

	select {
	case answer := <-waiter.answer:
		return answer, nil
	case <-time.After(b.cfg.Timeout):
		return nil, custom_error.Critical("approval %q timed out after %s with no answer", req.Question, b.cfg.Timeout)
	case <-b.stop:
		return nil, custom_error.Critical("approval %q was abandoned, the app is shutting down", req.Question)
	}
}

func validRequest(req *core_itf.ApprovalRequest) (*core_itf.ApprovalOption, error) {
	if req.Question == "" {
		return nil, custom_error.Critical("approval request has no question")
	}

	if len(req.Options) == 0 {
		return nil, custom_error.Critical("approval request %q has no options", req.Question)
	}

	recommended := []*core_itf.ApprovalOption{}

	for _, option := range req.Options {
		if option == nil || option.ID == "" || option.Label == "" {
			return nil, custom_error.Critical("approval request %q has an option without an id or label", req.Question)
		}

		if option.Recommended {
			recommended = append(recommended, option)
		}
	}

	if len(recommended) != 1 {
		return nil, custom_error.Critical(
			"approval request %q recommends %d of its %d options, it has to recommend exactly one",
			req.Question, len(recommended), len(req.Options))
	}

	if !req.Kind.Valid() {
		req.Kind = enums.ApproveDecision
	}

	return recommended[0], nil
}

func (b *v1) forget(requestID uuid.UUID) {
	b.locker.Lock()
	defer b.locker.Unlock()

	delete(b.waiting, requestID)
}

func (b *v1) Answer(answer *core_itf.ApprovalAnswer) error {
	b.locker.Lock()
	defer b.locker.Unlock()

	waiter, found := b.waiting[answer.RequestID]
	if !found {
		return custom_error.Critical("approval %v is not waiting for an answer", answer.RequestID)
	}

	if err := validAnswer(waiter.req, answer); err != nil {
		return err
	}

	delete(b.waiting, answer.RequestID)

	waiter.answer <- answer

	return nil
}

func validAnswer(req *core_itf.ApprovalRequest, answer *core_itf.ApprovalAnswer) error {
	if answer == nil {
		return custom_error.Critical("empty answer")
	}

	if !answer.Approved {
		if len(answer.OptionIDs) > 0 {
			return custom_error.Critical("approval %v was rejected, it cannot also pick an option", req.ID)
		}

		return nil
	}

	if len(answer.OptionIDs) == 0 {
		return custom_error.Critical("approval %v was approved with no option", req.ID)
	}

	if !req.MultiSelect && len(answer.OptionIDs) > 1 {
		return custom_error.Critical("approval %v accepts a single option only", req.ID)
	}

	for _, picked := range answer.OptionIDs {
		known := slices.ContainsFunc(req.Options, func(option *core_itf.ApprovalOption) bool {
			return option.ID == picked
		})

		if !known {
			return custom_error.Critical("approval %v has no option %q", req.ID, picked)
		}
	}

	return nil
}

func (b *v1) Pending() []*core_itf.ApprovalRequest {
	b.locker.Lock()

	requests := make([]*core_itf.ApprovalRequest, 0, len(b.waiting))
	for _, waiter := range b.waiting {
		requests = append(requests, snapshot(waiter.req))
	}

	b.locker.Unlock()

	slices.SortFunc(requests, func(a, b *core_itf.ApprovalRequest) int {
		return a.RequestedAt.Compare(b.RequestedAt)
	})

	return requests
}

func (b *v1) Awaiting(agentID uuid.UUID) bool {
	if agentID == uuid.Nil {
		return false
	}

	b.locker.Lock()
	defer b.locker.Unlock()

	for _, waiter := range b.waiting {
		if waiter.req.AgentID == agentID {
			return true
		}
	}

	return false
}

func (b *v1) Stop() {
	b.locker.Lock()
	defer b.locker.Unlock()

	if b.stopped {
		return
	}

	b.stopped = true
	close(b.stop)
}

func snapshot(req *core_itf.ApprovalRequest) *core_itf.ApprovalRequest {
	clone := *req
	clone.Options = make([]*core_itf.ApprovalOption, 0, len(req.Options))

	for _, option := range req.Options {
		copied := *option
		clone.Options = append(clone.Options, &copied)
	}

	return &clone
}
