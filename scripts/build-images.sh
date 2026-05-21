#!/usr/bin/env bash
# 构建全部微服务镜像 im/<域>/<服务>:dev（供 kind / 线上 registry 使用）
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

GO_BASE_IMAGE="${GO_BASE_IMAGE:-im-go-base:deps}"
RUNTIME_BASE_IMAGE="${RUNTIME_BASE_IMAGE:-im-go-base:runtime}"
TAG_SUFFIX="${TAG_SUFFIX:-dev}"

SERVICES=(
  "im/gateway/gateway-api:./apps/gateway/api:gateway-api"
  "im/user/user-api:./apps/user/api:user-api"
  "im/user/user-rpc:./apps/user/rpc:user-rpc"
  "im/friend/friend-api:./apps/friend/api:friend-api"
  "im/friend/friend-rpc:./apps/friend/rpc:friend-rpc"
  "im/group/group-api:./apps/group/api:group-api"
  "im/group/group-rpc:./apps/group/rpc:group-rpc"
  "im/conversation/conversation-api:./apps/conversation/api:conversation-api"
  "im/conversation/conversation-rpc:./apps/conversation/rpc:conversation-rpc"
  "im/message/message-api:./apps/message/api:message-api"
  "im/message/message-rpc:./apps/message/rpc:message-rpc"
  "im/notification/notification-api:./apps/notification/api:notification-api"
  "im/notification/notification-rpc:./apps/notification/rpc:notification-rpc"
  "im/push/push-api:./apps/push/api:push-api"
  "im/push/push-rpc:./apps/push/rpc:push-rpc"
  "im/cron/cron:./apps/cron:cron"
)

for spec in "${SERVICES[@]}"; do
  IFS=':' read -r image pkg bin <<<"$spec"
  echo "== build ${image}:${TAG_SUFFIX} (${bin}) =="
  docker build \
    -f deploy/Dockerfile.service \
    --build-arg GO_BASE_IMAGE="${GO_BASE_IMAGE}" \
    --build-arg RUNTIME_BASE_IMAGE="${RUNTIME_BASE_IMAGE}" \
    --build-arg SERVICE_PKG="${pkg}" \
    --build-arg BIN_NAME="${bin}" \
    -t "${image}:${TAG_SUFFIX}" \
    .
done

echo ""
echo "完成。部署: make up  或  make k8s-up"
