package model

import (
	"time"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/user"
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
	tag, featureID, variationID, sdkVersion string,
	featureVersion int32,
	sourceID SourceIDType,
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
		SourceID:       sourceID,
		SDKVersion:     sdkVersion,
		Metadata:       map[string]string{},
		Type:           EvaluationEventType,
	}
}
