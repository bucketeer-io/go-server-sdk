//go:generate mockgen -source=$GOFILE -package=$GOPACKAGE -destination=../../../test/mock/$GOPACKAGE/$GOFILE
package api

import "github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/user"

// Client is the client interface for the Bucketeer APIGateway service.
type RestClient interface {
	GetEvaluation(user *user.User, tag, featureID string) (*GetEvaluationResponse, error)
	RegisterEvents(events []*Event) (*RegisterEventsResponse, error)
}

type client struct {
	// APIKey is the key to use the Bucketeer APIGateway service.
	apiKey string
	// Host is the host name of the target service, e.g. api.example.com.
	host string
}

// NewClient creates a new Client.
//
// NewClient returns error if failed to dial gRPC.
func NewRestClient(apiKey, host string) RestClient {
	return &client{
		apiKey: apiKey,
		host:   host,
	}
}
