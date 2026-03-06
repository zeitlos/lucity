ARG GO_VERSION=1.26

FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /workspace
ENV GOWORK=off

# --- Layer 1: Module metadata (changes rarely) ---
COPY pkg/go.mod pkg/go.sum pkg/
COPY charts/go.mod charts/
COPY services/builder/go.mod services/builder/go.sum services/builder/

RUN cd services/builder && go mod download

# --- Layer 2: Source code (changes often) ---
COPY pkg/ pkg/
COPY charts/go.mod charts/
COPY charts/embed.go charts/
COPY charts/lucity-app/ charts/lucity-app/
COPY services/builder/ services/builder/

RUN cd services/builder && \
    CGO_ENABLED=0 go build -ldflags="-s -w" -o /app ./cmd/builder/...

# The builder needs a full runtime because railpack downloads and executes the
# `mise` binary at runtime to resolve language versions. mise is dynamically
# linked against glibc and needs /lib64/ld-linux-x86-64.so.2 + shared libs.
# distroless images lack these. bookworm-slim provides glibc + CA certs (~80MB)
# while staying minimal.
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /app /app

# buildctl is needed by the run-build subcommand (K8s Job pods) to submit
# builds to the BuildKit sidecar. Copied from the official BuildKit image.
COPY --from=moby/buildkit:v0.21 /usr/bin/buildctl /usr/bin/buildctl

ENTRYPOINT ["/app"]
