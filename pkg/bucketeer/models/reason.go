package models

type Reason struct {
	Type   ReasonType `json:"type,omitempty"`
	RuleID string     `json:"ruleId,omitempty"`
}

type ReasonType string

const (
	ReasonClient ReasonType = "CLIENT"
)
