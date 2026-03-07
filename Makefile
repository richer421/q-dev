APP_NAME := q-dev
BUILD_DIR := backend
BINARY := $(BUILD_DIR)/bin/$(APP_NAME)

# 热加载调试的子命令，默认 server，可通过 make dev CMD=xxx 覆盖
CMD ?= server

.PHONY: build run dev swagger sql lint test cover fe-install fe-dev fe-build fe-lint infra-up infra-down infra-logs docker-build docker-up docker-down clean

# ---------- 构建 & 运行 ----------

build:
	cd $(BUILD_DIR) && go build -o bin/$(APP_NAME) .

run: build
	cd $(BUILD_DIR) && ./bin/$(APP_NAME) server

# ---------- 热加载调试 ----------

dev:
	cd $(BUILD_DIR) && air -- $(CMD)

# ---------- 代码生成 ----------

swagger:
	cd $(BUILD_DIR) && go run github.com/swaggo/swag/cmd/swag init -o ./gen/docs --parseDependency

sql:
	cd $(BUILD_DIR) && go run ./gen/gorm_gen

# ---------- 代码检查 ----------

lint:
	cd $(BUILD_DIR) && go vet ./...
	cd $(BUILD_DIR) && golangci-lint run ./...

# ---------- 测试 ----------

test:
	cd $(BUILD_DIR) && go test ./... -v -count=1

cover:
	cd $(BUILD_DIR) && go test ./... -coverprofile=coverage.out -count=1
	cd $(BUILD_DIR) && go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: $(BUILD_DIR)/coverage.html"

# ---------- 前端 ----------

fe-install:
	cd frontend && pnpm install

fe-dev:
	cd frontend && pnpm dev

fe-build:
	cd frontend && pnpm build

fe-lint:
	cd frontend && pnpm lint
	cd frontend && pnpm format:check

# ---------- 基础设施（本地调试） ----------

COMPOSE := docker compose -f deploy/docker-compose.yml

infra-up:
	$(COMPOSE) up -d

infra-down:
	$(COMPOSE) down

infra-logs:
	$(COMPOSE) logs -f

# ---------- Docker（全栈部署） ----------

docker-build:
	docker build -t $(APP_NAME) -f deploy/Dockerfile .
	docker build -t $(APP_NAME)-web -f deploy/Dockerfile.frontend .

docker-up:
	$(COMPOSE) --profile deploy up -d

docker-down:
	$(COMPOSE) --profile deploy down

# ---------- 清理 ----------

clean:
	rm -rf $(BUILD_DIR)/bin
