package e2e

import (
	"flag"
	"time"
)

const (
	timeout = 10 * time.Second

	userID              = "bucketeer-go-server-user-id-1"
	featureID           = "feature-go-server-e2e-1"
	featureIDVariation1 = "value-1"
	featureIDVariation2 = "value-2"
	goalID              = "goal-go-server-e2e-1"
)

var (
	apiKey = flag.String("api-key", "", "API key for the Bucketeer APIGateway service")
	host   = flag.String("host", "", "Host name of the target service, e.g. api-dev.bucketeer.jp")
	port   = flag.Int("port", 443, "Port number of the target service, e.g. 443")
)
