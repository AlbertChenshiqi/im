# Kubernetes 部署

> **完整部署文档**（镜像构建/推送、kind 与云集群、Dashboard 可视化、Ingress 路由）：  
> **[DEPLOY.md](./DEPLOY.md)**

本地与线上统一使用 K8s；清单**按模块分目录**，便于查看与扩展。

## 目录结构

```
deploy/k8s/overlays/local/
├── namespace.yaml
├── kustomization.yaml
├── migrations/              # 供 postgres init
├── config/                  # make k8s-etc 生成的服务配置
│   ├── gateway/
│   ├── user/
│   └── ...
├── infra/                   # Postgres / Redis / RocketMQ
│   ├── kustomization.yaml
│   └── infra.yaml
└── apps/                    # 按业务模块拆分（make k8s-etc 生成）
    ├── gateway/
    │   ├── kustomization.yaml
    │   ├── gateway-api-configmap.yaml
    │   ├── gateway-api-deployment.yaml   # 默认 2 副本
    │   └── gateway-api-service.yaml      # ClientIP 会话保持
    ├── user/
    │   ├── user-api-*.yaml
    │   └── user-rpc-*.yaml
    ├── friend/
    ├── group/
    ├── conversation/
    ├── message/
    ├── notification/
    └── transfer/
```

## 本地开发（kind）

```bash
make image-base-build
make up
make seed
kubectl -n im-local get pods -l app.kubernetes.io/component=gateway
```

### Gateway 多节点

| 机制 | 说明 |
|------|------|
| **副本数** | 默认 `GATEWAY_REPLICAS=2`（`make k8s-etc GATEWAY_REPLICAS=3` 可改） |
| **RocketMQ** | Tag `gateway_push` + **广播消费**；每 Pod 全量收到消息，仅向本机 WS 连接投递（可选 `GATEWAY_INSTANCE_ID` 区分 group） |
| **Service** | `sessionAffinity: ClientIP`，长连接尽量粘在同一 Pod |
| **Redis 在线** | `online_gateways:{uid}` 记录持有连接的实例；仅当所有实例均无连接才清除 `online:{uid}` |

生产环境建议在 Ingress 层增加 **WebSocket 粘性**（如 nginx `affinity: cookie`），与 Service 会话保持配合。

## Make 目标

| 目标 | 说明 |
|------|------|
| `make k8s-etc` | 生成 `config/` + `apps/<module>/` |
| `make up` | 构建镜像并部署 |
| `make seed` | 测试数据 |
| `GATEWAY_REPLICAS=3 make k8s-etc` | 调整 gateway 副本后需重新 apply |

## 镜像

```bash
make image-base-build
make build-images    # im/user/user-api:dev 等
```

载入 kind 时使用 **`docker save` + `kind load image-archive` 单次导入**（避免连续 16 次 `kind load` 导致 containerd 连接失败）。若仍失败：

```bash
kind delete cluster --name im-local
make up
# 或逐张载入：KIND_LOAD_MODE=single make up
```

## 访问（kind NodePort 映射）

| 服务 | 地址 |
|------|------|
| Gateway WS | `ws://localhost:10000/gateway/v1/ws` |
| User API | `http://localhost:10100` |
| Postgres | `postgres://im:im@localhost:5432/im?sslmode=disable` |
| Redis | `localhost:6379` |
| RocketMQ NameServer | `localhost:9876` |

`make k8s-forward`：非 kind 集群时 API 端口转发。基础设施非 kind 时用 `make k8s-forward-infra`。
