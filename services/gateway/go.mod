module github.com/zeitlos/lucity/services/gateway

go 1.26.0

replace (
	github.com/zeitlos/lucity/pkg/auth => ../../pkg/auth
	github.com/zeitlos/lucity/pkg/graceful => ../../pkg/graceful
	github.com/zeitlos/lucity/pkg/logger => ../../pkg/logger
)

require (
	github.com/99designs/gqlgen v0.17.86
	github.com/go-playground/validator/v10 v10.30.1
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/rs/cors v1.11.1
	github.com/vektah/gqlparser/v2 v2.5.31
	github.com/zeitlos/lucity/pkg/graceful v0.0.0-00010101000000-000000000000
	github.com/zeitlos/lucity/pkg/logger v0.0.0-00010101000000-000000000000
)

require (
	github.com/agnivade/levenshtein v1.2.1 // indirect
	github.com/gabriel-vasile/mimetype v1.4.12 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/lmittmann/tint v1.1.3 // indirect
	github.com/sosodev/duration v1.3.1 // indirect
	golang.org/x/crypto v0.46.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.33.0 // indirect
)
