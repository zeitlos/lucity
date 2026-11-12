# Static sites (Railpack)

Static websites served with Caddy, no build step required. Provider value: `staticfile`. (For SPA
frameworks that build via `npm run build` — Vite, Astro, Next static export, etc. — see the Static Sites
section in `node.md`; this file is the plain-files/`Staticfile` provider.)

## Detection

Any of:
- A `Staticfile` in the root
- An `index.html` in the root
- A `public` directory
- `RAILPACK_STATIC_FILE_ROOT` set

## Root directory resolution (highest first)

1. `RAILPACK_STATIC_FILE_ROOT`
2. `root` in `Staticfile`
3. `public` directory if it exists
4. Current directory (`.`) if `index.html` exists in the root

## Staticfile

A `Staticfile` in the root configures serving:

```yaml
root: dist
index_fallback: true
```

`index_fallback` defaults to `false`. Enable it for single-page apps (React, Vue, Angular) so unmatched
routes fall back to `index.html`. Leave it off for multi-page sites so unknown paths return 404.

## Custom Caddyfile

Override the default server config by placing a `Caddyfile` at the repo root.

## Config variables

| Variable | Effect | Example |
| :-- | :-- | :-- |
| `RAILPACK_STATIC_FILE_ROOT` | Override the served root directory | `public` |

## Common failure modes

- **Wrong directory served (blank page / 404)** → set `RAILPACK_STATIC_FILE_ROOT` or `root` in `Staticfile`.
- **SPA deep links 404** → enable `index_fallback: true` in `Staticfile`.
- **Site needs a build step first** → this provider serves prebuilt files; use the Node SPA path (`node.md`) if a build is required.

Docs: https://railpack.com/languages/staticfile
