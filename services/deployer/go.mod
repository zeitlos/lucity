module github.com/zeitlos/lucity/services/deployer

go 1.26.0

replace (
	github.com/zeitlos/lucity/pkg/deployer => ../../pkg/deployer
	github.com/zeitlos/lucity/pkg/graceful => ../../pkg/graceful
	github.com/zeitlos/lucity/pkg/logger => ../../pkg/logger
)

require github.com/kelseyhightower/envconfig v1.4.0
