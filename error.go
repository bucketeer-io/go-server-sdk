package bucketeer_go_sdk

import (
	"context"

	event "github.com/ca-dp/bucketeer-go-sdk/proto/event/client"
)

type Error struct {
	UserID         string
	UserAttributes map[string]string
	FeatureID      string
	Message        string
	Experiment     *event.ExperimentEvent
	originalErr    error
}

func (e *Error) Error() string {
	return e.Message
}

func newError(ctx context.Context, err error, featureID string, experiment *event.ExperimentEvent, message string) *Error {
	return &Error{
		UserID:         userID(ctx),
		UserAttributes: attributes(ctx),
		FeatureID:      featureID,
		Message:        message,
		Experiment:     experiment,
		originalErr:    err,
	}
}
