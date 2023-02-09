package model

type RegisterEventsRequest struct {
	Events []*Event `json:"events,omitempty"`
}
