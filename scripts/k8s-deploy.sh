#!/usr/bin/env bash
# 指定服务：构建镜像 → 载入 kind/minikube → 调副本（可选）→ 滚动发布
#
# 用法：
#   ./scripts/k8s-deploy.sh message-rpc
#   ./scripts/k8s-deploy.sh gateway-api:2 message-rpc:1 cron
#   make deploy SVC=gateway-api:2,message-rpc
#
# 环境变量：
#   SVC / BINS   服务列表（与 positional 二选一，positional 优先）
#   REPLICAS     未写 :N 时的默认副本数
#   SKIP_BUILD=1 SKIP_LOAD=1 SKIP_ROLLOUT=1
#   K8S_NAMESPACE（默认 im-local）ROLLOUT_TIMEOUT（默认 180s）
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
# shellcheck source=scripts/im-services.sh
source "$ROOT/scripts/im-services.sh"

NS="${K8S_NAMESPACE:-im-local}"
ROLLOUT_TIMEOUT="${ROLLOUT_TIMEOUT:-180s}"
SKIP_BUILD="${SKIP_BUILD:-0}"
SKIP_LOAD="${SKIP_LOAD:-0}"
SKIP_ROLLOUT="${SKIP_ROLLOUT:-0}"

usage() {
  cat <<EOF
用法: $0 <服务[:副本]> [更多服务...]

  服务名为 Deployment/bin 名，例如 message-rpc、gateway-api。
  副本省略时不改 scale，仅滚动重启；可用 REPLICAS=2 作为默认副本。

示例:
  $0 message-rpc
  $0 gateway-api:2 message-rpc:1
  make deploy SVC=gateway-api:2,cron

可用服务:
$(im_list_bins | sed 's/^/  /')

环境变量: SKIP_BUILD SKIP_LOAD SKIP_ROLLOUT K8S_NAMESPACE ROLLOUT_TIMEOUT
EOF
}

# deploy_bins[i] deploy_replicas[i]（空字符串表示不调整副本）
declare -a deploy_bins=()
declare -a deploy_replicas=()

add_target() {
  local item=$1 bin replicas
  [[ -n "$item" ]] || return 0
  if [[ "$item" == *:* ]]; then
    bin="${item%%:*}"
    replicas="${item#*:}"
  else
    bin="$item"
    replicas="${REPLICAS:-}"
  fi
  if ! im_valid_bin "$bin"; then
    echo "未知服务: ${bin}" >&2
    echo "可用: $(im_bins_joined)" >&2
    exit 1
  fi
  if [[ -n "$replicas" ]] && ! [[ "$replicas" =~ ^[0-9]+$ ]]; then
    echo "无效副本数: ${replicas}（服务 ${bin}）" >&2
    exit 1
  fi
  deploy_bins+=("$bin")
  deploy_replicas+=("$replicas")
}

parse_input() {
  if [[ $# -gt 0 ]]; then
    local arg
    for arg in "$@"; do
      [[ "$arg" == "-h" || "$arg" == "--help" ]] && { usage; exit 0; }
      add_target "$arg"
    done
    return 0
  fi
  local raw="${SVC:-${BINS:-}}"
  if [[ -z "$raw" ]]; then
    usage >&2
    exit 1
  fi
  raw="${raw//,/ }"
  local item
  for item in $raw; do
    add_target "$item"
  done
}

parse_input "$@"

if [[ ${#deploy_bins[@]} -eq 0 ]]; then
  echo "请至少指定一个服务" >&2
  exit 1
fi

# 供 build-images / k8s-load-images 使用
BINS=$(IFS=,; echo "${deploy_bins[*]}")
export BINS

echo "== 部署目标 (namespace=${NS}) =="
local_i=0
while [[ $local_i -lt ${#deploy_bins[@]} ]]; do
  rep="${deploy_replicas[$local_i]}"
  if [[ -n "$rep" ]]; then
    echo "  ${deploy_bins[$local_i]} (replicas=${rep})"
  else
    echo "  ${deploy_bins[$local_i]} (replicas 不变)"
  fi
  local_i=$((local_i + 1))
done

ensure_base_images() {
  local base runtime
  base="${GO_BASE_IMAGE:-im-go-base:deps}"
  runtime="${RUNTIME_BASE_IMAGE:-im-go-base:runtime}"
  docker image inspect "$base" >/dev/null 2>&1 || {
    echo "未找到 ${base}，先执行: make image-base-build" >&2
    exit 1
  }
  docker image inspect "$runtime" >/dev/null 2>&1 || {
    echo "未找到 ${runtime}，先执行: make image-base-build" >&2
    exit 1
  }
}

if [[ "$SKIP_BUILD" != "1" ]]; then
  echo ""
  echo "== [1/3] 构建镜像 (${BINS}) =="
  ensure_base_images
  chmod +x scripts/build-images.sh
  GO_BASE_IMAGE="${GO_BASE_IMAGE:-im-go-base:deps}" \
    RUNTIME_BASE_IMAGE="${RUNTIME_BASE_IMAGE:-im-go-base:runtime}" \
    BINS="$BINS" ./scripts/build-images.sh
else
  echo ""
  echo "== [1/3] 跳过构建 (SKIP_BUILD=1) =="
fi

if [[ "$SKIP_LOAD" != "1" ]]; then
  echo ""
  echo "== [2/3] 载入集群 (${BINS}) =="
  chmod +x scripts/k8s-load-images.sh
  BINS="$BINS" ./scripts/k8s-load-images.sh
else
  echo ""
  echo "== [2/3] 跳过载入 (SKIP_LOAD=1) =="
fi

if [[ "$SKIP_ROLLOUT" != "1" ]]; then
  echo ""
  echo "== [3/3] 滚动发布 =="
  local_i=0
  while [[ $local_i -lt ${#deploy_bins[@]} ]]; do
    bin="${deploy_bins[$local_i]}"
    rep="${deploy_replicas[$local_i]}"
    if ! kubectl -n "$NS" get deployment "$bin" >/dev/null 2>&1; then
      echo "Deployment ${bin} 不存在于 namespace ${NS}，请先 make up" >&2
      exit 1
    fi
    if [[ -n "$rep" ]]; then
      echo "-- scale deployment/${bin} --replicas=${rep}"
      kubectl -n "$NS" scale "deployment/${bin}" --replicas="$rep"
    fi
    echo "-- rollout restart deployment/${bin}"
    kubectl -n "$NS" rollout restart "deployment/${bin}"
    kubectl -n "$NS" rollout status "deployment/${bin}" --timeout="$ROLLOUT_TIMEOUT"
    local_i=$((local_i + 1))
  done
else
  echo ""
  echo "== [3/3] 跳过滚动发布 (SKIP_ROLLOUT=1) =="
fi

echo ""
echo "完成: ${deploy_bins[*]}"
