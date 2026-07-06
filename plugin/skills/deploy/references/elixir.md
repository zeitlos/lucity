# Elixir (Railpack)

Elixir and Phoenix apps built into an OTP release.

## Detection

A `mix.exs` in the root directory.

## Version resolution (highest first)

Elixir:
1. Mise version files (`mise.toml`, `.tool-versions`)
2. `.elixir-version`
3. `mix.exs`
4. `RAILPACK_ELIXIR_VERSION` env var
5. Default `1.18`

Erlang/OTP:
1. Mise version files
2. `.erlang-version`
3. Auto from the resolved Elixir version
4. `RAILPACK_ERLANG_VERSION` env var
5. Default `27.3`

## Build behavior

Installs Elixir + Erlang, fetches prod deps (`mix deps.get --only prod`, `mix deps.compile`). If defined,
runs `mix assets.setup`, `mix assets.deploy`, and `mix ecto.deploy`. Compiles a release with
`mix compile` and `mix release`.

## Start command

Runs the release binary: `/app/_build/prod/rel/{app}/bin/{app} start`. No shell command to derive. For a
custom release name, set the start command via `configure_service`.

## Config variables

| Variable | Effect | Example |
| :-- | :-- | :-- |
| `RAILPACK_ELIXIR_VERSION` | Elixir version | `1.18` |
| `RAILPACK_ERLANG_VERSION` | Erlang/OTP version | `27.3` |

## Common failure modes

- **Phoenix doesn't bind the injected port** → set `PORT`-aware config in `config/runtime.exs` (e.g. `http: [ip: {0,0,0,0}, port: String.to_integer(System.get_env("PORT") || "4000")]`) — code fix, propose it.
- **`SECRET_KEY_BASE` / release env missing** → set required runtime vars with `set_variables`.
- **Asset deploy fails** → ensure `assets.deploy` is defined and its Node/esbuild deps are available.
- **Elixir/OTP version mismatch** → pin `RAILPACK_ELIXIR_VERSION` and/or `RAILPACK_ERLANG_VERSION`.

Docs: https://railpack.com/languages/elixir
