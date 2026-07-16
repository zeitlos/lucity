---
name: deploy
description: Deploy any local project to the user's Lucity platform. Triggers when the user says "deploy", "ship this", "put this online", "deploy to lucity", "get this running in the cloud", or points at a local codebase and asks to host it. Works for any application shape Railpack can build (Node, Python, Go, PHP, Java, Ruby, Rust, Elixir, Deno, Bun, .NET, static sites, and more) plus prebuilt images. Provisions PostgreSQL, Redis, and S3-compatible buckets, wires credentials, imports data, and self-heals failed rollouts.
---

# Deploy to Lucity

Take a local project and get it running on the user's Lucity platform: build, provision data
dependencies, wire credentials, deploy, self-heal, and hand back a live URL. Lucity builds with
Railpack (zero-config) and runs everything as standard Kubernetes + Helm, so anything you ship stays
ejectable.

All actions go through the `lucity` MCP tools (listed inline below). Never write to the user's source
repo — the platform only reads it.

## 1. Preflight

Run these before anything else. Stop and instruct the user if a check fails.

- **CLI present**: if the `lucity` MCP tools are not available (the server failed to connect), the CLI is not installed or not on `PATH`. Tell the user to `brew install zeitlos/tap/lucity`, then `lucity login`, then **restart the Claude Code session** so the plugin can start `lucity mcp`. Do not try to install it yourself — it cannot take effect in this session.
- **Auth**: call `get_account`. If it errors on auth, tell the user to run `lucity login` in a terminal, then retry.
- **Workspace**: `get_account` returns the active workspace. If none is active, tell the user to select one (`lucity workspace`). Every resource is workspace-scoped.
- **GitHub auth (separate token!)**: `get_account` passing does NOT mean GitHub access works — source-build calls (`list_github_repos`, `detect_services`, `add_service`) use a distinct GitHub/Logto token that can be expired even when `get_account` is green. So run `list_github_repos` as its own preflight check. If it fails with an expired-session/auth error, the user must run `lucity login` (a re-login refreshes that token); stop and tell them rather than retrying.
- **Code on GitHub**: Lucity builds from GitHub, not from local disk. Match the local repo's `origin` remote against `list_github_repos`. If it is not listed, the workspace's GitHub app cannot see it — the user either installs/authorizes the app for that repo or pushes the repo to GitHub. The platform never commits, pushes, or writes hooks; the user does that themselves.

## 2. Topology discovery

Figure out how many **deployable units** the codebase contains. A deployable unit is a directory with
its own dependency manifest or entry point that Railpack can build and run. Check these patterns:

- **Single app**: one manifest at the root. One service.
- **Monorepo**: multiple service subdirectories (npm/yarn/pnpm/bun workspaces, turborepo, nx, or plain folders each with their own manifest). Each long-running process is a service.
- **Sibling repos as one product**: separate local repos that form one system (e.g. `frontend` + `backend`). Ask the user which belong together — this is a business question.
- **Frontend + API split**: a static/SPA build plus an API server → two services.
- **Web + worker from one codebase**: same repo, a `Procfile` (or scripts) defining `web` and `worker`. Railpack picks one start command per service; deploy the codebase twice with different start commands (one per process type) or read the `Procfile`.
- **Static site**: no server, just built HTML/assets. Still a service (Railpack serves it with Caddy).
- **Prebuilt image**: the user already has an image in a registry → deploy the image directly, no build.
- **Dockerfile-only repo**: Railpack IGNORES Dockerfiles. Read the Dockerfile as *documentation* — extract the runtime version, start command, and exposed port — then configure the equivalent through Railpack knobs. If the image does something Railpack cannot reproduce, deploy the prebuilt image instead.

## 3. Reconcile detection with repo truth

Call `detect_services` to get the platform's Railpack detection facts (provider, versions, start
command) for the repo/subdir. Then **independently verify** against the repo:

- Version files: `.python-version`, `.nvmrc`, `.node-version`, `engines` in `package.json`, `go.mod`, `.tool-versions`, `runtime.txt`, `.ruby-version` / Gemfile `ruby`, `rust-toolchain.toml`, `.elixir-version`, `global.json`, `*.csproj` `TargetFramework`, `deno.json`.
- Lockfiles (which package manager), `Procfile` (start command + process types), framework configs.

On disagreement, or when detection is missing something, configure via the **three generic knobs**:

1. **Version pins** — `RAILPACK_<PROVIDER>_VERSION` style variables (exact spellings in `references/<provider>.md`).
2. **Custom install/build** — `RAILPACK_INSTALL_CMD` and `RAILPACK_BUILD_CMD` (set as variables; they override provider defaults).
3. **Custom start** — the service start command via `configure_service`.

These are set as service variables. Build-time `RAILPACK_*` variables must exist **before the first
build**, so pass them in the initial `add_service` variables (it accepts them). Consult
`references/<provider>.md` for the exact variable names and gotchas. If a detected provider has no
reference file, use the generic mechanisms above and https://railpack.com/languages/<provider>.

## 4. Universal platform contracts

Every service must satisfy these (details in `references/lucity-platform.md`):

- **Port**: the platform injects a `PORT` env var. The app MUST bind `0.0.0.0:$PORT`, not a hardcoded port or `localhost`. Most Railpack providers wire this automatically; verify the app respects `$PORT`.
- **Config via env only**: no config files baked with secrets; read everything from env vars.
- **Ephemeral filesystem**: containers lose local writes on restart. Persistent data → a volume (`create_volume`) or a bucket (`create_bucket`). Never assume local disk survives.
- **Env files are a config manifest, not a security task**: read a committed `.env`, `env.zip`, or `.env.example` for the KEYS it lists (which variables the app expects) so you know what to wire from platform resources and what to ask the user for. Never trust or copy the values. Do NOT hunt for or audit leaked secrets: the platform scans every release automatically (see §5) and flags them; duplicating that locally is wasted effort.
- **Wire by reference**: after `create_database`/`create_kv_store`/`create_bucket`, read the generated credential variables with `list_variables` and reference them in `set_variables` (ref, not literal). Never copy a credential value into a literal string.

## 5. Data dependencies scan

Scan the repo for data needs and provision them:

- **Client libraries** in the dependency manifest for PostgreSQL, Redis, or S3 (any language).
- **Env conventions**: `DATABASE_URL`, `REDIS_URL`, `S3_*` / `AWS_*` / bucket names.
- **Migrations / dumps / seeds**: `.sql` files, a `migrations/` dir, seed scripts.

Provision what the platform offers: `create_database` (PostgreSQL), `create_kv_store` (Redis),
`create_bucket` (S3-compatible). Wire each with a ref (see §4).

**External managed dependencies the platform does NOT provide** (Azure Blob Storage, Google Cloud
Storage, a hosted third-party API, SendGrid/Stripe/etc., an external database) need credentials you
cannot create. **Ask the user to provide them** and set them with `set_variables`. Do NOT stub, fake,
or hack placeholder values to force the app past its boot checks — a half-broken deploy is worse than a
clear "this app needs these credentials." Detect these from the dependency manifest (e.g.
`azure-storage-blob`, `@google-cloud/storage`, `boto3` pointed at non-Lucity endpoints) and from env
usage, and surface them before you deploy.

**Importing data**: if the app needs seed data or a dump that is not in the repo, **ask the user for
it** — do not silently deploy an empty database and call it done. To load it:
- **Small schema / migrations**: `run_sql`. (It executes from the conductor; against a remote platform
  it works, but for anything large prefer the dump path below.)
- **Bulk dumps**: `get_credentials` with `expose_publicly` for a temporary public endpoint, then a local
  `psql`/`pg_restore` (needs `sslmode=require` and an SNI-capable client, libpq ≥ 14).

Creating paid resources is a business question — ask before provisioning.

**Secret scanning is automatic.** The platform runs trufflehog + gitleaks as a `scan` phase of every
release; you do not scan locally. Read the structured result directly from `get_deploy_status` — the
release `scan` object gives `status` and `findingsCount` (no log parsing). Report any findings to the
user; `get_logs kind=scan` is only a fallback for raw detail.

## 6. Deploy loop with self-heal

Deploy, then poll and remediate. **Bound remediation to 3 iterations**, then report honestly instead of
looping forever.

1. `deploy` the service (this triggers a Railpack build + rollout).
2. Poll `get_deploy_status` every ~10s.
3. On failure, classify by the **signal**, not the technology:

> **`deploy` rebuilds from source (minutes). Config changes do not need it.** `set_variables`, `configure_service` (resources, port, start command), and volume mounts each roll the service out automatically with its **current image** in seconds. So every fix below that is config-only (a version pin for the *next* build aside) is applied and then you just poll `get_deploy_status` — **only call `deploy` again when you changed the source or a build-time `RAILPACK_*` variable.**

**Build FAILED** — `get_logs kind=build`:
- Version mismatch → pin `RAILPACK_<PROVIDER>_VERSION`.
- Missing system package → add `RAILPACK_BUILD_APT_PACKAGES` / `RAILPACK_DEPLOY_APT_PACKAGES` or `RAILPACK_PACKAGES` (see `references/railpack-config.md`).
- Wrong build command → set `RAILPACK_BUILD_CMD`.
Then `deploy` again.

**Deploy FAILED** — `get_logs kind=deploy`: bad config applied during rollout; read the message and fix the offending variable/setting.

**Rollout failure reasons** (from `get_deploy_status`):
- `OOM_KILLED` → double the memory with `configure_service` (up to the per-service ceiling). It rolls out automatically with the current image; just re-check status. No rebuild.
- `CRASH_LOOP` → `get_logs kind=runtime`. Usually wrong start command, `PORT` not honored, or a missing env var. Fix it with `configure_service` (start command) or `set_variables` — the change rolls out automatically with the current image; re-check status. No `deploy`/rebuild unless you changed source.
- `IMAGE_PULL_FAILED` → the image ref is wrong or unreachable (prebuilt-image deploys); fix the ref.
- `CONFIG_ERROR` → an invalid variable or setting; correct it.
- `UNSCHEDULABLE` → no node has room for the requested resources; lower the request or the user adds capacity.
- `QUOTA_EXCEEDED` → the workspace quota is hit; the user raises it in the dashboard (you cannot).
- `NOT_READY` / `DEADLINE_EXCEEDED` → the app started but never became healthy in time; check runtime logs for a slow or failing boot.

**NEVER silently modify the user's code.** If the fix requires a code change (e.g. the app truly
ignores `$PORT`), propose the change, explain why, and let the user apply it. Configuration lives on
the platform; code lives in the repo.

## 7. Verify & report

- `add_domain` to get a generated domain for the web service.
- `curl` the endpoint until it returns 200 (TLS provisioning can take ~1 minute; retry).
- Report back:
  - Live URL(s).
  - Services created and their start commands.
  - Resources provisioned (databases, KV stores, buckets, volumes).
  - **Every configuration decision and why** (version pins, custom commands, memory bumps).
  - Any secret-scan findings the platform reported (read from `get_deploy_status`'s `scan` object).
  - Remaining manual steps for the user.
- If a remediation needed knowledge that was missing from `references/`, note at the end which reference file deserves a new line. The skill improves by contribution.

## 8. Ask vs. decide

- **Decide** anything evidenced in the repo (provider, versions, start command, which resources the code needs, what to pin) — and report it.
- **Ask** only business questions: which repos form the product, whether it should be public, whether you may import a given data dump, whether you may create paid resources, and **any credentials for external managed services the platform does not provide** (Azure/GCS/third-party APIs) plus **any data dump the app needs that is not in the repo**. If a required external credential or dump is missing, stop and ask rather than shipping a broken deploy.

## MCP tools

Account/projects: `get_account`, `list_projects`, `get_project`, `create_project` (auto-creates a
`development` environment), `create_environment`.
Detection/source: `detect_services`, `list_github_repos`.
Services: `add_service` (accepts initial variables — put build-time `RAILPACK_*` pins here so the first
build sees them; optional `cpu`/`memory` to size it), `configure_service` (start command, resources),
`set_variables`, `list_variables`.
Resources: `create_database` (optional `cpu`/`memory`), `create_kv_store`, `create_bucket`,
`create_volume`, `get_credentials`, `run_sql`.
Deploy: `deploy`, `get_deploy_status`, `get_logs`, `rollback`, `add_domain`.

There are no delete tools — the user removes projects, services, and resources from the dashboard.

## References

- `references/railpack-config.md` — global Railpack config (build/install/start/packages, `railpack.json`, Procfile priority).
- `references/lucity-platform.md` — PORT injection, resource-quantity format, ID formats, variable refs, public DB access, rollout failure reasons.
- `references/<provider>.md` — per-language detection, version resolution, start command, config variables, and common failure fixes. One file per Railpack provider.
