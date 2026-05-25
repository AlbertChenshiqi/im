#!/usr/bin/env bash
# 清空 MySQL + Redis，并灌入 scripts/dev_reset_seed.sql 测试数据
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

REDIS_ADDR="${REDIS_ADDR:-localhost:6379}"
NS="${K8S_NAMESPACE:-im-local}"

echo "== TRUNCATE + seed MySQL =="
if [[ "${K8S_MODE:-1}" == "1" ]]; then
  kubectl -n "$NS" exec deploy/mysql -- mysql -uim -pim im \
    < "$ROOT/scripts/dev_reset_seed.sql"
elif command -v mysql >/dev/null 2>&1; then
  mysql -uim -pim im < "$ROOT/scripts/dev_reset_seed.sql"
elif docker inspect im-mysql >/dev/null 2>&1; then
  docker exec -i im-mysql mysql -uim -pim im < "$ROOT/scripts/dev_reset_seed.sql"
else
  echo "需要 kubectl 集群（make k8s-seed）或本机 mysql 客户端 / im-mysql 容器"
  exit 1
fi

echo "== FLUSH Redis =="
if [[ "${K8S_MODE:-1}" == "1" ]]; then
  kubectl -n "$NS" exec deploy/redis -- redis-cli FLUSHDB >/dev/null
elif command -v redis-cli >/dev/null 2>&1; then
  redis-cli -h "${REDIS_ADDR%%:*}" -p "${REDIS_ADDR##*:}" FLUSHDB >/dev/null
else
  echo "skip Redis（无 redis-cli；集群内请 K8S_MODE=1 make seed）"
fi
echo "Redis FLUSHDB ok"

echo ""
echo "测试数据已就绪:"
echo "  用户: dev-token user_id=1|2|3"
echo "  群聊 conv_id: group_1"
echo "  私信 conv_id: c2c_1_2 (好友), c2c_1_3 (非好友)"
