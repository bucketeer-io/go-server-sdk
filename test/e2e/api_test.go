package e2e

import (
	"encoding/json"
	"strconv"
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
			"tag":   tag,
			"state": strconv.Itoa(int(api.UserEvaluationsFULL)),
		},
		Type: api.GetEvaluationSizeMetricsEventType,
	})
	assert.NoError(t, err)
	metricsEvent, err := json.Marshal(&api.MetricsEvent{
		Timestamp: time.Now().Unix(),
		Event:     gesMetricsEvt,
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
				Event: metricsEvent,
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
