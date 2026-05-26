# apps — 业务微服务目录

所有 go-zero 服务与异步任务均在此目录；共享库见根目录 [`pkg/`](../pkg/)。

| 目录 | 端口 | 说明 |
|------|------|------|
| [gateway/api](gateway/api) | 10000 | WebSocket 网关 + 订阅 `im_sync` / `gateway_push` |
| [user](user) | API 10100 / RPC 20100 | 用户（含 HTTP 在线心跳 `POST /user/v1/online`） |
| [friend](friend) | 10200 / 20200 | 好友 |
| [group](group) | 10300 / 20300 | 群组 |
| [conversation](conversation) | 10400 / 20400 | 会话 |
| [message](message) | 10500 / 20500 | 消息（API 查历史；发送走 Gateway WS → RPC） |
| [notification](notification) | 10600 / 20600 | 系统通知 |
| **[transfer](transfer)** | **10800** | **RocketMQ 异步任务** |

## RocketMQ Topic

| Topic | 用途 |
|-------|------|
| `im_chat` | 单聊 `c2c`、群聊 `group`、撤回 `recall`、自定义 `custom`、落库 `persisted` |
| `im_push` | 离线 `offline_message`、系统公告 `system_announce` |
| `im_sync` | 已读 `read`、实时下行 `gateway_push`、上下线 `online`、好友 `friend`（后两者预留） |

## 实时推送数据流

```
im_chat (c2c|group) → transfer(realtime) → [在线] → im_sync/gateway_push → gateway WS
                     → transfer(inbox)   → im_sync/read → push-dispatch → 在线 badge / im_push/offline_message
```

## transfer 任务

| 任务 | Topic | Tag 订阅 | 说明 |
|------|-------|----------|------|
| message-persist | `im_chat` | `c2c \|\| group \|\| custom \|\| recall` | 落库 |
| inbox-unread | `im_chat` | 同上 | 未读 → `im_sync`/`read` |
| realtime-message | `im_chat` | 同上 | 在线正文 → `im_sync`/`gateway_push` |
| push-dispatch | `im_sync` | `read` | badge / 离线 |
| offline-push | `im_push` | `offline_message` | APNs/FCM |
| push-notification | `im_push` | `system_announce` | 系统公告 |

启动：`./bin/transfer -f apps/transfer/etc/transfer.yaml`  
Gateway：`./bin/gateway-api -f apps/gateway/api/etc/gateway-api.yaml`
