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
