#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"

gen_api() {
  local dir=$1 api=$2
  if [[ -f "$ROOT/$dir/$api" ]]; then
    (cd "$ROOT/$dir" && goctl api go -api "$api" -dir . --style go_zero)
  fi
}

gen_rpc() {
  local dir=$1 proto=$2
  if [[ -f "$ROOT/$dir/$proto" ]]; then
    (cd "$ROOT/$dir" && goctl rpc protoc "$proto" --go_out=. --go-grpc_out=. --zrpc_out=. --style go_zero)
  fi
}

gen_api apps/user/api user.api
gen_rpc apps/user/rpc user.proto

for svc in group friend conversation message notification push; do
  gen_api "apps/$svc/api" "${svc}.api"
  gen_rpc "apps/$svc/rpc" "${svc}.proto"
done

gen_api apps/gateway/api gateway.api

echo "go-zero codegen done"
