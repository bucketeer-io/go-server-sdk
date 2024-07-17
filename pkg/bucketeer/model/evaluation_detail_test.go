package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestConvertReason(t *testing.T) {
	t.Parallel()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
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
