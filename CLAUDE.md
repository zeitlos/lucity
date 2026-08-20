# Lucity

Open-source PaaS on Kubernetes with full ejectability. Monorepo with a single Go control-plane binary plus a Vue 3 dashboard.

## Project

- **Go workspace**: `go.work` with multi-module layout (Go 1.26)
- **Module path**: `github.com/zeitlos/lucity`
- **Monorepo**: `services/conductor` (control plane), `services/cashier` (billing), `services/dashboard` (Vue), `pkg/` (shared Go), `charts/` (Helm)
- **Platform images**: `ghcr.io/zeitlos/lucity/{conductor,cashier,dashboard,docs}`
- **User workload images**: Zot (self-hosted OCI registry, `localhost:5000` in dev)
- **Coding rules**: see `.claude/rules/` for architecture, Go, frontend, GraphQL, marketing, and working conventions

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

- **Kubernetes**: namespaces, labels, Helm release state (Secrets), operator CRDs
- **OCI Registry (Zot)**: built images, tags, digests
- **Identity Provider (OIDC/Logto)**: users, roles, authentication, workspace metadata

The platform is non-intrusive — its downtime does not affect running workloads.

### Single-Repository Model

- **User's source repo** (GitHub): read-only to the platform, never written to
- **No GitOps repo**: the conductor deploys workloads as standard Helm releases applied imperatively (Helm SDK). Deployment config is the chart values it computes; Helm stores the release state in-cluster.

### Multi-Tenant (Workspaces)

Each Lucity instance supports multiple workspaces. A workspace is the tenant boundary — it owns projects, environments, and members. Workspace context is derived from the JWT token (workspace claim). Kubernetes namespaces carry `lucity.dev/workspace` labels for discovery.

### Services

| Service | Port | Protocol | Purpose |
|---------|------|----------|---------|
| Conductor | 8080 (HTTP), 9090 (gRPC), 9004 (webhook HTTP) | HTTP+gRPC | Unified control plane: GraphQL, Helm release management, build orchestration, custom-domain reconciliation, GitHub webhook receiver |
| Cashier | 9005 (gRPC), 9006 (HTTP) | gRPC + HTTP | Stripe billing, metering, suspension callbacks |
| Dashboard | 5173 | HTTP | Vue 3 SPA |

### Conductor internal layout

```
services/conductor/
├── cmd/conductor/                  # main, config, OIDC login, GraphQL + gRPC + webhook servers
├── internal/
│   ├── api/
│   │   ├── graphql/                # gqlgen schema, resolvers, directives, models
│   │   └── webhook/                # GitHub webhook receiver (github/, http/)
│   ├── conductor/                  # Client facade tying the domain packages together
│   ├── deployer/                   # workload deployment
│   │   ├── helm/                   # imperative Helm release apply (services, dbs, volumes, envs)
│   │   └── values/                 # lucity-app values generation + validation
│   ├── buildjob/                   # build Job orchestration (kubernetes/)
│   ├── planner/                    # source detection + build planning (railpack/)
│   ├── source/                     # user source repo access (github/)
│   ├── environment/                # namespace lifecycle (kubernetes/)
│   ├── platform/                   # vendor-neutral value types + IDs (kubernetes/)
│   ├── resources/                  # resource allocation listing (for cashier)
│   ├── hostname/                   # custom-domain DNS verification
│   ├── directory/                  # user/workspace directory (logto/)
│   ├── dbquery/                    # database explorer query execution
│   └── transport/grpc/             # external gRPC (cashier callbacks)
```

### Communication

- **Dashboard ↔ Conductor**: GraphQL over HTTP (port 8080)
- **Cashier ↔ Conductor**: gRPC (port 9090), authenticated with internal ES256 JWTs — SuspendWorkspace, ListResourceAllocations
- **GitHub → Conductor**: HTTP webhooks (port 9004)
- **Long-running operations**: polling (watch registry for images, poll Helm/Kubernetes for rollout status)

### Shared Packages (`pkg/`)

graceful (server lifecycle), logger (slog + tint), auth (OIDC/JWT), labels (K8s label constants), tenant (workspace context), github (App + OAuth), logto (Logto Management API), to (pointer/conversion helpers), conductor + cashier (proto definitions for cross-service gRPC).

## Feature Development Workflow

1. **Research** — how do Railway, Heroku, Render, Fly.io handle this?
2. **Architecture fit** — does it respect stateless design? Is it ejectable?
3. **Day-2 operations cost** — can a small team run it?
4. **Design APIs** — GraphQL schema + (rare) gRPC proto definitions
5. **Design deployment values** — how does this affect the lucity-app chart values?
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
