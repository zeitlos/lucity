# Railpack global configuration

Lucity builds with [Railpack](https://railpack.com). Railpack is zero-config: it analyzes the source,
picks a provider, resolves tool versions, builds with BuildKit, and produces a runtime image. When the
defaults are wrong, override them. On Lucity you set these as **service variables** (via `add_service`
initial variables, or `set_variables`); build-time `RAILPACK_*` variables reach the Railpack build as
environment variables, so pin them **before the first build**.

Railpack ignores Dockerfiles entirely — it does not read or generate them. Treat a Dockerfile as
documentation only.

## Build configuration variables

| Variable | Effect |
| :-- | :-- |
| `RAILPACK_BUILD_CMD` | Command for the build step. Overrides any provider-supplied build commands. |
| `RAILPACK_INSTALL_CMD` | Command for the install step. Overrides provider install commands. All files are copied to the project root before it runs. |
| `RAILPACK_START_CMD` | Command the container runs on start. Highest-priority start override. |
| `RAILPACK_PACKAGES` | Additional Mise packages, space-separated, `pkg[@version]` (e.g. `jq@latest python@3.11`). Latest if version omitted. |
| `RAILPACK_BUILD_APT_PACKAGES` | Additional Apt packages during build. Space-separated list. |
| `RAILPACK_DEPLOY_APT_PACKAGES` | Additional Apt packages in the final runtime image. Space-separated list. |
| `RAILPACK_DISABLE_CACHES` | BuildKit cache keys to disable, or `*` for all. Space-separated list. |

Mise is preferred over Apt for language tools and CLIs (`RAILPACK_PACKAGES`). Reach for Apt only for
system libraries Mise cannot provide (e.g. `libpq-dev`, `ffmpeg`).

## Start command priority (highest wins)

1. `RAILPACK_START_CMD` environment variable
2. `deploy.startCommand` in `railpack.json`
3. `Procfile` process type (`web` → `worker` → first other)
4. Provider default

On Lucity, `configure_service` sets the service start command; prefer it over asking the user to add
`RAILPACK_START_CMD` when you just need to change the start command.

## Procfile

Heroku-style. Railpack auto-detects a `Procfile` in the repo root. Format:

```yaml
web: gunicorn --bind 0.0.0.0:$PORT main:app
worker: celery worker -A myapp.celery
```

Priority for the single container start command: `web`, then `worker`, then the first process type
defined. For a web + worker split, deploy the codebase as two services with different start commands.

## railpack.json (config file)

Railpack reads `railpack.json` from the repo root (override path with `RAILPACK_CONFIG_FILE`). It lives
in the user's repo, so prefer platform variables over asking the user to add one. Useful root fields:

| Field | Purpose |
| :-- | :-- |
| `provider` | Force a provider (see list below) instead of autodetection. |
| `packages` | Map of tool → version (e.g. `{ "node": "22", "python": "3.13" }`). |
| `buildAptPackages` | Apt packages during build. |
| `steps` | Custom install/build steps and commands. |
| `deploy.startCommand` | Container start command. |
| `deploy.variables` | Runtime env vars (e.g. `TZ`, `LANG`). |
| `deploy.aptPackages` | Apt packages in the final image. |

Force-provider values (case-insensitive): `node`, `python`, `golang`, `php`, `java`, `ruby`, `rust`,
`elixir`, `deno`, `dotnet`, `gleam`, `cpp`, `staticfile`, `shell`. Bun and `uv` are handled inside the
`node` and `python` providers respectively — there is no separate provider value.

## Version resolution

Providers resolve *fuzzy* versions (e.g. `22`, `3.13`) into a concrete release via Mise. Precedence,
per provider, is: `RAILPACK_<PROVIDER>_VERSION` env var → mise version files → language-specific version
files → provider default. Manifest constraints like `engines.pnpm: ^10.34.0` are simplified to the
major version (`10`) before resolution; pin an exact version or use `mise.toml` if you need precision.

Mise applies a `minimum_release_age` of 14 days by default: versions released in the last two weeks are
skipped. Override in the app's `mise.toml` if you need a brand-new release.

## Locale and timezone

Railpack sets no `TZ`, `LANG`, or `LC_ALL`. Only `en_US.UTF-8` is available in the runtime image. If the
app needs a UTF-8 locale or a fixed timezone, set `LANG=en_US.UTF-8` / `TZ=UTC` as runtime variables.

## .dockerignore

Railpack honors `.dockerignore`. With no `.dockerignore`, the entire repo is sent as build context.
Exclude `node_modules`, `.venv`, `vendor`, `.env`, and secrets to keep builds fast and images clean.

Full reference: https://railpack.com
