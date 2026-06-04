#!/usr/bin/env bash
# scripts/dev-stop.sh — Stop every dev service by killing its listener
# ports. Uses the shared SERVICE_PORTS map so it can never disagree with
# what scripts/dev.sh starts.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"

# shellcheck disable=SC1091
source "$ROOT/scripts/dev-ports.sh"

for svc in "${SERVICE_NAMES[@]}"; do
    for port in $(service_ports "$svc"); do
        lsof -ti :"$port" -sTCP:LISTEN | xargs kill 2>/dev/null || true
    done
done

echo "All services stopped."
