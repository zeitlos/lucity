# Bun (Railpack)

Bun is handled by the **Node provider** — there is no separate `bun` provider value. Force-provider is
`node` if you ever need it. See `node.md` for the full Node behavior; this file covers the Bun specifics.

## Detection

A `package.json` in the root. Bun is selected as the package manager when a `bun.lockb` or `bun.lock`
lock file is present (or via `packageManager` / `engines.bun`).

## Version resolution (highest first)

1. `RAILPACK_BUN_VERSION` env var
2. `.bun-version` file
3. `engines.bun` in `package.json`
4. `mise.toml` / `.tool-versions`
5. Default `latest`

## Node alongside Bun

Even when Bun is the package manager, Railpack also installs Node.js in these cases:

- A `packageManager` field is set (Corepack support).
- Any `package.json` script contains `node`.
- The app uses Astro or Vite.
- Native modules need compiling during install (`node-gyp`).

When Node is needed only during install (native modules), it is installed via Mise and respects the Node
version resolution order (see `node.md`).

## Build & start

Same as the Node provider: install → optional `build` script → start command from `start` script,
`main`, or root `index.*`. Static-site (SPA) detection applies too.

## Config variables

| Variable | Effect | Example |
| :-- | :-- | :-- |
| `RAILPACK_BUN_VERSION` | Bun version | `1.2` |
| `RAILPACK_NODE_VERSION` | Node version (when Node is also installed) | `22` |

Plus all Node config variables (`RAILPACK_NO_SPA`, `RAILPACK_SPA_OUTPUT_DIR`, ...); see `node.md`.

## Common failure modes

- **Native module build fails** → Node is auto-installed for `node-gyp`; pin `RAILPACK_NODE_VERSION` if the wrong Node version is picked.
- **No start command** → add a `start` script or set the start command via `configure_service`.
- **Bun version too new (blocked by Mise 14-day rule)** → pin an older `RAILPACK_BUN_VERSION` or adjust `minimum_release_age` in `mise.toml`.

Docs: https://railpack.com/languages/node
