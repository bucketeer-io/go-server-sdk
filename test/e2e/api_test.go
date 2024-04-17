package e2e

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/api"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/model"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/user"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/uuid"
)

func TestGetEvaluation(t *testing.T) {
	t.Parallel()
	client := newAPIClient(t)
	user := user.NewUser(userID, nil)
	res, _, err := client.GetEvaluation(model.NewGetEvaluationRequest(tag, featureID, user))
	assert.NoError(t, err)
	assert.Equal(t, featureID, res.Evaluation.FeatureID)
	assert.Equal(t, featureIDVariation2, res.Evaluation.VariationValue)
}

func TestRegisterEvents(t *testing.T) {
	t.Parallel()
	client := newAPIClient(t)
	user := user.NewUser(userID, nil)
	evaluationEvent, err := json.Marshal(&model.EvaluationEvent{
		Timestamp:      time.Now().Unix(),
		SourceID:       model.SourceIDGoServer,
		Tag:            tag,
		FeatureID:      featureID,
		FeatureVersion: 0,
		VariationID:    "",
		User:           user,
		Reason:         &model.Reason{Type: model.ReasonClient},
		Type:           model.EvaluationEventType,
	})
	assert.NoError(t, err)
	goalEvent, err := json.Marshal(&model.GoalEvent{
		Timestamp: time.Now().Unix(),
		SourceID:  model.SourceIDGoServer,
		Tag:       tag,
		GoalID:    goalID,
		UserID:    user.ID,
		Value:     0.0,
		User:      user,
		Type:      model.GoalEventType,
	})
	assert.NoError(t, err)
	sizeMetrics, err := json.Marshal(&model.SizeMetricsEvent{
		APIID: model.GetEvaluation,
		Labels: map[string]string{
			"tag": tag,
		},
		Type: model.SizeMetricsEventType,
	})
	assert.NoError(t, err)
	sizeMetricsEvent, err := json.Marshal(model.NewMetricsEvent(sizeMetrics))
	assert.NoError(t, err)
	internalError, err := json.Marshal(&model.InternalSDKErrorMetricsEvent{
		APIID:  model.GetEvaluation,
		Labels: map[string]string{"tag": tag},
		Type:   model.InternalSDKErrorMetricsEventType,
	})
	assert.NoError(t, err)
	iemetricsEvent, err := json.Marshal(model.NewMetricsEvent(internalError))
	assert.NoError(t, err)
	timeoutError, err := json.Marshal(&model.TimeoutErrorMetricsEvent{
		APIID:  model.GetEvaluation,
		Labels: map[string]string{"tag": tag},
		Type:   model.TimeoutErrorMetricsEventType,
	})
	assert.NoError(t, err)
	temetricsEvent, err := json.Marshal(model.NewMetricsEvent(timeoutError))
	assert.NoError(t, err)
	latency, err := json.Marshal(&model.LatencyMetricsEvent{
		APIID:         model.GetEvaluation,
		Labels:        map[string]string{"tag": tag},
		LatencySecond: 0.1,
		Type:          model.LatencyMetricsEventType,
	})
	assert.NoError(t, err)
	lmetricsEvent, err := json.Marshal(model.NewMetricsEvent(latency))
	assert.NoError(t, err)
	badRequest, err := json.Marshal(model.NewBadRequestErrorMetricsEvent(tag, model.GetEvaluation))
	assert.NoError(t, err)
	brmetricsEvent, err := json.Marshal(model.NewMetricsEvent(badRequest))
	assert.NoError(t, err)
	internalServerError, err := json.Marshal(model.NewInternalServerErrorMetricsEvent(tag, model.GetEvaluation))
	assert.NoError(t, err)
	iesmetricsEvent, err := json.Marshal(model.NewMetricsEvent(internalServerError))
	assert.NoError(t, err)
	req := model.NewRegisterEventsRequest(
		[]*model.Event{
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
				Event: sizeMetricsEvent,
			},
			{
				ID:    newUUID(t),
				Event: iemetricsEvent,
			},
			{
				ID:    newUUID(t),
				Event: temetricsEvent,
			},
			{
				ID:    newUUID(t),
				Event: lmetricsEvent,
			},
			{
				ID:    newUUID(t),
				Event: brmetricsEvent,
			},
			{
				ID:    newUUID(t),
				Event: iesmetricsEvent,
			},
		},
	)
	res, _, err := client.RegisterEvents(req)
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
