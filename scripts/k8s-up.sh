#!/usr/bin/env bash
# 本地 K8s 一键部署（kind 推荐）
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

CLUSTER="${K8S_CLUSTER:-kind}"
CLUSTER_NAME="${K8S_CLUSTER_NAME:-im-local}"
NS="im-local"
SKIP_BUILD="${SKIP_BUILD:-0}"

chmod +x scripts/write-k8s-etc.sh scripts/gen-k8s-manifests.sh scripts/k8s-load-images.sh

echo "== 生成 K8s 配置与清单 =="
./scripts/write-k8s-etc.sh
./scripts/gen-k8s-manifests.sh

if [[ "$SKIP_BUILD" != "1" ]]; then
  echo "== 构建应用镜像 =="
  make build-images
fi

if [[ "$CLUSTER" == "kind" ]]; then
  if ! kind get clusters 2>/dev/null | grep -qx "${CLUSTER_NAME}"; then
    echo "== 创建 kind 集群 ${CLUSTER_NAME} =="
    kind create cluster --name "${CLUSTER_NAME}" --config deploy/kind/im-local.yaml
    echo "== 等待节点就绪 =="
    kubectl wait --for=condition=ready node --all --timeout=180s
    sleep 5
  fi
fi

echo "== 载入镜像到集群 =="
K8S_CLUSTER="$CLUSTER" K8S_CLUSTER_NAME="$CLUSTER_NAME" ./scripts/k8s-load-images.sh

echo "== kubectl apply =="
kubectl apply -k deploy/k8s/overlays/local

echo "== 等待基础设施就绪 =="
kubectl -n "$NS" wait --for=condition=ready pod -l app=postgres --timeout=300s
kubectl -n "$NS" wait --for=condition=ready pod -l app=redis --timeout=120s
kubectl -n "$NS" wait --for=condition=ready pod -l app=kafka --timeout=360s

echo ""
echo "已部署到 namespace=${NS}"
echo "  Gateway WS:  ws://localhost:10000/v1/ws  (kind NodePort 30000)"
echo "  User API:    http://localhost:10100"
echo "  Conversation: http://localhost:10400"
echo "  Cron health: http://localhost:10800/health"
echo ""
echo "查看 Pod: kubectl -n ${NS} get pods"
echo "日志:     kubectl -n ${NS} logs -f deploy/gateway-api"
echo "非 kind 集群请: make k8s-forward"
