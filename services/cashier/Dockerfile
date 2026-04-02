ARG GO_VERSION=1.26

FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /workspace
ENV GOWORK=off

COPY pkg/go.mod pkg/go.sum pkg/
COPY services/cashier/go.mod services/cashier/go.sum services/cashier/
RUN cd services/cashier && go mod download

COPY pkg/ pkg/
COPY services/cashier/ services/cashier/
RUN cd services/cashier && \
    CGO_ENABLED=0 go build -ldflags="-s -w" -o /app ./cmd/cashier/...

FROM gcr.io/distroless/static-debian12

COPY --from=builder /app /app

ENTRYPOINT ["/app"]
