package bucketeer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithOptions(t *testing.T) {
	t.Parallel()
	tag := "go-server"
	apiKey := "apiKey"
	host := "host"
	port := 8443
	eventQueueCapacity := 100
	warnLogFuncCalled := false
	warnLogFunc := func(msg string) {
		warnLogFuncCalled = true
	}
	errorLogFuncCalled := false
	errorLogFunc := func(msg string) {
		errorLogFuncCalled = true
	}

	opts := []Option{
		WithTag(tag),
		WithAPIKey(apiKey),
		WithHost(host),
		WithPort(port),
		WithEventQueueCapacity(eventQueueCapacity),
		WithWarnLogFunc(warnLogFunc),
		WithErrorLogFunc(errorLogFunc),
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
	// Doesn't compare two func because the behavior is undefined
	// https://stackoverflow.com/questions/9643205/how-do-i-compare-two-functions-for-pointer-equality-in-the-latest-go-weekly
	dopts.warnLogFunc("call")
	assert.True(t, warnLogFuncCalled)
	dopts.errorLogFunc("call")
	assert.True(t, errorLogFuncCalled)
}
