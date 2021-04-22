package api

import (
	"context"
	"crypto/tls"

	"google.golang.org/grpc/credentials"
)

type perRPCCredentials struct {
	APIKey string
}

func newPerRPCCredentials(apiKey string) credentials.PerRPCCredentials {
	return perRPCCredentials{
		APIKey: apiKey,
	}
}

func (c perRPCCredentials) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": c.APIKey,
	}, nil
}

func (c perRPCCredentials) RequireTransportSecurity() bool {
	return true
}

func newTransportCredentials() credentials.TransportCredentials {
	return credentials.NewTLS(&tls.Config{
		MinVersion: tls.VersionTLS12,
	})
}
