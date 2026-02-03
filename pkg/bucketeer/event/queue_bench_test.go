package event

import (
	"sync"
	"testing"
	"time"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/model"
)

// createTestEvent creates a test event for benchmarking
func createTestEvent(id string) *model.Event {
	return model.NewEvent(id, []byte(`{"test": "data"}`))
}

// BenchmarkQueuePush benchmarks the push operation on the queue
func BenchmarkQueuePush(b *testing.B) {
	q := newQueue(&queueConfig{capacity: 100000})
	evt := createTestEvent("test-id")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = q.push(evt)
	}
}

// BenchmarkQueuePushPop benchmarks push and pop operations
func BenchmarkQueuePushPop(b *testing.B) {
	q := newQueue(&queueConfig{capacity: 1024})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		evt := createTestEvent("test-id")
		_ = q.push(evt)
		q.pop()
	}
}

// BenchmarkQueueContention benchmarks the queue under contention with multiple producers
func BenchmarkQueueContention(b *testing.B) {
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
			q := newQueue(&queueConfig{capacity: 100000})

			// Start consumers with polling
			var consumerWG sync.WaitGroup
			stopConsumers := make(chan struct{})
			for c := 0; c < bm.consumers; c++ {
				consumerWG.Add(1)
				go func() {
					defer consumerWG.Done()
					ticker := time.NewTicker(100 * time.Microsecond)
					defer ticker.Stop()
					for {
						select {
						case <-ticker.C:
							// Drain all available events (work-stealing)
							for {
								if _, ok := q.pop(); !ok {
									break
								}
							}
						case <-stopConsumers:
							return
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
					evt := createTestEvent("test-id")
					for i := 0; i < opsPerProducer; i++ {
						_ = q.push(evt)
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

// BenchmarkQueueThroughput benchmarks overall throughput
func BenchmarkQueueThroughput(b *testing.B) {
	q := newQueue(&queueConfig{capacity: 10000})

	// Start a consumer goroutine with polling
	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(100 * time.Microsecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				// Drain all available events
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

	for i := 0; i < b.N; i++ {
		evt := createTestEvent("test-id")
		_ = q.push(evt)
	}

	b.StopTimer()
	close(done)
}
