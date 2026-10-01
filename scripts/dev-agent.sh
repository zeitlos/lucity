#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
AGENT="${AGENT:-ops-agent}"
CONTEXT="${DEV_CONTEXT:-lucity-dev}"
NAMESPACE=lucity-agents
NAME="$AGENT-local"
IMAGE=lucity-switchboard:local
CHART="$ROOT/charts/lucity-agent"
VALUES="$ROOT/deployments/lucity-dev/$AGENT-values.yaml"
SECRETS="$ROOT/deployments/lucity-dev/$AGENT-secrets.yaml"
RUN="$ROOT/tmp/agents/$AGENT"

[[ -f "$VALUES" ]] || { echo "Error: $VALUES not found." >&2; exit 1; }
[[ -f "$SECRETS" ]] || { echo "Error: $SECRETS not found. Copy deployments/$AGENT-secrets.yaml.example and fill in values." >&2; exit 1; }

if [[ "$(kubectl --context "$CONTEXT" -n "$NAMESPACE" get deployment "$AGENT" -o jsonpath='{.status.replicas}' 2>/dev/null)" =~ ^[1-9] ]]; then
    echo "Error: $AGENT is running on $CONTEXT. Telegram allows one poller per bot token, so scale it to 0 first." >&2
    exit 1
fi

render() {
    helm template "$NAME" "$CHART" -n "$NAMESPACE" -f "$VALUES" --show-only "templates/$1"
}

cleanup() {
    for pid in $(jobs -p); do
        kill "$pid" 2>/dev/null || true
        wait "$pid" 2>/dev/null || true
    done
    if [[ -f "$RUN/identity.yaml" ]]; then
        kubectl --context "$CONTEXT" -n "$NAMESPACE" delete -f "$RUN/identity.yaml" --ignore-not-found >/dev/null || true
        rm -f "$RUN/identity.yaml" "$RUN/data/home/.kube/config"
    fi
}
trap cleanup EXIT

echo "Building $IMAGE"
docker build -q -f "$ROOT/services/switchboard/Dockerfile" -t "$IMAGE" "$ROOT" >/dev/null

rm -rf "$RUN/config" "$RUN/identity.yaml"
mkdir -p "$RUN/config" "$RUN/data/home/.kube"
render configmap.yaml >"$RUN/config/configmap.yaml"
yq '.data["CLAUDE.md"]' "$RUN/config/configmap.yaml" >"$RUN/config/CLAUDE.md"
yq '.data["settings.json"]' "$RUN/config/configmap.yaml" >"$RUN/config/settings.json"
yq '.data["mcp-servers.json"]' "$RUN/config/configmap.yaml" >"$RUN/config/mcp-servers.json"

port=19400
for url in $(jq -r '[.. | strings | select(test("^http://[a-z0-9-]+\\.[a-z0-9-]+\\.svc(\\.cluster\\.local)?:[0-9]+"))] | unique | .[]' "$RUN/config/mcp-servers.json"); do
    authority="${url#http://}"
    authority="${authority%%/*}"
    host="${authority%:*}"
    service="${host%%.*}"
    namespace="${host#*.}"
    namespace="${namespace%%.*}"
    kubectl --context "$CONTEXT" -n "$namespace" port-forward "svc/$service" "$port:${authority##*:}" >"$RUN/port-forward-$service.log" 2>&1 &
    jq --arg from "http://$authority" --arg to "http://host.docker.internal:$port" \
        'walk(if type == "string" then (split($from) | join($to)) else . end)' \
        "$RUN/config/mcp-servers.json" >"$RUN/config/mcp-servers.json.tmp"
    mv "$RUN/config/mcp-servers.json.tmp" "$RUN/config/mcp-servers.json"
    echo "Forwarding $service.$namespace to host.docker.internal:$port"
    port=$((port + 1))
done

if rbac="$(render rbac.yaml 2>/dev/null)"; then
    echo "Creating ServiceAccount $NAMESPACE/$NAME with the agent's RBAC on $CONTEXT"
    kubectl --context "$CONTEXT" create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl --context "$CONTEXT" apply -f - >/dev/null
    { render serviceaccount.yaml; echo "$rbac"; } >"$RUN/identity.yaml"
    kubectl --context "$CONTEXT" -n "$NAMESPACE" apply -f "$RUN/identity.yaml" >/dev/null
    token="$(kubectl --context "$CONTEXT" -n "$NAMESPACE" create token "$NAME" --duration 12h)"
    server="$(kubectl config view --raw --minify --flatten --context "$CONTEXT" -o jsonpath='{.clusters[0].cluster.server}')"
    authority_data="$(kubectl config view --raw --minify --flatten --context "$CONTEXT" -o jsonpath='{.clusters[0].cluster.certificate-authority-data}')"
    (
        umask 077
        cat >"$RUN/data/home/.kube/config" <<EOF
apiVersion: v1
kind: Config
clusters:
  - name: $CONTEXT
    cluster:
      server: $server
      certificate-authority-data: $authority_data
users:
  - name: $NAME
    user:
      token: $token
contexts:
  - name: $CONTEXT
    context:
      cluster: $CONTEXT
      user: $NAME
      namespace: $NAMESPACE
current-context: $CONTEXT
EOF
    )
fi

echo "Starting $AGENT, Ctrl-C to stop"
docker run --rm --init --name "$NAME" \
    --read-only --tmpfs /tmp:exec --tmpfs /workspace:exec,mode=1777 \
    --cap-drop ALL --security-opt no-new-privileges --memory 2g \
    --env-file <(render deployment.yaml | yq '.spec.template.spec.containers[0].env[] | .name + "=" + .value') \
    --env-file <(yq '.secrets | to_entries | .[] | .key + "=" + (.value | tostring)' "$SECRETS") \
    -v "$RUN/data:/data" \
    -v "$RUN/config/CLAUDE.md:/workspace/CLAUDE.md:ro" \
    -v "$RUN/config/settings.json:/etc/claude-code/managed-settings.json:ro" \
    -v "$RUN/config/mcp-servers.json:/etc/switchboard/mcp-servers.json:ro" \
    "$IMAGE"
