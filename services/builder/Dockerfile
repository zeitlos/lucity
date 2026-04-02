ARG GO_VERSION=1.26

FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /workspace
ENV GOWORK=off

COPY pkg/go.mod pkg/go.sum pkg/
COPY services/builder/go.mod services/builder/go.sum services/builder/
RUN cd services/builder && go mod download

COPY pkg/ pkg/
COPY services/builder/ services/builder/
RUN cd services/builder && \
    CGO_ENABLED=0 go build -ldflags="-s -w" -o /app ./cmd/builder/...

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /app /app

ENTRYPOINT ["/app"]
