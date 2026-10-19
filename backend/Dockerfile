# ---------- build stage ----------
FROM golang:1.23-alpine AS builder

WORKDIR /src

RUN apk add --no-cache git ca-certificates

# Dependencies first: this layer is cached as long as go.mod/go.sum are unchanged.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Static binary: the runtime image needs no Go toolchain and no libc.
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath -ldflags="-s -w" \
    -o /out/siswa-api ./cmd/api

# ---------- runtime stage ----------
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata curl \
    && adduser -D -u 10001 siswa

WORKDIR /app

COPY --from=builder /out/siswa-api /app/siswa-api

# Directory for the stock mutation audit log (STOCK_LOG_FILE).
RUN mkdir -p /app/logs && chown -R siswa:siswa /app

USER siswa

ENV APP_PORT=3069
EXPOSE 3069

HEALTHCHECK --interval=15s --timeout=3s --start-period=20s --retries=5 \
    CMD curl -fsS "http://127.0.0.1:${APP_PORT}/health" || exit 1

ENTRYPOINT ["/app/siswa-api"]
