package event

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/model"
)

// TestQueueHighPressure_FillAndDrain tests filling the queue to capacity
// and then draining all events, verifying FIFO order is maintained.
// Note: This is single-threaded since the queue is intentionally NOT thread-safe.
// The processor owns all synchronization.
func TestQueueHighPressure_FillAndDrain(t *testing.T) {
	// Use a power of 2 to get exact capacity
	const capacity = 65536 // 2^16

	q := newQueue(&queueConfig{capacity: capacity})

	// Verify actual capacity matches
	assert.Equal(t, capacity, q.cap(), "capacity should be exact power of 2")

	// Fill queue to capacity
	for i := 0; i < capacity; i++ {
		evt := model.NewEvent(fmt.Sprintf("evt-%d", i), []byte("data"))
		ok := q.push(evt)
		assert.True(t, ok, "push should succeed for event %d", i)
	}

	// Verify queue is full
	assert.Equal(t, capacity, q.len())

	// Try to push one more - should fail
	extraEvt := model.NewEvent("extra", []byte("data"))
	ok := q.push(extraEvt)
	assert.False(t, ok, "push should fail when queue is full")

	// Drain all events and verify FIFO order
	for i := 0; i < capacity; i++ {
		evt, ok := q.pop()
		assert.True(t, ok, "pop should succeed for event %d", i)
		assert.Equal(t, fmt.Sprintf("evt-%d", i), evt.ID, "events should be in FIFO order")
	}

	// Verify queue is empty
	assert.Equal(t, 0, q.len())
	evt, ok := q.pop()
	assert.False(t, ok, "pop should return false when queue is empty")
	assert.Nil(t, evt)
}

// TestQueueHighPressure_BurstTraffic simulates burst traffic where events
// arrive faster than they can be consumed, testing queue full behavior.
// Note: Single-threaded test since queue is not thread-safe.
func TestQueueHighPressure_BurstTraffic(t *testing.T) {
	const queueSize = 8192 // 2^13
	const burstSize = 12000
	q := newQueue(&queueConfig{capacity: queueSize})

	var (
		pushed  int64
		dropped int64
	)

	// Burst: push events as fast as possible
	for i := 0; i < burstSize; i++ {
		evt := model.NewEvent(fmt.Sprintf("burst-%d", i), []byte("data"))
		if !q.push(evt) {
			dropped++
		} else {
			pushed++
		}
	}

	t.Logf("Burst complete - Pushed: %d, Dropped: %d, Queue len: %d",
		pushed, dropped, q.len())

	// Verify queue is at capacity
	assert.Equal(t, queueSize, q.len())
	assert.Equal(t, int64(queueSize), pushed)
	assert.Equal(t, int64(burstSize-queueSize), dropped)

	// Now drain and verify we get the FIRST events (not the dropped ones)
	// Events 0 to queueSize-1 should be in the queue (FIFO - oldest kept)
	for i := 0; i < queueSize; i++ {
		evt, ok := q.pop()
		assert.True(t, ok)
		assert.Equal(t, fmt.Sprintf("burst-%d", i), evt.ID)
	}

	// Queue should be empty now
	assert.Equal(t, 0, q.len())

	// We can push again
	evt := model.NewEvent("after-burst", []byte("data"))
	ok := q.push(evt)
	assert.True(t, ok)
	assert.Equal(t, 1, q.len())
}

// Note: Concurrent tests (TestQueueHighPressure_ConcurrentProducersConsumers,
// TestQueueHighPressure_CloseUnderLoad, TestQueueHighPressure_ConsumerKeepsUp)
// were removed because the queue is now intentionally NOT thread-safe.
// The processor owns all synchronization and wraps queue access with its mutex.
// See processor_test.go for concurrent tests that use proper synchronization.
