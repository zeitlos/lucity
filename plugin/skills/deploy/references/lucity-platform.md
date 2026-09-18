# Lucity platform contracts

What every workload must satisfy on Lucity, and how the platform's shapes (IDs, quantities, variable
refs, failure reasons) work. Lucity runs standard Kubernetes + Helm under the hood; everything here maps
to plain infrastructure-as-code, so anything you deploy stays ejectable.

## PORT injection

The platform injects a `PORT` environment variable into every service. The app **must** bind
`0.0.0.0:$PORT`. Binding a hardcoded port, or binding `localhost` / `127.0.0.1`, makes the service
unreachable and the rollout fails readiness. Most Railpack providers already read `$PORT` (Node, Python
frameworks, .NET via `ASPNETCORE_URLS`, etc.); a few need the start command adjusted. If the app truly
ignores `$PORT`, that is a code fix — propose it, do not rewrite the code silently.

## Resource quantities

CPU and memory use Kubernetes quantity strings:

- CPU: millicores or cores — `250m`, `500m`, `1`, `2`.
- Memory: binary suffixes — `256Mi`, `512Mi`, `1Gi`, `2Gi`.

Pass `cpu` and `memory` (both together) to `add_service` / `create_database` to size a resource at
creation, or set them later with `configure_service`; omit them for platform defaults. Each service or
database has its own ceiling below the whole-workspace quota: a larger request is rejected, and
exhausting the quota surfaces as `QUOTA_EXCEEDED`.

## ID formats

Resources are workspace-scoped and identified by human-readable composite IDs:

- **Workspace**: the tenant boundary, derived from the login token (never passed as an argument).
- **Project**: workspace-scoped, human-readable; the project name *is* the project ID.
- **Environment**: named within a project. `create_project` auto-creates a `development` environment; add more with `create_environment`.
- **Service / resource**: named within a project + environment. Composite IDs read as `workspace/project/environment/service`.

You generally pass the readable name to each tool; the platform resolves scope from the token and the
project/environment you name.

## Variable references (wire, don't copy)

Provisioned resources (`create_database`, `create_kv_store`, `create_bucket`) generate their own
credential variables on the service. To connect an app, **reference** those variables rather than
copying secret values into literals:

1. `list_variables` to see the generated credential variable names.
2. `set_variables` to bind the app's expected key (e.g. `DATABASE_URL`, `REDIS_URL`, `S3_ENDPOINT`) as a **ref** to the generated variable.

A ref stays correct across rotations and never leaks a literal secret into your config or logs. Any
secret you find committed in the repo goes into `set_variables` too — and you tell the user what you
found and where.

## Build vs. runtime variables

Variables reach the app at runtime. `RAILPACK_*` variables additionally reach the **Railpack build** as
environment variables, so build-time pins (`RAILPACK_NODE_VERSION`, `RAILPACK_BUILD_CMD`, ...) must be
present before the first build. Pass them in the initial `add_service` variables. Changing a build-time
variable requires a fresh `deploy` (rebuild); changing a pure runtime variable takes effect on the next
rollout.

## Public database access (SNI + TLS)

`get_credentials` with `expose_publicly` mints a temporary public endpoint for a database (useful for
importing a bulk dump with a local `psql`). The endpoint requires:

- `sslmode=require` (TLS is mandatory).
- An SNI-capable client — libpq ≥ 14. Older clients that do not send SNI get an "SSL EOF detected" style error because routing is by SNI.

Use it for one-off imports, then rely on in-cluster refs for the running app.

## Rollout failure reasons

`get_deploy_status` returns one of these reasons on failure. Remedies:

| Reason | Meaning | Remedy |
| :-- | :-- | :-- |
| `OOM_KILLED` | Container exceeded its memory limit. | Double memory via `configure_service` (up to the per-service ceiling), re-check. No rebuild. |
| `CRASH_LOOP` | Container starts then exits repeatedly. | `get_logs kind=runtime`: wrong start command, `PORT` not honored, or missing env var. Fix + redeploy. |
| `IMAGE_PULL_FAILED` | The image ref cannot be pulled. | Fix the image reference (prebuilt-image deploys) or rebuild. |
| `CONFIG_ERROR` | Invalid configuration applied during rollout. | Read the message; correct the offending variable/setting. |
| `QUOTA_EXCEEDED` | Workspace resource quota hit. | User raises the quota in the dashboard; you cannot. |
| `UNSCHEDULABLE` | No node has room for the requested resources. | Lower CPU/memory requests, or the user adds capacity. |
| `NOT_READY` | Container ran but never passed readiness in time. | `get_logs kind=runtime`; usually a slow boot or a port-binding issue. |
| `DEADLINE_EXCEEDED` | Rollout did not complete within the deadline. | Same as `NOT_READY`; inspect runtime logs for a hanging start. |

Bound remediation to 3 iterations, then report honestly rather than looping.

## Internal networking (service to service)

Every service with a port gets a cluster-internal endpoint. Its shape is

```
lucity-app-<service>.<namespace>.svc.cluster.local:<port>
```

The namespace is `<workspace>-<project>-<environment>-<10-char hash>`; the hash is not derivable by
hand, so **never construct this hostname yourself**. Read it from `get_project` (or the `add_service`
result): each service lists `endpoints`, and the one with `type: INTERNAL` carries the host, while the
service's `port` is the port. Wire it into the consumer as a literal variable, e.g.
`API_URL=http://lucity-app-api.<namespace>.svc.cluster.local:3000`. Inside the same environment the
short name `lucity-app-<service>:<port>` resolves too; the fully qualified form is the safe default.

Internal reach is **per environment**: a NetworkPolicy blocks traffic between environments and
projects, so `development` cannot call `production`'s API by cluster DNS. Cross-environment calls go
through a public domain. Internal endpoints are plain HTTP/TCP, no TLS.

Databases, key-value stores, and buckets are different: their hosts arrive via variable refs (see
above), not via endpoints. Only service-to-service links use the internal endpoint.

## Ephemeral filesystem and volumes

Container disk is ephemeral: local writes vanish on restart, redeploy, or reschedule. Persistent data
needs a volume (`create_volume`) or an S3-compatible bucket (`create_bucket`). Never assume a file
written at runtime survives.

Volume rules, all enforced at creation or mount time:

- **Size is 10Gi to 1Ti.** Anything smaller is rejected (`volume size must be between 10Gi and 1Ti`),
  so a tiny config or upload directory still costs 10Gi. Default to `10Gi` unless the data is larger.
- **Grow-only.** A volume can be expanded later, never shrunk.
- **One service, one mount path.** A volume mounts into exactly one service at exactly one path; that
  service must run a single replica with autoscaling off. Web services that scale horizontally should
  use a bucket instead.
- **Mounting rolls out with the current image** (no rebuild). `create_volume` with `mount_service` +
  `mount_path` provisions and mounts in one call.
- **The mount root ships with `lost+found`.** Tools that insist on an empty data directory (MySQL
  `--initialize`, some migrators) must point at a subdirectory of the mount, never the mount root.
- **Ownership follows the service `user`.** For prebuilt images, pass `user` on `add_service` (e.g.
  `999` for postgres/mysql/redis images, `1000` for node-based images); the volume is then writable by
  that uid. Containers run with all capabilities dropped, so an entrypoint that `chown`s the data dir
  fails unless the uid already matches.

## Custom domains

Subdomains need `TXT` + `CNAME`, which every DNS provider supports. An apex (`example.com`) needs
either `redirect_to: www.example.com` (records: `A` for the apex, `CNAME` for the target) or, without
a redirect, an `ALIAS` record, which only providers with ALIAS/ANAME/CNAME flattening offer
(Cloudflare, Route 53, DNSimple, Porkbun). When the provider is unknown, use the redirect.
