package model

type TimeoutErrorCountMetricsEvent struct {
	Tag  string                 `json:"tag,omitempty"`
	Type metricsDetailEventType `json:"@type,omitempty"`
}

func NewTimeoutErrorCountMetricsEvent(tag string) *TimeoutErrorCountMetricsEvent {
	return &TimeoutErrorCountMetricsEvent{
		Tag:  tag,
		Type: TimeoutErrorCountMetricsEventType,
	}
}
