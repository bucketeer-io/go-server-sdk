package model

type SizeMetricsEvent struct {
	APIID    APIID                  `json:"api_id,omitempty"`
	Labels   map[string]string      `json:"labels,omitempty"`
	SizeByte int32                  `json:"sizeByte,omitempty"`
	Type     metricsDetailEventType `json:"@type,omitempty"`
}

func NewSizeMetricsEvent(tag string, sizeByte int32, api APIID) *SizeMetricsEvent {
	return &SizeMetricsEvent{
		APIID:    api,
		Labels:   map[string]string{"tag": tag},
		SizeByte: sizeByte,
		Type:     SizeMetricsEventType,
	}
}
