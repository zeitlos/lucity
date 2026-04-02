ARG GO_VERSION=1.26

FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /workspace
ENV GOWORK=off

# --- Layer 1: Module metadata (changes rarely) ---
COPY pkg/go.mod pkg/go.sum pkg/
COPY charts/go.mod charts/
COPY services/packager/go.mod services/packager/go.sum services/packager/

RUN cd services/packager && go mod download

# --- Layer 2: Source code (changes often) ---
COPY pkg/ pkg/
COPY charts/go.mod charts/
COPY charts/embed.go charts/
COPY charts/lucity-app/ charts/lucity-app/
COPY services/packager/ services/packager/

RUN cd services/packager && \
    CGO_ENABLED=0 go build -ldflags="-s -w" -o /app ./cmd/packager/...

FROM gcr.io/distroless/static-debian12

COPY --from=builder /app /app

ENTRYPOINT ["/app"]
