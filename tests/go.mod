module github.com/zeitlos/lucity/tests

go 1.26.0

replace github.com/zeitlos/lucity/pkg => ../pkg

require github.com/zeitlos/lucity/pkg v0.0.0-00010101000000-000000000000

require (
	github.com/coreos/go-oidc/v3 v3.17.0 // indirect
	github.com/go-jose/go-jose/v4 v4.1.3 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.3 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/oauth2 v0.34.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260203192932-546029d2fa20 // indirect
	google.golang.org/grpc v1.79.1 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)
