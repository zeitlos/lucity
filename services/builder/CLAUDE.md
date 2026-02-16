# Builder

Source-to-image service. Builds container images from source code using nixpack and pushes them to the OCI registry (Zot).

## Run

```sh
go run ./cmd/builder/...
```

gRPC on `:9001`.

## Architecture

```
cmd/builder/     Entry point, config, server wiring
grpc/            gRPC service implementation
nixpack/         nixpack integration for source-to-image builds
oci/             OCI image pushing and tagging
```

## gRPC API

Defined in `proto/builder/v1/builder.proto`:

- `BuildImage` — build a container image from a source repository and push to registry

## Configuration

- `PORT` — gRPC listen port (default: 9001)
- `LOG_LEVEL` — Log level (default: info)
- `REGISTRY_URL` — OCI registry URL
