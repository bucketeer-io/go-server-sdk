package models

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/version"
)

func TestNewGoalEvent(t *testing.T) {
	t.Parallel()
	e := NewGoalEvent(tag, goalID, 0.2, newUser(t, id))
	assert.IsType(t, &GoalEvent{}, e)
	assert.Equal(t, tag, e.Tag)
	assert.Equal(t, GoalEventType, e.Type)
	assert.Equal(t, id, e.User.ID)
	assert.Equal(t, goalID, e.GoalID)
	assert.Equal(t, 0.2, e.Value)
	assert.Equal(t, e.SourceID, SourceIDGoServer)
	assert.Equal(t, version.SDKVersion, e.SDKVersion)
}
