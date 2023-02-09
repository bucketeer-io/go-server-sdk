package model

type GetEvaluationLatencyMetricsEvent struct {
	Labels   map[string]string      `json:"labels,omitempty"`
	Duration *Duration              `json:"duration,omitempty"`
	Type     metricsDetailEventType `json:"@type,omitempty"`
}

func NewGetEvaluationLatencyMetricsEvent(tag, val string) *GetEvaluationLatencyMetricsEvent {
	return &GetEvaluationLatencyMetricsEvent{
		Labels: map[string]string{"tag": tag},
		Duration: &Duration{
			Type:  DurationType,
			Value: val,
		},
		Type: GetEvaluationLatencyMetricsEventType,
	}
}
