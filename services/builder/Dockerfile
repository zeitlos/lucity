FROM --platform=$BUILDPLATFORM golang:1.27-alpine AS builder

WORKDIR /workspace
ENV GOWORK=off

COPY pkg/go.mod pkg/go.sum pkg/
COPY services/builder/go.mod services/builder/go.sum services/builder/
RUN --mount=type=cache,target=/go/pkg/mod \
    cd services/builder && go mod download

COPY pkg/ pkg/
COPY services/builder/ services/builder/
ARG VERSION=dev
ARG TARGETOS TARGETARCH
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    cd services/builder && \
    GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go build -ldflags="-s -w -X main.Version=${VERSION}" -o /app ./cmd/builder/...

# debian-slim (not distroless) because the build runner shells out
# to git for cloning user source repos and the BuildKit client embeds
# CGO-friendly code paths in some transitive deps.
FROM ghcr.io/gitleaks/gitleaks:v8.30.1 AS gitleaks
FROM trufflesecurity/trufflehog:3.100.1@sha256:9c7a8325d43a4af2960d3f21d61a08981d8a1bf0542e007efa2e47097e386452 AS trufflehog

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates git && \
    rm -rf /var/lib/apt/lists/*

COPY --from=gitleaks /usr/bin/gitleaks /usr/local/bin/gitleaks
COPY --from=trufflehog /usr/bin/trufflehog /usr/local/bin/trufflehog
COPY --from=builder /app /app

ENTRYPOINT ["/app"]
