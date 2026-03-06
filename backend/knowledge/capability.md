# 核心业务能力

系统当前具备的能力清单。新增模块时在此注册。

| 能力 | 模块 | 入口 | 说明 |
|---|---|---|---|
| HelloWorld | hello_world | `GET /api/v1/hello-world` | 示例模块，验证脚手架各层联通 |

## Redis

- 通用缓存：通过 `infraredis.Client` 直接使用 go-redis 原生 API
- 分布式锁：通过 `infraredis.RS.NewMutex("lock-key")` 创建分布式互斥锁

## Kafka

- 生产者：`infrakafka.Produce(topic, key, value)` 异步发送消息
- 消费者注册：`infrakafka.Register(topic, groupID, handler, opts...)` 注册消费函数
- 同步/异步消费：默认同步，`WithAsync()` 启用异步消费
- 失败重试 + 死信队列：重试 N 次失败后发送到 `<topic>.DLQ`

## 运维端点

- `GET /healthz` — 存活探针，直接返回 200
- `GET /readyz` — 就绪探针，检查 MySQL/Redis/Kafka 连通性，全通 200，任一失败 503
- `/debug/pprof/*` — Go pprof 性能分析端点
