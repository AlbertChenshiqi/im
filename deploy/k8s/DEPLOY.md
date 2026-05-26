# IM 平台 Kubernetes 部署指南

本文档说明如何从**构建/拉取镜像**、**创建集群**、**部署应用**、**可视化运维**，到**配置 Ingress 路由**的完整流程。清单与脚本以仓库内 `deploy/k8s/overlays/local` 为准。

---

## 目录

1. [架构与端口](#1-架构与端口)
2. [环境准备](#2-环境准备)
3. [镜像：构建、推送与拉取](#3-镜像构建推送与拉取)
4. [方式 A：本地 kind 一键部署](#4-方式-a本地-kind-一键部署)
5. [方式 B：已有 K8s 集群部署](#5-方式-b已有-k8s-集群部署)
6. [Kubernetes 可视化面板](#6-kubernetes-可视化面板)
7. [配置说明（Namespace / ConfigMap / 服务发现）](#7-配置说明namespace--configmap--服务发现)
8. [Ingress 与对外路由](#8-ingress-与对外路由)
9. [日常运维](#9-日常运维)
10. [故障排查](#10-故障排查)

---

## 1. 架构与端口

| 层级 | 组件 | 说明 |
|------|------|------|
| 接入 | `gateway-api` | WebSocket `:10000`，多副本 + 会话保持 |
| API | `*-api` | 前端 REST `10100–10600` |
| RPC | `*-rpc` | 集群内 `20100–20600`，**不对外暴露** |
| 异步 | `transfer` | 健康检查 `:10800`，消费 RocketMQ |
| 基础设施 | Postgres / Redis / RocketMQ | 集群内 DNS；**kind 下映射到本机 5432/6379/9876** |

**Namespace**：`im-local`（见 `deploy/k8s/overlays/local/namespace.yaml`）

**kind 本地端口映射**（`deploy/kind/im-local.yaml` 将 NodePort 映射到本机）：

| 服务 | 集群内 Service 端口 | NodePort | 本机访问 |
|------|---------------------|----------|----------|
| gateway-api | 10000 | 30000 | `ws://localhost:10000/gateway/v1/ws` |
| user-api | 10100 | 30100 | `http://localhost:10100` |
| friend-api | 10200 | 30200 | `http://localhost:10200` |
| group-api | 10300 | 30300 | `http://localhost:10300` |
| conversation-api | 10400 | 30400 | `http://localhost:10400` |
| message-api | 10500 | 30500 | `http://localhost:10500` |
| notification-api | 10600 | 30600 | `http://localhost:10600` |
| transfer | 10800 | 30800 | `http://localhost:10800/health` |
| postgres | 5432 | 30432 | `localhost:5432`（`im/im`，库 `im`） |
| redis | 6379 | 30637 | `localhost:6379` |
| rocketmq-namesrv | 9876 | 30876 | `localhost:9876` |
| rocketmq-broker | 10911 | 30911 | `localhost:10911`（管理/排障，业务连 NameServer） |

RPC（`20100–20600`）在生成清单中为 **ClusterIP、无 NodePort**，仅 Pod 间通过服务名访问，例如 `message-rpc:20500`。

**本机调试**：kind 集群下可直接连上表「本机访问」地址，与 `apps/*/etc/*.yaml` 中 `localhost` 一致，**无需** `make k8s-forward-infra`。非 kind 集群仍用 `make k8s-forward-infra` 或 `kubectl port-forward`。

---

## 2. 环境准备

### 2.1 必需工具

| 工具 | 用途 | 安装示例（macOS） |
|------|------|-------------------|
| **Docker** | 构建镜像、kind 节点 | Docker Desktop |
| **kubectl** | 操作集群 | `brew install kubectl` |
| **kind** | 本地 K8s（推荐） | `brew install kind` |
| **helm** | 安装 Ingress / Dashboard（可选） | `brew install helm` |

验证：

```bash
docker version
kubectl version --client
kind version
```

### 2.2 仓库与 Go

```bash
git clone <your-repo> im && cd im
go version   # 建议 Go 1.22+，与 go.mod 一致
```

---

## 3. 镜像：构建、推送与拉取

### 3.1 镜像命名约定

构建脚本 `scripts/build-images.sh` 会生成 **16 个**业务镜像，标签默认为 `dev`：

```text
im/gateway/gateway-api:dev
im/user/user-api:dev
im/user/user-rpc:dev
…（各域 api + rpc）
im/transfer/transfer:dev
```

基础镜像（先构建一次）：

```bash
make image-base-build
# 生成 im-go-base:deps、im-go-base:runtime
```

### 3.2 本地构建全部服务镜像

```bash
make image-base-build    # 首次必须
make build-images      # 或 TAG_SUFFIX=v1.0.0 make build-images
```

查看本地镜像：

```bash
docker images 'im/*'
```

### 3.3 推送到镜像仓库（生产/测试集群）

以阿里云 ACR 为例（请替换为自己的 registry）：

**推荐做法**：对 `scripts/build-images.sh` 列出的每个镜像执行 tag / push（共 16 个，见脚本内 `SERVICES` 数组）：

```bash
REGISTRY=registry.example.com/im
TAG=v1.0.0

docker tag im/gateway/gateway-api:dev  ${REGISTRY}/gateway-api:${TAG}
docker push ${REGISTRY}/gateway-api:${TAG}
# …对其余 15 个镜像重复 tag / push
```

登录私有仓库：

```bash
docker login registry.example.com
# 或在 K8s 中创建 pull secret（见下文 5.2）
```

### 3.4 集群如何「拉」镜像

| 场景 | 方式 |
|------|------|
| **kind 本地** | 节点无法直接访问本机 Docker 守护进程，需 **`kind load`** 导入（`make up` 已自动执行） |
| **云厂商 K8s** | Deployment 中 `image` 指向仓库地址，`imagePullPolicy: Always`，配置 `imagePullSecrets` |
| **minikube** | `minikube image load im/gateway/gateway-api:dev` 或 push 到仓库 |

kind 载入逻辑见 `scripts/k8s-load-images.sh`：

- 默认 **`docker save` 打包 + `kind load image-archive`**（稳定）
- 失败时可：`KIND_LOAD_MODE=single make up` 逐张导入
- 仍失败：`kind delete cluster --name im-local && make up`

### 3.5 修改 Deployment 使用的镜像地址

生成清单默认：

```yaml
image: im/gateway/gateway-api:dev
imagePullPolicy: IfNotPresent
```

**生产**建议在 `deploy/k8s/overlays/local` 上增加 Kustomize 补丁，例如 `kustomization.yaml`：

```yaml
images:
  - name: im/gateway/gateway-api
    newName: registry.example.com/im/gateway-api
    newTag: v1.0.0
  # …为每个 im/* 镜像添加一项
```

然后 `kubectl apply -k deploy/k8s/overlays/local`。

---

## 4. 方式 A：本地 kind 一键部署

### 4.1 全流程（推荐）

```bash
make image-base-build   # 首次
make up                 # = k8s-etc + build-images + kind + load + kubectl apply
make seed               # 灌测试数据（可选）
```

等价命令：

```bash
make k8s-etc            # 生成 config/ 与 apps/*/ 清单
make build-images
make k8s-up
```

### 4.2 创建 kind 集群（手动）

`make up` 会在集群不存在时自动执行：

```bash
kind create cluster --name im-local --config deploy/kind/im-local.yaml
kubectl cluster-info --context kind-im-local
```

### 4.3 验证部署

```bash
kubectl -n im-local get pods
kubectl -n im-local get svc
kubectl -n im-local rollout status deployment/gateway-api --timeout=120s
```

访问：

- WebSocket：`ws://localhost:10000/gateway/v1/ws?token=<JWT>`
- 用户 API：`http://localhost:10100`
- Transfer 健康：`http://localhost:10800/health`

### 4.4 跳过重新构建镜像

```bash
SKIP_BUILD=1 make up
```

### 4.5 增量部署（改代码后推荐）

一条命令完成：**构建镜像 → 载入 kind → 调副本（可选）→ 滚动发布**：

```bash
make deploy SVC=message-rpc
make deploy SVC=gateway-api:2,message-rpc,transfer
make deploy SVC=gateway-api,message-rpc REPLICAS=1   # 未写 :N 的服务用同一副本数
```

等价脚本：`./scripts/k8s-deploy.sh gateway-api:2 message-rpc`

仅构建或仅载入：`make build-images SVC=...`、`make k8s-load-images SVC=...`（内部调用同一套 `SVC` 参数）。

### 4.6 调整 Gateway 副本数（改清单）

```bash
GATEWAY_REPLICAS=3 make k8s-etc
kubectl apply -k deploy/k8s/overlays/local
```

---

## 5. 方式 B：已有 K8s 集群部署

适用于阿里云 ACK、腾讯云 TKE、自建 kubeadm 集群等。

### 5.1 部署清单

```bash
make k8s-etc
# 按 3.5 节 patch 镜像地址为仓库 URL
kubectl apply -k deploy/k8s/overlays/local
```

### 5.2 私有仓库拉取凭证

```bash
kubectl -n im-local create secret docker-registry im-regcred \
  --docker-server=registry.example.com \
  --docker-username=<user> \
  --docker-password=<pass> \
  --docker-email=<email>

# 在 Deployment patch 或 default ServiceAccount 中引用：
# imagePullSecrets:
#   - name: im-regcred
```

### 5.3 无 NodePort 时访问服务

云集群常不映射 NodePort 到本机，使用端口转发：

```bash
make k8s-forward
# 或仅转发基础设施（本机跑 Go 进程调试）：
make k8s-forward-infra   # postgres:5432 redis:6379 rocketmq-namesrv:9876
```

### 5.4 等待基础设施

`k8s-up.sh` 会等待 Postgres、Redis、RocketMQ NameServer/Broker Ready。手动检查：

```bash
kubectl -n im-local get pods -l app=postgres
kubectl -n im-local logs -l app=rocketmq-broker --tail=50
```

---

## 6. Kubernetes 可视化面板

### 6.1 Kubernetes Dashboard（官方 Web UI）

**安装**（集群级，仅需一次）：

```bash
kubectl apply -f https://raw.githubusercontent.com/kubernetes/dashboard/v2.7.0/aio/deploy/recommended.yaml
```

创建管理员访问账号（**仅开发/测试环境**；生产请用 RBAC 最小权限）：

```bash
cat <<'EOF' | kubectl apply -f -
apiVersion: v1
kind: ServiceAccount
metadata:
  name: admin-user
  namespace: kubernetes-dashboard
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: admin-user
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
  - kind: ServiceAccount
    name: admin-user
    namespace: kubernetes-dashboard
EOF
```

**启动本地代理并打开浏览器**：

上文 `recommended.yaml`（v2.7）创建的 Service 名为 **`kubernetes-dashboard`**，不是 Helm v7 的 `kubernetes-dashboard-kong-proxy`：

```bash
# v2.7 aio（本文档安装方式）
kubectl -n kubernetes-dashboard port-forward svc/kubernetes-dashboard 8443:443
```

若用 **Helm 安装 Dashboard v7+**（带 Kong），才用：

```bash
kubectl -n kubernetes-dashboard port-forward svc/kubernetes-dashboard-kong-proxy 8443:443
```

浏览器访问：`https://localhost:8443`（接受自签名证书）。Pod 未 Ready 时需先 `kubectl -n kubernetes-dashboard get pods -w`。

**获取登录 Token**：

```bash
kubectl -n kubernetes-dashboard create token admin-user
```

在 Dashboard 登录页选择 **Token**，粘贴即可。

**在 Dashboard 中查看 IM**：

- 左上角 Namespace 选 **`im-local`**
- **Workloads → Deployments**：`gateway-api`、`user-api`、`transfer` 等
- **Service**：NodePort / ClusterIP
- **Pods → Logs**：查看业务日志
- **Config Maps**：`gateway-api-config` 等

### 6.2 其他常用可视化工具

| 工具 | 说明 |
|------|------|
| **[Lens](https://k8slens.dev/)** | 桌面客户端，免安装 Dashboard，直接选 kubeconfig |
| **k9s** | 终端 TUI：`brew install k9s` → `k9s -n im-local` |
| **Kuboard** | 国产 Web 面板，适合团队运维 |

kubectl 自带观察命令：

```bash
kubectl -n im-local get pods -o wide
kubectl -n im-local top pods    # 需 metrics-server
watch kubectl -n im-local get pods
```

---

## 7. 配置说明（Namespace / ConfigMap / 服务发现）

### 7.1 目录结构

```text
deploy/k8s/overlays/local/
├── namespace.yaml          # im-local
├── kustomization.yaml      # 聚合 infra + 各 app
├── migrations/             # Postgres 初始化 SQL
├── config/                 # 服务 yaml 源（集群 DNS）
│   ├── gateway/gateway-api.yaml
│   ├── message/message-rpc.yaml
│   └── …
├── infra/                  # Postgres、Redis、RocketMQ
└── apps/                   # 各模块 Deployment/Service/ConfigMap
    ├── gateway/
    ├── user/
    └── …
```

修改配置后：

```bash
./scripts/write-k8s-etc.sh              # 根据模板写 config/
GATEWAY_REPLICAS=2 ./scripts/gen-k8s-manifests.sh
kubectl apply -k deploy/k8s/overlays/local
kubectl -n im-local rollout restart deploy/gateway-api
```

### 7.2 集群内 DNS

Pod 内访问 RPC 使用 **Service 名**（同 namespace 可省略后缀）：

| 配置项示例 | 值 |
|------------|-----|
| `MessageRpc.Endpoints` | `message-rpc:20500` |
| `Postgres.DSN` | `postgres://im:im@postgres:5432/im?sslmode=disable` |
| `Redis.Addr` | `redis:6379` |
| `RocketMQ.NameServer` | `rocketmq-namesrv:9876` |

源文件：`deploy/k8s/overlays/local/config/*`，由 `scripts/write-k8s-etc.sh` 生成。

### 7.3 Gateway 多副本要点

- `GATEWAY_INSTANCE_ID` = Pod 名（见 Deployment env）
- Service **`sessionAffinity: ClientIP`**（长连接粘滞）
- RocketMQ `gateway_push` 使用**广播消费**，每 Pod 收全量，仅推本机 WS

生产环境请在 **Ingress** 上增加 cookie 粘性（见下节）。

### 7.4 密钥与生产配置

- JWT：`Auth.AccessSecret` 在 ConfigMap 中，**生产请改为 Secret + 外部 KMS**
- Postgres 密码：当前为开发默认值 `im/im`，生产请改 `infra` 与 DSN

---

## 8. Ingress 与对外路由

本地 **kind** 已通过 **NodePort + hostPort 映射** 暴露，可不装 Ingress。  
**测试/生产** 建议使用 Ingress Controller 统一 HTTPS 与域名路由。

### 8.1 安装 NGINX Ingress Controller

```bash
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update

helm upgrade --install ingress-nginx ingress-nginx/ingress-nginx \
  --namespace ingress-nginx \
  --create-namespace \
  --set controller.watchIngressWithoutClass=true
```

验证：

```bash
kubectl -n ingress-nginx get pods
kubectl -n ingress-nginx get svc
# 云 LB 场景：EXTERNAL-IP 即为入口
# kind：常为 NodePort，可用 port-forward：
kubectl -n ingress-nginx port-forward svc/ingress-nginx-controller 8080:80
```

### 8.2 路由设计建议

| 域名 | 后端 Service | 说明 |
|------|--------------|------|
| `api.im.example.com` | 各 `*-api` | REST，按路径前缀拆分 |
| `api.im.example.com/gateway/v1` | `gateway-api:10000` | WebSocket `/gateway/v1/ws`（可与 REST 同域） |

**不要** 将 `*-rpc` 暴露到 Ingress。

仓库示例清单：`deploy/k8s/examples/ingress-im.yaml`

```bash
# 1. 将文件中 im.example.com 改为你的域名
# 2. 确认 namespace 为 im-local
kubectl apply -f deploy/k8s/examples/ingress-im.yaml
```

### 8.3 WebSocket 关键注解

示例中已包含（可按需调整）：

```yaml
nginx.ingress.kubernetes.io/proxy-read-timeout: "3600"
nginx.ingress.kubernetes.io/proxy-send-timeout: "3600"
nginx.ingress.kubernetes.io/affinity: "cookie"
nginx.ingress.kubernetes.io/session-cookie-name: "im-gw-route"
```

客户端连接：

```text
wss://api.im.example.com/gateway/v1/ws?token=<JWT>
```

### 8.4 TLS（HTTPS / WSS）

**cert-manager + Let's Encrypt**（简要）：

```bash
helm repo add jetstack https://charts.jetstack.io
helm install cert-manager jetstack/cert-manager \
  --namespace cert-manager --create-namespace \
  --set crds.enabled=true
```

在 Ingress 中取消注释 `tls` 段，并配置 `cert-manager.io/cluster-issuer` 注解（详见 cert-manager 文档）。

### 8.5 路径与 API 路由说明

各 API 路由形如 `/{service}/v1/...`。示例 Ingress 将：

- `/user/v1` → `user-api`
- `/group/v1` → `group-api`
- `/conversation/v1` → `conversation-api`
- `/message/v1` → `message-api`
- `/gateway/v1` → `gateway-api`（含 WebSocket `/gateway/v1/ws`）
- …

若你新增了路由前缀，需同步修改 Ingress 的 `paths`。  
**发消息**走 WebSocket，不经过 message-api 的 HTTP POST。

### 8.6 仅 Gateway 单域名（简化）

若 REST 仍走各端口 NodePort，仅 Gateway 走 Ingress：

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: im-gateway
  namespace: im-local
  annotations:
    nginx.ingress.kubernetes.io/proxy-read-timeout: "3600"
    nginx.ingress.kubernetes.io/affinity: "cookie"
spec:
  ingressClassName: nginx
  rules:
```

（WebSocket 已包含在上文 `/gateway/v1` 前缀路由中，无需单独域名。）

### 8.7 DNS

将 `api.im.example.com` 解析到 Ingress Controller 的 **EXTERNAL-IP**（或 LB 地址）。

---

## 9. 日常运维

| 操作 | 命令 |
|------|------|
| 查看 Pod | `kubectl -n im-local get pods -o wide` |
| Gateway 日志 | `make k8s-logs` 或 `kubectl -n im-local logs -f deploy/gateway-api --tail=200` |
| 重启 Gateway | `kubectl -n im-local rollout restart deploy/gateway-api` |
| 扩缩容 | `kubectl -n im-local scale deploy/gateway-api --replicas=3` |
| 重新灌数据 | `make seed` |
| 卸载 | `make down`（`scripts/k8s-down.sh`） |
| 重新生成清单 | `make k8s-etc && kubectl apply -k deploy/k8s/overlays/local` |

**滚动升级镜像**：

```bash
TAG=v1.0.1 make build-images
# push 到仓库后更新 kustomize images 或 set image：
kubectl -n im-local set image deploy/gateway-api gateway-api=registry.example.com/im/gateway-api:v1.0.1
kubectl -n im-local rollout status deploy/gateway-api
```

---

## 10. 故障排查

### Pod ImagePullBackOff

- **业务镜像 `im/*:dev`**：kind 是否执行过 `make up` / `make k8s-load-images`？`docker images 'im/*:dev'` 是否存在？
- **基础设施 `postgres` / `redis` / `rocketmq`**：kind 节点拉 Docker Hub 常失败（代理未开、网络慢）。`make up` 默认会先 `k8s-prepull-infra`（本机 `docker pull` 再 `kind load`）。
- 若 `describe pod` 出现 `proxyconnect tcp: dial tcp 127.0.0.1:7897: connection refused`：在 **Docker Desktop** 关闭 HTTP/HTTPS 代理，或**先启动**本机代理（Clash 等）再 pull。
- **手动修复**（无需重装集群）：

```bash
make k8s-prepull-infra
kubectl -n im-local rollout restart deployment/postgres deployment/redis deployment/rocketmq-namesrv deployment/rocketmq-broker
kubectl -n im-local rollout status deployment/postgres --timeout=300s
```

- 云集群：镜像名、tag、`imagePullSecrets` 是否正确？

### Pod CrashLoopBackOff

```bash
kubectl -n im-local describe pod <pod-name>
kubectl -n im-local logs <pod-name> --previous
```

常见原因：Postgres/RocketMQ 未就绪、ConfigMap 中 DSN 错误。

### `kubectl wait` 报 no matching resources found

`apply` 后若立刻 `kubectl wait pod -l app=...`，ReplicaSet 可能尚未创建 Pod，会立即失败。`make up` 已改用 `kubectl rollout status deployment/...`；手动等待请同样使用 rollout。

### 基础设施 rollout 超时（postgres 等）

先查 Pod 是否 **ImagePullBackOff**（见上节），不要只靠加长等待时间。

```bash
kubectl -n im-local describe pod -l app=postgres | tail -20
kubectl -n im-local get pods -l 'app in (postgres,redis,rocketmq-namesrv)'
```

修复镜像后重试；`make up` 默认 `INFRA_TIMEOUT=600s`，可临时加大：`INFRA_TIMEOUT=900s make up`。

### WebSocket 连不上

- kind：确认 `localhost:10000` 与 kind 映射（`deploy/kind/im-local.yaml`）
- Ingress：是否配置 WSS、超时、cookie 粘性；Gateway 是否 Ready
- JWT：握手 query `token` 是否有效

### RPC 调用失败（集群内）

```bash
kubectl -n im-local exec -it deploy/gateway-api -- wget -qO- http://message-rpc:20500/health 2>/dev/null || true
kubectl -n im-local get endpoints message-rpc
```

确认 `message-rpc` Endpoints 非空。

### kind 载入镜像失败

```bash
kind delete cluster --name im-local
KIND_LOAD_MODE=single make up
```

---

## 附录：Make 目标速查

| 命令 | 说明 |
|------|------|
| `make image-base-build` | 构建 Go/Alpine 基础镜像 |
| `make deploy SVC=...` | 增量：构建 + 载入 + 滚动发布（可带 `:副本`） |
| `make build-images SVC=...` | 仅构建指定服务镜像 |
| `make k8s-etc` | 生成 K8s 配置与清单 |
| `make up` / `make k8s-up` | 构建 + 部署到 kind |
| `make down` | 删除 im-local 资源 |
| `make seed` | 测试数据 |
| `make k8s-forward` | API 端口转发 |
| `make k8s-forward-infra` | 中间件端口转发 |
| `make k8s-logs` | 跟踪 gateway 日志 |

更多架构说明见仓库根目录 [README.md](../../README.md)。
