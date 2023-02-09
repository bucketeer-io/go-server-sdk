//go:generate mockgen -source=$GOFILE -package=$GOPACKAGE -destination=../../../test/mock/$GOPACKAGE/$GOFILE
package api

import "github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/models"

// Client is the client interface for the Bucketeer APIGateway service.
type Client interface {
	GetEvaluation(req *models.GetEvaluationRequest) (*models.GetEvaluationResponse, error)
	RegisterEvents(req *models.RegisterEventsRequest) (*models.RegisterEventsResponse, error)
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
