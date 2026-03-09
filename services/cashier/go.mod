module github.com/zeitlos/lucity/services/cashier

go 1.26.0

replace github.com/zeitlos/lucity/pkg => ../../pkg

require (
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/stripe/stripe-go/v82 v82.1.0
	github.com/zeitlos/lucity/pkg v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.79.1
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/golang-jwt/jwt/v5 v5.2.3 // indirect
	github.com/lmittmann/tint v1.1.3 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260203192932-546029d2fa20 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)
