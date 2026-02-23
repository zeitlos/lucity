# Integration Tests

End-to-end tests that exercise the full Lucity platform through its GraphQL API and verify side effects with `kubectl` and `psql`.

## Running

Services must be running (`make dev`) and infrastructure must be up (`make infra && make infra-forward`) before running tests.

### Quick tests (gateway only)

```sh
make test-integration-short
```

### Full suite (all services + Minikube)

```sh
make test-integration
```

### Watch mode (auto-rerun on file changes)

```sh
make test-watch
```

Requires `watchexec` (`brew install watchexec`).

## Logs

Test output is written to `tmp/logs/tests.log`. Test runner status is in `tmp/dev/tests.status`.

```sh
cat tmp/logs/tests.log
```

Look for `--- FAIL` lines to identify failures and `ok`/`FAIL` at the end for overall result.

## Test Organization

All tests run sequentially via `TestIntegration` in `main_test.go`. Each test group is a subtest:

```
TestIntegration/Health       — gateway health + playground
TestIntegration/Auth         — JWT auth, unauth, invalid token
TestIntegration/Project      — create, list, get, not-found + kubectl ns/argocd verification
TestIntegration/Environment  — create staging, syncChart, delete staging
TestIntegration/Service      — detectServices, addService, getService
TestIntegration/Variables    — shared + service variables, overwrite, fromShared refs
TestIntegration/Database     — create, wait ready, connect, executeQuery, tables, tableData, delete
TestIntegration/Build        — buildService, poll buildStatus
TestIntegration/Deploy       — deploy, poll deployStatus, deployBuild, rollback + kubectl pod/deployment checks
TestIntegration/Domain       — setServiceDomain, verify httproute, remove domain
TestIntegration/Promote      — create staging, promote dev→staging, delete staging
TestIntegration/Eject        — GET /api/eject/{project}, verify zip archive
TestIntegration/GitHub       — githubRepositories (skips without GITHUB_TOKEN)
TestIntegration/Cleanup      — removeService, deleteProject, verify ns/argocd cleaned up
```

Tests share state via package-level variables:
- `testProjectName` — project created for this run (e.g., `inttest-abc123`)
- `testServiceName` — service added to the project (`vouch`)
- `testSourceURL` — GitHub source repo (`https://github.com/zeitlos/vouch`)
- `testDBName` — database created (`main`)
- `testBuildTag` / `testBuildDigest` — set after successful build

## Infrastructure Requirements

| Test Group | Services | External |
|------------|----------|----------|
| Health, Auth | gateway | — |
| Project, Environment | gateway, packager, deployer | Soft-serve, ArgoCD, Minikube |
| Service | gateway, packager, builder | Soft-serve |
| Variables | gateway, packager | Soft-serve |
| Database | gateway, packager, deployer | Soft-serve, ArgoCD, CNPG, Minikube |
| Build | gateway, builder | Zot, Docker |
| Deploy | all services | Zot, Docker, ArgoCD, Soft-serve, Minikube |
| Domain | gateway, packager, deployer | ArgoCD, Envoy Gateway |
| Promote | gateway, packager, deployer | ArgoCD, Soft-serve, Minikube |
| Eject | gateway, packager | Soft-serve |
| GitHub | gateway | Internet access |

## kubectl / psql Verification

Tests don't just trust GraphQL responses. They verify side effects:
- `kubectl get namespace` — namespaces actually created/deleted
- `kubectl get application.argoproj.io` — ArgoCD apps exist/removed
- `kubectl get cluster.postgresql.cnpg.io` — CNPG databases provisioned
- `kubectl get deployment` — workloads deployed
- `kubectl get httproute` — domains configured
- `kubectl get pods` — pods actually running

## Cleanup

Tests create a project named `inttest-<random>` and clean it up via `deleteProject` mutation in the Cleanup test phase and as a fallback in `TestMain`. If cleanup fails:

```sh
kubectl delete namespace -l lucity.dev/project=inttest-xxx --ignore-not-found
```

## Environment Variables

| Variable | Default | Purpose |
|----------|---------|---------|
| `GATEWAY_URL` | `http://localhost:8080` | Gateway endpoint |
| `JWT_SECRET` | `change-me-in-production` | JWT signing secret |
| `GITHUB_TOKEN` | (none) | GitHub OAuth token for repo tests (optional) |
