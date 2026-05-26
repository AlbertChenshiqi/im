#!/usr/bin/env bash
# 生成 K8s 本地 overlay 用配置（Service DNS：mysql / redis / rocketmq-namesrv / *-rpc）
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUT="$ROOT/deploy/k8s/overlays/local/config"
rm -rf "$OUT"
mkdir -p "$OUT"/{gateway,user,friend,group,conversation,message,notification,transfer}

SECRET='dev-secret-change-in-production'

if [[ "${HOST_INFRA:-0}" == "1" ]]; then
  HOST="${HOST_INFRA_ADDR:-host.docker.internal}"
  DSN="im:im@tcp(${HOST}:3306)/im?parseTime=true&charset=utf8mb4&loc=Local"
  REDIS="${HOST}:6379"
  ROCKETMQ="${HOST}:9876"
else
  DSN="im:im@tcp(mysql:3306)/im?parseTime=true&charset=utf8mb4&loc=Local"
  REDIS='redis:6379'
  ROCKETMQ='rocketmq-namesrv:9876'
fi

append_log() {
  local svc=$1 file=$2
  cat >>"$file" <<EOF
Log:
  ServiceName: ${svc}
  Mode: console
  Level: info
  Stat: false
EOF
}

write_api() {
  local domain=$1 port=$2
  local dev_line=""
  if [[ "$domain" == "user" || "$domain" == "group" ]]; then
    dev_line="  DevMode: true"
  fi
  local file="$OUT/${domain}/${domain}-api.yaml"
  cat >"$file" <<EOF
Name: ${domain}-api
Host: 0.0.0.0
Port: ${port}
Auth:
  AccessSecret: ${SECRET}
  AccessExpire: 604800
${dev_line}
MySQL:
  DSN: im:im@tcp(mysql:3306)/im?parseTime=true&charset=utf8mb4&loc=Local
EOF
  append_log "${domain}-api" "$file"
}

write_rpc() {
  local domain=$1 port=$2
  local file="$OUT/${domain}/${domain}-rpc.yaml"
  cat >"$file" <<EOF
Name: ${domain}.rpc
ListenOn: 0.0.0.0:${port}
MySQL:
  DSN: im:im@tcp(mysql:3306)/im?parseTime=true&charset=utf8mb4&loc=Local
EOF
  append_log "${domain}-rpc" "$file"
}

write_api gateway 10000
write_api user 10100
write_api friend 10200
write_api group 10300
write_api conversation 10400
write_api message 10500
write_api notification 10600

cat >"$OUT/user/user-rpc.yaml" <<EOF
Name: user.rpc
ListenOn: 0.0.0.0:20100
JwtAuth:
  AccessSecret: ${SECRET}
MySQL:
  DSN: im:im@tcp(mysql:3306)/im?parseTime=true&charset=utf8mb4&loc=Local
EOF
append_log user-rpc "$OUT/user/user-rpc.yaml"

write_rpc friend 20200
write_rpc group 20300
write_rpc conversation 20400
write_rpc message 20500
write_rpc notification 20600

cat >>"$OUT/user/user-api.yaml" <<EOF
Redis:
  Addr: ${REDIS}
OnlineTTLSeconds: 300
EOF

cat >>"$OUT/message/message-rpc.yaml" <<EOF
GroupRpc:
  Endpoints:
    - group-rpc:20300
RocketMQ:
  NameServer:
    - ${ROCKETMQ}
RedisStore:
  Addr: ${REDIS}
EOF

cat >>"$OUT/friend/friend-api.yaml" <<EOF
ConversationRpc:
  Endpoints:
    - conversation-rpc:20400
EOF

cat >>"$OUT/friend/friend-rpc.yaml" <<EOF
ConversationRpc:
  Endpoints:
    - conversation-rpc:20400
EOF

cat >>"$OUT/conversation/conversation-api.yaml" <<EOF
Redis:
  Addr: ${REDIS}
Conversation:
  DirectRecentDays: 0
EOF

cat >"$OUT/transfer/transfer.yaml" <<EOF
Name: transfer
HealthPort: 10800
MySQL:
  DSN: im:im@tcp(mysql:3306)/im?parseTime=true&charset=utf8mb4&loc=Local
Redis:
  Addr: ${REDIS}
RocketMQ:
  NameServer:
    - ${ROCKETMQ}
Transfer:
  InboxMergeMs: 100
  OfflineMergeSec: 10
  MemberBatch: 500
EOF

cat >"$OUT/gateway/gateway-api.yaml" <<EOF
Name: gateway-api
Host: 0.0.0.0
Port: 10000
MessageRpc:
  Endpoints:
    - message-rpc:20500
Redis:
  Addr: ${REDIS}
RocketMQ:
  NameServer:
    - ${ROCKETMQ}
WebSocket:
  OnlineTTL: 300
  HeartbeatInterval: 60
  HeartbeatMaxMiss: 3
  MaxMessageBytes: 65536
  ConnectionMode: multi
  AllowedOrigins:
    - "*"
Auth:
  AccessSecret: ${SECRET}
EOF
append_log gateway-api "$OUT/gateway/gateway-api.yaml"

cat >>"$OUT/notification/notification-rpc.yaml" <<EOF
RocketMQ:
  NameServer:
    - ${ROCKETMQ}
EOF

MIG="$ROOT/deploy/k8s/overlays/local/migrations"
mkdir -p "$MIG"
cp "$ROOT/migrations/001_init.sql" "$MIG/"

KUST="$ROOT/deploy/k8s/overlays/local/kustomization.yaml"
{
  echo "apiVersion: kustomize.config.k8s.io/v1beta1"
  echo "kind: Kustomization"
  echo ""
  echo "namespace: im-local"
  echo ""
  echo "resources:"
  echo "  - namespace.yaml"
  if [[ "${HOST_INFRA:-0}" != "1" ]]; then
    echo "  - infra"
  fi
  for mod in gateway user friend group conversation message notification transfer; do
    echo "  - apps/${mod}"
  done
  if [[ "${HOST_INFRA:-0}" != "1" ]]; then
    echo ""
    echo "configMapGenerator:"
    echo "  - name: mysql-init"
    echo "    files:"
    echo "      - 001_init.sql=migrations/001_init.sql"
    echo ""
    echo "generatorOptions:"
    echo "  disableNameSuffixHash: true"
  fi
} >"$KUST"

echo "k8s config written to $OUT (HOST_INFRA=${HOST_INFRA:-0})"
