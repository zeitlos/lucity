#!/usr/bin/env bash
#
# Test the trial billing system using Stripe Test Clocks (Simulations).
#
# Prerequisites:
#   1. Stripe CLI installed and authenticated (stripe login)
#   2. Cashier .env has sandbox price IDs
#   3. Optionally: make dev + stripe listen for webhook testing
#
# Usage:
#   ./scripts/test-trial-billing.sh [command]
#
# Commands:
#   create     - Verify subscription has thresholds + interval
#   threshold  - Heavy user triggers threshold invoice → suspend
#   interval   - Light user hits day 14 → suspend
#   add-plan   - Adding a plan clears trial params + resumes
#   cleanup    - Delete test clock, customer, and subscription
#   all        - Run create + add-plan + cleanup (default)

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
NC='\033[0m'

log()  { echo -e "${CYAN}[test]${NC} $*"; }
pass() { echo -e "${GREEN}[PASS]${NC} $*"; }
fail() { echo -e "${RED}[FAIL]${NC} $*"; exit 1; }
warn() { echo -e "${YELLOW}[WARN]${NC} $*"; }

# Load price IDs from cashier .env
CASHIER_ENV="$(cd "$(dirname "$0")/.." && pwd)/services/cashier/.env"
if [[ ! -f "$CASHIER_ENV" ]]; then
    fail "Missing $CASHIER_ENV — copy from .env.example and fill in sandbox values"
fi
source "$CASHIER_ENV"

ECO_CPU_METER="${STRIPE_ECO_CPU_METER_EVENT:-eco_cpu_usage}"

# State file to persist IDs across test runs
STATE_FILE="/tmp/lucity-trial-test-state.json"

save_state() {
    cat > "$STATE_FILE" <<EOF
{
  "clock_id": "$CLOCK_ID",
  "customer_id": "$CUSTOMER_ID",
  "subscription_id": "$SUB_ID",
  "frozen_time": $FROZEN_TIME
}
EOF
    log "State saved to $STATE_FILE"
}

load_state() {
    if [[ ! -f "$STATE_FILE" ]]; then
        fail "No state file found. Run 'create' test first."
    fi
    CLOCK_ID=$(jq -r .clock_id "$STATE_FILE")
    CUSTOMER_ID=$(jq -r .customer_id "$STATE_FILE")
    SUB_ID=$(jq -r .subscription_id "$STATE_FILE")
    FROZEN_TIME=$(jq -r .frozen_time "$STATE_FILE")
    log "Loaded state: clock=$CLOCK_ID customer=$CUSTOMER_ID sub=$SUB_ID"
}

# Wait for test clock to reach "ready" status after advancing
wait_clock_ready() {
    local clock_id="$1"
    local max_wait=60
    local elapsed=0
    while [[ $elapsed -lt $max_wait ]]; do
        local status
        status=$(stripe test_helpers test_clocks retrieve "$clock_id" 2>&1 | jq -r .status)
        if [[ "$status" == "ready" ]]; then
            return 0
        fi
        sleep 2
        elapsed=$((elapsed + 2))
    done
    fail "Test clock $clock_id did not reach ready status within ${max_wait}s"
}

# ─── Test: create ────────────────────────────────────────────────────────────

test_create() {
    log "Creating test clock + customer + subscription..."

    FROZEN_TIME=$(date +%s)

    # Create test clock
    CLOCK_ID=$(stripe test_helpers test_clocks create \
        --frozen-time="$FROZEN_TIME" \
        2>&1 | jq -r .id)
    log "Test clock: $CLOCK_ID (frozen at $(date -r "$FROZEN_TIME" '+%Y-%m-%d %H:%M'))"

    # Create customer on the clock
    CUSTOMER_ID=$(stripe customers create \
        --name "trial-test-$(date +%s)" \
        --test-clock="$CLOCK_ID" \
        -d "metadata[workspace]=test-trial-$$" \
        2>&1 | jq -r .id)
    log "Customer: $CUSTOMER_ID"

    # Create subscription with trial billing params via Stripe directly
    # (mirrors what CreateSubscription does)
    SUB_ID=$(stripe subscriptions create \
        --customer="$CUSTOMER_ID" \
        -d "payment_behavior=allow_incomplete" \
        -d "billing_thresholds[amount_gte]=500" \
        -d "pending_invoice_item_interval[interval]=day" \
        -d "pending_invoice_item_interval[interval_count]=14" \
        -d "metadata[workspace]=test-trial-$$" \
        -d "items[0][price]=${STRIPE_ECO_CPU_PRICE_ID}" \
        -d "items[1][price]=${STRIPE_ECO_MEM_PRICE_ID}" \
        -d "items[2][price]=${STRIPE_ECO_DISK_PRICE_ID}" \
        -d "items[3][price]=${STRIPE_PROD_CPU_PRICE_ID}" \
        -d "items[4][price]=${STRIPE_PROD_MEM_PRICE_ID}" \
        -d "items[5][price]=${STRIPE_PROD_DISK_PRICE_ID}" \
        2>&1 | jq -r .id)
    log "Subscription: $SUB_ID"

    # Create trial credit grant (500 cents, expires in 14 days + 1 hour)
    local expires_at=$((FROZEN_TIME + 14 * 86400 + 3600))
    stripe billing credit_grants create \
        --customer="$CUSTOMER_ID" \
        --name="Trial credit" \
        --category=promotional \
        -d "amount[type]=monetary" \
        -d "amount[monetary][currency]=eur" \
        -d "amount[monetary][value]=500" \
        -d "applicability_config[scope][price_type]=metered" \
        --expires-at="$expires_at" \
        > /dev/null 2>&1
    log "Credit grant created (500 cents, expires $(date -r "$expires_at" '+%Y-%m-%d %H:%M'))"

    # Verify subscription state
    local sub_json
    sub_json=$(stripe subscriptions retrieve "$SUB_ID" 2>&1)

    local status threshold interval_count
    status=$(echo "$sub_json" | jq -r .status)
    threshold=$(echo "$sub_json" | jq -r '.billing_thresholds.amount_gte // empty')
    interval_count=$(echo "$sub_json" | jq -r '.pending_invoice_item_interval.interval_count // empty')

    [[ "$status" == "active" ]] || fail "Expected status=active, got $status"
    [[ "$threshold" == "500" ]] || fail "Expected billing_thresholds.amount_gte=500, got $threshold"
    [[ "$interval_count" == "14" ]] || fail "Expected pending_invoice_item_interval.interval_count=14, got $interval_count"

    save_state

    pass "Subscription created with correct trial billing params"
    echo ""
    echo "  Status:              $status"
    echo "  Billing threshold:   €$(echo "scale=2; $threshold / 100" | bc)"
    echo "  Invoice interval:    ${interval_count} days"
    echo "  Items:               $(echo "$sub_json" | jq '.items.data | length') metered prices"
    echo ""
}

# ─── Test: threshold ─────────────────────────────────────────────────────────

test_threshold() {
    load_state
    log "Testing threshold invoice (heavy usage > €5)..."

    # Report enough CPU usage to exceed €5.
    # The exact amount depends on your meter price. We report a large value
    # to ensure it crosses the threshold.
    log "Reporting heavy usage..."
    stripe billing meter_events create \
        --event-name="$ECO_CPU_METER" \
        -d "payload[stripe_customer_id]=$CUSTOMER_ID" \
        -d "payload[value]=100000" \
        --timestamp="$FROZEN_TIME" \
        > /dev/null 2>&1

    # Advance clock by 1 day to trigger threshold processing
    local advance_to=$((FROZEN_TIME + 86400))
    log "Advancing clock to $(date -r "$advance_to" '+%Y-%m-%d %H:%M')..."
    stripe test_helpers test_clocks advance "$CLOCK_ID" \
        --frozen-time="$advance_to" \
        > /dev/null 2>&1
    FROZEN_TIME=$advance_to

    wait_clock_ready "$CLOCK_ID"

    # Check for invoices
    sleep 3  # Give webhook time to fire
    local invoices
    invoices=$(stripe invoices list --customer="$CUSTOMER_ID" 2>&1)

    local invoice_count
    invoice_count=$(echo "$invoices" | jq '.data | length')
    log "Found $invoice_count invoice(s)"

    if [[ "$invoice_count" -gt 0 ]]; then
        local latest_status latest_total
        latest_status=$(echo "$invoices" | jq -r '.data[0].status')
        latest_total=$(echo "$invoices" | jq -r '.data[0].total')
        echo ""
        echo "  Latest invoice status: $latest_status"
        echo "  Latest invoice total:  €$(echo "scale=2; $latest_total / 100" | bc)"
        echo ""
        pass "Threshold invoice generated"
    else
        warn "No invoices found yet. Check cashier logs for webhook activity."
        warn "The threshold may not have been reached with the reported usage amount."
        warn "Check: tail -f tmp/logs/cashier.log"
    fi

    save_state
    warn "Check cashier logs to verify workspace was suspended (no plan, no payment method)"
    echo "  tail -f tmp/logs/cashier.log | grep -E 'trial ended|suspend|resume'"
}

# ─── Test: interval ──────────────────────────────────────────────────────────

test_interval() {
    load_state
    log "Testing interval invoice (day 14, light usage)..."

    # Report small usage (well under €5)
    log "Reporting light usage..."
    stripe billing meter_events create \
        --event-name="$ECO_CPU_METER" \
        -d "payload[stripe_customer_id]=$CUSTOMER_ID" \
        -d "payload[value]=10" \
        --timestamp="$FROZEN_TIME" \
        > /dev/null 2>&1

    # Advance clock to day 14 from original frozen time
    local original_time
    original_time=$(jq -r .frozen_time "$STATE_FILE" 2>/dev/null || echo "$FROZEN_TIME")
    local advance_to=$((original_time + 14 * 86400))
    log "Advancing clock to day 14: $(date -r "$advance_to" '+%Y-%m-%d %H:%M')..."
    stripe test_helpers test_clocks advance "$CLOCK_ID" \
        --frozen-time="$advance_to" \
        > /dev/null 2>&1
    FROZEN_TIME=$advance_to

    wait_clock_ready "$CLOCK_ID"

    sleep 3  # Give webhook time to fire
    local invoices
    invoices=$(stripe invoices list --customer="$CUSTOMER_ID" 2>&1)

    local invoice_count
    invoice_count=$(echo "$invoices" | jq '.data | length')
    log "Found $invoice_count invoice(s)"

    if [[ "$invoice_count" -gt 0 ]]; then
        echo ""
        for i in $(seq 0 $((invoice_count - 1))); do
            local inv_status inv_total inv_id
            inv_id=$(echo "$invoices" | jq -r ".data[$i].id")
            inv_status=$(echo "$invoices" | jq -r ".data[$i].status")
            inv_total=$(echo "$invoices" | jq -r ".data[$i].total")
            echo "  Invoice $inv_id: status=$inv_status total=€$(echo "scale=2; $inv_total / 100" | bc)"
        done
        echo ""
        pass "Interval invoice generated at day 14"
    else
        warn "No invoices found. The interval invoice may not have fired yet."
    fi

    save_state
    warn "Check cashier logs to verify workspace was suspended"
    echo "  tail -f tmp/logs/cashier.log | grep -E 'trial ended|suspend|resume'"
}

# ─── Test: add-plan ──────────────────────────────────────────────────────────

test_add_plan() {
    load_state
    log "Testing plan addition (clears trial params + resets anchor)..."

    # Add a payment method (test token)
    local pm_id
    pm_id=$(stripe payment_methods create \
        --type=card \
        -d "card[token]=tok_visa" \
        2>&1 | jq -r .id)
    log "Payment method: $pm_id"

    stripe payment_methods attach "$pm_id" \
        --customer="$CUSTOMER_ID" \
        > /dev/null 2>&1

    stripe customers update "$CUSTOMER_ID" \
        -d "invoice_settings[default_payment_method]=$pm_id" \
        > /dev/null 2>&1
    log "Payment method attached and set as default"

    # Add hobby plan to subscription
    stripe subscription_items create \
        --subscription="$SUB_ID" \
        --price="$STRIPE_HOBBY_PRICE_ID" \
        -d "proration_behavior=create_prorations" \
        > /dev/null 2>&1
    log "Hobby plan added to subscription"

    # Clear trial billing params (mirrors ClearTrialBillingParams)
    stripe subscriptions update "$SUB_ID" \
        -d "billing_thresholds=" \
        -d "pending_invoice_item_interval=" \
        -d "billing_cycle_anchor=now" \
        -d "proration_behavior=create_prorations" \
        > /dev/null 2>&1
    log "Trial billing params cleared, anchor reset"

    # Verify
    local sub_json
    sub_json=$(stripe subscriptions retrieve "$SUB_ID" 2>&1)

    local status threshold interval item_count
    status=$(echo "$sub_json" | jq -r .status)
    threshold=$(echo "$sub_json" | jq -r '.billing_thresholds // "null"')
    interval=$(echo "$sub_json" | jq -r '.pending_invoice_item_interval // "null"')
    item_count=$(echo "$sub_json" | jq '.items.data | length')

    echo ""
    echo "  Status:                         $status"
    echo "  Billing thresholds:             $threshold"
    echo "  Pending invoice item interval:  $interval"
    echo "  Items:                          $item_count (expected 7: 6 metered + 1 plan)"
    echo ""

    [[ "$status" == "active" ]] || fail "Expected status=active, got $status"
    [[ "$threshold" == "null" ]] || fail "Expected billing_thresholds=null, got $threshold"
    [[ "$interval" == "null" ]] || fail "Expected pending_invoice_item_interval=null, got $interval"
    [[ "$item_count" == "7" ]] || fail "Expected 7 items, got $item_count"

    pass "Plan added, trial params cleared, anchor reset"
}

# ─── Test: cleanup ───────────────────────────────────────────────────────────

test_cleanup() {
    load_state
    log "Cleaning up test resources..."

    stripe subscriptions cancel "$SUB_ID" --confirm > /dev/null 2>&1 || true
    log "Subscription cancelled"

    stripe customers delete "$CUSTOMER_ID" --confirm > /dev/null 2>&1 || true
    log "Customer deleted"

    stripe test_helpers test_clocks delete "$CLOCK_ID" --confirm > /dev/null 2>&1 || true
    log "Test clock deleted"

    rm -f "$STATE_FILE"
    pass "Cleanup complete"
}

# ─── Main ────────────────────────────────────────────────────────────────────

# Check prerequisites
command -v stripe > /dev/null 2>&1 || fail "Stripe CLI not installed. Run: brew install stripe/stripe-cli/stripe"
command -v jq > /dev/null 2>&1 || fail "jq not installed. Run: brew install jq"

TEST="${1:-all}"

case "$TEST" in
    create)
        test_create
        ;;
    threshold)
        test_threshold
        ;;
    interval)
        test_interval
        ;;
    add-plan)
        test_add_plan
        ;;
    cleanup)
        test_cleanup
        ;;
    all)
        echo ""
        log "Running full trial billing test suite"
        echo "═══════════════════════════════════════════════════════"
        echo ""

        echo "── 1. Subscription Creation ──"
        test_create
        echo ""

        echo "── 2. Plan Addition (clears trial params) ──"
        test_add_plan
        echo ""

        echo "── Cleanup ──"
        test_cleanup
        echo ""

        echo "═══════════════════════════════════════════════════════"
        pass "All tests passed"
        echo ""
        warn "Threshold and interval tests require separate runs with"
        warn "fresh state since they conflict (both advance the clock)."
        echo ""
        echo "  To test threshold: ./scripts/test-trial-billing.sh create"
        echo "                     ./scripts/test-trial-billing.sh threshold"
        echo "                     ./scripts/test-trial-billing.sh cleanup"
        echo ""
        echo "  To test interval:  ./scripts/test-trial-billing.sh create"
        echo "                     ./scripts/test-trial-billing.sh interval"
        echo "                     ./scripts/test-trial-billing.sh cleanup"
        ;;
    *)
        echo "Usage: $0 {create|threshold|interval|add-plan|cleanup|all}"
        exit 1
        ;;
esac
