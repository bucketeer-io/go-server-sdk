package e2e

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer"
	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/user"
)

func TestStringVariation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	sdk := newSDK(t, ctx)
	defer func() {
		// Close
		err := sdk.Close(ctx)
		assert.NoError(t, err)
	}()
	var wg sync.WaitGroup
	testVariationFunc := func(ctx context.Context, userID, featureID, expectedVariation string) {
		defer wg.Done()
		u := newUser(t, userID)
		v := sdk.StringVariation(ctx, u, featureID, "default")
		assert.Equal(t, expectedVariation, v, "userID: %s, featureID: %s", userID, featureID)
	}

	// Get Variation by Default Strategy
	for i := 0; i < 3; i++ {
		wg.Add(1)
		userID := fmt.Sprintf("user-%d", i)
		go testVariationFunc(ctx, userID, featureIDString, featureIDStringVariation1)
	}

	// Get Variation by Targeting Users
	wg.Add(1)
	go testVariationFunc(ctx, targetUserID, featureIDString, featureIDStringTargetVariation)

	// Track
	wg.Add(1)
	go func(ctx context.Context, userID, goalID string) {
		defer wg.Done()
		user := newUser(t, userID)
		sdk.Track(ctx, user, goalID)
	}(ctx, userID, goalID)

	wg.Wait()
}

func TestBoolVariation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	sdk := newSDK(t, ctx)
	defer func() {
		// Close
		err := sdk.Close(ctx)
		assert.NoError(t, err)
	}()
	var wg sync.WaitGroup
	testVariationFunc := func(ctx context.Context, userID, featureID string, expectedVariation bool) {
		defer wg.Done()
		u := newUser(t, userID)
		v := sdk.BoolVariation(ctx, u, featureID, false)
		assert.Equal(t, expectedVariation, v, "userID: %s, featureID: %s", userID, featureID)
	}

	// Get Variation by Default Strategy
	for i := 0; i < 3; i++ {
		wg.Add(1)
		userID := fmt.Sprintf("user-%d", i)
		go testVariationFunc(ctx, userID, featureIDBoolean, true)
	}

	// Get Variation by Targeting Users
	wg.Add(1)
	go testVariationFunc(ctx, targetUserID, featureIDBoolean, featureIDBooleanTargetVariation)

	wg.Wait()
}

func TestIntVariation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	sdk := newSDK(t, ctx)
	defer func() {
		// Close
		err := sdk.Close(ctx)
		assert.NoError(t, err)
	}()
	var wg sync.WaitGroup
	testVariationFunc := func(ctx context.Context, userID, featureID string, expectedVariation int) {
		defer wg.Done()
		user := newUser(t, userID)
		v := sdk.IntVariation(ctx, user, featureID, -1)
		assert.Equal(t, expectedVariation, v, "userID: %s, featureID: %s", userID, featureID)
	}

	// Get Variation by Default Strategy
	for i := 0; i < 3; i++ {
		wg.Add(1)
		userID := fmt.Sprintf("user-%d", i)
		go testVariationFunc(ctx, userID, featureIDInt, featureIDIntVariation1)
	}

	// Get Variation by Targeting Users
	wg.Add(1)
	go testVariationFunc(ctx, targetUserID, featureIDInt, featureIDIntTargetVariation)

	wg.Wait()
}

func TestInt64Variation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	sdk := newSDK(t, ctx)
	defer func() {
		// Close
		err := sdk.Close(ctx)
		assert.NoError(t, err)
	}()
	var wg sync.WaitGroup
	testVariationFunc := func(ctx context.Context, userID, featureID string, expectedVariation int64) {
		defer wg.Done()
		user := newUser(t, userID)
		v := sdk.Int64Variation(ctx, user, featureID, -1000000000)
		assert.Equal(t, expectedVariation, v, "userID: %s, featureID: %s", userID, featureID)
	}

	// Get Variation by Default Strategy
	for i := 0; i < 3; i++ {
		wg.Add(1)
		userID := fmt.Sprintf("user-%d", i)
		go testVariationFunc(ctx, userID, featureIDInt64, featureIDInt64Variation1)
	}

	// Get Variation by Targeting Users
	wg.Add(1)
	go testVariationFunc(ctx, targetUserID, featureIDInt64, featureIDInt64TargetVariation)

	wg.Wait()
}

func TestFloat64Variation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	sdk := newSDK(t, ctx)
	defer func() {
		// Close
		err := sdk.Close(ctx)
		assert.NoError(t, err)
	}()
	var wg sync.WaitGroup
	testVariationFunc := func(ctx context.Context, userID, featureID string, expectedVariation float64) {
		defer wg.Done()
		u := newUser(t, userID)
		v := sdk.Float64Variation(ctx, u, featureID, -1.1)
		assert.Equal(t, expectedVariation, v, "userID: %s, featureID: %s", userID, featureID)
	}

	// Get Variation by Default Strategy
	for i := 0; i < 3; i++ {
		wg.Add(1)
		userID := fmt.Sprintf("user-%d", i)
		go testVariationFunc(ctx, userID, featureIDFloat, featureIDFloatVariation1)
	}

	// Get Variation by Targeting Users
	wg.Add(1)
	go testVariationFunc(ctx, targetUserID, featureIDFloat, featureIDFloatTargetVariation)

	wg.Wait()
}

func TestJSONVariation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	sdk := newSDK(t, ctx)
	defer func() {
		// Close
		err := sdk.Close(ctx)
		assert.NoError(t, err)
	}()
	var wg sync.WaitGroup
	type DstStruct struct {
		Str string `json:"str"`
		Int string `json:"int"`
	}
	testVariationFunc := func(ctx context.Context, userID, featureID string, expectedVariation interface{}) {
		defer wg.Done()
		u := newUser(t, userID)
		v := &DstStruct{}
		sdk.JSONVariation(ctx, u, featureID, v)
		assert.Equal(t, expectedVariation, v, "userID: %s, featureID: %s", userID, featureID)
	}

	// Get Variation by Default Strategy
	for i := 0; i < 3; i++ {
		wg.Add(1)
		userID := fmt.Sprintf("user-%d", i)
		go testVariationFunc(ctx, userID, featureIDJson, &DstStruct{Str: "str1", Int: "int1"})
	}

	// Get Variation by Targeting Users
	wg.Add(1)
	go testVariationFunc(ctx, targetUserID, featureIDJson, &DstStruct{Str: "str2", Int: "int2"})

	wg.Wait()
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
		bucketeer.WithEventFlushSize(5),
		bucketeer.WithEnableDebugLog(true),
	)
	assert.NoError(t, err)
	return sdk
}

func newUser(t *testing.T, id string) *user.User {
	t.Helper()
	return user.NewUser(id, map[string]string{"attr-key": "attr-value"})
}
