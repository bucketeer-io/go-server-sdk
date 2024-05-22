package e2e

import (
	"encoding/json"
	"testing"
	"time"

	ftproto "github.com/bucketeer-io/bucketeer/proto/feature"
	"github.com/stretchr/testify/assert"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/api"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/model"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/user"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/uuid"
)

func TestGetEvaluation(t *testing.T) {
	t.Parallel()
	client := newAPIClient(t, *apiKey)
	user := user.NewUser(userID, nil)
	res, _, err := client.GetEvaluation(model.NewGetEvaluationRequest(tag, featureID, user))
	assert.NoError(t, err)
	assert.Equal(t, featureID, res.Evaluation.FeatureID)
	assert.Equal(t, featureIDVariation2, res.Evaluation.VariationValue)
}

func TestGetFeatureFlags(t *testing.T) {
	t.Parallel()
	client := newAPIClient(t, *apiKeyServer)

	// Get all the features by tag
	featureFlagsID := ""
	requestedAt := int64(1)
	resp, _, err := client.GetFeatureFlags(model.NewGetFeatureFlagsRequest(tag, featureFlagsID, requestedAt))
	assert.NoError(t, err)
	assert.True(t, len(resp.Features) >= 1)
	assert.True(t, resp.FeatureFlagsId != featureFlagsID)
	assert.True(t, resp.RequestedAt > requestedAt)
	assert.True(t, resp.ForceUpdate)
	assert.True(t, findFeature(t, resp.Features, featureIDString))
	assert.True(t, findFeature(t, resp.Features, featureIDBoolean))
	assert.True(t, findFeature(t, resp.Features, featureIDInt))
	assert.True(t, findFeature(t, resp.Features, featureIDInt64))
	assert.True(t, findFeature(t, resp.Features, featureIDFloat))
	assert.True(t, findFeature(t, resp.Features, featureIDJson))

	time.Sleep(time.Second)

	// Use the `featureFlagsID` and `requestedAt` to get an empty response
	featureFlagsID = resp.FeatureFlagsId
	requestedAt = resp.RequestedAt
	resp, _, err = client.GetFeatureFlags(model.NewGetFeatureFlagsRequest(tag, featureFlagsID, requestedAt))
	assert.NoError(t, err)
	assert.Empty(t, resp.Features)
	assert.True(t, resp.FeatureFlagsId == featureFlagsID)
	assert.True(t, resp.RequestedAt > requestedAt)
	assert.False(t, resp.ForceUpdate)
}

func findFeature(t *testing.T, features []*ftproto.Feature, featureID string) bool {
	t.Helper()
	for _, f := range features {
		if f.Id == featureID {
			return true
		}
	}
	return false
}

func TestGetSegmentUsers(t *testing.T) {
	t.Parallel()
	client := newAPIClient(t, *apiKeyServer)

	segmentIDs := []string{""}
	requestedAt := int64(1)
	resp, _, err := client.GetSegmentUsers(model.NewGetSegmentUsersRequest(segmentIDs, requestedAt))
	assert.NoError(t, err)
	assert.True(t, len(resp.SegmentUsers) == 1)
	assert.Empty(t, resp.DeletedSegmentIds)
	assert.True(t, resp.RequestedAt > requestedAt)
	assert.True(t, resp.ForceUpdate)

	time.Sleep(time.Second)

	// Use the `segmentIDs` and `requestedAt` to get an empty response
	randomID := "random-id"
	segmentIDs = []string{resp.SegmentUsers[0].SegmentId, randomID}
	requestedAt = resp.RequestedAt
	resp, _, err = client.GetSegmentUsers(model.NewGetSegmentUsersRequest(segmentIDs, requestedAt))
	assert.NoError(t, err)
	assert.Empty(t, resp.SegmentUsers)
	assert.NotEmpty(t, resp.DeletedSegmentIds)
	assert.Contains(t, resp.DeletedSegmentIds, randomID)
	assert.True(t, resp.RequestedAt > requestedAt)
	assert.False(t, resp.ForceUpdate)
}

func TestRegisterEvents(t *testing.T) {
	t.Parallel()
	client := newAPIClient(t, *apiKey)
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

func newAPIClient(t *testing.T, apiKey string) api.Client {
	t.Helper()
	conf := &api.ClientConfig{
		APIKey: apiKey,
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
