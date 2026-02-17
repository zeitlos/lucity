#!/usr/bin/env bash
# scripts/dev.sh — Start all Lucity services with hot reload
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
LOG_DIR="$ROOT/tmp/logs"

# Ensure air is installed
if ! command -v air &>/dev/null; then
    echo "Installing air..."
    go install github.com/air-verse/air@latest
fi

# Create directories
mkdir -p "$LOG_DIR"
mkdir -p "$ROOT/tmp/air"/{gateway,builder,packager,deployer,webhook}

# Truncate logs (fresh session)
for svc in gateway builder packager deployer webhook dashboard; do
    > "$LOG_DIR/$svc.log"
done

# Kill stale processes
for port in 8080 9001 9002 9003 9004 5173; do
    lsof -ti :"$port" | xargs kill 2>/dev/null || true
done
sleep 1

# Track child PIDs for cleanup
PIDS=()
cleanup() {
    echo ""
    echo "Shutting down all services..."
    for pid in "${PIDS[@]}"; do
        kill "$pid" 2>/dev/null || true
    done
    wait 2>/dev/null
    echo "All services stopped."
}
trap cleanup EXIT INT TERM

# Start a Go service with air
start_service() {
    local name="$1"
    local dir="$ROOT/services/$name"
    echo "  $name"
    (cd "$dir" && air >> "$LOG_DIR/$name.log" 2>&1) &
    PIDS+=($!)
}

echo "Starting services..."
start_service gateway
start_service builder
start_service packager
start_service deployer
start_service webhook

# Start dashboard (Vite already has HMR)
echo "  dashboard"
(cd "$ROOT/services/dashboard" && npm run dev >> "$LOG_DIR/dashboard.log" 2>&1) &
PIDS+=($!)

echo ""
echo "All services started. Logs in tmp/logs/"
echo ""
echo "  Gateway:    http://localhost:8080 (playground: /playground)"
echo "  Dashboard:  http://localhost:5173"
echo "  Builder:    gRPC :9001"
echo "  Packager:   gRPC :9002"
echo "  Deployer:   gRPC :9003"
echo "  Webhook:    http://localhost:9004"
echo ""
echo "  Logs:       make dev-logs"
echo "  Stop:       Ctrl+C"
echo ""

# Wait for any child to exit (or Ctrl+C)
wait
