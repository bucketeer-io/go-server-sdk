#############################
# Variables
#############################
PROTO_TOP_DIR := $(shell cd ../bucketeer && pwd)
PROTO_FOLDERS := event/client feature gateway user
PROTO_OUTPUT := proto_output
IMPORT_PATH_FROM := github.com/ca-dp/bucketeer
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

#############################
# Proto
#############################
.PHONY: copy-protos
copy-protos: .gen-protos
	rm -rf proto
	mv ${PROTO_OUTPUT}/${IMPORT_PATH_FROM}/proto .

.PHONY: .gen-protos
.gen-protos:
	if [ -d ${PROTO_OUTPUT} ]; then rm -fr ${PROTO_OUTPUT}; fi; \
	mkdir ${PROTO_OUTPUT}; \
	for f in ${PROTO_FOLDERS}; do \
		protoc -I"${PROTO_TOP_DIR}" \
			-I"${GOPATH}/src/github.com/googleapis/googleapis" \
			--go_out=plugins=grpc:./${PROTO_OUTPUT} \
			${PROTO_TOP_DIR}/proto/$$f/*.proto; \
	done; \
	for file in $$(find ./${PROTO_OUTPUT} -name "*.go"); do \
		sed -i '' 's|${IMPORT_PATH_FROM}|${IMPORT_PATH_TO}|g' $$file; \
	done
