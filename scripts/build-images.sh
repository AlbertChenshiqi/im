#!/usr/bin/env bash
# 构建微服务镜像 im/<域>/<服务>:dev（供 kind / 线上 registry 使用）
#
# 全量：./scripts/build-images.sh
# 指定：BINS=message-rpc,gateway-api ./scripts/build-images.sh
# 一键：./scripts/k8s-deploy.sh message-rpc  或  make deploy SVC=message-rpc
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
# shellcheck source=scripts/im-services.sh
source "$ROOT/scripts/im-services.sh"

GO_BASE_IMAGE="${GO_BASE_IMAGE:-im-go-base:deps}"
RUNTIME_BASE_IMAGE="${RUNTIME_BASE_IMAGE:-im-go-base:runtime}"
TAG_SUFFIX="${TAG_SUFFIX:-dev}"

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

want_bin() {
  local bin=$1
  [[ ${#WANT_BINS[@]} -eq 0 ]] && return 0
  local w
  for w in "${WANT_BINS[@]}"; do
    if [[ "$bin" == "$w" ]]; then
      return 0
    fi
  done
  return 1
}

parse_want_bins

built=0
for spec in "${IM_SERVICE_BUILD_SPECS[@]}"; do
  IFS=':' read -r image pkg bin <<<"$spec"
  if ! want_bin "$bin"; then
    continue
  fi
  echo "== build ${image}:${TAG_SUFFIX} (${bin}) =="
  docker build \
    -f deploy/Dockerfile.service \
    --build-arg GO_BASE_IMAGE="${GO_BASE_IMAGE}" \
    --build-arg RUNTIME_BASE_IMAGE="${RUNTIME_BASE_IMAGE}" \
    --build-arg SERVICE_PKG="${pkg}" \
    --build-arg BIN_NAME="${bin}" \
    -t "${image}:${TAG_SUFFIX}" \
    .
  built=$((built + 1))
done

if [[ ${#WANT_BINS[@]} -gt 0 && "$built" -eq 0 ]]; then
  echo "未匹配到任何服务，请检查 BINS。可用 bin 名：" >&2
  im_list_bins | tr '\n' ' ' >&2
  echo >&2
  exit 1
fi

echo ""
if [[ ${#WANT_BINS[@]} -gt 0 ]]; then
  echo "完成（${built} 个镜像）。一键部署: make deploy SVC=${BINS}"
else
  echo "完成（全量 ${built} 个镜像）。首次部署: make up"
fi
