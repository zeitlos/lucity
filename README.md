# Lucity

Open-source Platform-as-a-Service on Kubernetes. Railway/Heroku-like developer experience with full ejectability.

## What is Lucity?

Lucity gives you the simplicity of a managed PaaS with the freedom of owning your infrastructure. Connect a GitHub repo, and Lucity builds, packages, and deploys your app to Kubernetes. When you're ready to move on, eject and take your standard Helm charts and ArgoCD configs with you.

## Key Principles

- **Ejectability** — leave anytime with standard K8s/Helm/ArgoCD configs
- **Stateless** — no platform database; state lives in Git, K8s, OCI Registry, and your IDP
- **Your repo is sacred** — the platform never writes to your source repository
- **Discovery over definition** — standard K8s labels, no custom CRDs
- **Standard tools** — Helm, ArgoCD, Gateway API, CloudNativePG, all open source

## Architecture

```
GitHub Repo ──webhook──► Lucity ──GitOps──► ArgoCD ──sync──► Kubernetes
                           │
                    ┌──────┼──────┐
                    │      │      │
                 Builder Packager Deployer
                    │      │      │
                 railpack Helm   ArgoCD
                    │    values    │
                    ▼      ▼      ▼
                 OCI Reg  Git   K8s Cluster
                 (Zot)  (Soft-serve)
```

## Getting Started

```sh
# Build all services
make build

# Run the gateway (GraphQL API)
make dev-gateway

# Run the dashboard
make dev-dashboard
```

## License

[AGPL-3.0](LICENSE)
