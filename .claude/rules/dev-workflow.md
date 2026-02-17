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
| Dashboard | 5173 | HTTP | `tmp/logs/dashboard.log` |

## Verifying Changes

After editing code, read the relevant log file(s) in `tmp/logs/` to check for errors. Air auto-rebuilds Go services on file changes, so new logs will appear after a few seconds. If logs show errors or services aren't running, ask the user to start them with `make dev`.

## Integration Tests

```sh
make test-integration          # full suite (needs all services + Docker)
make test-integration-short    # quick tests only (gateway sufficient)
```

Tests are in `tests/` — a separate Go module that hits the GraphQL API with generated JWT tokens.

## Paths

Always use absolute literal paths. Never use `$HOME` or `~` in commands.

## Environment Files

`.env` files are gitignored. Each service with required env vars has a `.env.example`. Air loads `.env` automatically via `env_files = [".env"]`.
