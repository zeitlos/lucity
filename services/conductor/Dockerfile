ARG GO_VERSION=1.26

FROM golang:${GO_VERSION}-alpine AS builder

ARG VERSION=dev

WORKDIR /workspace
ENV GOWORK=off

COPY pkg/go.mod pkg/go.sum pkg/
COPY charts/go.mod charts/go.sum charts/
COPY services/conductor/go.mod services/conductor/go.sum services/conductor/
RUN cd services/conductor && go mod download

COPY pkg/ pkg/
COPY charts/ charts/
COPY services/conductor/ services/conductor/
RUN cd services/conductor && \
    CGO_ENABLED=0 go build -ldflags="-s -w -X main.Version=${VERSION}" -o /app ./cmd/conductor/...

# debian-slim rather than distroless: conductor doubles as the in-pod
# build runner ("conductor run-build") which shells out to git +
# buildkit and needs a real OS.
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates git && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /app /app

ENTRYPOINT ["/app"]
