package bucketeer

import "github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/log"

// Option is the functional options type (Functional Options Pattern) to set sdk options.
type Option func(*options)

type options struct {
	tag                string
	apiKey             string
	host               string
	port               int
	eventQueueCapacity int
	enableDebugLog     bool
	errorLogger        log.BaseLogger
}

var defaultOptions = options{
	tag:                "",
	apiKey:             "",
	host:               "",
	port:               443,
	eventQueueCapacity: 100_000,
	enableDebugLog:     false,
	errorLogger:        log.DefaultErrorLogger,
}

// WithTag sets tag specified in getting evaluations. (Default: "")
func WithTag(tag string) Option {
	return func(opts *options) {
		opts.tag = tag
	}
}

// WithAPIKey sets apiKey to use Bucketeer API Service. (Default: "")
func WithAPIKey(apiKey string) Option {
	return func(opts *options) {
		opts.apiKey = apiKey
	}
}

// WithHost sets host name for the Bucketeer API service. (Default: "")
func WithHost(host string) Option {
	return func(opts *options) {
		opts.host = host
	}
}

// WithPort sets port for the Bucketeer API service. (Default: 443)
func WithPort(port int) Option {
	return func(opts *options) {
		opts.port = port
	}
}

// WithEventQueueCapacity sets a capacity of the event queue. (Default: 100_000)
//
// The SDK buffers events up to the capacity in memory before processing.
// If the capacity is exceeded, events will be discarded.
func WithEventQueueCapacity(eventQueueCapacity int) Option {
	return func(opts *options) {
		opts.eventQueueCapacity = eventQueueCapacity
	}
}

// WithEnableDebugLog sets if outpus debug logs or not. (Default: false)
//
// Debug logs are for Bucketeer SDK developers.
func WithEnableDebugLog(enableDebugLog bool) Option {
	return func(opts *options) {
		opts.enableDebugLog = enableDebugLog
	}
}

// WithErrorLogger sets a looger to output error logs. (Default: log.DefaultErrorLogger)
//
// Error logs are for Bucketeer SDK users.
func WithErrorLogger(errorLogger log.BaseLogger) Option {
	return func(opts *options) {
		opts.errorLogger = errorLogger
	}
}
