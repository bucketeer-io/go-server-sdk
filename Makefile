#############################
# Variables
#############################
IMPORT_PATH_TO := github.com/bucketeer-io/go-server-sdk
GOTESTSUM_VERSION := v1.13.0

#############################
# Go
#############################
.PHONY: all
all: deps mockgen fmt lint build test

.PHONY: deps
deps:
	go mod tidy
	go mod vendor

.PHONY: mockgen
mockgen:
	go generate -run="mockgen" ./pkg/...
	make fmt

.PHONY: fmt
fmt:
	goimports -local ${IMPORT_PATH_TO} -w ./pkg ./test

.PHONY: fmt-check
fmt-check:
	test -z "$$(goimports -local ${IMPORT_PATH_TO} -d ./pkg ./test)"

.PHONY: lint
lint:
	golangci-lint run ./pkg/... ./test/...

.PHONY: build
build:
	go build ./pkg/... ./test/...

.PHONY: test
test:
	TZ=UTC CGO_ENABLED=1 go run gotest.tools/gotestsum@$(GOTESTSUM_VERSION) \
		--format pkgname \
		-- -v -race ./pkg/...

.PHONY: coverage
coverage:
	go test -race -covermode=atomic -coverprofile=coverage.out -coverpkg=./pkg/... ./pkg/...

.PHONY: e2e
e2e:
	TZ=UTC CGO_ENABLED=1 go run gotest.tools/gotestsum@$(GOTESTSUM_VERSION) \
		--format pkgname \
		-- -v -race ./test/e2e/... -args \
		-api-key=${API_KEY} \
		-api-key-server=${API_KEY_SERVER} \
		-api-endpoint=${API_ENDPOINT} \
		-scheme=${SCHEME}
