package event

import (
	"errors"
	"sync"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/model"
)

// queue is a hybrid event queue that combines a ring buffer for storage
// with a notification channel for consumer signaling.
// This design provides:
// - Reduced GC pressure through pre-allocated ring buffer slots
// - Lower push latency by avoiding channel send operations for data
// - Compatibility with Go's select statement for consumer notification
type queue struct {
	ring     *ringBuffer   // Pre-allocated ring buffer for event storage
	notify   chan struct{} // Signal-only channel (capacity 1) for wake-up
	done     chan struct{} // Close signaling channel
	closedMu sync.RWMutex
	closed   bool
}

type queueConfig struct {
	capacity int
}

func newQueue(conf *queueConfig) *queue {
	return &queue{
		ring:   newRingBuffer(conf.capacity),
		notify: make(chan struct{}, 1),
		done:   make(chan struct{}),
	}
}

// push adds an event to the queue.
// Returns an error if the queue is closed or full.
func (q *queue) push(evt *model.Event) error {
	q.closedMu.RLock()
	defer q.closedMu.RUnlock()

	if q.closed {
		return errors.New("queue is already closed")
	}

	if !q.ring.Push(evt) {
		// When ring buffer is full, discard evt and return an error without waiting.
		// If we wait for space, there is a risk that a user app will experience serious slowdown.
		return errors.New("queue is full")
	}

	// Non-blocking signal to notify consumers that data is available.
	// We use a buffered channel of size 1, so at most one signal is queued.
	// If a signal is already pending, this select will fall through to default,
	// which is fine because the consumer will drain all available events.
	select {
	case q.notify <- struct{}{}:
	default:
	}

	return nil
}

// notifyCh returns the notification channel for consumers to select on.
// Consumers should drain the ring buffer when they receive a notification.
func (q *queue) notifyCh() <-chan struct{} {
	return q.notify
}

// doneCh returns the done channel for consumers to detect queue closure.
func (q *queue) doneCh() <-chan struct{} {
	return q.done
}

// pop removes and returns an event from the queue.
// Returns nil and false if the queue is empty.
func (q *queue) pop() (*model.Event, bool) {
	return q.ring.Pop()
}

// len returns the current number of events in the queue.
// Used by tests to verify queue state.
//
//nolint:unused
func (q *queue) len() int {
	return q.ring.Len()
}

// close closes the queue and signals all consumers to stop.
func (q *queue) close() {
	q.closedMu.Lock()
	defer q.closedMu.Unlock()

	if q.closed {
		return
	}

	q.closed = true
	close(q.done)

	// Send final notification to wake up any waiting consumers
	select {
	case q.notify <- struct{}{}:
	default:
	}
}
