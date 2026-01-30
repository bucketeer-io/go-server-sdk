package bucketeer

import (
	"time"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/log"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/model"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/version"
)

// Option is the functional options type (Functional Options Pattern) to set sdk options.
type Option func(*options)

type options struct {
	enableLocalEvaluation bool
	cachePollingInterval  time.Duration
	tag                   string
	apiKey                string
	apiEndpoint           string
	scheme                string
	host                  string
	port                  int
	sdkVersion            string
	eventQueueCapacity    int
	numEventFlushWorkers  int
	sourceID              int32
	eventFlushInterval    time.Duration
	eventFlushSize        int
	enableDebugLog        bool
	enableEventStats      bool
	errorLogger           log.BaseLogger

	// Retry configuration (global)
	maxRetries           int           // Maximum retry attempts per HTTP request (Default: 3)
	retryInitialInterval time.Duration // Initial wait between retries (Default: 1s)
	retryMaxInterval     time.Duration // Maximum wait between retries (Default: 10s)
	retryMultiplier      float64       // Exponential backoff multiplier (Default: 2.0, internal, not exposed)
}

var defaultOptions = options{
	enableLocalEvaluation: false,
	cachePollingInterval:  1 * time.Minute,
	tag:                   "",
	apiKey:                "",
	apiEndpoint:           "",
	scheme:                "https",
	host:                  "",
	port:                  443,
	sdkVersion:            version.SDKVersion,
	eventQueueCapacity:    100_000,
	numEventFlushWorkers:  50,
	sourceID:              model.SourceIDGoServer.Int32(),
	eventFlushInterval:    10 * time.Second,
	eventFlushSize:        100,
	enableDebugLog:        false,
	enableEventStats:      true,
	errorLogger:           log.DefaultErrorLogger,

	// Retry configuration defaults
	maxRetries:           3,
	retryInitialInterval: 1 * time.Second,
	retryMaxInterval:     10 * time.Second,
	retryMultiplier:      2.0,
}

// WithEnableLocalEvaluation sets whether the user will be evaluated locally or not. (Default: false)
//
// Evaluate the end user locally in the SDK instead of on the server.
// Note: To evaluate the user locally, you must create an API key and select the server-side role.
func WithEnableLocalEvaluation(enable bool) Option {
	return func(opts *options) {
		opts.enableLocalEvaluation = enable
	}
}

// WithCachePollingInterval sets the polling interval for cache updating. (Default: 1 min)
//
// Note: To use the cache you must set the `WithEnableLocalEvaluation` to true.
func WithCachePollingInterval(interval time.Duration) Option {
	return func(opts *options) {
		opts.cachePollingInterval = interval
	}
}

// WithTag sets the specified tag when the flag is created. (Default: "")
//
// When empty, it will get all the flags in the environment.
// Tags are recommended for better response time and lower network traffic.
func WithTag(tag string) Option {
	return func(opts *options) {
		opts.tag = tag
	}
}

// WithAPIKey sets apiKey to use Bucketeer service. (Default: "")
func WithAPIKey(apiKey string) Option {
	return func(opts *options) {
		opts.apiKey = apiKey
	}
}

// WithAPIEndpoint sets apiEndpoint for the Bucketeer service. (Default: "")
func WithAPIEndpoint(apiEndpoint string) Option {
	return func(opts *options) {
		opts.apiEndpoint = apiEndpoint
	}
}

// WithScheme sets scheme to use Bucketeer service. (Default: "https")
func WithScheme(scheme string) Option {
	return func(opts *options) {
		opts.scheme = scheme
	}
}

// Deprecated: Use `WithAPIEndpoint` instead. This will be removed soon.
// WithHost sets host name for the Bucketeer service. (Default: "")
func WithHost(host string) Option {
	return func(opts *options) {
		opts.host = host
	}
}

// Deprecated: Use `WithScheme` instead. This will be removed soon.
// WithPort sets port for the Bucketeer service. (Default: 443)
func WithPort(port int) Option {
	return func(opts *options) {
		opts.port = port
	}
}

// WithWrapperSDKVersion sets the SDK version explicitly. (Default: version.SDKVersion)
// IMPORTANT: This option is intended for internal use only.
// It should NOT be set by developers directly integrating this SDK.
// Use this option ONLY when another SDK acts as a proxy and wraps this native SDK.
// In such cases, set this value to the version of the proxy SDK.
func WithWrapperSDKVersion(sdkVersion string) Option {
	return func(opts *options) {
		opts.sdkVersion = sdkVersion
	}
}

// WithWrapperSourceID sets the source ID explicitly. (Default: model.SourceIDGoServer)
// IMPORTANT: This option is intended for internal use only.
// It should NOT be set by developers directly integrating this SDK.
// Use this option ONLY when another SDK acts as a proxy and wraps this native SDK.
// In such cases, set this value to the sourceID of the proxy SDK.
// The sourceID is used to identify the origin of the request.
func WithWrapperSourceID(sourceID int32) Option {
	return func(opts *options) {
		opts.sourceID = sourceID
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

// WithNumEventFlushWorkers sets a number of workers to flush events. (Default: 50)
func WithNumEventFlushWorkers(numEventFlushWorkers int) Option {
	return func(opts *options) {
		opts.numEventFlushWorkers = numEventFlushWorkers
	}
}

// WithEventFlushInterval sets a interval of flushing events. (Default: 10s)
//
// Each worker sends the events to Bucketeer service every time eventFlushInterval elapses or
// its buffer exceeds eventFlushSize.
func WithEventFlushInterval(eventFlushInterval time.Duration) Option {
	return func(opts *options) {
		opts.eventFlushInterval = eventFlushInterval
	}
}

// WithEventFlushSize sets a size of the buffer for each worker. (Default: 100)
//
// Each worker sends the events to Bucketeer service every time eventFlushInterval elapses or
// its buffer exceeds eventFlushSize.
func WithEventFlushSize(eventFlushSize int) Option {
	return func(opts *options) {
		opts.eventFlushSize = eventFlushSize
	}
}

// WithEnableEventStats enables or disables event processing statistics tracking. (Default: true)
//
// When enabled, the SDK tracks event counts (created, sent, dropped, retried) which can be
// retrieved via sdk.EventStats(). The overhead is minimal (~1-2ns per event using atomic operations),
// but can be disabled for zero overhead if not needed.
func WithEnableEventStats(enable bool) Option {
	return func(opts *options) {
		opts.enableEventStats = enable
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

// WithMaxRetries sets the maximum number of retry attempts for HTTP requests. (Default: 3)
//
// When an HTTP request fails due to a retryable error (network errors, timeouts,
// or specific HTTP status codes like 429, 500, 502, 503, 504), the SDK will
// automatically retry up to this many times.
// Set to 0 to disable retries.
func WithMaxRetries(maxRetries int) Option {
	return func(opts *options) {
		opts.maxRetries = maxRetries
	}
}

// WithRetryInitialInterval sets the initial interval between retries. (Default: 1s)
//
// The SDK uses exponential backoff for retries:
// interval = initialInterval * (multiplier ^ attemptNumber)
// with ±25% jitter added to prevent thundering herd issues.
func WithRetryInitialInterval(interval time.Duration) Option {
	return func(opts *options) {
		opts.retryInitialInterval = interval
	}
}

// WithRetryMaxInterval sets the maximum interval between retries. (Default: 10s)
//
// The exponential backoff interval is capped at this value.
// This prevents excessively long waits between retry attempts.
func WithRetryMaxInterval(interval time.Duration) Option {
	return func(opts *options) {
		opts.retryMaxInterval = interval
	}
}
