#!/usr/bin/env bash
# 将本地构建的 im/* 镜像载入 kind / minikube 节点
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
TAG="${IMAGE_TAG:-dev}"
CLUSTER="${K8S_CLUSTER:-kind}"
CLUSTER_NAME="${K8S_CLUSTER_NAME:-im-local}"
LOAD_RETRIES="${LOAD_RETRIES:-5}"

SERVICES=(
  "im/gateway/gateway-api"
  "im/user/user-api"
  "im/user/user-rpc"
  "im/friend/friend-api"
  "im/friend/friend-rpc"
  "im/group/group-api"
  "im/group/group-rpc"
  "im/conversation/conversation-api"
  "im/conversation/conversation-rpc"
  "im/message/message-api"
  "im/message/message-rpc"
  "im/notification/notification-api"
  "im/notification/notification-rpc"
  "im/push/push-api"
  "im/push/push-rpc"
  "im/cron/cron"
)

image_refs() {
  local ref
  for ref in "${SERVICES[@]}"; do
    echo "${ref}:${TAG}"
  done
}

wait_kind_containerd() {
  local node="${CLUSTER_NAME}-control-plane"
  echo "== 等待 kind 节点 containerd 就绪 (${node}) =="
  local i
  for ((i = 1; i <= 60; i++)); do
    if docker exec "$node" ctr -n k8s.io version >/dev/null 2>&1; then
      echo "containerd 已就绪"
      return 0
    fi
    sleep 2
  done
  echo "containerd 未就绪，可尝试: kind delete cluster --name ${CLUSTER_NAME} && make up" >&2
  return 1
}

retry() {
  local n attempt=1 delay=3
  n=$LOAD_RETRIES
  while true; do
    if "$@"; then
      return 0
    fi
    if ((attempt >= n)); then
      return 1
    fi
    echo "重试 ${attempt}/${n}（${delay}s 后）..."
    sleep "$delay"
    attempt=$((attempt + 1))
    delay=$((delay + 2))
  done
}

load_kind_archive() {
  wait_kind_containerd

  local -a refs=()
  while IFS= read -r line; do
    refs+=("$line")
  done < <(image_refs)

  echo "== 校验本地镜像 =="
  local ref
  for ref in "${refs[@]}"; do
    docker image inspect "$ref" >/dev/null || {
      echo "缺少镜像 ${ref}，请先 make build-images" >&2
      exit 1
    }
  done

  local archive
  archive=$(mktemp -t im-kind-images.XXXXXX.tar)
  trap 'rm -f "$archive"' EXIT

  echo "== docker save ${#refs[@]} 个镜像 -> 临时归档 =="
  docker save "${refs[@]}" -o "$archive"

  echo "== kind load image-archive（单次导入，避免 containerd 过载）=="
  retry kind load image-archive "$archive" --name "${CLUSTER_NAME}"
}

load_kind_one_by_one() {
  wait_kind_containerd
  local spec ref
  for spec in "${SERVICES[@]}"; do
    ref="${spec}:${TAG}"
    echo "== kind load ${ref} =="
    docker image inspect "$ref" >/dev/null
    retry kind load docker-image "$ref" --name "${CLUSTER_NAME}"
    sleep 1
  done
}

load_kind() {
  if [[ "${KIND_LOAD_MODE:-archive}" == "single" ]]; then
    load_kind_one_by_one
  else
    load_kind_archive
  fi
}

load_minikube() {
  local spec ref
  for spec in "${SERVICES[@]}"; do
    ref="${spec}:${TAG}"
    echo "== minikube load ${ref} =="
    docker image inspect "$ref" >/dev/null
    retry minikube image load "$ref" -p "${CLUSTER_NAME}"
  done
}

case "$CLUSTER" in
  kind)
    command -v kind >/dev/null || { echo "需要安装 kind: https://kind.sigs.k8s.io/"; exit 1; }
    load_kind
    ;;
  minikube)
    command -v minikube >/dev/null || { echo "需要安装 minikube"; exit 1; }
    load_minikube
    ;;
  *)
    echo "K8S_CLUSTER 仅支持 kind 或 minikube，当前: $CLUSTER" >&2
    exit 1
    ;;
esac

echo "镜像已载入 ${CLUSTER}/${CLUSTER_NAME}"
