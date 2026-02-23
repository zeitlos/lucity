#!/usr/bin/env bash
set -euo pipefail

# Discover all CNPG read-write services across namespaces and let the user pick one.

# Find all -rw services managed by CNPG.
RAW=$(kubectl get svc --all-namespaces -l cnpg.io/cluster -o custom-columns='NS:.metadata.namespace,NAME:.metadata.name' --no-headers 2>/dev/null | grep '\-rw$' || true)

if [ -z "$RAW" ]; then
  echo "No CNPG databases found in the cluster."
  exit 1
fi

# Build options: "label|namespace|service" per line.
OPTIONS=()
while IFS= read -r line; do
  ns=$(echo "$line" | awk '{print $1}')
  svc=$(echo "$line" | awk '{print $2}')

  # Parse project and environment from namespace: {project}-{environment}
  project="" env=""
  if [[ "$ns" =~ ^(.+)-(development|staging|production|pr-[0-9]+)$ ]]; then
    project="${BASH_REMATCH[1]}"
    env="${BASH_REMATCH[2]}"
  fi

  # Extract the database name: strip the namespace prefix + -lucity-app-pg- prefix, then -rw suffix.
  db_name="${svc#"${ns}"-lucity-app-pg-}"
  db_name="${db_name%-rw}"

  if [ -n "$project" ] && [ -n "$env" ]; then
    OPTIONS+=("$project  ›  $env  ›  $db_name|$ns|$svc")
  else
    OPTIONS+=("$ns  ›  $db_name|$ns|$svc")
  fi
done <<< "$RAW"

# Single result — skip selection.
if [ ${#OPTIONS[@]} -eq 1 ]; then
  CHOICE="${OPTIONS[0]}"
else
  # Build display labels.
  DISPLAY=()
  for opt in "${OPTIONS[@]}"; do
    DISPLAY+=("${opt%%|*}")
  done

  if command -v fzf >/dev/null 2>&1; then
    SELECTED=$(printf '%s\n' "${DISPLAY[@]}" | fzf --prompt="Select database › " --height=~10 --reverse) || exit 0
    for opt in "${OPTIONS[@]}"; do
      if [ "${opt%%|*}" = "$SELECTED" ]; then
        CHOICE="$opt"
        break
      fi
    done
  else
    echo "Select a database to forward:"
    echo ""
    for i in "${!DISPLAY[@]}"; do
      echo "  $((i+1))) ${DISPLAY[$i]}"
    done
    echo ""
    read -rp "Enter number: " NUM
    if ! [[ "$NUM" =~ ^[0-9]+$ ]] || [ "$NUM" -lt 1 ] || [ "$NUM" -gt ${#OPTIONS[@]} ]; then
      echo "Invalid selection."
      exit 1
    fi
    CHOICE="${OPTIONS[$((NUM-1))]}"
  fi
fi

NS=$(echo "$CHOICE" | cut -d'|' -f2)
SVC=$(echo "$CHOICE" | cut -d'|' -f3)
LABEL="${CHOICE%%|*}"

echo ""
echo "Forwarding $LABEL → localhost:5432"
echo "Press Ctrl+C to stop."
echo ""
kubectl port-forward -n "$NS" "svc/$SVC" 5432:5432
