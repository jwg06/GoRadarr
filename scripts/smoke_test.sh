#!/usr/bin/env bash
set -euo pipefail

PORT=17879
TMPDIR="$(mktemp -d)"
BINARY="./bin/goradarr-smoke"
PID=""

cleanup() {
    if [ -n "$PID" ]; then
        kill "$PID" 2>/dev/null || true
        wait "$PID" 2>/dev/null || true
    fi
    rm -rf "$TMPDIR"
    rm -f "$BINARY"
}
trap cleanup EXIT

echo "==> Building binary..."
go build -o "$BINARY" ./cmd/goradarr

echo "==> Starting server on port $PORT (data dir: $TMPDIR)..."
GORADARR_PORT="$PORT" \
GORADARR_DATA_ROOT_DIR="$TMPDIR" \
GORADARR_DATABASE_DSN="$TMPDIR/goradarr.db" \
GORADARR_AUTH_ENABLED=false \
GORADARR_SCHEDULER_ENABLED=false \
GORADARR_LOG_TARGET=stdout \
    "$BINARY" &
PID=$!

BASE="http://localhost:$PORT"

echo "==> Waiting for server to become ready..."
for i in $(seq 1 20); do
    if curl -sf "$BASE/api/v1/ping" >/dev/null 2>&1; then
        echo "    Server ready after ${i} attempts."
        break
    fi
    if [ "$i" -eq 20 ]; then
        echo "ERROR: Server did not become ready in time." >&2
        exit 1
    fi
    sleep 0.5
done

FAILED=0
check() {
    local url="$1"
    local desc="$2"
    if curl -sf "$url" >/dev/null 2>&1; then
        echo "  OK  $desc"
    else
        echo "  FAIL $desc ($url)" >&2
        FAILED=1
    fi
}

echo "==> Testing endpoints..."
check "$BASE/api/v1/ping"              "GET /api/v1/ping"
check "$BASE/api/v1/system/status"     "GET /api/v1/system/status"
check "$BASE/api/v1/movie"             "GET /api/v1/movie"
check "$BASE/api/v1/qualityprofile"    "GET /api/v1/qualityprofile"
check "$BASE/api/v1/queue/status"      "GET /api/v1/queue/status"
check "$BASE/openapi.yaml"             "GET /openapi.yaml"
check "$BASE/metrics"                  "GET /metrics"

if [ "$FAILED" -ne 0 ]; then
    echo "==> Smoke test FAILED." >&2
    exit 1
fi

echo "==> Smoke test PASSED."
exit 0
