# GitOps Conventions

## Two-Repository Model

Each project has two repositories:

1. **Source repo** (user-owned, e.g. `github.com/acme/myapp`) — the platform reads from this but NEVER writes to it
2. **GitOps repo** (platform-managed, on Soft-serve) — the platform owns and manages this entirely

## lucity-app Chart

User workloads are deployed via the `lucity-app` Helm chart. The packager generates `values.yaml` files — it never touches templates. The chart handles:

- Deployments (web services, workers)
- Services (ClusterIP)
- HTTPRoutes (Gateway API for public traffic)
- CronJobs (scheduled tasks)
- ConfigMaps (environment config)
- CNPG Clusters (PostgreSQL)
- Redis instances

## GitOps Repo Structure

```
gitops/{project}/
├── base/
│   ├── Chart.yaml              # Depends on lucity-app chart
│   └── values.yaml             # Shared: services, databases, base config
└── environments/
    ├── development/
    │   └── values.yaml         # Overrides: replicas, image tag, debug settings
    ├── staging/
    │   └── values.yaml         # Overrides: promoted image tag
    ├── production/
    │   └── values.yaml         # Overrides: replicas, resources, HA config
    └── pr-142/
        └── values.yaml         # Ephemeral: minimal resources, PR-specific tag
```

Each environment's Chart.yaml references `../../base` as a dependency.

## Commit Messages

Semantic commits in GitOps repos:

- `deploy(development): api a1b2c3d` — new image deployed
- `promote(staging): api a1b2c3d from development` — image promoted
- `env(create): pr-142 from development` — environment created
- `env(delete): pr-142` — environment removed
- `config(production): update replicas to 3` — configuration change

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

# OCI Image labels (set by Builder)
org.opencontainers.image.source: "https://github.com/acme/myapp"
org.opencontainers.image.revision: "a1b2c3d"
lucity.dev/built-by: "lucity-builder"
lucity.dev/service: "api"
```

## Ejection

When a user ejects, they receive:

```
ejected-project/
├── base/                          # Shared Helm values
├── environments/                  # Per-environment overrides
├── chart/                         # Complete lucity-app chart (templates + values)
│   ├── Chart.yaml
│   ├── templates/
│   └── values.yaml
├── argocd/                        # ArgoCD Application manifests
│   └── applications/
└── README.md                      # Setup guide: prerequisites, commands, how to modify
```

The ejected output is fully self-contained — no Lucity dependencies. Users point their own ArgoCD at this repo and run independently.
