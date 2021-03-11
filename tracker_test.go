package bucketeer

import (
	"errors"
	"testing"
	"time"

	gofakeit "github.com/brianvoe/gofakeit/v6"
	event "github.com/ca-dp/bucketeer-go-server-sdk/proto/event/client"
	"github.com/ca-dp/bucketeer-go-server-sdk/proto/feature"
	"github.com/ca-dp/bucketeer-go-server-sdk/proto/user"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/require"
)

func Test_tracker_Track(t *testing.T) {
	const (
		goalID = "goal-id"
		userID = "user-id"
		value  = 1.2345
	)
	attrs := map[string]string{
		"class": "a",
	}
	evaluation := &feature.Evaluation{}
	gofakeit.Struct(evaluation)
	now := time.Now()
	goalEvent := &event.GoalEvent{
		Timestamp:   now.Unix(),
		GoalId:      goalID,
		UserId:      userID,
		Value:       value,
		User:        &user.User{Id: userID, Data: attrs},
		Evaluations: []*feature.Evaluation{evaluation},
	}
	expect, err := ptypes.MarshalAny(goalEvent)
	require.NoError(t, err)

	// test: track is nil
	var track *tracker
	require.NotPanics(t, func() {
		track.Track(goalID, value)
	})

	// test: track
	ch := make(chan *event.Event, 1)
	track = &tracker{
		ch: ch,
		now: func() time.Time {
			return now
		},
		userID:      "user-id",
		attributes:  attrs,
		evaluations: []*feature.Evaluation{evaluation},
		logFunc:     nil,
	}
	track.Track(goalID, value)
	require.Len(t, track.ch, 1)
	e := <-ch
	require.NotEmpty(t, e.Id)
	require.Equal(t, expect, e.Event)

}

func Test_tracker_log(t *testing.T) {
	var called bool
	track := &tracker{
		logFunc: func(err error) {
			called = true
		},
	}
	fakeErr := errors.New("error")

	// test: success
	track.log(fakeErr)
	require.True(t, called)

	// test: logFunc is nil
	track.logFunc = nil
	require.NotPanics(t, func() {
		track.log(fakeErr)
	})
}
