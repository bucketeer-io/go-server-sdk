package model

type InternalErrorMetricsEvent struct {
	APIID  APIID                  `json:"api_id,omitempty"`
	Labels map[string]string      `json:"labels,omitempty"`
	Type   metricsDetailEventType `json:"@type,omitempty"`
}

func NewInternalErrorMetricsEvent(tag string, api APIID) *InternalErrorMetricsEvent {
	return &InternalErrorMetricsEvent{
		APIID:  api,
		Labels: map[string]string{"tag": tag},
		Type:   InternalErrorMetricsEventType,
	}
}
