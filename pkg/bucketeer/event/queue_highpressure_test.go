package event

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/model"
)

const (
	// Note: ring buffer rounds to next power of 2, so 100000 becomes 131072
	highPressureQueueCapacity = 131072 // 2^17, a power of 2 for exact capacity
	highPressureEventCount    = 100000
)

// TestQueueHighPressure_FillAndDrain tests filling the queue to capacity
// and then draining all events, verifying FIFO order is maintained.
func TestQueueHighPressure_FillAndDrain(t *testing.T) {
	// Use a power of 2 to get exact capacity
	const capacity = 65536 // 2^16

	q := newQueue(&queueConfig{capacity: capacity})

	// Verify actual capacity matches
	assert.Equal(t, capacity, q.cap(), "capacity should be exact power of 2")

	// Fill queue to capacity
	for i := 0; i < capacity; i++ {
		evt := model.NewEvent(fmt.Sprintf("evt-%d", i), []byte("data"))
		err := q.push(evt)
		assert.NoError(t, err, "push should succeed for event %d", i)
	}

	// Verify queue is full
	assert.Equal(t, capacity, q.len())

	// Try to push one more - should fail
	extraEvt := model.NewEvent("extra", []byte("data"))
	err := q.push(extraEvt)
	assert.Error(t, err, "push should fail when queue is full")
	assert.Contains(t, err.Error(), "queue is full")

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

// TestQueueHighPressure_ConcurrentProducersConsumers tests the queue under
// high contention with multiple producers and consumers.
func TestQueueHighPressure_ConcurrentProducersConsumers(t *testing.T) {
	q := newQueue(&queueConfig{capacity: highPressureQueueCapacity})

	numProducers := 8
	numConsumers := 4
	eventsPerProducer := highPressureEventCount / numProducers

	var (
		totalPushed   int64
		totalPopped   int64
		totalDropped  int64
		producerWG    sync.WaitGroup
		consumerWG    sync.WaitGroup
		stopConsumers = make(chan struct{})
	)

	// Start consumers with polling (work-stealing pattern)
	for c := 0; c < numConsumers; c++ {
		consumerWG.Add(1)
		go func() {
			defer consumerWG.Done()
			localPopped := int64(0)
			ticker := time.NewTicker(time.Millisecond)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					// Drain all available events (work-stealing)
					for {
						_, ok := q.pop()
						if !ok {
							break
						}
						localPopped++
					}
				case <-stopConsumers:
					// Final drain
					for {
						_, ok := q.pop()
						if !ok {
							break
						}
						localPopped++
					}
					atomic.AddInt64(&totalPopped, localPopped)
					return
				}
			}
		}()
	}

	// Start producers
	for p := 0; p < numProducers; p++ {
		producerWG.Add(1)
		go func(producerID int) {
			defer producerWG.Done()
			localPushed := int64(0)
			localDropped := int64(0)
			for i := 0; i < eventsPerProducer; i++ {
				evt := model.NewEvent(fmt.Sprintf("p%d-evt-%d", producerID, i), []byte("data"))
				if err := q.push(evt); err != nil {
					localDropped++
				} else {
					localPushed++
				}
			}
			atomic.AddInt64(&totalPushed, localPushed)
			atomic.AddInt64(&totalDropped, localDropped)
		}(p)
	}

	// Wait for all producers to finish
	producerWG.Wait()

	// Give consumers time to drain remaining events
	time.Sleep(100 * time.Millisecond)

	// Stop consumers
	close(stopConsumers)
	consumerWG.Wait()

	// Verify accounting
	t.Logf("Total pushed: %d, Total popped: %d, Total dropped: %d",
		totalPushed, totalPopped, totalDropped)

	assert.Equal(t, totalPushed, totalPopped, "all pushed events should be consumed")
	assert.Equal(t, int64(highPressureEventCount), totalPushed+totalDropped,
		"pushed + dropped should equal total attempts")

	// Queue should be empty
	assert.Equal(t, 0, q.len())
}

// TestQueueHighPressure_BurstTraffic simulates burst traffic where events
// come in faster than they can be consumed, then verifies recovery.
func TestQueueHighPressure_BurstTraffic(t *testing.T) {
	// Use power of 2 for exact capacity (8192 = 2^13)
	queueSize := 8192
	burstSize := 12000 // More than queue can hold

	q := newQueue(&queueConfig{capacity: queueSize})

	// Verify actual capacity
	actualCap := q.cap()
	assert.Equal(t, queueSize, actualCap, "should be exact power of 2")

	var (
		pushed  int64
		dropped int64
	)

	// Burst: push events as fast as possible (no consumer running)
	for i := 0; i < burstSize; i++ {
		evt := model.NewEvent(fmt.Sprintf("burst-%d", i), []byte("data"))
		if err := q.push(evt); err != nil {
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

	// Now start draining and verify we get the FIRST events (not the dropped ones)
	// Events 0 to queueSize-1 should be in the queue (FIFO - oldest kept)
	for i := 0; i < queueSize; i++ {
		evt, ok := q.pop()
		assert.True(t, ok)
		expectedID := fmt.Sprintf("burst-%d", i)
		assert.Equal(t, expectedID, evt.ID, "should get events in FIFO order")
	}

	// Queue should be empty now
	assert.Equal(t, 0, q.len())

	// We can push again
	evt := model.NewEvent("after-burst", []byte("data"))
	err := q.push(evt)
	assert.NoError(t, err)
	assert.Equal(t, 1, q.len())
}

// TestQueueHighPressure_CloseUnderLoad tests closing the queue while
// producers are still pushing events.
func TestQueueHighPressure_CloseUnderLoad(t *testing.T) {
	q := newQueue(&queueConfig{capacity: highPressureQueueCapacity})

	var (
		pushAttempts   int64
		pushSuccesses  int64
		pushAfterClose int64
		producerWG     sync.WaitGroup
	)

	numProducers := 4

	// Start producers that will keep pushing
	for p := 0; p < numProducers; p++ {
		producerWG.Add(1)
		go func(producerID int) {
			defer producerWG.Done()
			for i := 0; i < 50000; i++ {
				atomic.AddInt64(&pushAttempts, 1)
				evt := model.NewEvent(fmt.Sprintf("p%d-evt-%d", producerID, i), []byte("data"))
				err := q.push(evt)
				if err == nil {
					atomic.AddInt64(&pushSuccesses, 1)
				} else if err.Error() == "queue is already closed" {
					atomic.AddInt64(&pushAfterClose, 1)
				}
			}
		}(p)
	}

	// Let producers run for a bit
	time.Sleep(10 * time.Millisecond)

	// Close the queue while producers are running
	q.close()

	// Wait for producers to finish
	producerWG.Wait()

	t.Logf("Push attempts: %d, Successes: %d, After close: %d",
		pushAttempts, pushSuccesses, pushAfterClose)

	// Verify some pushes succeeded before close
	assert.Greater(t, pushSuccesses, int64(0), "some pushes should succeed before close")

	// Verify pushes after close were rejected
	assert.Greater(t, pushAfterClose, int64(0), "some pushes should be rejected after close")

	// Verify queue is marked as closed
	assert.True(t, q.isClosed(), "queue should be closed")
}

// TestQueueHighPressure_ConsumerKeepsUp tests the scenario where
// consumers can keep up with producers (no drops expected).
func TestQueueHighPressure_ConsumerKeepsUp(t *testing.T) {
	q := newQueue(&queueConfig{capacity: 1000}) // Small buffer

	numEvents := 100000
	var (
		pushed     int64
		popped     int64
		consumerWG sync.WaitGroup
		done       = make(chan struct{})
	)

	// Start aggressive consumer with polling
	consumerWG.Add(1)
	go func() {
		defer consumerWG.Done()
		ticker := time.NewTicker(100 * time.Microsecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				for {
					_, ok := q.pop()
					if !ok {
						break
					}
					atomic.AddInt64(&popped, 1)
				}
			case <-done:
				// Final drain
				for {
					_, ok := q.pop()
					if !ok {
						break
					}
					atomic.AddInt64(&popped, 1)
				}
				return
			}
		}
	}()

	// Producer with small delays to let consumer keep up
	for i := 0; i < numEvents; i++ {
		evt := model.NewEvent(fmt.Sprintf("evt-%d", i), []byte("data"))
		// Retry a few times if queue is full
		for retry := 0; retry < 100; retry++ {
			if err := q.push(evt); err == nil {
				atomic.AddInt64(&pushed, 1)
				break
			}
			time.Sleep(time.Microsecond)
		}
	}

	// Signal consumer to stop and drain
	time.Sleep(50 * time.Millisecond)
	close(done)
	consumerWG.Wait()

	t.Logf("Pushed: %d, Popped: %d", pushed, popped)

	// All events should be processed
	assert.Equal(t, pushed, popped, "consumer should process all pushed events")
	assert.Equal(t, int64(numEvents), pushed, "all events should be pushed with retries")
}

// BenchmarkQueueHighPressure benchmarks queue operations under high pressure
func BenchmarkQueueHighPressure(b *testing.B) {
	q := newQueue(&queueConfig{capacity: 100000})

	// Start consumer with polling
	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(100 * time.Microsecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				for {
					if _, ok := q.pop(); !ok {
						break
					}
				}
			case <-done:
				return
			}
		}
	}()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		evt := model.NewEvent("bench", []byte("data"))
		for pb.Next() {
			q.push(evt)
		}
	})

	b.StopTimer()
	close(done)
}
