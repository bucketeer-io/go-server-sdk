//go:generate mockgen -source=$GOFILE -package=$GOPACKAGE -destination=../../../test/mock/$GOPACKAGE/$GOFILE
package api

import (
	"errors"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/model"
)

var (
	ErrEmptyAPIKey   = errors.New("api key must not be empty")
	ErrInvalidScheme = errors.New("scheme must be http or https")
	ErrEmptyHost     = errors.New("host must not be empty")
)

// Client is the client interface for the Bucketeer APIGateway service.
type Client interface {
	GetEvaluation(req *model.GetEvaluationRequest) (*model.GetEvaluationResponse, int, error)
	GetFeatureFlags(req *model.GetFeatureFlagsRequest) (*model.GetFeatureFlagsResponse, int, error)
	GetSegmentUsers(req *model.GetSegmentUsersRequest) (*model.GetSegmentUsersResponse, int, error)
	RegisterEvents(req *model.RegisterEventsRequest) (*model.RegisterEventsResponse, int, error)
}

type client struct {
	apiKey string
	scheme string
	host   string
}

// ClientConfig is the config for Client.
type ClientConfig struct {
	// APIKey is the key to use the Bucketeer APIGateway service.
	APIKey string

	// Scheme is the scheme of the target service. This must be "http" or "https".
	Scheme string

	// Host is the host name of the target service, e.g. api.example.com.
	Host string
}

// Validate validates the ClientConfig.
func (c *ClientConfig) Validate() error {
	if c.APIKey == "" {
		return ErrEmptyAPIKey
	}
	if c.Scheme != "http" && c.Scheme != "https" {
		return ErrInvalidScheme
	}
	if c.Host == "" {
		return ErrEmptyHost
	}
	return nil
}

// NewClient creates a new Client.
func NewClient(conf *ClientConfig) (Client, error) {
	client := &client{
		scheme: string(conf.Scheme),
		apiKey: conf.APIKey,
		host:   conf.Host,
	}
	if err := conf.Validate(); err != nil {
		return nil, err
	}
	return client, nil
}
