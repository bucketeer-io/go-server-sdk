package models

type RegisterEventsRequest struct {
	Events []*Event `json:"events,omitempty"`
}
