package event

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/api"
	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/log"
	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/model"
	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/user"
	mockapi "github.com/ca-dp/bucketeer-go-server-sdk/test/mock/api"
)

const (
	processorUserID      = "user-id"
	processorFeatureID   = "feature-id"
	processorVariationID = "variation-id"
	processorGoalID      = "goal-id"
)

type registerEventsResponseError struct {
	Retriable bool   `json:"retriable,omitempty"`
	Message   string `json:"message,omitempty"`
}

func TestPushEvaluationEvent(t *testing.T) {
	t.Parallel()
	p := newProcessorForTestPushEvent(t, 10)
	user := newUser(t, processorUserID)
	evaluation := newEvaluation(t, processorFeatureID, processorVariationID)
	p.PushEvaluationEvent(context.Background(), user, evaluation)
	evt := <-p.evtQueue.eventCh()
	evalationEvt := &model.EvaluationEvent{}
	err := json.Unmarshal(evt.Event, evalationEvt)
	assert.NoError(t, err)
}

func TestPushDefaultEvaluationEvent(t *testing.T) {
	t.Parallel()
	p := newProcessorForTestPushEvent(t, 10)
	user := newUser(t, processorUserID)
	p.PushDefaultEvaluationEvent(context.Background(), user, processorFeatureID)
	evt := <-p.evtQueue.eventCh()
	evalationEvt := &model.EvaluationEvent{}
	err := json.Unmarshal(evt.Event, evalationEvt)
	assert.NoError(t, err)
}

func TestPushGoalEvent(t *testing.T) {
	t.Parallel()
	p := newProcessorForTestPushEvent(t, 10)
	user := newUser(t, processorUserID)
	p.PushGoalEvent(context.Background(), user, processorGoalID, 1.1)
	evt := <-p.evtQueue.eventCh()
	goalEvt := &model.GoalEvent{}
	err := json.Unmarshal(evt.Event, goalEvt)
	assert.NoError(t, err)
}

func TestPushLatencyMetricsEvent(t *testing.T) {
	t.Parallel()
	p := newProcessorForTestPushEvent(t, 10)
	p.PushLatencyMetricsEvent(context.Background(), time.Duration(1), model.GetEvaluation)
	evt := <-p.evtQueue.eventCh()
	metricsEvt := &model.MetricsEvent{}
	err := json.Unmarshal(evt.Event, metricsEvt)
	assert.NoError(t, err)
	gelMetricsEvt := &model.LatencyMetricsEvent{}
	err = json.Unmarshal(metricsEvt.Event, gelMetricsEvt)
	assert.NoError(t, err)
}

func TestPushSizeMetricsEvent(t *testing.T) {
	t.Parallel()
	p := newProcessorForTestPushEvent(t, 10)
	p.PushSizeMetricsEvent(context.Background(), 1, model.GetEvaluation)
	evt := <-p.evtQueue.eventCh()
	metricsEvt := &model.MetricsEvent{}
	err := json.Unmarshal(evt.Event, metricsEvt)
	assert.NoError(t, err)
	gesMetricsEvt := &model.SizeMetricsEvent{}
	err = json.Unmarshal(metricsEvt.Event, gesMetricsEvt)
	assert.NoError(t, err)
}

func TestPushTimeoutErrorMetricsEvent(t *testing.T) {
	t.Parallel()
	p := newProcessorForTestPushEvent(t, 10)
	p.PushTimeoutErrorMetricsEvent(context.Background(), model.GetEvaluation)
	evt := <-p.evtQueue.eventCh()
	metricsEvt := &model.MetricsEvent{}
	err := json.Unmarshal(evt.Event, metricsEvt)
	assert.NoError(t, err)
	tecMetricsEvt := &model.TimeoutErrorMetricsEvent{}
	err = json.Unmarshal(metricsEvt.Event, tecMetricsEvt)
	assert.NoError(t, err)
}

func TestPushInternalErrorMetricsEvent(t *testing.T) {
	t.Parallel()
	p := newProcessorForTestPushEvent(t, 10)
	p.PushInternalErrorMetricsEvent(context.Background(), model.GetEvaluation)
	evt := <-p.evtQueue.eventCh()
	metricsEvt := &model.MetricsEvent{}
	err := json.Unmarshal(evt.Event, metricsEvt)
	assert.NoError(t, err)
	iecMetricsEvt := &model.InternalErrorMetricsEvent{}
	err = json.Unmarshal(metricsEvt.Event, iecMetricsEvt)
	assert.NoError(t, err)
}

func TestPushEvent(t *testing.T) {
	t.Parallel()
	tests := []struct {
		desc               string
		eventQueueCapacity int
		encodedEvt         []byte
		isErr              bool
	}{
		{
			desc:               "return error when failed to push event",
			eventQueueCapacity: 0,
			encodedEvt:         newEncodedEvaluationEvent(t, processorFeatureID),
			isErr:              true,
		},
		{
			desc:               "success",
			eventQueueCapacity: 10,
			encodedEvt:         newEncodedEvaluationEvent(t, processorFeatureID),
			isErr:              false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			p := newProcessorForTestPushEvent(t, tt.eventQueueCapacity)
			err := p.pushEvent(tt.encodedEvt)
			if tt.isErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				evt := <-p.evtQueue.eventCh()
				evalationEvt := &model.EvaluationEvent{}
				err := json.Unmarshal(evt.Event, evalationEvt)
				assert.NoError(t, err)
			}
		})
	}
}

func newProcessorForTestPushEvent(t *testing.T, eventQueueCapacity int) *processor {
	t.Helper()
	return &processor{
		evtQueue: newQueue(&queueConfig{
			capacity: eventQueueCapacity,
		}),
		loggers: log.NewLoggers(&log.LoggersConfig{
			EnableDebugLog: false,
			ErrorLogger:    log.DiscardErrorLogger,
		}),
	}
}

func newUser(t *testing.T, id string) *user.User {
	t.Helper()
	return &user.User{ID: id}
}

func newEvaluation(t *testing.T, featureID, variationID string) *model.Evaluation {
	t.Helper()
	return &model.Evaluation{
		FeatureID:      featureID,
		FeatureVersion: 0,
		VariationID:    variationID,
		Reason:         &model.Reason{Type: model.ReasonClient},
	}
}

func newEncodedEvaluationEvent(t *testing.T, featureID string) []byte {
	t.Helper()
	evaluationEvt := &model.EvaluationEvent{FeatureID: featureID}
	encodedEvt, err := json.Marshal(evaluationEvt)
	assert.NoError(t, err)
	return encodedEvt
}

func TestFlushEvents(t *testing.T) {
	t.Parallel()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	tests := []struct {
		desc             string
		setup            func(*processor, []*model.Event)
		events           []*model.Event
		expectedQueueLen int
	}{
		{
			desc: "do nothing when events length is 0",
			setup: func(p *processor, events []*model.Event) {
				p.apiClient.(*mockapi.MockClient).EXPECT().RegisterEvents(&model.RegisterEventsRequest{Events: events}).Times(0)
			},
			events:           make([]*model.Event, 0, 10),
			expectedQueueLen: 0,
		},
		{
			desc: "re-push all events when failed to register events",
			setup: func(p *processor, events []*model.Event) {
				p.apiClient.(*mockapi.MockClient).EXPECT().RegisterEvents(&model.RegisterEventsRequest{Events: events}).Return(
					nil,
					api.NewErrStatus(http.StatusInternalServerError),
				)
			},
			events:           []*model.Event{{ID: "id-0"}, {ID: "id-1"}, {ID: "id-2"}},
			expectedQueueLen: 3,
		},
		{
			desc: "faled to re-push all events when failed to register events if queue is closed",
			setup: func(p *processor, events []*model.Event) {
				p.evtQueue.close()
				p.apiClient.(*mockapi.MockClient).EXPECT().RegisterEvents(&model.RegisterEventsRequest{Events: events}).Return(
					nil,
					api.NewErrStatus(http.StatusInternalServerError),
				)
			},
			events:           []*model.Event{{ID: "id-0"}, {ID: "id-1"}, {ID: "id-2"}},
			expectedQueueLen: 0,
		},
		{
			desc: "re-push events when register events res contains retriable errors",
			setup: func(p *processor, events []*model.Event) {
				p.apiClient.(*mockapi.MockClient).EXPECT().RegisterEvents(&model.RegisterEventsRequest{Events: events}).Return(
					&model.RegisterEventsResponse{
						Errors: map[string]*model.RegisterEventsResponseError{
							"id-0": {Retriable: true, Message: "retriable"},
							"id-1": {Retriable: false, Message: "non retriable"},
						},
					},
					nil,
				)
			},
			events:           []*model.Event{{ID: "id-0"}, {ID: "id-1"}, {ID: "id-2"}},
			expectedQueueLen: 1,
		},
		{
			desc: "faled to re-push events when register events res contains retriable errors if queue is closed",
			setup: func(p *processor, events []*model.Event) {
				p.evtQueue.close()
				p.apiClient.(*mockapi.MockClient).EXPECT().RegisterEvents(&model.RegisterEventsRequest{Events: events}).Return(
					&model.RegisterEventsResponse{
						Errors: map[string]*model.RegisterEventsResponseError{
							"id-0": {Retriable: true, Message: "retriable"},
							"id-1": {Retriable: false, Message: "non retriable"},
						},
					},
					nil,
				)
			},
			events:           []*model.Event{{ID: "id-0"}, {ID: "id-1"}, {ID: "id-2"}},
			expectedQueueLen: 0,
		},
		{
			desc: "success",
			setup: func(p *processor, events []*model.Event) {
				p.apiClient.(*mockapi.MockClient).EXPECT().RegisterEvents(&model.RegisterEventsRequest{Events: events}).Return(
					&model.RegisterEventsResponse{
						Errors: make(map[string]*model.RegisterEventsResponseError),
					},
					nil,
				)
			},
			events:           []*model.Event{{ID: "id-0"}, {ID: "id-1"}, {ID: "id-2"}},
			expectedQueueLen: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			p := newProcessorForTestWorker(t, mockCtrl)
			if tt.setup != nil {
				tt.setup(p, tt.events)
			}
			p.flushEvents(tt.events)
			assert.Len(t, p.evtQueue.eventCh(), tt.expectedQueueLen)
		})
	}
}

func TestClose(t *testing.T) {
	t.Parallel()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	tests := []struct {
		desc    string
		setup   func(*processor)
		timeout time.Duration
		isErr   bool
	}{
		{
			desc:    "return error when ctx is canceled",
			setup:   nil,
			timeout: 1 * time.Millisecond,
			isErr:   true,
		},
		{
			desc: "success",
			setup: func(p *processor) {
				go p.startWorkers()
			},
			timeout: 1 * time.Minute,
			isErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			p := newProcessorForTestWorker(t, mockCtrl)
			if tt.setup != nil {
				tt.setup(p)
			}
			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()
			err := p.Close(ctx)
			if tt.isErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func newProcessorForTestWorker(t *testing.T, mockCtrl *gomock.Controller) *processor {
	t.Helper()
	return &processor{
		evtQueue: newQueue(&queueConfig{
			capacity: 10,
		}),
		numFlushWorkers: 3,
		flushInterval:   1 * time.Minute,
		flushSize:       10,
		flushTimeout:    10 * time.Second,
		apiClient:       mockapi.NewMockClient(mockCtrl),
		loggers: log.NewLoggers(&log.LoggersConfig{
			EnableDebugLog: false,
			ErrorLogger:    log.DiscardErrorLogger,
		}),
		closeCh: make(chan struct{}),
	}
}
