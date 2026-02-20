#!/usr/bin/env bash
# scripts/dev.sh — Start all Lucity services with hot reload and live status
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
LOG_DIR="$ROOT/tmp/logs"
STATUS_DIR="$ROOT/tmp/dev"
MONITOR="$ROOT/scripts/dev-monitor.sh"

SERVICES=(gateway builder packager deployer webhook)
ALL_SERVICES=(gateway builder packager deployer webhook dashboard)
PORTS=(8080 9001 9002 9003 9004 5173)

# Colors
GREEN=$'\033[32m'
YELLOW=$'\033[33m'
RED=$'\033[31m'
DIM=$'\033[2m'
BOLD=$'\033[1m'
RESET=$'\033[0m'
CLEAR_LINE=$'\033[K'

# Ensure air is installed
if ! command -v air &>/dev/null; then
    echo "Installing air..."
    go install github.com/air-verse/air@latest
fi

# Create directories
mkdir -p "$LOG_DIR" "$STATUS_DIR"
mkdir -p "$ROOT/tmp/air"/{gateway,builder,packager,deployer,webhook}

# Truncate logs (fresh session)
for svc in "${ALL_SERVICES[@]}"; do
    > "$LOG_DIR/$svc.log"
done

# Clean status files
rm -f "$STATUS_DIR"/*.status

# Kill stale processes
for port in "${PORTS[@]}"; do
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
    tput cnorm 2>/dev/null || true
    echo "All services stopped."
}
trap cleanup EXIT INT TERM

# Start a Go service with air, piped through the monitor
start_service() {
    local name="$1"
    local dir="$ROOT/services/$name"
    (cd "$dir" && air 2>&1) | bash "$MONITOR" "$name" "$STATUS_DIR" "$LOG_DIR/$name.log" &
    PIDS+=($!)
}

# Start dashboard (Vite) with monitor
start_dashboard() {
    local now
    now=$(date +%s)
    cat > "$STATUS_DIR/dashboard.status" <<EOF
state=starting
build_start=0
build_end=0
build_duration=
run_start=$now
last_error=
EOF
    (
        cd "$ROOT/services/dashboard"
        npm run dev 2>&1 | while IFS= read -r line; do
            echo "$line" >> "$LOG_DIR/dashboard.log"
            # Strip ANSI for matching
            clean=$(printf '%s' "$line" | tr -d '[:cntrl:]' | sed 's/\[[0-9;]*[a-zA-Z]//g')
            if [[ "$clean" == *"ready in"* ]] || [[ "$clean" == *"Local:"*"http"* ]]; then
                now=$(date +%s)
                cat > "$STATUS_DIR/dashboard.status" <<INNER
state=running
build_start=0
build_end=0
build_duration=
run_start=$now
last_error=
INNER
            fi
        done
        echo "state=stopped" > "$STATUS_DIR/dashboard.status"
    ) &
    PIDS+=($!)
}

# Format seconds into human-readable duration
format_duration() {
    local secs="$1"
    if [[ "$secs" -lt 60 ]]; then
        printf '%ds' "$secs"
    elif [[ "$secs" -lt 3600 ]]; then
        printf '%dm %ds' "$((secs / 60))" "$((secs % 60))"
    else
        printf '%dh %dm' "$((secs / 3600))" "$((secs % 3600 / 60))"
    fi
}

# Read a value from a status file
read_status() {
    local file="$1" key="$2" default="${3:-}"
    if [[ -f "$file" ]]; then
        local val
        val=$(grep "^${key}=" "$file" 2>/dev/null | head -1 | cut -d= -f2-)
        echo "${val:-$default}"
    else
        echo "$default"
    fi
}

# --- Main ---

# Start all services
for svc in "${SERVICES[@]}"; do
    start_service "$svc"
done
start_dashboard

# Print header
printf '\n'
printf '  %slucity dev%s\n' "$BOLD" "$RESET"
printf '\n'

# The status table occupies lines below the header.
# We use cursor-save/restore to redraw in place.
# Print initial blank lines to reserve space (header + 6 services + footer = 10 lines)
TABLE_LINES=$(( ${#ALL_SERVICES[@]} + 3 ))  # header + services + blank + footer
for ((i = 0; i < TABLE_LINES; i++)); do
    echo ""
done

# Hide cursor
tput civis 2>/dev/null || true

# Status display loop (foreground — Ctrl+C triggers cleanup)
while true; do
    now=$(date +%s)

    # Move cursor up to the start of the table
    printf '\033[%dA' "$TABLE_LINES"

    # Header
    printf '  %s%-13s %-12s %-10s %s%s%s\n' "$BOLD" "SERVICE" "STATUS" "BUILD" "UPTIME" "$RESET" "$CLEAR_LINE"

    for svc in "${ALL_SERVICES[@]}"; do
        local_file="$STATUS_DIR/$svc.status"
        state=$(read_status "$local_file" "state" "starting")
        build_duration=$(read_status "$local_file" "build_duration" "")
        run_start=$(read_status "$local_file" "run_start" "$now")
        build_start=$(read_status "$local_file" "build_start" "0")

        status_str=""
        build_str=""
        uptime_str=""

        case "$state" in
            running)
                status_str="${GREEN}● running${RESET}"
                if [[ -n "$build_duration" && "$build_duration" != "0" ]]; then
                    build_str="${DIM}${build_duration}s${RESET}"
                fi
                if [[ -n "$run_start" && "$run_start" -gt 0 ]]; then
                    uptime_str=$(format_duration $((now - run_start)))
                fi
                ;;
            building)
                elapsed=""
                if [[ -n "$build_start" && "$build_start" -gt 0 ]]; then
                    elapsed="$((now - build_start))s"
                fi
                status_str="${YELLOW}● building${RESET}"
                build_str="${YELLOW}${elapsed:-...}${RESET}"
                ;;
            failed)
                status_str="${RED}✗ failed${RESET}"
                if [[ -n "$build_duration" ]]; then
                    build_str="${DIM}${build_duration}s${RESET}"
                fi
                ;;
            starting)
                status_str="${DIM}○ starting${RESET}"
                ;;
            stopped)
                status_str="${RED}○ stopped${RESET}"
                ;;
            *)
                status_str="${DIM}○ ${state}${RESET}"
                ;;
        esac

        # Print the row — use %b for ANSI interpretation
        printf '  %-13s ' "$svc"
        printf '%-22b' "$status_str"
        printf '%-16b' "$build_str"
        printf '%-10s' "$uptime_str"
        printf '%s\n' "$CLEAR_LINE"
    done

    printf '%s\n' "$CLEAR_LINE"
    printf '  %sLogs: tmp/logs/   Stop: Ctrl+C%s%s\n' "$DIM" "$RESET" "$CLEAR_LINE"

    sleep 1
done
