# Go Conventions

## Philosophy

Write idiomatic Go. Follow [Effective Go](https://go.dev/doc/effective_go) and [The Zen of Go](https://the-zen-of-go.netlify.app/) by Dave Cheney. Key principles:

- Each package fulfils a single purpose
- Handle errors explicitly
- Return early rather than nesting deeply
- Write for clarity, not cleverness
- A little copying is better than a little dependency
- Simplicity is not a goal, it is the prerequisite

## Imports

Three groups separated by blank lines: stdlib, external, internal.

```go
import (
    "context"
    "fmt"
    "log/slog"

    "github.com/kelseyhightower/envconfig"

    "github.com/zeitlos/lucity/pkg/auth"
    "github.com/zeitlos/lucity/pkg/graceful"
)
```

Always use full module paths. No relative imports.

## Naming

- **Exported**: PascalCase (`CreateProject`, `Environment`, `NewServer`)
- **Unexported**: camelCase (`parseImageTag`, `releaseName`)
- **Constructors**: `New<Type>(...) (*Type, error)`
- **Method receivers**: short names — `s *Server`, `c *Client`, `b *Builder`
- **Interfaces**: semantic names (Builder, Deployer, Planner). No forced `-er` suffix.
- **Packages**: single lowercase word (`auth`, `labels`, `deployer`, `hostname`)
- **No `Get`/`List` prefixes**: follow stdlib convention — `Repositories()` not `GetRepositories()` or `ListRepositories()`. Use the noun directly. `Create`, `Delete`, `Update` verbs are fine since they denote actions.
- **Local variables and parameters**: spell out domain words. `deployment` not `dep`, `replicaSet` not `rs`, `environmentID` not `envID`, `currentHash` not `hash` when scope matters. Readability over shortness. Exceptions, only because they're universal: `ctx context.Context`, short method receivers (`c *Client`, `s *Server`), `i`/`j`/`k` loop indices, established acronyms in camelCase (`ID`, `URL`, `JSON`, `HTTP`).

## Error Handling

Return errors directly when there's nothing meaningful to add. Wrap with `fmt.Errorf("...: %w", err)` only when the call site adds context the underlying error doesn't already have. Avoid `"failed to X: ..."` boilerplate that just stutters the operation back.

Use `slog.Error()` before `os.Exit(1)` in main.

### Blank line before err handling

Always separate the call producing an error from the `if err != nil` block with a blank line. Same goes for early-return `if` checks following any statement.

```go
parts := strings.SplitN(s, "/", 5)

if len(parts) != 5 || slices.Contains(parts, "") {
    return DeploymentID{}, fmt.Errorf("invalid deployment id %q", s)
}

return DeploymentID{
    Workspace:   parts[0],
    Project:     parts[1],
    Environment: parts[2],
    Service:     parts[3],
    Hash:        parts[4],
}, nil
```

## Logging

`log/slog` with structured key-value pairs:

```go
slog.Info("connecting to registry", "url", config.RegistryURL)
slog.Error("failed to build image", "error", err, "service", serviceName)
```

Levels: debug, info, warn, error. Colored output via tint handler.

## Configuration

`envconfig` struct tags. Load in `main()`, pass values to constructors.

```go
type Config struct {
    Port     string `envconfig:"PORT" default:"8080"`
    LogLevel string `envconfig:"LOG_LEVEL" default:"info"`
}
```

## Graceful Shutdown

```go
ctx, cancel := graceful.Context()
defer cancel()
graceful.Serve(ctx, graphqlServer, grpcServer)
```

Server interface: `Label() string`, `Start() error`, `Shutdown(context.Context) error`.

## File Organization

- **Naming**: feature-based — `project.go`, `environment.go`, `build.go`
- **Generated code**: `.gen.go` suffix or `generated.go`
- **Structure within file**: package decl, imports, types/interfaces, constructors, methods, helpers
- **Tests**: `_test.go` suffix. Standard `testing` package.

## Dependencies

After adding, removing, or upgrading dependencies, run `go mod tidy` in **every affected service directory**. Docker builds use `GOWORK=off`, so each service needs a complete `go.sum` — the workspace won't save you in CI. When in doubt, tidy all services:

```sh
for svc in services/*/; do (cd "$svc" && go mod tidy); done
```

## Patterns

- **Context**: always first parameter — `ctx context.Context`
- **Options pattern**: `With<Option>(value)` for optional constructor params
- **Dependency injection**: interfaces passed to constructors, not globals
