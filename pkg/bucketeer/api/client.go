//go:generate mockgen -source=$GOFILE -package=$GOPACKAGE -destination=../../../test/mock/$GOPACKAGE/$GOFILE
package api

import (
	"context"
	"errors"
	"time"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/model"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/retry"
)

var (
	ErrEmptyAPIKey      = errors.New("api key must not be empty")
	ErrInvalidScheme    = errors.New("scheme must be http or https")
	ErrEmptyAPIEndpoint = errors.New("api endpoint must not be empty")
)

// Client is the client interface for the Bucketeer APIGateway service.
type Client interface {
	// GetEvaluation retrieves evaluation for a single feature flag with retry support.
	// Used in server-side evaluation mode.
	GetEvaluation(
		ctx context.Context,
		req *model.GetEvaluationRequest,
		deadline time.Time,
	) (*model.GetEvaluationResponse, int, error)

	// GetFeatureFlags retrieves feature flags with retry support.
	GetFeatureFlags(
		ctx context.Context,
		req *model.GetFeatureFlagsRequest,
		deadline time.Time,
	) (*model.GetFeatureFlagsResponse, int, error)

	// GetSegmentUsers retrieves segment users with retry support.
	GetSegmentUsers(
		ctx context.Context,
		req *model.GetSegmentUsersRequest,
		deadline time.Time,
	) (*model.GetSegmentUsersResponse, int, error)

	// RegisterEvents registers events with retry support.
	RegisterEvents(
		ctx context.Context,
		req *model.RegisterEventsRequest,
		deadline time.Time,
	) (*model.RegisterEventsResponse, int, error)
}

type client struct {
	apiKey      string
	apiEndpoint string
	scheme      string

	// Retry configuration
	retryConfig retry.Config
}

// ClientConfig is the config for Client.
type ClientConfig struct {
	// APIKey is the key to use the Bucketeer APIGateway service.
	APIKey string

	// APIEndpoint is the backend endpoint, e.g. api.example.com.
	APIEndpoint string

	// Scheme is the scheme of the target service. This must be "http" or "https".
	Scheme string

	// Retry settings
	MaxRetries           int
	RetryInitialInterval time.Duration
	RetryMaxInterval     time.Duration
	RetryMultiplier      float64
}

// Validate validates the ClientConfig.
func (c *ClientConfig) Validate() error {
	if c.APIKey == "" {
		return ErrEmptyAPIKey
	}
	if c.Scheme != "http" && c.Scheme != "https" {
		return ErrInvalidScheme
	}
	if c.APIEndpoint == "" {
		return ErrEmptyAPIEndpoint
	}
	return nil
}

// NewClient creates a new Client.
func NewClient(conf *ClientConfig) (Client, error) {
	if err := conf.Validate(); err != nil {
		return nil, err
	}

	// Set default retry values if not provided
	maxRetries := conf.MaxRetries
	if maxRetries == 0 {
		maxRetries = 3
	}
	initialInterval := conf.RetryInitialInterval
	if initialInterval == 0 {
		initialInterval = 1 * time.Second
	}
	maxInterval := conf.RetryMaxInterval
	if maxInterval == 0 {
		maxInterval = 10 * time.Second
	}
	multiplier := conf.RetryMultiplier
	if multiplier == 0 {
		multiplier = 2.0
	}

	client := &client{
		scheme:      string(conf.Scheme),
		apiKey:      conf.APIKey,
		apiEndpoint: conf.APIEndpoint,
		retryConfig: retry.Config{
			MaxRetries:      maxRetries,
			InitialInterval: initialInterval,
			MaxInterval:     maxInterval,
			Multiplier:      multiplier,
		},
	}
	return client, nil
}
