# Webhook

GitHub webhook receiver. Processes push and pull request events, routing them to the builder, packager, and deployer services.

## Run

```sh
go run ./cmd/webhook/...
```

HTTP on `:9004`.

## Architecture

```
cmd/webhook/     Entry point, config, server wiring
github/          GitHub webhook signature validation and event parsing
```

## Webhook Events

- **Push to default branch** → trigger build → update GitOps repo → deploy
- **PR opened** → create preview environment → build → deploy
- **PR updated (synchronize)** → rebuild → update preview
- **PR closed/merged** → delete preview environment

## Configuration

- `PORT` — HTTP listen port (default: 9004)
- `LOG_LEVEL` — Log level (default: info)
- `WEBHOOK_SECRET` — GitHub webhook secret for signature validation
- `BUILDER_ADDR` — Builder gRPC address
- `PACKAGER_ADDR` — Packager gRPC address
- `DEPLOYER_ADDR` — Deployer gRPC address
