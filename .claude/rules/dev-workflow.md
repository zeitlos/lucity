# Development Workflow

## Starting Services

Always use the Makefile targets. Never manually pass env vars on the command line or use `go run` directly.

```sh
make dev-gateway    # HTTP on :8080
make dev-builder    # gRPC on :9001
make dev-packager   # gRPC on :9002
make dev-deployer   # gRPC on :9003
make dev-webhook    # HTTP on :9004
make dev-dashboard  # HTTP on :5173
```

The Makefile targets `cd` into the service directory, export variables from the local `.env` file via `set -a`, then run the service. This is where secrets like `JWT_SECRET` and `GITHUB_CLIENT_ID` live.

Never do this:

```sh
# BAD — bypasses .env, error-prone, hard to reproduce
GITHUB_APP_ID=123 GITHUB_CLIENT_ID=abc JWT_SECRET=foo go run ./services/gateway/cmd/gateway/...
```

## Port Cleanup Before Starting

Before starting a service, kill any existing process on its port. Stale processes from previous runs cause "address already in use" errors and phantom debugging.

```sh
lsof -ti :8080 | xargs kill 2>/dev/null   # before gateway
lsof -ti :9001 | xargs kill 2>/dev/null   # before builder
lsof -ti :9002 | xargs kill 2>/dev/null   # before packager
lsof -ti :9003 | xargs kill 2>/dev/null   # before deployer
lsof -ti :9004 | xargs kill 2>/dev/null   # before webhook
lsof -ti :5173 | xargs kill 2>/dev/null   # before dashboard
```

Use `kill` (SIGTERM) — all Go services use the `graceful` package which handles SIGTERM cleanly. Only use `kill -9` if a process is stuck.

## Verifying Services

After starting a service in the background, always verify it is actually running. Never assume it started successfully.

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

Allow 2-3 seconds after starting before checking — Go compilation + startup takes a moment.

## Restarting After Code Changes

When code changes require a restart:

1. Kill the running process on its port
2. Start the service again with `make dev-<service>`
3. Wait for startup
4. Verify it came back up

```sh
lsof -ti :8080 | xargs kill 2>/dev/null
make dev-gateway &
sleep 3
curl -sf http://localhost:8080/health
```

Never leave a stale process running while starting a new one on the same port. The new process will fail silently with "address already in use".

## Session Cleanup

At the end of a development session, clean up all Lucity processes:

```sh
for port in 8080 9001 9002 9003 9004 5173; do
    lsof -ti :$port | xargs kill 2>/dev/null
done
```

## Paths

Always use absolute literal paths. Never use `$HOME` or `~` in commands — they may not expand in all shell contexts (Makefile recipes, subshells, tool invocations).

```sh
# Good
/Users/christian/Code/lucity/services/gateway/.env

# Bad
$HOME/Code/lucity/services/gateway/.env
~/Code/lucity/services/gateway/.env
```

## Environment Files

Each service that requires env vars has a `.env.example` in its directory. To set up a new service:

```sh
cp services/gateway/.env.example services/gateway/.env
# Edit .env with actual values
```

`.env` files are gitignored. Never commit them. The `.env.example` files document required variables.

**Services with required env vars (no defaults):**

- **Gateway**: `GITHUB_APP_ID`, `GITHUB_CLIENT_ID`, `GITHUB_CLIENT_SECRET`, `JWT_SECRET`
- **Builder**: `JWT_SECRET`, `REGISTRY_TOKEN`
- **Packager**: `JWT_SECRET`

**Services with only optional/defaulted vars:**

- **Deployer** (port 9003), **Webhook** (port 9004) — work without any `.env` file

## Service Quick Reference

| Service | Port | Protocol | Verify | Required Env |
|---------|------|----------|--------|-------------|
| Gateway | 8080 | HTTP | `curl -sf localhost:8080/health` | GITHUB_APP_ID, GITHUB_CLIENT_ID, GITHUB_CLIENT_SECRET, JWT_SECRET |
| Builder | 9001 | gRPC | `lsof -i :9001 \| grep LISTEN` | JWT_SECRET, REGISTRY_TOKEN |
| Packager | 9002 | gRPC | `lsof -i :9002 \| grep LISTEN` | JWT_SECRET |
| Deployer | 9003 | gRPC | `lsof -i :9003 \| grep LISTEN` | none |
| Webhook | 9004 | HTTP | `curl -sf localhost:9004/health` | none |
| Dashboard | 5173 | HTTP | `curl -sf localhost:5173` | none (needs gateway) |
