#############################
# Variables
#############################
IMPORT_PATH_TO := github.com/ca-dp/bucketeer-go-server-sdk

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
	go test -race ./pkg/...

.PHONY: coverage
coverage:
	go test -race -covermode=atomic -coverprofile=coverage.out -coverpkg=./pkg/... ./pkg/...

.PHONY: e2e
e2e:
	go test -race ./test/e2e/... \
		-args -api-key=${API_KEY} -host=${HOST} -port=${PORT}
