# 核心抽象

系统的关键契约与扩展点。

## 统一响应

- **包**: `q-dev/http/common`
- **函数**: `OK(c, data)` 成功响应 / `Fail(c, err)` 失败响应
- **格式**: `{"code": 0, "message": "ok", "data": ...}`

## 路由版本管理

- **包**: `q-dev/http/router`
- **规则**: 每个版本一个文件（v1.go / v2.go），每个模块一个 `registerXxx` 函数
- **路径**: `/api/v1/...`、`/api/v2/...`

## 配置

- **包**: `q-dev/conf`
- **加载**: `conf.Load("conf/config.yaml")`
- **读取**: `conf.C.Server.Port`、`conf.C.MySQL.DSN()`

## CLI 命令

- **包**: `q-dev/cmd`
- **规则**: 每个子命令独立文件，`init()` 中注册到 `rootCmd`
- **使用**: `go run main.go server -c conf/config.yaml`

## Infra 初始化

各基础设施组件独立 Init/Close，在 cmd/server.go 中按顺序调用：
- 启动：`redis.Init() → kafka.Init() → kafka.StartConsumers()`
- 关停：`kafka.StopConsumers() → kafka.Close() → redis.Close()`
