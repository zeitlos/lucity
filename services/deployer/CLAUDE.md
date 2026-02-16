# Deployer

ArgoCD integration and environment lifecycle management. Creates ArgoCD Applications, polls sync status, and manages environment promotion.

## Run

```sh
go run ./cmd/deployer/...
```

gRPC on `:9003`.

## Architecture

```
cmd/deployer/    Entry point, config, server wiring
grpc/            gRPC service implementation
argocd/          ArgoCD Application CRUD and sync management
environment/     Environment lifecycle (create, delete, promote)
```

## gRPC API

Defined in `proto/deployer/v1/deployer.proto`:

- `CreateArgoApp` / `DeleteArgoApp` — manage ArgoCD Applications
- `GetSyncStatus` — poll environment sync status
- `SyncEnvironment` — trigger ArgoCD sync

## Configuration

- `PORT` — gRPC listen port (default: 9003)
- `LOG_LEVEL` — Log level (default: info)
- `ARGOCD_ADDR` — ArgoCD server address
- `ARGOCD_TOKEN` — ArgoCD API token
