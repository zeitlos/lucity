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

## Registry

User workload images are stored in Zot (self-hosted OCI registry) at `lucity-infra-zot.lucity-system.svc.cluster.local:5000`.

- **Pods** access the registry via cluster DNS (works because CoreDNS resolves `*.svc.cluster.local`)
- **Kubelet** (containerd) accesses the registry via fixed ClusterIP `10.96.100.100:5000` (kubelet uses node DNS, not CoreDNS)
- **Insecure (HTTP)**: containerd is configured with `/etc/containerd/certs.d/10.96.100.100:5000/hosts.toml` on each node
- **Gateway config**: `REGISTRY_URL` (for pod-to-registry) uses DNS; `REGISTRY_IMAGE_PREFIX` (for image refs in Helm values) uses ClusterIP

### Containerd Insecure Registry Config

The worker nodes run Flatcar Linux with a read-only `/usr/share/containerd/config.toml`. To enable insecure HTTP registry access:

1. `/etc/containerd/config.toml` — copy of base config + registry `config_path`
2. `/etc/containerd/certs.d/10.96.100.100:5000/hosts.toml` — insecure HTTP config
3. `/etc/systemd/system/containerd.service.d/override.conf` — points `CONTAINERD_CONFIG` to `/etc/containerd/config.toml`

This config is not persisted by flex.plane — if nodes are replaced, it must be reapplied. Future: automate via DaemonSet or flex.plane MachineDeployment bootstrap config.

## Images

All service images published to `ghcr.io/zeitlos/lucity/<service>` tagged with `main` (from CI) and commit-based semver.
