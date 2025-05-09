package model

type Reason struct {
	Type   ReasonType `json:"type,omitempty"`
	RuleID string     `json:"ruleId,omitempty"`
}

type ReasonType string

// Evaluation reasons for feature flag results:
const (
	// Successful evaluations:
	ReasonTarget  ReasonType = "TARGET"  // Evaluated using individual targeting.
	ReasonRule    ReasonType = "RULE"    // Evaluated using a custom rule.
	ReasonDefault ReasonType = "DEFAULT" // Evaluated using the default strategy.
	// Deprecated: Use the reason with the ERROR prefix instead.
	ReasonClient       ReasonType = "CLIENT"        // The flag is missing in the cache; the default value was returned.
	ReasonOffVariation ReasonType = "OFF_VARIATION" // Evaluated using the off variation.
	ReasonPrerequisite ReasonType = "PREREQUISITE"  // Evaluated using a prerequisite.
	// Error evaluations:
	ReasonErrorNoEvaluations             ReasonType = "ERROR_NO_EVALUATIONS"                // No evaluations were performed.
	ReasonErrorFlagNotFound              ReasonType = "ERROR_FLAG_NOT_FOUND"                // The specified feature flag was not found.
	ReasonErrorWrongType                 ReasonType = "ERROR_WRONG_TYPE"                    // The variation type does not match the expected type.
	ReasonErrorUserIDNotSpecified        ReasonType = "ERROR_USER_ID_NOT_SPECIFIED"         // User ID was not specified in the evaluation request.
	ReasonErrorFeatureFlagIDNotSpecified ReasonType = "ERROR_FEATURE_FLAG_ID_NOT_SPECIFIED" // Feature flag ID was not specified in the evaluation request.
	ReasonErrorException                 ReasonType = "ERROR_EXCEPTION"                     // An unexpected error occurred during evaluation.
)
