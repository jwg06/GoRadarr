# Stage 1: Build frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# Stage 2: Build Go binary (Debian-based — git/ca-certs/tzdata pre-installed, no apk needed)
FROM golang:1.26 AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN rm -rf ./internal/server/frontend && mkdir -p ./internal/server
COPY --from=frontend-builder /app/frontend/dist ./internal/server/frontend
RUN CGO_ENABLED=0 GOOS=linux go build       -ldflags="-s -w -X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo dev)"       -o /app/bin/goradarr       ./cmd/goradarr

FROM gcr.io/distroless/static-debian12:nonroot
LABEL org.opencontainers.image.title="GoRadarr"       org.opencontainers.image.description="Radarr rebuilt in Go — blazing fast movie collection manager"       org.opencontainers.image.url="https://github.com/jwg06/GoRadarr"       org.opencontainers.image.source="https://github.com/jwg06/GoRadarr"       org.opencontainers.image.licenses="GPL-3.0"

COPY --from=go-builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=go-builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=go-builder /app/bin/goradarr /goradarr

VOLUME ["/config", "/media"]
EXPOSE 7878

ENV GORADARR_HOST=0.0.0.0     GORADARR_PORT=7878     GORADARR_DATABASE_DSN=/config/goradarr.db     GORADARR_DATA_ROOT_DIR=/config

HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3   CMD ["/goradarr", "-healthcheck"]

USER nonroot:nonroot
ENTRYPOINT ["/goradarr"]
