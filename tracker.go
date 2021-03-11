package bucketeer

import (
	"time"

	event "github.com/ca-dp/bucketeer-go-server-sdk/proto/event/client"
	"github.com/ca-dp/bucketeer-go-server-sdk/proto/feature"
	"github.com/ca-dp/bucketeer-go-server-sdk/proto/user"
	"github.com/golang/protobuf/ptypes"
	"github.com/google/uuid"
)

type Tracker interface {
	Track(goalID string, value float64)
}

type tracker struct {
	ch          chan<- *event.Event
	now         func() time.Time
	userID      string
	attributes  map[string]string
	evaluations []*feature.Evaluation
	logFunc     func(error)
}

func (t *tracker) Track(goalID string, value float64) {
	if t == nil {
		return
	}
	t.sendGoalEvent(goalID, value)
}

func (t *tracker) sendGoalEvent(goalID string, value float64) {
	goal := &event.GoalEvent{
		Timestamp:   t.now().Unix(),
		GoalId:      goalID,
		UserId:      t.userID,
		Value:       value,
		User:        &user.User{Id: t.userID, Data: t.attributes},
		Evaluations: t.evaluations,
	}
	any, err := ptypes.MarshalAny(goal)
	if err == nil {
		t.ch <- &event.Event{
			Id:    uuid.New().String(),
			Event: any,
		}
		return
	}
	t.log(&Error{
		Message:        "failed to marshal goal event",
		UserID:         t.userID,
		UserAttributes: t.attributes,
		Goal:           goal,
		originalErr:    err,
	})
}

func (t *tracker) log(err error) {
	if t.logFunc == nil {
		return
	}
	t.logFunc(err)
}
