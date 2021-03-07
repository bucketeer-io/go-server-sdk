package event

import (
	"errors"

	protoevent "github.com/ca-dp/bucketeer-go-server-sdk/proto/event/client"
)

type queue struct {
	evtCh chan *protoevent.Event
}

type queueConfig struct {
	capacity int
}

func newQueue(conf *queueConfig) *queue {
	return &queue{
		evtCh: make(chan *protoevent.Event, conf.capacity),
	}
}

// TODO: remove nolint comment
// nolint
func (q *queue) push(evt *protoevent.Event) error {
	select {
	case q.evtCh <- evt:
		return nil
	default:
		return errors.New("queue is full")
	}
}

// TODO: remove nolint comment
// nolint
func (q *queue) eventCh() <-chan *protoevent.Event {
	return q.evtCh
}
