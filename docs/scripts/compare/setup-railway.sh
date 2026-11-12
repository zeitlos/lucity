#!/usr/bin/env bash
# Puts the vouch demo (web service from COMPARE_REPO, default rolandderroboter/vouch, plus Postgres) on Railway,
# seeds it, and saves the ids the capture script needs.
#
# Usage: scripts/compare/setup-railway.sh [project-name]
#        scripts/compare/setup-railway.sh --teardown
# Needs: `railway login` done in your own terminal, and Railway's GitHub app
# allowed on the repo.
set -euo pipefail

here="$(cd "$(dirname "$0")" && pwd)"
docs="$(cd "$here/../.." && pwd)"
state="$docs/.shots/compare/railway.json"
mkdir -p "$(dirname "$state")"

railway whoami >/dev/null 2>&1 || { echo "Not signed in. Run 'railway login' in your terminal first." >&2; exit 1; }

if [ "${1:-}" = "--teardown" ]; then
  project="$(node -e 'console.log(require(process.argv[1]).project || "")' "$state")"
  [ -n "$project" ] || { echo "No Railway project saved in $state" >&2; exit 1; }
  railway delete --project "$project" --yes
  node -e 'const fs=require("fs");const s=JSON.parse(fs.readFileSync(process.argv[1]));for(const k of ["project","environment","service","database","deployment","appURL"])delete s[k];fs.writeFileSync(process.argv[1],JSON.stringify(s,null,2)+"\n")' "$state"
  exit 0
fi

repo="${COMPARE_REPO:-rolandderroboter/vouch}"
name="${1:-vouch-compare}"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
cd "$work"

railway init --name "$name" --json
railway add --database postgres --json
railway add --repo "$repo" --service vouch --variables 'DATABASE_URL=${{Postgres.DATABASE_URL}}' --json
railway domain --service vouch --json || true

echo "Waiting for the first deployment..."
railway status --json > status.json
url="$(railway domain list --service vouch --json | node -e 'let s="";process.stdin.on("data",d=>s+=d).on("end",()=>{const j=JSON.parse(s);const all=JSON.stringify(j).match(/[a-z0-9-]+\.up\.railway\.app/);console.log(all?"https://"+all[0]:"")})')"
[ -n "$url" ] || { echo "Could not read the service domain, see 'railway domain list --service vouch'" >&2; exit 1; }

for _ in $(seq 1 120); do
  if curl -fsS -o /dev/null "$url/api/posts"; then break; fi
  sleep 5
done
node "$here/seed.mjs" "$url"

railway deployment list --service vouch --json > deployments.json
node - "$state" "$url" <<'EOF'
const fs = require('fs');
const [statePath, appURL] = process.argv.slice(2);
const status = JSON.parse(fs.readFileSync('status.json', 'utf8'));
const deployments = JSON.parse(fs.readFileSync('deployments.json', 'utf8'));
const nodes = (value) => (value?.edges ?? value ?? []).map((edge) => edge.node ?? edge);
const services = nodes(status.services);
const find = (pattern) => services.find((service) => pattern.test(service.name))?.id;
const deploymentList = Array.isArray(deployments) ? deployments : nodes(deployments.deployments);
const state = fs.existsSync(statePath) ? JSON.parse(fs.readFileSync(statePath, 'utf8')) : {};
Object.assign(state, {
  project: status.id,
  environment: nodes(status.environments).find((environment) => environment.name === 'production')?.id ?? nodes(status.environments)[0]?.id,
  service: find(/^vouch$/),
  database: find(/postgres/i),
  deployment: deploymentList[0]?.id,
  appURL,
});
fs.writeFileSync(statePath, JSON.stringify(state, null, 2) + '\n');
console.log(state);
EOF
