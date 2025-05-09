package model

type Reason struct {
	Type   ReasonType `json:"type,omitempty"`
	RuleID string     `json:"ruleId,omitempty"`
}

type ReasonType string

const (
	ReasonTarget  ReasonType = "TARGET"
	ReasonRule    ReasonType = "RULE"
	ReasonDefault ReasonType = "DEFAULT"
	// Deprecated: Do not use.
	ReasonClient                         ReasonType = "CLIENT"
	ReasonOffVariation                   ReasonType = "OFF_VARIATION"
	ReasonPrerequisite                   ReasonType = "PREREQUISITE"
	ReasonErrorNoEvaluations             ReasonType = "ERROR_NO_EVALUATIONS"
	ReasonErrorFlagNotFound              ReasonType = "ERROR_FLAG_NOT_FOUND"
	ReasonErrorWrongType                 ReasonType = "ERROR_WRONG_TYPE"
	ReasonErrorUserIDNotSpecified        ReasonType = "ERROR_USER_ID_NOT_SPECIFIED"
	ReasonErrorFeatureFlagIDNotSpecified ReasonType = "ERROR_FEATURE_FLAG_ID_NOT_SPECIFIED"
	ReasonErrorException                 ReasonType = "ERROR_EXCEPTION"
)
