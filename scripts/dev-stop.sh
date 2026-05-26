#!/usr/bin/env bash
# 停止 run-local.sh 启动的本机 Go 进程（按监听端口）
set -euo pipefail

PORTS=(20100 20200 20300 20400 20500 20600 10000 10100 10200 10300 10400 10500 10600 10800)

stopped=0
for port in "${PORTS[@]}"; do
  pids=$(lsof -ti ":$port" 2>/dev/null || true)
  [[ -z "$pids" ]] && continue
  echo "停止端口 ${port} 上的进程: ${pids}"
  kill $pids 2>/dev/null || true
  stopped=$((stopped + 1))
done

if [[ "$stopped" -eq 0 ]]; then
  echo "未发现 run-local 占用的端口"
else
  echo "本机微服务已停止"
fi
