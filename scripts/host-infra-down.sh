#!/usr/bin/env bash
# 停止本机 Docker 基础设施（docker compose）
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

COMPOSE_FILE="${COMPOSE_INFRA:-deploy/docker/docker-compose.yml}"

echo "== 停止本机基础设施 =="
docker compose -f "$COMPOSE_FILE" stop

echo "本机基础设施已停止（数据卷 im-mysql-data 保留）"
echo "删除容器: docker compose -f ${COMPOSE_FILE} down"
echo "删除数据: docker volume rm im-mysql-data"
