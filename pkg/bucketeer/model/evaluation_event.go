package model

import (
	"time"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/user"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/version"
)

type EvaluationEvent struct {
	Timestamp      int64             `json:"timestamp,omitempty"`
	FeatureID      string            `json:"featureId,omitempty"`
	FeatureVersion int32             `json:"featureVersion,omitempty"`
	VariationID    string            `json:"variationId,omitempty"`
	User           *user.User        `json:"user,omitempty"`
	Reason         *Reason           `json:"reason,omitempty"`
	Tag            string            `json:"tag,omitempty"`
	SourceID       SourceIDType      `json:"sourceId,omitempty"`
	SDKVersion     string            `json:"sdkVersion,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	Type           EventType         `json:"@type,omitempty"`
}

func NewEvaluationEvent(
	tag, featureID, variationID string,
	featureVersion int32,
	user *user.User,
	reason *Reason,
) *EvaluationEvent {
	return &EvaluationEvent{
		Tag:            tag,
		Timestamp:      time.Now().Unix(),
		FeatureID:      featureID,
		FeatureVersion: featureVersion,
		VariationID:    variationID,
		User:           user,
		Reason:         reason,
		SourceID:       SourceIDGoServer,
		SDKVersion:     version.SDKVersion,
		Metadata:       map[string]string{},
		Type:           EvaluationEventType,
	}
}
