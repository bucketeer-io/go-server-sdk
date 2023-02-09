package model

type GetEvaluationSizeMetricsEvent struct {
	Labels   map[string]string      `json:"labels,omitempty"`
	SizeByte int32                  `json:"sizeByte,omitempty"`
	Type     metricsDetailEventType `json:"@type,omitempty"`
}

func NewGetEvaluationSizeMetricsEvent(tag string, sizeByte int32) *GetEvaluationSizeMetricsEvent {
	return &GetEvaluationSizeMetricsEvent{
		Labels:   map[string]string{"tag": tag},
		SizeByte: sizeByte,
		Type:     GetEvaluationSizeMetricsEventType,
	}
}
