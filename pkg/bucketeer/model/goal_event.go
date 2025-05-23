package model

import (
	"time"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/user"
)

type GoalEvent struct {
	Timestamp  int64             `json:"timestamp,omitempty"`
	GoalID     string            `json:"goalId,omitempty"`
	UserID     string            `json:"userId,omitempty"`
	Value      float64           `json:"value,omitempty"`
	User       *user.User        `json:"user,omitempty"`
	Tag        string            `json:"tag,omitempty"`
	SourceID   SourceIDType      `json:"sourceId,omitempty"`
	SDKVersion string            `json:"sdkVersion,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	Type       EventType         `json:"@type,omitempty"`
}

func NewGoalEvent(
	tag, goalID, sdkVersion string,
	value float64,
	sourceID SourceIDType,
	user *user.User,
) *GoalEvent {
	return &GoalEvent{
		Tag:        tag,
		Timestamp:  time.Now().Unix(),
		GoalID:     goalID,
		UserID:     user.ID,
		Value:      value,
		User:       user,
		SourceID:   sourceID,
		SDKVersion: sdkVersion,
		Metadata:   map[string]string{},
		Type:       GoalEventType,
	}
}
