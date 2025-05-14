package model

type Reason struct {
	Type   ReasonType `json:"type,omitempty"`
	RuleID string     `json:"ruleId,omitempty"`
}

type ReasonType string

// Evaluation reasons for feature flag results:
const (
	// Successful evaluations:
	// Evaluated using individual targeting.
	ReasonTarget ReasonType = "TARGET"
	// Evaluated using a custom rule.
	ReasonRule ReasonType = "RULE"
	// Evaluated using the default strategy.
	ReasonDefault ReasonType = "DEFAULT"
	// The flag is missing in the cache; the default value was returned.
	// Deprecated: Use the reason with the ERROR prefix instead.
	ReasonClient ReasonType = "CLIENT"
	// Evaluated using the off variation.
	ReasonOffVariation ReasonType = "OFF_VARIATION"
	// Evaluated using a prerequisite.
	ReasonPrerequisite ReasonType = "PREREQUISITE"
	// Error evaluations:
	// No evaluations were performed.
	ReasonErrorNoEvaluations ReasonType = "ERROR_NO_EVALUATIONS"
	// The specified feature flag was not found.
	ReasonErrorFlagNotFound ReasonType = "ERROR_FLAG_NOT_FOUND"
	// The variation type does not match the expected type.
	ReasonErrorWrongType ReasonType = "ERROR_WRONG_TYPE"
	// User ID was not specified in the evaluation request.
	ReasonErrorUserIDNotSpecified ReasonType = "ERROR_USER_ID_NOT_SPECIFIED"
	// Feature flag ID was not specified in the evaluation request.
	ReasonErrorFeatureFlagIDNotSpecified ReasonType = "ERROR_FEATURE_FLAG_ID_NOT_SPECIFIED"
	// An unexpected error occurred during evaluation.
	ReasonErrorException ReasonType = "ERROR_EXCEPTION"
)
