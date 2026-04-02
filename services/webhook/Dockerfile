ARG GO_VERSION=1.26

FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /workspace
ENV GOWORK=off

# --- Layer 1: Module metadata (changes rarely) ---
COPY pkg/go.mod pkg/go.sum pkg/
COPY services/webhook/go.mod services/webhook/go.sum services/webhook/

RUN cd services/webhook && go mod download

# --- Layer 2: Source code (changes often) ---
COPY pkg/ pkg/
COPY services/webhook/ services/webhook/

RUN cd services/webhook && \
    CGO_ENABLED=0 go build -ldflags="-s -w" -o /app ./cmd/webhook/...

FROM gcr.io/distroless/static-debian12

COPY --from=builder /app /app

ENTRYPOINT ["/app"]
