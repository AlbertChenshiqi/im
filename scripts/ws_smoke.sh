#!/usr/bin/env bash
# 手工验收 Gateway WebSocket（需 websocat: brew install websocat）
set -euo pipefail

GW_BASE="${GW_BASE:-ws://localhost:10000/v1/ws}"
USER_API="${USER_API:-http://localhost:10100}"

echo "== dev-token =="
DEV=$(curl -s -X POST "$USER_API/v1/auth/dev-token" \
  -H 'Content-Type: application/json' \
  -d '{"user_id":1}')
TOKEN=$(echo "$DEV" | python3 -c "import sys,json; print(json.load(sys.stdin).get('token',''))")
if [[ -z "$TOKEN" ]]; then
  echo "dev-token failed: $DEV"
  echo "ensure user-api Auth.DevMode=true"
  exit 1
fi
echo "token ok for user_id=1"

if ! command -v websocat >/dev/null 2>&1; then
  echo "install websocat to run interactive WS test"
  exit 0
fi

GW="${GW_BASE}?token=${TOKEN}"
echo "== connect with token in URL =="
(
  sleep 1
  echo '{"type":"ping"}'
  sleep 2
) | websocat -n1 "$GW"

echo "done (see auth_ok / pong above)"
