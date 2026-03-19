# Stage 1: Build frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
# Only copy package files first for layer caching
COPY frontend/package*.json ./
RUN npm ci --prefer-offline 2>/dev/null || npm install
COPY frontend/ ./
RUN npm run build

# Stage 2: Build Go binary
FROM golang:1.22-alpine AS go-builder
RUN apk add --no-cache git ca-certificates tzdata
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist
RUN CGO_ENABLED=0 GOOS=linux go build \
      -ldflags="-s -w -X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo dev)" \
      -o /app/bin/goradarr \
      ./cmd/goradarr

# Stage 3: Minimal final image
FROM gcr.io/distroless/static-debian12:nonroot
LABEL org.opencontainers.image.title="GoRadarr" \
      org.opencontainers.image.description="Radarr rebuilt in Go — blazing fast movie collection manager" \
      org.opencontainers.image.url="https://github.com/jwg06/GoRadarr" \
      org.opencontainers.image.source="https://github.com/jwg06/GoRadarr" \
      org.opencontainers.image.licenses="GPL-3.0"

# Copy timezone data and TLS certs from builder
COPY --from=go-builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=go-builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy binary
COPY --from=go-builder /app/bin/goradarr /goradarr

# Data directories (will be mounted as volumes)
# /config — database + config file
# /media  — movie root folders
VOLUME ["/config", "/media"]

EXPOSE 7878

ENV GORADARR_HOST=0.0.0.0 \
    GORADARR_PORT=7878 \
    GORADARR_DATABASE_DSN=/config/goradarr.db \
    GORADARR_DATA_ROOT_DIR=/config

HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
  CMD ["/goradarr", "-healthcheck"] || exit 1

USER nonroot:nonroot

ENTRYPOINT ["/goradarr"]
