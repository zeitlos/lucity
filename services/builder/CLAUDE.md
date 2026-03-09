# Builder

Source-to-image service. Detects services from source code using railpack, builds container images, and pushes them to the OCI registry (Zot).

## Run

```sh
go run ./cmd/builder/...
```

gRPC on `:9001`.

## Architecture

```
cmd/builder/     Entry point, config, server wiring
grpc/            gRPC service implementation
engine/          Build engine interface and implementations
  engine.go      Engine interface (Detect, Build)
  local.go       Local engine: railpack CLI + docker push
build/           In-memory build state tracker
```

## gRPC API

Defined in `pkg/builder/builder.proto`:

- `DetectServices` — scan a source repo and return detected services with framework info
- `StartBuild` — start an async container image build, returns a build ID
- `BuildStatus` — poll the status of an in-progress or completed build

## Configuration

- `PORT` — gRPC listen port (default: 9001)
- `LOG_LEVEL` — Log level (default: info)
- `REGISTRY_URL` — OCI registry URL (default: localhost:5000)
- `WORK_DIR` — Temp directory for cloning and building (default: /tmp/lucity-builds)
