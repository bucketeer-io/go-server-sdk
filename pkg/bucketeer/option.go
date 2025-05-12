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
	errorLogger           log.BaseLogger
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
	eventFlushInterval:    1 * time.Minute,
	eventFlushSize:        100,
	enableDebugLog:        false,
	errorLogger:           log.DefaultErrorLogger,
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

// WithSDKVersion sets the SDK version. (Default: version.SDKVersion)
func WithSDKVersion(sdkVersion string) Option {
	return func(opts *options) {
		opts.sdkVersion = sdkVersion
	}
}

// WithSourceID sets the source ID. (Default: model.SourceIDGoServer)
func WithSourceID(sourceID int32) Option {
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

// WithEventFlushInterval sets a interval of flushing events. (Default: 1 min)
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
