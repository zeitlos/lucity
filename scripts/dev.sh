#!/usr/bin/env bash
# scripts/dev.sh — Start all Lucity services with hot reload and live status
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
LOG_DIR="$ROOT/tmp/logs"
STATUS_DIR="$ROOT/tmp/dev"
MONITOR="$ROOT/scripts/dev-monitor.sh"

SERVICES=(gateway builder packager deployer webhook cashier)
ALL_SERVICES=(gateway builder packager deployer webhook cashier dashboard)
PORTS=(8080 9001 9002 9003 9004 9005 9006 5173)

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
mkdir -p "$ROOT/tmp/air"/{gateway,builder,packager,deployer,webhook,cashier}

# Truncate logs (fresh session)
for svc in "${ALL_SERVICES[@]}"; do
    > "$LOG_DIR/$svc.log"
done

# Clean status files
rm -f "$STATUS_DIR"/*.status

# Kill stale processes (only listeners, not clients like browsers)
for port in "${PORTS[@]}"; do
    lsof -ti :"$port" -sTCP:LISTEN | xargs kill 2>/dev/null || true
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
    printf '  %-14s%-12s%-10s%s%s\n' "SERVICE" "STATUS" "BUILD" "UPTIME" "$CLEAR_LINE"

    for svc in "${ALL_SERVICES[@]}"; do
        svc_file="$STATUS_DIR/$svc.status"
        state=$(read_status "$svc_file" "state" "starting")
        build_duration=$(read_status "$svc_file" "build_duration" "")
        run_start=$(read_status "$svc_file" "run_start" "$now")
        build_start=$(read_status "$svc_file" "build_start" "0")

        status_text=""; build_text=""; uptime_text=""
        status_color="$DIM"; build_color="$DIM"

        case "$state" in
            running)
                status_text="● running"
                status_color="$GREEN"
                if [[ -n "$build_duration" && "$build_duration" != "0" ]]; then
                    build_text="${build_duration}s"
                fi
                if [[ -n "$run_start" && "$run_start" -gt 0 ]]; then
                    uptime_text=$(format_duration $((now - run_start)))
                fi
                ;;
            building)
                status_text="● building"
                status_color="$YELLOW"
                if [[ -n "$build_start" && "$build_start" -gt 0 ]]; then
                    build_text="$((now - build_start))s"
                else
                    build_text="..."
                fi
                build_color="$YELLOW"
                ;;
            failed)
                status_text="✗ failed"
                status_color="$RED"
                if [[ -n "$build_duration" ]]; then
                    build_text="${build_duration}s"
                fi
                ;;
            starting)
                status_text="○ starting"
                ;;
            stopped)
                status_text="○ stopped"
                status_color="$RED"
                ;;
            *)
                status_text="○ ${state}"
                ;;
        esac

        # Pad each field to fixed width as plain text, then colorize
        col_svc=$(printf '%-14s' "$svc")
        col_status=$(printf '%-12s' "$status_text")
        col_build=$(printf '%-10s' "$build_text")

        printf '  %s%s%s%s%s%s%s%s\n' \
            "$col_svc" \
            "$status_color" "$col_status" "$RESET" \
            "$build_color" "$col_build" "$RESET" \
            "$uptime_text$CLEAR_LINE"
    done

    printf '%s\n' "$CLEAR_LINE"
    printf '  %sLogs: tmp/logs/   Stop: Ctrl+C%s%s\n' "$DIM" "$RESET" "$CLEAR_LINE"

    sleep 1
done
