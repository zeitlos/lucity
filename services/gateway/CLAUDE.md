# Gateway

GraphQL API gateway. Entry point for all frontend and external API requests. Delegates to backend services via gRPC.

## Run

```sh
go run ./cmd/gateway/...
```

## Key URLs

- GraphQL playground: http://localhost:8080/playground

## GraphQL Code Generation

```sh
go generate ./graphql/resolver.go
```

## Architecture

```
cmd/gateway/     Entry point, config, server wiring
graphql/         gqlgen configuration and resolvers
  schema/        *.graphqls files (domain-split)
  resolver.go    Root resolver, go:generate directive
  gqlgen.yml     gqlgen configuration
handler/         Business logic, gRPC client calls, type conversion
```

## Configuration

- `PORT` — HTTP listen port (default: 8080)
- `LOG_LEVEL` — Log level: debug, info, warn, error (default: info)
- `BUILDER_ADDR` — Builder gRPC address
- `PACKAGER_ADDR` — Packager gRPC address
- `DEPLOYER_ADDR` — Deployer gRPC address
