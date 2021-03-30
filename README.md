# Bucketeer Server-side SDK for Go

- master branch, v1.X.X tag and later: It's the official implementation.
- v0-dev branch and v0.X.X tag: It's the unofficial implementation based on the [PR](https://github.com/ca-dp/bucketeer-go-server-sdk/pull/6). We don't recommend using these versions.

## Setup

Install the prerequisite tools.

- [Go 1.15.X](https://golang.org/dl/)
- [mockgen](https://github.com/golang/mock)
- [goimports](https://pkg.go.dev/golang.org/x/tools/cmd/goimports)
- [golangci-lint](https://golangci-lint.run/usage/install/)

Install the dependencies.

```
make deps
```

## Development

### SDK

Generate mocks.

```
make mockgen
```

Format.

```
make fmt
```

Lint.

```
make lint
```

Build.

```
make build
```

Run unit tests.

```
make test
```

Run unit tests to get a coverage.

```
make coverage
```

Run e2e tests.

```
make e2e API_KEY=<API_KEY> HOST=<HOST> PORT=<PORT>

# e.g.
make e2e API_KEY="xxxxxxxx" HOST="api-dev.bucketeer.jp" PORT=443
```

### Example

First, you need to move to the example directory.

```
cd example
```

Build.

```
make build
```

Start the example server.

```
make start API_KEY=<API_KEY> HOST=<HOST> PORT=<PORT> \
    ENABLE_DEBUG_LOG=<ENABLE_DEBUG_LOG> FEATURE_ID=<FEATURE_ID> GOAL_ID=<GOAL_ID>

# e.g.
make start API_KEY="xxxxxxxx" HOST="api-dev.bucketeer.jp" PORT="443" \
    ENABLE_DEBUG_LOG="false" FEATURE_ID="go-example-1" GOAL_ID="go-example-goal-1"
```

Send requests to the example server.

```
# variation
# e.g.
curl --cookie "user_id=user-1" http://localhost:8080/variation

# track
# e.g.
curl -X POST --cookie "user_id=user-1" http://localhost:8080/track
```

## SDK User Docs

- [Tutorial](https://bucketeer.io/docs/#/./server-side-sdk-tutorial-go)
- [Integration](https://bucketeer.io/docs/#/./server-side-sdk-reference-guides-go)
- [Go Doc](https://pkg.go.dev/github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer)
