# Q-DEV

AI 驱动的全栈开发脚手架。为 AI 时代的工程实践而生——让 AI 理解你的架构，让开发回归本质。

Q-DEV 不是又一个 CRUD 模板，而是一套面向 AI 协作开发的工程基座。清晰的分层约束让 AI 精准生成代码，规范的目录契约让人机协作零摩擦。

## 技术栈

- **后端**: Go + Gin + GORM Gen + Cobra
- **前端**: 待定
- **基础设施**: MySQL / Redis / Kafka
- **工程化**: Swagger 自动文档 / golangci-lint / Makefile 一键操作

## 后端架构

采用 DDD 分层架构，严格单向依赖：`http → app → domain → infra`
app 层：业务能力编排
domain 层：核心能力抽象
通过 knowledge 模块快速了解业务知识

```
backend/
├── main.go                     # 入口，仅调用 cmd.Execute()
├── cmd/                        # CLI 命令层（Cobra）
│   ├── root.go                 #   根命令，加载配置，支持 -c 指定配置文件
│   └── server.go               #   server 子命令，启动 HTTP 服务 + 优雅关停
├── conf/                       # 配置层
│   ├── conf.go                 #   Config 结构体 + Load() + 全局变量 C
│   └── config.yaml             #   YAML 配置（server/mysql/redis/kafka）
├── http/                       # HTTP 接口层（Gin）
│   ├── server.go               #   NewServer() 初始化 Engine + 中间件 + 路由
│   ├── router/                 #   路由注册，按版本管理
│   │   ├── router.go           #     Register() 入口，挂载 /api，分发到各版本
│   │   └── v1.go               #     v1 版本路由，按模块拆分 registerXxx
│   ├── api/                    #   请求处理器，对接 app 层
│   ├── common/                 #   通用工具（统一响应 OK/Fail）
│   └── middleware/             #   自定义中间件（auth/cors 等按需添加）
├── app/                        # 应用层 — 用例编排
│   └── <module>/
│       ├── app.go              #   应用服务（产品能力与业务能力的编排）
│       └── vo/                 #   值对象（入参/出参）
├── domain/                     # 领域层 — 核心业务逻辑抽象内聚
│   └── <module>/
│       └── <module>.go         #   领域模型 + 领域服务
├── infra/                      # 基础设施层 — 技术实现
│   ├── mysql/
│   │   ├── model/              #     GORM 模型定义
│   │   └── dao/                #     GORM Gen 生成的类型安全查询（自动生成，勿手动修改）
│   ├── redis/                  #     缓存
│   └── kafka/                  #     消息队列
├── knowledge/                  # 知识层 — 项目的"自我描述"，纯 Markdown，不参与运行时
│   ├── README.md               #   知识库说明
│   ├── semantic.md             #   核心语义：项目定位、系统边界、架构分层
│   ├── capability.md           #   核心业务能力清单
│   ├── model.md                #   核心数据模型与实体关系
│   └── abstraction.md          #   核心抽象：统一响应、路由版本、配置、CLI 命令
└── gen/                        # 代码生成
    ├── docs/                   #   Swagger 文档（自动生成）
    └── gorm_gen/
        └── main.go             #   GORM Gen 生成脚本，离线运行，无需连接数据库
```

## Makefile

```bash
make swagger    # 生成 Swagger 文档到 gen/docs/
make sql        # 根据 model 结构体生成类型安全查询代码到 dao/
make lint       # go vet + golangci-lint 代码检查
```


