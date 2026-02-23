#!/usr/bin/env bash
# scripts/test-watch.sh — Watch for changes and re-run integration tests
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
LOG_DIR="$ROOT/tmp/logs"
STATUS_DIR="$ROOT/tmp/dev"
LOG_FILE="$LOG_DIR/tests.log"
STATUS_FILE="$STATUS_DIR/tests.status"

# Colors
GREEN=$'\033[32m'
RED=$'\033[31m'
YELLOW=$'\033[33m'
DIM=$'\033[2m'
BOLD=$'\033[1m'
RESET=$'\033[0m'

mkdir -p "$LOG_DIR" "$STATUS_DIR"

write_status() {
    cat > "$STATUS_FILE" <<EOF
state=$1
run_start=${2:-}
last_error=${3:-}
EOF
}

# Check for watchexec
if ! command -v watchexec &>/dev/null; then
    echo "${RED}watchexec is not installed.${RESET}"
    echo "Install it with: ${BOLD}brew install watchexec${RESET}"
    echo "Or: ${BOLD}cargo install watchexec-cli${RESET}"
    exit 1
fi

# Parse arguments
TEST_ARGS="${*:--v -count=1 -run TestIntegration ./...}"

echo "${BOLD}Lucity Integration Test Watcher${RESET}"
echo "${DIM}Watching: tests/ services/ pkg/ for .go and .graphqls changes${RESET}"
echo "${DIM}Log file: $LOG_FILE${RESET}"
echo "${DIM}Test args: $TEST_ARGS${RESET}"
echo ""

# Run tests with watchexec
# --restart: kill previous test run on new changes
# --debounce: wait 1.5s after last change before re-running
# --exts: only watch Go and GraphQL schema files
watchexec \
    -w "$ROOT/tests/" \
    -w "$ROOT/services/" \
    -w "$ROOT/pkg/" \
    --exts go,graphqls \
    --debounce 1500 \
    --restart \
    -- bash -c "
        echo '${YELLOW}Running tests...${RESET}'
        write_status() {
            cat > '$STATUS_FILE' <<INNER_EOF
state=\$1
run_start=\$(date +%s)
last_error=\${2:-}
INNER_EOF
        }
        write_status running

        cd '$ROOT/tests'
        if go test $TEST_ARGS 2>&1 | tee '$LOG_FILE'; then
            write_status passed
            echo ''
            echo '${GREEN}${BOLD}PASS${RESET}'
        else
            # Extract last FAIL line for status
            LAST_FAIL=\$(grep -E '^--- FAIL' '$LOG_FILE' | tail -1 || true)
            write_status failed \"\$LAST_FAIL\"
            echo ''
            echo '${RED}${BOLD}FAIL${RESET}'
        fi
    "
