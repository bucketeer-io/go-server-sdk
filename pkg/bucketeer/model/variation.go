package model

type Variation struct {
	ID          string `json:"id,omitempty"`
	Value       string `json:"value,omitempty"` // number or even json object
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}
