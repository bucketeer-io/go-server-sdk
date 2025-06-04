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
	variationName := "variation-name"
	userID := "user-id"
	reasonType := ReasonErrorException
	tests := []struct {
		desc          string
		value         interface{}
		expectedValue BKTEvaluationDetails[interface{}]
	}{
		{
			desc:  "valueType: bool",
			value: true,
			expectedValue: BKTEvaluationDetails[interface{}]{
				FeatureID:      featureID,
				FeatureVersion: featureVersion,
				UserID:         userID,
				VariationID:    variationID,
				Reason:         EvaluationReasonErrorException,
				VariationName:  variationName,
				VariationValue: true,
			},
		},
		{
			desc:  "valueType: int",
			value: 100,
			expectedValue: BKTEvaluationDetails[interface{}]{
				FeatureID:      featureID,
				FeatureVersion: featureVersion,
				UserID:         userID,
				VariationID:    variationID,
				Reason:         EvaluationReasonErrorException,
				VariationName:  variationName,
				VariationValue: 100,
			},
		},
		{
			desc:  "valueType: string",
			value: "value",
			expectedValue: BKTEvaluationDetails[interface{}]{
				FeatureID:      featureID,
				FeatureVersion: featureVersion,
				UserID:         userID,
				VariationID:    variationID,
				Reason:         EvaluationReasonErrorException,
				VariationName:  variationName,
				VariationValue: "value",
			},
		},
		{
			desc:  "valueType: json",
			value: &DstStruct{Str: "str", Int: "int"},
			expectedValue: BKTEvaluationDetails[interface{}]{
				FeatureID:      featureID,
				FeatureVersion: featureVersion,
				UserID:         userID,
				VariationID:    variationID,
				Reason:         EvaluationReasonErrorException,
				VariationName:  variationName,
				VariationValue: &DstStruct{Str: "str", Int: "int"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			actual := NewEvaluationDetails(featureID, userID, variationID, variationName, featureVersion, reasonType, tt.value)
			assert.Equal(t, tt.expectedValue.VariationValue, actual.VariationValue)
			assert.Equal(t, featureID, actual.FeatureID)
			assert.Equal(t, featureVersion, actual.FeatureVersion)
			assert.Equal(t, userID, actual.UserID)
			assert.Equal(t, variationID, actual.VariationID)
			assert.Equal(t, variationName, actual.VariationName)
		})
	}
}

func TestConvertEvaluationReason(t *testing.T) {
	t.Parallel()
	featureID := "feature-id"
	var featureVersion int32 = 1
	variationID := "variation-id"
	variationName := "variation-name"
	userID := "user-id"
	value := true
	tests := []struct {
		desc           string
		reasonType     ReasonType
		expectedReason EvaluationReason
	}{
		{
			desc:           "reasonType: ReasonTarget",
			reasonType:     ReasonTarget,
			expectedReason: EvaluationReasonTarget,
		},
		{
			desc:           "reasonType: ReasonClient",
			reasonType:     ReasonClient,
			expectedReason: EvaluationReasonClient,
		},
		{
			desc:           "reasonType: ReasonDefault",
			reasonType:     ReasonDefault,
			expectedReason: EvaluationReasonDefault,
		},
		{
			desc:           "reasonType: ReasonPrerequisite",
			reasonType:     ReasonPrerequisite,
			expectedReason: EvaluationReasonPrerequisite,
		},
		{
			desc:           "reasonType: ReasonOffVariation",
			reasonType:     ReasonOffVariation,
			expectedReason: EvaluationReasonOffVariation,
		},
		{
			desc:           "reasonType: ReasonRule",
			reasonType:     ReasonRule,
			expectedReason: EvaluationReasonRule,
		},
		{
			desc:           "reasonType: ReasonErrorNoEvaluations",
			reasonType:     ReasonErrorNoEvaluations,
			expectedReason: EvaluationReasonErrorNoEvaluations,
		},
		{
			desc:           "reasonType: ReasonErrorFlagNotFound",
			reasonType:     ReasonErrorFlagNotFound,
			expectedReason: EvaluationReasonErrorFlagNotFound,
		},
		{
			desc:           "reasonType: ReasonErrorWrongType",
			reasonType:     ReasonErrorWrongType,
			expectedReason: EvaluationReasonErrorWrongType,
		},
		{
			desc:           "reasonType: ReasonErrorUserIDNotSpecified",
			reasonType:     ReasonErrorUserIDNotSpecified,
			expectedReason: EvaluationReasonErrorUserIDNotSpecified,
		},
		{
			desc:           "reasonType: ReasonErrorFeatureFlagIDNotSpecified",
			reasonType:     ReasonErrorFeatureFlagIDNotSpecified,
			expectedReason: EvaluationReasonErrorFeatureFlagIDNotSpecified,
		},
		{
			desc:           "reasonType: ReasonErrorException",
			reasonType:     ReasonErrorException,
			expectedReason: EvaluationReasonErrorException,
		},
		{
			desc:           "reasonType: ReasonErrorCacheNotFound",
			reasonType:     ReasonErrorCacheNotFound,
			expectedReason: EvaluationReasonErrorCacheNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			t.Parallel()
			actual := NewEvaluationDetails(featureID, userID, variationID, variationName, featureVersion, tt.reasonType, value)
			assert.Equal(t, tt.expectedReason, actual.Reason)
		})
	}
}
