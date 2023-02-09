package models

type InternalErrorCountMetricsEvent struct {
	Tag  string                 `json:"tag,omitempty"`
	Type metricsDetailEventType `json:"@type,omitempty"`
}

func NewInternalErrorCountMetricsEvent(tag string) *InternalErrorCountMetricsEvent {
	return &InternalErrorCountMetricsEvent{
		Tag:  tag,
		Type: InternalErrorCountMetricsEventType,
	}
}
