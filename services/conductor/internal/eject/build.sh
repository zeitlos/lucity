#!/usr/bin/env bash
set -euo pipefail

railpack_version=@RAILPACK_VERSION@
buildkit_image=moby/buildkit:v0.28.0
buildkit_container=lucity-buildkit
platform=${BUILD_PLATFORM:-linux/amd64}
root=$(cd "$(dirname "$0")" && pwd)

usage() {
  cat >&2 <<'EOF'
Usage: ./build.sh <environment> <registry> [service...]

Rebuilds the services of an environment that are built from a repository, the
way Lucity builds them, and pushes each one to <registry>/<service>:<tag>.
values/<environment>.yaml is then updated to use the new images.

  ./build.sh production registry.example.com/shop
  ./build.sh production registry.example.com/shop web worker

Needs git, docker, curl and yq (https://github.com/mikefarah/yq), and a
docker login for the registry. Images are built for linux/amd64, as on Lucity.
Set BUILD_PLATFORM to build for something else, such as linux/arm64.
EOF
  exit 2
}

fail() {
  echo "build.sh: $*" >&2
  exit 1
}

case ${1:-} in
  -h | --help | "") usage ;;
esac

[ $# -ge 2 ] || usage

environment=$1
registry=${2%/}
shift 2

values=$root/values/$environment.yaml
[ -f "$values" ] || fail "no values file for environment '$environment' at $values"

for tool in git docker curl tar yq; do
  command -v "$tool" >/dev/null || fail "$tool is required but not installed"
done

case $(yq --version) in
  *mikefarah*) ;;
  *) fail "yq has to be mikefarah/yq version 4, see https://github.com/mikefarah/yq" ;;
esac

yq -0 -n '"probe"' >/dev/null 2>&1 || fail "this yq is too old, update it to a current version 4"
docker info >/dev/null 2>&1 || fail "docker is not running"

railpack=${XDG_CACHE_HOME:-$HOME/.cache}/lucity/railpack-$railpack_version/railpack
work=$(mktemp -d)
started_buildkit=

cleanup() {
  rm -rf "$work"

  if [ -n "$started_buildkit" ]; then
    docker rm -f "$buildkit_container" >/dev/null 2>&1 || true
  fi
}

trap cleanup EXIT

sha256() {
  if command -v sha256sum >/dev/null; then
    sha256sum "$1" | awk '{print $1}'
  else
    shasum -a 256 "$1" | awk '{print $1}'
  fi
}

install_railpack() {
  local target archive url expected

  case "$(uname -s)-$(uname -m)" in
    Darwin-arm64) target=arm64-apple-darwin ;;
    Darwin-x86_64) target=x86_64-apple-darwin ;;
    Linux-aarch64 | Linux-arm64) target=arm64-unknown-linux-musl ;;
    Linux-x86_64) target=x86_64-unknown-linux-musl ;;
    *) fail "Railpack has no release for $(uname -s) $(uname -m)" ;;
  esac

  archive=railpack-v$railpack_version-$target.tar.gz
  url=https://github.com/railwayapp/railpack/releases/download/v$railpack_version

  echo "Downloading Railpack $railpack_version"
  curl -fsSL -o "$work/$archive" "$url/$archive"
  curl -fsSL -o "$work/checksums.txt" "$url/checksums.txt"

  expected=$(awk -v name="$archive" '$2 == name {print $1}' "$work/checksums.txt")

  if [ -z "$expected" ] || [ "$expected" != "$(sha256 "$work/$archive")" ]; then
    fail "checksum mismatch for $archive"
  fi

  tar -xzf "$work/$archive" -C "$work" railpack
  mkdir -p "$(dirname "$railpack")"
  mv "$work/railpack" "$railpack"
}

start_buildkit() {
  if [ -n "${BUILDKIT_HOST:-}" ]; then
    return
  fi

  export BUILDKIT_HOST=docker-container://$buildkit_container

  if [ -n "$(docker ps --quiet --filter "name=^$buildkit_container\$")" ]; then
    return
  fi

  echo "Starting BuildKit"
  docker run --detach --rm --privileged --name "$buildkit_container" "$buildkit_image" >/dev/null
  started_buildkit=1

  for _ in $(seq 30); do
    if docker exec "$buildkit_container" buildctl debug workers >/dev/null 2>&1; then
      return
    fi

    sleep 1
  done

  fail "BuildKit did not come up"
}

build() {
  local service=$1 repository context tag clone image key value
  local variables=() flags=()

  repository=$(SERVICE=$service yq '.services[strenv(SERVICE)].annotations["lucity.dev/source-repo"] // ""' "$values")
  context=$(SERVICE=$service yq '.services[strenv(SERVICE)].annotations["lucity.dev/source-context"] // ""' "$values")
  tag=$(SERVICE=$service yq '.services[strenv(SERVICE)].image.tag // ""' "$values")

  [ -n "$repository" ] || fail "$service is not built from a repository, so there is nothing to rebuild"
  [[ $tag =~ ^[0-9a-f]{7,40}$ ]] || fail "$service has no commit to build, its image tag is '$tag'"

  image=$registry/$service:$tag
  clone=$work/$service

  echo
  echo "Building $service from $repository at $tag"
  git clone --quiet --filter=blob:none --no-checkout -- "$repository" "$clone"
  git -C "$clone" checkout --quiet "$tag"

  SERVICE=$service yq -0 '.services[strenv(SERVICE)].env // {} | to_entries | map([.key, .value]) | flatten | .[]' "$values" >"$work/$service.env"

  while IFS= read -r -d '' key && IFS= read -r -d '' value; do
    case $key in
      PATH | HOME | TMPDIR | BUILDKIT_HOST | DOCKER_* | LD_* | DYLD_*) flags+=(--env "$key=$value") ;;
      *)
        variables+=("$key=$value")
        flags+=(--env "$key")
        ;;
    esac
  done <"$work/$service.env"

  env ${variables[@]+"${variables[@]}"} "$railpack" build \
    --name "$image" --platform "$platform" ${flags[@]+"${flags[@]}"} "$clone/${context#/}"

  docker push "$image"

  SERVICE=$service REPOSITORY=$registry/$service yq -i '
    .services[strenv(SERVICE)].image.repository = strenv(REPOSITORY) |
    del(.services[strenv(SERVICE)].image.digest)
  ' "$values"

  echo "Pushed $image and pointed $service at it in values/$environment.yaml"
}

services=("$@")

if [ ${#services[@]} -eq 0 ]; then
  names=$(yq '.services // {} | to_entries | .[] | select(.value.annotations["lucity.dev/source-repo"]) | .key' "$values")

  while IFS= read -r name; do
    if [ -n "$name" ]; then
      services+=("$name")
    fi
  done <<<"$names"
fi

if [ ${#services[@]} -eq 0 ]; then
  echo "Every service in $environment runs a prebuilt image, so there is nothing to rebuild."
  exit 0
fi

[ -x "$railpack" ] || install_railpack
start_buildkit

for service in "${services[@]}"; do
  build "$service"
done
