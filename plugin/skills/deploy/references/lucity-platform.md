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

Set these with `configure_service`. When remediating `OOM_KILLED`, double the memory (`512Mi` → `1Gi`)
and re-check; stay within the workspace quota (`QUOTA_EXCEEDED` if you exceed it).

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
| `OOM_KILLED` | Container exceeded its memory limit. | Double memory via `configure_service` (cap at quota), re-check. No rebuild. |
| `CRASH_LOOP` | Container starts then exits repeatedly. | `get_logs kind=runtime`: wrong start command, `PORT` not honored, or missing env var. Fix + redeploy. |
| `IMAGE_PULL_FAILED` | The image ref cannot be pulled. | Fix the image reference (prebuilt-image deploys) or rebuild. |
| `CONFIG_ERROR` | Invalid configuration applied during rollout. | Read the message; correct the offending variable/setting. |
| `QUOTA_EXCEEDED` | Workspace resource quota hit. | User raises the quota in the dashboard; you cannot. |
| `UNSCHEDULABLE` | No node has room for the requested resources. | Lower CPU/memory requests, or the user adds capacity. |
| `NOT_READY` | Container ran but never passed readiness in time. | `get_logs kind=runtime`; usually a slow boot or a port-binding issue. |
| `DEADLINE_EXCEEDED` | Rollout did not complete within the deadline. | Same as `NOT_READY`; inspect runtime logs for a hanging start. |

Bound remediation to 3 iterations, then report honestly rather than looping.

## Ephemeral filesystem

Container disk is ephemeral: local writes vanish on restart, redeploy, or reschedule. Persistent data
needs a volume (`create_volume`) or an S3-compatible bucket (`create_bucket`). Never assume a file
written at runtime survives.
