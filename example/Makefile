.PHONY: all
all: deps fmt lint build

.PHONY: deps
deps:
	go mod tidy
	go mod vendor

.PHONY: fmt
fmt:
	goimports -local github.com/ca-dp/bucketeer-go-server-sdk/example -w .

.PHONY: lint
lint:
	golangci-lint run --config ../.golangci.yml ./...

.PHONY: build
build:
	go build -o main ./...


.PHONY: start
start:
	./main
