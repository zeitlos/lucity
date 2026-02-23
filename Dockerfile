ARG GO_VERSION=1.26

FROM golang:${GO_VERSION}-alpine AS builder

ARG SERVICE

WORKDIR /workspace
ENV GOWORK=off

# --- Layer 1: Module metadata (changes rarely) ---
# Copy go.mod/go.sum from all workspace modules so `go mod download`
# can resolve the full dependency graph via replace directives.
COPY pkg/go.mod pkg/go.sum pkg/
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
