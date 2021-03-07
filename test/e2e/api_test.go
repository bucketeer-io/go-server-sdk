package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer"
	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/api"
	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/uuid"
	protoevent "github.com/ca-dp/bucketeer-go-server-sdk/proto/event/client"
	protofeature "github.com/ca-dp/bucketeer-go-server-sdk/proto/feature"
	protogateway "github.com/ca-dp/bucketeer-go-server-sdk/proto/gateway"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetEvaluation(t *testing.T) {
	t.Parallel()
	client := newAPIClient(t)
	defer client.Close()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	user, err := bucketeer.NewUser(userID, nil)
	require.NoError(t, err)
	req := &protogateway.GetEvaluationsRequest{
		Tag:       tag,
		User:      user.User,
		FeatureId: featureID,
	}
	res, err := client.GetEvaluations(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, protofeature.UserEvaluations_FULL, res.State)
	assert.Len(t, res.Evaluations.Evaluations, 1)
	assert.Equal(t, featureIDVariation2, res.Evaluations.Evaluations[0].Variation.Value)
}

func TestRegisterEvents(t *testing.T) {
	t.Parallel()
	client := newAPIClient(t)
	defer client.Close()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	user, err := bucketeer.NewUser(userID, nil)
	require.NoError(t, err)
	evaluationEvent, err := ptypes.MarshalAny(&protoevent.EvaluationEvent{
		Timestamp:      time.Now().Unix(),
		FeatureId:      featureID,
		FeatureVersion: 0,
		UserId:         user.Id,
		VariationId:    "",
		User:           user.User,
		Reason:         &protofeature.Reason{Type: protofeature.Reason_CLIENT},
	})
	require.NoError(t, err)
	goalEvent, err := ptypes.MarshalAny(&protoevent.GoalEvent{
		Timestamp:   time.Now().Unix(),
		GoalId:      goalID,
		UserId:      user.Id,
		Value:       0.0,
		User:        user.User,
		Evaluations: []*protofeature.Evaluation{},
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
	require.NoError(t, err)
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
	require.NoError(t, err)
	return client
}

func newUUID(t *testing.T) string {
	t.Helper()
	id, err := uuid.NewV4()
	require.NoError(t, err)
	return id.String()
}
