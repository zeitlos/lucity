#!/bin/bash
# Seed the registry pull secret into all existing workload namespaces.
# Run after deploying the lucity chart with registryPull values.
#
# The deployer automatically creates the secret for new environments,
# but existing namespaces need this one-time seed.
#
# Usage: ./scripts/seed-registry-pull-secret.sh --context lucity-prod

set -euo pipefail

CONTEXT_FLAG=""
if [[ "${1:-}" == "--context" ]]; then
  CONTEXT_FLAG="--context $2"
fi

SOURCE_NS="lucity-system"
SOURCE_SECRET="lucity-registry-pull"
TARGET_SECRET="lucity-registry"

echo "Reading source secret $SOURCE_SECRET from $SOURCE_NS..."
SOURCE_JSON=$(kubectl $CONTEXT_FLAG get secret "$SOURCE_SECRET" -n "$SOURCE_NS" -o json)
SOURCE_TYPE=$(echo "$SOURCE_JSON" | jq -r '.type')
SOURCE_DATA=$(echo "$SOURCE_JSON" | jq -r '.data[".dockerconfigjson"]')

NAMESPACES=$(kubectl $CONTEXT_FLAG get namespaces -l "lucity.dev/managed-by=lucity" -o jsonpath='{.items[*].metadata.name}')

if [[ -z "$NAMESPACES" ]]; then
  echo "No workload namespaces found."
  exit 0
fi

COUNT=0
for NS in $NAMESPACES; do
  echo "  $NS"
  kubectl $CONTEXT_FLAG apply -f - <<YAML
apiVersion: v1
kind: Secret
metadata:
  name: $TARGET_SECRET
  namespace: $NS
  labels:
    lucity.dev/managed-by: lucity
type: $SOURCE_TYPE
data:
  .dockerconfigjson: $SOURCE_DATA
YAML
  COUNT=$((COUNT + 1))
done

echo "Done. Seeded $TARGET_SECRET into $COUNT namespaces."
