package bucketeer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang/protobuf/proto" // nolint:staticcheck
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/api"
	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/event"
	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/log"
	protofeature "github.com/ca-dp/bucketeer-go-server-sdk/proto/feature"
	protogateway "github.com/ca-dp/bucketeer-go-server-sdk/proto/gateway"
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
	BoolVariation(ctx context.Context, user *User, featureID string, defaultValue bool) bool

	// IntVariation returns the value of a feature flag (whose variations are ints) for the given user.
	//
	// IntVariation returns defaultValue if an error occurs.
	IntVariation(ctx context.Context, user *User, featureID string, defaultValue int) int

	// Int64Variation returns the value of a feature flag (whose variations are int64s) for the given user.
	//
	// Int64Variation returns defaultValue if an error occurs.
	Int64Variation(ctx context.Context, user *User, featureID string, defaultValue int64) int64

	// Float64Variation returns the value of a feature flag (whose variations are float64s) for the given user.
	//
	// Float64Variation returns defaultValue if an error occurs.
	Float64Variation(ctx context.Context, user *User, featureID string, defaultValue float64) float64

	// StringVariation returns the value of a feature flag (whose variations are strings) for the given user.
	//
	// StringVariation returns defaultValue if an error occurs.
	StringVariation(ctx context.Context, user *User, featureID, defaultValue string) string

	// JSONVariation parses the value of a feature flag (whose variations are jsons) for the given user,
	// and stores the result in dst.
	JSONVariation(ctx context.Context, user *User, featureID string, dst interface{})

	// Track reports that a user has performed a goal event.
	//
	// TODO: Track doesn't work correctly until Bucketeer service implements the new goal tracking architecture.
	Track(ctx context.Context, user *User, goalID string)

	// TrackValue reports that a user has performed a goal event, and associates it with a custom value.
	//
	// TODO: TrackValue doesn't work correctly until Bucketeer service implements the new goal tracking architecture.
	TrackValue(ctx context.Context, user *User, goalID string, value float64)

	// Close tears down all SDK activities and resources, after ensuring that all events have been delivered.
	Close()
}

type sdk struct {
	tag            string
	apiClient      api.Client
	eventProcessor event.Processor
	loggers        *log.Loggers
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
	clientConf := &api.ClientConfig{
		APIKey: dopts.apiKey,
		Host:   dopts.host,
		Port:   dopts.port,
	}
	client, err := api.NewClient(ctx, clientConf)
	if err != nil {
		return nil, fmt.Errorf("bucketeer: failed to new api client: %w", err)
	}
	processorConf := &event.ProcessorConfig{
		NumFlushWorkers: dopts.numEventFlushWorkers,
		FlushInterval:   dopts.eventFlushInterval,
		FlushSize:       dopts.eventFlushSize,
		QueueCapacity:   dopts.eventQueueCapacity,
		APIClient:       client,
		Loggers:         loggers,
	}
	processor := event.NewProcessor(processorConf)
	return &sdk{
		tag:            dopts.tag,
		apiClient:      client,
		eventProcessor: processor,
		loggers:        loggers,
	}, nil
}

func (s *sdk) BoolVariation(ctx context.Context, user *User, featureID string, defaultValue bool) bool {
	evaluation, err := s.getEvaluation(ctx, user, featureID)
	if err != nil {
		s.logVariationError(err, "BoolVariation", user.Id, featureID)
		s.eventProcessor.PushDefaultEvaluationEvent(ctx, user.User, featureID)
		return defaultValue
	}
	variation := evaluation.Variation.Value
	v, err := strconv.ParseBool(variation)
	if err != nil {
		s.logVariationError(err, "BoolVariation", user.Id, featureID)
		s.eventProcessor.PushDefaultEvaluationEvent(ctx, user.User, featureID)
		return defaultValue
	}
	s.eventProcessor.PushEvaluationEvent(ctx, user.User, evaluation)
	return v
}

func (s *sdk) IntVariation(ctx context.Context, user *User, featureID string, defaultValue int) int {
	evaluation, err := s.getEvaluation(ctx, user, featureID)
	if err != nil {
		s.logVariationError(err, "IntVariation", user.Id, featureID)
		s.eventProcessor.PushDefaultEvaluationEvent(ctx, user.User, featureID)
		return defaultValue
	}
	variation := evaluation.Variation.Value
	v, err := strconv.ParseInt(variation, 10, 64)
	if err != nil {
		s.logVariationError(err, "IntVariation", user.Id, featureID)
		s.eventProcessor.PushDefaultEvaluationEvent(ctx, user.User, featureID)
		return defaultValue
	}
	s.eventProcessor.PushEvaluationEvent(ctx, user.User, evaluation)
	return int(v)
}

func (s *sdk) Int64Variation(ctx context.Context, user *User, featureID string, defaultValue int64) int64 {
	evaluation, err := s.getEvaluation(ctx, user, featureID)
	if err != nil {
		s.logVariationError(err, "Int64Variation", user.Id, featureID)
		s.eventProcessor.PushDefaultEvaluationEvent(ctx, user.User, featureID)
		return defaultValue
	}
	variation := evaluation.Variation.Value
	v, err := strconv.ParseInt(variation, 10, 64)
	if err != nil {
		s.logVariationError(err, "Int64Variation", user.Id, featureID)
		s.eventProcessor.PushDefaultEvaluationEvent(ctx, user.User, featureID)
		return defaultValue
	}
	s.eventProcessor.PushEvaluationEvent(ctx, user.User, evaluation)
	return v
}

func (s *sdk) Float64Variation(ctx context.Context, user *User, featureID string, defaultValue float64) float64 {
	evaluation, err := s.getEvaluation(ctx, user, featureID)
	if err != nil {
		s.logVariationError(err, "Float64Variation", user.Id, featureID)
		s.eventProcessor.PushDefaultEvaluationEvent(ctx, user.User, featureID)
		return defaultValue
	}
	variation := evaluation.Variation.Value
	v, err := strconv.ParseFloat(variation, 64)
	if err != nil {
		s.logVariationError(err, "Float64Variation", user.Id, featureID)
		s.eventProcessor.PushDefaultEvaluationEvent(ctx, user.User, featureID)
		return defaultValue
	}
	s.eventProcessor.PushEvaluationEvent(ctx, user.User, evaluation)
	return v
}

func (s *sdk) StringVariation(ctx context.Context, user *User, featureID, defaultValue string) string {
	evaluation, err := s.getEvaluation(ctx, user, featureID)
	if err != nil {
		s.logVariationError(err, "StringVariation", user.Id, featureID)
		s.eventProcessor.PushDefaultEvaluationEvent(ctx, user.User, featureID)
		return defaultValue
	}
	variation := evaluation.Variation.Value
	s.eventProcessor.PushEvaluationEvent(ctx, user.User, evaluation)
	return variation
}

func (s *sdk) JSONVariation(ctx context.Context, user *User, featureID string, dst interface{}) {
	evaluation, err := s.getEvaluation(ctx, user, featureID)
	if err != nil {
		s.logVariationError(err, "JSONVariation", user.Id, featureID)
		s.eventProcessor.PushDefaultEvaluationEvent(ctx, user.User, featureID)
		return
	}
	variation := evaluation.Variation.Value
	err = json.Unmarshal([]byte(variation), dst)
	if err != nil {
		s.logVariationError(err, "JSONVariation", user.Id, featureID)
		s.eventProcessor.PushDefaultEvaluationEvent(ctx, user.User, featureID)
		return
	}
	s.eventProcessor.PushEvaluationEvent(ctx, user.User, evaluation)
}

func (s *sdk) getEvaluation(ctx context.Context, user *User, featureID string) (*protofeature.Evaluation, error) {
	req := &protogateway.GetEvaluationRequest{
		Tag:       s.tag,
		User:      user.User,
		FeatureId: featureID,
	}
	reqStart := time.Now()
	res, err := s.apiClient.GetEvaluation(ctx, req)
	if err != nil || res == nil {
		if status.Code(err) == codes.DeadlineExceeded {
			s.eventProcessor.PushTimeoutErrorCountMetricsEvent(ctx, s.tag)
		} else {
			s.eventProcessor.PushInternalErrorCountMetricsEvent(ctx, s.tag)
		}
		return nil, fmt.Errorf("failed to get evaluation: %w", err)
	}
	s.eventProcessor.PushGetEvaluationLatencyMetricsEvent(ctx, time.Since(reqStart), s.tag)
	s.eventProcessor.PushGetEvaluationSizeMetricsEvent(ctx, proto.Size(res), s.tag)
	if err := s.validateGetEvaluationResponse(res, featureID); err != nil {
		return nil, fmt.Errorf("invalid get evaluation response: %w", err)
	}
	return res.Evaluation, nil
}

// Require res is not nil.
func (s *sdk) validateGetEvaluationResponse(res *protogateway.GetEvaluationResponse, featureID string) error {
	if res.Evaluation == nil {
		return errors.New("evaluation is nil")
	}
	if res.Evaluation.FeatureId != featureID {
		return fmt.Errorf(
			"feature id doesn't match: actual %s != expected %s",
			res.Evaluation.FeatureId,
			featureID,
		)
	}
	if res.Evaluation.Variation == nil {
		return errors.New("variation is nil")
	}
	return nil
}

func (s *sdk) logVariationError(err error, methodName, userID, featureID string) {
	s.loggers.Errorf(
		"bucketeer: %s returns default value (err: %v, userID: %s, featureID: %s)",
		methodName,
		err,
		userID,
		featureID,
	)
}

func (s *sdk) Track(ctx context.Context, user *User, goalID string) {
	s.TrackValue(ctx, user, goalID, 0.0)
}

func (s *sdk) TrackValue(ctx context.Context, user *User, goalID string, value float64) {
	s.eventProcessor.PushGoalEvent(ctx, user.User, goalID, value)
}

func (s *sdk) Close() {
	// TODO: implement later
}
