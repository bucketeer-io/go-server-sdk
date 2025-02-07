package model

type Clause struct {
	ID        string   `json:"id"`
	Attribute string   `json:"attribute"`
	Operator  string   `json:"operator"`
	Values    []string `json:"values"`
}
