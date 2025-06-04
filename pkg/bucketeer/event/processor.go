//go:generate mockgen -source=$GOFILE -package=$GOPACKAGE -destination=../../../test/mock/$GOPACKAGE/$GOFILE
package event

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/user"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/api"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/log"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/model"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/uuid"
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
	PushEvaluationEvent(user *user.User, evaluation *model.Evaluation)

	// PushDefaultEvaluationEvent pushes the default evaluation event to the queue.
	PushDefaultEvaluationEvent(user *user.User, featureID string)

	// PushDefaultEvaluationEventWithReason pushes the default evaluation event with a specific reason to the queue.
	PushDefaultEvaluationEventWithReason(user *user.User, featureID string, reason model.ReasonType)

	// PushGoalEvent pushes the goal event to the queue.
	PushGoalEvent(user *user.User, GoalID string, value float64)

	// PushLatencyMetricsEvent pushes the get evaluation latency metrics event to the queue.
	PushLatencyMetricsEvent(duration time.Duration, api model.APIID)

	// PushSizeMetricsEvent pushes the get evaluation size metrics event to the queue.
	PushSizeMetricsEvent(sizeByte int, api model.APIID)

	// Close tears down all Processor activities, after ensuring that all events have been delivered.
	Close(ctx context.Context) error

	// PushErrorEvent pushes the error event to the queue.
	PushErrorEvent(err error, api model.APIID)

	// PushEvent pushes events to the queue.
	PushEvent(encoded []byte) error
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
	sdkVersion      string
	sourceID        model.SourceIDType
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

	// SDKVersion is the SDK version.
	SDKVersion string

	// SourceID is the source ID of the SDK.
	SourceID model.SourceIDType
}

// NewProcessor creates a new Processor.
func NewProcessor(ctx context.Context, conf *ProcessorConfig) Processor {
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
		sdkVersion:      conf.SDKVersion,
		sourceID:        conf.SourceID,
	}
	go p.startWorkers(ctx)
	return p
}

func (p *processor) PushEvaluationEvent(
	user *user.User,
	evaluation *model.Evaluation,
) {
	evaluationEvt := model.NewEvaluationEvent(
		p.tag,
		evaluation.FeatureID,
		evaluation.VariationID,
		p.sdkVersion,
		evaluation.FeatureVersion,
		p.sourceID,
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
	if err := p.PushEvent(encodedEvaluationEvt); err != nil {
		p.loggers.Errorf(
			"bucketeer/event: PushEvaluationEvent failed (err: %v, UserID: %s, featureID: %s)",
			err,
			user.ID,
			evaluation.FeatureID,
		)
		return
	}
}

func (p *processor) PushDefaultEvaluationEvent(user *user.User, featureID string) {
	p.PushDefaultEvaluationEventWithReason(user, featureID, model.ReasonErrorException)
}

func (p *processor) PushDefaultEvaluationEventWithReason(user *user.User, featureID string, reason model.ReasonType) {
	evaluationEvt := model.NewEvaluationEvent(
		p.tag,
		featureID,
		"",
		p.sdkVersion,
		0,
		p.sourceID,
		user,
		&model.Reason{Type: reason},
	)
	encodedEvaluationEvt, err := json.Marshal(evaluationEvt)
	if err != nil {
		p.loggers.Errorf(
			"bucketeer/event: PushDefaultEvaluationEventWithReason failed (err: %v, UserID: %s, featureID: %s)",
			err,
			user.ID,
			featureID,
		)
		return
	}
	if err := p.PushEvent(encodedEvaluationEvt); err != nil {
		p.loggers.Errorf(
			"bucketeer/event: PushDefaultEvaluationEventWithReason failed (err: %v, UserID: %s, featureID: %s)",
			err,
			user.ID,
			featureID,
		)
		return
	}
}

func (p *processor) PushGoalEvent(user *user.User, GoalID string, value float64) {
	goalEvt := model.NewGoalEvent(p.tag, GoalID, p.sdkVersion, value, p.sourceID, user)
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
	if err := p.PushEvent(encodedGoalEvt); err != nil {
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

func (p *processor) PushLatencyMetricsEvent(duration time.Duration, api model.APIID) {
	gelMetricsEvt := model.NewLatencyMetricsEvent(p.tag, duration.Seconds(), api)
	encodedGELMetricsEvt, err := json.Marshal(gelMetricsEvt)
	if err != nil {
		p.loggers.Errorf("bucketeer/event: PushLatencyMetricsEvent failed (err: %v, tag: %s)", err, p.tag)
		return
	}
	metricsEvt := model.NewMetricsEvent(encodedGELMetricsEvt, p.sourceID, p.sdkVersion)
	encodedMetricsEvt, err := json.Marshal(metricsEvt)
	if err != nil {
		p.loggers.Errorf("bucketeer/event: PushLatencyMetricsEvent failed (err: %v, tag: %s)", err, p.tag)
		return
	}
	if err := p.PushEvent(encodedMetricsEvt); err != nil {
		p.loggers.Errorf("bucketeer/event: PushLatencyMetricsEvent failed (err: %v, tag: %s)", err, p.tag)
		return
	}
}

func (p *processor) PushSizeMetricsEvent(sizeByte int, api model.APIID) {
	gesMetricsEvt := model.NewSizeMetricsEvent(p.tag, int32(sizeByte), api)
	encodedGESMetricsEvt, err := json.Marshal(gesMetricsEvt)
	if err != nil {
		p.loggers.Errorf("bucketeer/event: PushSizeMetricsEvent failed (err: %v, tag: %s)", err, p.tag)
		return
	}
	metricsEvt := model.NewMetricsEvent(encodedGESMetricsEvt, p.sourceID, p.sdkVersion)
	encodedMetricsEvt, err := json.Marshal(metricsEvt)
	if err != nil {
		p.loggers.Errorf("bucketeer/event: PushSizeMetricsEvent failed (err: %v, tag: %s)", err, p.tag)
		return
	}
	if err := p.PushEvent(encodedMetricsEvt); err != nil {
		p.loggers.Errorf("bucketeer/event: PushSizeMetricsEvent failed (err: %v, tag: %s)", err, p.tag)
		return
	}
}

func (p *processor) pushTimeoutErrorMetricsEvent(api model.APIID) {
	tecMetricsEvt := model.NewTimeoutErrorMetricsEvent(p.tag, api)

	encodedTECMetricsEvt, err := json.Marshal(tecMetricsEvt)
	if err != nil {
		p.loggers.Errorf("bucketeer/event: pushTimeoutErrorMetricsEvent failed (err: %v, tag: %s)", err, p.tag)
		return
	}
	metricsEvt := model.NewMetricsEvent(encodedTECMetricsEvt, p.sourceID, p.sdkVersion)
	encodedMetricsEvt, err := json.Marshal(metricsEvt)
	if err != nil {
		p.loggers.Errorf("bucketeer/event: pushTimeoutErrorMetricsEvent failed (err: %v, tag: %s)", err, p.tag)
		return
	}
	if err := p.PushEvent(encodedMetricsEvt); err != nil {
		p.loggers.Errorf("bucketeer/event: pushTimeoutErrorMetricsEvent failed (err: %v, tag: %s)", err, p.tag)
		return
	}
}

func (p *processor) pushInternalSDKErrorMetricsEvent(api model.APIID, err error) {
	iecMetricsEvt := model.NewInternalSDKErrorMetricsEvent(p.tag, api, err.Error())
	encodedIECMetricsEvt, err := json.Marshal(iecMetricsEvt)
	if err != nil {
		p.loggers.Errorf("bucketeer/event: pushInternalSDKErrorMetricsEvent failed (err: %v, tag: %s)", err, p.tag)
		return
	}
	metricsEvt := model.NewMetricsEvent(encodedIECMetricsEvt, p.sourceID, p.sdkVersion)
	encodedMetricsEvt, err := json.Marshal(metricsEvt)
	if err != nil {
		p.loggers.Errorf("bucketeer/event: pushInternalSDKErrorMetricsEvent failed (err: %v, tag: %s", err, p.tag)
		return
	}
	if err := p.PushEvent(encodedMetricsEvt); err != nil {
		p.loggers.Errorf("bucketeer/event: pushInternalSDKErrorMetricsEvent failed (err: %v, tag: %s", err, p.tag)
		return
	}
}

func (p *processor) pushErrorStatusCodeMetricsEvent(api model.APIID, code int, err error) {
	var evt interface{}
	switch {
	// Update error metrics report
	// https://github.com/bucketeer-io/bucketeer/issues/799
	case 300 <= code && code < 400:
		evt = model.NewRedirectionRequestErrorMetricsEvent(p.tag, api, code)
	case code == http.StatusBadRequest:
		evt = model.NewBadRequestErrorMetricsEvent(p.tag, api)
		// We don't generate error events for unaturothized and forbidden errors
		// because they can't be reported using an invalid API key, so we only log them.
	case code == http.StatusUnauthorized || code == http.StatusForbidden:
		p.loggers.Errorf("bucketeer/event: failed to request (err: %v, tag: %s, api: %d)", err, p.tag, api)
		return
	case code == http.StatusNotFound:
		evt = model.NewNotFoundErrorMetricsEvent(p.tag, api)
	case code == http.StatusMethodNotAllowed:
		evt = model.NewInternalSDKErrorMetricsEvent(p.tag, api, err.Error())
	case code == http.StatusRequestTimeout:
		evt = model.NewTimeoutErrorMetricsEvent(p.tag, api)
	case code == http.StatusRequestEntityTooLarge:
		evt = model.NewPayloadTooLargeErrorMetricsEvent(p.tag, api)
	case code == 499:
		evt = model.NewClientClosedRequestErrorMetricsEvent(p.tag, api)
	case code == http.StatusInternalServerError:
		evt = model.NewInternalServerErrorMetricsEvent(p.tag, api)
	case code == http.StatusServiceUnavailable || code == http.StatusBadGateway:
		evt = model.NewServiceUnavailableErrorMetricsEvent(p.tag, api)
	default:
		evt = model.NewUnknownErrorMetricsEvent(p.tag, code, err.Error(), api)
	}
	encodedESCMetricsEvt, err := json.Marshal(evt)
	if err != nil {
		p.loggers.Errorf("bucketeer/event: pushErrorStatusCodeMetricsEvent failed (err: %v, tag: %s)", err, p.tag)
		return
	}
	metricsEvt := model.NewMetricsEvent(encodedESCMetricsEvt, p.sourceID, p.sdkVersion)
	encodedMetricsEvt, err := json.Marshal(metricsEvt)
	if err != nil {
		p.loggers.Errorf("bucketeer/event: pushErrorStatusCodeMetricsEvent failed (err: %v, tag: %s", err, p.tag)
		return
	}
	if err := p.PushEvent(encodedMetricsEvt); err != nil {
		p.loggers.Errorf("bucketeer/event: pushErrorStatusCodeMetricsEvent failed (err: %v, tag: %s", err, p.tag)
		return
	}
}

func (p *processor) PushNetworkErrorMetricsEvent(api model.APIID) {
	neMetricsEvt := model.NewNetworkErrorMetricsEvent(p.tag, api)
	encodedNEMetricsEvt, err := json.Marshal(neMetricsEvt)
	if err != nil {
		p.loggers.Errorf("bucketeer/event: PushNetworkErrorMetricsEvent failed (err: %v, tag: %s)", err, p.tag)
		return
	}
	metricsEvt := model.NewMetricsEvent(encodedNEMetricsEvt, p.sourceID, p.sdkVersion)
	encodedMetricsEvt, err := json.Marshal(metricsEvt)
	if err != nil {
		p.loggers.Errorf("bucketeer/event: PushNetworkErrorMetricsEvent failed (err: %v, tag: %s", err, p.tag)
		return
	}
	if err := p.PushEvent(encodedMetricsEvt); err != nil {
		p.loggers.Errorf("bucketeer/event: PushNetworkErrorMetricsEvent failed (err: %v, tag: %s", err, p.tag)
		return
	}
}

func (p *processor) PushEvent(encoded []byte) error {
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

func (p *processor) startWorkers(ctx context.Context) {
	for i := 0; i < p.numFlushWorkers; i++ {
		p.workerWG.Add(1)
		go p.runWorkerProcessLoop(ctx)
	}
	p.workerWG.Wait()
	close(p.closeCh)
}

func (p *processor) runWorkerProcessLoop(ctx context.Context) {
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
				p.flushEvents(ctx, events)
				return
			}
			events = append(events, evt)
			if len(events) < p.flushSize {
				continue
			}
			p.flushEvents(ctx, events)
			events = make([]*model.Event, 0, p.flushSize)
		case <-ticker.C:
			p.flushEvents(ctx, events)
			events = make([]*model.Event, 0, p.flushSize)
		}
	}
}

func (p *processor) flushEvents(ctx context.Context, events []*model.Event) {
	if len(events) == 0 {
		return
	}
	res, _, err := p.apiClient.RegisterEvents(model.NewRegisterEventsRequest(events, p.sourceID))
	if err != nil {
		p.loggers.Debugf("bucketeer/event: failed to register events: %v", err)
		// Re-push all events to the event queue.
		for _, evt := range events {
			if err := p.evtQueue.push(evt); err != nil {
				p.loggers.Errorf("bucketeer/event: failed to re-push event: %v", err)
			}
		}
		p.PushErrorEvent(err, model.RegisterEvents)
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

func (p *processor) PushErrorEvent(err error, apiID model.APIID) {
	code, ok := api.GetStatusCode(err)
	if !ok {
		switch {
		case strings.Contains(err.Error(), syscall.EHOSTUNREACH.Error()) ||
			strings.Contains(err.Error(), syscall.ECONNREFUSED.Error()):
			p.PushNetworkErrorMetricsEvent(apiID)
		case os.IsTimeout(err):
			p.pushTimeoutErrorMetricsEvent(apiID)
		default:
			p.pushInternalSDKErrorMetricsEvent(apiID, err)
		}
		return
	}
	if code == http.StatusGatewayTimeout {
		p.pushTimeoutErrorMetricsEvent(apiID)
	} else {
		p.pushErrorStatusCodeMetricsEvent(apiID, code, err)
	}
}
