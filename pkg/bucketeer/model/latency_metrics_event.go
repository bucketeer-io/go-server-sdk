package model

type LatencyMetricsEvent struct {
	APIID         APIID                  `json:"apiId,omitempty"`
	Labels        map[string]string      `json:"labels,omitempty"`
	LatencySecond float64                `json:"latencySecond,omitempty"`
	Type          metricsDetailEventType `json:"@type,omitempty"`
}

//nolint:lll
const LatencyMetricsEventType metricsDetailEventType = "type.googleapis.com/bucketeer.event.client.LatencyMetricsEvent"

func NewLatencyMetricsEvent(tag string, second float64, api APIID) *LatencyMetricsEvent {
	return &LatencyMetricsEvent{
		APIID:         api,
		Labels:        map[string]string{"tag": tag},
		LatencySecond: second,
		Type:          LatencyMetricsEventType,
	}
}
