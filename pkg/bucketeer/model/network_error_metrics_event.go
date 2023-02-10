package model

//nolint:lll
const NetworkErrorMetricsEventType metricsDetailEventType = "type.googleapis.com/bucketeer.event.client.NetworkErrorMetricsEvent"

type NetworkErrorMetricsEvent struct {
	APIID  APIID                  `json:"api_id,omitempty"`
	Labels map[string]string      `json:"labels,omitempty"`
	Type   metricsDetailEventType `json:"@type,omitempty"`
}

func NewNetworkErrorMetricsEvent(tag string, api APIID) *NetworkErrorMetricsEvent {
	return &NetworkErrorMetricsEvent{
		APIID:  api,
		Labels: map[string]string{"tag": tag},
		Type:   NetworkErrorMetricsEventType,
	}
}
