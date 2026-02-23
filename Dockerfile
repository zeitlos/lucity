ARG GO_VERSION=1.26

FROM golang:${GO_VERSION}-alpine AS builder

ARG SERVICE

WORKDIR /workspace
ENV GOWORK=off

# --- Layer 1: Module metadata (changes rarely) ---
# Copy go.mod/go.sum from all workspace modules so `go mod download`
# can resolve the full dependency graph via replace directives.
COPY pkg/auth/go.mod pkg/auth/go.sum pkg/auth/
COPY pkg/builder/go.mod pkg/builder/go.sum pkg/builder/
COPY pkg/deployer/go.mod pkg/deployer/go.sum pkg/deployer/
COPY pkg/github/go.mod pkg/github/go.sum pkg/github/
COPY pkg/graceful/go.mod pkg/graceful/
COPY pkg/labels/go.mod pkg/labels/
COPY pkg/logger/go.mod pkg/logger/go.sum pkg/logger/
COPY pkg/packager/go.mod pkg/packager/go.sum pkg/packager/
COPY pkg/webhook/go.mod pkg/webhook/go.sum pkg/webhook/
COPY charts/go.mod charts/
COPY services/${SERVICE}/go.mod services/${SERVICE}/go.sum services/${SERVICE}/

RUN cd services/${SERVICE} && go mod download

# --- Layer 2: Source code (changes often) ---
COPY pkg/ pkg/
COPY charts/go.mod charts/
COPY charts/embed.go charts/
COPY charts/lucity-app/ charts/lucity-app/
COPY services/${SERVICE}/ services/${SERVICE}/

RUN cd services/${SERVICE} && \
    CGO_ENABLED=0 go build -ldflags="-s -w" -o /app ./cmd/${SERVICE}/...

FROM gcr.io/distroless/static-debian12

COPY --from=builder /app /app

ENTRYPOINT ["/app"]
