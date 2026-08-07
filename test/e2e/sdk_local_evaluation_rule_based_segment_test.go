package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/user"
)

// waitForLocalCacheReady polls the local evaluation path until the SDK's feature flag
// and segment users caches are ready, so the tests proceed as soon as possible instead
// of sleeping for a fixed duration. While either cache is not ready yet, the local
// evaluation fails and the SDK returns the given default value, which no fixture
// variation uses.
func waitForLocalCacheReady(ctx context.Context, t *testing.T, sdk bucketeer.SDK, u *user.User, featureID string) {
	t.Helper()
	assert.Eventually(t, func() bool {
		return sdk.StringVariation(ctx, u, featureID, "default") != "default"
	}, 15*time.Second, 200*time.Millisecond, "timed out waiting for the local cache updates")
}

// TestLocalRuleBasedSegmentMultipleRules verifies the segment rule evaluation semantics
// using the local evaluation path: rules are OR-ed, and the clauses within a rule are AND-ed.
// See main_test.go for the segment and flag fixtures configuration.
func TestLocalRuleBasedSegmentMultipleRules(t *testing.T) {
	t.Parallel()
	tests := []struct {
		desc     string
		user     *user.User
		expected string
	}{
		{
			desc: "match rule 1: country equals AND age within the greater/less bounds",
			user: user.NewUser("rule-based-user-1", map[string]string{
				"country": "japan",
				"age":     "30",
			}),
			expected: ruleBasedSegmentMatchedVariation,
		},
		{
			desc: "match rule 2: email starts-with AND plan in",
			user: user.NewUser("rule-based-user-2", map[string]string{
				"email": "test@bucketeer.io",
				"plan":  "premium",
			}),
			expected: ruleBasedSegmentMatchedVariation,
		},
		{
			desc: "no match: age is out of the less bound (clauses are AND-ed)",
			user: user.NewUser("rule-based-user-3", map[string]string{
				"country": "japan",
				"age":     "65",
			}),
			expected: ruleBasedSegmentDefaultVariation,
		},
		{
			desc: "no match: plan is not in the values (clauses are AND-ed)",
			user: user.NewUser("rule-based-user-4", map[string]string{
				"email": "test@bucketeer.io",
				"plan":  "free",
			}),
			expected: ruleBasedSegmentDefaultVariation,
		},
		{
			desc: "no match: the rule's attribute is missing entirely",
			user: user.NewUser("rule-based-user-5", map[string]string{
				"country": "japan",
				// The age attribute required by rule 1 is missing,
				// and no attribute required by rule 2 is set
			}),
			expected: ruleBasedSegmentDefaultVariation,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	sdk := newLocalSDK(t, ctx)
	defer func() {
		// Close
		err := sdk.Close(ctx)
		assert.NoError(t, err)
	}()

	waitForLocalCacheReady(ctx, t, sdk, tests[0].user, featureIDRuleBasedSegment)

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			actual := sdk.StringVariation(ctx, tt.user, featureIDRuleBasedSegment, "default")
			assert.Equal(t, tt.expected, actual, "userID: %s, featureID: %s", tt.user.ID, featureIDRuleBasedSegment)
		})
	}
}

// TestLocalRuleBasedSegmentMixedListAndRules verifies that a user belongs to a mixed segment
// (uploaded user list AND rules) when the user is in the list OR matches any rule.
func TestLocalRuleBasedSegmentMixedListAndRules(t *testing.T) {
	t.Parallel()
	tests := []struct {
		desc     string
		user     *user.User
		expected string
	}{
		{
			desc: "match by the uploaded user list only (attributes don't match the rules)",
			user: user.NewUser(ruleBasedSegmentListedUserID, map[string]string{
				"country": "france",
			}),
			expected: ruleBasedSegmentMatchedVariation,
		},
		{
			desc: "match by the rules only (user is not in the uploaded user list)",
			user: user.NewUser("rule-based-user-not-listed", map[string]string{
				"country": "japan",
				"age":     "25",
			}),
			expected: ruleBasedSegmentMatchedVariation,
		},
		{
			desc: "no match: user is not in the uploaded user list and doesn't match the rules",
			user: user.NewUser("rule-based-user-no-match", map[string]string{
				"country": "france",
			}),
			expected: ruleBasedSegmentDefaultVariation,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	sdk := newLocalSDK(t, ctx)
	defer func() {
		// Close
		err := sdk.Close(ctx)
		assert.NoError(t, err)
	}()

	waitForLocalCacheReady(ctx, t, sdk, tests[0].user, featureIDRuleBasedSegment)

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			actual := sdk.StringVariation(ctx, tt.user, featureIDRuleBasedSegment, "default")
			assert.Equal(t, tt.expected, actual, "userID: %s, featureID: %s", tt.user.ID, featureIDRuleBasedSegment)
		})
	}
}

// TestLocalRuleBasedSegmentWithAttributeClause verifies a flag rule that combines
// a SEGMENT clause with an additional attribute clause in the same rule
// (segment membership AND region equals "tokyo").
func TestLocalRuleBasedSegmentWithAttributeClause(t *testing.T) {
	t.Parallel()
	tests := []struct {
		desc     string
		user     *user.User
		expected string
	}{
		{
			desc: "match: in the segment by rules AND the region attribute matches",
			user: user.NewUser("segment-attribute-user-1", map[string]string{
				"country": "japan",
				"age":     "30",
				"region":  "tokyo",
			}),
			expected: ruleBasedSegmentMatchedVariation,
		},
		{
			desc: "no match: in the segment by rules but the region attribute doesn't match",
			user: user.NewUser("segment-attribute-user-2", map[string]string{
				"country": "japan",
				"age":     "30",
				"region":  "osaka",
			}),
			expected: ruleBasedSegmentDefaultVariation,
		},
		{
			desc: "no match: the region attribute matches but the user is not in the segment",
			user: user.NewUser("segment-attribute-user-3", map[string]string{
				"country": "france",
				"region":  "tokyo",
			}),
			expected: ruleBasedSegmentDefaultVariation,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	sdk := newLocalSDK(t, ctx)
	defer func() {
		// Close
		err := sdk.Close(ctx)
		assert.NoError(t, err)
	}()

	waitForLocalCacheReady(ctx, t, sdk, tests[0].user, featureIDSegmentAndAttribute)

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			actual := sdk.StringVariation(ctx, tt.user, featureIDSegmentAndAttribute, "default")
			assert.Equal(t, tt.expected, actual, "userID: %s, featureID: %s", tt.user.ID, featureIDSegmentAndAttribute)
		})
	}
}

// TestLocalListOnlySegmentBackwardCompatibility verifies that a segment configured
// only with an uploaded user list (no rules) still evaluates as before.
func TestLocalListOnlySegmentBackwardCompatibility(t *testing.T) {
	t.Parallel()
	tests := []struct {
		desc     string
		user     *user.User
		expected string
	}{
		{
			desc:     "match by the uploaded user list",
			user:     newUser(t, targetSegmentUserID),
			expected: featureIDStringVariation3,
		},
		{
			desc:     "no match: user is not in the uploaded user list",
			user:     newUser(t, "list-only-segment-user-no-match"),
			expected: featureIDStringVariation1,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	sdk := newLocalSDK(t, ctx)
	defer func() {
		// Close
		err := sdk.Close(ctx)
		assert.NoError(t, err)
	}()

	waitForLocalCacheReady(ctx, t, sdk, tests[0].user, featureIDString)

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			actual := sdk.StringVariation(ctx, tt.user, featureIDString, "default")
			assert.Equal(t, tt.expected, actual, "userID: %s, featureID: %s", tt.user.ID, featureIDString)
		})
	}
}
