package model

type Reason struct {
	Type   ReasonType `json:"type,omitempty"`
	RuleID string     `json:"ruleId,omitempty"`
}

type ReasonType string

const (
	ReasonTarget       ReasonType = "TARGET"
	ReasonRule         ReasonType = "RULE"
	ReasonDefault      ReasonType = "DEFAULT"
	ReasonClient       ReasonType = "CLIENT"
	ReasonOffVariation ReasonType = "OFF_VARIATION"
	ReasonPrerequisite ReasonType = "PREREQUISITE"
)
