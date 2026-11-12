# C/C++ (Railpack)

C/C++ apps built with CMake or Meson. Provider value: `cpp`.

## Detection

A `CMakeLists.txt` or a `meson.build` in the root directory.

## Versions

The latest versions of CMake (or Meson) and Ninja are installed during the build. Pin specific tool
versions with a `mise.toml` or `RAILPACK_PACKAGES` if needed.

## Build behavior

Builds into `/build` and runs the executable in that directory **whose name matches the name of the
project's root directory**. The source tree is not included in the final container by default — only the
build directory.

## Start command

The matching executable in `/build`. If your binary's name differs from the root directory name, or you
have multiple targets, set the start command via `configure_service` to point at the correct binary.

## Config

Use the generic Railpack knobs (see `railpack-config.md`):
- `RAILPACK_PACKAGES` to pin CMake/Meson/Ninja versions or add tools.
- `RAILPACK_BUILD_APT_PACKAGES` for system libraries needed at build.
- `RAILPACK_DEPLOY_APT_PACKAGES` for shared libraries needed at runtime.

## Common failure modes

- **No executable found / wrong binary run** → ensure the target name matches the root directory name, or set the start command explicitly.
- **Missing dev library at build** → add `RAILPACK_BUILD_APT_PACKAGES`.
- **Missing shared library at runtime** → add `RAILPACK_DEPLOY_APT_PACKAGES`.
- **App doesn't bind the injected port** → read `PORT` and bind `0.0.0.0:$PORT` (code fix; propose it).

Docs: https://railpack.com/languages/cpp
