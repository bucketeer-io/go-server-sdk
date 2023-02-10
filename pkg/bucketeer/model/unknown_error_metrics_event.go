package model

//nolint:lll
const UnknownErrorMetricsEventType metricsDetailEventType = "type.googleapis.com/bucketeer.event.client.UnknownErrorMetricsEvent"

type UnknownErrorMetricsEvent struct {
	APIID  APIID                  `json:"api_id,omitempty"`
	Labels map[string]string      `json:"labels,omitempty"`
	Type   metricsDetailEventType `json:"@type,omitempty"`
}

func NewUnknownErrorMetricsEvent(tag string, api APIID) *UnknownErrorMetricsEvent {
	return &UnknownErrorMetricsEvent{
		APIID:  api,
		Labels: map[string]string{"tag": tag},
		Type:   UnknownErrorMetricsEventType,
	}
}
