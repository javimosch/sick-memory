#!/usr/bin/env bash
set -euo pipefail

# test-am-cloud.sh — verify am-cloud admin API accessibility
#
# Confirms the am-cloud panel API is reachable and the worker token
# authenticates correctly. Returns 0 on success, non-zero on failure.

API_BASE="${AM_API_URL:-https://automaintainer.intrane.fr}"
TOKEN="${AM_WORKER_TOKEN:-}"
REPO="${AM_REPO:-71a96b}"

if [ -z "$TOKEN" ]; then
  echo "FAIL: AM_WORKER_TOKEN is not set" >&2
  exit 1
fi

echo "Connecting to am-cloud admin API at $API_BASE ..."

# Test: read memories (expect: empty array or valid JSON)
resp=$(curl -s -w "\n%{http_code}" \
  "$API_BASE/api/agent/memories?repo=$REPO&token=$TOKEN" \
  --connect-timeout 10 --max-time 15)

http_code=$(echo "$resp" | tail -1)
body=$(echo "$resp" | sed '$d')

if [ "$http_code" != "200" ]; then
  echo "FAIL: HTTP $http_code — $body" >&2
  exit 1
fi

echo "OK: HTTP 200 — API accessible and authenticated"
echo "Response: $body"
echo ""
echo "am-cloud admin access: CONFIRMED"
