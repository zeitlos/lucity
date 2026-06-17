# Deployment Conventions

## Single-Repository Model

The platform reads from one repository and writes to none:

1. **Source repo** (user-owned, e.g. `github.com/acme/myapp`) — the platform reads from this to build images but NEVER writes to it. Not a commit, not a file, not a config.

There is no platform-managed GitOps repo. Deployment state lives in the cluster: Helm release state (stored as Secrets by Helm) plus the discovery labels on each resource. The conductor applies releases imperatively through the Helm SDK (`helm upgrade --install` semantics). No external Git server, no ArgoCD in the deploy path.

## lucity-app Chart

User workloads are deployed via the `lucity-app` Helm chart. The deployer generates the chart **values** — it never touches templates. The chart handles:

- Deployments (web services, workers)
- Services (ClusterIP)
- HTTPRoutes (Gateway API for public traffic)
- CronJobs (scheduled tasks)
- ConfigMaps (environment config)
- CNPG Clusters (PostgreSQL)
- Redis instances

## Release Model

One Helm release per project per environment. Each release renders the `lucity-app` chart from merged values:

- **Base values**: services, databases, shared config — what exists.
- **Per-environment overrides**: image tag/digest, replica count, resource limits — what differs.

Operations are imperative and idempotent: the deployer computes the new values, then re-applies the release. A change to replicas, an image bump, or a new variable is a fresh release revision. Helm keeps the revision history, so rollback is a Helm rollback.

A promotion from development to staging copies an image digest from one environment's values into another and re-applies. No rebuild, no repackaging.

## Environment Lifecycle

- **Permanent environments** (development, staging, production): created with the project, persist until project deletion
- **Ephemeral environments** (PR previews): auto-created on PR open, auto-deleted on PR merge/close
- **Promotion**: copies image tags between environments, never rebuilds

## Label Conventions

All labels use the `lucity.dev/` prefix:

```yaml
# Namespace labels (discovery)
lucity.dev/project: "myapp"
lucity.dev/environment: "production"
lucity.dev/ephemeral: "true"              # PR environments only

# OCI Image labels (set by the builder)
org.opencontainers.image.source: "https://github.com/acme/myapp"
org.opencontainers.image.revision: "a1b2c3d"
lucity.dev/built-by: "lucity-builder"
lucity.dev/service: "api"
```

## Ejection

When a user ejects, they receive a self-contained tree rendered from their Helm releases:

```
ejected-project/
├── base/                          # Shared Helm values
├── environments/                  # Per-environment overrides
├── chart/                         # Complete lucity-app chart (templates + values)
│   ├── Chart.yaml
│   ├── templates/
│   └── values.yaml
├── argocd/                        # ArgoCD Application manifests (optional GitOps setup)
│   └── applications/
└── README.md                      # Setup guide: prerequisites, commands, how to modify
```

The ejected output is fully self-contained — no Lucity dependencies. Lucity itself deploys via imperative Helm, but the eject bundle ships ready-made ArgoCD Applications so you can run the workloads under your own GitOps if you prefer. Either way: `helm upgrade` by hand, or point your own ArgoCD at the repo.
