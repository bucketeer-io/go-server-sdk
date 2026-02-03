package event

import (
	"errors"
	"sync"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/model"
)

// queue is a thread-safe event queue using a ring buffer for storage
// with a notification channel for consumer signaling.
// This design provides:
// - Reduced GC pressure through pre-allocated ring buffer slots
// - Lower push latency by avoiding channel send operations for data
// - Compatibility with Go's select statement for consumer notification
// - We use a single mutex instead of RWMutex for all operations to avoid double-locking overhead)
type queue struct {
	// Ring buffer fields
	buffer   []*model.Event // Pre-allocated slots
	head     uint64         // Write position
	tail     uint64         // Read position
	mask     uint64         // capacity - 1 (for power-of-2 indexing)
	capacity uint64

	// Signaling channels
	notify chan struct{} // Signal-only channel (capacity 1) for wake-up
	done   chan struct{} // Close signaling channel

	// Single mutex protects all state (buffer, head, tail, closed)
	mu     sync.Mutex
	closed bool
}

type queueConfig struct {
	capacity int
}

func newQueue(conf *queueConfig) *queue {
	cap64 := nextPowerOf2(uint64(conf.capacity))
	return &queue{
		buffer:   make([]*model.Event, cap64),
		capacity: cap64,
		mask:     cap64 - 1,
		notify:   make(chan struct{}, 1),
		done:     make(chan struct{}),
	}
}

// nextPowerOf2 returns the next power of 2 >= n.
// If n is already a power of 2, it returns n.
func nextPowerOf2(n uint64) uint64 {
	if n == 0 {
		return 1
	}
	// Check if n is already a power of 2
	if n&(n-1) == 0 {
		return n
	}
	// Find the position of the highest bit set
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n |= n >> 32
	return n + 1
}

// push adds an event to the queue.
// Returns an error if the queue is closed or full.
func (q *queue) push(evt *model.Event) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return errors.New("queue is already closed")
	}

	// Check if buffer is full
	if q.head-q.tail >= q.capacity {
		return errors.New("queue is full")
	}

	// Store the event at the head position
	q.buffer[q.head&q.mask] = evt
	q.head++

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
	q.mu.Lock()
	defer q.mu.Unlock()

	// Check if buffer is empty
	if q.head == q.tail {
		return nil, false
	}

	// Retrieve the event at the tail position
	evt := q.buffer[q.tail&q.mask]
	q.buffer[q.tail&q.mask] = nil // Clear the slot to allow GC
	q.tail++

	return evt, true
}

// len returns the current number of events in the queue.
// Used by tests to verify queue state.
//
//nolint:unused
func (q *queue) len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return int(q.head - q.tail)
}

// cap returns the capacity of the queue.
// This does not require a lock as capacity is immutable after initialization.
//
//nolint:unused
func (q *queue) cap() int {
	return int(q.capacity)
}

// close closes the queue and signals all consumers to stop.
func (q *queue) close() {
	q.mu.Lock()
	defer q.mu.Unlock()

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
