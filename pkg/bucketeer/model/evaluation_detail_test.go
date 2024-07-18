package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEvaluationDetail(t *testing.T) {
	t.Parallel()
	type DstStruct struct {
		Str string `json:"str"`
		Int string `json:"int"`
	}
	featureID := "feature-id"
	var featureVersion int32 = 1
	variationID := "variation-id"
	userID := "user-id"
	reason := EvaluationReasonClient
	tests := []struct {
		desc          string
		value         interface{}
		expectedValue EvaluationDetail[interface{}]
	}{
		{
			desc:  "valueType: bool",
			value: true,
			expectedValue: EvaluationDetail[interface{}]{
				FeatureID:      featureID,
				FeatureVersion: featureVersion,
				UserID:         userID,
				VariationID:    variationID,
				Reason:         reason,
				Value:          true,
			},
		},
		{
			desc:  "valueType: int",
			value: 100,
			expectedValue: EvaluationDetail[interface{}]{
				FeatureID:      featureID,
				FeatureVersion: featureVersion,
				UserID:         userID,
				VariationID:    variationID,
				Reason:         reason,
				Value:          100,
			},
		},
		{
			desc:  "valueType: string",
			value: "value",
			expectedValue: EvaluationDetail[interface{}]{
				FeatureID:      featureID,
				FeatureVersion: featureVersion,
				UserID:         userID,
				VariationID:    variationID,
				Reason:         reason,
				Value:          "value",
			},
		},
		{
			desc:  "valueType: json",
			value: &DstStruct{Str: "str", Int: "int"},
			expectedValue: EvaluationDetail[interface{}]{
				FeatureID:      featureID,
				FeatureVersion: featureVersion,
				UserID:         userID,
				VariationID:    variationID,
				Reason:         reason,
				Value:          &DstStruct{Str: "str", Int: "int"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			actual := NewEvaluationDetail(featureID, userID, variationID, featureVersion, reason, tt.value)
			assert.Equal(t, tt.expectedValue.Value, actual.Value)
			assert.Equal(t, featureID, actual.FeatureID)
			assert.Equal(t, featureVersion, actual.FeatureVersion)
			assert.Equal(t, userID, actual.UserID)
			assert.Equal(t, variationID, actual.VariationID)
			assert.Equal(t, reason, actual.Reason)
		})
	}
}
func TestConvertReason(t *testing.T) {
	t.Parallel()
	tests := []struct {
		desc          string
		reasonType    ReasonType
		expectedValue EvaluationReason
	}{
		{
			desc:          "reasonType: ReasonTarget",
			reasonType:    ReasonTarget,
			expectedValue: EvaluationReasonTarget,
		},
		{
			desc:          "reasonType: ReasonClient",
			reasonType:    ReasonClient,
			expectedValue: EvaluationReasonClient,
		},
		{
			desc:          "reasonType: ReasonDefault",
			reasonType:    ReasonDefault,
			expectedValue: EvaluationReasonDefault,
		},
		{
			desc:          "reasonType: ReasonPrerequisite",
			reasonType:    ReasonPrerequisite,
			expectedValue: EvaluationReasonPrerequisite,
		},
		{
			desc:          "reasonType: ReasonOffVariation",
			reasonType:    ReasonOffVariation,
			expectedValue: EvaluationReasonOffVariation,
		},
		{
			desc:          "reasonType: ReasonRule",
			reasonType:    ReasonRule,
			expectedValue: EvaluationReasonRule,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			actual := ConvertEvaluationReason(tt.reasonType)
			assert.Equal(t, tt.expectedValue, actual)
		})
	}
}
