package event

import (
	"errors"

	protoevent "github.com/ca-dp/bucketeer-go-server-sdk/proto/event/client"
)

type queue struct {
	evtCh chan *protoevent.Event
}

func newQueue(capacity int) *queue {
	return &queue{
		evtCh: make(chan *protoevent.Event, capacity),
	}
}

// TODO: remove nolint comment
// nolint
func (q *queue) push(evt *protoevent.Event) error {
	select {
	case q.evtCh <- evt:
		return nil
	default:
		return errors.New("event: queue is full")
	}
}

// TODO: remove nolint comment
// nolint
func (q *queue) eventCh() <-chan *protoevent.Event {
	return q.evtCh
}
