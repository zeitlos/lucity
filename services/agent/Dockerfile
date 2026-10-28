ARG GO_VERSION=1.27

FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine AS builder

WORKDIR /workspace
ENV GOWORK=off

COPY pkg/go.mod pkg/go.sum pkg/
COPY services/agent/go.mod services/agent/go.sum services/agent/
RUN --mount=type=cache,target=/go/pkg/mod \
    cd services/agent && go mod download

COPY pkg/ pkg/
COPY services/agent/ services/agent/
ARG TARGETOS TARGETARCH
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    cd services/agent && \
    GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/agent ./cmd/agent

FROM --platform=$BUILDPLATFORM node:24-bookworm-slim AS harnesses

ARG TARGETARCH
WORKDIR /opt/harness
COPY services/agent/harnesses/package.json services/agent/harnesses/package-lock.json ./
RUN NPM_CPU=$([ "$TARGETARCH" = "arm64" ] && echo arm64 || echo x64) && \
    OPENCODE=$([ "$TARGETARCH" = "arm64" ] && echo opencode-linux-arm64 || echo opencode-linux-x64-baseline) && \
    npm ci --omit=dev --ignore-scripts --no-audit --no-fund --os=linux --cpu="$NPM_CPU" --libc=glibc && \
    ln -s "../$OPENCODE/bin/opencode" node_modules/.bin/opencode && \
    npm cache clean --force

FROM node:24-bookworm-slim

RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates curl git ripgrep && \
    rm -rf /var/lib/apt/lists/* && \
    mkdir -p /etc/agent /etc/harness /data && \
    ln -s /etc/harness/claude-code /etc/claude-code && \
    ln -s /etc/harness/codex /etc/codex && \
    ln -s /etc/harness/opencode /etc/opencode && \
    chown 1000:1000 /etc/harness /data

COPY --from=harnesses /opt/harness /opt/harness
COPY --from=builder /out/agent /usr/local/bin/agent
ENV PATH=/opt/harness/node_modules/.bin:$PATH

USER 1000:1000
WORKDIR /data
EXPOSE 8080
ENTRYPOINT ["agent"]
