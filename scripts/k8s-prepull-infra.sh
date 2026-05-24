#!/usr/bin/env bash
# 在本机 Docker 拉取基础设施镜像并载入 kind（避免节点内拉取走失效代理）
set -euo pipefail
CLUSTER_NAME="${K8S_CLUSTER_NAME:-im-local}"

INFRA_IMAGES=(
  postgres:16-alpine
  redis:7-alpine
  apache/rocketmq:5.3.2
)

pull_one() {
  local img=$1
  if docker image inspect "$img" >/dev/null 2>&1; then
    echo "  已有镜像 ${img}"
  else
    echo "  docker pull ${img}"
    docker pull "$img"
  fi
}

echo "== 拉取基础设施镜像（使用本机 Docker 配置）=="
for img in "${INFRA_IMAGES[@]}"; do
  pull_one "$img"
done

if command -v kind >/dev/null 2>&1 && kind get clusters 2>/dev/null | grep -qx "${CLUSTER_NAME}"; then
  echo "== 载入 kind 集群 ${CLUSTER_NAME} =="
  for img in "${INFRA_IMAGES[@]}"; do
    echo "  kind load ${img}"
    kind load docker-image "$img" --name "${CLUSTER_NAME}"
  done
else
  echo "（未检测到 kind 集群 ${CLUSTER_NAME}，跳过 kind load）"
fi

echo "完成"
