# Lucity

Open-source PaaS on Kubernetes with full ejectability. Monorepo with a single Go control-plane binary plus a Vue 3 dashboard.

## Project

- **Go workspace**: `go.work` with multi-module layout (Go 1.26)
- **Module path**: `github.com/zeitlos/lucity`
- **Monorepo**: `services/conductor` (control plane), `services/cashier` (billing), `services/dashboard` (Vue), `pkg/` (shared Go), `charts/` (Helm)
- **Platform images**: `ghcr.io/zeitlos/lucity/{conductor,cashier,dashboard,docs}`
- **User workload images**: Zot (self-hosted OCI registry, `localhost:5000` in dev)
- **Coding rules**: see `.claude/rules/` for Go, Vue, GraphQL, GitOps, general, architecture, and marketing conventions

## Build & Run

- **Conductor**: `go run ./cmd/conductor/...` from `services/conductor/`, or `make dev-conductor`
- **Cashier**: `go run ./cmd/cashier/...` from `services/cashier/`, or `make dev-cashier`
- **Dashboard**: `npm run dev` from `services/dashboard/`
- **Build all**: `make build`
- **GraphQL codegen**: `go generate ./internal/api/graphql/resolver.go` from `services/conductor/`
- **Proto codegen**: `make proto`

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

### Multi-Tenant (Workspaces)

Each Lucity instance supports multiple workspaces. A workspace is the tenant boundary — it owns projects, environments, and members. Workspace context is derived from the JWT token (workspace claim). Kubernetes namespaces carry `lucity.dev/workspace` labels for discovery.

### Services

| Service | Port | Protocol | Purpose |
|---------|------|----------|---------|
| Conductor | 8080 (HTTP), 9090 (gRPC), 9004 (webhook HTTP) | HTTP+gRPC | Unified control plane: GraphQL, GitOps repo management, ArgoCD lifecycle, build orchestration, GitHub webhook receiver |
| Cashier | 9005 (gRPC), 9006 (HTTP) | gRPC + HTTP | Stripe billing, metering, suspension callbacks |
| Dashboard | 5173 | HTTP | Vue 3 SPA |

### Conductor internal layout

```
services/conductor/
├── cmd/conductor/                  # main, OIDC, GraphQL server, sessions, run-build
├── internal/
│   ├── api/
│   │   ├── graphql/                # gqlgen schema + resolvers
│   │   ├── handler/                # business logic, takes WorkspaceID explicitly
│   │   ├── webhook/                # GitHub webhook receiver
│   │   └── deploy/                 # in-memory deploy run tracker
│   ├── deployer/
│   │   ├── backend.go              # Backend interface (vendor-neutral)
│   │   └── argo/                   # GitOps + ArgoCD impl
│   ├── builder/                    # source detection + build Job orchestration
│   ├── inproc/                     # gRPC server impls registered on bufconn
│   │   ├── packager/
│   │   ├── deployer/
│   │   └── builder/
│   ├── kube/                       # namespace lifecycle + label resolution
│   ├── domain/                     # vendor-neutral value types
│   └── transport/grpc/             # external gRPC (cashier callbacks)
```

### Communication

- **Dashboard ↔ Conductor**: GraphQL over HTTP (port 8080)
- **Cashier ↔ Conductor**: gRPC (port 9090) — SuspendWorkspace, ListResourceAllocations
- **GitHub → Conductor**: HTTP webhooks (port 9004)
- **Internal modules**: in-process bufconn-backed gRPC inside the conductor binary
- **Long-running operations**: polling (watch registry for images, poll ArgoCD for sync status)

### Shared Packages (`pkg/`)

graceful (server lifecycle), logger (slog + tint), auth (OIDC/JWT), labels (K8s label constants), tenant (workspace context), github (App + OAuth), logto (Logto Management API), conductor + cashier (proto definitions for cross-service gRPC).

## Feature Development Workflow

1. **Research** — how do Railway, Heroku, Render, Fly.io handle this?
2. **Architecture fit** — does it respect stateless design? Is it ejectable?
3. **Day-2 operations cost** — can a small team run it?
4. **Design APIs** — GraphQL schema + (rare) gRPC proto definitions
5. **Design GitOps structure** — how does this affect the lucity-app chart values?
6. **Design frontend** — Vue pages, composables, GraphQL queries
7. **Implement minimal** — ship the smallest useful version first
8. **Test** — GraphQL playground, dashboard end-to-end, Go unit tests
9. **Iterate** — extend with more advanced capabilities

## Smoke Testing

```sh
make dev          # conductor + cashier + dashboard with hot reload
```

Then:

- GraphQL playground: http://localhost:8080/playground
- Dashboard: http://localhost:5173/
- Webhook receiver: POST to http://localhost:9004/webhooks/github

Logs land in `tmp/logs/{conductor,cashier,dashboard}.log`.

## Known Issues

_None yet — this is a fresh scaffold._
