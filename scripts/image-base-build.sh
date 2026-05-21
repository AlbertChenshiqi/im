#!/usr/bin/env bash
# 构建 im-go-base 基础镜像（供 K8s 本地镜像构建使用）
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

GO_VERSION="${GO_VERSION:-1.24}"
TOOLCHAIN_TAG="im-go-base:${GO_VERSION}"
DEPS_TAG="im-go-base:deps"
RUNTIME_TAG="im-go-base:runtime"

echo "== 1/3 工具链镜像 ${TOOLCHAIN_TAG} =="
docker build \
  --target toolchain \
  -t "${TOOLCHAIN_TAG}" \
  -f deploy/Dockerfile.base \
  .

echo "== 2/3 依赖缓存镜像 ${DEPS_TAG} =="
docker build \
  --target deps \
  -t "${DEPS_TAG}" \
  -f deploy/Dockerfile.base \
  .

echo "== 3/3 运行层镜像 ${RUNTIME_TAG} =="
docker build \
  --target runtime \
  -t "${RUNTIME_TAG}" \
  -f deploy/Dockerfile.base \
  .

echo ""
echo "完成。后续: make build-images"
docker images im-go-base --format 'table {{.Repository}}:{{.Tag}}\t{{.Size}}'
