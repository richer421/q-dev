# 可观测性设计文档（Metric + Trace）

## 背景

项目当前零可观测性组件。需要为 HTTP 层和 MySQL 层接入 metric 和 trace，后端统一通过 OpenTelemetry Collector 收集。

## 方案

OpenTelemetry Go SDK 全家桶。统一管理 metric + trace，Gin 和 GORM 使用社区插件自动采集。

## 目录结构

新增：
- `pkg/otel/otel.go` — OTel 初始化（TracerProvider + MeterProvider）+ PrometheusHandler()
- `http/middleware/otel.go` — otelgin 中间件封装

修改：
- `conf/conf.go` — 新增 OTelConfig / PrometheusConfig
- `conf/config.yaml` — 新增 otel 配置段
- `http/server.go` — 注册 OTel 中间件 + /metrics 路由
- `cmd/server.go` — 初始化 OTel + defer shutdown

## 配置

```yaml
otel:
  enabled: false
  service_name: "q-dev"
  endpoint: "localhost:4317"
  prometheus:
    enabled: true
    path: "/metrics"
```

```go
type OTelConfig struct {
    Enabled     bool             `yaml:"enabled"`
    ServiceName string           `yaml:"service_name"`
    Endpoint    string           `yaml:"endpoint"`
    Prometheus  PrometheusConfig `yaml:"prometheus"`
}

type PrometheusConfig struct {
    Enabled bool   `yaml:"enabled"`
    Path    string `yaml:"path"`
}
```

## OTel 初始化（pkg/otel/otel.go）

```go
func Init(cfg conf.OTelConfig) (shutdown func(context.Context) error, err error)
func PrometheusHandler() http.Handler
```

Init 逻辑：
1. 创建 Resource（service.name）
2. Trace：OTLP gRPC exporter → TracerProvider（BatchSpanProcessor）→ 设为全局
3. Metric：OTLP gRPC PeriodicReader + Prometheus exporter → MeterProvider → 设为全局
4. enabled: false 时直接返回空 shutdown

## HTTP 集成

中间件注册顺序：
```go
r.Use(middleware.OTel())        // 最先，确保所有请求被采集
r.Use(middleware.Logger())
r.Use(middleware.Recovery())
```

Prometheus /metrics 端点：
```go
if conf.C.OTel.Prometheus.Enabled {
    r.GET(conf.C.OTel.Prometheus.Path, gin.WrapH(otel.PrometheusHandler()))
}
```

## cmd/server.go 集成

```go
shutdown, err := otel.Init(conf.C.OTel)
if err != nil {
    logger.Fatalf("otel init: %s", err)
}
defer shutdown(context.Background())
```

## GORM 集成

本次只列依赖，实际集成在 MySQL 初始化实现时做：
```go
db.Use(otelgorm.NewPlugin())
```

## 不做的事

- 不做自定义 metric 定义
- 不做 trace sampling 配置（默认 AlwaysSample）
- 不做 Redis/Kafka instrumentation
- GORM otelgorm 插件本次不实现

## 依赖

- go.opentelemetry.io/otel
- go.opentelemetry.io/otel/sdk
- go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc
- go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc
- go.opentelemetry.io/otel/exporters/prometheus
- go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin
- github.com/uptrace/opentelemetry-go-extra/otelgorm（MySQL 时用）
