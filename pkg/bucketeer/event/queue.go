package event

import (
	"errors"
	"sync"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/model"
)

// queue is a thread-safe event queue using a ring buffer for storage.
// This design provides:
// - Reduced GC pressure through pre-allocated ring buffer slots
// - Lower push latency (no channel operations on push)
// - Single mutex for all operations (no double-locking overhead)
// - Workers poll the queue using adaptive ticker intervals
type queue struct {
	// Ring buffer fields
	buffer   []*model.Event // Pre-allocated slots
	head     uint64         // Write position
	tail     uint64         // Read position
	mask     uint64         // capacity - 1 (for power-of-2 indexing)
	capacity uint64

	// Close signaling channel
	done chan struct{}

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

	return nil
}

// doneCh returns the done channel for consumers to detect queue closure.
func (q *queue) doneCh() <-chan struct{} {
	return q.done
}

// pop removes and returns an event from the queue.
// Returns nil and false if the queue is empty.
//
//nolint:unused
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

// popMany removes and returns up to n events from the queue.
// Returns an empty slice if the queue is empty.
// This is more efficient than calling pop() n times as it only acquires the lock once.
func (q *queue) popMany(n int) []*model.Event {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Calculate how many events are available
	available := int(q.head - q.tail)
	if available == 0 {
		return nil
	}

	// Take the minimum of available and requested
	count := available
	if n < count {
		count = n
	}

	// Pre-allocate result slice
	result := make([]*model.Event, count)

	// Copy events to result
	for i := 0; i < count; i++ {
		idx := q.tail & q.mask
		result[i] = q.buffer[idx]
		q.buffer[idx] = nil // Clear the slot to allow GC
		q.tail++
	}

	return result
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

// isClosed returns whether the queue is closed.
// Used by tests to verify queue state.
//
//nolint:unused
func (q *queue) isClosed() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.closed
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
}
