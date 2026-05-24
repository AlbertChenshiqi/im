#!/usr/bin/env bash
# 将本地构建的 im/* 镜像载入 kind / minikube 节点
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
# shellcheck source=scripts/im-services.sh
source "$ROOT/scripts/im-services.sh"
TAG="${IMAGE_TAG:-dev}"
CLUSTER="${K8S_CLUSTER:-kind}"
CLUSTER_NAME="${K8S_CLUSTER_NAME:-im-local}"
LOAD_RETRIES="${LOAD_RETRIES:-5}"

# 与 build-images.sh 一致：BINS=message-rpc,gateway-api 仅载入指定镜像（bin 名）
parse_want_bins() {
  WANT_BINS=()
  [[ -n "${BINS:-}" ]] || return 0
  local raw="${BINS//,/ }"
  local name
  for name in $raw; do
    [[ -n "$name" ]] || continue
    WANT_BINS+=("$name")
  done
}

want_image_ref() {
  local ref=$1
  [[ ${#WANT_BINS[@]} -eq 0 ]] && return 0
  local base="${ref##*/}"
  local w
  for w in "${WANT_BINS[@]}"; do
    if [[ "$base" == "$w" ]]; then
      return 0
    fi
  done
  return 1
}

parse_want_bins

image_refs() {
  local spec image
  for spec in "${IM_SERVICE_BUILD_SPECS[@]}"; do
    IFS=':' read -r image _pkg _bin <<<"$spec"
    want_image_ref "$image" || continue
    echo "${image}:${TAG}"
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

  if [[ ${#refs[@]} -eq 0 ]]; then
    echo "未匹配到任何镜像，请检查 BINS（bin 名，如 message-rpc）" >&2
    exit 1
  fi

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
  # 用 RETURN + 展开路径：EXIT 在函数返回后触发时 local archive 已失效，set -u 会报 unbound
  trap "rm -f '${archive}'" RETURN

  echo "== docker save ${#refs[@]} 个镜像 -> 临时归档 =="
  docker save "${refs[@]}" -o "$archive"

  echo "== kind load image-archive（单次导入，避免 containerd 过载）=="
  retry kind load image-archive "$archive" --name "${CLUSTER_NAME}"

  rm -f "$archive"
  trap - RETURN
}

load_kind_one_by_one() {
  wait_kind_containerd
  local spec image ref
  for spec in "${IM_SERVICE_BUILD_SPECS[@]}"; do
    IFS=':' read -r image _pkg _bin <<<"$spec"
    want_image_ref "$image" || continue
    ref="${image}:${TAG}"
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
  local spec image ref
  for spec in "${IM_SERVICE_BUILD_SPECS[@]}"; do
    IFS=':' read -r image _pkg _bin <<<"$spec"
    want_image_ref "$image" || continue
    ref="${image}:${TAG}"
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
