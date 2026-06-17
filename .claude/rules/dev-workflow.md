# Development Workflow

## Services

The user starts services with `make dev` (all services with hot reload via air) or individual `make dev-<service>` targets. Claude never starts or stops services — ask the user if services need to be running.

| Service | Port | Protocol | Log file |
|---------|------|----------|----------|
| Conductor | 8080 (HTTP), 9090 (gRPC), 9004 (webhook) | HTTP+gRPC | `tmp/logs/conductor.log` |
| Cashier | 9005 (gRPC), 9006 (HTTP) | gRPC + HTTP | `tmp/logs/cashier.log` |
| Dashboard | 5173 | HTTP | `tmp/logs/dashboard.log` |

The conductor is the unified control-plane binary; it serves the GraphQL API, applies Helm releases, orchestrates builds, reconciles custom domains, and receives GitHub webhooks. Its internal packages (deployer / planner / buildjob) are plain Go packages behind interfaces — no internal network hops. Cashier is the one separate service, reached over gRPC.

## Verifying Changes

After editing code, read the relevant log file(s) in `tmp/logs/` to check for errors. Air auto-rebuilds Go services on file changes, so new logs will appear after a few seconds. If logs show errors or services aren't running, ask the user to start them with `make dev`.

## Integration Tests

The integration test suite (`tests/`) is currently in disrepair after the conductor merge. Verification leans on:

- `go build ./...` (workspace-wide compile check)
- `go vet ./...`
- Manual smoke tests via the GraphQL playground at http://localhost:8080/playground
- Manual end-to-end clicks through the dashboard at http://localhost:5173/
- `tmp/logs/conductor.log` inspection while running `make dev`

When the suite is restored, the targets will be `make test-integration`, `make test-integration-short`, `make test-watch`. Output goes to `tmp/logs/tests.log`.

## Paths

Always use absolute literal paths. Never use `$HOME` or `~` in commands.

## Environment Files

`.env` files are gitignored. Each service with required env vars has a `.env.example`. Air loads `.env` automatically via `env_files = [".env"]`.
