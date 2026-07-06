# Rust (Railpack)

Rust apps compiled to a release binary.

## Detection

A `Cargo.toml` in the root directory.

## Version resolution (highest first)

1. Mise version files (`mise.toml`, `.tool-versions`)
2. `toolchain.channel` in `rust-toolchain.toml`
3. `package.rust-version` in `Cargo.toml`
4. `.rust-version` or `rust-version.txt`
5. `RAILPACK_RUST_VERSION` env var
6. `package.edition` in `Cargo.toml`
7. Default `1.89`

## Build behavior

Installs Rust + system deps, installs crates, compiles to a binary. Caches `~/.cargo/registry`
(`cargo_registry`), `~/.cargo/git` (`cargo_git`), and `target` (`cargo_target`).

## Start command

Runs `./bin/<project-name>` (the compiled binary). No shell command to derive. For a workspace with
multiple binaries or a non-default name, set the start command via `configure_service`.

## Config variables

| Variable | Effect | Example |
| :-- | :-- | :-- |
| `RAILPACK_RUST_VERSION` | Rust version | `1.85.1` |

Runtime default set: `ROCKET_ADDRESS=0.0.0.0` (so Rocket binds all interfaces).

## Common failure modes

- **Wrong Rust version** → pin `RAILPACK_RUST_VERSION` or add `rust-toolchain.toml`.
- **Missing system library at build (e.g. OpenSSL)** → add `RAILPACK_BUILD_APT_PACKAGES` (e.g. `libssl-dev pkg-config`).
- **Binary doesn't bind the injected port** → read `PORT` and bind `0.0.0.0:$PORT` (code fix for most frameworks; Rocket address is preset).
- **Wrong binary in a workspace** → set the start command to the correct `./bin/<name>`.

Docs: https://railpack.com/languages/rust
