#!/usr/bin/env bash
# 本地 K8s 一键部署（kind 推荐）
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

CLUSTER="${K8S_CLUSTER:-kind}"
CLUSTER_NAME="${K8S_CLUSTER_NAME:-im-local}"
NS="im-local"
SKIP_BUILD="${SKIP_BUILD:-0}"
PREPULL_INFRA="${PREPULL_INFRA:-1}"
HOST_INFRA="${HOST_INFRA:-0}"

chmod +x scripts/write-k8s-etc.sh scripts/gen-k8s-manifests.sh scripts/k8s-load-images.sh scripts/k8s-prepull-infra.sh scripts/host-infra-up.sh

echo "== 生成 K8s 配置与清单 =="
HOST_INFRA="$HOST_INFRA" ./scripts/write-k8s-etc.sh
HOST_INFRA="$HOST_INFRA" ./scripts/gen-k8s-manifests.sh

if [[ "$SKIP_BUILD" != "1" ]]; then
  echo "== 构建应用镜像 =="
  make build-images
fi

if [[ "$CLUSTER" == "kind" ]]; then
  KIND_CONFIG="deploy/kind/im-local.yaml"
  if [[ "$HOST_INFRA" == "1" ]]; then
    KIND_CONFIG="deploy/kind/im-local-host-infra.yaml"
  fi
  if ! kind get clusters 2>/dev/null | grep -qx "${CLUSTER_NAME}"; then
    echo "== 创建 kind 集群 ${CLUSTER_NAME}（${KIND_CONFIG}）=="
    kind create cluster --name "${CLUSTER_NAME}" --config "$KIND_CONFIG"
    echo "== 等待节点就绪 =="
    kubectl wait --for=condition=ready node --all --timeout=180s
    sleep 5
  elif [[ "$HOST_INFRA" == "1" ]] && docker port "${CLUSTER_NAME}-control-plane" 2>/dev/null | grep -q '5432/tcp'; then
    echo "警告: 当前 kind 集群仍映射 5432/6379/9876，与本机 Docker 基础设施冲突" >&2
    echo "  请执行: kind delete cluster --name ${CLUSTER_NAME} && HOST_INFRA=1 make up" >&2
  fi
fi

if [[ "$HOST_INFRA" == "1" ]]; then
  echo "== 清理集群内基础设施（HOST_INFRA=1）=="
  kubectl -n "$NS" delete deployment mysql redis rocketmq-namesrv rocketmq-broker --ignore-not-found --wait=true --timeout=120s 2>/dev/null || \
    kubectl -n "$NS" delete deployment mysql redis rocketmq-namesrv rocketmq-broker --ignore-not-found --wait=false
  kubectl -n "$NS" delete svc mysql redis rocketmq-namesrv rocketmq-broker --ignore-not-found
  echo "== 启动本机基础设施（跳过 kind load 基础设施镜像）=="
  HOST_INFRA_ADDR="${HOST_INFRA_ADDR:-host.docker.internal}" ./scripts/host-infra-up.sh
elif [[ "$CLUSTER" == "kind" && "$PREPULL_INFRA" == "1" ]]; then
  echo "== 预拉取基础设施镜像（MySQL / Redis / RocketMQ）=="
  K8S_CLUSTER_NAME="$CLUSTER_NAME" ./scripts/k8s-prepull-infra.sh
fi

echo "== 载入应用镜像到集群 =="
K8S_CLUSTER="$CLUSTER" K8S_CLUSTER_NAME="$CLUSTER_NAME" ./scripts/k8s-load-images.sh

echo "== kubectl apply =="
kubectl apply -k deploy/k8s/overlays/local

# rollout status 会等 ReplicaSet 创建 Pod；kubectl wait pod 在 Pod 尚未出现时即报 no matching resources
wait_rollout() {
  local deploy=$1 timeout=$2
  echo "== 等待 deployment/${deploy} 就绪 =="
  kubectl -n "$NS" rollout status "deployment/${deploy}" --timeout="${timeout}"
}

echo "== 等待基础设施就绪 =="
if [[ "$HOST_INFRA" == "1" ]]; then
  echo "（HOST_INFRA=1，基础设施在本机 Docker，已跳过集群内 rollout）"
else
  INFRA_TIMEOUT="${INFRA_TIMEOUT:-600s}"
  wait_rollout mysql "${INFRA_TIMEOUT}"
  wait_rollout redis "${INFRA_TIMEOUT}"
  wait_rollout rocketmq-namesrv "${INFRA_TIMEOUT}"
  wait_rollout rocketmq-broker "${INFRA_TIMEOUT}"
fi

echo ""
echo "已部署到 namespace=${NS}"
echo "  Gateway WS:  ws://localhost:10000/gateway/v1/ws  (kind NodePort 30000)"
echo "  User API:    http://localhost:10100"
echo "  Conversation: http://localhost:10400"
echo "  Cron health: http://localhost:10800/health"
echo "  基础设施:   mysql localhost:3306  redis localhost:6379  rocketmq localhost:9876"
if [[ "$HOST_INFRA" == "1" ]]; then
  echo "  （HOST_INFRA=1：基础设施在本机 Docker，Pod 经 host.docker.internal 访问）"
fi
echo ""
echo "查看 Pod: kubectl -n ${NS} get pods"
echo "日志:     kubectl -n ${NS} logs -f deploy/gateway-api"
echo "非 kind 集群请: make k8s-forward"
