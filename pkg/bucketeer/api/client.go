//go:generate mockgen -source=$GOFILE -package=$GOPACKAGE -destination=../../../test/mock/$GOPACKAGE/$GOFILE
package api

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	protogateway "github.com/ca-dp/bucketeer-go-server-sdk/proto/gateway"
)

// Client is the client interface for the Bucketeer APIGateway service.
type Client interface {
	// GatewayClient defines GetEvaluation and RegisterEvents methods.
	protogateway.GatewayClient

	// Close tears down the connection.
	Close() error
}

type client struct {
	protogateway.GatewayClient
	conn *grpc.ClientConn
}

// ClientConfig is the config for Client.
type ClientConfig struct {
	// APIKey is the key to use the Bucketeer APIGateway service.
	APIKey string

	// Host is the host name of the target service, e.g. api-dev.bucketeer.jp.
	Host string

	// Port is the port number of the target service, e.g. 443.
	Port int
}

// NewClient creates a new Client.
//
// NewClient returns error if failed to dial gRPC.
func NewClient(ctx context.Context, conf *ClientConfig) (Client, error) {
	perRPCCreds := newPerRPCCredentials(conf.APIKey)
	transportCreds := newTransportCredentials()
	dialOptions := []grpc.DialOption{
		grpc.WithPerRPCCredentials(perRPCCreds),
		grpc.WithTransportCredentials(transportCreds),
		grpc.WithBlock(),
	}
	conn, err := grpc.DialContext(
		ctx,
		fmt.Sprintf("%s:%d", conf.Host, conf.Port),
		dialOptions...,
	)
	if err != nil {
		return nil, fmt.Errorf("bucketeer/api: failed to dial gRPC: %w", err)
	}
	return &client{
		GatewayClient: protogateway.NewGatewayClient(conn),
		conn:          conn,
	}, nil
}

func (c *client) Close() error {
	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("bucketeer/api: failed to close conn: %v", err)
	}
	return nil
}
