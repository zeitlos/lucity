ARG GO_VERSION=1.27

FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine AS builder

WORKDIR /workspace
ENV GOWORK=off

COPY pkg/go.mod pkg/go.sum pkg/
COPY services/cashier/go.mod services/cashier/go.sum services/cashier/
RUN --mount=type=cache,target=/go/pkg/mod \
    cd services/cashier && go mod download

COPY pkg/ pkg/
COPY services/cashier/ services/cashier/
ARG TARGETOS TARGETARCH
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    cd services/cashier && \
    GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go build -ldflags="-s -w" -o /app ./cmd/cashier/...

FROM gcr.io/distroless/static-debian12

COPY --from=builder /app /app

ENTRYPOINT ["/app"]
