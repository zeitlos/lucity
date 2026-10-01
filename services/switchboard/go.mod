module github.com/zeitlos/lucity/services/switchboard

go 1.26.0

replace github.com/zeitlos/lucity/pkg => ../../pkg

require (
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/zeitlos/lucity/pkg v0.0.0-00010101000000-000000000000
)

require github.com/lmittmann/tint v1.2.0 // indirect
