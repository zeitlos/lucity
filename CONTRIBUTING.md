# Contributing to Lucity

Lucity is still in early development. We're not accepting external contributions yet - the architecture is moving fast and we'd rather not waste your time with PRs that conflict with in-flight changes.

That said, we'd love to hear from you:

- **Bug reports:** open an issue using the bug report template
- **Feature requests:** open an issue using the feature request template
- **Questions:** use GitHub Discussions

The rest of this guide covers the local development setup if you want to explore the codebase.

## Architecture in one breath

Lucity is a single Go control-plane binary (**conductor**) plus a Vue 3 **dashboard** and a separate **cashier** billing service. The conductor serves the GraphQL API, builds images, applies workloads as standard Helm releases, and receives GitHub webhooks. There is no GitOps repo and no ArgoCD in the deploy path; deployment state lives in the cluster as Helm releases. See [Architecture](https://lucity.cloud/architecture/how-it-works) for the full picture.

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

Installs the Gateway API CRDs and Envoy Gateway, then deploys the `lucity-infra` chart into the `lucity-system` namespace: the OCI registry (Zot), the PostgreSQL operator (CloudNativePG), the observability stack, and a Gateway resource.

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
| Envoy Gateway | `:8880` |
| Logto (admin) | `:3002` |
| VictoriaMetrics | `:8428` |
| Grafana | `:3000` |

Deployed services with a configured hostname are accessible at `http://<name>.lucity.local:8880` via Envoy Gateway and Gateway API HTTPRoutes.

### 5. Generate internal JWT keys

```sh
make generate-internal-keys
```

Writes the ES256 keypair used to authenticate the conductor ↔ cashier gRPC boundary. Reference the generated paths from each service's `.env`.

### 6. Configure services

The conductor and cashier each ship a `.env.example`. Copy and fill in the values:

```sh
cp services/conductor/.env.example services/conductor/.env
cp services/cashier/.env.example services/cashier/.env
```

Key configuration:

| Service | Required Variables |
|---------|-------------------|
| Conductor | GitHub App (`GITHUB_APP_ID`, `GITHUB_CLIENT_ID`, `GITHUB_CLIENT_SECRET`, private key), OIDC/Logto (`OIDC_*`, `LOGTO_*`), registry URLs, `WORKLOAD_DOMAIN` |
| Cashier | `STRIPE_*`, `LOGTO_*` (only needed when working on billing) |

Point the conductor at Zot's fixed ClusterIP (assigned in `deployments/minikube/values.yaml`, e.g. `10.96.100.50:5000`) for the image refs it writes into Helm values, so kubelet can pull them.

> **Why a ClusterIP, not a DNS name?** Docker on minikube uses the host DNS resolver, not CoreDNS. Cluster-internal DNS names like `*.svc.cluster.local` don't resolve for image pulls. The fixed ClusterIP works because `--insecure-registry` already covers the service CIDR.

Billing is optional in dev: leave `CASHIER_ADDR` unset on the conductor to run without cashier.

### 7. Start all services

```sh
make dev
```

Conductor, cashier, and dashboard start with hot reload (air for the Go services, Vite for the dashboard). Dashboard at http://localhost:5173, GraphQL playground at http://localhost:8080/playground.

## Quick Reference

```sh
make minikube                # 1. Create cluster (one-time)
make infra                   # 2. CRDs + Envoy Gateway + lucity-infra chart
make dns                     # 3. Wildcard DNS for *.lucity.local (one-time)
make infra-forward           # 4. Port-forward infrastructure
make generate-internal-keys  # 5. ES256 keypair for conductor <-> cashier
make dev                     # 6. Start services with hot reload
```

## Services

| Service | Port | Protocol | Purpose |
|---------|------|----------|---------|
| Conductor | 8080 (HTTP), 9090 (gRPC), 9004 (webhook) | HTTP + gRPC | Unified control plane: GraphQL API, Helm release management, build orchestration, GitHub webhooks |
| Cashier | 9005 (gRPC), 9006 (HTTP) | gRPC + HTTP | Stripe billing, metering, suspension callbacks |
| Dashboard | 5173 | HTTP | Vue 3 SPA for project and environment management |

`builder` is a separate image (`ghcr.io/zeitlos/lucity/builder`) that the conductor runs inside a per-build Kubernetes Job, not a long-running service.

## Makefile Targets

| Target | Description |
|--------|-------------|
| `make minikube` | Create minikube cluster with insecure registry config |
| `make infra` | Install CRDs + Envoy Gateway + deploy the `lucity-infra` chart |
| `make dns` | Set up wildcard DNS for `*.lucity.local` (one-time) |
| `make infra-forward` | Port-forward infrastructure to localhost |
| `make generate-internal-keys` | Generate the ES256 keypair for internal gRPC auth |
| `make dev` | Start all services with hot reload |
| `make dev-<service>` | Start one service (e.g. `make dev-conductor`) |
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
