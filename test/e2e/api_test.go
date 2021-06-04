package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"

	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer"
	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/api"
	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/uuid"
	protoevent "github.com/ca-dp/bucketeer-go-server-sdk/proto/event/client"
	protofeature "github.com/ca-dp/bucketeer-go-server-sdk/proto/feature"
	protogateway "github.com/ca-dp/bucketeer-go-server-sdk/proto/gateway"
)

func TestGetEvaluation(t *testing.T) {
	t.Parallel()
	client := newAPIClient(t)
	defer client.Close()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	user := bucketeer.NewUser(userID, nil)
	req := &protogateway.GetEvaluationRequest{
		Tag:       tag,
		User:      user.User,
		FeatureId: featureID,
	}
	res, err := client.GetEvaluation(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, featureID, res.Evaluation.FeatureId)
	assert.Equal(t, featureIDVariation2, res.Evaluation.VariationValue)
}

func TestRegisterEvents(t *testing.T) {
	t.Parallel()
	client := newAPIClient(t)
	defer client.Close()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	user := bucketeer.NewUser(userID, nil)
	evaluationEvent, err := ptypes.MarshalAny(&protoevent.EvaluationEvent{
		Timestamp:      time.Now().Unix(),
		FeatureId:      featureID,
		FeatureVersion: 0,
		UserId:         user.Id,
		VariationId:    "",
		User:           user.User,
		Reason:         &protofeature.Reason{Type: protofeature.Reason_CLIENT},
	})
	assert.NoError(t, err)
	goalEvent, err := ptypes.MarshalAny(&protoevent.GoalEvent{
		Timestamp: time.Now().Unix(),
		GoalId:    goalID,
		UserId:    user.Id,
		Value:     0.0,
		User:      user.User,
	})
	req := &protogateway.RegisterEventsRequest{
		Events: []*protoevent.Event{
			{
				Id:    newUUID(t),
				Event: evaluationEvent,
			},
			{
				Id:    newUUID(t),
				Event: goalEvent,
			},
		},
	}
	res, err := client.RegisterEvents(ctx, req)
	assert.NoError(t, err)
	assert.Len(t, res.Errors, 0)
}

func newAPIClient(t *testing.T) api.Client {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	conf := &api.ClientConfig{
		APIKey: *apiKey,
		Host:   *host,
		Port:   *port,
	}
	client, err := api.NewClient(ctx, conf)
	assert.NoError(t, err)
	return client
}

func newUUID(t *testing.T) string {
	t.Helper()
	id, err := uuid.NewV4()
	assert.NoError(t, err)
	return id.String()
}
