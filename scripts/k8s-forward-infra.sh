#!/usr/bin/env bash
# 将集群内 Postgres / Redis / Kafka 转发到本机（供 make run-all 本机进程调试）
set -euo pipefail
NS="${K8S_NAMESPACE:-im-local}"

echo "转发 postgres:5432 redis:6379 kafka:9092（需已 make k8s-up，Ctrl+C 停止）"
kubectl -n "$NS" port-forward svc/postgres 5432:5432 &
kubectl -n "$NS" port-forward svc/redis 6379:6379 &
kubectl -n "$NS" port-forward svc/kafka 9092:9092 &
wait
