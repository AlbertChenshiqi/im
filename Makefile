.PHONY: deps build build-images image-base-build test loadtest gen run-all \
	k8s-etc k8s-up k8s-down k8s-forward k8s-forward-infra k8s-seed k8s-logs \
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

# 构建容器镜像（K8s / kind 使用，需 Docker 守护进程）
image-base-build:
	chmod +x scripts/image-base-build.sh
	./scripts/image-base-build.sh

build-images:
	@docker image inspect $(GO_BASE_IMAGE) >/dev/null 2>&1 || { \
		echo "未找到 $(GO_BASE_IMAGE)，先执行: make image-base-build"; \
		exit 1; \
	}
	@docker image inspect $(RUNTIME_BASE_IMAGE) >/dev/null 2>&1 || { \
		echo "未找到 $(RUNTIME_BASE_IMAGE)，先执行: make image-base-build"; \
		exit 1; \
	}
	chmod +x scripts/build-images.sh
	GO_BASE_IMAGE=$(GO_BASE_IMAGE) RUNTIME_BASE_IMAGE=$(RUNTIME_BASE_IMAGE) ./scripts/build-images.sh

# --- 本地 Kubernetes（默认开发方式）---
k8s-etc:
	chmod +x scripts/write-k8s-etc.sh scripts/gen-k8s-manifests.sh
	./scripts/write-k8s-etc.sh
	GATEWAY_REPLICAS=$(GATEWAY_REPLICAS) ./scripts/gen-k8s-manifests.sh

k8s-up:
	chmod +x scripts/k8s-up.sh scripts/k8s-load-images.sh
	./scripts/k8s-up.sh

k8s-down:
	chmod +x scripts/k8s-down.sh
	./scripts/k8s-down.sh

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

# 常用别名
up: k8s-up
down: k8s-down
seed: k8s-seed

# 本机进程跑服务（需先 k8s-up，另开终端 make k8s-forward-infra）
run-all: build
	./scripts/run-local.sh

migrate:
	psql "postgres://im:im@localhost:5432/im?sslmode=disable" -f migrations/001_init.sql

test:
	go test $$(go list ./... | grep -v scripts/loadtest)

loadtest:
	go run ./scripts/loadtest
