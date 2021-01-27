package bucketeer_go_sdk

import (
	"context"
	"errors"
	"testing"
	"time"

	event "github.com/ca-dp/bucketeer-go-sdk/proto/event/client"
	"github.com/stretchr/testify/require"
)

func TestNewError(t *testing.T) {
	const (
		featureID = "featureid"
		message   = "msg"
		userID    = "userid"
	)
	attrs := map[string]string{
		"foo": "bar",
	}
	experiment := &event.ExperimentEvent{
		Timestamp:    time.Now().Unix(),
		ExperimentId: "experiment_id",
		FeatureId:    "feature_id",
	}
	err := errors.New("error")
	ctx := NewContext(context.Background(), userID, attrs)

	expect := &Error{
		UserID:         userID,
		UserAttributes: attrs,
		FeatureID:      featureID,
		Message:        message,
		Experiment:     experiment,
		originalErr:    err,
	}

	ret := newError(ctx, err, featureID, experiment, message)
	require.Equal(t, expect, ret)
	require.Equal(t, message, ret.Error())
}
