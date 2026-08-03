#!/usr/bin/env bash
# TESTRA Production Smoke Test
# Usage: GATEWAY_URL=https://app.testra.io ./scripts/prod-smoke-test.sh

set -euo pipefail

GATEWAY_URL=${GATEWAY_URL:-https://app.testra.io}
HEALTH_PATH="/health"
LOGIN_PATH="/login"

echo "== TESTRA Production Smoke Test =="
echo "Gateway: ${GATEWAY_URL}"

# 1. Ping the outer gateway
echo "[1/3] Pinging outer gateway..."
HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "${GATEWAY_URL}/")
if [[ "${HTTP_STATUS}" -lt 200 || "${HTTP_STATUS}" -ge 400 ]]; then
  echo "FAIL: Gateway returned HTTP ${HTTP_STATUS}"
  exit 1
fi
echo "OK: Gateway HTTP ${HTTP_STATUS}"

# 2. Hit the API gateway health endpoint
echo "[2/3] Checking API /health endpoint..."
HEALTH_BODY=$(curl -sf "${GATEWAY_URL}${HEALTH_PATH}")
if ! echo "${HEALTH_BODY}" | grep -q '"status":"ok"'; then
  echo "FAIL: /health did not return expected status. Body: ${HEALTH_BODY}"
  exit 1
fi
echo "OK: API /health healthy"

# 3. Verify the frontend login route renders without runtime/static errors
echo "[3/3] Checking frontend login route..."
LOGIN_BODY=$(curl -sf "${GATEWAY_URL}${LOGIN_PATH}" || true)
if [[ -z "${LOGIN_BODY}" ]]; then
  echo "FAIL: /login returned an empty body"
  exit 1
fi
if echo "${LOGIN_BODY}" | grep -qiE "Application error|Internal Server Error|__NEXT_DATA__.*\"err|/\\_next/static/.*/error|Error:.* at "; then
  echo "FAIL: /login contains static generation or runtime error markers"
  exit 1
fi
echo "OK: /login rendered cleanly"

echo "== All smoke tests passed =="
