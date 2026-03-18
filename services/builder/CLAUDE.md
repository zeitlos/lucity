# Builder

Source-to-image service. Detects services from source code using railpack, builds container images, and pushes them to the OCI registry (Zot).

## Run

```sh
go run ./cmd/builder/...
```

gRPC on `:9001`. Requires a Kubernetes cluster (creates build Jobs + BuildKit).

## Architecture

```
cmd/builder/     Entry point, config, server wiring, K8s Job build runner
grpc/            gRPC service implementation
engine/          Build engine interface and K8s implementation
  engine.go      Engine interface (Detect, Build)
  kubernetes.go  K8s Jobs + BuildKit engine
  detect.go      Railpack-based source detection
build/           Build state tracking (K8s Job annotations)
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
- `BUILD_IMAGE` — Container image for build Job pods (required)
- `BUILD_NAMESPACE` — K8s namespace for BuildKit and build Jobs (default: lucity-builds)
- `BUILDKIT_ADDR` — TCP address of BuildKit service
- `WORK_DIR` — Temp directory for cloning and building (default: /tmp/lucity-builds)
