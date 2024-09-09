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
	EvaluationReasonTarget       EvaluationReason = "TARGET"
	EvaluationReasonRule         EvaluationReason = "RULE"
	EvaluationReasonDefault      EvaluationReason = "DEFAULT"
	EvaluationReasonClient       EvaluationReason = "CLIENT"
	EvaluationReasonOffVariation EvaluationReason = "OFF_VARIATION"
	EvaluationReasonPrerequisite EvaluationReason = "PREREQUISITE"
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
	default:
		return EvaluationReasonDefault
	}
}
