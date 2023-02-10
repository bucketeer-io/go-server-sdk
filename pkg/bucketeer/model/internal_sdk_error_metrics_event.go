package model

type InternalSDKErrorMetricsEvent struct {
	APIID  APIID                  `json:"api_id,omitempty"`
	Labels map[string]string      `json:"labels,omitempty"`
	Type   metricsDetailEventType `json:"@type,omitempty"`
}

//nolint:lll
const InternalSDKErrorMetricsEventType metricsDetailEventType = "type.googleapis.com/bucketeer.event.client.InternalSdkErrorMetricsEvent"

func NewInternalSDKErrorMetricsEvent(tag string, api APIID) *InternalSDKErrorMetricsEvent {
	return &InternalSDKErrorMetricsEvent{
		APIID:  api,
		Labels: map[string]string{"tag": tag},
		Type:   InternalSDKErrorMetricsEventType,
	}
}
