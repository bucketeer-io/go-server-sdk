package model

import (
	"encoding/json"
	"time"

	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/version"
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

//nolint:lll
const (
	LatencyMetricsEventType       metricsDetailEventType = "type.googleapis.com/bucketeer.event.client.LatencyMetricsEvent"
	SizeMetricsEventType          metricsDetailEventType = "type.googleapis.com/bucketeer.event.client.SizeMetricsEvent"
	TimeoutErrorMetricsEventType  metricsDetailEventType = "type.googleapis.com/bucketeer.event.client.TimeoutErrorMetricsEvent"
	InternalErrorMetricsEventType metricsDetailEventType = "type.googleapis.com/bucketeer.event.client.InternalErrorMetricsEvent"
)

type APIID int

const (
	GetEvaluation  APIID = 1
	RegisterEvents APIID = 3
)
