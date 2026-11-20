# Lucity platform contracts

What every workload must satisfy on Lucity, and how the platform's shapes (IDs, quantities, variable
refs, failure reasons) work. Lucity runs standard Kubernetes + Helm under the hood; everything here maps
to plain infrastructure-as-code, so anything you deploy stays ejectable.

## PORT injection

The platform injects a `PORT` environment variable into every service: `8080` for repository services
unless the port is changed in the service settings. Framework defaults (`8000`, `3000`) only
appear as fallbacks like `${PORT:-8000}` and never take effect. The service port is also the port in its
internal endpoint, so settle it before wiring other services to it; changing it later silently breaks
any consumer still pointing at the old one. The app **must** bind
`0.0.0.0:$PORT`. Binding a hardcoded port, or binding `localhost` / `127.0.0.1`, makes the service
unreachable and the rollout fails readiness. Most Railpack providers already read `$PORT` (Node, Python
frameworks, .NET via `ASPNETCORE_URLS`, etc.); a few need the start command adjusted. If the app truly
ignores `$PORT`, that is a code fix — propose it, do not rewrite the code silently.

## Resource quantities

CPU and memory use Kubernetes quantity strings:

- CPU: millicores or cores — `250m`, `500m`, `1`, `2`.
- Memory: binary suffixes — `256Mi`, `512Mi`, `1Gi`, `2Gi`.

Give CPU and memory together when adding a service or creating a database to size it at creation, or
change them later in the service settings; leave both out for platform defaults (a service gets `512Mi`
of memory). Each service or
database has its own ceiling below the whole-workspace quota: a larger request is rejected, and
exhausting the quota surfaces as `QUOTA_EXCEEDED`.

## ID formats

Every project, environment, service and resource has an **id** and a **name**:

- The **id** is what tools take. It is lowercase letters, digits and hyphens, 2 to 16 characters, and never changes. Composite ids start with the workspace: project `workspace/project`, environment `workspace/project/environment`, service or resource `workspace/project/environment/<id>`. The workspace segment may be left out.
- The **name** is what people read in the dashboard. It can contain spaces and can be renamed, so never pass a name where a tool expects an id.

Everything you create (projects, environments, services, databases, key-value stores, buckets,
volumes) takes a name and an optional id. Leave the id out and the platform derives it from the name,
cut to 16 characters with a counter added when it is taken. Always continue with the id the tool
returns. A new project comes with a `development` environment; create more environments as needed. Inside one environment, services, databases, key-value stores, buckets and volumes
share one set of ids, so a database can't reuse a service's id.

## Variable references (wire, don't copy)

Provisioned resources (databases, key-value stores, buckets) generate their own credential variables. To connect an app, **reference** those variables rather than
copying secret values into literals:

1. List the variables to see the generated credential variable names.
2. Set the app's expected key (e.g. `DATABASE_URL`, `REDIS_URL`, `S3_ENDPOINT`) as a **ref** to the generated variable.

A ref stays correct across rotations and never leaks a literal secret into your config or logs.
Secrets that no platform resource provides take one of two routes, and neither passes a value through
you:

- **App-owned** (`SECRET_KEY_BASE`, `AUTH_SECRET`, a JWT secret): set it as a generated variable,
  which stores a random value you never see. A key that already has a value keeps it.
- **Everything else** (third-party API keys, `RAILS_MASTER_KEY`, anything that must match a value
  elsewhere): the user adds it in the dashboard, in the service's **Variables** tab. That holds even
  when the value sits in a committed or local file, so never copy it into a variable yourself.

Listing variables returns keys and refs, never values. To confirm what a variable holds, check it
against the value you expect. The check answers match, mismatch, missing, or ref.

## Build vs. runtime variables

Variables reach the app at runtime. The **literal values set on the service itself** also reach the
Railpack build; refs and shared variables never do. A variable matters at build time when Railpack reads
it (`RAILPACK_NODE_VERSION`, `RAILPACK_BUILD_CMD`, ...) or when a framework inlines it into the bundle
(`NEXT_PUBLIC_*`, `VITE_*`, `NUXT_PUBLIC_*`, `REACT_APP_*`, `PUBLIC_*` in SvelteKit and Astro). Those must
be present before the first build, which starts as soon as a repository service is added, so pass them
in its initial variables as literal values. Changing one requires a fresh deploy (rebuild); changing a pure runtime variable takes effect on the next rollout.

## HTTPS ends at the edge

Public traffic is HTTPS up to the platform's edge; the app itself receives plain HTTP with
`X-Forwarded-Proto: https` and `X-Forwarded-For`. A framework that doesn't trust the proxy builds
`http://` URLs, redirects in a loop when it forces HTTPS, rejects CSRF origins, or won't set secure
cookies. The fix is one setting in the app, a code change to propose to the user:

- **Laravel**: trust all proxies, `$middleware->trustProxies(at: '*');` in `bootstrap/app.php`
  (Laravel 11+), or `protected $proxies = '*';` in `TrustProxies` (older).
- **Express**: `app.set('trust proxy', true)`.
- **Django**: `SECURE_PROXY_SSL_HEADER = ('HTTP_X_FORWARDED_PROTO', 'https')`, plus the public origin
  in `CSRF_TRUSTED_ORIGINS` and the host in `ALLOWED_HOSTS`.
- **Rails**: `config.assume_ssl = true` (7.1+) next to `config.force_ssl = true`.

## Public database access (SNI + TLS)

`lucity db expose <db>` gives a database a temporary public hostname (useful for importing a bulk dump
with a local `psql`/`pg_restore`) and allows only the address of the machine that ran it,
`lucity db credentials <db>` lists its connection details as the `PLATFORM` entry, and
`lucity db unexpose <db>` removes the hostname and its allowed addresses again. The user runs these in
their own terminal so the credentials never reach you, and the restore has to run from that same
machine. Other addresses are added in the dashboard (Connect tab, Allowed addresses). The endpoint requires:

- `sslmode=require` (TLS is mandatory). The `PLATFORM` entry's `uri` already carries it. The hostname has a publicly trusted certificate, so node-postgres, Prisma and `sslmode=verify-full` work without a CA file or `no-verify`.
- An SNI-capable client — libpq ≥ 14. Older clients that do not send SNI get an "SSL EOF detected" style error because routing is by SNI.
- An allowed address. A client from any other address gets the same "SSL EOF detected" error.

Use it for one-off imports, then rely on in-cluster refs for the running app.

## Rollout failure reasons

The deploy status reports one of these reasons on failure. Remedies:

| Reason | Meaning | Remedy |
| :-- | :-- | :-- |
| `OOM_KILLED` | Container exceeded its memory limit. | Double memory in the service settings (up to the per-service ceiling), re-check. No rebuild. |
| `CRASH_LOOP` | Container starts then exits repeatedly. | Runtime logs: wrong start command, `PORT` not honored, or missing env var. Fix the service settings or variables, re-check. No rebuild. |
| `IMAGE_PULL_FAILED` | The image ref cannot be pulled. | Fix the image reference (prebuilt-image deploys) or rebuild. |
| `CONFIG_ERROR` | Invalid configuration applied during rollout. | Read the message; correct the offending variable/setting. |
| `QUOTA_EXCEEDED` | Workspace resource quota hit. | User raises the quota in the dashboard; you cannot. |
| `UNSCHEDULABLE` | No node has room for the requested resources. | Lower CPU/memory requests, or the user adds capacity. |
| `NOT_READY` | Container ran but never passed readiness in time. | Runtime logs; usually a slow boot or a port-binding issue. |
| `DEADLINE_EXCEEDED` | Rollout did not complete within the deadline. | Same as `NOT_READY`; inspect runtime logs for a hanging start. |

Bound remediation to 3 iterations, then report honestly rather than looping.

## Release queue

Each environment runs one release (build, scan, deploy) at a time, and builds share a small
cluster-wide cap. Cancelling a queued or running build ends its release as `CANCELLED`, and the service
keeps serving what it served. Several services added back to back therefore build one after another; a
later build showing as queued is waiting its turn. The newest release after a config change is a
settings-only one that can read `SUPERSEDED` with no build info while the service is healthy, so follow
the release you are tracking.

## Generated domains

Adding a domain without a hostname generates a platform subdomain right away, but until the service's
first rollout succeeds the reply shows only the internal endpoint. The public domain appears in the
deploy status endpoints once the service is up.

## Internal networking (service to service)

Every service with a port gets a cluster-internal endpoint. Its shape is

```
lucity-app-<service>.<namespace>.svc.cluster.local:<port>
```

The namespace is `<workspace>-<project>-<environment>-<10-char hash>`; the hash is not derivable by
hand, so **never construct this hostname yourself**. Read it from the project (or from the result of
adding the service): each service lists its endpoints, and the internal one carries the host, while the
service's port is the port. Wire it into the consumer as a literal variable, e.g.
`API_URL=http://lucity-app-api.<namespace>.svc.cluster.local:3000`. Inside the same environment the
short name `lucity-app-<service>:<port>` resolves too; the fully qualified form is the safe default.

Internal reach is **per environment**: a NetworkPolicy blocks traffic between environments and
projects, so `development` cannot call `production`'s API by cluster DNS. Cross-environment calls go
through a public domain. Internal endpoints are plain HTTP/TCP, no TLS.

Databases, key-value stores, and buckets are different: their hosts arrive via variable refs (see
above), not via endpoints. Only service-to-service links use the internal endpoint.

## Ephemeral filesystem and volumes

Container disk is ephemeral: local writes vanish on restart, redeploy, or reschedule. Persistent data
needs a volume or an S3-compatible bucket. Never assume a file
written at runtime survives.

Volume rules, all enforced at creation or mount time:

- **Size is 10Gi to 128Gi.** Anything outside that is rejected (`size must be between 10Gi and 128Gi`),
  so a tiny config or upload directory still costs 10Gi. Default to `10Gi` unless the data is larger.
- **Grow-only.** A volume can be expanded later, never shrunk.
- **One service, one mount path.** A volume mounts into exactly one service at exactly one path; that
  service must run a single replica with autoscaling off. Web services that scale horizontally should
  use a bucket instead.
- **Mounting rolls out with the current image** (no rebuild). A volume can be created and mounted into
  its service in one step.
- **The mount root ships with `lost+found`.** Tools that insist on an empty data directory (MySQL
  `--initialize`, some migrators) must point at a subdirectory of the mount, never the mount root.
- **Ownership follows the service's run-as user.** For prebuilt images, set it when adding the service
  (e.g. `999` for postgres/mysql/redis images, `1000` for node-based images); the volume is then
  writable by that uid. Containers run with all capabilities dropped, so an entrypoint that `chown`s the data dir
  fails unless the uid already matches.

## Custom domains

Subdomains need `TXT` + `CNAME`, which every DNS provider supports. An apex (`example.com`) needs
either a redirect to `www.example.com` (add `www` first; records: `A` for the apex, `CNAME` for the target) or, without
a redirect, an `ALIAS` record, which only providers with ALIAS/ANAME/CNAME flattening offer
(Cloudflare, Route 53, DNSimple, Porkbun). When the provider is unknown, use the redirect.
