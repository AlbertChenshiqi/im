#!/usr/bin/env bash
# 在本机 Docker 启动 Postgres / Redis / RocketMQ（docker compose）
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

COMPOSE_FILE="${COMPOSE_INFRA:-deploy/docker/docker-compose.yml}"
# 本机 go run 默认 127.0.0.1；kind Pod 经宿主机访问时: HOST_INFRA_ADDR=host.docker.internal
BROKER_IP="${HOST_INFRA_ADDR:-127.0.0.1}"
BROKER_RUNTIME="$ROOT/deploy/docker/broker.runtime.conf"

sed "s/brokerIP1 = host.docker.internal/brokerIP1 = ${BROKER_IP}/" \
  "$ROOT/deploy/docker/broker.conf" >"$BROKER_RUNTIME"
export BROKER_CONF="$BROKER_RUNTIME"

docker volume create im-postgres-data >/dev/null 2>&1 || true

INFRA_CONTAINERS=(im-postgres im-redis im-rocketmq-namesrv im-rocketmq-broker im-rocketmq-dashboard)

# 固定 container_name 的残留容器会阻塞 compose
for c in "${INFRA_CONTAINERS[@]}"; do
  if ! docker inspect "$c" >/dev/null 2>&1; then
    continue
  fi
  project="$(docker inspect "$c" --format '{{index .Config.Labels "com.docker.compose.project"}}' 2>/dev/null || true)"
  if [[ "$project" != "im-infra" ]]; then
    echo "移除未由 compose 管理的旧容器 ${c} ..."
    docker rm -f "$c" >/dev/null
  fi
done

# 旧版 host-infra-up.sh 手动创建的 im-infra 网络会与 compose 冲突
if docker network inspect im-infra >/dev/null 2>&1; then
  compose_net="$(docker network inspect im-infra --format '{{index .Labels "com.docker.compose.network"}}' 2>/dev/null || true)"
  if [[ "$compose_net" != "default" ]]; then
    echo "移除未由 compose 管理的旧网络 im-infra ..."
    docker network rm im-infra 2>/dev/null || {
      echo "错误: 无法删除网络 im-infra（仍有容器连接）。请先: docker rm -f im-postgres im-redis im-rocketmq-namesrv im-rocketmq-broker im-rocketmq-dashboard && docker network rm im-infra" >&2
      exit 1
    }
  fi
fi

# kind 默认配置会把 5432/6379/9876 映射到 control-plane，与本机 compose 冲突
kind_cp="im-local-control-plane"
if docker inspect "$kind_cp" >/dev/null 2>&1; then
  if docker port "$kind_cp" 2>/dev/null | grep -qE '0\.0\.0\.0:(5432|6379|9876)$'; then
    echo "错误: kind 集群 im-local 已占用 5432/6379/9876，无法同时启动本机 compose 基础设施。" >&2
    echo "  若要用本机 Docker 跑 Postgres/Redis/RocketMQ:" >&2
    echo "    kind delete cluster --name im-local" >&2
    echo "    HOST_INFRA=1 make up   # 使用 im-local-host-infra.yaml，不再映射上述端口" >&2
    echo "  若继续用集群内基础设施: 无需 make host-infra-up" >&2
    exit 1
  fi
fi

echo "== 启动本机基础设施（${COMPOSE_FILE}）=="
docker compose -f "$COMPOSE_FILE" up -d --wait

# NameServer 会缓存 broker 注册地址；brokerIP1 变更后需重启，否则本机 go run 仍解析到 host.docker.internal
if docker inspect im-rocketmq-namesrv >/dev/null 2>&1; then
  registered=$(
    docker exec im-rocketmq-namesrv sh -c \
      'cd /home/rocketmq/rocketmq-5.3.2 2>/dev/null || cd /home/rocketmq/rocketmq-5.3.1 2>/dev/null || exit 0; bin/mqadmin clusterList -n localhost:9876 2>/dev/null' \
      | awk '/broker-a/ {print $6; exit}'
  )
  expect="${BROKER_IP}:10911"
  if [[ -n "$registered" && "$registered" != "$expect" ]]; then
    echo "RocketMQ broker 注册为 ${registered}，期望 ${expect}，重启 namesrv/broker ..."
    docker compose -f "$COMPOSE_FILE" restart rocketmq-namesrv rocketmq-broker
    sleep 12
  fi
fi

echo ""
echo "本机基础设施已就绪（kind Pod 经 ${BROKER_IP} 访问）"
echo "  Postgres:  localhost:5432  (im/im, db=im)"
echo "  Redis:     localhost:6379"
echo "  RocketMQ:  localhost:9876"
echo "  RMQ 看板:  http://localhost:8082"
echo ""
echo "停止: make host-infra-down"
