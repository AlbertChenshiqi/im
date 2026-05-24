.PHONY: deps build build-images image-base-build deploy k8s-load-images k8s-prepull-infra test loadtest gen run-all dev-stop \
	k8s-etc k8s-up k8s-down k8s-forward k8s-forward-infra k8s-seed k8s-logs \
	host-infra-up host-infra-down \
	up down seed migrate

GO_BASE_IMAGE ?= im-go-base:deps
RUNTIME_BASE_IMAGE ?= im-go-base:runtime
GATEWAY_REPLICAS ?= 2

deps:
	go mod tidy

gen:
	./scripts/gen-gozero.sh

# 本机二进制（非容器内运行）
build:
	go build -o bin/gateway-api ./apps/gateway/api
	go build -o bin/user-api ./apps/user/api
	go build -o bin/user-rpc ./apps/user/rpc
	go build -o bin/friend-api ./apps/friend/api
	go build -o bin/friend-rpc ./apps/friend/rpc
	go build -o bin/group-api ./apps/group/api
	go build -o bin/group-rpc ./apps/group/rpc
	go build -o bin/conversation-api ./apps/conversation/api
	go build -o bin/conversation-rpc ./apps/conversation/rpc
	go build -o bin/message-api ./apps/message/api
	go build -o bin/message-rpc ./apps/message/rpc
	go build -o bin/notification-api ./apps/notification/api
	go build -o bin/notification-rpc ./apps/notification/rpc
	go build -o bin/push-api ./apps/push/api
	go build -o bin/push-rpc ./apps/push/rpc
	go build -o bin/cron ./apps/cron

image-base-build:
	chmod +x scripts/image-base-build.sh
	./scripts/image-base-build.sh

# 增量部署（推荐）：构建 → 载入 kind → 调副本（可选）→ 滚动发布
#   make deploy SVC=message-rpc
#   make deploy SVC=gateway-api:2,message-rpc,cron
#   make deploy SVC=gateway-api,message-rpc REPLICAS=1
deploy:
	chmod +x scripts/k8s-deploy.sh scripts/build-images.sh scripts/k8s-load-images.sh
	SVC="$(SVC)" BINS="$(BINS)" REPLICAS="$(REPLICAS)" ./scripts/k8s-deploy.sh

# 仅构建 / 仅载入（需 SVC 或 BINS，与 deploy 相同）
build-images:
	@docker image inspect $(GO_BASE_IMAGE) >/dev/null 2>&1 || { \
		echo "未找到 $(GO_BASE_IMAGE)，先执行: make image-base-build"; exit 1; }
	@docker image inspect $(RUNTIME_BASE_IMAGE) >/dev/null 2>&1 || { \
		echo "未找到 $(RUNTIME_BASE_IMAGE)，先执行: make image-base-build"; exit 1; }
	SKIP_LOAD=1 SKIP_ROLLOUT=1 $(MAKE) deploy SVC="$(SVC)" BINS="$(BINS)"

k8s-load-images:
	SKIP_BUILD=1 SKIP_ROLLOUT=1 $(MAKE) deploy SVC="$(SVC)" BINS="$(BINS)"

# --- 本地 Kubernetes ---
k8s-etc:
	chmod +x scripts/write-k8s-etc.sh scripts/gen-k8s-manifests.sh
	HOST_INFRA="$(HOST_INFRA)" ./scripts/write-k8s-etc.sh
	GATEWAY_REPLICAS=$(GATEWAY_REPLICAS) ./scripts/gen-k8s-manifests.sh

# 本机 Docker 跑 Postgres / Redis / RocketMQ（deploy/docker/docker-compose.yml）
host-infra-up:
	chmod +x scripts/host-infra-up.sh
	./scripts/host-infra-up.sh

host-infra-down:
	chmod +x scripts/host-infra-down.sh
	./scripts/host-infra-down.sh

# 直接使用 compose（等价于 host-infra-up/down）
infra-up: host-infra-up
infra-down: host-infra-down

k8s-up:
	chmod +x scripts/k8s-up.sh scripts/k8s-load-images.sh scripts/k8s-prepull-infra.sh scripts/host-infra-up.sh
	HOST_INFRA="$(HOST_INFRA)" ./scripts/k8s-up.sh

# 仅预拉取并载入 Postgres / Redis / RocketMQ（修复 ImagePullBackOff）
k8s-prepull-infra:
	chmod +x scripts/k8s-prepull-infra.sh
	./scripts/k8s-prepull-infra.sh

k8s-down:
	chmod +x scripts/k8s-down.sh
	DELETE_CLUSTER="$(DELETE_CLUSTER)" DELETE_DASHBOARD="$(DELETE_DASHBOARD)" ./scripts/k8s-down.sh

# 停止本机 go run 进程（run-all 启动的）
dev-stop:
	chmod +x scripts/dev-stop.sh
	./scripts/dev-stop.sh

k8s-forward:
	chmod +x scripts/k8s-forward.sh
	./scripts/k8s-forward.sh

k8s-forward-infra:
	chmod +x scripts/k8s-forward-infra.sh
	./scripts/k8s-forward-infra.sh

k8s-seed:
	K8S_MODE=1 ./scripts/reset-dev-data.sh

k8s-logs:
	kubectl -n im-local logs -f deploy/gateway-api --tail=100

# 本机调试（默认）：Docker 基础设施 + 本机 go run
up: host-infra-up

down: dev-stop host-infra-down

seed:
	K8S_MODE=0 ./scripts/reset-dev-data.sh

# Kubernetes 部署（需要时）
#   make k8s-up
#   HOST_INFRA=1 make k8s-up
#   DELETE_CLUSTER=1 make k8s-down

run-all: build
	./scripts/run-local.sh

migrate:
	@if command -v psql >/dev/null 2>&1; then \
		psql "postgres://im:im@localhost:5432/im?sslmode=disable" -v ON_ERROR_STOP=1 -f migrations/001_init.sql; \
	elif docker inspect im-postgres >/dev/null 2>&1; then \
		docker exec -i im-postgres psql -U im -d im -v ON_ERROR_STOP=1 < migrations/001_init.sql; \
	else \
		echo "需要 psql 或运行中的 im-postgres 容器（先 make up）"; exit 1; \
	fi

test:
	go test $$(go list ./... | grep -v scripts/loadtest)

loadtest:
	go run ./scripts/loadtest
