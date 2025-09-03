#!/usr/bin/env bash
set -euo pipefail

BASE_URL=${BASE_URL:-http://127.0.0.1:8080}

cleanup_pf() {
  [[ -n "${PF_PID:-}" ]] && kill "$PF_PID" || true
}
trap cleanup_pf EXIT

echo "[port-forward] Forwarding gateway-service 80 -> 8080"
kubectl -n default port-forward svc/gateway-service 8080:80 >/tmp/pf2.log 2>&1 &
PF_PID=$!
sleep 3
name="student_$(date +%s)"
telephone="1$(date +%s)"
password="p@ssw0rd"

env_json=$(mktemp)
cat > "$env_json" <<JSON
{
  "id": "smartfox-env",
  "name": "smartfox",
  "values": [
    {"key": "base_url", "value": "$BASE_URL", "enabled": true},
    {"key": "name", "value": "$name", "enabled": true},
    {"key": "telephone", "value": "$telephone", "enabled": true},
    {"key": "password", "value": "$password", "enabled": true}
  ]
}
JSON

npm -g i newman >/dev/null 2>&1 || sudo npm -g i newman

newman run .github/tests/basic.postman_collection.json -e "$env_json" --reporters cli
