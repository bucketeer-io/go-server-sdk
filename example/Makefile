.PHONY: all
all: deps fmt lint build

.PHONY: deps
deps:
	go mod tidy
	go mod vendor

.PHONY: fmt
fmt:
	goimports -local github.com/bucketeer-io/go-server-sdk/example -w .

.PHONY: lint
lint:
	golangci-lint run --config ../.golangci.yml ./...

.PHONY: build
build:
	go build -o main ./...

.PHONY: start
start:
	./main \
		-bucketeer-tag=${TAG} \
		-bucketeer-api-key=${API_KEY} \
		-bucketeer-api-endpoint=${API_ENDPOINT} \
		-bucketeer-enable-debug-log=${ENABLE_DEBUG_LOG} \
		-bucketeer-feature-id=${FEATURE_ID} \
		-bucketeer-goal-id=${GOAL_ID}
