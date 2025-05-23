package model

import (
	"encoding/json"
	"time"
)

type MetricsEvent struct {
	Timestamp  int64             `json:"timestamp,omitempty"`
	Event      json.RawMessage   `json:"event,omitempty"`
	SourceID   SourceIDType      `json:"sourceId,omitempty"`
	SDKVersion string            `json:"sdkVersion,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	Type       EventType         `json:"@type,omitempty"`
}

func NewMetricsEvent(
	encoded json.RawMessage,
	sourceID SourceIDType,
	sdkVersion string,
) *MetricsEvent {
	return &MetricsEvent{
		Timestamp:  time.Now().Unix(),
		Event:      encoded,
		SourceID:   sourceID,
		SDKVersion: sdkVersion,
		Metadata:   map[string]string{},
		Type:       MetricsEventType,
	}
}

type metricsDetailEventType string

type APIID int

// The API IDs must match the IDs defined on the main repository
// https://github.com/bucketeer-io/bucketeer/blob/main/proto/event/client/event.proto
const (
	GetEvaluation    APIID = 1
	RegisterEvents   APIID = 3
	GetFeatureFlags  APIID = 4
	GetSegmentUsers  APIID = 5
	SDKGetEvaluation APIID = 100
)
