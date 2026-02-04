//go:generate mockgen -source=$GOFILE -package=$GOPACKAGE -destination=../../../test/mock/$GOPACKAGE/$GOFILE
package event

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/user"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/api"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/log"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/model"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/uuid"
)

// ProcessorStats contains event processing statistics.
type ProcessorStats struct {
	EventsCreated int64 // Events successfully pushed to queue
	EventsSent    int64 // Events successfully sent to server
	EventsDropped int64 // Events dropped due to queue full
	EventsRetried int64 // Events re-pushed after API error
}

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

	// PushDefaultEvaluationEvent pushes the default evaluation event with a specific reason to the queue.
	PushDefaultEvaluationEvent(user *user.User, featureID string, reason model.ReasonType)

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

	// Stats returns current event processing statistics.
	Stats() ProcessorStats
}

type processor struct {
	evtQueue        *queue
	numFlushWorkers int
	flushInterval   time.Duration
	flushSize       int
	flushTimeout    time.Duration
	apiClient       api.Client
	loggers         *log.Loggers
	closeCh         chan struct{} // Signals dispatcher completion to Close()
	workerSem       chan struct{} // Semaphore: limits concurrent workers AND tracks completion
	tag             string
	sdkVersion      string
	sourceID        model.SourceIDType

	// Stats tracking
	enableStats   bool
	eventsCreated int64 // atomic
	eventsSent    int64 // atomic
	eventsDropped int64 // atomic
	eventsRetried int64 // atomic
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

	// EnableStats enables event processing statistics tracking.
	// When true, the processor tracks event counts (created, sent, dropped, retried).
	EnableStats bool
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
		workerSem:       make(chan struct{}, conf.NumFlushWorkers), // Semaphore limits concurrent workers
		tag:             conf.Tag,
		sdkVersion:      conf.SDKVersion,
		sourceID:        conf.SourceID,
		enableStats:     conf.EnableStats,
	}
	go p.runDispatcher(ctx)
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

func (p *processor) PushDefaultEvaluationEvent(user *user.User, featureID string, reason model.ReasonType) {
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
			"bucketeer/event: PushDefaultEvaluationEvent failed (err: %v, UserID: %s, featureID: %s)",
			err,
			user.ID,
			featureID,
		)
		return
	}
	if err := p.PushEvent(encodedEvaluationEvt); err != nil {
		p.loggers.Errorf(
			"bucketeer/event: PushDefaultEvaluationEvent failed (err: %v, UserID: %s, featureID: %s)",
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
		if p.enableStats {
			atomic.AddInt64(&p.eventsDropped, 1)
		}
		return fmt.Errorf("failed to push event: %w", err)
	}
	if p.enableStats {
		atomic.AddInt64(&p.eventsCreated, 1)
	}
	return nil
}

// Stats returns current event processing statistics.
func (p *processor) Stats() ProcessorStats {
	return ProcessorStats{
		EventsCreated: atomic.LoadInt64(&p.eventsCreated),
		EventsSent:    atomic.LoadInt64(&p.eventsSent),
		EventsDropped: atomic.LoadInt64(&p.eventsDropped),
		EventsRetried: atomic.LoadInt64(&p.eventsRetried),
	}
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

// pollInterval is the fixed interval for the dispatcher to poll the queue.
// 1 second keeps CPU overhead minimal while the shutdown drain handles any
// remaining events that haven't been polled yet.
const pollInterval = 1 * time.Second

// runDispatcher is the single goroutine that polls the queue and dispatches
// batches to workers. This design:
// - Uses only 1 polling goroutine instead of N workers
// - Spawns worker goroutines on-demand via semaphore
// - Under low load: 1 goroutine polling
// - Under high load: 1 dispatcher + up to numFlushWorkers workers
func (p *processor) runDispatcher(ctx context.Context) {
	defer p.loggers.Debug("bucketeer/event: runDispatcher done")

	// Fixed poll ticker - single dispatcher polls the queue
	pollTicker := time.NewTicker(pollInterval)
	defer pollTicker.Stop()

	// User-configured flush ticker for partial batches
	flushTicker := time.NewTicker(p.flushInterval)
	defer flushTicker.Stop()

	for {
		select {
		case <-pollTicker.C:
			// Drain queue in batches (single lock per batch via popMany)
			for {
				events := p.evtQueue.popMany(p.flushSize)
				if len(events) == 0 {
					break
				}
				p.submitBatch(ctx, events)
				// Reset flush ticker since we just sent a batch
				flushTicker.Reset(p.flushInterval)
			}

		case <-flushTicker.C:
			// Flush any events that accumulated (partial batch)
			events := p.evtQueue.popMany(p.flushSize)
			if len(events) > 0 {
				p.submitBatch(ctx, events)
			}

		case <-p.evtQueue.doneCh():
			// Queue is closed, drain remaining events using workers (parallel)
			for {
				events := p.evtQueue.popMany(p.flushSize)
				if len(events) == 0 {
					break
				}
				p.submitBatch(ctx, events)
			}
			// Wait for all in-flight workers by acquiring all semaphore permits
			// When we hold all permits, all workers must have finished
			p.waitForWorkers()
			close(p.closeCh)
			return
		}
	}
}

// waitForWorkers blocks until all in-flight workers complete.
// It works by acquiring all semaphore permits - when we hold all permits,
// no workers can be running.
func (p *processor) waitForWorkers() {
	for i := 0; i < cap(p.workerSem); i++ {
		p.workerSem <- struct{}{} // Acquire permit (blocks if worker holds it)
	}
	// All permits acquired = all workers done
	// Release them
	for i := 0; i < cap(p.workerSem); i++ {
		<-p.workerSem
	}
}

// submitBatch submits a batch to a worker goroutine if one is available.
// If all workers are busy, it flushes synchronously (backpressure).
// Note: batch ownership is transferred to this function (no copy needed since
// popMany already returns a new slice).
func (p *processor) submitBatch(ctx context.Context, batch []*model.Event) {
	if len(batch) == 0 {
		return
	}

	select {
	case p.workerSem <- struct{}{}: // Acquire semaphore permit
		go func() {
			defer func() {
				<-p.workerSem // Release semaphore permit
			}()
			p.flushEvents(ctx, batch)
		}()
	default:
		// All workers busy - flush synchronously (backpressure)
		p.loggers.Debug("bucketeer/event: all workers busy, flushing synchronously")
		p.flushEvents(ctx, batch)
	}
}

func (p *processor) flushEvents(ctx context.Context, events []*model.Event) {
	if len(events) == 0 {
		return
	}

	// Calculate deadline: use 80% of flush interval or flush timeout, whichever is smaller
	// Minimum 5 seconds to allow retry attempts
	deadlineDuration := p.flushInterval * 8 / 10
	if p.flushTimeout > 0 && p.flushTimeout < deadlineDuration {
		deadlineDuration = p.flushTimeout
	}
	const minDeadline = 5 * time.Second
	if deadlineDuration < minDeadline {
		deadlineDuration = minDeadline
	}
	deadline := time.Now().Add(deadlineDuration)

	res, _, err := p.apiClient.RegisterEvents(ctx, model.NewRegisterEventsRequest(events, p.sourceID), deadline)
	if err != nil {
		p.loggers.Debugf("bucketeer/event: failed to register events: %v", err)
		// Re-push all events to the event queue.
		var retriedCount int64
		for _, evt := range events {
			if err := p.evtQueue.push(evt); err != nil {
				p.loggers.Errorf("bucketeer/event: failed to re-push event: %v", err)
			} else {
				retriedCount++
			}
		}
		if p.enableStats {
			atomic.AddInt64(&p.eventsRetried, retriedCount)
		}
		p.PushErrorEvent(err, model.RegisterEvents)
		return
	}

	// Count successful sends and handle partial errors
	var sentCount int64
	var retriedCount int64
	if len(res.Errors) > 0 {
		p.loggers.Debugf("bucketeer/event: register events response contains errors, len: %d", len(res.Errors))
		// Re-push events returned retriable error to the event queue.
		for _, evt := range events {
			resErr, ok := res.Errors[evt.ID]
			if !ok {
				// No error for this event, it was sent successfully
				sentCount++
				continue
			}
			if !resErr.Retriable {
				p.loggers.Errorf("bucketeer/event: non retriable error: %s", resErr.Message)
				// Non-retriable errors are counted as sent (won't be retried)
				sentCount++
				continue
			}
			if err := p.evtQueue.push(evt); err != nil {
				p.loggers.Errorf("bucketeer/event: failed to re-push retriable event: %v", err)
			} else {
				retriedCount++
			}
		}
	} else {
		// All events sent successfully
		sentCount = int64(len(events))
	}
	if p.enableStats {
		atomic.AddInt64(&p.eventsSent, sentCount)
		atomic.AddInt64(&p.eventsRetried, retriedCount)
	}
}

// errorType represents the classification of an error for metrics reporting.
type errorType int

const (
	errorTypeUnknown errorType = iota
	errorTypeNetwork
	errorTypeTimeout
	errorTypeStatusCode
	errorTypeInternal
)

// classifyError determines the type of error for metrics reporting.
// Uses proper Go error type assertions instead of string matching.
func classifyError(err error) (errorType, int) {
	if err == nil {
		return errorTypeUnknown, 0
	}
	// Check HTTP status code first (highest priority)
	if errType, code := classifyHTTPStatusError(err); errType != errorTypeUnknown {
		return errType, code
	}
	// Check context errors (deadline/cancellation)
	if errType := classifyContextError(err); errType != errorTypeUnknown {
		return errType, 0
	}
	// Check network-related errors
	if errType := classifyNetworkError(err); errType != errorTypeUnknown {
		return errType, 0
	}
	return errorTypeInternal, 0
}

// classifyHTTPStatusError checks for HTTP status code errors from the API client.
func classifyHTTPStatusError(err error) (errorType, int) {
	code, ok := api.GetStatusCode(err)
	if ok && code != http.StatusOK {
		return errorTypeStatusCode, code
	}
	return errorTypeUnknown, 0
}

// classifyContextError checks for context-related errors (deadline exceeded, canceled).
func classifyContextError(err error) errorType {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return errorTypeTimeout
	case errors.Is(err, context.Canceled):
		return errorTypeNetwork // Treat cancellation as network issue
	default:
		return errorTypeUnknown
	}
}

// classifyNetworkError checks for network-related errors (net.Error, DNS, syscall, TLS, EOF).
func classifyNetworkError(err error) errorType {
	// Check for EOF errors (connection closed unexpectedly)
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return errorTypeNetwork
	}

	// Check TLS/certificate errors (x509 errors are network-layer issues)
	if isTLSError(err) {
		return errorTypeNetwork
	}

	// Check net.Error interface (includes timeout check)
	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return errorTypeTimeout
		}
		return errorTypeNetwork
	}

	// Check url.Error (wraps network errors from http.Client)
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		if urlErr.Timeout() {
			return errorTypeTimeout
		}
		// Recursively check the wrapped error
		if urlErr.Err != nil {
			return classifyNetworkError(urlErr.Err)
		}
		return errorTypeNetwork
	}

	// Check DNS errors
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		if dnsErr.IsTimeout {
			return errorTypeTimeout
		}
		return errorTypeNetwork
	}

	// Check syscall errors (connection refused, reset, broken pipe, etc.)
	if isNetworkSyscallError(err) {
		return errorTypeNetwork
	}

	return errorTypeUnknown
}

// isTLSError checks if the error is a TLS/certificate-related error.
// These are classified as network errors since they prevent connection establishment.
func isTLSError(err error) bool {
	// Check for x509 certificate errors
	var x509Err x509.CertificateInvalidError
	if errors.As(err, &x509Err) {
		return true
	}

	var x509UnknownErr x509.UnknownAuthorityError
	if errors.As(err, &x509UnknownErr) {
		return true
	}

	var x509HostErr x509.HostnameError
	if errors.As(err, &x509HostErr) {
		return true
	}

	// Fallback: check error message for TLS-related strings
	// This catches TLS errors that may be wrapped or have custom types.
	// Only use specific prefixes (tls:, x509:) to avoid false positives
	// from generic words like "certificate" in unrelated error messages.
	errMsg := err.Error()
	return strings.Contains(errMsg, "tls:") ||
		strings.Contains(errMsg, "x509:")
}

// isNetworkSyscallError checks if the error is a network-related syscall error.
func isNetworkSyscallError(err error) bool {
	var syscallErr syscall.Errno
	if !errors.As(err, &syscallErr) {
		return false
	}
	switch syscallErr {
	// Connection errors
	case syscall.ECONNREFUSED, syscall.ECONNRESET, syscall.ECONNABORTED, syscall.EPIPE:
		return true
	// Network unreachable errors
	case syscall.EHOSTUNREACH, syscall.ENETUNREACH, syscall.ENETDOWN, syscall.ENETRESET:
		return true
	// Network-layer timeout errors: treat ETIMEDOUT as a network error since
	// it is raised by the underlying syscall layer; this also acts as a
	// fallback detection for timeouts not caught by higher-level logic.
	case syscall.ETIMEDOUT:
		return true
	default:
		return false
	}
}

func (p *processor) PushErrorEvent(err error, apiID model.APIID) {
	errType, statusCode := classifyError(err)

	switch errType {
	case errorTypeNetwork:
		p.PushNetworkErrorMetricsEvent(apiID)
	case errorTypeTimeout:
		p.pushTimeoutErrorMetricsEvent(apiID)
	case errorTypeStatusCode:
		if statusCode == http.StatusGatewayTimeout {
			p.pushTimeoutErrorMetricsEvent(apiID)
		} else {
			p.pushErrorStatusCodeMetricsEvent(apiID, statusCode, err)
		}
	case errorTypeInternal, errorTypeUnknown:
		p.pushInternalSDKErrorMetricsEvent(apiID, err)
	}
}
