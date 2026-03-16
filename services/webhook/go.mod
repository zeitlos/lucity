module github.com/zeitlos/lucity/services/webhook

go 1.26.0

replace github.com/zeitlos/lucity/pkg => ../../pkg

require (
	github.com/google/go-github/v68 v68.0.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/zeitlos/lucity/pkg v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.79.1
)

require (
	github.com/bradleyfalzon/ghinstallation/v2 v2.17.0 // indirect
	github.com/coreos/go-oidc/v3 v3.17.0 // indirect
	github.com/go-jose/go-jose/v4 v4.1.3 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.2 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.3 // indirect
	github.com/google/go-github/v75 v75.0.0 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/lmittmann/tint v1.1.3 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/oauth2 v0.34.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260203192932-546029d2fa20 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)
