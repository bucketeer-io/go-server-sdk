package model

import "strconv"

//nolint:lll
const UnknownErrorMetricsEventType metricsDetailEventType = "type.googleapis.com/bucketeer.event.client.UnknownErrorMetricsEvent"

type UnknownErrorMetricsEvent struct {
	APIID  APIID                  `json:"apiId,omitempty"`
	Labels map[string]string      `json:"labels,omitempty"`
	Type   metricsDetailEventType `json:"@type,omitempty"`
}

func NewUnknownErrorMetricsEvent(tag string, code int, api APIID) *UnknownErrorMetricsEvent {
	return &UnknownErrorMetricsEvent{
		APIID:  api,
		Labels: map[string]string{"tag": tag, "response_code": strconv.Itoa(code)},
		Type:   UnknownErrorMetricsEventType,
	}
}
