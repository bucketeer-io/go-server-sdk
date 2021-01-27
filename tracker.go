package bucketeer_go_sdk

import (
	"time"

	event "github.com/ca-dp/bucketeer-go-sdk/proto/event/client"
	"github.com/ca-dp/bucketeer-go-sdk/proto/user"
)

type Tracker interface {
	Track(goalID string, value float64)
}

type tracker struct {
	ch             chan<- *event.ExperimentEvent
	now            func() time.Time
	experimentID   string
	featureID      string
	featureVersion int32
	variationID    string
	userID         string
	attributes     map[string]string
}

func (t *tracker) Track(goalID string, value float64) {
	if t == nil {
		return
	}

	experiment := &event.ExperimentEvent{
		Timestamp:      t.now().Unix(),
		ExperimentId:   t.experimentID,
		FeatureId:      t.featureID,
		FeatureVersion: t.featureVersion,
		UserId:         t.userID,
		VariationId:    t.variationID,
		GoalId:         goalID,
		Value:          value,
		User:           &user.User{Id: t.userID, Data: t.attributes},
	}
	t.ch <- experiment
}
