package model

type LatencyMetricsEvent struct {
	APIID    APIID                  `json:"api_id,omitempty"`
	Labels   map[string]string      `json:"labels,omitempty"`
	Duration *Duration              `json:"duration,omitempty"`
	Type     metricsDetailEventType `json:"@type,omitempty"`
}

func NewLatencyMetricsEvent(tag, val string, api APIID) *LatencyMetricsEvent {
	return &LatencyMetricsEvent{
		APIID:  api,
		Labels: map[string]string{"tag": tag},
		Duration: &Duration{
			Type:  DurationType,
			Value: val,
		},
		Type: LatencyMetricsEventType,
	}
}
