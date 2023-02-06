//go:generate mockgen -source=$GOFILE -package=$GOPACKAGE -destination=../../../test/mock/$GOPACKAGE/$GOFILE
package api

// Client is the client interface for the Bucketeer APIGateway service.
type Client interface {
	GetEvaluation(req *GetEvaluationRequest) (*GetEvaluationResponse, error)
	RegisterEvents(req *RegisterEventsRequest) (*RegisterEventsResponse, error)
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
	// _, err := client.ping()
	// if err != nil {
	// 	return nil, fmt.Errorf("bucketeer/api: failed to ping to the server: %w", err)
	// }
	return client, nil
}
