#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

API_BASE_URL="${API_BASE_URL:-http://127.0.0.1:5001}"
export API_BASE_URL

SERVER_PID=""

cleanup() {
  if [[ -n "$SERVER_PID" ]]; then
    kill "$SERVER_PID" 2>/dev/null || true
    wait "$SERVER_PID" 2>/dev/null || true
  fi
}

trap cleanup EXIT

wait_for_api() {
  for _ in $(seq 1 60); do
    if curl -sf "$API_BASE_URL/health" >/dev/null 2>&1; then
      return 0
    fi
    sleep 1
  done
  echo "API did not become ready at $API_BASE_URL" >&2
  return 1
}

if curl -sf "$API_BASE_URL/health" >/dev/null 2>&1; then
  echo "Using existing API at $API_BASE_URL"
else
  echo "Starting API for Playwright tests..."
  go run ./cmd/seed/main.go
  go run ./cmd/server/main.go &
  SERVER_PID=$!
  wait_for_api
  echo "API ready at $API_BASE_URL"
fi

npm run test:api
