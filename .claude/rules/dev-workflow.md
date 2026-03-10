# Development Workflow

## Services

The user starts services with `make dev` (all services with hot reload via air) or individual `make dev-<service>` targets. Claude never starts or stops services — ask the user if services need to be running.

| Service | Port | Protocol | Log file |
|---------|------|----------|----------|
| Gateway | 8080 | HTTP/GraphQL | `tmp/logs/gateway.log` |
| Builder | 9001 | gRPC | `tmp/logs/builder.log` |
| Packager | 9002 | gRPC | `tmp/logs/packager.log` |
| Deployer | 9003 | gRPC | `tmp/logs/deployer.log` |
| Webhook | 9004 | HTTP | `tmp/logs/webhook.log` |
| Cashier | 9005 | gRPC + 9006 HTTP | `tmp/logs/cashier.log` |
| Dashboard | 5173 | HTTP | `tmp/logs/dashboard.log` |

## Verifying Changes

After editing code, read the relevant log file(s) in `tmp/logs/` to check for errors. Air auto-rebuilds Go services on file changes, so new logs will appear after a few seconds. If logs show errors or services aren't running, ask the user to start them with `make dev`.

## Integration Tests

```sh
make test-integration          # full suite (all services + Minikube infrastructure)
make test-integration-short    # quick tests (gateway only)
make test-watch                # auto-rerun on file changes (requires watchexec)
```

Tests are in `tests/` — a separate Go module that hits the GraphQL API with generated JWT tokens. Tests verify side effects with `kubectl` (namespaces, ArgoCD apps, deployments, CNPG clusters) and `psql`.

| Runner | Log file | Status file |
|--------|----------|-------------|
| Tests  | `tmp/logs/tests.log` | `tmp/dev/tests.status` |

After making changes that affect the GraphQL API or backend services, read `tmp/logs/tests.log` to verify integration tests still pass. If the test runner is in watch mode (`make test-watch`), tests re-run automatically.

See `tests/CLAUDE.md` for test organization, infrastructure requirements, and cleanup instructions.

## Paths

Always use absolute literal paths. Never use `$HOME` or `~` in commands.

## Environment Files

`.env` files are gitignored. Each service with required env vars has a `.env.example`. Air loads `.env` automatically via `env_files = [".env"]`.
