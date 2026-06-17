ARG GO_VERSION=1.26

FROM golang:${GO_VERSION}-alpine AS builder

ARG VERSION=dev

WORKDIR /workspace
ENV GOWORK=off

COPY pkg/go.mod pkg/go.sum pkg/
COPY charts/go.mod charts/
COPY services/conductor/go.mod services/conductor/go.sum services/conductor/
RUN cd services/conductor && go mod download

COPY pkg/ pkg/
COPY charts/ charts/
COPY services/conductor/ services/conductor/
RUN cd services/conductor && \
    CGO_ENABLED=0 go build -ldflags="-s -w -X main.Version=${VERSION}" -o /app ./cmd/conductor/...

ARG TARGETARCH=amd64
RUN set -eux; \
    miseVersion="$(cat "$(go env GOMODCACHE)"/github.com/railwayapp/railpack@*/core/mise/version.txt)"; \
    case "${TARGETARCH}" in \
      amd64) miseArch=linux-x64 ;; \
      arm64) miseArch=linux-arm64 ;; \
      *) echo "unsupported arch ${TARGETARCH}" >&2; exit 1 ;; \
    esac; \
    mkdir -p /railpack/mise; \
    wget -qO /tmp/mise.tar.gz "https://github.com/jdx/mise/releases/download/v${miseVersion}/mise-v${miseVersion}-${miseArch}.tar.gz"; \
    tar -xzf /tmp/mise.tar.gz -C /tmp; \
    cp "$(find /tmp -type f -name mise -perm -u+x | head -n1)" "/railpack/mise/mise-${miseVersion}"; \
    chmod 0755 "/railpack/mise/mise-${miseVersion}"

FROM debian:bookworm-slim

RUN set -eux; \
    apt-get update; \
    apt-get install -y --no-install-recommends ca-certificates git curl xz-utils; \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder --chown=65534:65534 /railpack /tmp/railpack
COPY --from=builder /app /app

ENTRYPOINT ["/app"]
