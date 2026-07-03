ARG GO_VERSION=1.26

FROM golang:${GO_VERSION}-alpine AS builder

ARG VERSION=dev

WORKDIR /workspace
ENV GOWORK=off

COPY pkg/go.mod pkg/go.sum pkg/
COPY services/builder/go.mod services/builder/go.sum services/builder/
RUN cd services/builder && go mod download

COPY pkg/ pkg/
COPY services/builder/ services/builder/
RUN cd services/builder && \
    CGO_ENABLED=0 go build -ldflags="-s -w -X main.Version=${VERSION}" -o /app ./cmd/builder/...

# debian-slim (not distroless) because the build runner shells out
# to git for cloning user source repos and the BuildKit client embeds
# CGO-friendly code paths in some transitive deps.
FROM ghcr.io/gitleaks/gitleaks:latest AS gitleaks
FROM trufflesecurity/trufflehog:latest AS trufflehog

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates git && \
    rm -rf /var/lib/apt/lists/*

COPY --from=gitleaks /usr/bin/gitleaks /usr/local/bin/gitleaks
COPY --from=trufflehog /usr/bin/trufflehog /usr/local/bin/trufflehog
COPY --from=builder /app /app

ENTRYPOINT ["/app"]
