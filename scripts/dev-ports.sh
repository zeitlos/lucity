#!/usr/bin/env bash
# scripts/dev-ports.sh — Single source of truth for which dev ports each
# service listens on. Sourced by scripts/dev.sh and scripts/dev-stop.sh so
# the SKIP filter and the "stop everything" path stay in agreement.
#
# Kept bash 3.2 compatible (the system bash macOS ships) — no associative
# arrays. SERVICE_NAMES is the full list; service_ports maps name -> ports.

SERVICE_NAMES=(conductor cashier dashboard)

# service_ports <name> — echo the space-separated ports for a service.
service_ports() {
    case "$1" in
        conductor) echo "8080 9004 9090" ;;
        cashier)   echo "9005 9006" ;;
        dashboard) echo "5173" ;;
        *)         echo "" ;;
    esac
}
