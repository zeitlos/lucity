ARG GO_VERSION=1.26

FROM golang:${GO_VERSION}-alpine AS builder

ARG SERVICE

WORKDIR /workspace

# Copy shared packages and charts (for dependency resolution via replace directives)
COPY pkg/ pkg/
COPY charts/go.mod charts/
COPY charts/embed.go charts/
COPY charts/lucity-app/ charts/lucity-app/

# Copy the target service
COPY services/${SERVICE}/ services/${SERVICE}/

# Build statically-linked binary without the Go workspace —
# each service's go.mod has replace directives that resolve local deps.
ENV GOWORK=off
RUN cd services/${SERVICE} && \
    CGO_ENABLED=0 go build -ldflags="-s -w" -o /app ./cmd/${SERVICE}/...

FROM gcr.io/distroless/static-debian12

COPY --from=builder /app /app

ENTRYPOINT ["/app"]
