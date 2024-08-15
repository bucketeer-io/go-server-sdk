package bucketeer

//go:generate mockgen -source=$GOFILE -package=$GOPACKAGE -destination=../../test/mock/$GOPACKAGE/$GOFILE
import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	// nolint:staticcheck
	iotag "go.opencensus.io/tag"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/api"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/cache"
	cacheprocessor "github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/cache/processor"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/evaluator"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/event"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/log"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/model"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/user"
)

// SDK is the Bucketeer SDK.
//
// SDK represents the ability to get the value of a feature flag and to track goal events
// by communicating with the Bucketeer service.
//
// A user application should instantiate a single SDK instance for the lifetime of the application and share it.
// SDK is safe for concurrent use by multiple goroutines.
type SDK interface {
	// BoolVariation returns the value of a feature flag (whose variations are booleans) for the given user.
	//
	// BoolVariation returns defaultValue if an error occurs.
	BoolVariation(ctx context.Context, user *user.User, featureID string, defaultValue bool) bool

	BoolVariationDetail(
		ctx context.Context,
		user *user.User,
		featureID string,
		defaultValue bool) model.BKTEvaluationDetail[bool]

	// IntVariation returns the value of a feature flag (whose variations are ints) for the given user.
	//
	// IntVariation returns defaultValue if an error occurs.
	IntVariation(ctx context.Context, user *user.User, featureID string, defaultValue int) int

	IntVariationDetail(
		ctx context.Context,
		user *user.User,
		featureID string,
		defaultValue int) model.BKTEvaluationDetail[int]

	// Int64Variation returns the value of a feature flag (whose variations are int64s) for the given user.
	//
	// Int64Variation returns defaultValue if an error occurs.
	Int64Variation(ctx context.Context, user *user.User, featureID string, defaultValue int64) int64

	Int64VariationDetail(
		ctx context.Context,
		user *user.User,
		featureID string,
		defaultValue int64) model.BKTEvaluationDetail[int64]

	// Float64Variation returns the value of a feature flag (whose variations are float64s) for the given user.
	//
	// Float64Variation returns defaultValue if an error occurs.
	Float64Variation(ctx context.Context, user *user.User, featureID string, defaultValue float64) float64

	Float64VariationDetail(
		ctx context.Context,
		user *user.User,
		featureID string,
		defaultValue float64) model.BKTEvaluationDetail[float64]

	// StringVariation returns the value of a feature flag (whose variations are strings) for the given user.
	//
	// StringVariation returns defaultValue if an error occurs.
	StringVariation(ctx context.Context, user *user.User, featureID, defaultValue string) string

	StringVariationDetail(
		ctx context.Context,
		user *user.User,
		featureID,
		defaultValue string) model.BKTEvaluationDetail[string]

	// Deprecated JSONVariation is deprecated. Please use ObjectVariation instead.
	JSONVariation(ctx context.Context, user *user.User, featureID string, dst interface{})

	ObjectVariation(
		ctx context.Context,
		user *user.User,
		featureID string,
		defaultValue interface{}) interface{}

	ObjectVariationDetail(
		ctx context.Context,
		user *user.User,
		featureID string,
		defaultValue interface{}) model.BKTEvaluationDetail[interface{}]

	// Track reports that a user has performed a goal event.
	//
	// TODO: Track doesn't work correctly until Bucketeer service implements the new goal tracking architecture.
	Track(ctx context.Context, user *user.User, GoalID string)

	// TrackValue reports that a user has performed a goal event, and associates it with a custom value.
	//
	// TODO: TrackValue doesn't work correctly until Bucketeer service implements the new goal tracking architecture.
	TrackValue(ctx context.Context, user *user.User, GoalID string, value float64)

	// Close tears down all SDK activities and resources, after ensuring that all events have been delivered.
	//
	// After calling this, the SDK should no longer be used.
	Close(ctx context.Context) error
}

var (
	errResponseNil                 = errors.New("invalid get evaluation response: res is nil")
	errResponseEvaluationNil       = errors.New("invalid get evaluation response: evaluation is nil")
	errResponseVariationValueEmpty = errors.New("invalid get evaluation response: variation value is empty")
	errResponseDifferentFeatureIDs = "invalid get evaluation response: feature id doesn't match: actual %s != expected %s"
)

type sdk struct {
	enableLocalEvaluation     bool
	tag                       string
	apiClient                 api.Client
	eventProcessor            event.Processor
	featureFlagCacheProcessor cacheprocessor.FeatureFlagProcessor
	segmentUserCacheProcessor cacheprocessor.SegmentUserProcessor
	featureFlagsCache         cache.FeaturesCache
	segmentUsersCache         cache.SegmentUsersCache
	loggers                   *log.Loggers
}

// NewSDK creates a new Bucketeer SDK.
func NewSDK(ctx context.Context, opts ...Option) (SDK, error) {
	dopts := defaultOptions
	for _, opt := range opts {
		opt(&dopts)
	}
	loggerConf := &log.LoggersConfig{
		EnableDebugLog: dopts.enableDebugLog,
		ErrorLogger:    dopts.errorLogger,
	}
	loggers := log.NewLoggers(loggerConf)
	client, err := api.NewClient(&api.ClientConfig{APIKey: dopts.apiKey, Host: dopts.host})
	if err != nil {
		return nil, fmt.Errorf("bucketeer: failed to new api client: %w", err)
	}
	processorConf := &event.ProcessorConfig{
		QueueCapacity:   dopts.eventQueueCapacity,
		NumFlushWorkers: dopts.numEventFlushWorkers,
		FlushInterval:   dopts.eventFlushInterval,
		FlushSize:       dopts.eventFlushSize,
		APIClient:       client,
		Loggers:         loggers,
		Tag:             dopts.tag,
	}
	processor := event.NewProcessor(ctx, processorConf)
	if !dopts.enableLocalEvaluation {
		// Evaluate the end user on the server
		return &sdk{
			enableLocalEvaluation: dopts.enableLocalEvaluation,
			tag:                   dopts.tag,
			apiClient:             client,
			eventProcessor:        processor,
			loggers:               loggers,
		}, nil
	}
	// Evaluate the user locally
	// Set the cache processors to update the flags and segment users cache in the background
	cacheInMemory := cache.NewInMemoryCache()
	fcp := newFeatureFlagCacheProcessor(
		cacheInMemory,
		dopts.cachePollingInterval,
		client,
		processor,
		dopts.tag,
		loggers,
	)
	sucp := newSegmentUserCacheProcessor(
		cacheInMemory,
		dopts.cachePollingInterval,
		client,
		processor,
		loggers,
	)
	// Run the cache processors to update the cache in background
	fcp.Run()
	sucp.Run()

	return &sdk{
		enableLocalEvaluation:     dopts.enableLocalEvaluation,
		tag:                       dopts.tag,
		apiClient:                 client,
		eventProcessor:            processor,
		featureFlagCacheProcessor: fcp,
		segmentUserCacheProcessor: sucp,
		featureFlagsCache:         cache.NewFeaturesCache(cacheInMemory),
		segmentUsersCache:         cache.NewSegmentUsersCache(cacheInMemory),
		loggers:                   loggers,
	}, nil
}

func newFeatureFlagCacheProcessor(
	cache cache.InMemoryCache,
	pollingInterval time.Duration,
	apiClient api.Client,
	eventProcessor event.Processor,
	tag string,
	loggers *log.Loggers) cacheprocessor.FeatureFlagProcessor {
	conf := &cacheprocessor.FeatureFlagProcessorConfig{
		Cache:                   cache,
		PollingInterval:         pollingInterval,
		APIClient:               apiClient,
		PushLatencyMetricsEvent: eventProcessor.PushLatencyMetricsEvent,
		PushSizeMetricsEvent:    eventProcessor.PushSizeMetricsEvent,
		PushErrorEvent:          eventProcessor.PushErrorEvent,
		Tag:                     tag,
		Loggers:                 loggers,
	}
	return cacheprocessor.NewFeatureFlagProcessor(conf)
}

func newSegmentUserCacheProcessor(
	cache cache.InMemoryCache,
	pollingInterval time.Duration,
	apiClient api.Client,
	eventProcessor event.Processor,
	loggers *log.Loggers) cacheprocessor.SegmentUserProcessor {
	conf := &cacheprocessor.SegmentUserProcessorConfig{
		Cache:                   cache,
		PollingInterval:         pollingInterval,
		APIClient:               apiClient,
		PushLatencyMetricsEvent: eventProcessor.PushLatencyMetricsEvent,
		PushSizeMetricsEvent:    eventProcessor.PushSizeMetricsEvent,
		PushErrorEvent:          eventProcessor.PushErrorEvent,
		Loggers:                 loggers,
	}
	return cacheprocessor.NewSegmentUserProcessor(conf)
}

func (s *sdk) BoolVariation(ctx context.Context, user *user.User, featureID string, defaultValue bool) bool {
	return getEvaluationDetail[bool](s, ctx, user, featureID, defaultValue, "BoolVariation").VariationValue
}

func (s *sdk) BoolVariationDetail(
	ctx context.Context,
	user *user.User,
	featureID string,
	defaultValue bool) model.BKTEvaluationDetail[bool] {
	return getEvaluationDetail[bool](s, ctx, user, featureID, defaultValue, "BoolVariationDetail")
}

func (s *sdk) IntVariation(ctx context.Context, user *user.User, featureID string, defaultValue int) int {
	return getEvaluationDetail[int](s, ctx, user, featureID, defaultValue, "IntVariation").VariationValue
}

func (s *sdk) IntVariationDetail(
	ctx context.Context,
	user *user.User,
	featureID string,
	defaultValue int) model.BKTEvaluationDetail[int] {
	return getEvaluationDetail[int](s, ctx, user, featureID, defaultValue, "IntVariationDetail")
}

func (s *sdk) Int64Variation(ctx context.Context, user *user.User, featureID string, defaultValue int64) int64 {
	return getEvaluationDetail[int64](s, ctx, user, featureID, defaultValue, "Int64Variation").VariationValue
}

func (s *sdk) Int64VariationDetail(
	ctx context.Context,
	user *user.User,
	featureID string,
	defaultValue int64) model.BKTEvaluationDetail[int64] {
	return getEvaluationDetail[int64](s, ctx, user, featureID, defaultValue, "Int64VariationDetail")
}

func (s *sdk) Float64Variation(ctx context.Context, user *user.User, featureID string, defaultValue float64) float64 {
	return getEvaluationDetail[float64](s, ctx, user, featureID, defaultValue, "Float64Variation").VariationValue
}

func (s *sdk) Float64VariationDetail(
	ctx context.Context,
	user *user.User,
	featureID string,
	defaultValue float64) model.BKTEvaluationDetail[float64] {
	return getEvaluationDetail[float64](s, ctx, user, featureID, defaultValue, "Float64VariationDetail")
}

func (s *sdk) StringVariation(ctx context.Context, user *user.User, featureID, defaultValue string) string {
	return getEvaluationDetail[string](s, ctx, user, featureID, defaultValue, "StringVariation").VariationValue
}

func (s *sdk) StringVariationDetail(
	ctx context.Context,
	user *user.User,
	featureID,
	defaultValue string) model.BKTEvaluationDetail[string] {
	return getEvaluationDetail[string](s, ctx, user, featureID, defaultValue, "StringVariationDetail")
}

func (s *sdk) JSONVariation(ctx context.Context, user *user.User, featureID string, dst interface{}) {
	evaluation, err := s.getEvaluation(ctx, user, featureID)
	if err != nil {
		s.logVariationError(err, "JSONVariation", user.ID, featureID)
		s.eventProcessor.PushDefaultEvaluationEvent(user, featureID)
		return
	}
	variation := evaluation.VariationValue
	err = json.Unmarshal([]byte(variation), dst)
	if err != nil {
		s.logVariationError(err, "JSONVariation", user.ID, featureID)
		s.eventProcessor.PushDefaultEvaluationEvent(user, featureID)
		return
	}
	s.eventProcessor.PushEvaluationEvent(user, evaluation)
}

func (s *sdk) ObjectVariation(
	ctx context.Context,
	user *user.User,
	featureID string,
	defaultValue interface{}) interface{} {
	return getEvaluationDetail[interface{}](s, ctx, user, featureID, defaultValue, "ObjectVariation").VariationValue
}

func (s *sdk) ObjectVariationDetail(
	ctx context.Context,
	user *user.User,
	featureID string,
	defaultValue interface{}) model.BKTEvaluationDetail[interface{}] {
	return getEvaluationDetail[interface{}](s, ctx, user, featureID, defaultValue, "ObjectVariationDetail")
}

func getEvaluationDetail[T model.EvaluationValue](
	s *sdk,
	ctx context.Context,
	user *user.User,
	featureID string,
	defaultValue T,
	logMethodName string,
) model.BKTEvaluationDetail[T] {
	var err error
	var value T
	evaluation, err := s.getEvaluation(ctx, user, featureID)
	if err != nil {
		s.logVariationError(err, logMethodName, user.ID, featureID)
		s.eventProcessor.PushDefaultEvaluationEvent(user, featureID)
		return model.NewEvaluationDetail(
			featureID,
			user.ID,
			"",
			"",
			0,
			model.ReasonClient,
			defaultValue,
		)
	}
	variation := evaluation.VariationValue
	switch any(defaultValue).(type) {
	case int:
		var parsedValue int64
		parsedValue, err = strconv.ParseInt(variation, 10, 64)
		if err == nil {
			value = any(int(parsedValue)).(T)
		}
	case int64:
		var parsedValue int64
		parsedValue, err = strconv.ParseInt(variation, 10, 64)
		if err == nil {
			value = any(parsedValue).(T)
		}
	case float64:
		var parsedValue float64
		parsedValue, err = strconv.ParseFloat(variation, 64)
		if err == nil {
			value = any(parsedValue).(T)
		}
	case string:
		value = any(variation).(T)
	case bool:
		var parsedValue bool
		parsedValue, err = strconv.ParseBool(variation)
		if err == nil {
			value = any(parsedValue).(T)
		}
	case interface{}:
		parsedValue := defaultValue
		err = json.Unmarshal([]byte(variation), &parsedValue)
		if err == nil {
			value = parsedValue
		}
	default:
		err = fmt.Errorf("unsupported type: %T", defaultValue)
	}
	if err != nil {
		s.logVariationError(err, logMethodName, user.ID, featureID)
		s.eventProcessor.PushDefaultEvaluationEvent(user, featureID)

		return model.NewEvaluationDetail[T](
			featureID,
			user.ID,
			"",
			"",
			0,
			model.ReasonClient,
			defaultValue,
		)
	}
	s.eventProcessor.PushEvaluationEvent(user, evaluation)

	return model.NewEvaluationDetail[T](
		featureID,
		user.ID,
		evaluation.VariationID,
		evaluation.VariationName,
		evaluation.FeatureVersion,
		evaluation.Reason.Type,
		value,
	)
}

func (s *sdk) getEvaluation(ctx context.Context, user *user.User, featureID string) (*model.Evaluation, error) {
	if !user.Valid() {
		return nil, fmt.Errorf("invalid user: %v", user)
	}
	if s.enableLocalEvaluation {
		// Evaluate the end user locally
		return s.getEvaluationLocally(ctx, user, featureID)
	}
	// Evaluate the end user on the server
	return s.getEvaluationRemotely(ctx, user, featureID)
}

func (s *sdk) getEvaluationLocally(
	ctx context.Context,
	user *user.User,
	featureID string,
) (*model.Evaluation, error) {
	reqStart := time.Now()
	eval := evaluator.NewEvaluator(s.tag, s.featureFlagsCache, s.segmentUsersCache)
	evaluation, err := eval.Evaluate(user, featureID)
	if err != nil {
		s.loggers.Errorf("bucketeer: failed to evaluate user locally: %w", err)
		// This error must be reported as a SDK internal error
		e := fmt.Errorf("internal error while evaluating user locally: %w", err)
		s.eventProcessor.PushErrorEvent(e, model.SDKGetEvaluation)
		return nil, err
	}
	s.eventProcessor.PushLatencyMetricsEvent(time.Since(reqStart), model.SDKGetEvaluation)
	go s.collectMetrics(ctx, featureID, reqStart)
	return evaluation, nil
}

func (s *sdk) getEvaluationRemotely(
	ctx context.Context,
	user *user.User,
	featureID string,
) (*model.Evaluation, error) {
	reqStart := time.Now()
	res, size, err := s.apiClient.GetEvaluation(model.NewGetEvaluationRequest(
		s.tag, featureID,
		user,
	))
	if err != nil {
		s.eventProcessor.PushErrorEvent(err, model.GetEvaluation)
		return nil, fmt.Errorf("failed to get evaluation: %w", err)
	}
	s.eventProcessor.PushLatencyMetricsEvent(time.Since(reqStart), model.GetEvaluation)
	s.eventProcessor.PushSizeMetricsEvent(size, model.GetEvaluation)
	// Validate the response from the server
	if err := s.validateGetEvaluationResponse(res, featureID); err != nil {
		s.eventProcessor.PushErrorEvent(err, model.GetEvaluation)
		return nil, err
	}
	go s.collectMetrics(ctx, featureID, reqStart)
	return res.Evaluation, nil
}

func (s *sdk) validateGetEvaluationResponse(res *model.GetEvaluationResponse, featureID string) error {
	if res == nil {
		return errResponseNil
	}
	if res.Evaluation == nil {
		return errResponseEvaluationNil
	}
	if res.Evaluation.FeatureID != featureID {
		return fmt.Errorf(
			errResponseDifferentFeatureIDs,
			res.Evaluation.FeatureID,
			featureID,
		)
	}
	if res.Evaluation.VariationValue == "" {
		return errResponseVariationValueEmpty
	}
	return nil
}

// Collect metrics to OpenCensus
func (s *sdk) collectMetrics(
	ctx context.Context,
	featureID string,
	startTime time.Time) {
	code := http.StatusOK
	mutators := []iotag.Mutator{
		iotag.Insert(keyFeatureID, featureID),
		iotag.Insert(keyStatus, strconv.Itoa(code)),
	}
	ctx, err := newMetricsContext(ctx, mutators)
	if err != nil {
		s.loggers.Errorf("bucketeer: failed to create metrics context (featureID: %s, statusCode: %d)",
			featureID,
			code,
		)
		s.eventProcessor.PushErrorEvent(err, model.GetEvaluation)
		return
	}
	count(ctx)
	measure(ctx, time.Since(startTime))
}

func (s *sdk) logVariationError(err error, methodName, UserID, featureID string) {
	s.loggers.Errorf(
		"bucketeer: %s returns default value (err: %v, UserID: %s, featureID: %s)",
		methodName,
		err,
		UserID,
		featureID,
	)
}

func (s *sdk) Track(ctx context.Context, user *user.User, GoalID string) {
	s.TrackValue(ctx, user, GoalID, 0.0)
}

func (s *sdk) TrackValue(ctx context.Context, user *user.User, GoalID string, value float64) {
	if !user.Valid() {
		s.loggers.Errorf("bucketeer: failed to track due to invalid user (user: %v, GoalID: %v, value: %g)",
			user,
			GoalID,
			value,
		)
		return
	}
	s.eventProcessor.PushGoalEvent(user, GoalID, value)
}

func (s *sdk) Close(ctx context.Context) error {
	if err := s.eventProcessor.Close(ctx); err != nil {
		return fmt.Errorf("bucketeer: failed to close event processor: %v", err)
	}
	if s.enableLocalEvaluation {
		s.featureFlagCacheProcessor.Close()
		s.segmentUserCacheProcessor.Close()
	}
	return nil
}

type nopSDK struct {
}

// NewNopSDK creates a new no-op Bucketeer SDK.
//
// It never requests to Bucketeer Service, and just returns default value when XXXVariation methods called.
func NewNopSDK() SDK {
	return &nopSDK{}
}

func (s *nopSDK) BoolVariation(
	ctx context.Context,
	user *user.User,
	featureID string,
	defaultValue bool) bool {
	return defaultValue
}

func (s *nopSDK) BoolVariationDetail(
	ctx context.Context,
	user *user.User,
	featureID string,
	defaultValue bool) model.BKTEvaluationDetail[bool] {
	return model.NewEvaluationDetail(
		featureID,
		user.ID,
		"no-op",
		"no-op-name",
		0,
		model.ReasonDefault,
		defaultValue,
	)
}

func (s *nopSDK) IntVariation(ctx context.Context, user *user.User, featureID string, defaultValue int) int {
	return defaultValue
}

func (s *nopSDK) IntVariationDetail(
	ctx context.Context,
	user *user.User,
	featureID string,
	defaultValue int) model.BKTEvaluationDetail[int] {
	return model.NewEvaluationDetail(
		featureID,
		user.ID,
		"no-op",
		"no-op-name",
		0,
		model.ReasonDefault,
		defaultValue,
	)
}

func (s *nopSDK) Int64Variation(
	ctx context.Context,
	user *user.User,
	featureID string,
	defaultValue int64) int64 {
	return defaultValue
}

func (s *nopSDK) Int64VariationDetail(
	ctx context.Context,
	user *user.User,
	featureID string,
	defaultValue int64) model.BKTEvaluationDetail[int64] {
	return model.NewEvaluationDetail(
		featureID,
		user.ID,
		"no-op",
		"no-op-name",
		0,
		model.ReasonDefault,
		defaultValue,
	)
}

func (s *nopSDK) Float64Variation(
	ctx context.Context,
	user *user.User,
	featureID string,
	defaultValue float64,
) float64 {
	return defaultValue
}

func (s *nopSDK) Float64VariationDetail(
	ctx context.Context,
	user *user.User,
	featureID string,
	defaultValue float64,
) model.BKTEvaluationDetail[float64] {
	return model.NewEvaluationDetail(
		featureID,
		user.ID,
		"no-op",
		"no-op-name",
		0,
		model.ReasonDefault,
		defaultValue,
	)
}

func (s *nopSDK) StringVariation(
	ctx context.Context,
	user *user.User,
	featureID,
	defaultValue string,
) string {
	return defaultValue
}

func (s *nopSDK) StringVariationDetail(
	ctx context.Context,
	user *user.User,
	featureID,
	defaultValue string,
) model.BKTEvaluationDetail[string] {
	return model.NewEvaluationDetail(
		featureID,
		user.ID,
		"no-op",
		"no-op-name",
		0,
		model.ReasonDefault,
		defaultValue,
	)
}

func (s *nopSDK) JSONVariation(ctx context.Context, user *user.User, featureID string, defaultValue interface{}) {
}

func (s *nopSDK) ObjectVariation(
	ctx context.Context,
	user *user.User,
	featureID string,
	defaultValue interface{}) interface{} {
	return defaultValue
}

func (s *nopSDK) ObjectVariationDetail(
	ctx context.Context,
	user *user.User,
	featureID string,
	defaultValue interface{}) model.BKTEvaluationDetail[interface{}] {
	return model.NewEvaluationDetail(
		featureID,
		user.ID,
		"no-op",
		"no-op-name",
		0,
		model.ReasonDefault,
		defaultValue,
	)
}

func (s *nopSDK) Track(ctx context.Context, user *user.User, GoalID string) {
}

func (s *nopSDK) TrackValue(ctx context.Context, user *user.User, GoalID string, value float64) {
}

func (s *nopSDK) Close(ctx context.Context) error {
	return nil
}
