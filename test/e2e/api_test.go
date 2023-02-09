package e2e

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/api"
	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/models"
	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/user"
	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/uuid"
)

func TestGetEvaluation(t *testing.T) {
	t.Parallel()
	client := newAPIClient(t)
	user := user.NewUser(userID, nil)
	res, err := client.GetEvaluation(&models.GetEvaluationRequest{Tag: tag, User: user, FeatureID: featureID})
	assert.NoError(t, err)
	assert.Equal(t, featureID, res.Evaluation.FeatureID)
	assert.Equal(t, featureIDVariation2, res.Evaluation.VariationValue)
}

func TestRegisterEvents(t *testing.T) {
	t.Parallel()
	client := newAPIClient(t)
	user := user.NewUser(userID, nil)
	evaluationEvent, err := json.Marshal(&models.EvaluationEvent{
		Timestamp:      time.Now().Unix(),
		SourceID:       models.SourceIDGoServer,
		Tag:            tag,
		FeatureID:      featureID,
		FeatureVersion: 0,
		VariationID:    "",
		User:           user,
		Reason:         &models.Reason{Type: models.ReasonClient},
		Type:           models.EvaluationEventType,
	})
	assert.NoError(t, err)
	goalEvent, err := json.Marshal(&models.GoalEvent{
		Timestamp: time.Now().Unix(),
		SourceID:  models.SourceIDGoServer,
		Tag:       tag,
		GoalID:    goalID,
		UserID:    user.ID,
		Value:     0.0,
		User:      user,
		Type:      models.GoalEventType,
	})
	assert.NoError(t, err)
	gesMetricsEvt, err := json.Marshal(&models.GetEvaluationSizeMetricsEvent{
		Labels: map[string]string{
			"tag": tag,
		},
		Type: models.GetEvaluationSizeMetricsEventType,
	})
	assert.NoError(t, err)
	gmetricsEvent, err := json.Marshal(&models.MetricsEvent{
		Timestamp: time.Now().Unix(),
		Event:     gesMetricsEvt,
		Type:      models.MetricsEventType,
	})
	assert.NoError(t, err)
	iecMetricsEvt, err := json.Marshal(&models.InternalErrorCountMetricsEvent{
		Tag:  tag,
		Type: models.InternalErrorCountMetricsEventType,
	})
	assert.NoError(t, err)
	imetricsEvent, err := json.Marshal(&models.MetricsEvent{
		Timestamp: time.Now().Unix(),
		Event:     iecMetricsEvt,
		Type:      models.MetricsEventType,
	})
	assert.NoError(t, err)
	tecMetricsEvent, err := json.Marshal(&models.TimeoutErrorCountMetricsEvent{
		Tag:  tag,
		Type: models.TimeoutErrorCountMetricsEventType,
	})
	assert.NoError(t, err)
	tmetricsEvent, err := json.Marshal(&models.MetricsEvent{
		Timestamp: time.Now().Unix(),
		Event:     tecMetricsEvent,
		Type:      models.MetricsEventType,
	})
	assert.NoError(t, err)
	elmMetricsEvent, err := json.Marshal(&models.GetEvaluationLatencyMetricsEvent{
		Labels: map[string]string{"tag": tag},
		Duration: &models.Duration{
			Type:  models.DurationType,
			Value: "5s",
		},
		Type: models.GetEvaluationLatencyMetricsEventType,
	})
	assert.NoError(t, err)
	emetricsEvent, err := json.Marshal(&models.MetricsEvent{
		Timestamp: time.Now().Unix(),
		Event:     elmMetricsEvent,
		Type:      models.MetricsEventType,
	})
	assert.NoError(t, err)
	req := &models.RegisterEventsRequest{
		Events: []*models.Event{
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
