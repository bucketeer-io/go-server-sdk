//go:generate mockgen -source=$GOFILE -package=$GOPACKAGE -destination=../../../test/mock/$GOPACKAGE/$GOFILE
package event

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/api"
	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/log"
	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/uuid"
	protoevent "github.com/ca-dp/bucketeer-go-server-sdk/proto/event/client"
	protofeature "github.com/ca-dp/bucketeer-go-server-sdk/proto/feature"
	protogateway "github.com/ca-dp/bucketeer-go-server-sdk/proto/gateway"
	protouser "github.com/ca-dp/bucketeer-go-server-sdk/proto/user"
)

// Processor defines the interface for processing events.
//
// When one of the Push method called, Processor pushes an event to the in-memory event queue.
//
// In the background, Processor pulls the events from the event queue,
// and sends them to the Bucketeer service using the Bucketeer API Client
// every time the specified time elapses or the specified capacity is exceeded.
type Processor interface {
	// PushEvaluationEvent pushes the evaluation event to the queue.
	PushEvaluationEvent(ctx context.Context, user *protouser.User, evaluation *protofeature.Evaluation)

	// PushDefaultEvaluationEvent pushes the default evaluation event to the queue.
	PushDefaultEvaluationEvent(ctx context.Context, user *protouser.User, featureID string)

	// PushEvaluationEvent pushes the goal event to the queue.
	PushGoalEvent(ctx context.Context, user *protouser.User, goalID string, value float64)

	// PushGetEvaluationLatencyMetricsEvent pushes the get evaluation latency metrics event to the queue.
	PushGetEvaluationLatencyMetricsEvent(ctx context.Context, duration time.Duration, tag string)

	// PushGetEvaluationSizeMetricsEvent pushes the get evaluation size metrics event to the queue.
	PushGetEvaluationSizeMetricsEvent(ctx context.Context, sizeByte int, tag string)

	// PushTimeoutErrorCountMetricsEvent pushes the timeout error count metrics event to the queue.
	PushTimeoutErrorCountMetricsEvent(ctx context.Context, tag string)

	// PushInternalErrorCountMetricsEvent pushes the internal error count metrics event to the queue.
	PushInternalErrorCountMetricsEvent(ctx context.Context, tag string)

	// Close tears down all Processor activities, after ensuring that all events have been delivered.
	Close(ctx context.Context) error
}

type processor struct {
	evtQueue        *queue
	numFlushWorkers int
	flushInterval   time.Duration
	flushSize       int
	flushTimeout    time.Duration
	apiClient       api.Client
	loggers         *log.Loggers
	closeCh         chan struct{}
	workerWG        sync.WaitGroup
}

// ProcessorConfig is the config for Processor.
type ProcessorConfig struct {
	// QueueCapacity is a capacity of the event queue.
	//
	// The queue buffers events up to the capacity in memory before processing.
	// If the capacity is exceeded, events will be discarded.
	QueueCapacity int

	// NumFlushWorkers is a number of workers to flush events.
	NumFlushWorkers int

	// FlushInterval is a interval of flushing events.
	//
	// Each worker sends the events to Bucketeer service every time eventFlushInterval elapses or
	// its buffer exceeds eventFlushSize.
	FlushInterval time.Duration

	// FlushSize is a size of the buffer for each worker.
	//
	// Each worker sends the events to Bucketeer service every time EventFlushInterval elapses or
	// its buffer exceeds EventFlushSize is exceeded.
	FlushSize int

	// APIClient is the client for Bucketeer service.
	APIClient api.Client

	// Loggers is the Bucketeer SDK Loggers.
	Loggers *log.Loggers
}

// NewProcessor creates a new Processor.
func NewProcessor(conf *ProcessorConfig) Processor {
	q := newQueue(&queueConfig{
		capacity: conf.QueueCapacity,
	})
	flushTimeout := 10 * time.Second
	p := &processor{
		evtQueue:        q,
		numFlushWorkers: conf.NumFlushWorkers,
		flushInterval:   conf.FlushInterval,
		flushSize:       conf.FlushSize,
		flushTimeout:    flushTimeout,
		apiClient:       conf.APIClient,
		loggers:         conf.Loggers,
		closeCh:         make(chan struct{}),
	}
	go p.startWorkers()
	return p
}

func (p *processor) PushEvaluationEvent(
	ctx context.Context,
	user *protouser.User,
	evaluation *protofeature.Evaluation,
) {
	evaluationEvt := &protoevent.EvaluationEvent{
		Timestamp:      time.Now().Unix(),
		FeatureId:      evaluation.FeatureId,
		FeatureVersion: evaluation.FeatureVersion,
		UserId:         user.Id,
		VariationId:    evaluation.VariationId,
		User:           user,
		Reason:         evaluation.Reason,
	}
	anyEvaluationEvt, err := ptypes.MarshalAny(evaluationEvt)
	if err != nil {
		p.loggers.Errorf(
			"bucketeer/event: PushEvaluationEvent failed (err: %v, userID: %s, featureID: %s)",
			err,
			user.Id,
			evaluation.FeatureId,
		)
		return
	}
	if err := p.pushEvent(anyEvaluationEvt); err != nil {
		p.loggers.Errorf(
			"bucketeer/event: PushEvaluationEvent failed (err: %v, userID: %s, featureID: %s)",
			err,
			user.Id,
			evaluation.FeatureId,
		)
		return
	}
}

func (p *processor) PushDefaultEvaluationEvent(ctx context.Context, user *protouser.User, featureID string) {
	evaluationEvt := &protoevent.EvaluationEvent{
		Timestamp:      time.Now().Unix(),
		FeatureId:      featureID,
		FeatureVersion: 0,
		UserId:         user.Id,
		VariationId:    "",
		User:           user,
		Reason:         &protofeature.Reason{Type: protofeature.Reason_CLIENT},
	}
	anyEvaluationEvt, err := ptypes.MarshalAny(evaluationEvt)
	if err != nil {
		p.loggers.Errorf(
			"bucketeer/event: PushDefaultEvaluationEvent failed (err: %v, userID: %s, featureID: %s)",
			err,
			user.Id,
			featureID,
		)
		return
	}
	if err := p.pushEvent(anyEvaluationEvt); err != nil {
		p.loggers.Errorf(
			"bucketeer/event: PushDefaultEvaluationEvent failed (err: %v, userID: %s, featureID: %s)",
			err,
			user.Id,
			featureID,
		)
		return
	}
}

func (p *processor) PushGoalEvent(ctx context.Context, user *protouser.User, goalID string, value float64) {
	goalEvt := &protoevent.GoalEvent{
		Timestamp: time.Now().Unix(),
		GoalId:    goalID,
		UserId:    user.Id,
		Value:     value,
		User:      user,
	}
	anyGoalEvt, err := ptypes.MarshalAny(goalEvt)
	if err != nil {
		p.loggers.Errorf(
			"bucketeer/event: PushGoalEvent failed (err: %v, userID: %s, goalID: %s, value: %g)",
			err,
			user.Id,
			goalID,
			value,
		)
		return
	}
	if err := p.pushEvent(anyGoalEvt); err != nil {
		p.loggers.Errorf(
			"bucketeer/event: PushGoalEvent failed (err: %v, userID: %s, goalID: %s, value: %g)",
			err,
			user.Id,
			goalID,
			value,
		)
		return
	}
}

func (p *processor) PushGetEvaluationLatencyMetricsEvent(ctx context.Context, duration time.Duration, tag string) {
	gelMetricsEvt := &protoevent.GetEvaluationLatencyMetricsEvent{
		Labels:   map[string]string{"tag": tag, "state": protofeature.UserEvaluations_FULL.String()},
		Duration: ptypes.DurationProto(duration),
	}
	anyGELMetricsEvt, err := ptypes.MarshalAny(gelMetricsEvt)
	if err != nil {
		p.loggers.Errorf("bucketeer/event: PushGetEvaluationLatencyMetricsEvent failed (err: %v, tag: %s)", err, tag)
		return
	}
	metricsEvt := &protoevent.MetricsEvent{
		Timestamp: time.Now().Unix(),
		Event:     anyGELMetricsEvt,
	}
	anyMetricsEvt, err := ptypes.MarshalAny(metricsEvt)
	if err != nil {
		p.loggers.Errorf("bucketeer/event: PushGetEvaluationLatencyMetricsEvent failed (err: %v, tag: %s)", err, tag)
		return
	}
	if err := p.pushEvent(anyMetricsEvt); err != nil {
		p.loggers.Errorf("bucketeer/event: PushGetEvaluationLatencyMetricsEvent failed (err: %v, tag: %s)", err, tag)
		return
	}
}

func (p *processor) PushGetEvaluationSizeMetricsEvent(ctx context.Context, sizeByte int, tag string) {
	gesMetricsEvt := &protoevent.GetEvaluationSizeMetricsEvent{
		Labels:   map[string]string{"tag": tag, "state": protofeature.UserEvaluations_FULL.String()},
		SizeByte: int32(sizeByte),
	}
	anyGESMetricsEvt, err := ptypes.MarshalAny(gesMetricsEvt)
	if err != nil {
		p.loggers.Errorf("bucketeer/event: PushGetEvaluationSizeMetricsEvent failed (err: %v, tag: %s)", err, tag)
		return
	}
	metricsEvt := &protoevent.MetricsEvent{
		Timestamp: time.Now().Unix(),
		Event:     anyGESMetricsEvt,
	}
	anyMetricsEvt, err := ptypes.MarshalAny(metricsEvt)
	if err != nil {
		p.loggers.Errorf("bucketeer/event: PushGetEvaluationSizeMetricsEvent failed (err: %v, tag: %s)", err, tag)
		return
	}
	if err := p.pushEvent(anyMetricsEvt); err != nil {
		p.loggers.Errorf("bucketeer/event: PushGetEvaluationSizeMetricsEvent failed (err: %v, tag: %s)", err, tag)
		return
	}
}

func (p *processor) PushTimeoutErrorCountMetricsEvent(ctx context.Context, tag string) {
	tecMetricsEvt := &protoevent.TimeoutErrorCountMetricsEvent{Tag: tag}
	anyTECMetricsEvt, err := ptypes.MarshalAny(tecMetricsEvt)
	if err != nil {
		p.loggers.Errorf("bucketeer/event: PushTimeoutErrorCountMetricsEvent failed (err: %v, tag: %s)", err, tag)
		return
	}
	metricsEvt := &protoevent.MetricsEvent{
		Timestamp: time.Now().Unix(),
		Event:     anyTECMetricsEvt,
	}
	anyMetricsEvt, err := ptypes.MarshalAny(metricsEvt)
	if err != nil {
		p.loggers.Errorf("bucketeer/event: PushTimeoutErrorCountMetricsEvent failed (err: %v, tag: %s)", err, tag)
		return
	}
	if err := p.pushEvent(anyMetricsEvt); err != nil {
		p.loggers.Errorf("bucketeer/event: PushTimeoutErrorCountMetricsEvent failed (err: %v, tag: %s)", err, tag)
		return
	}
}

func (p *processor) PushInternalErrorCountMetricsEvent(ctx context.Context, tag string) {
	iecMetricsEvt := &protoevent.InternalErrorCountMetricsEvent{Tag: tag}
	anyIECMetricsEvt, err := ptypes.MarshalAny(iecMetricsEvt)
	if err != nil {
		p.loggers.Errorf("bucketeer/event: PushInternalErrorCountMetricsEvent failed (err: %v, tag: %s)", err, tag)
		return
	}
	metricsEvt := &protoevent.MetricsEvent{
		Timestamp: time.Now().Unix(),
		Event:     anyIECMetricsEvt,
	}
	anyMetricsEvt, err := ptypes.MarshalAny(metricsEvt)
	if err != nil {
		p.loggers.Errorf("bucketeer/event: PushInternalErrorCountMetricsEvent failed (err: %v, tag: %s", err, tag)
		return
	}
	if err := p.pushEvent(anyMetricsEvt); err != nil {
		p.loggers.Errorf("bucketeer/event: PushInternalErrorCountMetricsEvent failed (err: %v, tag: %s", err, tag)
		return
	}
}

func (p *processor) pushEvent(anyEvt *anypb.Any) error {
	id, err := uuid.NewV4()
	if err != nil {
		return fmt.Errorf("failed to new uuid v4: %w", err)
	}
	evt := &protoevent.Event{
		Id:    id.String(),
		Event: anyEvt,
	}
	if err := p.evtQueue.push(evt); err != nil {
		return fmt.Errorf("failed to push event: %w", err)
	}
	return nil
}

func (p *processor) Close(ctx context.Context) error {
	p.evtQueue.close()
	select {
	case <-p.closeCh:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("bucketeer/event: ctx is canceled: %v", ctx.Err())
	}
}

func (p *processor) startWorkers() {
	for i := 0; i < p.numFlushWorkers; i++ {
		p.workerWG.Add(1)
		go p.runWorkerProcessLoop()
	}
	p.workerWG.Wait()
	close(p.closeCh)
}

func (p *processor) runWorkerProcessLoop() {
	defer func() {
		p.loggers.Debug("bucketeer/event: runWorkerProcessLoop done")
		p.workerWG.Done()
	}()
	events := make([]*protoevent.Event, 0, p.flushSize)
	ticker := time.NewTicker(p.flushInterval)
	defer ticker.Stop()
	for {
		select {
		case evt, ok := <-p.evtQueue.eventCh():
			if !ok {
				p.flushEvents(events)
				return
			}
			events = append(events, evt)
			if len(events) < p.flushSize {
				continue
			}
			p.flushEvents(events)
			events = make([]*protoevent.Event, 0, p.flushSize)
		case <-ticker.C:
			p.flushEvents(events)
			events = make([]*protoevent.Event, 0, p.flushSize)
		}
	}
}

func (p *processor) flushEvents(events []*protoevent.Event) {
	if len(events) == 0 {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), p.flushTimeout)
	defer cancel()
	req := &protogateway.RegisterEventsRequest{Events: events}
	res, err := p.apiClient.RegisterEvents(ctx, req)
	if err != nil {
		p.loggers.Debugf("bucketeer/event: failed to register events: %v", err)
		// Re-push all events to the event queue.
		for _, evt := range events {
			if err := p.evtQueue.push(evt); err != nil {
				p.loggers.Errorf("bucketeer/event: failed to re-push event: %v", err)
			}
		}
		return
	}
	if len(res.Errors) > 0 {
		p.loggers.Debugf("bucketeer/event: register events response contains errors, len: %d", len(res.Errors))
		// Re-push events returned retriable error to the event queue.
		for _, evt := range events {
			resErr, ok := res.Errors[evt.Id]
			if !ok {
				continue
			}
			if !resErr.Retriable {
				p.loggers.Errorf("bucketeer/event: non retriable error: %s", resErr.Message)
				continue
			}
			if err := p.evtQueue.push(evt); err != nil {
				p.loggers.Errorf("bucketeer/event: failed to re-push retriable event: %v", err)
			}
		}
	}
}
