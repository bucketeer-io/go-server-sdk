package event

import (
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/model"
)

// queue is a ring buffer for event storage.
// NOT thread-safe - the processor is responsible for synchronization.
type queue struct {
	buffer   []*model.Event // Pre-allocated slots
	head     uint64         // Write position
	tail     uint64         // Read position
	mask     uint64         // capacity - 1 (for power-of-2 indexing)
	capacity uint64
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
// Returns false if the queue is full.
// NOT thread-safe - caller must hold lock.
func (q *queue) push(evt *model.Event) bool {
	if q.head-q.tail >= q.capacity {
		return false
	}

	q.buffer[q.head&q.mask] = evt
	q.head++
	return true
}

// pop removes and returns an event from the queue.
// Returns nil and false if the queue is empty.
// NOT thread-safe - caller must hold lock.
//
//nolint:unused
func (q *queue) pop() (*model.Event, bool) {
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

// popMany removes and returns events from the queue.
// - min: minimum number of events required (returns nil if queue has fewer)
// - max: maximum number of events to return
// Returns nil if queue has fewer than min events.
// NOT thread-safe - caller must hold lock.
func (q *queue) popMany(min, max int) []*model.Event {
	// Calculate how many events are available
	available := int(q.head - q.tail)
	if available < min {
		return nil
	}

	// Take the minimum of available and max
	count := available
	if max < count {
		count = max
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
//
//nolint:unused
func (q *queue) len() int {
	return int(q.head - q.tail)
}

// cap returns the capacity of the queue.
// This is safe to call without locking as capacity is immutable.
//
//nolint:unused
func (q *queue) cap() int {
	return int(q.capacity)
}
