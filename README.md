# Q-DEV

AI 驱动的全栈开发脚手架。为 AI 时代的工程实践而生——让 AI 理解你的架构，让开发回归本质。

Q-DEV 不是又一个 CRUD 模板，而是一套面向 AI 协作开发的工程基座。清晰的分层约束让 AI 精准生成代码，规范的目录契约让人机协作零摩擦。

## 快速开始

### 安装

```bash
# macOS / Linux (推荐)
curl -fsSL https://github.com/richer421/q-dev/releases/latest/download/install.sh | bash

# 或指定版本
curl -fsSL https://github.com/richer421/q-dev/releases/download/v1.0.0/install.sh | bash
```

安装完成后：

```bash
qdev init my-project        # 交互式创建
qdev init my-project --backend-only  # 纯后端��式
```

<details>
<summary>其他安装方式</summary>

#### 手动下载

从 [Releases](https://github.com/richer421/q-dev/releases) 页面下载对应平台的二进制文件：

| 平台 | 文件 |
|------|------|
| macOS (Intel) | `qdev-darwin-amd64` |
| macOS (Apple Silicon) | `qdev-darwin-arm64` |
| Linux (x64) | `qdev-linux-amd64` |
| Linux (ARM64) | `qdev-linux-arm64` |
| Windows | `qdev-windows-amd64.exe` |

```bash
chmod +x qdev-*
sudo mv qdev-* /usr/local/bin/qdev
```

#### 从源码构建

```bash
git clone https://github.com/richer421/q-dev.git
cd q-dev/qdev-cli
go build -o qdev .
```

</details>

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端 | Go 1.25 + Gin + GORM Gen + Cobra |
| 前端 | React + Umi Max + Ant Design Pro |
| 数据库 | MySQL 8.0 / Redis 7 / Kafka 3.7 |
| 可观测性 | OpenTelemetry + Jaeger + Prometheus |
| 工程化 | Swagger / golangci-lint / Makefile / Air / Docker Compose |

## 项目结构

```
├── backend/                    # 后端服务
│   ├── main.go                 # 入口
│   ├── cmd/                    # CLI 命令（Cobra）
│   │   ├── root.go             #   根命令，加载配置
│   │   └── server.go           #   server 子命令
│   ├── conf/                   # 配置层
│   │   ├── conf.go             #   Config 结构体 + Load()
│   │   └── config.yaml         #   YAML 配置
│   ├── http/                   # HTTP 接口层
│   │   ├── server.go           #   Gin Engine 初始化
│   │   ├── router/             #   路由注册
│   │   ├── api/                #   请求处理器
│   │   ├── sdk/                #   HTTP 客户端 SDK
│   │   ├── common/             #   统一响应
│   │   └── middleware/         #   中间件
│   ├── app/                    # 应用层（用例编排）
│   ├── domain/                 # 领域层（核心业务）
│   ├── infra/                  # 基础设施层
│   │   ├── mysql/              #   MySQL + GORM Gen
│   │   ├── redis/              #   Redis + 分布式锁
│   │   └── kafka/              #   Kafka 生产者/消费者
│   ├── pkg/                    # 共享包
│   │   ├── logger/             #   Zap 日志
│   │   └── otel/               #   OpenTelemetry
│   ├── knowledge/              # 知识层（AI 理解入口）
│   └── gen/                    # 代码生成
│       ├── docs/               #   Swagger 文档
│       └── gorm_gen/           #   GORM Gen 脚本
├── frontend/                   # 前端应用（Ant Design Pro）
│   ├── .umirc.ts               #   Umi 配置
│   └── src/                    #   源码
├── deploy/                     # 部署配置
│   ├── Dockerfile              #   后端镜像
│   ├── Dockerfile.frontend     #   前端镜像
│   ├── docker-compose.yml      #   全栈编排
│   ├── nginx.conf              #   Nginx 配置
│   ├── otel-collector.yaml     #   OTel Collector
│   └── prometheus.yml          #   Prometheus 配置
├── qdev-cli/                   # 脚手架 CLI 工具
└── Makefile                    # 构建命令
```

## 架构设计

采用 DDD 分层架构，严格单向依赖：`http → app → domain → infra`

```
┌─────────────────────────────────────────────────────────┐
│                      HTTP Layer                          │
│  请求处理 · 参数校验 · 统一响应 · 路由版本管理            │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│                      App Layer                           │
│  业务能力编排 · VO 转换 · 用例协调                        │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│                     Domain Layer                         │
│  核心业务逻辑抽象内聚 · 领域服务                          │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│                     Infra Layer                          │
│  MySQL · Redis · Kafka · 技术实现                        │
└─────────────────────────────────────────────────────────┘
```

## Makefile 命令

```bash
# 构建 & 运行
make build          # 编译到 backend/bin/q-dev
make run            # 编译 + 运行 server
make dev            # Air 热重载

# 代码生成
make swagger        # 生成 Swagger 文档
make sql            # 生成 GORM Gen 代码

# 代码检查
make lint           # go vet + golangci-lint
make test           # 运行测试
make cover          # 生成覆盖率报告

# 前端
make fe-install     # 安装前端依赖
make fe-dev         # 前端开发服务器
make fe-build       # 前端构建

# 基础设施
make infra-up       # 启动 MySQL/Redis/Kafka/Jaeger/Prometheus
make infra-down     # 停止基础设施

# Docker
make docker-build   # 构建镜像
make docker-up      # 全栈部署
```

## 开发约定

### 分层规则

- 严格单向依赖，禁止反向引用
- 所有方法第一个参数为 `context.Context`
- domain 层不感知 HTTP/配置，直接调用 DAO
- app 层负责 VO ↔ Model 转换

### 命名规范

| 类型 | 规范 | 示例 |
|------|------|------|
| Go 包名 | snake_case | `hello_world` |
| API 路由 | kebab-case | `/api/v1/hello-world` |
| 数据库表名 | snake_case 复数 | `hello_worlds` |

### 新增模块流程

1. `infra/mysql/model/` — 定义 GORM 模型
2. `gen/gorm_gen/main.go` — 注册模型，`make sql` 生成 DAO
3. `domain/<module>/` — 实现领域服务
4. `app/<module>/` — 实现应用服务 + VO
5. `http/api/` — 实现 API 处理器
6. `http/router/v1.go` — 注册路由
7. `http/sdk/` — 实现客户端 SDK
8. `knowledge/` — 更新文档
9. `make swagger` — 更新 API 文档

## 统一响应格式

```json
{"code": 0, "message": "ok", "data": ...}       // 成功
{"code": -1, "message": "错误信息"}               // 失败
```

## 运维端点

| 端点 | 用途 |
|------|------|
| `/healthz` | 存活探针 |
| `/readyz` | 就绪探针（检查依赖） |
| `/debug/pprof/*` | Go 性能分析 |

## qdev-cli 脚手架工具

```bash
# 交互式创建
./qdev init my-project

# 纯后端模式
./qdev init my-project --backend-only

# 指定模板仓库
./qdev init my-project --repo https://github.com/user/q-dev

# 指定版本
./qdev init my-project --tag v1.0.0

# 私有仓库认证
./qdev init my-project --git-token ghp_xxx
./qdev init my-project --ssh-key ~/.ssh/id_rsa

# 强制覆盖
./qdev init my-project --force
```

## License

MIT
