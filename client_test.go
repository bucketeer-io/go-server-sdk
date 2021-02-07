package bucketeer

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	const (
		envAPIKey          = "BUCKETEER_API_KEY"
		featureIDBoolTrue  = "TestBoolFeatTrue"
		featureIDInt10     = "TestIntFeat10"
		featureIDFloat123  = "TestFloatFeat1-23"
		featureIDStringFoo = "TestStringFeatFoo"
		featureIDJSONHi    = "TestJsonFeatHi"
		goalID             = "TestGoal"
	)
	apiKey := os.Getenv(envAPIKey)
	if apiKey == "" {
		t.Skip("API key is required")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = NewDefault(ctx)

	// test: NewClient
	const interval = time.Minute
	config := &Config{
		APIKey:             APIKey(apiKey),
		Host:               "api-media.bucketeer.jp:443",
		Tag:                "test",
		EventFlushInterval: interval,
		Log:                func(error) {},
	}
	cli, err := NewClient(ctx, config)
	require.NoError(t, err)
	require.NotNil(t, cli)
	defer cli.Close()
	clientS := cli.(*client)
	require.Equal(t, defaultBufferSize, clientS.eventBufferSize)
	require.Equal(t, interval, clientS.eventFlushInterval)

	// test: bool success
	boolValue, track := cli.BoolVariation(ctx, featureIDBoolTrue, false)
	require.True(t, boolValue)
	require.NotNil(t, track)
	track.Track(goalID, 1)

	// test: bool default value
	boolValue, track = cli.BoolVariation(ctx, "foo", false)
	require.False(t, boolValue)
	require.NotNil(t, track)

	// test: int success
	int64Value, track := cli.Int64Variation(ctx, featureIDInt10, 1)
	require.Equal(t, int64(10), int64Value)
	require.NotNil(t, track)

	// test: int default value
	int64Value, track = cli.Int64Variation(ctx, "foo", 1)
	require.Equal(t, int64(1), int64Value)
	require.NotNil(t, track)

	// test: float success
	floatValue, track := cli.Float64Variation(ctx, featureIDFloat123, 3.45)
	require.Equal(t, 1.23, floatValue)
	require.NotNil(t, track)

	// test: float default value
	floatValue, track = cli.Float64Variation(ctx, "foo", 3.45)
	require.Equal(t, 3.45, floatValue)
	require.NotNil(t, track)

	// test: string success
	stringValue, track := cli.StringVariation(ctx, featureIDStringFoo, "bar")
	require.Equal(t, "foo", stringValue)
	require.NotNil(t, track)

	// test: string default value
	stringValue, track = cli.StringVariation(ctx, "foo", "bar")
	require.Equal(t, "bar", stringValue)
	require.NotNil(t, track)

	// test: json success
	result := &struct {
		Value string
	}{
		Value: "default",
	}
	track = cli.JSONVariation(ctx, featureIDJSONHi, result)
	require.Equal(t, "hi", result.Value)
	require.NotNil(t, track)

	// test: json default value
	result.Value = "default"
	track = cli.JSONVariation(ctx, "foo", result)
	require.Equal(t, "default", result.Value)
	require.NotNil(t, track)

	// test: event channel
	events := clientS.events
	require.Len(t, events, 10)

	// context cancel
	cancel()
}
