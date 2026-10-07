module github.com/zeitlos/lucity/services/exporter

go 1.27.0

replace github.com/zeitlos/lucity/pkg => ../../pkg

require (
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/ovh/go-ovh v1.9.0
	github.com/prometheus/client_golang v1.24.1
	github.com/zeitlos/lucity/pkg v0.0.0-20260706084318-3ef7184f18d8
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/klauspost/compress v1.20.1 // indirect
	github.com/lmittmann/tint v1.2.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/prometheus/client_model v0.6.3 // indirect
	github.com/prometheus/common v0.72.0 // indirect
	github.com/prometheus/procfs v0.22.0 // indirect
	golang.org/x/oauth2 v0.37.0 // indirect
	golang.org/x/sys v0.48.0 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
	gopkg.in/ini.v1 v1.67.3 // indirect
)
