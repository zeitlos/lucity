#!/usr/bin/env bash
# scripts/dev-monitor.sh — Monitors a service's output and writes status files
# Usage: dev-monitor.sh <service-name> <status-dir> <log-file>
#
# Reads stdin line by line, detects air build events, writes state to a status
# file, and forwards all output to the log file.

set -uo pipefail

SERVICE="$1"
STATUS_DIR="$2"
LOG_FILE="$3"
STATUS_FILE="$STATUS_DIR/$SERVICE.status"

write_status() {
    cat > "$STATUS_FILE" <<EOF
state=$state
build_start=$build_start
build_end=$build_end
build_duration=$build_duration
run_start=$run_start
last_error=$last_error
EOF
}

# State
state="starting"
build_start=0
build_end=0
build_duration=""
run_start=$(date +%s)
last_error=""

write_status

while IFS= read -r line; do
    # Write raw line to log file
    echo "$line" >> "$LOG_FILE"

    # Strip ANSI escape codes for pattern matching
    clean=$(printf '%s' "$line" | sed $'s/\033\\[[0-9;]*[a-zA-Z]//g')

    case "$clean" in
        *"building..."*)
            state="building"
            build_start=$(date +%s)
            build_end=0
            build_duration=""
            last_error=""
            write_status
            ;;
        *"running..."*)
            now=$(date +%s)
            if [[ "$build_start" -gt 0 ]]; then
                build_duration=$((now - build_start))
            fi
            state="running"
            build_end=$now
            run_start=$now
            write_status
            ;;
        *"failed to build"*)
            now=$(date +%s)
            if [[ "$build_start" -gt 0 ]]; then
                build_duration=$((now - build_start))
            fi
            state="failed"
            build_end=$now
            # Extract error message
            last_error="${clean##*error: }"
            last_error="${last_error:0:120}"
            write_status
            ;;
    esac
done

# Process ended
state="stopped"
write_status
