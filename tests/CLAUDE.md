# Integration Tests

End-to-end tests that exercise the full Lucity platform through the conductor's GraphQL API and verify side effects with `kubectl` and `psql`.

> **Status:** the suite is in disrepair after the conductor merge — the platform consolidated from separate gRPC services into the single conductor binary, and deploys moved from GitOps/ArgoCD to imperative Helm releases. The targets below are the intended interface; expect breakage until the suite is restored. Day-to-day verification currently leans on `go build ./...`, `go vet ./...`, the GraphQL playground, and `tmp/logs/conductor.log`.

## Running

Services must be running (`make dev`) and infrastructure must be up (`make infra && make infra-forward`) before running tests.

```sh
make test-integration-short   # quick subset
make test-integration         # full suite (+ Minikube side-effect checks)
make test-watch               # auto-rerun on file changes (requires watchexec)
```

## Logs

Test output is written to `tmp/logs/tests.log`. Runner status is in `tmp/dev/tests.status`.

```sh
cat tmp/logs/tests.log
```

Look for `--- FAIL` lines to identify failures and `ok`/`FAIL` at the end for the overall result.

## Test Organization

All tests run sequentially via `TestIntegration` in `main_test.go`. Each test group is a subtest covering one slice of the platform:

```
TestIntegration/Health       — conductor health + GraphQL playground
TestIntegration/Auth         — JWT auth, unauth, invalid token
TestIntegration/Project      — create, list, get, not-found (+ kubectl namespace checks)
TestIntegration/Environment  — create, sync, delete
TestIntegration/Service      — detect services, add service, get service
TestIntegration/Variables    — shared + service variables, overwrite, fromShared refs
TestIntegration/Database     — create, wait ready, connect, query, tables, delete
TestIntegration/Build        — build service, poll build status
TestIntegration/Deploy       — deploy, poll status, rollback (+ kubectl pod/deployment checks)
TestIntegration/Domain       — set service domain, verify HTTPRoute, remove domain
TestIntegration/Promote      — promote dev → staging
TestIntegration/Eject        — fetch eject archive, verify contents
TestIntegration/Cleanup      — remove service, delete project, verify resources gone
```

Tests share state via package-level variables (e.g. `testProjectName`, `testServiceName`, `testSourceURL`, `testDBName`, `testBuildTag` / `testBuildDigest`).

## Side-Effect Verification

Tests don't just trust GraphQL responses. They verify cluster state:

- `kubectl get namespace` — namespaces actually created/deleted
- `helm list` / `kubectl get secret -l owner=helm` — Helm releases applied/removed
- `kubectl get cluster.postgresql.cnpg.io` — CNPG databases provisioned
- `kubectl get deployment` / `kubectl get pods` — workloads deployed and running
- `kubectl get httproute` — domains configured

## Cleanup

Tests create a project named `inttest-<random>` and clean it up via the `deleteProject` mutation (and as a fallback in `TestMain`). If cleanup fails:

```sh
kubectl delete namespace -l lucity.dev/project=inttest-xxx --ignore-not-found
```

## Environment Variables

| Variable | Default | Purpose |
|----------|---------|---------|
| `CONDUCTOR_URL` | `http://localhost:8080` | Conductor GraphQL endpoint |
| `AUTH_TEST_SECRET` | `change-me-in-production` | HS256 test token secret (must match the conductor's `AUTH_TEST_SECRET`) |
