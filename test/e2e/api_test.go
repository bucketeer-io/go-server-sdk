package e2e

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/api"
	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/user"
	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/uuid"
)

func TestGetEvaluation(t *testing.T) {
	t.Parallel()
	client := newAPIClient(t)
	user := user.NewUser(userID, nil)
	res, err := client.GetEvaluation(&api.GetEvaluationRequest{Tag: tag, User: user, FeatureID: featureID})
	assert.NoError(t, err)
	assert.Equal(t, featureID, res.Evaluation.FeatureID)
	assert.Equal(t, featureIDVariation2, res.Evaluation.VariationValue)
}

func TestRegisterEvents(t *testing.T) {
	t.Parallel()
	client := newAPIClient(t)
	user := user.NewUser(userID, nil)
	evaluationEvent, err := json.Marshal(api.EvaluationEvent{
		Timestamp:      time.Now().Unix(),
		SourceID:       api.SourceIDGoServer,
		Tag:            tag,
		FeatureID:      featureID,
		FeatureVersion: 0,
		VariationID:    "",
		User:           user,
		Reason:         &api.Reason{Type: api.ReasonClient},
		Type:           api.EvaluationEventType,
	})
	assert.NoError(t, err)
	goalEvent, err := json.Marshal(&api.GoalEvent{
		Timestamp: time.Now().Unix(),
		SourceID:  api.SourceIDGoServer,
		Tag:       tag,
		GoalID:    goalID,
		UserID:    user.ID,
		Value:     0.0,
		User:      user,
		Type:      api.GoalEventType,
	})
	assert.NoError(t, err)
	gesMetricsEvt, err := json.Marshal(&api.GetEvaluationSizeMetricsEvent{
		Labels: map[string]string{
			"tag": tag,
		},
		Type: api.GetEvaluationSizeMetricsEventType,
	})
	assert.NoError(t, err)
	gmetricsEvent, err := json.Marshal(&api.MetricsEvent{
		Timestamp: time.Now().Unix(),
		Event:     gesMetricsEvt,
		Type:      api.MetricsEventType,
	})
	assert.NoError(t, err)
	iecMetricsEvt, err := json.Marshal(&api.InternalErrorCountMetricsEvent{
		Tag:  tag,
		Type: api.InternalErrorCountMetricsEventType,
	})
	assert.NoError(t, err)
	imetricsEvent, err := json.Marshal(&api.MetricsEvent{
		Timestamp: time.Now().Unix(),
		Event:     iecMetricsEvt,
		Type:      api.MetricsEventType,
	})
	assert.NoError(t, err)
	tecMetricsEvent, err := json.Marshal(&api.TimeoutErrorCountMetricsEvent{
		Tag:  tag,
		Type: api.TimeoutErrorCountMetricsEventType,
	})
	assert.NoError(t, err)
	tmetricsEvent, err := json.Marshal(&api.MetricsEvent{
		Timestamp: time.Now().Unix(),
		Event:     tecMetricsEvent,
		Type:      api.MetricsEventType,
	})
	assert.NoError(t, err)
	elmMetricsEvent, err := json.Marshal(&api.GetEvaluationLatencyMetricsEvent{
		Labels: map[string]string{"tag": tag},
		Duration: &api.Duration{
			Type:  api.DurationType,
			Value: "5s",
		},
		Type: api.GetEvaluationLatencyMetricsEventType,
	})
	assert.NoError(t, err)
	emetricsEvent, err := json.Marshal(&api.MetricsEvent{
		Timestamp: time.Now().Unix(),
		Event:     elmMetricsEvent,
		Type:      api.MetricsEventType,
	})
	assert.NoError(t, err)
	req := &api.RegisterEventsRequest{
		Events: []*api.Event{
			{
				ID:    newUUID(t),
				Event: evaluationEvent,
			},
			{
				ID:    newUUID(t),
				Event: goalEvent,
			},
			{
				ID:    newUUID(t),
				Event: gmetricsEvent,
			},
			{
				ID:    newUUID(t),
				Event: imetricsEvent,
			},
			{
				ID:    newUUID(t),
				Event: tmetricsEvent,
			},
			{
				ID:    newUUID(t),
				Event: emetricsEvent,
			},
		},
	}
	res, err := client.RegisterEvents(req)
	assert.NoError(t, err)
	assert.Len(t, res.Errors, 0)
}

func newAPIClient(t *testing.T) api.Client {
	t.Helper()
	conf := &api.ClientConfig{
		APIKey: *apiKey,
		Host:   *host,
	}
	client, err := api.NewClient(conf)
	assert.NoError(t, err)
	return client
}

func newUUID(t *testing.T) string {
	t.Helper()
	id, err := uuid.NewV4()
	assert.NoError(t, err)
	return id.String()
}
