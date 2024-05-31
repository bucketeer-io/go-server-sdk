module github.com/bucketeer-io/go-server-sdk/example

go 1.21

toolchain go1.21.10

require github.com/bucketeer-io/go-server-sdk v0.0.0-00010101000000-000000000000

require (
	github.com/bucketeer-io/bucketeer/proto v0.0.0-20240520044049-69f55ef9a09f // indirect
	go.opencensus.io v0.24.0 // indirect
	golang.org/x/net v0.25.0 // indirect
	golang.org/x/sys v0.20.0 // indirect
	golang.org/x/text v0.15.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240513163218-0867130af1f8 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240509183442-62759503f434 // indirect
	google.golang.org/grpc v1.64.0 // indirect
	google.golang.org/protobuf v1.34.1 // indirect
)

replace github.com/bucketeer-io/go-server-sdk => ../
