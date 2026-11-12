# Go (Railpack)

Go apps compiled to a static binary. Provider value: `golang`.

## Detection

Any of:
- `go.mod` in the root
- `go.work` in the root (Go workspaces)
- `main.go` in the root

## Version resolution (highest first)

1. Mise version files (`mise.toml`, `.tool-versions`)
2. `go.mod`
3. `RAILPACK_GO_VERSION` env var
4. Default `1.23`

## Build behavior

Installs dependencies and builds a static binary with `-ldflags="-w -s"`, output named `out`. Caches
`~/.cache/go-build`. The main package to build is chosen in order:

1. `RAILPACK_GO_WORKSPACE_MODULE` (workspaces)
2. `RAILPACK_GO_BIN` (a command under `cmd/`)
3. The root directory if it has Go files
4. The first subdirectory in `cmd/`
5. For workspaces, the first module with a `main.go`
6. `main.go` in the root

## Start command

Runs the compiled `out` binary. There is no shell start command to derive — the binary is the entry
point. If the wrong package is built, correct it with `RAILPACK_GO_BIN` or `RAILPACK_GO_WORKSPACE_MODULE`.

## CGO

Static build with `CGO_ENABLED=0` by default. For CGO, set `CGO_ENABLED=1`: Railpack adds `gcc`, `g++`,
`libc6-dev` at build and `libc6` at runtime.

## Go workspaces

With a `go.work` file, Railpack discovers and copies all module dependencies and builds the first module
with a `main.go`. Select a specific module with `RAILPACK_GO_WORKSPACE_MODULE`.

## Config variables

| Variable | Effect | Example |
| :-- | :-- | :-- |
| `RAILPACK_GO_VERSION` | Go version | `1.22` |
| `RAILPACK_GO_BIN` | Which `cmd/` command to build | `server` |
| `RAILPACK_GO_WORKSPACE_MODULE` | Which workspace module to build | `api` |
| `CGO_ENABLED` | Enable CGO (non-static binary) | `1` |

## Common failure modes

- **Multiple `cmd/` binaries, wrong one built** → set `RAILPACK_GO_BIN`.
- **Workspace builds the wrong module** → set `RAILPACK_GO_WORKSPACE_MODULE`.
- **Binary crashes with a dynamic-linking error** → the app needs CGO; set `CGO_ENABLED=1`.
- **App doesn't listen on the injected port** → read `PORT` in code and bind `0.0.0.0:$PORT` (code fix; propose it).

Docs: https://railpack.com/languages/golang
