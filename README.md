# Bucketeer Server-side SDK for Go

[Bucketeer](https://bucketeer.io) is an open-source platform created by [CyberAgent](https://www.cyberagent.co.jp/en) to help teams make better decisions, reduce deployment lead time and release risk through feature flags. Bucketeer offers advanced features like dark launches and staged rollouts that perform limited releases based on user attributes, devices, and other segments.

[Getting started](https://docs.bucketeer.io/getting-started) using Bucketeer SDK.

## Installation

See our [documentation](https://docs.bucketeer.io/sdk/server-side/go) to install the SDK.

## Contributing

We would ❤️ for you to contribute to Bucketeer and help improve it! Anyone can use and enjoy it!

Please follow our contribution guide [here](https://docs.bucketeer.io/contribution-guide/contributing).

## Setup

#### Install the prerequisite tools

- [Go 1.15.X](https://golang.org/dl/)
- [mockgen](https://github.com/golang/mock)
- [goimports](https://pkg.go.dev/golang.org/x/tools/cmd/goimports)
- [golangci-lint](https://golangci-lint.run/usage/install/)

#### Install the dependencies

```
make deps
```

## Development

### SDK

####  Generate mocks

```
make mockgen
```

#### Format

```
make fmt
```

#### Lint

```
make lint
```

#### Build

```
make build
```

#### Run unit tests

```
make test
```

#### Run unit tests to get coverage

```
make coverage
```

#### Run e2e tests

```
make e2e API_KEY=<API_KEY> HOST=<HOST> PORT=<PORT>

# e.g.
make e2e API_KEY="xxxxxxxx" HOST="api-dev.bucketeer.jp" PORT=443
```

### Example

Go to the example directory.

```
cd example
```

#### Build

```
make build
```

#### Start the example server

```
make start TAG=<TAG> API_KEY=<API_KEY> HOST=<HOST> PORT=<PORT> \
    ENABLE_DEBUG_LOG=<ENABLE_DEBUG_LOG> FEATURE_ID=<FEATURE_ID> GOAL_ID=<GOAL_ID>

# e.g.
make start TAG="go-server" API_KEY="xxxxxxxx" HOST="api.example.com" PORT="443" \
    ENABLE_DEBUG_LOG="false" FEATURE_ID="go-example-1" GOAL_ID="go-example-goal-1"
```

#### Send requests to the example server

```
# variation
# e.g.
curl --cookie "user_id=user-1" http://localhost:8080/variation

# track
# e.g.
curl -X POST --cookie "user_id=user-1" http://localhost:8080/track
```

If you want to use a published SDK instead of a local one, you can change the [go.mod](https://github.com/bucketeer-io/go-server-sdk/blob/master/go.mod) in the example directory.

Please check the SDK [releases](https://github.com/bucketeer-io/go-server-sdk/releases)
