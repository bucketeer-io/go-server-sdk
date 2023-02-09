package models

import "encoding/json"

type Event struct {
	ID                   string          `json:"id,omitempty"`
	Event                json.RawMessage `json:"event,omitempty"`
	EnvironmentNamespace string          `json:"environmentNamespace,omitempty"`
}

type EventType string

const (
	GoalEventType       EventType = "type.googleapis.com/bucketeer.event.client.GoalEvent"
	EvaluationEventType EventType = "type.googleapis.com/bucketeer.event.client.EvaluationEvent"
	MetricsEventType    EventType = "type.googleapis.com/bucketeer.event.client.MetricsEvent"
)

func NewEvent(id string, encoded []byte) *Event {
	return &Event{
		ID:    id,
		Event: encoded,
	}
}
