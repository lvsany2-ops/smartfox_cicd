#!/usr/bin/env bash
set -euo pipefail

BASE_URL=http://127.0.0.1:8080

cleanup_pf() {
  [[ -n "${PF_PID:-}" ]] && kill "$PF_PID" || true
}
trap cleanup_pf EXIT

echo "[port-forward] Forwarding gateway-service 80 -> 8080"
kubectl -n default port-forward svc/gateway-service 8080:80 >/tmp/pf.log 2>&1 &
PF_PID=$!
sleep 3

echo "[health] GET /health"
curl -fsS "$BASE_URL/health" | grep -qi 'OK'

epoch=$(date +%s) # 10 digits
name="student_${epoch}"
phone="1${epoch}" # 11 digits total
pass="p@ssw0rd"

echo "[register] POST /api/auth/register"
reg=$(curl -fsS -X POST "$BASE_URL/api/auth/register" \
  -H 'Content-Type: application/json' \
  -d "{\"name\":\"$name\",\"telephone\":\"$phone\",\"password\":\"$pass\",\"role\":\"student\"}")
echo "$reg" | grep -q '"code":\s*200'

echo "[login] POST /api/auth/login"
login=$(curl -fsS -X POST "$BASE_URL/api/auth/login" \
  -H 'Content-Type: application/json' \
  -d "{\"telephone\":\"$phone\",\"password\":\"$pass\"}")
token=$(echo "$login" | jq -r '.data.token')
[[ "$token" != "null" && -n "$token" ]]

echo "[profile] GET /api/auth/profile"
curl -fsS -H "Authorization: $token" "$BASE_URL/api/auth/profile" | grep -q '"user_id"'

echo "[experiments] GET /api/student/experiments"
curl -fsS -H "Authorization: $token" "$BASE_URL/api/student/experiments?page=1&limit=10&status=all" | grep -q '"status":\s*"success"'

echo "E2E smoke passed"
