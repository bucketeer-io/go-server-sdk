package bucketeer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/log"
)

func TestWithOptions(t *testing.T) {
	t.Parallel()
	enableLocalEvaluation := true
	cachePollingInterval := 30 * time.Second
	tag := "go-server"
	apiKey := "apiKey"
	apiEndpoint := "apiEndpoint"
	scheme := "http"
	host := "host"
	port := 8443
	eventQueueCapacity := 100
	numEventFlushWorkers := 1
	eventFlushInterval := 1 * time.Second
	eventFlushSize := 10
	enableDebugLog := true
	errorLogger := log.DiscardErrorLogger

	opts := []Option{
		WithEnableLocalEvaluation(enableLocalEvaluation),
		WithCachePollingInterval(cachePollingInterval),
		WithTag(tag),
		WithAPIKey(apiKey),
		WithAPIEndpoint(apiEndpoint),
		WithScheme(scheme),
		WithHost(host),
		WithPort(port),
		WithEventQueueCapacity(eventQueueCapacity),
		WithNumEventFlushWorkers(numEventFlushWorkers),
		WithEventFlushInterval(eventFlushInterval),
		WithEventFlushSize(eventFlushSize),
		WithEnableDebugLog(enableDebugLog),
		WithErrorLogger(errorLogger),
	}
	dopts := defaultOptions
	for _, opt := range opts {
		opt(&dopts)
	}

	assert.Equal(t, enableLocalEvaluation, dopts.enableLocalEvaluation)
	assert.Equal(t, cachePollingInterval, dopts.cachePollingInterval)
	assert.Equal(t, tag, dopts.tag)
	assert.Equal(t, apiKey, dopts.apiKey)
	assert.Equal(t, apiEndpoint, dopts.apiEndpoint)
	assert.Equal(t, scheme, dopts.scheme)
	assert.Equal(t, host, dopts.host)
	assert.Equal(t, port, dopts.port)
	assert.Equal(t, eventQueueCapacity, dopts.eventQueueCapacity)
	assert.Equal(t, numEventFlushWorkers, dopts.numEventFlushWorkers)
	assert.Equal(t, eventFlushInterval, dopts.eventFlushInterval)
	assert.Equal(t, eventFlushSize, dopts.eventFlushSize)
	assert.Equal(t, enableDebugLog, dopts.enableDebugLog)
	assert.Equal(t, errorLogger, dopts.errorLogger)
}

func TestDefaultRetryOptions(t *testing.T) {
	t.Parallel()

	// Verify default retry configuration values
	assert.Equal(t, 3, defaultOptions.maxRetries, "default maxRetries should be 3")
	assert.Equal(t, 1*time.Second, defaultOptions.retryInitialInterval, "default retryInitialInterval should be 1s")
	assert.Equal(t, 10*time.Second, defaultOptions.retryMaxInterval, "default retryMaxInterval should be 10s")
	assert.Equal(t, 2.0, defaultOptions.retryMultiplier, "default retryMultiplier should be 2.0")
}

func TestWithMaxRetries(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    int
		expected int
	}{
		{
			name:     "set to 5 retries",
			input:    5,
			expected: 5,
		},
		{
			name:     "set to 0 disables retries",
			input:    0,
			expected: 0,
		},
		{
			name:     "set to 1 retry",
			input:    1,
			expected: 1,
		},
		{
			name:     "set to large value",
			input:    100,
			expected: 100,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dopts := defaultOptions
			WithMaxRetries(tc.input)(&dopts)
			assert.Equal(t, tc.expected, dopts.maxRetries)
		})
	}
}

func TestWithRetryInitialInterval(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    time.Duration
		expected time.Duration
	}{
		{
			name:     "set to 500ms",
			input:    500 * time.Millisecond,
			expected: 500 * time.Millisecond,
		},
		{
			name:     "set to 2s",
			input:    2 * time.Second,
			expected: 2 * time.Second,
		},
		{
			name:     "set to 100ms",
			input:    100 * time.Millisecond,
			expected: 100 * time.Millisecond,
		},
		{
			name:     "set to 0 disables initial wait",
			input:    0,
			expected: 0,
		},
		{
			name:     "set to 5s",
			input:    5 * time.Second,
			expected: 5 * time.Second,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dopts := defaultOptions
			WithRetryInitialInterval(tc.input)(&dopts)
			assert.Equal(t, tc.expected, dopts.retryInitialInterval)
		})
	}
}

func TestWithRetryMaxInterval(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    time.Duration
		expected time.Duration
	}{
		{
			name:     "set to 5s",
			input:    5 * time.Second,
			expected: 5 * time.Second,
		},
		{
			name:     "set to 30s",
			input:    30 * time.Second,
			expected: 30 * time.Second,
		},
		{
			name:     "set to 1m",
			input:    1 * time.Minute,
			expected: 1 * time.Minute,
		},
		{
			name:     "set to 0",
			input:    0,
			expected: 0,
		},
		{
			name:     "set to 15s",
			input:    15 * time.Second,
			expected: 15 * time.Second,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dopts := defaultOptions
			WithRetryMaxInterval(tc.input)(&dopts)
			assert.Equal(t, tc.expected, dopts.retryMaxInterval)
		})
	}
}

func TestWithRetryOptionsAllTogether(t *testing.T) {
	t.Parallel()

	maxRetries := 5
	retryInitialInterval := 2 * time.Second
	retryMaxInterval := 20 * time.Second

	opts := []Option{
		WithMaxRetries(maxRetries),
		WithRetryInitialInterval(retryInitialInterval),
		WithRetryMaxInterval(retryMaxInterval),
	}

	dopts := defaultOptions
	for _, opt := range opts {
		opt(&dopts)
	}

	assert.Equal(t, maxRetries, dopts.maxRetries)
	assert.Equal(t, retryInitialInterval, dopts.retryInitialInterval)
	assert.Equal(t, retryMaxInterval, dopts.retryMaxInterval)
	// retryMultiplier is internal and should remain at default
	assert.Equal(t, 2.0, dopts.retryMultiplier)
}

func TestWithRetryOptionsWithOtherOptions(t *testing.T) {
	t.Parallel()

	// Test that retry options work alongside other SDK options
	enableLocalEvaluation := true
	cachePollingInterval := 30 * time.Second
	tag := "go-server"
	apiKey := "apiKey"
	maxRetries := 5
	retryInitialInterval := 500 * time.Millisecond
	retryMaxInterval := 15 * time.Second

	opts := []Option{
		WithEnableLocalEvaluation(enableLocalEvaluation),
		WithCachePollingInterval(cachePollingInterval),
		WithTag(tag),
		WithAPIKey(apiKey),
		WithMaxRetries(maxRetries),
		WithRetryInitialInterval(retryInitialInterval),
		WithRetryMaxInterval(retryMaxInterval),
	}

	dopts := defaultOptions
	for _, opt := range opts {
		opt(&dopts)
	}

	// Verify retry options
	assert.Equal(t, maxRetries, dopts.maxRetries)
	assert.Equal(t, retryInitialInterval, dopts.retryInitialInterval)
	assert.Equal(t, retryMaxInterval, dopts.retryMaxInterval)

	// Verify other options are also set correctly
	assert.Equal(t, enableLocalEvaluation, dopts.enableLocalEvaluation)
	assert.Equal(t, cachePollingInterval, dopts.cachePollingInterval)
	assert.Equal(t, tag, dopts.tag)
	assert.Equal(t, apiKey, dopts.apiKey)
}

func TestWithRetryOptionsOverwrite(t *testing.T) {
	t.Parallel()

	// Test that options can be overwritten when applied multiple times
	dopts := defaultOptions

	// First application
	WithMaxRetries(10)(&dopts)
	assert.Equal(t, 10, dopts.maxRetries)

	// Second application overwrites
	WithMaxRetries(5)(&dopts)
	assert.Equal(t, 5, dopts.maxRetries)

	// Same for intervals
	WithRetryInitialInterval(2 * time.Second)(&dopts)
	assert.Equal(t, 2*time.Second, dopts.retryInitialInterval)

	WithRetryInitialInterval(500 * time.Millisecond)(&dopts)
	assert.Equal(t, 500*time.Millisecond, dopts.retryInitialInterval)

	WithRetryMaxInterval(30 * time.Second)(&dopts)
	assert.Equal(t, 30*time.Second, dopts.retryMaxInterval)

	WithRetryMaxInterval(5 * time.Second)(&dopts)
	assert.Equal(t, 5*time.Second, dopts.retryMaxInterval)
}
