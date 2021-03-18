package bucketeer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/log"
)

func TestWithOptions(t *testing.T) {
	t.Parallel()
	tag := "go-server"
	apiKey := "apiKey"
	host := "host"
	port := 8443
	numEventFlushWorkers := 1
	eventFlushInterval := 1 * time.Second
	eventFlushSize := 10
	eventQueueCapacity := 100
	enableDebugLog := true
	errorLogger := log.DiscardErrorLogger

	opts := []Option{
		WithTag(tag),
		WithAPIKey(apiKey),
		WithHost(host),
		WithPort(port),
		WithNumEventFlushWorkers(numEventFlushWorkers),
		WithEventFlushInterval(eventFlushInterval),
		WithEventFlushSize(eventFlushSize),
		WithEventQueueCapacity(eventQueueCapacity),
		WithEnableDebugLog(enableDebugLog),
		WithErrorLogger(errorLogger),
	}
	dopts := defaultOptions
	for _, opt := range opts {
		opt(&dopts)
	}

	assert.Equal(t, tag, dopts.tag)
	assert.Equal(t, apiKey, dopts.apiKey)
	assert.Equal(t, host, dopts.host)
	assert.Equal(t, port, dopts.port)
	assert.Equal(t, numEventFlushWorkers, dopts.numEventFlushWorkers)
	assert.Equal(t, eventFlushInterval, dopts.eventFlushInterval)
	assert.Equal(t, eventFlushSize, dopts.eventFlushSize)
	assert.Equal(t, eventQueueCapacity, dopts.eventQueueCapacity)
	assert.Equal(t, enableDebugLog, dopts.enableDebugLog)
	assert.Equal(t, errorLogger, dopts.errorLogger)
}
