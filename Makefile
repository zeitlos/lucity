.PHONY: build proto dev-gateway dev-builder dev-packager dev-deployer dev-webhook dev-dashboard generate-graphql lint

# Build all Go services
build:
	go build ./services/gateway/cmd/gateway/...
	go build ./services/builder/cmd/builder/...
	go build ./services/packager/cmd/packager/...
	go build ./services/deployer/cmd/deployer/...
	go build ./services/webhook/cmd/webhook/...

# Generate protobuf code (requires protoc, protoc-gen-go, protoc-gen-go-grpc)
proto:
	cd pkg/builder && go generate ./...
	cd pkg/deployer && go generate ./...
	cd pkg/packager && go generate ./...
	cd pkg/webhook && go generate ./...

# Generate GraphQL resolvers (requires gqlgen)
generate-graphql:
	cd services/gateway && go generate ./graphql/resolver.go

# Run individual services
dev-gateway:
	cd services/gateway && set -a && . .env 2>/dev/null && set +a && go run ./cmd/gateway/...

dev-builder:
	cd services/builder && set -a && . .env 2>/dev/null && set +a && go run ./cmd/builder/...

dev-packager:
	cd services/packager && set -a && . .env 2>/dev/null && set +a && go run ./cmd/packager/...

dev-deployer:
	cd services/deployer && set -a && . .env 2>/dev/null && set +a && go run ./cmd/deployer/...

dev-webhook:
	cd services/webhook && set -a && . .env 2>/dev/null && set +a && go run ./cmd/webhook/...

dev-dashboard:
	cd services/dashboard && npm run dev

# Lint
lint:
	cd services/dashboard && npm run lint

# Sync workspace
sync:
	go work sync
