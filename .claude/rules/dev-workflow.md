# Development Workflow

## Starting All Services

Use `make dev` to start everything with hot reload:

```sh
make dev
```

This starts all 5 Go services with `air` (auto-rebuild on file changes) and the dashboard with Vite (HMR). All logs go to `tmp/logs/`.

- Kills stale processes on all ports before starting
- Truncates log files at the start of each session
- Ctrl+C stops everything cleanly

## Viewing Logs

```sh
make dev-logs            # all services combined
make dev-logs-gateway    # single service
make dev-logs-builder    # single service
```

Log files are at `tmp/logs/<service>.log`. Claude should read these files to verify changes are working.

## Stopping Services

```sh
make dev-stop            # kill all service ports
# or Ctrl+C if make dev is running in the foreground
```

## Starting Individual Services (without air)

For running a single service without hot reload:

```sh
make dev-gateway    # HTTP on :8080
make dev-builder    # gRPC on :9001
make dev-packager   # gRPC on :9002
make dev-deployer   # gRPC on :9003
make dev-webhook    # HTTP on :9004
make dev-dashboard  # HTTP on :5173
```

## Integration Tests

Run against a live stack (requires `make dev` running):

```sh
make test-integration          # full suite (needs all services + Docker)
make test-integration-short    # tier 1+2 only (no Docker needed)
```

Tests are in `tests/` — a separate Go module that hits the GraphQL API. Uses JWT tokens generated from `pkg/auth` for authentication.

## Verifying Services

**HTTP services** — curl the health endpoint:

```sh
curl -sf http://localhost:8080/health   # gateway
curl -sf http://localhost:9004/health   # webhook
curl -sf http://localhost:5173          # dashboard
```

**gRPC services** — check the port is listening:

```sh
lsof -i :9001 | grep LISTEN   # builder
lsof -i :9002 | grep LISTEN   # packager
lsof -i :9003 | grep LISTEN   # deployer
```

## Air Configuration

Each Go service has an `.air.toml` in its directory. Air:
- Watches the service directory and its `pkg/` dependencies
- Loads `.env` via `env_files = [".env"]`
- Builds to `tmp/air/<service>/`
- Sends SIGINT for graceful shutdown

## Paths

Always use absolute literal paths. Never use `$HOME` or `~` in commands.

```sh
# Good
/Users/christian/Code/lucity/services/gateway/.env

# Bad
$HOME/Code/lucity/services/gateway/.env
~/Code/lucity/services/gateway/.env
```

## Environment Files

Each service that requires env vars has a `.env.example` in its directory:

```sh
cp services/gateway/.env.example services/gateway/.env
# Edit .env with actual values
```

`.env` files are gitignored. Never commit them.

## Service Quick Reference

| Service | Port | Protocol | Verify | Required Env |
|---------|------|----------|--------|-------------|
| Gateway | 8080 | HTTP | `curl -sf localhost:8080/health` | GITHUB_APP_ID, GITHUB_CLIENT_ID, GITHUB_CLIENT_SECRET, JWT_SECRET |
| Builder | 9001 | gRPC | `lsof -i :9001 \| grep LISTEN` | JWT_SECRET, REGISTRY_TOKEN |
| Packager | 9002 | gRPC | `lsof -i :9002 \| grep LISTEN` | JWT_SECRET |
| Deployer | 9003 | gRPC | `lsof -i :9003 \| grep LISTEN` | none |
| Webhook | 9004 | HTTP | `curl -sf localhost:9004/health` | none |
| Dashboard | 5173 | HTTP | `curl -sf localhost:5173` | none (needs gateway) |
