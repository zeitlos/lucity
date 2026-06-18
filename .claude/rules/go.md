# Go

Write idiomatic Go; the standard idioms are assumed. Only the project-specific rules below.

- **Workspace module resolution**: services must NOT `require` internal `github.com/zeitlos/lucity/*` modules in their `go.mod`. The workspace resolves them; a require directive makes Go try to fetch from GitHub and fail.
- **Tidy per service**: Docker builds use `GOWORK=off`, so every service needs a complete `go.sum`. After changing dependencies, run `go mod tidy` in each affected service directory, not just the workspace root.
- **Never edit generated code** (`*.gen.go`, `generated.go`). Regenerate from the source schema or proto.
- **No `Get`/`List` prefixes** on accessors — use the noun (`Repositories()`, not `ListRepositories()`). `Create`/`Update`/`Delete` verbs are fine.
- **Dependency injection through constructors**, never globals. Optional params via functional `With*` options.
- **Spell out domain names** in locals and params (`environmentID`, not `envID`). Exceptions: `ctx`, short receivers, loop indices, established acronyms.
- **gRPC only at real process boundaries.** Inside the control plane, prefer plain packages behind narrow interfaces over network hops.
