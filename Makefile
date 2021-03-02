#############################
# Go
#############################
.PHONY: all
all: deps fmt lint build test

.PHONY: deps
deps:
	go mod tidy
	go mod vendor

.PHONY: fmt
fmt:
	gofmt -s -w ./pkg

.PHONY: fmt-check
fmt-check:
	test -z "$$(gofmt -d ./pkg)"

.PHONY: lint
lint:
	golangci-lint run ./pkg/...

.PHONY: build
build:
	go build ./pkg/...

.PHONY: test
test:
	go test -race ./pkg/...

.PHONY: coverage
coverage:
	go test -race -v -coverpkg=./pkg/... -covermode=atomic -coverprofile=coverage.out ./pkg/...

#############################
# Proto
#############################
PROTO_TOP_DIR := $(shell cd ../bucketeer && pwd)
PROTO_FOLDERS := event/client feature gateway user
PROTO_OUTPUT := proto_output
IMPORT_PATH_FROM := github.com/ca-dp/bucketeer
IMPORT_PATH_TO := github.com/ca-dp/bucketeer-go-server-sdk

.PHONY: copy-protos
copy-protos: .gen-protos
	rm -rf proto
	mv $(PROTO_OUTPUT)/$(IMPORT_PATH_FROM)/proto .

.PHONY: .gen-protos
.gen-protos:
	if [ -d ${PROTO_OUTPUT} ]; then rm -fr ${PROTO_OUTPUT}; fi; \
	mkdir ${PROTO_OUTPUT}; \
	for f in ${PROTO_FOLDERS}; do \
		protoc -I"$(PROTO_TOP_DIR)" \
			-I"${GOPATH}/src/github.com/googleapis/googleapis" \
			--go_out=plugins=grpc:./$(PROTO_OUTPUT) \
			$(PROTO_TOP_DIR)/proto/$$f/*.proto; \
	done; \
	for file in $$(find ./${PROTO_OUTPUT} -name "*.go"); do \
		sed -i '' 's|${IMPORT_PATH_FROM}|${IMPORT_PATH_TO}|g' $$file; \
	done
