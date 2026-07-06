# Node.js (Railpack)

Node.js apps with npm, yarn, pnpm, or bun, plus first-class static-site support. For Bun as the primary
runtime see `bun.md`.

## Detection

A `package.json` in the root directory.

## Version resolution (highest first)

1. `RAILPACK_NODE_VERSION` env var
2. `engines.node` in `package.json`
3. `.nvmrc`
4. `.node-version`
5. `mise.toml` / `.tool-versions`
6. Default `22`

Only actively maintained LTS versions are officially supported. Bun version resolves via
`RAILPACK_BUN_VERSION` → `.bun-version` → `engines.bun` → mise files → `latest`.

## Package manager detection (highest first)

1. `packageManager` field in `package.json` (installed via Corepack)
2. Lock files: `pnpm-lock.yaml` (pnpm), `bun.lockb`/`bun.lock` (bun), `.yarnrc.yml`/`.yarnrc.yaml` (Yarn Berry), `yarn.lock` (Yarn 1)
3. `engines` field (`engines.pnpm`, `engines.bun`, `engines.yarn`)
4. Default npm

## Build behavior

Installs dependencies, runs the `build` script if present, then serves. Monorepos (npm/yarn/bun
`workspaces`, or `pnpm-workspace.yaml`) work with no config; ensure build/start scripts are in the root
`package.json` or use a config file. Native modules (`node-gyp`) are configured automatically. Puppeteer
and Playwright pull in their system dependencies automatically.

## Start command derivation (highest first)

1. `start` script in `package.json`
2. `main` field in `package.json`
3. `index.js` or `index.ts` in the root

What breaks it: no `start` script, no `main`, and no root `index.*` → no start command. Set one via
`configure_service` (or `RAILPACK_START_CMD`). Nuxt defaults to `node .output/server/index.mjs`.

## Static sites (SPA mode)

Railpack serves a statically built project with Caddy when it detects Vite, Astro, Next.js
(`output: 'export'`), CRA, Angular, React Router (`ssr: false`), or Expo Web. Output dir defaults to
`dist` (`out` for Next static export, `build/client/` for React Router). Force or disable:

- `RAILPACK_SPA_OUTPUT_DIR=<dir>` forces SPA mode and serves that dir.
- `RAILPACK_NO_SPA=1` disables SPA mode (deploy as a Node server instead).

If your Next.js app uses `next start` (SSR), SPA mode stays off — it deploys as a server.

## Config variables

| Variable | Effect | Example |
| :-- | :-- | :-- |
| `RAILPACK_NODE_VERSION` | Node version | `22` |
| `RAILPACK_BUN_VERSION` | Bun version | `1.2` |
| `RAILPACK_NO_SPA` | Disable static-site mode | `true` |
| `RAILPACK_SPA_OUTPUT_DIR` | Static output directory | `dist` |
| `RAILPACK_PRUNE_DEPS` | Remove dev dependencies | `true` |
| `RAILPACK_NODE_PRUNE_CMD` | Custom prune command | `npm prune --omit=dev` |
| `RAILPACK_NODE_INSTALL_PATTERNS` | Extra files to include for install | `prisma` |
| `RAILPACK_ANGULAR_PROJECT` | Angular project to build | `my-app` |

Runtime defaults set: `NODE_ENV=production`, `CI=true`, and several `NPM_CONFIG_*` / `YARN_PRODUCTION`
flags.

## Common failure modes

- **No start command found** → add a `start` script, or set the start command via `configure_service`.
- **SPA served instead of the API server (or vice versa)** → `RAILPACK_NO_SPA=1` to force a server, or `RAILPACK_SPA_OUTPUT_DIR` to force static.
- **Wrong static output directory (blank page / 404)** → set `RAILPACK_SPA_OUTPUT_DIR` to the real build output.
- **Wrong package-manager version** → set `packageManager` in `package.json` or the matching `engines` field.

Docs: https://railpack.com/languages/node
