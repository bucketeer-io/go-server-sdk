package model

type EvaluationDetail[T EvaluationValue] struct {
	FeatureID      string
	FeatureVersion int32
	UserID         string
	VariationID    string
	Reason         EvaluationReason
	Value          T
}

type EvaluationValue interface {
	int | int64 | float64 | string | bool | interface{}
}

type EvaluationReason string

const (
	EvaluationReasonTarget       EvaluationReason = "TARGET"
	EvaluationReasonRule         EvaluationReason = "RULE"
	EvaluationReasonDefault      EvaluationReason = "DEFAULT"
	EvaluationReasonClient       EvaluationReason = "CLIENT"
	EvaluationReasonOffVariation EvaluationReason = "OFF_VARIATION"
	EvaluationReasonPrerequisite EvaluationReason = "PREREQUISITE"
)

func NewEvaluationDetail[T EvaluationValue](
	featureID, userID, variationID string,
	featureVersion int32,
	reason EvaluationReason,
	value T,
) EvaluationDetail[T] {
	return EvaluationDetail[T]{
		FeatureID:      featureID,
		FeatureVersion: featureVersion,
		VariationID:    variationID,
		UserID:         userID,
		Reason:         reason,
		Value:          value,
	}
}

func ConvertEvaluationReason(reasonType ReasonType) EvaluationReason {
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
	default:
		return EvaluationReasonDefault
	}
}
