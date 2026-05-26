# 微服务清单（build / load / deploy 共用）
# shellcheck shell=bash
IM_SERVICE_BUILD_SPECS=(
  "im/gateway/gateway-api:./apps/gateway/api:gateway-api"
  "im/user/user-api:./apps/user/api:user-api"
  "im/user/user-rpc:./apps/user/rpc:user-rpc"
  "im/friend/friend-api:./apps/friend/api:friend-api"
  "im/friend/friend-rpc:./apps/friend/rpc:friend-rpc"
  "im/group/group-api:./apps/group/api:group-api"
  "im/group/group-rpc:./apps/group/rpc:group-rpc"
  "im/conversation/conversation-api:./apps/conversation/api:conversation-api"
  "im/conversation/conversation-rpc:./apps/conversation/rpc:conversation-rpc"
  "im/message/message-api:./apps/message/api:message-api"
  "im/message/message-rpc:./apps/message/rpc:message-rpc"
  "im/notification/notification-api:./apps/notification/api:notification-api"
  "im/notification/notification-rpc:./apps/notification/rpc:notification-rpc"
  "im/transfer/transfer:./apps/transfer:transfer"
)

im_list_bins() {
  local spec _img _pkg bin
  for spec in "${IM_SERVICE_BUILD_SPECS[@]}"; do
    IFS=':' read -r _img _pkg bin <<<"$spec"
    echo "$bin"
  done
}

im_valid_bin() {
  local want=$1 spec _img _pkg bin
  for spec in "${IM_SERVICE_BUILD_SPECS[@]}"; do
    IFS=':' read -r _img _pkg bin <<<"$spec"
    [[ "$bin" == "$want" ]] && return 0
  done
  return 1
}

im_bins_joined() {
  im_list_bins | tr '\n' ',' | sed 's/,$//'
}
