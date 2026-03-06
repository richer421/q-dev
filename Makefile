APP_NAME := q-dev
BUILD_DIR := backend
BINARY := $(BUILD_DIR)/bin/$(APP_NAME)

.PHONY: build run swagger sql lint docker-build docker-up docker-down clean

# ---------- 构建 & 运行 ----------

build:
	cd $(BUILD_DIR) && go build -o bin/$(APP_NAME) .

run: build
	cd $(BUILD_DIR) && ./bin/$(APP_NAME) server

# ---------- 代码生成 ----------

swagger:
	cd $(BUILD_DIR) && go run github.com/swaggo/swag/cmd/swag init -o ./gen/docs --parseDependency

sql:
	cd $(BUILD_DIR) && go run ./gen/gorm_gen

# ---------- 代码检查 ----------

lint:
	cd $(BUILD_DIR) && go vet ./...
	cd $(BUILD_DIR) && golangci-lint run ./...

# ---------- Docker ----------

docker-build:
	docker build -t $(APP_NAME) -f deploy/Dockerfile .

docker-up:
	docker compose -f deploy/docker-compose.yml up -d

docker-down:
	docker compose -f deploy/docker-compose.yml down

# ---------- 清理 ----------

clean:
	rm -rf $(BUILD_DIR)/bin
