ARG GO_VERSION=1.26

FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine AS builder

WORKDIR /workspace
ENV GOWORK=off

COPY pkg/go.mod pkg/go.sum pkg/
COPY services/switchboard/go.mod services/switchboard/go.sum services/switchboard/
RUN --mount=type=cache,target=/go/pkg/mod \
    cd services/switchboard && go mod download

COPY pkg/ pkg/
COPY services/switchboard/ services/switchboard/
ARG TARGETOS TARGETARCH
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    cd services/switchboard && \
    GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/switchboard ./cmd/switchboard

FROM --platform=$BUILDPLATFORM alpine:3.24 AS tools

RUN apk add --no-cache ca-certificates curl git tar
ARG TARGETARCH
ARG KUBECTL_VERSION=v1.35.9
ARG KUBERNETES_MCP_VERSION=v0.0.67
ARG VICTORIALOGS_MCP_VERSION=v1.9.0
ARG VICTORIAMETRICS_MCP_VERSION=v1.20.2
RUN ARCHIVE_ARCH=$([ "$TARGETARCH" = "arm64" ] && echo arm64 || echo x86_64) && \
    mkdir -p /out && \
    curl -fsSLo /out/kubectl "https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/${TARGETARCH}/kubectl" && \
    curl -fsSLo /out/kubernetes-mcp-server "https://github.com/containers/kubernetes-mcp-server/releases/download/${KUBERNETES_MCP_VERSION}/kubernetes-mcp-server-linux-${TARGETARCH}" && \
    curl -fsSL "https://github.com/VictoriaMetrics/mcp-victorialogs/releases/download/${VICTORIALOGS_MCP_VERSION}/mcp-victorialogs_Linux_${ARCHIVE_ARCH}.tar.gz" | tar -xz -C /out mcp-victorialogs && \
    curl -fsSL "https://github.com/VictoriaMetrics/mcp-victoriametrics/releases/download/${VICTORIAMETRICS_MCP_VERSION}/mcp-victoriametrics_Linux_${ARCHIVE_ARCH}.tar.gz" | tar -xz -C /out mcp-victoriametrics && \
    chmod 0755 /out/*
RUN git clone --depth 1 https://github.com/zeitlos/lucity.git /lucity && rm -rf /lucity/.git

FROM --platform=$BUILDPLATFORM node:24-bookworm-slim AS agent

ARG TARGETARCH
ARG CLAUDE_AGENT_ACP_VERSION=0.85.0
RUN NPM_CPU=$([ "$TARGETARCH" = "arm64" ] && echo arm64 || echo x64) && \
    npm install --global --prefix /opt/agent --os=linux --cpu="$NPM_CPU" --libc=glibc \
      @agentclientprotocol/claude-agent-acp@${CLAUDE_AGENT_ACP_VERSION} && \
    npm cache clean --force

FROM node:24-bookworm-slim

COPY --from=tools /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=tools /out/ /usr/local/bin/
COPY --from=tools /lucity /opt/lucity
COPY --from=agent /opt/agent /opt/agent
COPY --from=builder /out/switchboard /usr/local/bin/switchboard
ENV PATH=/opt/agent/bin:$PATH

USER node
WORKDIR /workspace
ENTRYPOINT ["switchboard"]
