ARG GO_VERSION=1.26

FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /workspace
ENV GOWORK=off

COPY pkg/go.mod pkg/go.sum pkg/
COPY services/deployer/go.mod services/deployer/go.sum services/deployer/
RUN cd services/deployer && go mod download

COPY pkg/ pkg/
COPY services/deployer/ services/deployer/
RUN cd services/deployer && \
    CGO_ENABLED=0 go build -ldflags="-s -w" -o /app ./cmd/deployer/...

FROM gcr.io/distroless/static-debian12

COPY --from=builder /app /app

ENTRYPOINT ["/app"]
