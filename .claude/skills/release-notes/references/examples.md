# Example releases (house style reference)

These are two real, published Lucity releases. They are the ground truth for tone,
structure, grouping, and level of detail. Match them. Note how each opens with the
social-preview image, then a short personal paragraph, then grouped user-facing bullets
with `(#NN)` PR references, and how open-source building blocks are credited with links
while the underlying cloud/infra vendor is never named.

---

## v26.6.2 (2026-06-29) — theme: "fire new features" during a heatwave

<img width="1280" height="640" alt="lucity-social-preview" src="https://github.com/user-attachments/assets/e94c8584-1588-45fe-bfef-b1949a3c4947" />


A heatwave in central europe gave me the perfect excuse to stay in my basement office and work on this new Lucity release.

This release brings some fire new features, such as volumes, real-time resource metrics, one-click object storage, highly-available PostgreSQL, and automatic credential discovery. Additionally variable handling gets a security upgrade with concealment by default, and rollouts are now graceful with near-zero downtime.

Variables

- Variables are now concealed by default and only revealed on demand. (#36)
- Variables auto-discovered from Redis and object storage are tagged with their source, so there is no ambiguity about where a value came from. (#31)
- Paste a `.env`-formatted block and the editor splits it into individual variables automatically. (#31)
- Variables are sorted alphabetically. (#31)
- Services automatically re-deploy when shared variables change. (#31)
- A note in the UI makes it clear that variables are available both at build time and at runtime.

Persistent Volumes

- Create, mount, unmount, expand, and delete persistent volumes from the dashboard. (#34)
- A slider controls the requested size; resize (expand) works in place without data migration. (#34)
- Live usage percentage is shown as a background indicator on the volume node in the canvas. (#34)

Resource Metrics

- CPU and memory usage are surfaced in real time in the service panel's Metrics tab. (#35)
- Volume usage percentage is shown in the volume panel alongside capacity. (#34)

Object Storage

- Provision an S3-compatible bucket in one click. No CLI, no cloud console, no separate credential step. (#29)
- Bucket credentials are automatically discovered and available as shared environment variables across the environment. (#31)

PostgreSQL

- Databases are now highly available by default: replicated, multi-node, tolerant of a single-node failure. (#32)
- CPU, memory, and disk size are configurable from the dashboard settings tab. Disk defaults to 16 GiB on a power-of-2 ladder, up to the configured quota. (#32)
- Scaling mutations now surface database status correctly so the UI stays in sync during resource changes. (#32)

Redis

- Redis instances backed by Valkey are available for provisioning.
- Connection credentials are automatically discovered and available as shared environment variables. (#31)

Deployments and Rollouts

- Rollouts are graceful by default: a readiness probe gates traffic until the new replica is healthy, and a preStop hook drains in-flight requests before the old pod terminates. Near-zero downtime on every deploy. (#28)

Eject

- The eject command is now available again. (#30)

Landing Page

- Redesigned landing page with new visual identity, dark mode support, and a more personal, direct tone. (#33)

---

## v26.6.1 (2026-06-19) — theme: control-plane rewrite (internal change, framed by user impact)

<img width="1280" height="640" alt="lucity-social-preview" src="https://github.com/user-attachments/assets/1a69a16f-18b0-46fe-8c99-8753b35344dc" />


Since 26.4.1 Lucity's control plane has been completely re-written. You won't directly see this change, but you might feel it in the stability of the platform and the cadence of new features. The whole control plane is essentially one binary now, called conductor.

Previously the control plane consisted of a pile of services that were vibe-coded together. Soon after the first release, it became clear that this architecture will not scale and the code is a nightmare to maintain.

On top of that foundation, this release brings security improvement, experimental support for Redis, exposing Postgres databases to the internet and build time variables. Here's a list of the complete highlights:

Deployments
- Immutable digest-based deploys. Every deployment now pins the exact image digest, so a deploy is fully reproducible and promotion between environments copies the digest instead of rebuilding.
- Self-healing releases. The platform continuously reconciles your deployments and corrects drift on its own.
- Avoiding GitHub rate limits. The deployed commit (message + SHA) is shown in the dashboard, stamped at deploy time so it stays fast and doesn't hit GitHub rate limits.
- Custom start commands are now handled properly and validated before they run.
- Variables now power builds and runtime. Your project variables are injected at build time as well as into the running container, with the dashboard making that clear.

New: managed Redis
- Add a Redis key-value store (powered by [Valkey](https://valkey.io/)) to any project, right from the dashboard. This is a first preview, so expect it to grow.

Databases
- Encrypted internal connections. Cluster-internal database traffic now uses TLS.
- Switched to the standard PostgreSQL image flavor (ships with pgvector) and bumped default memory requests/limits for more headroom.

Security & isolation
- Hardened build jobs: mutual TLS and user namespaces for the build runner, automatic certificate rotation, and link-local addresses blocked from the build network policy.
- Least-privilege workloads: service account tokens are no longer auto-mounted into your containers, and unused build-job service accounts/roles were removed.
- Tighter GitHub access: tokens are now scoped to just the repo being read, the installation ID is gone from the API surface, and build IDs are workspace-validated.
- Repository names containing dots or underscores are now sanitized correctly (#27).

Reliability & operations
- New observability stack built on VictoriaMetrics and Grafana (migrated off Signoz/ClickHouse), with host and [Hubble](https://github.com/cilium/hubble) network metrics, HTTP gateway alert rules, and Telegram alerting.
- Global hostname uniqueness checks for custom domains.
- Real client IPs preserved end to end via PROXY protocol.
- Fixed a graceful-shutdown bug and a billing bug that could suspend a workspace too early.
- Fixed the dashboard showing version dev in production.
- Empty projects now show their service and repo counts in the dashboard.

Platform images & build
- ARM builds for tagged releases (multi-arch images).
- Per-service Dockerfiles on a slim Debian base with [mise](https://mise.jdx.dev/) preinstalled, for smaller, faster images.
- Updated [Zot](https://zotregistry.dev/) (now with UI and CVE scanning), [CloudNativePG](https://cloudnative-pg.io/), and Argo CD charts.

Docs & site
- Added a blog to the docs site, included blog posts in the sitemap, sharper landing-page copy and CTAs, and removed an outdated quick-start guide.

---

## What to notice

- **Opening image, then a blank line, then prose.** The `<img>` tag is literally the first line.
- **The intro is a person talking, not a product changelog.** It sets a scene ("a heatwave", "my basement office"), or gives honest context ("a pile of services that were vibe-coded together"), then names the 3-6 headline items in one or two sentences. The skill does **not** write this paragraph; it leaves a notes placeholder and the maintainer writes it in their own voice. These examples are the target that placeholder is aiming at.
- **Group headings are plain text lines**, not markdown `##` headers. Order groups by how much a user will care (new capabilities first, polish and ops later).
- **Bullets are use-case oriented.** "Paste a `.env`-formatted block and the editor splits it" beats "added .env parsing". Say what the user can now do and why it helps.
- **`(#NN)` PR references are appended where a PR exists**, not forced onto every line.
- **Open-source building blocks get credited with links** (Valkey, Hubble, Zot, mise, CloudNativePG). The backing cloud/infra vendor is never named.
- **Internal-only work is omitted** unless it changes the day-to-day experience. The whole v26.6.1 rewrite paragraph is the exception that proves the rule: it's internal, but it's framed entirely in terms of what the user feels (stability, feature cadence).
- **No em dashes.** Periods, commas, colons, semicolons, or parentheses instead.
