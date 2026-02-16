# General Conventions

## API Naming

Concise, precise, consistent. Use domain terminology.

- **Plural for lists**: `projects`, `environments`, `services`
- **Singular + ID for single items**: `project(id: ID!)`, `environment(id: ID!)`
- **Action verbs for mutations**: `CreateProject`, `PromoteEnvironment`, `DeleteService`
- **No obscure abbreviations**: prefer clarity over brevity
- **Consistent across layers**: same names in GraphQL schema, Go resolvers, and gRPC methods

## Module Pattern

Each module is three things:

1. **Go library** — importable package for programmatic use
2. **CLI tool** — cobra-based CLI for local use and debugging
3. **gRPC service** — network-accessible API for inter-service communication

```bash
# CLI usage
lucity-builder build --source ./src/api --registry registry.example.com/myapp/api
# gRPC usage (called by gateway or other services)
# Programmatic usage (imported as Go package)
```

## Dependency Injection

Through constructors, not globals. Optional parameters via functional options (`With*` pattern).

```go
server, err := gateway.New(
    gateway.WithPort(config.Port),
    gateway.WithBuilderClient(builderConn),
)
```

## File Organization

Feature-based (group by domain), not layer-based. A `project.go` file contains types, functions, and methods related to projects.

## Configuration

Environment-driven. Never hardcode secrets. Every service has a `.env.example` documenting required variables.

## Generated Code

Never manually edit files with `.gen.go` suffix or `generated.go`. Regenerate from source (schema, proto).

## Deployment

- Docker images tagged with git commit hash
- Standard Helm charts (no HULL)
- ArgoCD for GitOps delivery
