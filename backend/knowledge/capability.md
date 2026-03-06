# 核心业务能力

系统当前具备的能力清单。新增模块时在此注册。

| 能力 | 模块 | 入口 | 说明 |
|---|---|---|---|
| HelloWorld CRUD | hello_world | `/api/v1/hello-world` | 示例模块，完整 CRUD，验证脚手架各层联通 |

### HelloWorld 接口清单

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/hello-world | 分页列表 |
| GET | /api/v1/hello-world/:id | 详情 |
| POST | /api/v1/hello-world | 创建 |
| PUT | /api/v1/hello-world/:id | 更新 |
| DELETE | /api/v1/hello-world/:id | 删除 |

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

## 前端

- 框架：Ant Design Pro（umi max）
- UI：antd 5.x
- 开发地址：`http://localhost:8000`
- API 代理：`/api/*` → `http://localhost:8080`
