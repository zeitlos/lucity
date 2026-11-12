#!/usr/bin/env bash
# Puts the vouch demo (COMPARE_REPO, default rolandderroboter/vouch, plus Heroku Postgres) on Heroku, seeds it,
# and saves the ids the capture script needs. Heroku has no buckets, so the
# demo runs without image uploads on every platform.
#
# Usage: scripts/compare/setup-heroku.sh [app-name]
#        scripts/compare/setup-heroku.sh --teardown
# Needs: the Heroku CLI, signed in with `heroku login` in your own terminal.
set -euo pipefail

here="$(cd "$(dirname "$0")" && pwd)"
docs="$(cd "$here/../.." && pwd)"
state="$docs/.shots/compare/heroku.json"
mkdir -p "$(dirname "$state")"

heroku auth:whoami >/dev/null 2>&1 || { echo "Not signed in. Run 'heroku login' in your terminal first." >&2; exit 1; }

if [ "${1:-}" = "--teardown" ]; then
  app="$(node -e 'console.log(require(process.argv[1]).app || "")' "$state")"
  [ -n "$app" ] || { echo "No Heroku app saved in $state" >&2; exit 1; }
  heroku apps:destroy --app "$app" --confirm "$app"
  node -e 'const fs=require("fs");const s=JSON.parse(fs.readFileSync(process.argv[1]));for(const k of ["app","build","addon","appURL"])delete s[k];fs.writeFileSync(process.argv[1],JSON.stringify(s,null,2)+"\n")' "$state"
  exit 0
fi

app="${1:-vouch-compare-$(date +%m%d)}"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

heroku create "$app" --region eu
heroku addons:create heroku-postgresql:essential-0 --app "$app" --wait
heroku config:set PGSSLMODE=no-verify --app "$app"

git clone --quiet "https://github.com/${COMPARE_REPO:-rolandderroboter/vouch}" "$work/vouch"
cd "$work/vouch"
git push --quiet "https://git.heroku.com/$app.git" HEAD:main

url="$(heroku apps:info --app "$app" --json | node -e 'let s="";process.stdin.on("data",d=>s+=d).on("end",()=>console.log(JSON.parse(s).app.web_url.replace(/\/$/,"")))')"
for _ in $(seq 1 60); do
  if curl -fsS -o /dev/null "$url/api/posts"; then break; fi
  sleep 5
done
node "$here/seed.mjs" "$url"

heroku builds --app "$app" --json > "$work/builds.json"
heroku addons --app "$app" --json > "$work/addons.json"
node - "$state" "$app" "$url" "$work" <<'EOF'
const fs = require('fs');
const [statePath, app, appURL, work] = process.argv.slice(2);
const builds = JSON.parse(fs.readFileSync(`${work}/builds.json`, 'utf8'));
const addons = JSON.parse(fs.readFileSync(`${work}/addons.json`, 'utf8'));
const state = fs.existsSync(statePath) ? JSON.parse(fs.readFileSync(statePath, 'utf8')) : {};
Object.assign(state, {
  app,
  build: builds.sort((a, b) => b.created_at.localeCompare(a.created_at))[0]?.id,
  addon: addons.find((addon) => addon.addon_service?.name === 'heroku-postgresql')?.id,
  appURL,
});
fs.writeFileSync(statePath, JSON.stringify(state, null, 2) + '\n');
console.log(state);
EOF
