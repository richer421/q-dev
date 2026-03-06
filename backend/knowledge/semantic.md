# 核心语义

## 项目

- **名称**: Q-DEV
- **定位**: AI 驱动的全栈开发脚手架
- **领域**: AI 全栈工程化

## 系统边界

做什么：
- 提供后端 DDD 分层脚手架（Go + Gin + GORM Gen + Cobra）
- 提供代码生成能力（Swagger 文档 / GORM Gen 类型安全查询）
- 提供 CLI 多命令入口（server / migrate / cron 等）
- 提供 knowledge 层，让 AI 理解项目结构并参与开发

不做什么：
- 不包含具体业务实现，仅提供工程基座
- 不做反包设计，domain 直接依赖 infra，务实优先

## 架构分层

`http → app → domain → infra`，严格单向依赖，上层依赖下层。

| 层 | 职责 |
|---|---|
| cmd | CLI 命令入口，Cobra 注册子命令 |
| conf | 配置加载，YAML → 全局变量 conf.C |
| http | Gin 路由、中间件、API 处理器、统一响应 |
| app | 用例编排，产品能力与业务能力的编排 |
| domain | 核心业务逻辑抽象内聚 |
| infra | 技术实现（MySQL / Redis / Kafka） |
| knowledge | 项目知识库，供 AI 理解系统（不参与运行时） |
| gen | 代码生成产物（Swagger 文档 / GORM Gen 查询） |
