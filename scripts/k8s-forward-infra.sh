#!/usr/bin/env bash
# 将集群内 Postgres / Redis / RocketMQ 转发到本机（非 kind 或 NodePort 未映射时使用）
set -euo pipefail
NS="${K8S_NAMESPACE:-im-local}"

echo "转发 mysql:3306 redis:6379 rocketmq-namesrv:9876（需已 make up，Ctrl+C 停止）"
echo "提示: kind 本地集群已 NodePort 映射到本机同端口，一般无需本脚本，可直接 make run-all"
kubectl -n "$NS" port-forward svc/postgres 5432:5432 &
kubectl -n "$NS" port-forward svc/redis 6379:6379 &
kubectl -n "$NS" port-forward svc/rocketmq-namesrv 9876:9876 &
wait
