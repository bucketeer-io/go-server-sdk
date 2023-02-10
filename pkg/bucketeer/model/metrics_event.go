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
	GetEvaluationLatencyMetricsEventType metricsDetailEventType = "type.googleapis.com/bucketeer.event.client.GetEvaluationLatencyMetricsEvent"
	GetEvaluationSizeMetricsEventType    metricsDetailEventType = "type.googleapis.com/bucketeer.event.client.GetEvaluationSizeMetricsEvent"
	TimeoutErrorCountMetricsEventType    metricsDetailEventType = "type.googleapis.com/bucketeer.event.client.TimeoutErrorCountMetricsEvent"
	InternalErrorMetricsEventType        metricsDetailEventType = "type.googleapis.com/bucketeer.event.client.InternalErrorMetricsEvent"
)

type APIID int

const (
	UnknownAPI APIID = iota
	GetEvaluation
	GetEvaluations
	RegisterEvents
)
