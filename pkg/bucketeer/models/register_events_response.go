package models

type RegisterEventsResponse struct {
	Errors map[string]*RegisterEventsResponseError `json:"errors,omitempty"`
}

type RegisterEventsResponseError struct {
	Retriable bool   `json:"retriable,omitempty"`
	Message   string `json:"message,omitempty"`
}
