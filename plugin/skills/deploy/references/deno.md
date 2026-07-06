# Deno (Railpack)

Deno apps with zero config.

## Detection

A `deno.json` or `deno.jsonc` in the root directory.

## Version resolution (highest first)

1. `RAILPACK_DENO_VERSION` env var
2. Mise version files (`.deno-version`, `.tool-versions`, `mise.toml`)
3. Default `2`

## Build behavior

Installs Deno and caches dependencies with `deno cache`.

## Start command derivation

1. A `main.ts`, `main.js`, `main.mjs`, or `main.mts` in the project root
2. Otherwise, the first `.ts` / `.js` / `.mjs` / `.mts` file found

The selected file runs with `deno run --allow-all`.

What breaks it: no main file and an ambiguous set of scripts → set the start command explicitly via
`configure_service` (e.g. `deno run --allow-net --allow-env server.ts`).

## Config variables

| Variable | Effect | Example |
| :-- | :-- | :-- |
| `RAILPACK_DENO_VERSION` | Deno version | `1.41` |

## Common failure modes

- **Wrong entry file selected** → set the start command via `configure_service`.
- **Permission errors at runtime** → tighten/loosen the `--allow-*` flags in a custom start command (the default is `--allow-all`).
- **App doesn't bind the injected port** → read `Deno.env.get("PORT")` and bind `0.0.0.0` (code fix; propose it).

Docs: https://railpack.com/languages/deno
