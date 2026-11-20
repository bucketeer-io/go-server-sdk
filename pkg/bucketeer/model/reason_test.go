package model

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"

	ftproto "github.com/bucketeer-io/bucketeer/v2/proto/feature"
	gwproto "github.com/bucketeer-io/bucketeer/v2/proto/gateway"
)

// The SDK exchanges the reason with the server by its name,
// and the local evaluation converts the reason of the evaluation module by its name too.
func TestReasonTypeIsTheNameOfTheServerReason(t *testing.T) {
	t.Parallel()
	patterns := []struct {
		reasonType ReasonType
		expected   ftproto.Reason_Type
	}{
		{reasonType: ReasonTarget, expected: ftproto.Reason_TARGET},
		{reasonType: ReasonRule, expected: ftproto.Reason_RULE},
		{reasonType: ReasonDefault, expected: ftproto.Reason_DEFAULT},
		{reasonType: ReasonClient, expected: ftproto.Reason_CLIENT},
		{reasonType: ReasonOffVariation, expected: ftproto.Reason_OFF_VARIATION},
		{reasonType: ReasonPrerequisite, expected: ftproto.Reason_PREREQUISITE},
		{reasonType: ReasonHoldout, expected: ftproto.Reason_HOLDOUT},
	}
	for _, p := range patterns {
		t.Run(string(p.reasonType), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, p.expected.String(), string(p.reasonType))
		})
	}
}

func TestReasonJSON(t *testing.T) {
	t.Parallel()
	patterns := []struct {
		desc     string
		reason   *Reason
		expected string
	}{
		{
			desc:     "holdout",
			reason:   &Reason{Type: ReasonHoldout, HoldoutGroupID: "holdout-group-id"},
			expected: `{"type":"HOLDOUT","holdoutGroupId":"holdout-group-id"}`,
		},
		{
			desc:     "rule: no holdout group ID",
			reason:   &Reason{Type: ReasonRule, RuleID: "rule-id"},
			expected: `{"type":"RULE","ruleId":"rule-id"}`,
		},
		{
			desc:     "default: no holdout group ID",
			reason:   &Reason{Type: ReasonDefault},
			expected: `{"type":"DEFAULT"}`,
		},
	}
	for _, p := range patterns {
		t.Run(p.desc, func(t *testing.T) {
			t.Parallel()
			encoded, err := json.Marshal(p.reason)
			require.NoError(t, err)
			assert.Equal(t, p.expected, string(encoded))

			decoded := &Reason{}
			require.NoError(t, json.Unmarshal(encoded, decoded))
			assert.Equal(t, p.reason, decoded)
		})
	}
}

// In the remote evaluation the server returns the HOLDOUT reason with the holdout group ID.
// The response is encoded as the API gateway encodes it.
func TestGetEvaluationResponseHoldoutFromGatewayJSON(t *testing.T) {
	t.Parallel()
	patterns := []struct {
		desc     string
		reason   *ftproto.Reason
		expected *Reason
	}{
		{
			desc:     "holdout",
			reason:   &ftproto.Reason{Type: ftproto.Reason_HOLDOUT, HoldoutGroupId: "holdout-group-id"},
			expected: &Reason{Type: ReasonHoldout, HoldoutGroupID: "holdout-group-id"},
		},
		{
			desc:     "rule: the holdout group ID is empty",
			reason:   &ftproto.Reason{Type: ftproto.Reason_RULE, RuleId: "rule-id"},
			expected: &Reason{Type: ReasonRule, RuleID: "rule-id"},
		},
	}
	for _, p := range patterns {
		t.Run(p.desc, func(t *testing.T) {
			t.Parallel()
			encoded, err := protojson.MarshalOptions{EmitUnpopulated: true}.Marshal(&gwproto.GetEvaluationResponse{
				Evaluation: &ftproto.Evaluation{
					Id:             "feature-id:3:user-id",
					FeatureId:      "feature-id",
					FeatureVersion: 3,
					UserId:         "user-id",
					VariationId:    "variation-id",
					VariationName:  "variation-name",
					VariationValue: "true",
					Reason:         p.reason,
				},
			})
			require.NoError(t, err)

			actual := &GetEvaluationResponse{}
			require.NoError(t, json.Unmarshal(encoded, actual))
			require.NotNil(t, actual.Evaluation)
			assert.Equal(t, p.expected, actual.Evaluation.Reason)
			assert.Equal(t, "variation-id", actual.Evaluation.VariationID)
			assert.Equal(t, NewEvaluationDetails(
				"feature-id", "user-id", "variation-id", "variation-name", 3, p.expected.Type, true,
			).Reason, EvaluationReason(p.reason.Type.String()))
		})
	}
}
