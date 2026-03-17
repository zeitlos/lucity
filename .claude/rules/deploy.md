# Deployment

## Image Tags

All service images are tagged with `git describe` output (e.g. `0.0.0-304-gdc54d87`). The Helm chart's `appVersion` is set to this tag during CI — each service template falls back to `.Chart.AppVersion` when no per-service `image.tag` is set.

Never use `--set global.image.tag` or per-service `--set gateway.image.tag` — the chart version pins the image tag via `appVersion`.

## Production Deploys

Always use the published OCI chart from GHCR, never the local `./charts/lucity` directory:

```sh
make deploy-prod VERSION=<tag>
```

This runs:

```sh
helm upgrade --install lucity \
  oci://ghcr.io/zeitlos/lucity/charts/lucity \
  --version <tag> \
  --kube-context lucity-prod \
  -n lucity-system --create-namespace \
  -f deployments/lucity-prod/values.yaml \
  -f deployments/lucity-prod/secrets.yaml
```

Do **not** use `--reuse-values`. The Makefile target applies the full values files on every deploy, which is the correct approach — it prevents stale values from accumulating.

## Infrastructure Deploys

```sh
make deploy-prod-infra VERSION=<tag>
```

Same pattern — OCI chart from GHCR with explicit values file. Uses `infra-secrets.yaml` (separate from platform secrets).

## Secrets

Production secrets are split into two files (both gitignored):

- **`deployments/lucity-prod/secrets.yaml`** — platform secrets (GitHub App, ArgoCD/Soft-serve tokens, Stripe, Logto M2M, SSH keys)
- **`deployments/lucity-prod/infra-secrets.yaml`** — infrastructure secrets (Zot htpasswd, Soft-serve admin key, Rybbit, Logto)

Copy from the corresponding `.example` files for first deploy.

## Verifying a Deploy

After deploying, verify pods are running:

```sh
kubectl get pods -n lucity-system -l app.kubernetes.io/instance=lucity --context lucity-prod
```

Check image tags on running pods:

```sh
kubectl get pods -n lucity-system -l app.kubernetes.io/instance=lucity --context lucity-prod \
  -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.spec.containers[*].image}{"\n"}{end}'
```
