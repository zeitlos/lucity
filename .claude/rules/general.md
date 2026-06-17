# General Conventions

## API Naming

Concise, precise, consistent. Use domain terminology.

- **Plural for lists**: `projects`, `environments`, `services`
- **Singular + ID for single items**: `project(id: ID!)`, `environment(id: ID!)`
- **Action verbs for mutations**: `CreateProject`, `PromoteEnvironment`, `DeleteService`
- **No obscure abbreviations**: prefer clarity over brevity
- **Consistent across layers**: same names in GraphQL schema, Go resolvers, and gRPC methods

## Service Layout

The platform is a small set of binaries, not a fleet of microservices:

1. **Conductor** — the unified control plane. Its internal packages (deployer, planner, buildjob, ...) are plain Go packages behind narrow interfaces, not standalone services.
2. **Cashier** — a separate billing service, reachable over gRPC (conductor ↔ cashier).
3. **Builder** — a separate binary that runs inside a per-build Kubernetes Job, not a long-running service.

Inside the conductor, prefer plain packages and interfaces. Reach for gRPC only at a real process boundary (cashier).

## Dependency Injection

Through constructors, not globals. Optional parameters via functional options (`With*` pattern).

```go
server, err := conductor.New(
    conductor.WithPort(config.Port),
    conductor.WithDeployer(deployer),
)
```

## File Organization

Feature-based (group by domain), not layer-based. A `project.go` file contains types, functions, and methods related to projects.

## Configuration

Environment-driven. Never hardcode secrets. Every service has a `.env.example` documenting required variables.

## Generated Code

Never manually edit files with `.gen.go` suffix or `generated.go`. Regenerate from source (schema, proto).

## Deployment

- Docker images tagged with `git describe` output
- Standard Helm charts (no HULL)
- Workloads applied as imperative Helm releases by the conductor
