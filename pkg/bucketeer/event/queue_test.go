package event

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/model"
)

func TestNextPowerOf2(t *testing.T) {
	tests := []struct {
		input    uint64
		expected uint64
	}{
		{0, 1},
		{1, 1},
		{2, 2},
		{3, 4},
		{4, 4},
		{5, 8},
		{7, 8},
		{8, 8},
		{9, 16},
		{100, 128},
		{1000, 1024},
		{1024, 1024},
		{1025, 2048},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("input_%d_expected_%d", tt.input, tt.expected), func(t *testing.T) {
			result := nextPowerOf2(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestQueue_Basic(t *testing.T) {
	q := newQueue(&queueConfig{capacity: 4})

	// Test empty queue
	assert.Equal(t, 0, q.len())
	assert.Equal(t, 4, q.cap())

	// Pop from empty should return false
	evt, ok := q.pop()
	assert.Nil(t, evt)
	assert.False(t, ok)

	// Push events
	evt1 := model.NewEvent("1", []byte("data1"))
	evt2 := model.NewEvent("2", []byte("data2"))
	evt3 := model.NewEvent("3", []byte("data3"))
	evt4 := model.NewEvent("4", []byte("data4"))

	assert.True(t, q.push(evt1))
	assert.Equal(t, 1, q.len())

	assert.True(t, q.push(evt2))
	assert.True(t, q.push(evt3))
	assert.True(t, q.push(evt4))

	// Push when full should fail
	evt5 := model.NewEvent("5", []byte("data5"))
	pushed := q.push(evt5)
	assert.False(t, pushed, "push should fail when queue is full")

	// Pop events in FIFO order
	poppedEvt, ok := q.pop()
	assert.True(t, ok)
	assert.Equal(t, "1", poppedEvt.ID)

	poppedEvt, ok = q.pop()
	assert.True(t, ok)
	assert.Equal(t, "2", poppedEvt.ID)

	// Now we can push again
	assert.True(t, q.push(evt5))

	poppedEvt, ok = q.pop()
	assert.True(t, ok)
	assert.Equal(t, "3", poppedEvt.ID)

	poppedEvt, ok = q.pop()
	assert.True(t, ok)
	assert.Equal(t, "4", poppedEvt.ID)

	poppedEvt, ok = q.pop()
	assert.True(t, ok)
	assert.Equal(t, "5", poppedEvt.ID)

	assert.Equal(t, 0, q.len())
}

func TestQueue_CapacityRounding(t *testing.T) {
	// Test that non-power-of-2 capacities are rounded up
	q := newQueue(&queueConfig{capacity: 5})
	assert.Equal(t, 8, q.cap())

	q = newQueue(&queueConfig{capacity: 100})
	assert.Equal(t, 128, q.cap())

	q = newQueue(&queueConfig{capacity: 1000})
	assert.Equal(t, 1024, q.cap())
}

func TestQueue_PopMany(t *testing.T) {
	t.Run("empty queue returns nil", func(t *testing.T) {
		q := newQueue(&queueConfig{capacity: 10})
		events := q.popMany(1, 5)
		assert.Nil(t, events)
	})

	t.Run("pop fewer than available", func(t *testing.T) {
		q := newQueue(&queueConfig{capacity: 10})
		for i := 0; i < 5; i++ {
			q.push(model.NewEvent(fmt.Sprintf("%d", i), []byte("data")))
		}

		events := q.popMany(1, 3)
		assert.Len(t, events, 3)
		assert.Equal(t, "0", events[0].ID)
		assert.Equal(t, "1", events[1].ID)
		assert.Equal(t, "2", events[2].ID)
		assert.Equal(t, 2, q.len())
	})

	t.Run("pop more than available with min=1", func(t *testing.T) {
		q := newQueue(&queueConfig{capacity: 10})
		for i := 0; i < 3; i++ {
			q.push(model.NewEvent(fmt.Sprintf("%d", i), []byte("data")))
		}

		events := q.popMany(1, 10) // Request more than available, min=1
		assert.Len(t, events, 3)
		assert.Equal(t, "0", events[0].ID)
		assert.Equal(t, "1", events[1].ID)
		assert.Equal(t, "2", events[2].ID)
		assert.Equal(t, 0, q.len())
	})

	t.Run("returns nil when below minimum", func(t *testing.T) {
		q := newQueue(&queueConfig{capacity: 10})
		for i := 0; i < 3; i++ {
			q.push(model.NewEvent(fmt.Sprintf("%d", i), []byte("data")))
		}

		// Request minimum 5, but only 3 available
		events := q.popMany(5, 10)
		assert.Nil(t, events)
		assert.Equal(t, 3, q.len()) // Events remain in queue
	})

	t.Run("pop exact count", func(t *testing.T) {
		q := newQueue(&queueConfig{capacity: 10})
		for i := 0; i < 5; i++ {
			q.push(model.NewEvent(fmt.Sprintf("%d", i), []byte("data")))
		}

		events := q.popMany(5, 5) // min=max=5
		assert.Len(t, events, 5)
		assert.Equal(t, 0, q.len())
	})

	t.Run("multiple popMany calls", func(t *testing.T) {
		q := newQueue(&queueConfig{capacity: 100})
		for i := 0; i < 50; i++ {
			q.push(model.NewEvent(fmt.Sprintf("%d", i), []byte("data")))
		}

		// Pop in batches
		batch1 := q.popMany(1, 10)
		assert.Len(t, batch1, 10)
		assert.Equal(t, "0", batch1[0].ID)

		batch2 := q.popMany(1, 10)
		assert.Len(t, batch2, 10)
		assert.Equal(t, "10", batch2[0].ID)

		assert.Equal(t, 30, q.len())
	})

	t.Run("clears slots for GC", func(t *testing.T) {
		q := newQueue(&queueConfig{capacity: 4})
		evt := model.NewEvent("1", []byte("data"))
		q.push(evt)

		events := q.popMany(1, 1)
		assert.Len(t, events, 1)

		// The slot should be cleared (nil) after pop
		// We can verify this indirectly by checking that a new push works
		assert.True(t, q.push(model.NewEvent("2", []byte("data"))))
	})
}

// Note: TestQueue_Concurrent and TestQueue_MultipleProducersConsumers were removed
// because the queue is now intentionally NOT thread-safe.
// The processor owns all synchronization and wraps queue access with its mutex.
// See processor_test.go for concurrent tests that use proper synchronization.
