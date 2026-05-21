#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DSN='postgres://im:im@localhost:5432/im?sslmode=disable'
SECRET='dev-secret-change-in-production'

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
  local name=$1 port=$2
  local dev_line=""
  if [[ "$name" == "user" || "$name" == "group" ]]; then
    dev_line="  DevMode: true"
  fi
  cat > "$ROOT/apps/$name/api/etc/${name}-api.yaml" <<EOF
Name: ${name}-api
Host: 0.0.0.0
Port: ${port}
Auth:
  AccessSecret: ${SECRET}
  AccessExpire: 604800
${dev_line}
Postgres:
  DSN: ${DSN}
EOF
  append_log "${name}-api" "$ROOT/apps/$name/api/etc/${name}-api.yaml"
}

write_rpc() {
  local name=$1 port=$2
  # 注意：zrpc.RpcServerConf.Auth 是 bool，不能写 Auth.AccessSecret；仅 user.rpc 在 config 里单独声明 Auth 结构体
  cat > "$ROOT/apps/$name/rpc/etc/${name}.yaml" <<EOF
Name: ${name}.rpc
ListenOn: 0.0.0.0:${port}
Postgres:
  DSN: ${DSN}
EOF
  append_log "${name}-rpc" "$ROOT/apps/$name/rpc/etc/${name}.yaml"
}

write_user_rpc() {
  cat > "$ROOT/apps/user/rpc/etc/user.yaml" <<EOF
Name: user.rpc
ListenOn: 0.0.0.0:20100
JwtAuth:
  AccessSecret: ${SECRET}
Postgres:
  DSN: ${DSN}
EOF
  append_log user-rpc "$ROOT/apps/user/rpc/etc/user.yaml"
}

write_api gateway 10000
write_api user 10100
write_api friend 10200
write_api group 10300
write_api conversation 10400
write_api message 10500
write_api notification 10600
write_api push 10700

write_user_rpc
write_rpc friend 20200
write_rpc group 20300
write_rpc conversation 20400
write_rpc message 20500
write_rpc notification 20600
write_rpc push 20700

cat >> "$ROOT/apps/push/rpc/etc/push.yaml" <<EOF
RedisStore:
  Addr: localhost:6379
EOF

# message-api 仅查询历史，发消息走 Gateway WebSocket

cat >> "$ROOT/apps/message/rpc/etc/message.yaml" <<EOF
GroupRpc:
  Endpoints:
    - 127.0.0.1:20300
Kafka:
  Brokers:
    - localhost:9092
RedisStore:
  Addr: localhost:6379
EOF

cat >> "$ROOT/apps/friend/api/etc/friend-api.yaml" <<EOF
ConversationRpc:
  Endpoints:
    - 127.0.0.1:20400
EOF

cat >> "$ROOT/apps/friend/rpc/etc/friend.yaml" <<EOF
ConversationRpc:
  Endpoints:
    - 127.0.0.1:20400
EOF

cat >> "$ROOT/apps/conversation/api/etc/conversation-api.yaml" <<EOF
Redis:
  Addr: localhost:6379
Conversation:
  DirectRecentDays: 0
EOF

cat >> "$ROOT/apps/push/api/etc/push-api.yaml" <<EOF
Redis:
  Addr: localhost:6379
EOF

cat > "$ROOT/apps/cron/etc/cron.yaml" <<EOF
Name: cron
HealthPort: 10800
Postgres:
  DSN: ${DSN}
Redis:
  Addr: localhost:6379
Kafka:
  Brokers:
    - localhost:9092
Cron:
  InboxMergeMs: 100
  OfflineMergeSec: 10
  MemberBatch: 500
EOF

cat > "$ROOT/apps/gateway/api/etc/gateway-api.yaml" <<EOF
Name: gateway-api
Host: 0.0.0.0
Port: 10000
MessageRpc:
  Endpoints:
    - 127.0.0.1:20500
Redis:
  Addr: localhost:6379
Kafka:
  Brokers:
    - localhost:9092
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
append_log gateway-api "$ROOT/apps/gateway/api/etc/gateway-api.yaml"

cat >> "$ROOT/apps/notification/rpc/etc/notification.yaml" <<EOF
Kafka:
  Brokers:
    - localhost:9092
EOF

echo "etc written"
