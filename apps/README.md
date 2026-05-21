# apps — 业务微服务目录

所有 go-zero 服务与异步任务均在此目录；共享库见根目录 [`pkg/`](../pkg/)。

| 目录 | 端口 | 说明 |
|------|------|------|
| [gateway/api](gateway/api) | 10000 | WebSocket 网关 + 订阅 `im.gateway.push` 下行 |
| [user](user) | API 10100 / RPC 20100 | 用户（开发期 `POST /v1/auth/dev-token` 按 user_id 签发 JWT） |
| [friend](friend) | 10200 / 20200 | 好友 |
| [group](group) | 10300 / 20300 | 群组 |
| [conversation](conversation) | 10400 / 20400 | 会话 |
| [message](message) | 10500 / 20500 | 消息（API 仅查历史；发送走 Gateway WS → RPC） |
| [notification](notification) | 10600 / 20600 | 系统通知 |
| [push](push) | 10700 / 20700 | 在线心跳（可选） |
| **[cron](cron)** | **10800**（健康检查） | **Kafka 异步任务** |

端口规则：业务 API `10XYZ` ↔ RPC `20XYZ`；**cron 无 RPC**。

## 实时推送数据流

```
message.send → cron(realtime-message) → [在线] → im.gateway.push → gateway WS → 客户端
            → cron(inbox-unread) → inbox.updated → cron(push-dispatch) → [在线] gateway.push (badge)
                                                                        → [离线] push.offline
```

WS 鉴权成功时 Gateway 写入 Redis `online:{uid}`；`ping` 续期；断开时删除。cron 向**在线**群成员/单聊对端推送 `message` / `badge`（无需按群 subscribe）。

## cron 服务包含的任务

| 任务 | Kafka Topic | 说明 |
|------|-------------|------|
| message-persist | `im.message.send` | 消息落库、更新会话摘要 |
| inbox-unread | `im.message.send` | 万人群未读扇出 → `im.inbox.updated` |
| realtime-message | `im.message.send` | 在线用户 WebSocket 消息正文 |
| push-dispatch | `im.inbox.updated` | 在线 badge / 离线 `im.push.offline` |
| offline-push | `im.push.offline` | APNs/FCM 合并推送 |
| push-notification | `im.notification.system` | 系统通知 WebSocket / 离线 |

启动：`./bin/cron -f apps/cron/etc/cron.yaml`  
Gateway：`./bin/gateway-api -f apps/gateway/api/etc/gateway-api.yaml`
