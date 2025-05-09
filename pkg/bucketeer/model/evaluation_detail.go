package model

type BKTEvaluationDetails[T EvaluationValue] struct {
	FeatureID      string
	FeatureVersion int32
	UserID         string
	VariationID    string
	VariationName  string
	VariationValue T
	Reason         EvaluationReason
}

type EvaluationValue interface {
	int | int64 | float64 | string | bool | interface{}
}

type EvaluationReason string

const (
	EvaluationReasonTarget  EvaluationReason = "TARGET"
	EvaluationReasonRule    EvaluationReason = "RULE"
	EvaluationReasonDefault EvaluationReason = "DEFAULT"
	// Deprecated: Use the reason with the ERROR prefix instead.
	EvaluationReasonClient                         EvaluationReason = "CLIENT"
	EvaluationReasonOffVariation                   EvaluationReason = "OFF_VARIATION"
	EvaluationReasonPrerequisite                   EvaluationReason = "PREREQUISITE"
	EvaluationReasonErrorNoEvaluations             EvaluationReason = "ERROR_NO_EVALUATIONS"
	EvaluationReasonErrorFlagNotFound              EvaluationReason = "ERROR_FLAG_NOT_FOUND"
	EvaluationReasonErrorWrongType                 EvaluationReason = "ERROR_WRONG_TYPE"
	EvaluationReasonErrorUserIDNotSpecified        EvaluationReason = "ERROR_USER_ID_NOT_SPECIFIED"
	EvaluationReasonErrorFeatureFlagIDNotSpecified EvaluationReason = "ERROR_FEATURE_FLAG_ID_NOT_SPECIFIED"
	EvaluationReasonErrorException                 EvaluationReason = "ERROR_EXCEPTION"
)

func NewEvaluationDetails[T EvaluationValue](
	featureID, userID, variationID, variationName string,
	featureVersion int32,
	reasonType ReasonType,
	value T,
) BKTEvaluationDetails[T] {
	return BKTEvaluationDetails[T]{
		FeatureID:      featureID,
		FeatureVersion: featureVersion,
		UserID:         userID,
		VariationID:    variationID,
		VariationName:  variationName,
		VariationValue: value,
		Reason:         convertEvaluationReason(reasonType),
	}
}

func convertEvaluationReason(reasonType ReasonType) EvaluationReason {
	switch reasonType {
	case ReasonTarget:
		return EvaluationReasonTarget
	case ReasonRule:
		return EvaluationReasonRule
	case ReasonDefault:
		return EvaluationReasonDefault
	case ReasonClient:
		return EvaluationReasonClient
	case ReasonOffVariation:
		return EvaluationReasonOffVariation
	case ReasonPrerequisite:
		return EvaluationReasonPrerequisite
	case ReasonErrorNoEvaluations:
		return EvaluationReasonErrorNoEvaluations
	case ReasonErrorFlagNotFound:
		return EvaluationReasonErrorFlagNotFound
	case ReasonErrorWrongType:
		return EvaluationReasonErrorWrongType
	case ReasonErrorUserIDNotSpecified:
		return EvaluationReasonErrorUserIDNotSpecified
	case ReasonErrorFeatureFlagIDNotSpecified:
		return EvaluationReasonErrorFeatureFlagIDNotSpecified
	case ReasonErrorException:
		return EvaluationReasonErrorException
	default:
		return EvaluationReasonDefault
	}
}
