package bucketeer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"unsafe"

	// nolint:staticcheck
	iotag "go.opencensus.io/tag"

	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/api"
	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/event"
	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/log"
	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/user"
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

	// IntVariation returns the value of a feature flag (whose variations are ints) for the given user.
	//
	// IntVariation returns defaultValue if an error occurs.
	IntVariation(ctx context.Context, user *user.User, featureID string, defaultValue int) int

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

type sdk struct {
	tag            string
	apiClient      api.Client
	eventProcessor event.Processor
	loggers        *log.Loggers
}

const SourceIDGoServer = 5

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
	processor := event.NewProcessor(processorConf)
	return &sdk{
		tag:            dopts.tag,
		apiClient:      client,
		eventProcessor: processor,
		loggers:        loggers,
	}, nil
}

func (s *sdk) BoolVariation(ctx context.Context, user *user.User, featureID string, defaultValue bool) bool {
	evaluation, err := s.getEvaluation(ctx, user, featureID)
	if err != nil {
		s.logVariationError(err, "BoolVariation", user.ID, featureID)
		s.eventProcessor.PushDefaultEvaluationEvent(ctx, user, featureID)
		return defaultValue
	}
	variation := evaluation.VariationValue
	v, err := strconv.ParseBool(variation)
	if err != nil {
		s.logVariationError(err, "BoolVariation", user.ID, featureID)
		s.eventProcessor.PushDefaultEvaluationEvent(ctx, user, featureID)
		return defaultValue
	}
	s.eventProcessor.PushEvaluationEvent(ctx, user, evaluation)
	return v
}

func (s *sdk) IntVariation(ctx context.Context, user *user.User, featureID string, defaultValue int) int {
	evaluation, err := s.getEvaluation(ctx, user, featureID)
	if err != nil {
		s.logVariationError(err, "IntVariation", user.ID, featureID)
		s.eventProcessor.PushDefaultEvaluationEvent(ctx, user, featureID)
		return defaultValue
	}
	variation := evaluation.VariationValue
	v, err := strconv.ParseInt(variation, 10, 64)
	if err != nil {
		s.logVariationError(err, "IntVariation", user.ID, featureID)
		s.eventProcessor.PushDefaultEvaluationEvent(ctx, user, featureID)
		return defaultValue
	}
	s.eventProcessor.PushEvaluationEvent(ctx, user, evaluation)
	return int(v)
}

func (s *sdk) Int64Variation(ctx context.Context, user *user.User, featureID string, defaultValue int64) int64 {
	evaluation, err := s.getEvaluation(ctx, user, featureID)
	if err != nil {
		s.logVariationError(err, "Int64Variation", user.ID, featureID)
		s.eventProcessor.PushDefaultEvaluationEvent(ctx, user, featureID)
		return defaultValue
	}
	variation := evaluation.VariationValue
	v, err := strconv.ParseInt(variation, 10, 64)
	if err != nil {
		s.logVariationError(err, "Int64Variation", user.ID, featureID)
		s.eventProcessor.PushDefaultEvaluationEvent(ctx, user, featureID)
		return defaultValue
	}
	s.eventProcessor.PushEvaluationEvent(ctx, user, evaluation)
	return v
}

func (s *sdk) Float64Variation(ctx context.Context, user *user.User, featureID string, defaultValue float64) float64 {
	evaluation, err := s.getEvaluation(ctx, user, featureID)
	if err != nil {
		s.logVariationError(err, "Float64Variation", user.ID, featureID)
		s.eventProcessor.PushDefaultEvaluationEvent(ctx, user, featureID)
		return defaultValue
	}
	variation := evaluation.VariationValue
	v, err := strconv.ParseFloat(variation, 64)
	if err != nil {
		s.logVariationError(err, "Float64Variation", user.ID, featureID)
		s.eventProcessor.PushDefaultEvaluationEvent(ctx, user, featureID)
		return defaultValue
	}
	s.eventProcessor.PushEvaluationEvent(ctx, user, evaluation)
	return v
}

func (s *sdk) StringVariation(ctx context.Context, user *user.User, featureID, defaultValue string) string {
	evaluation, err := s.getEvaluation(ctx, user, featureID)
	if err != nil {
		s.logVariationError(err, "StringVariation", user.ID, featureID)
		s.eventProcessor.PushDefaultEvaluationEvent(ctx, user, featureID)
		return defaultValue
	}
	variation := evaluation.VariationValue
	s.eventProcessor.PushEvaluationEvent(ctx, user, evaluation)
	return variation
}

func (s *sdk) JSONVariation(ctx context.Context, user *user.User, featureID string, dst interface{}) {
	evaluation, err := s.getEvaluation(ctx, user, featureID)
	if err != nil {
		s.logVariationError(err, "JSONVariation", user.ID, featureID)
		s.eventProcessor.PushDefaultEvaluationEvent(ctx, user, featureID)
		return
	}
	variation := evaluation.VariationValue
	err = json.Unmarshal([]byte(variation), dst)
	if err != nil {
		s.logVariationError(err, "JSONVariation", user.ID, featureID)
		s.eventProcessor.PushDefaultEvaluationEvent(ctx, user, featureID)
		return
	}
	s.eventProcessor.PushEvaluationEvent(ctx, user, evaluation)
}

func (s *sdk) getEvaluation(ctx context.Context, user *user.User, featureID string) (*api.Evaluation, error) {
	if !user.Valid() {
		return nil, fmt.Errorf("invalid user: %v", user)
	}
	res, err := s.callGetEvaluationAPI(ctx, user, s.tag, featureID)
	if err != nil {
		return nil, fmt.Errorf("failed to call get evaluation api: %w", err)
	}
	if err := s.validateGetEvaluationResponse(res, featureID); err != nil {
		return nil, fmt.Errorf("invalid get evaluation response: %w", err)
	}
	return res.Evaluation, nil
}

func (s *sdk) callGetEvaluationAPI(
	ctx context.Context,
	user *user.User,
	tag, featureID string,
) (*api.GetEvaluationResponse, error) {
	var gserr error
	reqStart := time.Now()
	defer func() {
		status, ok := api.ConvertToErrStatus(gserr)
		if !ok {
			return
		}
		mutators := []iotag.Mutator{
			iotag.Insert(keyFeatureID, featureID),
			iotag.Insert(keyStatus, strconv.Itoa(status.GetStatusCode())),
		}
		ctx, err := newMetricsContext(ctx, mutators)
		if err != nil {
			return
		}

		count(ctx)
		measure(ctx, time.Since(reqStart))
	}()

	res, err := s.apiClient.GetEvaluation(&api.GetEvaluationRequest{Tag: tag, User: user, FeatureID: featureID})
	if err != nil {
		gserr = err // set HTTP status error
		status, ok := api.ConvertToErrStatus(gserr)
		if !ok {
			s.eventProcessor.PushInternalErrorCountMetricsEvent(ctx)
			return nil, fmt.Errorf("failed to get evaluation: %w", err)
		}
		if status.GetStatusCode() == http.StatusGatewayTimeout {
			s.eventProcessor.PushTimeoutErrorCountMetricsEvent(ctx)
		} else {
			s.eventProcessor.PushInternalErrorCountMetricsEvent(ctx)
		}
		return nil, fmt.Errorf("failed to get evaluation: %w", err)
	}
	s.eventProcessor.PushGetEvaluationLatencyMetricsEvent(ctx, time.Since(reqStart))
	s.eventProcessor.PushGetEvaluationSizeMetricsEvent(ctx, int(unsafe.Sizeof(res)))
	return res, nil
}

func (s *sdk) validateGetEvaluationResponse(res *api.GetEvaluationResponse, featureID string) error {
	if res == nil {
		return errors.New("res is nil")
	}
	if res.Evaluation == nil {
		return errors.New("evaluation is nil")
	}
	if res.Evaluation.FeatureID != featureID {
		return fmt.Errorf(
			"feature id doesn't match: actual %s != expected %s",
			res.Evaluation.FeatureID,
			featureID,
		)
	}
	if res.Evaluation.VariationValue == "" {
		return errors.New("variation value is empty")
	}
	return nil
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
	s.eventProcessor.PushGoalEvent(ctx, user, GoalID, value)
}

func (s *sdk) Close(ctx context.Context) error {
	if err := s.eventProcessor.Close(ctx); err != nil {
		return fmt.Errorf("bucketeer: failed to close event processor: %v", err)
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

func (s *nopSDK) BoolVariation(ctx context.Context, user *user.User, featureID string, defaultValue bool) bool {
	return defaultValue
}

func (s *nopSDK) IntVariation(ctx context.Context, user *user.User, featureID string, defaultValue int) int {
	return defaultValue
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
