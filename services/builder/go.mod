module github.com/zeitlos/lucity/services/builder

go 1.26.0

replace (
	github.com/zeitlos/lucity/pkg/builder => ../../pkg/builder
	github.com/zeitlos/lucity/pkg/graceful => ../../pkg/graceful
	github.com/zeitlos/lucity/pkg/logger => ../../pkg/logger
)

require (
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/zeitlos/lucity/pkg/builder v0.0.0-00010101000000-000000000000
	github.com/zeitlos/lucity/pkg/graceful v0.0.0-00010101000000-000000000000
	github.com/zeitlos/lucity/pkg/logger v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.79.1
)

require (
	github.com/lmittmann/tint v1.1.3 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)
