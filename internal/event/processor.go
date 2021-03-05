//go:generate mockgen -source=$GOFILE -package=$GOPACKAGE -destination=../../test/mock/$GOPACKAGE/$GOFILE
package event

import (
	"context"

	protoevent "github.com/ca-dp/bucketeer-go-server-sdk/proto/event/client"
)

// Processor defines the interface for processing events.
//
// When one of the Push method called, Processor pushes the given event to the event queue.
//
// In the background, Processor pulls the events from the event queue,
// and sends them to the Bucketeer server using the Bucketeer API Client
// every time the specified time elapses or the specified capacity is exceeded.
type Processor interface {
	// PushEvaluationEvent pushes the given evaluation event to the event queue.
	PushEvaluationEvent(ctx context.Context, evt protoevent.EvaluationEvent)

	// PushEvaluationEvent pushes the given goal event to the event queue.
	PushGoalEvent(ctx context.Context, evt protoevent.GoalEvent)

	// PushEvaluationEvent pushes the given metrics event to the event queue.
	PushMetricsEvent(ctx context.Context, evt protoevent.MetricsEvent)

	// Close shuts down all Processor activity, after ensuring that all events have been delivered.
	Close()
}

// TODO: implement below

type processor struct {
	evtQueue *queue
}

type ProcessorConfig struct {
	EventQueueCapacity int
}

func NewProcessor(conf *ProcessorConfig) Processor {
	return &processor{
		evtQueue: newQueue(conf.EventQueueCapacity),
	}
}

func (p *processor) PushEvaluationEvent(ctx context.Context, evt protoevent.EvaluationEvent) {}

func (p *processor) PushGoalEvent(ctx context.Context, evt protoevent.GoalEvent) {}

func (p *processor) PushMetricsEvent(ctx context.Context, evt protoevent.MetricsEvent) {}

func (p *processor) Close() {}
