module github.com/zeitlos/lucity/services/webhook

go 1.26.0

replace (
	github.com/zeitlos/lucity/pkg/graceful => ../../pkg/graceful
	github.com/zeitlos/lucity/pkg/logger => ../../pkg/logger
	github.com/zeitlos/lucity/pkg/webhook => ../../pkg/webhook
)

require github.com/kelseyhightower/envconfig v1.4.0
