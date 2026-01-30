package event

import (
	"fmt"
	"sync"
	"sync/atomic"
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

func TestRingBuffer_Basic(t *testing.T) {
	rb := newRingBuffer(4)

	// Test empty buffer
	assert.True(t, rb.IsEmpty())
	assert.False(t, rb.IsFull())
	assert.Equal(t, 0, rb.Len())
	assert.Equal(t, 4, rb.Cap())

	// Pop from empty should return false
	evt, ok := rb.Pop()
	assert.Nil(t, evt)
	assert.False(t, ok)

	// Push events
	evt1 := model.NewEvent("1", []byte("data1"))
	evt2 := model.NewEvent("2", []byte("data2"))
	evt3 := model.NewEvent("3", []byte("data3"))
	evt4 := model.NewEvent("4", []byte("data4"))

	assert.True(t, rb.Push(evt1))
	assert.Equal(t, 1, rb.Len())
	assert.False(t, rb.IsEmpty())

	assert.True(t, rb.Push(evt2))
	assert.True(t, rb.Push(evt3))
	assert.True(t, rb.Push(evt4))
	assert.True(t, rb.IsFull())

	// Push when full should fail
	evt5 := model.NewEvent("5", []byte("data5"))
	assert.False(t, rb.Push(evt5))

	// Pop events in FIFO order
	poppedEvt, ok := rb.Pop()
	assert.True(t, ok)
	assert.Equal(t, "1", poppedEvt.ID)

	poppedEvt, ok = rb.Pop()
	assert.True(t, ok)
	assert.Equal(t, "2", poppedEvt.ID)

	// Now we can push again
	assert.True(t, rb.Push(evt5))

	poppedEvt, ok = rb.Pop()
	assert.True(t, ok)
	assert.Equal(t, "3", poppedEvt.ID)

	poppedEvt, ok = rb.Pop()
	assert.True(t, ok)
	assert.Equal(t, "4", poppedEvt.ID)

	poppedEvt, ok = rb.Pop()
	assert.True(t, ok)
	assert.Equal(t, "5", poppedEvt.ID)

	assert.True(t, rb.IsEmpty())
}

func TestRingBuffer_CapacityRounding(t *testing.T) {
	// Test that non-power-of-2 capacities are rounded up
	rb := newRingBuffer(5)
	assert.Equal(t, 8, rb.Cap())

	rb = newRingBuffer(100)
	assert.Equal(t, 128, rb.Cap())

	rb = newRingBuffer(1000)
	assert.Equal(t, 1024, rb.Cap())
}

func TestRingBuffer_DrainTo(t *testing.T) {
	rb := newRingBuffer(8)

	// Push 5 events
	for i := 1; i <= 5; i++ {
		rb.Push(model.NewEvent(string(rune('0'+i)), []byte("data")))
	}
	assert.Equal(t, 5, rb.Len())

	// Drain max 3 events
	events := make([]*model.Event, 0, 10)
	events, count := rb.DrainTo(events, 3)
	assert.Equal(t, 3, count)
	assert.Len(t, events, 3)
	assert.Equal(t, 2, rb.Len())

	// Drain remaining (ask for 10, should only get 2)
	events = make([]*model.Event, 0, 10)
	events, count = rb.DrainTo(events, 10)
	assert.Equal(t, 2, count)
	assert.Len(t, events, 2)
	assert.True(t, rb.IsEmpty())
}

func TestRingBuffer_Concurrent(t *testing.T) {
	rb := newRingBuffer(1024)
	numOps := 10000

	var wg sync.WaitGroup

	// Producer
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < numOps; i++ {
			evt := model.NewEvent("test", []byte("data"))
			for !rb.Push(evt) {
				// Spin until we can push
			}
		}
	}()

	// Consumer
	consumed := 0
	wg.Add(1)
	go func() {
		defer wg.Done()
		for consumed < numOps {
			if _, ok := rb.Pop(); ok {
				consumed++
			}
		}
	}()

	wg.Wait()
	assert.Equal(t, numOps, consumed)
	assert.True(t, rb.IsEmpty())
}

func TestRingBuffer_MultipleProducersConsumers(t *testing.T) {
	rb := newRingBuffer(4096)
	numProducers := 4
	numConsumers := 2
	opsPerProducer := 5000

	var producerWG sync.WaitGroup
	var consumerWG sync.WaitGroup

	produced := int64(0)
	consumed := int64(0)

	// Start consumers
	stopConsumers := make(chan struct{})
	for i := 0; i < numConsumers; i++ {
		consumerWG.Add(1)
		go func() {
			defer consumerWG.Done()
			localConsumed := int64(0)
			for {
				select {
				case <-stopConsumers:
					// Drain remaining
					for {
						if _, ok := rb.Pop(); ok {
							localConsumed++
						} else {
							break
						}
					}
					atomicAdd(&consumed, localConsumed)
					return
				default:
					if _, ok := rb.Pop(); ok {
						localConsumed++
					}
				}
			}
		}()
	}

	// Start producers
	for i := 0; i < numProducers; i++ {
		producerWG.Add(1)
		go func() {
			defer producerWG.Done()
			localProduced := int64(0)
			for j := 0; j < opsPerProducer; j++ {
				evt := model.NewEvent("test", []byte("data"))
				for !rb.Push(evt) {
					// Spin until we can push
				}
				localProduced++
			}
			atomicAdd(&produced, localProduced)
		}()
	}

	producerWG.Wait()
	close(stopConsumers)
	consumerWG.Wait()

	totalProduced := atomicLoad(&produced)
	totalConsumed := atomicLoad(&consumed)

	assert.Equal(t, int64(numProducers*opsPerProducer), totalProduced)
	assert.Equal(t, totalProduced, totalConsumed)
}

func atomicAdd(addr *int64, delta int64) {
	for {
		old := atomic.LoadInt64(addr)
		if atomic.CompareAndSwapInt64(addr, old, old+delta) {
			return
		}
	}
}

func atomicLoad(addr *int64) int64 {
	return atomic.LoadInt64(addr)
}

// Benchmarks for ring buffer
func BenchmarkRingBufferPush(b *testing.B) {
	rb := newRingBuffer(100000)
	evt := model.NewEvent("test-id", []byte(`{"test": "data"}`))

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		rb.Push(evt)
	}
}

func BenchmarkRingBufferPushPop(b *testing.B) {
	rb := newRingBuffer(1024)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		evt := model.NewEvent("test-id", []byte(`{"test": "data"}`))
		rb.Push(evt)
		rb.Pop()
	}
}

func BenchmarkRingBufferContention(b *testing.B) {
	benchmarks := []struct {
		name      string
		producers int
		consumers int
	}{
		{"1P1C", 1, 1},
		{"4P1C", 4, 1},
		{"4P4C", 4, 4},
		{"8P4C", 8, 4},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			rb := newRingBuffer(100000)

			// Start consumers
			var consumerWG sync.WaitGroup
			stopConsumers := make(chan struct{})
			for c := 0; c < bm.consumers; c++ {
				consumerWG.Add(1)
				go func() {
					defer consumerWG.Done()
					for {
						select {
						case <-stopConsumers:
							return
						default:
							rb.Pop()
						}
					}
				}()
			}

			// Divide work among producers
			opsPerProducer := b.N / bm.producers
			var producerWG sync.WaitGroup

			b.ResetTimer()
			b.ReportAllocs()

			for p := 0; p < bm.producers; p++ {
				producerWG.Add(1)
				go func() {
					defer producerWG.Done()
					evt := model.NewEvent("test-id", []byte(`{"test": "data"}`))
					for i := 0; i < opsPerProducer; i++ {
						rb.Push(evt)
					}
				}()
			}

			producerWG.Wait()
			b.StopTimer()

			close(stopConsumers)
			consumerWG.Wait()
		})
	}
}

func BenchmarkRingBufferDrainTo(b *testing.B) {
	rb := newRingBuffer(1024)
	evt := model.NewEvent("test-id", []byte(`{"test": "data"}`))

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Push 100 events
		for j := 0; j < 100; j++ {
			rb.Push(evt)
		}
		// Drain all
		events := make([]*model.Event, 0, 100)
		rb.DrainTo(events, 100)
	}
}
