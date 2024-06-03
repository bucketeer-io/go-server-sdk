package model

import (
	"encoding/json"
	"time"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/version"
)

type MetricsEvent struct {
	Timestamp  int64             `json:"timestamp,omitempty"`
	Event      json.RawMessage   `json:"event,omitempty"`
	SourceID   SourceIDType      `json:"sourceId,omitempty"`
	SDKVersion string            `json:"sdkVersion,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	Type       EventType         `json:"@type,omitempty"`
}

func NewMetricsEvent(encoded json.RawMessage) *MetricsEvent {
	return &MetricsEvent{
		Timestamp:  time.Now().Unix(),
		Event:      encoded,
		SourceID:   SourceIDGoServer,
		SDKVersion: version.SDKVersion,
		Metadata:   map[string]string{},
		Type:       MetricsEventType,
	}
}

type metricsDetailEventType string

type APIID int

// The API IDs must match the IDs defined on the main repository
// https://github.com/bucketeer-io/bucketeer/blob/main/proto/event/client/event.proto
const (
	GetEvaluation   APIID = 1
	RegisterEvents  APIID = 3
	GetFeatureFlags APIID = 4
	GetSegmentUsers APIID = 5
	SDKGetVariation APIID = 100
)
