package model

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/version"
)

func TestNewGoalEvent(t *testing.T) {
	t.Parallel()
	e := NewGoalEvent(tag, goalID, version.SDKVersion, 0.2, SourceIDGoServer, newUser(t, id))
	assert.IsType(t, &GoalEvent{}, e)
	assert.Equal(t, tag, e.Tag)
	assert.Equal(t, GoalEventType, e.Type)
	assert.Equal(t, id, e.User.ID)
	assert.Equal(t, goalID, e.GoalID)
	assert.Equal(t, 0.2, e.Value)
	assert.Equal(t, SourceIDGoServer, e.SourceID)
	assert.Equal(t, version.SDKVersion, e.SDKVersion)
}
