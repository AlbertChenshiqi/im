#!/usr/bin/env bash
# 删除 IM 的 K8s 部署；可选删除 kind 集群与 kubernetes-dashboard
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

NS="im-local"
DELETE_CLUSTER="${DELETE_CLUSTER:-0}"
DELETE_DASHBOARD="${DELETE_DASHBOARD:-1}"
CLUSTER_NAME="${K8S_CLUSTER_NAME:-im-local}"

if ! kubectl cluster-info >/dev/null 2>&1; then
  echo "kubectl 无法连接集群，跳过 kubectl delete"
else
  echo "== 删除 ${NS} 应用清单 =="
  kubectl delete -k deploy/k8s/overlays/local --ignore-not-found --wait=false 2>/dev/null || true
  kubectl delete namespace "$NS" --ignore-not-found --timeout=120s 2>/dev/null || true

  if [[ "$DELETE_DASHBOARD" == "1" ]]; then
    echo "== 删除 kubernetes-dashboard =="
    kubectl delete namespace kubernetes-dashboard --ignore-not-found --timeout=120s 2>/dev/null || true
  fi
fi

if [[ "$DELETE_CLUSTER" == "1" ]] && command -v kind >/dev/null; then
  if kind get clusters 2>/dev/null | grep -qx "${CLUSTER_NAME}"; then
    echo "== 删除 kind 集群 ${CLUSTER_NAME} =="
    kind delete cluster --name "${CLUSTER_NAME}"
  fi
fi

echo ""
echo "K8s 部署已清理（数据卷 PVC 若在集群内可能仍保留于 kind 节点，删集群后一并消失）"
echo "本机调试: make up && make migrate && make run-all"
