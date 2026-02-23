# Lucity

Open-source PaaS on Kubernetes with full ejectability. Monorepo with Go backend services and Vue 3 dashboard.

## Project

- **Go workspace**: `go.work` with multi-module layout (Go 1.26)
- **Module path**: `github.com/zeitlos/lucity`
- **Monorepo**: `services/` (6 services), `pkg/` (4 shared packages), `charts/` (Helm), `proto/` (protobuf)
- **Platform images**: `ghcr.io/zeitlos/lucity` (GHCR for Lucity's own service images)
- **User workload images**: Zot (self-hosted OCI registry, `localhost:5000` in dev)
- **Coding rules**: see `.claude/rules/` for Go, Vue, GraphQL, GitOps, general, architecture, and marketing conventions

## Build & Run

- **Go services**: `go run ./cmd/<service>/...` from each service directory, or `make dev-<service>` from root
- **Dashboard**: `npm run dev` from `services/dashboard/`
- **Build all**: `make build`
- **GraphQL codegen**: `go generate ./graphql/resolver.go` from gateway dir
- **Proto codegen**: `make proto` (requires buf)

## Architecture

### Stateless Design

The platform has no central database. All state is derived from external systems:

- **Git (Soft-serve)**: GitOps repos, Helm values, environment config
- **Kubernetes**: namespaces, labels, ArgoCD Applications, operator CRDs
- **OCI Registry (Zot)**: built images, tags, digests
- **Identity Provider (OIDC)**: users, roles, authentication

The platform is non-intrusive — its downtime does not affect running workloads.

### Two-Repository Model

- **User's source repo** (GitHub): read-only to the platform, never written to
- **Platform's GitOps repo** (Soft-serve): managed entirely by the platform, contains Helm values per environment

### Single-Tenant

Each Lucity instance serves one organization. No organization header or multi-tenant scoping — the instance IS the organization.

### Services

| Service | Port | Protocol | Purpose |
|---------|------|----------|---------|
| Gateway | 8080 | HTTP/GraphQL | API entry point, delegates to backend services |
| Builder | 9001 | gRPC | Source-to-image via railpack, OCI push to registry |
| Packager | 9002 | gRPC | Helm values generation, Soft-serve repo management, ejection |
| Deployer | 9003 | gRPC | ArgoCD Application lifecycle, sync status, promotion |
| Webhook | 9004 | HTTP | GitHub webhook reception, event routing |
| Dashboard | 5173 | HTTP | Vue 3 SPA for project/environment management |

### Communication

- **Dashboard ↔ Gateway**: GraphQL over HTTP
- **Gateway ↔ Backend services**: gRPC (short-lived commands)
- **Long-running operations**: polling (watch registry for images, poll ArgoCD for sync status)
- **External triggers**: GitHub webhooks → Webhook service

### Shared Packages (`pkg/`)

graceful (server lifecycle), logger (slog + tint), auth (OIDC/JWT), labels (K8s label constants)

## Feature Development Workflow

1. **Research** — how do Railway, Heroku, Render, Fly.io handle this?
2. **Architecture fit** — does it respect stateless design? Is it ejectable?
3. **Day-2 operations cost** — can a small team run it?
4. **Design APIs** — GraphQL schema + gRPC proto definitions
5. **Design GitOps structure** — how does this affect the lucity-app chart values?
6. **Design frontend** — Vue pages, composables, GraphQL queries
7. **Implement minimal** — ship the smallest useful version first
8. **Test** — GraphQL playground, dashboard end-to-end, Go unit tests
9. **Iterate** — extend with more advanced capabilities

## Smoke Testing

### Gateway

```sh
cd services/gateway && go run ./cmd/gateway/...
```

Verify GraphQL playground at http://localhost:8080/playground.

### Dashboard

```sh
cd services/dashboard && npm run dev
```

Requires gateway on :8080. Verify at http://localhost:5173/.

### Other Services

```sh
make dev-builder    # gRPC on :9001
make dev-packager   # gRPC on :9002
make dev-deployer   # gRPC on :9003
make dev-webhook    # HTTP on :9004
```

### Integration Tests

```sh
make test-integration          # full suite (all services + Minikube)
make test-integration-short    # quick tests (gateway only)
make test-watch                # auto-rerun on file changes (requires watchexec)
```

Test output is logged to `tmp/logs/tests.log`. After running tests, read this file to check results. Look for `--- FAIL` to find failures and `ok`/`FAIL` for overall status. See `tests/CLAUDE.md` for full details.

## Known Issues

_None yet — this is a fresh scaffold._
