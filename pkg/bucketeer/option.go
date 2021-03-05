package bucketeer

// Option is the functional options type (Functional Options Pattern) to set sdk options.
type Option func(*options)

type options struct {
	tag                string
	apiKey             string
	host               string
	port               int
	eventQueueCapacity int
	warnLogFunc        func(string)
	errorLogFunc       func(string)
}

var defaultOptions = options{
	tag:                "",
	apiKey:             "",
	host:               "",
	port:               443,
	eventQueueCapacity: 100_000,
	warnLogFunc:        nil,
	errorLogFunc:       nil,
}

// WithTag sets tag specified in getting evaluations. Default: ""
func WithTag(tag string) Option {
	return func(opts *options) {
		opts.tag = tag
	}
}

// WithAPIKey sets apiKey to use Bucketeer API Service. Default: ""
func WithAPIKey(apiKey string) Option {
	return func(opts *options) {
		opts.apiKey = apiKey
	}
}

// WithHost sets host name for the Bucketeer API service. Default: ""
func WithHost(host string) Option {
	return func(opts *options) {
		opts.host = host
	}
}

// WithPort sets port for the Bucketeer API service. Default: 443
func WithPort(port int) Option {
	return func(opts *options) {
		opts.port = port
	}
}

// WithEventQueueCapacity sets a capacity of the event queue. Default: 100_000
//
// The SDK buffers events up to the capacity in memory before processing.
// If the capacity is exceeded, events will be discarded.
func WithEventQueueCapacity(eventQueueCapacity int) Option {
	return func(opts *options) {
		opts.eventQueueCapacity = eventQueueCapacity
	}
}

// WithWarnLogFunc sets a function to output warn log. Default: nil (not output warn log)
func WithWarnLogFunc(warnLogFunc func(string)) Option {
	return func(opts *options) {
		opts.warnLogFunc = warnLogFunc
	}
}

// WithErrorLogFunc sets a function to output error log. Default: nil (not output error log)
func WithErrorLogFunc(errorLogFunc func(string)) Option {
	return func(opts *options) {
		opts.errorLogFunc = errorLogFunc
	}
}
