---
name: deploy
description: Deploy any local project to the user's Lucity platform. Triggers when the user says "deploy", "ship this", "put this online", "deploy to lucity", "get this running in the cloud", or points at a local codebase and asks to host it. Works for any application shape Railpack can build (Node, Python, Go, PHP, Java, Ruby, Rust, Elixir, Deno, Bun, .NET, static sites, and more) plus prebuilt images. Provisions PostgreSQL, Redis, and S3-compatible buckets, wires credentials, imports data, and self-heals failed rollouts.
---

# Deploy to Lucity

Take a local project and get it running on the user's Lucity platform: build, provision data
dependencies, wire credentials, deploy, self-heal, and hand back a live URL. Lucity builds with
Railpack (zero-config) and runs everything as standard Kubernetes + Helm, so anything you ship stays
ejectable.

All actions go through the tools of the Lucity MCP server. This skill says what to do; the tools'
own descriptions say exactly how, with which arguments. They come from the CLI the user runs, so when
the two disagree, trust the tool descriptions. Never write to the user's source repo — the platform
only reads it.

## 1. Preflight

Run these before anything else. Stop and instruct the user if a check fails.

- **CLI present**: if the `lucity` MCP tools are not available (the server failed to connect), the CLI is not installed or not on `PATH`. Tell the user to `brew install zeitlos/tap/lucity`, then `lucity login`, then **restart the Claude Code session** so the plugin can start `lucity mcp`. Do not try to install it yourself — it cannot take effect in this session.
- **Auth**: look up the signed-in account. If that errors on auth, tell the user to run `lucity login` in a terminal, then retry.
- **Workspace**: the account lookup returns the platform URL, the active workspace, and every workspace the user belongs to. If none is active, tell the user to select one (`lucity workspace`). If they belong to **more than one**, ask which workspace the project belongs in before creating anything, even when one is active: where a project lives is a business decision. Name the platform URL in the same question so a stale login to an old platform is caught. Tool IDs must start with the active workspace. To switch, the user runs `lucity workspace <id>` in a terminal; the MCP server picks it up on the next call, and the account lookup confirms which workspace is active.
- **GitHub auth (separate token!)**: a working account lookup does NOT mean GitHub access works — listing repositories, detecting services and adding a repository service use a distinct GitHub token that can be expired even when the account lookup is green. So list the GitHub repositories as its own preflight check. If it fails with an expired-session/auth error, the user must run `lucity login` (a re-login refreshes that token); stop and tell them rather than retrying.
- **Code on GitHub**: Lucity builds from GitHub, not from local disk. Match the local repo's `origin` remote against the repositories the workspace can see. If it is not listed, the workspace's GitHub app cannot see it — the user either installs/authorizes the app for that repo or pushes the repo to GitHub. The platform never commits, pushes, or writes hooks; the user does that themselves.

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

Have the platform detect the services in the repo or subfolder to get Railpack's detection facts
(provider, versions, start command). Then **independently verify** against the repo:

- Version files: `.python-version`, `.nvmrc`, `.node-version`, `engines` in `package.json`, `go.mod`, `.tool-versions`, `runtime.txt`, `.ruby-version` / Gemfile `ruby`, `rust-toolchain.toml`, `.elixir-version`, `global.json`, `*.csproj` `TargetFramework`, `deno.json`.
- Lockfiles (which package manager), `Procfile` (start command + process types), framework configs.

**Railpack only sees the service's build directory.** When a service builds from a subfolder (a
monorepo's `backend/`), version files at the repo root (`.python-version`, `.nvmrc`, `.tool-versions`)
are invisible to it and the build silently falls back to the provider default. If the version file is
not inside the build directory, pin `RAILPACK_<PROVIDER>_VERSION` in the initial variables of the service when you add it.

On disagreement, or when detection is missing something, configure via the **three generic knobs**:

1. **Version pins** — `RAILPACK_<PROVIDER>_VERSION` style variables (exact spellings in `references/<provider>.md`).
2. **Custom install/build** — `RAILPACK_INSTALL_CMD` and `RAILPACK_BUILD_CMD` (set as variables; they override provider defaults).
3. **Custom start** — the start command in the service settings. It runs without a shell, replacing the image's entrypoint: write variables as `$(PORT)`, and wrap `&&`, pipes or `${VAR:-default}` in `sh -c '…'`.

These are set as service variables. Build-time variables (`RAILPACK_*`, and values a frontend
framework inlines into its bundle such as `NEXT_PUBLIC_*`, `VITE_*`, `NUXT_PUBLIC_*`) must exist
**before the first build**, and adding a repository service starts that build right away, so pass
them in the service's initial variables as literal values. One set afterwards only reaches the next
rebuild. Builds see literal values set on the service itself, never refs or shared variables.
Avoid single quotes in `RAILPACK_INSTALL_CMD` / `RAILPACK_BUILD_CMD`: Railpack wraps them in
`sh -c '…'` without escaping, so a quote inside ends the string early and backslashes get eaten
(`printf '%s\n'` writes a literal `n`). Use double quotes, or put the logic in a script in the repo.
Consult `references/<provider>.md` for the exact variable names and gotchas. If a detected provider has no
reference file, use the generic mechanisms above and https://railpack.com/languages/<provider>.

## 4. Universal platform contracts

Every service must satisfy these (details in `references/lucity-platform.md`):

- **Port**: the platform injects a `PORT` env var, **`8080` for repository services** unless you change the port in the service settings. The framework's usual port (`8000`, `3000`) is only the fallback in a start command like `${PORT:-8000}` and never applies. The app MUST bind `0.0.0.0:$PORT`, not a hardcoded port or `localhost`. Most Railpack providers wire this automatically; verify the app respects `$PORT`. **Settle each service's port before writing its internal URL into another service**: changing it later silently breaks every consumer that still points at the old port.
- **Config via env only**: no config files baked with secrets; read everything from env vars.
- **HTTPS ends at the edge**: the app receives plain HTTP with `X-Forwarded-Proto: https`. Frameworks that
  don't trust the proxy build `http://` URLs, loop on HTTPS redirects, reject CSRF origins, or drop
  secure cookies. Each needs a one-line setting, which is a code change to propose (see
  `references/lucity-platform.md` for Laravel, Express, Django and Rails).
- **Memory**: a service gets **512Mi** unless sized. A dev-style start (`nodemon`, `next dev`, `vite`,
  `tsc --watch`, a `start` script that runs a watcher) gets `OOM_KILLED` there and isn't meant for
  production anyway: use the production start command, and only then size memory.
- **Migrations and seeds**: there is no separate pre-start step yet. Chain them into the start command
  with a shell, e.g. `sh -c 'npm run migrate && exec npm start'` (`exec` so the app gets the stop
  signal). They run on every instance start, so they must be safe to repeat. Laravel (Railpack's start
  script) and Django (its default start command) already migrate.
- **Ephemeral filesystem**: containers lose local writes on restart. Persistent data → a volume (**10Gi minimum**, one service per volume, single replica) or a bucket (for anything that scales). Never assume local disk survives.
- **Service-to-service links use the internal endpoint**: read the service's internal endpoint from the project (or from the result of adding the service) plus the service port, and set it as a literal variable on the consumer (e.g. `API_URL=http://lucity-app-api.<namespace>.svc.cluster.local:3000`). Never build the hostname by hand (the namespace carries a hash) and never expect cross-environment reach: internal DNS only works within one environment.
- **Env files are a config manifest, not a security task**: read a committed `.env`, `env.zip`, or `.env.example` for the KEYS it lists (which variables the app expects) so you know what to wire from platform resources, what to generate, and what the user has to add. Never trust or copy the values. Do NOT hunt for or audit leaked secrets: the platform scans every release automatically (see §5) and flags them; duplicating that locally is wasted effort.
- **Wire by reference**: after creating a database, key-value store or bucket, list the credential variables it generated and reference them (ref, not literal). Never copy a credential value into a literal string. Provision these before adding the service and pass the refs in its initial variables (listing the environment's variables shows which refs exist before the service does), so the first rollout already has them; for an existing service, set them as variables. Builds never see refs, only literal values.
- **Generate app-owned secrets**: a key the app only needs to be random (`SECRET_KEY_BASE`, `AUTH_SECRET`, a JWT secret) gets a generated value, in the initial variables or when setting variables later, which stores a random value you never see. A key that already has a value keeps it, so repeating the call is safe.
- **Secret values never reach you**: the tools return keys, refs, and the results of variable checks, never variable values or credentials. To confirm a value, check the variable against the value you expect. Don't route around this by running `lucity vars list` or `lucity db credentials` yourself or by asking the user to paste a secret into the chat.

## 5. Data dependencies scan

Scan the repo for data needs and provision them:

- **Client libraries** in the dependency manifest for PostgreSQL, Redis, or S3 (any language).
- **Env conventions**: `DATABASE_URL`, `REDIS_URL`, `S3_*` / `AWS_*` / bucket names.
- **Migrations / dumps / seeds**: `.sql` files, a `migrations/` dir, seed scripts.

Provision what the platform offers: a PostgreSQL database, a Redis-compatible key-value store, an
S3-compatible bucket. Wire each with a ref (see §4).

**External managed dependencies the platform does NOT provide** (Azure Blob Storage, Google Cloud
Storage, a hosted third-party API, SendGrid/Stripe/etc., an external database) need credentials you
cannot create, and so do values that must match something elsewhere, like `RAILS_MASTER_KEY`. **Ask
the user to add them** in the dashboard, in the service's **Variables** tab, instead of pasting them
into the chat. Multi-line values (PEM keys, certificates, JSON credentials) keep their line breaks when
pasted there, or the user runs `lucity vars set <service> --from-file KEY=./key.pem`. Name the exact keys, then confirm they exist by listing the service's variables. Do NOT stub, fake,
or hack placeholder values to force the app past its boot checks — a half-broken deploy is worse than a
clear "this app needs these credentials." Detect these from the dependency manifest (e.g.
`azure-storage-blob`, `@google-cloud/storage`, `boto3` pointed at non-Lucity endpoints) and from env
usage, and surface them before you deploy.

**Importing data**: if the app needs seed data or a dump that is not in the repo, **ask the user for
it** — do not silently deploy an empty database and call it done. To load it:
- **Small schema / migrations**: run the SQL through the server. (For anything large, prefer the dump
  path below.)
- **Rows you need outside this conversation** (a file, a local database, `jq`): `lucity db query <db> "<sql>" --json`
  in a shell, up to 1,000 rows per statement. It returns data, not credentials, so you may run it yourself.
- **Bulk dumps**: the credentials a local `psql`/`pg_restore` needs are never shown to you, so the user
  runs the import from their own terminal. Hand them the steps with the database id filled in:
  `lucity db expose <db>` (temporary public hostname, open only to the machine that ran it),
  `lucity db credentials <db>` (its `PLATFORM` entry's `uri` already has `sslmode=require`), the
  restore from that same machine with libpq ≥ 14 (it must send SNI), then `lucity db unexpose <db>`. Never run `lucity db credentials` yourself or ask the user
  to paste its output. Once they are done, spot-check the data with a query.

Creating paid resources is a business question — ask before provisioning.

**Secret scanning is automatic.** The platform runs trufflehog + gitleaks as a `scan` phase of every
release; you do not scan locally. Read the result from the release's deploy status: the scan's
outcome and, when it found something, each finding's rule, file, line and whether the credential is
verified live. Report them to the user; verified ones need rotating.

## 6. Deploy loop with self-heal

Poll the first build, then remediate. **Bound remediation to 3 iterations**, then report honestly
instead of looping forever.

1. Adding a repository service already started the first Railpack build + rollout (an image service
   rolls out without a build). Do **not** trigger a deploy now, which only builds a second time.
2. Follow the deploy status of the release that adding the service returned, and let the status call
   wait for it to settle rather than polling in a loop. Config changes made meanwhile create newer
   releases that would otherwise hide the build, and one issued while the rollout runs can fail with
   "another operation ... is in progress", so retry it after a few seconds.
   - **Always ask for the release you are tracking.** The newest release after a config change is a
     settings-only one that can show as `SUPERSEDED` with no build info even though the service is
     healthy. That is not a failure.
   - **Releases run one at a time per environment** (and builds share a small cluster-wide cap). Adding
     several services back to back does not build them in parallel: the later builds sit queued until
     the earlier release finishes. A queued build is waiting its turn, not stuck. A build queued for the
     wrong commit can be cancelled.
3. On failure, classify by the **signal**, not the technology:

> **A deploy rebuilds from source (minutes). Config changes do not need it.** Setting variables, changing service settings (resources, port, start command), and mounting volumes each roll the service out automatically with its **current image** in seconds. So every fix below that is config-only (a version pin for the *next* build aside) is applied and then you just follow the deploy status again. **Only deploy again when you changed the source or a variable the build reads** (`RAILPACK_*`, framework build-time values). A build that fetches content which changes without a commit (a static site pulling from a CMS) needs a deploy that skips the build cache, or it reuses the old result.

**Build FAILED** — read the build logs:
- Version mismatch → pin `RAILPACK_<PROVIDER>_VERSION`.
- Missing system package → add `RAILPACK_BUILD_APT_PACKAGES` / `RAILPACK_DEPLOY_APT_PACKAGES` or `RAILPACK_PACKAGES` (see `references/railpack-config.md`).
- Wrong build command → set `RAILPACK_BUILD_CMD`.
- Killed while building (exit code 137, or `buildkit solve failed … EOF`) → the build ran out of memory, often while prerendering; the builder prints a hint line. There is no setting for build memory: reduce what the build does at once (fewer prerendered pages, lower parallelism) or report it to the user.
Then deploy again.

**Deploy FAILED** — read the deploy logs: bad config applied during rollout; read the message and fix the offending variable/setting.

**Rollout failure reasons** (from the deploy status):
- `OOM_KILLED` → double the memory in the service settings (up to the per-service ceiling). It rolls out automatically with the current image; just re-check status. No rebuild.
- `CRASH_LOOP` → read the runtime logs (they include instances that already crashed or were replaced, so there is no need to swap the start command for `sleep` to look around). Usually wrong start command, `PORT` not honored, or a missing env var. Fix the start command in the service settings or set the variable — the change rolls out automatically with the current image; re-check status. No deploy/rebuild unless you changed source.
- `IMAGE_PULL_FAILED` → the image ref is wrong or unreachable (prebuilt-image deploys); fix the ref.
- `CONFIG_ERROR` → an invalid variable or setting; correct it.
- `UNSCHEDULABLE` → no node has room for the requested resources; lower the request or the user adds capacity.
- `QUOTA_EXCEEDED` → the workspace quota is hit; the user raises it in the dashboard (you cannot).
- `NOT_READY` / `DEADLINE_EXCEEDED` → the app started but never became healthy in time; check runtime logs for a slow or failing boot.
- An instance that is stuck while code and config are fine → restart the service (fresh instances, same image, no build). Don't set a dummy variable to force a restart.

**NEVER silently modify the user's code.** If the fix requires a code change (e.g. the app truly
ignores `$PORT`), propose the change, explain why, and let the user apply it. Configuration lives on
the platform; code lives in the repo.

## 7. Verify & report

- Add a generated domain to the web service. Until the service's first rollout has succeeded, the
  reply lists only the internal endpoint; that is not a failure. The public domain appears in the
  deploy status endpoints once the service is up, so read the URL from there.
- `curl` the endpoint until it returns 200 (TLS provisioning can take ~1 minute; retry).
- Report back:
  - Live URL(s).
  - Services created and their start commands.
  - Resources provisioned (databases, KV stores, buckets, volumes).
  - **Every configuration decision and why** (version pins, custom commands, memory bumps).
  - Any secret-scan findings the platform reported (read from the deploy status).
  - Remaining manual steps for the user.
- If a remediation needed knowledge that was missing from `references/`, note at the end which reference file deserves a new line. The skill improves by contribution.

## 8. Ask vs. decide

- **Decide** anything evidenced in the repo (provider, versions, start command, which resources the code needs, what to pin) — and report it.
- **Ask** only business questions: which repos form the product, whether it should be public, whether you may import a given data dump, whether you may create paid resources, and **any credentials for external managed services the platform does not provide** (Azure/GCS/third-party APIs, which the user adds in the dashboard, see §5) plus **any data dump the app needs that is not in the repo**. If a required external credential or dump is missing, stop and ask rather than shipping a broken deploy.

## Tools

The server's tool list and descriptions are the reference for names, arguments and limits. If a step
in this skill has no matching tool, the user's CLI is older than this skill: do what the available
tools allow, and tell the user to run `lucity update` and start a new session.

There are no delete tools — the user removes projects, services, and resources from the dashboard.

## References

- `references/railpack-config.md` — global Railpack config (build/install/start/packages, `railpack.json`, Procfile priority).
- `references/lucity-platform.md` — PORT injection (8080 default), resource-quantity format, ID formats, variable refs, public DB access, rollout failure reasons, release queue, generated domains, internal DNS shape, volume rules (10Gi floor), custom domains (apex redirect vs ALIAS).
- `references/<provider>.md` — per-language detection, version resolution, start command, config variables, and common failure fixes. One file per Railpack provider.
