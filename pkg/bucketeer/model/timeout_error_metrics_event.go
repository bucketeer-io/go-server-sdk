package model

type TimeoutErrorMetricsEvent struct {
	APIID  APIID                  `json:"api_id,omitempty"`
	Labels map[string]string      `json:"labels,omitempty"`
	Type   metricsDetailEventType `json:"@type,omitempty"`
}

func NewTimeoutErrorMetricsEvent(tag string, api APIID) *TimeoutErrorMetricsEvent {
	return &TimeoutErrorMetricsEvent{
		APIID:  api,
		Labels: map[string]string{"tag": tag},
		Type:   TimeoutErrorMetricsEventType,
	}
}
