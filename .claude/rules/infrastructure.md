# Infrastructure

## Production Cluster

- **Managed by flex.plane** — the virtualization platform built by zeitlos.software (that's us)
- **Hosted control plane** — no CP overhead on worker nodes
- **Source code**: `~/Code/flex.plane` (refer here for CCM, CSI, networking details)
- **Provider**: Hetzner Cloud, Nuremberg (nbg1)
- **Workers**: 3x CX33 (shared CPU, 4 vCPU, 8GB RAM, 80GB disk)
- **kubectl context**: `lucity-prod`
- **Namespace**: `lucity-system` (all platform components)

## Cluster Components

| Component | Purpose |
|-----------|---------|
| Cilium | CNI + Gateway API controller |
| cert-manager | TLS certificates (Let's Encrypt DNS01 via Azure DNS) |
| external-dns | Automatic DNS records (Azure DNS, watches Gateway API HTTPRoutes) |
| hcloud-csi | Hetzner Cloud storage (`hcloud-volumes` StorageClass) |

## Domains

- **`lucity.cloud`** — platform (API, dashboard, docs, webhooks) via path-based routing
- **`lucity.app`** — user workloads via subdomain routing (`{service}-{env}.lucity.app`)
- Both zones managed in Azure DNS (resource group: `lucity-prod`)

## Helm Charts

- **`lucity-infra`** — infrastructure: Zot, Soft-serve, ArgoCD, Gateway, certificates
- **`lucity`** — platform services: gateway, builder, packager, deployer, webhook, dashboard, docs
- **Deployment profiles**: `deployments/lucity-prod/` (infra-values.yaml + values.yaml)

## Images

All service images published to `ghcr.io/zeitlos/lucity/<service>` tagged with `main` (from CI) and commit-based semver.
