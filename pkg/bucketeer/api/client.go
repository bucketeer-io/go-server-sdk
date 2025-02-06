//go:generate mockgen -source=$GOFILE -package=$GOPACKAGE -destination=../../../test/mock/$GOPACKAGE/$GOFILE
package api

import (
	gwproto "github.com/bucketeer-io/bucketeer/proto/gateway"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/model"
)

// Client is the client interface for the Bucketeer APIGateway service.
type Client interface {
	GetEvaluation(req *model.GetEvaluationRequest) (*model.GetEvaluationResponse, int, error)
	GetFeatureFlags(req *model.GetFeatureFlagsRequest) (*model.GetFeatureFlagsResponse, int, error)
	GetSegmentUsers(req *model.GetSegmentUsersRequest) (*gwproto.GetSegmentUsersResponse, int, error)
	RegisterEvents(req *model.RegisterEventsRequest) (*model.RegisterEventsResponse, int, error)
}

type client struct {
	apiKey string
	host   string
}

// ClientConfig is the config for Client.
type ClientConfig struct {
	// APIKey is the key to use the Bucketeer APIGateway service.
	APIKey string

	// Host is the host name of the target service, e.g. api.example.com.
	Host string
}

// NewClient creates a new Client.
//
// NewClient returns error if failed to dial gRPC.
func NewClient(conf *ClientConfig) (Client, error) {
	client := &client{
		apiKey: conf.APIKey,
		host:   conf.Host,
	}
	return client, nil
}
