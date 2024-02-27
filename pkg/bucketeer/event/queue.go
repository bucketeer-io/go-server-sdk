package event

import (
	"errors"
	"sync"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/model"
)

type queue struct {
	evtCh    chan *model.Event
	closedMu sync.RWMutex
	closed   bool
}

type queueConfig struct {
	capacity int
}

func newQueue(conf *queueConfig) *queue {
	return &queue{
		evtCh: make(chan *model.Event, conf.capacity),
	}
}

func (q *queue) push(evt *model.Event) error {
	q.closedMu.RLock()
	defer q.closedMu.RUnlock()
	if q.closed {
		return errors.New("queue is already closed")
	}
	select {
	case q.evtCh <- evt:
		return nil
	default:
		// When evtCh is full, discards evt and returns an error without waiting for a space in evtCh.
		// If we wait for the space, there is a risk that a user app will be a serious slowdown.
		return errors.New("queue is full")
	}
}

func (q *queue) eventCh() <-chan *model.Event {
	return q.evtCh
}

func (q *queue) close() {
	q.closedMu.Lock()
	defer q.closedMu.Unlock()
	close(q.evtCh)
	q.closed = true
}
