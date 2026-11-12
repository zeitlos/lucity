# Gleam (Railpack)

Gleam apps built as an Erlang shipment.

## Detection

A `gleam.toml` in the root directory.

## Versions

Gleam and Erlang both default to the latest version. Erlang is present in both build and runtime images;
Gleam is only present during the build. Override with a mise version file (`mise.toml`, `.tool-versions`)
or `RAILPACK_PACKAGES`.

## Build behavior

Builds with `gleam export erlang-shipment` and runs `./build/erlang-shipment/entrypoint.sh run`. The
source tree is not included in the final container by default.

## Config variables

| Variable | Effect | Example |
| :-- | :-- | :-- |
| `RAILPACK_GLEAM_INCLUDE_SOURCE` | If truthy, include the source tree in the final container | `1` |
| `RAILPACK_PACKAGES` | Pin Gleam/Erlang via Mise (e.g. `gleam@1.4 erlang@27`) | `gleam@1.4` |

## Common failure modes

- **Need a specific Gleam or Erlang version** → add a `mise.toml` or set `RAILPACK_PACKAGES`.
- **Runtime needs source files** → set `RAILPACK_GLEAM_INCLUDE_SOURCE` to a truthy value.
- **App doesn't bind the injected port** → read `PORT` in the Gleam app and bind `0.0.0.0` (code fix; propose it).

Docs: https://railpack.com/languages/gleam
