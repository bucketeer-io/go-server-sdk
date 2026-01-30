package event

import (
	"sync"
	"sync/atomic"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/model"
)

// ringBuffer is a thread-safe ring buffer for events with pre-allocated slots.
// It uses atomic operations for head/tail pointers and a mutex for slot access.
// The capacity is always rounded up to the next power of 2 for efficient indexing.
type ringBuffer struct {
	buffer   []*model.Event // Pre-allocated slots
	head     uint64         // Write position (atomic)
	tail     uint64         // Read position (atomic)
	mask     uint64         // capacity - 1 (for power-of-2 indexing)
	capacity uint64
	mu       sync.Mutex // Protects slot access during push/pop
}

// newRingBuffer creates a new ring buffer with the given capacity.
// The capacity is rounded up to the next power of 2.
func newRingBuffer(capacity int) *ringBuffer {
	cap64 := nextPowerOf2(uint64(capacity))
	return &ringBuffer{
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

// Push adds an event to the ring buffer.
// Returns false if the buffer is full (non-blocking).
func (rb *ringBuffer) Push(evt *model.Event) bool {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	head := atomic.LoadUint64(&rb.head)
	tail := atomic.LoadUint64(&rb.tail)

	// Check if buffer is full
	if head-tail >= rb.capacity {
		return false
	}

	// Store the event at the head position
	rb.buffer[head&rb.mask] = evt
	atomic.AddUint64(&rb.head, 1)

	return true
}

// Pop removes and returns an event from the ring buffer.
// Returns nil and false if the buffer is empty.
func (rb *ringBuffer) Pop() (*model.Event, bool) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	head := atomic.LoadUint64(&rb.head)
	tail := atomic.LoadUint64(&rb.tail)

	// Check if buffer is empty
	if head == tail {
		return nil, false
	}

	// Retrieve the event at the tail position
	evt := rb.buffer[tail&rb.mask]
	rb.buffer[tail&rb.mask] = nil // Clear the slot to allow GC
	atomic.AddUint64(&rb.tail, 1)

	return evt, true
}

// Len returns the current number of elements in the ring buffer.
func (rb *ringBuffer) Len() int {
	head := atomic.LoadUint64(&rb.head)
	tail := atomic.LoadUint64(&rb.tail)
	return int(head - tail)
}

// Cap returns the capacity of the ring buffer.
func (rb *ringBuffer) Cap() int {
	return int(rb.capacity)
}

// IsEmpty returns true if the ring buffer is empty.
func (rb *ringBuffer) IsEmpty() bool {
	return rb.Len() == 0
}

// IsFull returns true if the ring buffer is full.
func (rb *ringBuffer) IsFull() bool {
	return rb.Len() >= int(rb.capacity)
}

// DrainTo drains up to maxItems from the ring buffer into the provided slice.
// Returns the slice with appended events and the count of events drained.
func (rb *ringBuffer) DrainTo(events []*model.Event, maxItems int) ([]*model.Event, int) {
	count := 0
	for count < maxItems {
		evt, ok := rb.Pop()
		if !ok {
			break
		}
		events = append(events, evt)
		count++
	}
	return events, count
}
