module github.com/zeitlos/lucity/services/webhook

go 1.26.0

replace (
	github.com/zeitlos/lucity/pkg/graceful => ../../pkg/graceful
	github.com/zeitlos/lucity/pkg/logger => ../../pkg/logger
	github.com/zeitlos/lucity/pkg/webhook => ../../pkg/webhook
)

require (
	github.com/google/go-github/v68 v68.0.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/zeitlos/lucity/pkg/graceful v0.0.0-00010101000000-000000000000
	github.com/zeitlos/lucity/pkg/logger v0.0.0-00010101000000-000000000000
)

require (
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/lmittmann/tint v1.1.3 // indirect
)
