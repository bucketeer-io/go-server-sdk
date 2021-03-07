//go:generate mockgen -source=$GOFILE -package=$GOPACKAGE -destination=../../../test/mock/$GOPACKAGE/$GOFILE
package event

import (
	"context"
	"time"

	protofeature "github.com/ca-dp/bucketeer-go-server-sdk/proto/feature"
	protouser "github.com/ca-dp/bucketeer-go-server-sdk/proto/user"
)

// Processor defines the interface for processing events.
//
// When one of the Push method called, Processor pushes an event to the event queue.
//
// In the background, Processor pulls the events from the event queue,
// and sends them to the Bucketeer server using the Bucketeer API Client
// every time the specified time elapses or the specified capacity is exceeded.
type Processor interface {
	// PushEvaluationEvent pushes the evaluation event to the event queue.
	PushEvaluationEvent(ctx context.Context, user *protouser.User, evaluation *protofeature.Evaluation)

	// PushDefaultEvaluationEvent pushes the default evaluation event to the event queue.
	PushDefaultEvaluationEvent(ctx context.Context, user *protouser.User, featureID string)

	// PushEvaluationEvent pushes the goal event to the event queue.
	PushGoalEvent(ctx context.Context, user *protouser.User, goalID string, value float64)

	// PushGetEvaluationLatencyMetricsEvent pushes the get evaluation latency metrics event to the event queue.
	PushGetEvaluationLatencyMetricsEvent(ctx context.Context, duration time.Duration, tag, state string)

	// PushGetEvaluationSizeMetricsEvent pushes the get evaluation size metrics event to the event queue.
	PushGetEvaluationSizeMetricsEvent(ctx context.Context, sizeByte int, tag, state string)

	// PushTimeoutErrorCountMetricsEvent pushes the timeout error count metrics event to the event queue.
	PushTimeoutErrorCountMetricsEvent(ctx context.Context, tag string)

	// PushInternalErrorCountMetricsEvent pushes the internal error count metrics event to the event queue.
	PushInternalErrorCountMetricsEvent(ctx context.Context, tag string)

	// Close tears down all Processor activities, after ensuring that all events have been delivered.
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
	q := newQueue(&queueConfig{capacity: conf.EventQueueCapacity})
	return &processor{
		evtQueue: q,
	}
}

func (p *processor) PushEvaluationEvent(
	ctx context.Context,
	user *protouser.User,
	evaluation *protofeature.Evaluation,
) {
}

func (p *processor) PushDefaultEvaluationEvent(ctx context.Context, user *protouser.User, featureID string) {
}

func (p *processor) PushGoalEvent(ctx context.Context, user *protouser.User, goalID string, value float64) {
}

func (p *processor) PushGetEvaluationLatencyMetricsEvent(
	ctx context.Context,
	duration time.Duration,
	tag, state string,
) {
}

func (p *processor) PushGetEvaluationSizeMetricsEvent(ctx context.Context, sizeByte int, tag, state string) {
}

func (p *processor) PushTimeoutErrorCountMetricsEvent(ctx context.Context, tag string) {
}

func (p *processor) PushInternalErrorCountMetricsEvent(ctx context.Context, tag string) {
}

func (p *processor) Close() {
}
