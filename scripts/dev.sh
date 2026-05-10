#!/usr/bin/env sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
ROOT_DIR=$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)
BACKEND_DIR="$ROOT_DIR/backend"
FRONTEND_DIR="$ROOT_DIR/frontend"

BACKEND_PID=""
FRONTEND_PID=""

cleanup() {
  trap - INT TERM EXIT

  if [ -n "$BACKEND_PID" ] && kill -0 "$BACKEND_PID" 2>/dev/null; then
    echo "[dev] stopping backend pid: $BACKEND_PID"
    kill "$BACKEND_PID" 2>/dev/null || true
  fi

  if [ -n "$FRONTEND_PID" ] && kill -0 "$FRONTEND_PID" 2>/dev/null; then
    echo "[dev] stopping frontend pid: $FRONTEND_PID"
    kill "$FRONTEND_PID" 2>/dev/null || true
  fi

  wait "$BACKEND_PID" 2>/dev/null || true
  wait "$FRONTEND_PID" 2>/dev/null || true
}

trap cleanup INT TERM EXIT

for cache_path in "$FRONTEND_DIR/node_modules/.vite" "$FRONTEND_DIR/.vite"; do
  if [ -e "$cache_path" ]; then
    echo "[dev] removing Vite cache: $cache_path"
    rm -rf "$cache_path"
  fi
done

echo "[dev] starting backend from $BACKEND_DIR"
(cd "$BACKEND_DIR" && go run ./cmd/server) &
BACKEND_PID=$!
echo "[dev] backend pid: $BACKEND_PID"

echo "[dev] starting frontend from $FRONTEND_DIR"
(cd "$FRONTEND_DIR" && npm run dev -- --host 0.0.0.0 --port 5173 --force) &
FRONTEND_PID=$!
echo "[dev] frontend pid: $FRONTEND_PID"

echo "[dev] press Ctrl+C to stop both"

wait "$BACKEND_PID" "$FRONTEND_PID"
