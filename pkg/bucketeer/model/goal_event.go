package models

import (
	"time"

	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/user"

	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/version"
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

func NewGoalEvent(tag, goalID string, value float64, user *user.User) *GoalEvent {
	return &GoalEvent{
		Tag:        tag,
		Timestamp:  time.Now().Unix(),
		GoalID:     goalID,
		UserID:     user.ID,
		Value:      value,
		User:       user,
		SourceID:   SourceIDGoServer,
		SDKVersion: version.SDKVersion,
		Metadata:   map[string]string{},
		Type:       GoalEventType,
	}
}
