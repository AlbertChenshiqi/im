#!/usr/bin/env bash
# 将常用 API 端口转发到本机（minikube / 无 NodePort 映射时）
set -euo pipefail
NS="${K8S_NAMESPACE:-im-local}"

echo "转发到本机（Ctrl+C 停止）..."
kubectl -n "$NS" port-forward svc/gateway-api 10000:10000 &
kubectl -n "$NS" port-forward svc/user-api 10100:10100 &
kubectl -n "$NS" port-forward svc/conversation-api 10400:10400 &
kubectl -n "$NS" port-forward svc/message-api 10500:10500 &
kubectl -n "$NS" port-forward svc/cron 10800:10800 &
wait
