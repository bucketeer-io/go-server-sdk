package bucketeer

import (
	"fmt"

	event "github.com/ca-dp/bucketeer-go-server-sdk/proto/event/client"
)

type Error struct {
	UserID         string
	UserAttributes map[string]string
	FeatureID      string
	Message        string
	Goal           *event.GoalEvent
	Evaluation     *event.EvaluationEvent
	Retryable      bool
	originalErr    error
}

func (e *Error) Error() string {
	return fmt.Sprintf("bucketeer: %s", e.Message)
}

func (e *Error) Unwrap() error {
	return e.originalErr
}
