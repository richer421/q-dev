# 健康检查 & pprof 设计

## 概述

为 Q-DEV 添加健康检查探针和 pprof 性能分析端点，全部挂在主 Gin Engine 上。

## 端点

| 路径 | 用途 | 响应 |
|------|------|------|
| `GET /healthz` | 存活探针 | 直接返回 200 `{"status":"ok"}` |
| `GET /readyz` | 就绪探针 | 检查 MySQL/Redis/Kafka 连通性，全通 200，任一失败 503 |
| `/debug/pprof/*` | pprof 性能分析 | 标准 Go pprof 端点 |

## 实现位置

在 `http/router/router.go` 的 `Register()` 函数中注册，不走 `/api` 前缀，直接挂在根路径。

## readyz 检查逻辑

- MySQL：`db.Raw("SELECT 1").Scan()`
- Redis：`Client.Ping(ctx)`
- Kafka：`Producer.GetMetadata(nil, true, 3000)`
- 任一组件失败：返回 503，响应体标明哪个组件不健康

## pprof

使用 `gin-contrib/pprof` 库一行注册，挂载到 `/debug/pprof`。
