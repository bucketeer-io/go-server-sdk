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
		defaultValue bool) model.EvaluationDetail[bool]

	// IntVariation returns the value of a feature flag (whose variations are ints) for the given user.
	//
	// IntVariation returns defaultValue if an error occurs.
	IntVariation(ctx context.Context, user *user.User, featureID string, defaultValue int) int

	IntVariationDetail(
		ctx context.Context,
		user *user.User,
		featureID string,
		defaultValue int) model.EvaluationDetail[int]

	// Int64Variation returns the value of a feature flag (whose variations are int64s) for the given user.
	//
	// Int64Variation returns defaultValue if an error occurs.
	Int64Variation(ctx context.Context, user *user.User, featureID string, defaultValue int64) int64

	// Float64Variation returns the value of a feature flag (whose variations are float64s) for the given user.
	//
	// Float64Variation returns defaultValue if an error occurs.
	Float64Variation(ctx context.Context, user *user.User, featureID string, defaultValue float64) float64

	// StringVariation returns the value of a feature flag (whose variations are strings) for the given user.
	//
	// StringVariation returns defaultValue if an error occurs.
	StringVariation(ctx context.Context, user *user.User, featureID, defaultValue string) string

	// JSONVariation parses the value of a feature flag (whose variations are jsons) for the given user,
	// and stores the result in dst.
	JSONVariation(ctx context.Context, user *user.User, featureID string, dst interface{})

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
	evaluation, err := s.getEvaluation(ctx, user, featureID)
	if err != nil {
		s.logVariationError(err, "BoolVariation", user.ID, featureID)
		s.eventProcessor.PushDefaultEvaluationEvent(user, featureID)
		return defaultValue
	}
	variation := evaluation.VariationValue
	v, err := strconv.ParseBool(variation)
	if err != nil {
		s.logVariationError(err, "BoolVariation", user.ID, featureID)
		s.eventProcessor.PushDefaultEvaluationEvent(user, featureID)
		return defaultValue
	}
	s.eventProcessor.PushEvaluationEvent(user, evaluation)
	return v
}

func (s *sdk) BoolVariationDetail(
	ctx context.Context,
	user *user.User,
	featureID string,
	defaultValue bool) model.EvaluationDetail[bool] {
	evaluation, err := s.getEvaluation(ctx, user, featureID)
	if err != nil {
		s.logVariationError(err, "BoolVariationDetail", user.ID, featureID)
		s.eventProcessor.PushDefaultEvaluationEvent(user, featureID)
		return model.NewEvaluationDetail[bool](
			featureID,
			user.ID,
			"",
			0,
			model.EvaluationReasonClient,
			defaultValue,
		)
	}
	variation := evaluation.VariationValue
	v, err := strconv.ParseBool(variation)
	if err != nil {
		s.logVariationError(err, "BoolVariationDetail", user.ID, featureID)
		s.eventProcessor.PushDefaultEvaluationEvent(user, featureID)
		return model.NewEvaluationDetail[bool](
			featureID,
			user.ID,
			evaluation.VariationID,
			evaluation.FeatureVersion,
			model.ConvertEvaluationReason(evaluation.Reason.Type),
			defaultValue,
		)
	}
	s.eventProcessor.PushEvaluationEvent(user, evaluation)
	return model.NewEvaluationDetail[bool](
		featureID,
		user.ID,
		evaluation.VariationID,
		evaluation.FeatureVersion,
		model.ConvertEvaluationReason(evaluation.Reason.Type),
		v,
	)
}

func (s *sdk) IntVariation(ctx context.Context, user *user.User, featureID string, defaultValue int) int {
	evaluation, err := s.getEvaluation(ctx, user, featureID)
	if err != nil {
		s.logVariationError(err, "IntVariation", user.ID, featureID)
		s.eventProcessor.PushDefaultEvaluationEvent(user, featureID)
		return defaultValue
	}
	variation := evaluation.VariationValue
	v, err := strconv.ParseInt(variation, 10, 64)
	if err != nil {
		s.logVariationError(err, "IntVariation", user.ID, featureID)
		s.eventProcessor.PushDefaultEvaluationEvent(user, featureID)
		return defaultValue
	}
	s.eventProcessor.PushEvaluationEvent(user, evaluation)
	return int(v)
}

func (s *sdk) IntVariationDetail(
	ctx context.Context,
	user *user.User,
	featureID string,
	defaultValue int) model.EvaluationDetail[int] {
	evaluation, err := s.getEvaluation(ctx, user, featureID)
	if err != nil {
		s.logVariationError(err, "IntVariationDetail", user.ID, featureID)
		s.eventProcessor.PushDefaultEvaluationEvent(user, featureID)
		return model.NewEvaluationDetail[int](featureID, user.ID, "", 0, model.EvaluationReasonClient, defaultValue)
	}
	variation := evaluation.VariationValue
	v, err := strconv.ParseInt(variation, 10, 64)
	if err != nil {
		s.logVariationError(err, "IntVariationDetail", user.ID, featureID)
		s.eventProcessor.PushDefaultEvaluationEvent(user, featureID)
		return model.NewEvaluationDetail[int](
			featureID,
			user.ID,
			evaluation.VariationID,
			evaluation.FeatureVersion,
			model.EvaluationReasonClient,
			defaultValue,
		)
	}
	s.eventProcessor.PushEvaluationEvent(user, evaluation)
	return model.NewEvaluationDetail[int](
		featureID,
		user.ID,
		evaluation.VariationID,
		evaluation.FeatureVersion,
		model.ConvertEvaluationReason(evaluation.Reason.Type),
		int(v),
	)
}

func (s *sdk) Int64Variation(ctx context.Context, user *user.User, featureID string, defaultValue int64) int64 {
	evaluation, err := s.getEvaluation(ctx, user, featureID)
	if err != nil {
		s.logVariationError(err, "Int64Variation", user.ID, featureID)
		s.eventProcessor.PushDefaultEvaluationEvent(user, featureID)
		return defaultValue
	}
	variation := evaluation.VariationValue
	v, err := strconv.ParseInt(variation, 10, 64)
	if err != nil {
		s.logVariationError(err, "Int64Variation", user.ID, featureID)
		s.eventProcessor.PushDefaultEvaluationEvent(user, featureID)
		return defaultValue
	}
	s.eventProcessor.PushEvaluationEvent(user, evaluation)
	return v
}

func (s *sdk) Float64Variation(ctx context.Context, user *user.User, featureID string, defaultValue float64) float64 {
	evaluation, err := s.getEvaluation(ctx, user, featureID)
	if err != nil {
		s.logVariationError(err, "Float64Variation", user.ID, featureID)
		s.eventProcessor.PushDefaultEvaluationEvent(user, featureID)
		return defaultValue
	}
	variation := evaluation.VariationValue
	v, err := strconv.ParseFloat(variation, 64)
	if err != nil {
		s.logVariationError(err, "Float64Variation", user.ID, featureID)
		s.eventProcessor.PushDefaultEvaluationEvent(user, featureID)
		return defaultValue
	}
	s.eventProcessor.PushEvaluationEvent(user, evaluation)
	return v
}

func (s *sdk) StringVariation(ctx context.Context, user *user.User, featureID, defaultValue string) string {
	evaluation, err := s.getEvaluation(ctx, user, featureID)
	if err != nil {
		s.logVariationError(err, "StringVariation", user.ID, featureID)
		s.eventProcessor.PushDefaultEvaluationEvent(user, featureID)
		return defaultValue
	}
	variation := evaluation.VariationValue
	s.eventProcessor.PushEvaluationEvent(user, evaluation)
	return variation
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
	defaultValue bool) model.EvaluationDetail[bool] {
	return model.NewEvaluationDetail[bool](
		featureID,
		user.ID,
		"no-op",
		0,
		model.EvaluationReasonDefault,
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
	defaultValue int) model.EvaluationDetail[int] {
	return model.EvaluationDetail[int]{
		FeatureID:      featureID,
		FeatureVersion: 0,
		UserID:         user.ID,
		VariationID:    "no-op",
		Reason:         model.EvaluationReasonDefault,
		Value:          defaultValue,
	}
}

func (s *nopSDK) Int64Variation(ctx context.Context, user *user.User, featureID string, defaultValue int64) int64 {
	return defaultValue
}

func (s *nopSDK) Float64Variation(
	ctx context.Context,
	user *user.User,
	featureID string,
	defaultValue float64,
) float64 {
	return defaultValue
}

func (s *nopSDK) StringVariation(
	ctx context.Context,
	user *user.User,
	featureID,
	defaultValue string,
) string {
	return defaultValue
}

func (s *nopSDK) JSONVariation(ctx context.Context, user *user.User, featureID string, dst interface{}) {
}

func (s *nopSDK) Track(ctx context.Context, user *user.User, GoalID string) {
}

func (s *nopSDK) TrackValue(ctx context.Context, user *user.User, GoalID string, value float64) {
}

func (s *nopSDK) Close(ctx context.Context) error {
	return nil
}
