//go:generate mockgen -source=$GOFILE -package=$GOPACKAGE -destination=../../../test/mock/$GOPACKAGE/$GOFILE
package event

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/user"

	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/api"
	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/log"
	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/model"
	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/uuid"
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
	PushEvaluationEvent(ctx context.Context, user *user.User, evaluation *model.Evaluation)

	// PushDefaultEvaluationEvent pushes the default evaluation event to the queue.
	PushDefaultEvaluationEvent(ctx context.Context, user *user.User, featureID string)

	// PushGoalEvent pushes the goal event to the queue.
	PushGoalEvent(ctx context.Context, user *user.User, GoalID string, value float64)

	// PushLatencyMetricsEvent pushes the get evaluation latency metrics event to the queue.
	PushLatencyMetricsEvent(ctx context.Context, duration time.Duration, api model.APIID)

	// PushSizeMetricsEvent pushes the get evaluation size metrics event to the queue.
	PushSizeMetricsEvent(ctx context.Context, sizeByte int, api model.APIID)

	// PushTimeoutErrorMetricsEvent pushes the timeout error count metrics event to the queue.
	PushTimeoutErrorMetricsEvent(ctx context.Context, api model.APIID)

	// PushInternalErrorMetricsEvent pushes the internal error count metrics event to the queue.
	PushInternalErrorMetricsEvent(ctx context.Context, api model.APIID)

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
	tag             string
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

	// Tag is the Feature Flag tag
	//
	// The tag is set when a Feature Flag is created, and can be retrieved from the admin console.
	Tag string
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
		tag:             conf.Tag,
	}
	go p.startWorkers()
	return p
}

func (p *processor) PushEvaluationEvent(
	ctx context.Context,
	user *user.User,
	evaluation *model.Evaluation,
) {
	evaluationEvt := model.NewEvaluationEvent(
		p.tag,
		evaluation.FeatureID,
		evaluation.VariationID,
		evaluation.FeatureVersion,
		user,
		evaluation.Reason,
	)
	encodedEvaluationEvt, err := json.Marshal(evaluationEvt)
	if err != nil {
		p.loggers.Errorf(
			"bucketeer/event: PushEvaluationEvent failed (err: %v, UserID: %s, featureID: %s)",
			err,
			user.ID,
			evaluation.FeatureID,
		)
		return
	}
	if err := p.pushEvent(encodedEvaluationEvt); err != nil {
		p.loggers.Errorf(
			"bucketeer/event: PushEvaluationEvent failed (err: %v, UserID: %s, featureID: %s)",
			err,
			user.ID,
			evaluation.FeatureID,
		)
		return
	}
}

func (p *processor) PushDefaultEvaluationEvent(ctx context.Context, user *user.User, featureID string) {
	evaluationEvt := model.NewEvaluationEvent(
		p.tag,
		featureID,
		"",
		0,
		user,
		&model.Reason{Type: model.ReasonClient},
	)
	encodedEvaluationEvt, err := json.Marshal(evaluationEvt)
	if err != nil {
		p.loggers.Errorf(
			"bucketeer/event: PushDefaultEvaluationEvent failed (err: %v, UserID: %s, featureID: %s)",
			err,
			user.ID,
			featureID,
		)
		return
	}
	if err := p.pushEvent(encodedEvaluationEvt); err != nil {
		p.loggers.Errorf(
			"bucketeer/event: PushDefaultEvaluationEvent failed (err: %v, UserID: %s, featureID: %s)",
			err,
			user.ID,
			featureID,
		)
		return
	}
}

func (p *processor) PushGoalEvent(ctx context.Context, user *user.User, GoalID string, value float64) {
	goalEvt := model.NewGoalEvent(p.tag, GoalID, value, user)
	encodedGoalEvt, err := json.Marshal(goalEvt)
	if err != nil {
		p.loggers.Errorf(
			"bucketeer/event: PushGoalEvent failed (err: %v, UserID: %s, GoalID: %s, value: %g)",
			err,
			user.ID,
			GoalID,
			value,
		)
		return
	}
	if err := p.pushEvent(encodedGoalEvt); err != nil {
		p.loggers.Errorf(
			"bucketeer/event: PushGoalEvent failed (err: %v, UserID: %s, GoalID: %s, value: %g)",
			err,
			user.ID,
			GoalID,
			value,
		)
		return
	}
}

func (p *processor) PushLatencyMetricsEvent(ctx context.Context, duration time.Duration, api model.APIID) {
	val := fmt.Sprintf("%ds", duration.Microseconds()/1000)
	gelMetricsEvt := model.NewLatencyMetricsEvent(p.tag, val, api)
	encodedGELMetricsEvt, err := json.Marshal(gelMetricsEvt)
	if err != nil {
		p.loggers.Errorf("bucketeer/event: PushLatencyMetricsEvent failed (err: %v, tag: %s)", err, p.tag)
		return
	}
	metricsEvt := model.NewMetricsEvent(encodedGELMetricsEvt)
	encodedMetricsEvt, err := json.Marshal(metricsEvt)
	if err != nil {
		p.loggers.Errorf("bucketeer/event: PushLatencyMetricsEvent failed (err: %v, tag: %s)", err, p.tag)
		return
	}
	if err := p.pushEvent(encodedMetricsEvt); err != nil {
		p.loggers.Errorf("bucketeer/event: PushLatencyMetricsEvent failed (err: %v, tag: %s)", err, p.tag)
		return
	}
}

func (p *processor) PushSizeMetricsEvent(ctx context.Context, sizeByte int, api model.APIID) {
	gesMetricsEvt := model.NewSizeMetricsEvent(p.tag, int32(sizeByte), api)
	encodedGESMetricsEvt, err := json.Marshal(gesMetricsEvt)
	if err != nil {
		p.loggers.Errorf("bucketeer/event: PushSizeMetricsEvent failed (err: %v, tag: %s)", err, p.tag)
		return
	}
	metricsEvt := model.NewMetricsEvent(encodedGESMetricsEvt)
	encodedMetricsEvt, err := json.Marshal(metricsEvt)
	if err != nil {
		p.loggers.Errorf("bucketeer/event: PushSizeMetricsEvent failed (err: %v, tag: %s)", err, p.tag)
		return
	}
	if err := p.pushEvent(encodedMetricsEvt); err != nil {
		p.loggers.Errorf("bucketeer/event: PushSizeMetricsEvent failed (err: %v, tag: %s)", err, p.tag)
		return
	}
}

func (p *processor) PushTimeoutErrorMetricsEvent(ctx context.Context, api model.APIID) {
	tecMetricsEvt := model.NewTimeoutErrorMetricsEvent(p.tag, api)

	encodedTECMetricsEvt, err := json.Marshal(tecMetricsEvt)
	if err != nil {
		p.loggers.Errorf("bucketeer/event: PushTimeoutErrorMetricsEvent failed (err: %v, tag: %s)", err, p.tag)
		return
	}
	metricsEvt := model.NewMetricsEvent(encodedTECMetricsEvt)
	encodedMetricsEvt, err := json.Marshal(metricsEvt)
	if err != nil {
		p.loggers.Errorf("bucketeer/event: PushTimeoutErrorMetricsEvent failed (err: %v, tag: %s)", err, p.tag)
		return
	}
	if err := p.pushEvent(encodedMetricsEvt); err != nil {
		p.loggers.Errorf("bucketeer/event: PushTimeoutErrorMetricsEvent failed (err: %v, tag: %s)", err, p.tag)
		return
	}
}

func (p *processor) PushInternalErrorMetricsEvent(ctx context.Context, api model.APIID) {
	iecMetricsEvt := model.NewInternalErrorMetricsEvent(p.tag, api)
	encodedIECMetricsEvt, err := json.Marshal(iecMetricsEvt)
	if err != nil {
		p.loggers.Errorf("bucketeer/event: PushInternalErrorMetricsEvent failed (err: %v, tag: %s)", err, p.tag)
		return
	}
	metricsEvt := model.NewMetricsEvent(encodedIECMetricsEvt)
	encodedMetricsEvt, err := json.Marshal(metricsEvt)
	if err != nil {
		p.loggers.Errorf("bucketeer/event: PushInternalErrorMetricsEvent failed (err: %v, tag: %s", err, p.tag)
		return
	}
	if err := p.pushEvent(encodedMetricsEvt); err != nil {
		p.loggers.Errorf("bucketeer/event: PushInternalErrorMetricsEvent failed (err: %v, tag: %s", err, p.tag)
		return
	}
}

func (p *processor) pushEvent(encoded []byte) error {
	id, err := uuid.NewV4()
	if err != nil {
		return fmt.Errorf("failed to model.New uuid v4: %w", err)
	}
	evt := model.NewEvent(id.String(), encoded)
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
	events := make([]*model.Event, 0, p.flushSize)
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
			events = make([]*model.Event, 0, p.flushSize)
		case <-ticker.C:
			p.flushEvents(events)
			events = make([]*model.Event, 0, p.flushSize)
		}
	}
}

func (p *processor) flushEvents(events []*model.Event) {
	if len(events) == 0 {
		return
	}
	res, err := p.apiClient.RegisterEvents(&model.RegisterEventsRequest{Events: events})
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
			resErr, ok := res.Errors[evt.ID]
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
