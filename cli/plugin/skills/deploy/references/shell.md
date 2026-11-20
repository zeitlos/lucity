# Shell scripts (Railpack)

Apps whose entry point is a shell script. Provider value: `shell`.

## Detection

Any of:
- A `start.sh` in the root directory
- `RAILPACK_SHELL_SCRIPT` set to a valid script file

## Interpreter detection

Railpack reads the script's shebang to choose the interpreter: `bash`, `sh`, `dash`, `zsh` (zsh is auto-
installed; the others are in the base image). No shebang → `sh`. Non-POSIX shells (`fish`, `mksh`, `ksh`)
cannot be auto-detected and fall back to `bash`; install and invoke them yourself if required.

## Build behavior

Optionally runs `RAILPACK_INSTALL_CMD` before the build to prepare the environment (create config,
download artifacts, run a setup script). Files created there are available in the build and the final
image.

## Config variables

| Variable | Effect | Example |
| :-- | :-- | :-- |
| `RAILPACK_SHELL_SCRIPT` | Script to execute | `deploy.sh` |
| `RAILPACK_INSTALL_CMD` | Custom install/setup step | `bash setup.sh` |
| `RAILPACK_PACKAGES` | Mise tools to install | `jq@latest python@3.11` |
| `RAILPACK_BUILD_APT_PACKAGES` | Apt packages at build | `git` |
| `RAILPACK_DEPLOY_APT_PACKAGES` | Apt packages at runtime | `ffmpeg` |

## Common failure modes

- **Wrong interpreter** → add an explicit shebang (e.g. `#!/bin/bash`) to the script.
- **Missing tool at runtime** → add it via `RAILPACK_PACKAGES` (Mise) or `RAILPACK_DEPLOY_APT_PACKAGES` (Apt).
- **Script isn't detected** → name it `start.sh` or set `RAILPACK_SHELL_SCRIPT`.
- **Long-running server doesn't bind the injected port** → have the script's process bind `0.0.0.0:$PORT`.

Docs: https://railpack.com/languages/shell
