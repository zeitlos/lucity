# Packager

Helm values generation and GitOps repository management. Creates and manages Soft-serve repos, generates lucity-app chart values, and handles ejection.

## Run

```sh
go run ./cmd/packager/...
```

gRPC on `:9002`.

## Architecture

```
cmd/packager/    Entry point, config, server wiring
grpc/            gRPC service implementation
chart/           Values generation for lucity-app chart
gitops/          Soft-serve repository management
eject/           Ejection logic (export project for independent operation)
```

## gRPC API

Defined in `proto/packager/v1/packager.proto`:

- `InitProject` — create GitOps repo for a new project
- `AddService` / `RemoveService` — manage service definitions in base values
- `CreateEnvironment` / `DeleteEnvironment` — manage environment configs
- `Promote` — copy image tags between environments
- `Eject` — export complete project for independent operation

## Configuration

- `PORT` — gRPC listen port (default: 9002)
- `LOG_LEVEL` — Log level (default: info)
- `SOFTSERVE_ADDR` — Soft-serve SSH address
- `SOFTSERVE_KEY` — SSH key for Soft-serve access
