ARG GO_VERSION=1.26

FROM golang:${GO_VERSION}-alpine AS builder

ARG VERSION=dev

WORKDIR /workspace
ENV GOWORK=off

# --- Layer 1: Module metadata (changes rarely) ---
COPY pkg/go.mod pkg/go.sum pkg/
COPY services/gateway/go.mod services/gateway/go.sum services/gateway/

RUN cd services/gateway && go mod download

# --- Layer 2: Source code (changes often) ---
COPY pkg/ pkg/
COPY services/gateway/ services/gateway/

RUN cd services/gateway && \
    CGO_ENABLED=0 go build -ldflags="-s -w -X main.Version=${VERSION}" -o /app ./cmd/gateway/...

FROM gcr.io/distroless/static-debian12

COPY --from=builder /app /app

ENTRYPOINT ["/app"]
