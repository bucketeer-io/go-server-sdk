package e2e

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/user"
)

func TestStringVariation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc      string
		user      *user.User
		featureID string
		expected  string
	}{
		{
			desc:      "get Variation by Default Strategy",
			user:      newUser(t, "user-1"),
			featureID: featureIDString,
			expected:  featureIDStringVariation1,
		},
		{
			desc:      "get Variation by Targeting Strategy",
			user:      newUser(t, targetUserID),
			featureID: featureIDString,
			expected:  featureIDStringTargetVariation,
		},
		{
			desc:      "get Variation by Segment user",
			user:      newUser(t, targetSegmentUserID),
			featureID: featureIDString,
			expected:  featureIDStringVariation3,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	sdk := newSDK(t, ctx)
	defer func() {
		// Close
		err := sdk.Close(ctx)
		assert.NoError(t, err)
	}()

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			actual := sdk.StringVariation(ctx, tt.user, tt.featureID, "default")
			assert.Equal(t, tt.expected, actual, "userID: %s, featureID: %s", tt.user.ID, tt.featureID)
		})
	}
}

func TestBoolVariation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc      string
		user      *user.User
		featureID string
		expected  bool
	}{
		{
			desc:      "get Variation by Default Strategy",
			user:      newUser(t, "user-1"),
			featureID: featureIDBoolean,
			expected:  true,
		},
		{
			desc:      "get Variation by Targeting Strategy",
			user:      newUser(t, targetUserID),
			featureID: featureIDBoolean,
			expected:  featureIDBooleanTargetVariation,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	sdk := newSDK(t, ctx)
	defer func() {
		// Close
		err := sdk.Close(ctx)
		assert.NoError(t, err)
	}()

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			actual := sdk.BoolVariation(ctx, tt.user, tt.featureID, false)
			assert.Equal(t, tt.expected, actual, "userID: %s, featureID: %s", tt.user.ID, tt.featureID)
		})
	}
}

func TestIntVariation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc      string
		user      *user.User
		featureID string
		expected  int
	}{
		{
			desc:      "get Variation by Default Strategy",
			user:      newUser(t, "user-1"),
			featureID: featureIDInt,
			expected:  featureIDIntVariation1,
		},
		{
			desc:      "get Variation by Targeting Strategy",
			user:      newUser(t, targetUserID),
			featureID: featureIDInt,
			expected:  featureIDIntTargetVariation,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	sdk := newSDK(t, ctx)
	defer func() {
		// Close
		err := sdk.Close(ctx)
		assert.NoError(t, err)
	}()

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			actual := sdk.IntVariation(ctx, tt.user, tt.featureID, -1)
			assert.Equal(t, tt.expected, actual, "userID: %s, featureID: %s", tt.user.ID, tt.featureID)
		})
	}
}

func TestInt64Variation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc      string
		user      *user.User
		featureID string
		expected  int64
	}{
		{
			desc:      "get Variation by Default Strategy",
			user:      newUser(t, "user-1"),
			featureID: featureIDInt64,
			expected:  featureIDInt64Variation1,
		},
		{
			desc:      "get Variation by Targeting Strategy",
			user:      newUser(t, targetUserID),
			featureID: featureIDInt64,
			expected:  featureIDInt64TargetVariation,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	sdk := newSDK(t, ctx)
	defer func() {
		// Close
		err := sdk.Close(ctx)
		assert.NoError(t, err)
	}()

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			actual := sdk.Int64Variation(ctx, tt.user, tt.featureID, -1000000000)
			assert.Equal(t, tt.expected, actual, "userID: %s, featureID: %s", tt.user.ID, tt.featureID)
		})
	}
}

func TestFloat64Variation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc      string
		user      *user.User
		featureID string
		expected  float64
	}{
		{
			desc:      "get Variation by Default Strategy",
			user:      newUser(t, "user-1"),
			featureID: featureIDFloat,
			expected:  featureIDFloatVariation1,
		},
		{
			desc:      "get Variation by Targeting Strategy",
			user:      newUser(t, targetUserID),
			featureID: featureIDFloat,
			expected:  featureIDFloatTargetVariation,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	sdk := newSDK(t, ctx)
	defer func() {
		// Close
		err := sdk.Close(ctx)
		assert.NoError(t, err)
	}()

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			actual := sdk.Float64Variation(ctx, tt.user, tt.featureID, -1.1)
			assert.Equal(t, tt.expected, actual, "userID: %s, featureID: %s", tt.user.ID, tt.featureID)
		})
	}
}

func TestJSONVariation(t *testing.T) {
	t.Parallel()

	type TestJson struct {
		Str string `json:"str"`
		Int string `json:"int"`
	}

	tests := []struct {
		desc      string
		user      *user.User
		featureID string
		expected  *TestJson
	}{
		{
			desc:      "get Variation by Default Strategy",
			user:      newUser(t, "user-1"),
			featureID: featureIDJson,
			expected:  &TestJson{Str: "str1", Int: "int1"},
		},
		{
			desc:      "get Variation by Targeting Strategy",
			user:      newUser(t, targetUserID),
			featureID: featureIDJson,
			expected:  &TestJson{Str: "str2", Int: "int2"},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	sdk := newSDK(t, ctx)
	defer func() {
		// Close
		err := sdk.Close(ctx)
		assert.NoError(t, err)
	}()

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			v := &TestJson{}
			sdk.JSONVariation(ctx, tt.user, tt.featureID, v)
			assert.Equal(t, tt.expected, v, "userID: %s, featureID: %s", tt.user.ID, tt.featureID)
		})
	}
}

func newSDK(t *testing.T, ctx context.Context) bucketeer.SDK {
	t.Helper()
	sdk, err := bucketeer.NewSDK(
		ctx,
		bucketeer.WithTag(tag),
		bucketeer.WithAPIKey(*apiKey),
		bucketeer.WithHost(*host),
		bucketeer.WithPort(*port),
		bucketeer.WithEventQueueCapacity(100),
		bucketeer.WithNumEventFlushWorkers(3),
		bucketeer.WithEventFlushSize(1),
		bucketeer.WithEnableDebugLog(true),
	)
	assert.NoError(t, err)
	return sdk
}

func newUser(t *testing.T, id string) *user.User {
	t.Helper()
	return user.NewUser(id, map[string]string{"attr-key": "attr-value"})
}
