# Contributing to Lucity

Lucity is still in early development. We're not accepting external contributions yet - the architecture is moving fast and we'd rather not waste your time with PRs that conflict with in-flight changes.

That said, we'd love to hear from you:

- **Bug reports:** open an issue using the bug report template
- **Feature requests:** open an issue using the feature request template
- **Questions:** use GitHub Discussions

The rest of this guide covers the local development setup if you want to explore the codebase.

## Prerequisites

- [Go 1.26+](https://go.dev/dl/)
- [Node.js 20+](https://nodejs.org/)
- [Docker](https://docs.docker.com/get-docker/)
- [Minikube](https://minikube.sigs.k8s.io/docs/start/)
- [Helm](https://helm.sh/docs/intro/install/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [crane](https://github.com/google/go-containerregistry/tree/main/cmd/crane) (image push)
- [air](https://github.com/air-verse/air) (hot reload)
- A [GitHub App](https://docs.github.com/en/apps/creating-github-apps) configured for OAuth

## Getting Started

### 1. Create the cluster

```sh
make minikube
```

Starts minikube with `--insecure-registry "10.96.0.0/12"` so Docker on the node trusts any ClusterIP-based registry over HTTP. This covers the entire Kubernetes service CIDR. See [minikube registry docs](https://minikube.sigs.k8s.io/docs/handbook/registry/#enabling-insecure-registries).

### 2. Deploy infrastructure

```sh
make infra
```

Installs Gateway API CRDs, Envoy Gateway, and deploys Zot (OCI registry), Soft-serve (Git server), ArgoCD, and a Gateway resource via Helm into the `lucity-system` namespace.

### 3. Set up local DNS

```sh
make dns
```

Configures [dnsmasq](https://thekelleys.org.uk/dnsmasq/doc.html) so `*.lucity.local` resolves to `127.0.0.1`. Run once; survives reboots. Requires Homebrew (installs dnsmasq if not present).

This lets you access deployed services by hostname, e.g. `http://myapp.lucity.local:8880`.

### 4. Port-forward infrastructure

```sh
make infra-forward
```

Exposes infrastructure on localhost:

| Service | Local Port |
|---------|-----------|
| Zot (OCI registry) | `:5000` |
| Soft-serve (SSH) | `:23231` |
| Soft-serve (HTTP) | `:23232` |
| ArgoCD | `:8443` |
| Envoy Gateway | `:8880` |

Deployed services with a configured hostname are accessible at `http://<name>.lucity.local:8880` via Envoy Gateway and Gateway API HTTPRoutes.

### 5. Generate API tokens

```sh
make infra-tokens
```

> Requires `make infra-forward` to be running.

Prints an ArgoCD token and a Soft-serve token. Add them to the service `.env` files:

| Token | Goes into |
|-------|-----------|
| `ARGOCD_TOKEN` | `services/deployer/.env` |
| `SOFTSERVE_TOKEN` | `services/deployer/.env`, `services/packager/.env` |

### 6. Configure services

Each service has a `.env.example`. Copy and fill in the values:

```sh
cp services/gateway/.env.example services/gateway/.env
cp services/builder/.env.example services/builder/.env
cp services/packager/.env.example services/packager/.env
cp services/deployer/.env.example services/deployer/.env
```

Key configuration:

| Service | Required Variables |
|---------|-------------------|
| Gateway | `GITHUB_APP_ID`, `GITHUB_CLIENT_ID`, `GITHUB_CLIENT_SECRET`, `REGISTRY_IMAGE_PREFIX` |
| Builder | `REGISTRY_INSECURE=true` (default for local Zot) |
| Packager | `SOFTSERVE_SSH_KEY_PATH`, `SOFTSERVE_TOKEN` |
| Deployer | `ARGOCD_TOKEN`, `SOFTSERVE_TOKEN` |

Set `REGISTRY_IMAGE_PREFIX` to Zot's fixed ClusterIP (assigned in `deployments/minikube/values.yaml`):

```
REGISTRY_IMAGE_PREFIX=10.96.100.50:5000
```

> **Why a ClusterIP, not a DNS name?** Docker on minikube uses the host DNS resolver, not CoreDNS. Cluster-internal DNS names like `*.svc.cluster.local` don't resolve for image pulls. The fixed ClusterIP works because `--insecure-registry` already covers the service CIDR.

### 7. Start all services

```sh
make dev
```

Dashboard at http://localhost:5173, GraphQL playground at http://localhost:8080/playground.

## Quick Reference

```sh
make minikube        # 1. Create cluster (one-time)
make infra           # 2. CRDs + Envoy Gateway + Helm deploy
make dns             # 3. Wildcard DNS for *.lucity.local (one-time)
make infra-forward   # 4. Port-forward services
make infra-tokens    # 5. Generate tokens → paste into .env files
make dev             # 6. Start services with hot reload
```

## Services

| Service | Port | Protocol | Purpose |
|---------|------|----------|---------|
| Gateway | 8080 | HTTP/GraphQL | API entry point, delegates to backend services |
| Builder | 9001 | gRPC | Source-to-image builds via railpack, pushes to Zot |
| Packager | 9002 | gRPC | GitOps repo management, Helm values generation |
| Deployer | 9003 | gRPC | ArgoCD Application lifecycle, sync, promotion |
| Webhook | 9004 | HTTP | GitHub webhook reception and event routing |
| Dashboard | 5173 | HTTP | Vue 3 SPA for project and environment management |

## Makefile Targets

| Target | Description |
|--------|-------------|
| `make minikube` | Create minikube cluster with insecure registry config |
| `make infra` | Install CRDs + Envoy Gateway + deploy Zot, Soft-serve, ArgoCD |
| `make dns` | Set up wildcard DNS for `*.lucity.local` (one-time) |
| `make infra-forward` | Port-forward infrastructure to localhost |
| `make infra-tokens` | Generate ArgoCD + Soft-serve API tokens |
| `make dev` | Start all services with hot reload |
| `make dev-<service>` | Start one service (e.g. `make dev-gateway`) |
| `make dev-logs` | Tail all service logs |
| `make dev-stop` | Stop all services |
| `make build` | Build all Go services |
| `make proto` | Regenerate protobuf code |
| `make generate-graphql` | Regenerate GraphQL resolvers |
| `make lint` | Run dashboard linter |
| `make test-integration` | Run integration tests (requires `make dev`) |
| `make infra-down` | Uninstall infrastructure from cluster |

## Further Reading

- [Architecture](https://lucity.cloud/architecture/how-it-works): how the pieces fit together
- [Concepts](https://lucity.cloud/getting-started/concepts): projects, services, environments
- [Self-Hosting](https://lucity.cloud/getting-started/self-hosting): hosting Lucity on your own hardware
