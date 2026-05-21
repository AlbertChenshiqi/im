#!/usr/bin/env bash
# 清空 PostgreSQL + Redis，并灌入 scripts/dev_reset_seed.sql 测试数据
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

DSN="${POSTGRES_DSN:-postgres://im:im@localhost:5432/im?sslmode=disable}"
REDIS_ADDR="${REDIS_ADDR:-localhost:6379}"
NS="${K8S_NAMESPACE:-im-local}"

echo "== TRUNCATE + seed PostgreSQL =="
if [[ "${K8S_MODE:-1}" == "1" ]]; then
  kubectl -n "$NS" exec deploy/postgres -- psql -U im -d im -v ON_ERROR_STOP=1 \
    < "$ROOT/scripts/dev_reset_seed.sql"
elif command -v psql >/dev/null 2>&1; then
  psql "$DSN" -v ON_ERROR_STOP=1 -f "$ROOT/scripts/dev_reset_seed.sql"
else
  echo "需要 kubectl 集群（make seed）或本机 psql（先 make k8s-forward-infra）"
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
echo ""
echo "示例:"
echo "  curl -s -X POST http://localhost:10100/v1/auth/dev-token -H 'Content-Type: application/json' -d '{\"user_id\":1}'"
echo "  curl -s http://localhost:10400/v1/conversations -H \"Authorization: Bearer \$TOKEN\""
