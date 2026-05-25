#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

export MYSQL_DSN="im:im@tcp(localhost:3306)/im?parseTime=true&charset=utf8mb4&loc=Local"
export REDIS_ADDR="${REDIS_ADDR:-localhost:6379}"
export ROCKETMQ_NAMESERVER="${ROCKETMQ_NAMESERVER:-localhost:9876}"
export JWT_SECRET="${JWT_SECRET:-dev-secret-change-in-production}"

mkdir -p bin
make build

run() {
  local name=$1
  local bin=$2
  local cfg=$3
  if [[ -n "${4:-}" ]] && lsof -i ":$4" >/dev/null 2>&1; then
    echo "skip $name (port $4 in use)"
    return
  fi
  "./bin/$bin" -f "$cfg" &
  echo "$name pid=$!"
}

run user-rpc user-rpc apps/user/rpc/etc/user.yaml 20100
run friend-rpc friend-rpc apps/friend/rpc/etc/friend.yaml 20200
run group-rpc group-rpc apps/group/rpc/etc/group.yaml 20300
run conversation-rpc conversation-rpc apps/conversation/rpc/etc/conversation.yaml 20400
run message-rpc message-rpc apps/message/rpc/etc/message.yaml 20500
run notification-rpc notification-rpc apps/notification/rpc/etc/notification.yaml 20600
run push-rpc push-rpc apps/push/rpc/etc/push.yaml 20700

sleep 2

run gateway gateway-api apps/gateway/api/etc/gateway-api.yaml 10000
run user-api user-api apps/user/api/etc/user-api.yaml 10100
run friend-api friend-api apps/friend/api/etc/friend-api.yaml 10200
run group-api group-api apps/group/api/etc/group-api.yaml 10300
run conversation-api conversation-api apps/conversation/api/etc/conversation-api.yaml 10400
run message-api message-api apps/message/api/etc/message-api.yaml 10500
run notification-api notification-api apps/notification/api/etc/notification-api.yaml 10600
run push-api push-api apps/push/api/etc/push-api.yaml 10700

run cron cron apps/cron/etc/cron.yaml 10800

echo "All services started."
echo "Gateway WS: ws://localhost:10000/gateway/v1/ws"
echo "User API:   http://localhost:10100"
