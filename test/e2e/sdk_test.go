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

func TestSDK(t *testing.T) {
	t.Parallel()
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
		user := newUser(t, userID)
		v := sdk.StringVariation(ctx, user, featureID, "default")
		assert.Equal(t, expectedVariation, v)
	}

	// Get Variation by Default Strategy (featureIDVariation1)
	for i := 0; i < 3; i++ {
		wg.Add(1)
		userID := fmt.Sprintf("user-%d", i)
		go testVariationFunc(ctx, userID, featureID, featureIDVariation1)
	}

	// Get Variation by Targeting Users (featureIDVariation2)
	wg.Add(1)
	go testVariationFunc(ctx, userID, featureID, featureIDVariation2)

	// Track
	wg.Add(1)
	go func(ctx context.Context, userID, goalID string) {
		defer wg.Done()
		user := newUser(t, userID)
		sdk.Track(ctx, user, goalID)
	}(ctx, userID, goalID)

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
