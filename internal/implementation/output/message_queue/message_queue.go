package message_queue

import (
	"sync"

	"hexago/internal/helpers/custom_error"
	output_itf "hexago/internal/interface/output"

	"github.com/google/uuid"
)

const eventBufferSize = 16

type subscription struct {
	qID   uuid.UUID
	event string
}

type v1 struct {
	locker  sync.Mutex
	streams map[subscription]chan any
}

func InitV1() output_itf.MessageQ {
	return &v1{
		streams: map[subscription]chan any{},
	}
}

func (q *v1) Emit(qID uuid.UUID, event output_itf.MQEvent, data any) error {
	key, err := newSubscription(qID, event)
	if err != nil {
		return err
	}

	q.locker.Lock()
	defer q.locker.Unlock()

	stream, found := q.streams[key]
	if !found {
		return nil
	}

	select {
	case stream <- data:
		return nil
	default:
		return custom_error.Bypass("event %s of queue %v dropped, channel is full", key.event, qID)
	}
}

func (q *v1) Subscribe(qID uuid.UUID, event output_itf.MQEvent) (<-chan any, error) {
	key, err := newSubscription(qID, event)
	if err != nil {
		return nil, err
	}

	q.locker.Lock()
	defer q.locker.Unlock()

	if _, found := q.streams[key]; found {
		return nil, custom_error.Critical("event %s of queue %v is already subscribed", key.event, qID)
	}

	stream := make(chan any, eventBufferSize)
	q.streams[key] = stream

	return stream, nil
}

func (q *v1) Unsubscribe(qID uuid.UUID, event output_itf.MQEvent) error {
	key, err := newSubscription(qID, event)
	if err != nil {
		return err
	}

	q.locker.Lock()
	defer q.locker.Unlock()

	stream, found := q.streams[key]
	if !found {
		return custom_error.Critical("event %s of queue %v is not subscribed", key.event, qID)
	}

	delete(q.streams, key)
	close(stream)

	return nil
}

func newSubscription(qID uuid.UUID, event output_itf.MQEvent) (subscription, error) {
	if qID == uuid.Nil {
		return subscription{}, custom_error.Critical("queue id is empty")
	}

	if event == nil || event.Event() == "" {
		return subscription{}, custom_error.Critical("event of queue %v is empty", qID)
	}

	return subscription{
		qID:   qID,
		event: event.Event(),
	}, nil
}
