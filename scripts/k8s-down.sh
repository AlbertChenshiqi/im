#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

NS="im-local"
DELETE_CLUSTER="${DELETE_CLUSTER:-0}"
CLUSTER_NAME="${K8S_CLUSTER_NAME:-im-local}"

kubectl delete -k deploy/k8s/overlays/local --ignore-not-found

if [[ "$DELETE_CLUSTER" == "1" ]] && command -v kind >/dev/null; then
  kind delete cluster --name "${CLUSTER_NAME}" || true
fi

echo "已删除 namespace ${NS} 资源（PVC 可能保留）"
echo "删除 kind 集群: DELETE_CLUSTER=1 ./scripts/k8s-down.sh"
