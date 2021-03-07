package bucketeer

import (
	"testing"

	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/log"
	"github.com/stretchr/testify/assert"
)

func TestWithOptions(t *testing.T) {
	t.Parallel()
	tag := "go-server"
	apiKey := "apiKey"
	host := "host"
	port := 8443
	eventQueueCapacity := 100
	enableDebugLog := true
	errorLogger := log.DiscardErrorLogger

	opts := []Option{
		WithTag(tag),
		WithAPIKey(apiKey),
		WithHost(host),
		WithPort(port),
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
	assert.Equal(t, eventQueueCapacity, dopts.eventQueueCapacity)
	assert.Equal(t, enableDebugLog, dopts.enableDebugLog)
	assert.Equal(t, errorLogger, dopts.errorLogger)
}
