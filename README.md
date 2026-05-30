<div align="center">

# Easy IM

### Production-grade instant messaging backend for 10K-member groups  
### 面向万人群的生产级即时通讯后端

**go-zero · WebSocket · RocketMQ · Redis · MySQL · Kubernetes**

[English](#-english) · [中文](#-中文) · [Architecture 架构详解](docs/architecture-zh.md) · [Frontend Integration 前端对接](docs/frontend-integration.md)

<br />

[![Go](https://img.shields.io/badge/Go-1.22%2B-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev/)
[![go-zero](https://img.shields.io/badge/go--zero-microservices-006AFF?style=flat-square)](https://go-zero.dev/)
[![RocketMQ](https://img.shields.io/badge/RocketMQ-async%20pipeline-D77310?style=flat-square)](https://rocketmq.apache.org/)
[![WebSocket](https://img.shields.io/badge/WebSocket-real--time-010101?style=flat-square)](https://developer.mozilla.org/en-US/docs/Web/API/WebSockets_API)
[![Kubernetes](https://img.shields.io/badge/Kubernetes-ready-326CE5?style=flat-square&logo=kubernetes&logoColor=white)](deploy/k8s/DEPLOY.md)

<br />

**If this project helps you, please give it a ⭐ — it means a lot!**  
**如果对你有帮助，欢迎 Star ⭐ 支持一下！**

</div>

---

## 🇬🇧 English

### What is Easy IM?

**Easy IM** is an open-source, production-oriented instant messaging backend built on [go-zero](https://go-zero.dev/). It targets **large group chats (10K+ members)** with a clean separation: REST for queries and management, WebSocket for sending messages and real-time push, and **RocketMQ** for async write pipelines.

### Highlights

- **Microservices** — user, friend, group, conversation, message, notification, gateway, transfer
- **WebSocket-first writes** — send messages only via Gateway `:10000`; REST for history & inbox
- **RocketMQ pipeline** — topics `im_chat` / `im_push` / `im_sync` with tag-based routing
- **10K-member groups** — Redis batch unread, inbox merge, online-only realtime push
- **K8s-ready** — one-command local cluster via `kind` + Kustomize overlays
- **Frontend-friendly** — JWT auth, `conv_id` rules, `input`-based message model ([docs](docs/frontend-integration.md))

### Architecture (overview)

```
Client ──REST──► API services (10100–10600)
       ──WS────► Gateway :10000 ──zrpc──► message/group RPC
                        ▲
                        └── RocketMQ im_sync/gateway_push (broadcast)

message/rpc ──► im_chat ──► transfer (persist · unread · realtime · offline push)
```

### Quick Start

**Requirements:** Go 1.22+, Docker, `kubectl`, `kind`

```bash
git clone https://github.com/AlbertChenshiqi/im.git
cd im

make image-base-build   # first time only
make up                 # build images → kind cluster → deploy all services
make seed               # seed test data

curl http://localhost:10000/gateway/v1/health
curl http://localhost:10800/health
```

**Dev token (Auth.DevMode: true):**

```bash
curl -s -X POST http://localhost:10100/user/v1/auth/dev-token \
  -H 'Content-Type: application/json' \
  -d '{"user_id":1}'
```

**WebSocket:**

```
ws://localhost:10000/gateway/v1/ws?token=<JWT>
```

See [docs/architecture-zh.md](docs/architecture-zh.md) for ports, RocketMQ tags, Redis keys, and deployment details.

### Services

| Service | API | RPC | Role |
|---------|-----|-----|------|
| gateway | **10000** | — | WebSocket + MQ downstream |
| user | 10100 | 20100 | Auth, profile |
| friend | 10200 | 20200 | Friendships |
| group | 10300 | 20300 | Groups & members |
| conversation | 10400 | 20400 | Inbox, read receipts |
| message | 10500 | 20500 | History (read-only REST) |
| notification | 10600 | 20600 | System notifications |
| transfer | 10800 | — | Async MQ workers |

### Documentation

| Doc | Description |
|-----|-------------|
| [docs/architecture-zh.md](docs/architecture-zh.md) | Full architecture, flows, ops (中文) |
| [docs/frontend-integration.md](docs/frontend-integration.md) | Client protocol & REST/WS API |
| [deploy/k8s/DEPLOY.md](deploy/k8s/DEPLOY.md) | Kubernetes deployment guide |
| [apps/README.md](apps/README.md) | Service directory overview |

### Contributing

Issues and PRs are welcome. Please open an issue before large changes.

---

## 🇨🇳 中文

### 什么是 Easy IM？

**Easy IM** 是基于 [go-zero](https://go-zero.dev/) 的**生产级开源 IM 后端**，面向**万人群**场景：REST 负责查询与管理，**WebSocket 统一负责发消息与实时下行**，写路径通过 **RocketMQ** 异步解耦。

### 核心特性

- **微服务拆分** — user / friend / group / conversation / message / notification / gateway / transfer
- **WebSocket 写、REST 读** — 发消息仅走 Gateway `:10000`；历史、会话列表等走各域 API
- **RocketMQ 异步管道** — `im_chat` / `im_push` / `im_sync` 三 Topic + Tag 订阅路由
- **万人群防风暴** — Redis 批处理未读、`read` 事件合并、仅在线成员实时推送
- **K8s 一键部署** — `kind` 本地集群 + Kustomize，与线上一致
- **前端友好** — JWT 鉴权、`conv_id` 规范、`input` 消息体（见 [前端对接文档](docs/frontend-integration.md)）

### 架构概览

```
客户端 ──REST──► 各域 API（10100–10600）
       ──WS────► Gateway :10000 ──zrpc──► message/group RPC
                        ▲
                        └── RocketMQ im_sync/gateway_push（广播消费）

message/rpc ──► im_chat ──► transfer（落库 · 未读 · 实时 · 离线推送）
```

### 快速开始

**环境：** Go 1.22+、Docker、`kubectl`、`kind`

```bash
git clone https://github.com/AlbertChenshiqi/im.git
cd im

make image-base-build   # 首次构建基础镜像
make up                 # 构建镜像 → kind 集群 → 部署全部服务
make seed               # 灌测试数据

curl http://localhost:10000/gateway/v1/health
curl http://localhost:10800/health
```

**开发鉴权（DevMode）：**

```bash
curl -s -X POST http://localhost:10100/user/v1/auth/dev-token \
  -H 'Content-Type: application/json' \
  -d '{"user_id":1}'
```

**WebSocket 连接：**

```
ws://localhost:10000/gateway/v1/ws?token=<JWT>
```

端口规划、RocketMQ Tag、Redis Key、生产部署等详见 [docs/architecture-zh.md](docs/architecture-zh.md)。

### 服务一览

| 模块 | API | RPC | 职责 |
|------|-----|-----|------|
| gateway | **10000** | — | WebSocket + MQ 下行 |
| user | 10100 | 20100 | 注册登录、资料 |
| friend | 10200 | 20200 | 好友关系 |
| group | 10300 | 20300 | 群组与成员 |
| conversation | 10400 | 20400 | 会话列表、已读 |
| message | 10500 | 20500 | 历史消息（只读 REST） |
| notification | 10600 | 20600 | 系统通知 |
| transfer | 10800 | — | RocketMQ 异步任务 |

### 文档

| 文档 | 说明 |
|------|------|
| [docs/architecture-zh.md](docs/architecture-zh.md) | 完整架构、流程图、运维（中文） |
| [docs/frontend-integration.md](docs/frontend-integration.md) | 客户端协议与 REST/WS 对接 |
| [deploy/k8s/DEPLOY.md](deploy/k8s/DEPLOY.md) | Kubernetes 部署指南 |
| [apps/README.md](apps/README.md) | 微服务目录说明 |

### 参与贡献

欢迎提交 Issue 与 PR。较大改动建议先开 Issue 讨论。

---

## ⭐ Star History / Star 趋势

<a href="https://star-history.com/#AlbertChenshiqi/im&Date">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/svg?repos=AlbertChenshiqi/im&type=Date&theme=dark" />
    <img alt="Star History Chart — Easy IM star growth over time" src="https://api.star-history.com/svg?repos=AlbertChenshiqi/im&type=Date" />
  </picture>
</a>

<p align="center">
  <sub>
    <a href="https://star-history.com/#AlbertChenshiqi/im&Date">View interactive chart 查看交互式图表</a>
  </sub>
</p>

---

<div align="center">

**Easy IM** — Real-time messaging, built for scale.  
**Easy IM** — 为规模而生的实时通讯后端。

</div>
