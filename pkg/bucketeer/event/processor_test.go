package event

import (
	"context"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/log"
	protoevent "github.com/ca-dp/bucketeer-go-server-sdk/proto/event/client"
	protofeature "github.com/ca-dp/bucketeer-go-server-sdk/proto/feature"
	protouser "github.com/ca-dp/bucketeer-go-server-sdk/proto/user"
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
	p := newProcessor(t, 10)
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
	p := newProcessor(t, 10)
	user := newUser(t, processorUserID)
	p.PushDefaultEvaluationEvent(context.Background(), user, processorFeatureID)
	evt := <-p.evtQueue.eventCh()
	evalationEvt := &protoevent.EvaluationEvent{}
	err := ptypes.UnmarshalAny(evt.Event, evalationEvt)
	assert.NoError(t, err)
}

func TestPushGoalEvent(t *testing.T) {
	t.Parallel()
	p := newProcessor(t, 10)
	user := newUser(t, processorUserID)
	p.PushGoalEvent(context.Background(), user, processorGoalID, 1.1)
	evt := <-p.evtQueue.eventCh()
	goalEvt := &protoevent.GoalEvent{}
	err := ptypes.UnmarshalAny(evt.Event, goalEvt)
	assert.NoError(t, err)
}

func TestPushGetEvaluationLatencyMetricsEvent(t *testing.T) {
	t.Parallel()
	p := newProcessor(t, 10)
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
	p := newProcessor(t, 10)
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
	p := newProcessor(t, 10)
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
	p := newProcessor(t, 10)
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
			p := newProcessor(t, tt.eventQueueCapacity)
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

func newProcessor(t *testing.T, eventQueueCapacity int) *processor {
	t.Helper()
	return &processor{
		evtQueue: newQueue(&queueConfig{
			capacity: eventQueueCapacity},
		),
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
