# One image, two programs: `migrate` and `server` come from the same commit so the schema and the code that
# depends on it can never arrive separately. The deployment runs `migrate` as an init container first.

FROM golang:1.26-alpine AS builder

WORKDIR /src

# Dependencies on their own layer, so editing Go files does not re-download them.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO off: the PostgreSQL driver is pure Go, so the binaries run on a base image without a matching libc.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server \
    && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate

FROM alpine:3.21

# tzdata: every business date is a Taipei date. The binary also embeds the zone database, so this is a second
# line of defence rather than the only one.
RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -H -u 10001 board

COPY --from=builder /out/server /usr/local/bin/server
COPY --from=builder /out/migrate /usr/local/bin/migrate

USER board

EXPOSE 8080

# For `docker run` by hand; Kubernetes uses its own probes.
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget --quiet --spider http://127.0.0.1:8080/health || exit 1

ENTRYPOINT ["/usr/local/bin/server"]
