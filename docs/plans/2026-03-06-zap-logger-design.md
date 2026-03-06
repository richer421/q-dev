# Zap Logger 设计文档

## 背景

当前后端使用标准库 `log` 和 Gin 默认 Logger，缺乏结构化日志、级别控制和文件输出能力。需要引入 zap 作为统一日志组件。

## 方案

zap + lumberjack。zap 做日志引擎，lumberjack 做文件轮转。

## 目录与文件

新增：
- `pkg/logger/logger.go` — 初始化 + 全局 format 函数
- `http/middleware/logger.go` — Gin 请求日志中间件
- `http/middleware/recovery.go` — Gin panic 恢复中间件

修改：
- `conf/conf.go` — 新增 LogConfig / LogFileConfig
- `conf/config.yaml` — 新增 log 配置段
- `http/server.go` — 替换 gin.Logger/Recovery
- `cmd/server.go` — 替换标准库 log
- `cmd/root.go` — 初始化 logger

## 配置

```yaml
log:
  level: "info"             # debug / info / warn / error
  format: "console"         # console / json
  file:
    enabled: false
    path: "logs/app.log"
    max_size: 100           # 单文件最大 MB
    max_age: 30             # 旧文件保留天数
    compress: true          # gzip 压缩旧文件
```

对应结构体：

```go
type LogConfig struct {
    Level  string        `yaml:"level"`
    Format string        `yaml:"format"`
    File   LogFileConfig `yaml:"file"`
}

type LogFileConfig struct {
    Enabled bool   `yaml:"enabled"`
    Path    string `yaml:"path"`
    MaxSize int    `yaml:"max_size"`
    MaxAge  int    `yaml:"max_age"`
    Compress bool  `yaml:"compress"`
}
```

## API

只暴露 SugaredLogger 的 format 风格：

```go
func Init(cfg conf.LogConfig)
func Sync()

func Debugf(template string, args ...interface{})
func Infof(template string, args ...interface{})
func Warnf(template string, args ...interface{})
func Errorf(template string, args ...interface{})
func Fatalf(template string, args ...interface{})
```

## Init 逻辑

1. 根据 format 选择 encoder（console / json）
2. 始终创建 stdout WriteSyncer
3. file.enabled 时额外创建 lumberjack WriteSyncer，用 zapcore.NewTee 合并
4. 根据 level 设置日志级别
5. 赋值给包级全局 SugaredLogger

## Gin 集成

- middleware/logger.go：记录 method、path、status、latency、client_ip
- middleware/recovery.go：捕获 panic，用 logger.Errorf 输出堆栈，返回 500
- server.go 替换 gin.Logger() 和 gin.Recovery()

## 不做的事

- 不封装 zap.Field 风格 API
- 不做日志 context 传递
- 不做按天切文件名，用 lumberjack 按大小轮转 + maxAge 清理

## 依赖

- go.uber.org/zap
- gopkg.in/natefinch/lumberjack.v2
