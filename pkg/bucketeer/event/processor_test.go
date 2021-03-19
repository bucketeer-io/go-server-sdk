package event

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/log"
	protoevent "github.com/ca-dp/bucketeer-go-server-sdk/proto/event/client"
	protofeature "github.com/ca-dp/bucketeer-go-server-sdk/proto/feature"
	protogateway "github.com/ca-dp/bucketeer-go-server-sdk/proto/gateway"
	protouser "github.com/ca-dp/bucketeer-go-server-sdk/proto/user"
	mockapi "github.com/ca-dp/bucketeer-go-server-sdk/test/mock/api"
)

const (
	processorTag         = "go-server"
	processorUserID      = "user-id"
	processorFeatureID   = "feature-id"
	processorVariationID = "variation-id"
	processorGoalID      = "goal-id"
)

func TestPushEvaluationEvent(t *testing.T) {
	t.Parallel()
	p := newProcessorForTestPushEvent(t, 10)
	user := newUser(t, processorUserID)
	evaluation := newEvaluation(t, processorFeatureID, processorVariationID)
	p.PushEvaluationEvent(context.Background(), user, evaluation)
	evt := <-p.evtQueue.eventCh()
	evalationEvt := &protoevent.EvaluationEvent{}
	err := ptypes.UnmarshalAny(evt.Event, evalationEvt)
	assert.NoError(t, err)
}

func TestPushDefaultEvaluationEvent(t *testing.T) {
	t.Parallel()
	p := newProcessorForTestPushEvent(t, 10)
	user := newUser(t, processorUserID)
	p.PushDefaultEvaluationEvent(context.Background(), user, processorFeatureID)
	evt := <-p.evtQueue.eventCh()
	evalationEvt := &protoevent.EvaluationEvent{}
	err := ptypes.UnmarshalAny(evt.Event, evalationEvt)
	assert.NoError(t, err)
}

func TestPushGoalEvent(t *testing.T) {
	t.Parallel()
	p := newProcessorForTestPushEvent(t, 10)
	user := newUser(t, processorUserID)
	p.PushGoalEvent(context.Background(), user, processorGoalID, 1.1)
	evt := <-p.evtQueue.eventCh()
	goalEvt := &protoevent.GoalEvent{}
	err := ptypes.UnmarshalAny(evt.Event, goalEvt)
	assert.NoError(t, err)
}

func TestPushGetEvaluationLatencyMetricsEvent(t *testing.T) {
	t.Parallel()
	p := newProcessorForTestPushEvent(t, 10)
	p.PushGetEvaluationLatencyMetricsEvent(context.Background(), time.Duration(1), processorTag)
	evt := <-p.evtQueue.eventCh()
	metricsEvt := &protoevent.MetricsEvent{}
	err := ptypes.UnmarshalAny(evt.Event, metricsEvt)
	assert.NoError(t, err)
	gelMetricsEvt := &protoevent.GetEvaluationLatencyMetricsEvent{}
	err = ptypes.UnmarshalAny(metricsEvt.Event, gelMetricsEvt)
	assert.NoError(t, err)
}

func TestPushGetEvaluationSizeMetricsEvent(t *testing.T) {
	t.Parallel()
	p := newProcessorForTestPushEvent(t, 10)
	p.PushGetEvaluationSizeMetricsEvent(context.Background(), 1, processorTag)
	evt := <-p.evtQueue.eventCh()
	metricsEvt := &protoevent.MetricsEvent{}
	err := ptypes.UnmarshalAny(evt.Event, metricsEvt)
	assert.NoError(t, err)
	gesMetricsEvt := &protoevent.GetEvaluationSizeMetricsEvent{}
	err = ptypes.UnmarshalAny(metricsEvt.Event, gesMetricsEvt)
	assert.NoError(t, err)
}

func TestPushTimeoutErrorCountMetricsEvent(t *testing.T) {
	t.Parallel()
	p := newProcessorForTestPushEvent(t, 10)
	p.PushTimeoutErrorCountMetricsEvent(context.Background(), processorTag)
	evt := <-p.evtQueue.eventCh()
	metricsEvt := &protoevent.MetricsEvent{}
	err := ptypes.UnmarshalAny(evt.Event, metricsEvt)
	assert.NoError(t, err)
	tecMetricsEvt := &protoevent.TimeoutErrorCountMetricsEvent{}
	err = ptypes.UnmarshalAny(metricsEvt.Event, tecMetricsEvt)
	assert.NoError(t, err)
}

func TestPushInternalErrorCountMetricsEvent(t *testing.T) {
	t.Parallel()
	p := newProcessorForTestPushEvent(t, 10)
	p.PushInternalErrorCountMetricsEvent(context.Background(), processorTag)
	evt := <-p.evtQueue.eventCh()
	metricsEvt := &protoevent.MetricsEvent{}
	err := ptypes.UnmarshalAny(evt.Event, metricsEvt)
	assert.NoError(t, err)
	iecMetricsEvt := &protoevent.InternalErrorCountMetricsEvent{}
	err = ptypes.UnmarshalAny(metricsEvt.Event, iecMetricsEvt)
	assert.NoError(t, err)
}

func TestPushEvent(t *testing.T) {
	t.Parallel()
	tests := []struct {
		desc               string
		eventQueueCapacity int
		anyEvt             *anypb.Any
		isErr              bool
	}{
		{
			desc:               "return error when failed to push event",
			eventQueueCapacity: 0,
			anyEvt:             newAnyEvaluationEvent(t, processorFeatureID),
			isErr:              true,
		},
		{
			desc:               "success",
			eventQueueCapacity: 10,
			anyEvt:             newAnyEvaluationEvent(t, processorFeatureID),
			isErr:              false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			p := newProcessorForTestPushEvent(t, tt.eventQueueCapacity)
			err := p.pushEvent(tt.anyEvt)
			if tt.isErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				evt := <-p.evtQueue.eventCh()
				evalationEvt := &protoevent.EvaluationEvent{}
				err := ptypes.UnmarshalAny(evt.Event, evalationEvt)
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

func newUser(t *testing.T, id string) *protouser.User {
	t.Helper()
	return &protouser.User{Id: id}
}

func newEvaluation(t *testing.T, featureID, variationID string) *protofeature.Evaluation {
	t.Helper()
	return &protofeature.Evaluation{
		FeatureId:      featureID,
		FeatureVersion: 0,
		VariationId:    variationID,
		Reason:         &protofeature.Reason{Type: protofeature.Reason_CLIENT},
	}
}

func newAnyEvaluationEvent(t *testing.T, featureID string) *anypb.Any {
	t.Helper()
	evaluationEvt := &protoevent.EvaluationEvent{FeatureId: featureID}
	anyEvt, err := ptypes.MarshalAny(evaluationEvt)
	require.NoError(t, err)
	return anyEvt
}

func TestFlushEvents(t *testing.T) {
	t.Parallel()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	tests := []struct {
		desc             string
		setup            func(*processor, []*protoevent.Event)
		events           []*protoevent.Event
		expectedQueueLen int
	}{
		{
			desc: "do nothing when events length is 0",
			setup: func(p *processor, events []*protoevent.Event) {
				req := &protogateway.RegisterEventsRequest{Events: events}
				p.apiClient.(*mockapi.MockClient).EXPECT().RegisterEvents(gomock.Any(), req).Times(0)
			},
			events:           make([]*protoevent.Event, 0, 10),
			expectedQueueLen: 0,
		},
		{
			desc: "re-push all events when failed to register events",
			setup: func(p *processor, events []*protoevent.Event) {
				req := &protogateway.RegisterEventsRequest{Events: events}
				p.apiClient.(*mockapi.MockClient).EXPECT().RegisterEvents(gomock.Any(), req).Return(
					nil,
					status.Error(codes.Internal, "error"),
				)
			},
			events:           []*protoevent.Event{{Id: "id-0"}, {Id: "id-1"}, {Id: "id-2"}},
			expectedQueueLen: 3,
		},
		{
			desc: "faled to re-push all events when failed to register events if queue is closed",
			setup: func(p *processor, events []*protoevent.Event) {
				p.evtQueue.close()
				req := &protogateway.RegisterEventsRequest{Events: events}
				p.apiClient.(*mockapi.MockClient).EXPECT().RegisterEvents(gomock.Any(), req).Return(
					nil,
					status.Error(codes.Internal, "error"),
				)
			},
			events:           []*protoevent.Event{{Id: "id-0"}, {Id: "id-1"}, {Id: "id-2"}},
			expectedQueueLen: 0,
		},
		{
			desc: "re-push events when register events res contains retriable errors",
			setup: func(p *processor, events []*protoevent.Event) {
				req := &protogateway.RegisterEventsRequest{Events: events}
				p.apiClient.(*mockapi.MockClient).EXPECT().RegisterEvents(gomock.Any(), req).Return(
					&protogateway.RegisterEventsResponse{
						Errors: map[string]*protogateway.RegisterEventsResponse_Error{
							"id-0": {Retriable: true, Message: "retriable"},
							"id-1": {Retriable: false, Message: "non retriable"},
						},
					},
					nil,
				)
			},
			events:           []*protoevent.Event{{Id: "id-0"}, {Id: "id-1"}, {Id: "id-2"}},
			expectedQueueLen: 1,
		},
		{
			desc: "faled to re-push events when register events res contains retriable errors if queue is closed",
			setup: func(p *processor, events []*protoevent.Event) {
				p.evtQueue.close()
				req := &protogateway.RegisterEventsRequest{Events: events}
				p.apiClient.(*mockapi.MockClient).EXPECT().RegisterEvents(gomock.Any(), req).Return(
					&protogateway.RegisterEventsResponse{
						Errors: map[string]*protogateway.RegisterEventsResponse_Error{
							"id-0": {Retriable: true, Message: "retriable"},
							"id-1": {Retriable: false, Message: "non retriable"},
						},
					},
					nil,
				)
			},
			events:           []*protoevent.Event{{Id: "id-0"}, {Id: "id-1"}, {Id: "id-2"}},
			expectedQueueLen: 0,
		},
		{
			desc: "success",
			setup: func(p *processor, events []*protoevent.Event) {
				req := &protogateway.RegisterEventsRequest{Events: events}
				p.apiClient.(*mockapi.MockClient).EXPECT().RegisterEvents(gomock.Any(), req).Return(
					&protogateway.RegisterEventsResponse{
						Errors: make(map[string]*protogateway.RegisterEventsResponse_Error),
					},
					nil,
				)
			},
			events:           []*protoevent.Event{{Id: "id-0"}, {Id: "id-1"}, {Id: "id-2"}},
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

func newProcessorForTestWorker(t *testing.T, mockCtrl *gomock.Controller) *processor {
	t.Helper()
	return &processor{
		evtQueue: newQueue(&queueConfig{
			capacity: 10,
		}),
		flushTimeout: 10 * time.Second,
		apiClient:    mockapi.NewMockClient(mockCtrl),
		loggers: log.NewLoggers(&log.LoggersConfig{
			EnableDebugLog: false,
			ErrorLogger:    log.DiscardErrorLogger,
		}),
	}
}
