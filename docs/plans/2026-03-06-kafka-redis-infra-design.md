# Kafka & Redis Infra 层实现设计

## 概述

为 Q-DEV 脚手架实现 Redis 和 Kafka 基础设施层，采用薄封装 + 全局变量方案，与现有 MySQL DAO 全局 `Q` 变量风格保持一致。

## 技术选型

| 组件 | 库 | 版本 |
|------|-----|------|
| Redis 客户端 | `github.com/redis/go-redis/v9` | latest |
| 分布式锁 | `github.com/go-redsync/redsync/v4` | latest |
| Kafka 客户端 | `github.com/confluentinc/confluent-kafka-go/v2` | latest |

## 目录结构

```
infra/
├── mysql/          # 已有
│   ├── model/
│   └── dao/
├── redis/
│   └── redis.go    # Init / Close / Client / RS
└── kafka/
    └── kafka.go    # Init / Close / Produce / Register / StartConsumers / StopConsumers
```

---

## Redis 设计

### 全局变量

```go
var (
    Client *redis.Client      // go-redis 客户端
    RS     *redsync.Redsync   // redsync 分布式锁实例
)
```

### API

```go
func Init(cfg conf.RedisConfig) error   // 创建 Client + Ping + 初始化 RS
func Close() error                       // 关闭连接
```

### 功能

- **缓存**：直接使用 `redis.Client` 原生 API（Get/Set/Del 等）
- **分布式锁**：通过 `redis.RS.NewMutex("lock-key")` 创建锁，调用 Lock/Unlock

### 不做额外封装

保持薄封装原则，业务层直接调用 go-redis 和 redsync 原生 API。

---

## Kafka 设计

### 全局变量

```go
var Producer *confluent.Producer
```

### API

```go
// 生产者
func Init(cfg conf.KafkaConfig) error
func Close()
func Produce(topic string, key, value []byte) error

// 消费者注册 + 启动
type HandleFunc func(msg *confluent.Message) error
type ConsumerOption func(*consumerConfig)

func Register(topic, groupID string, handler HandleFunc, opts ...ConsumerOption)
func StartConsumers(ctx context.Context) error
func StopConsumers()

// 选项
func WithAsync() ConsumerOption           // 异步消费模式
func WithMaxRetries(n int) ConsumerOption // 最大重试次数，默认 3
```

### Consumer 注册模式

1. 业务模块在启动阶段调用 `kafka.Register("order.created", "order-group", handler)`
2. `cmd/server.go` 在 HTTP 服务启动后调用 `kafka.StartConsumers(ctx)` 统一拉起消费者
3. 每个注册项启动一个 goroutine，内部循环 Poll 消费

### 消费流程

1. Poll 取到消息 → 调用 handler
2. handler 成功 → 提交 offset
3. handler 失败 → 原地重试，最多 N 次（默认 3）
4. 仍失败 → 将消息发到 `<topic>.DLQ`（死信队列） → 提交 offset → 继续消费

### 同步 vs 异步消费

- **同步（默认）**：handler 返回后才提交 offset，继续 Poll 下一条
- **异步（WithAsync）**：handler 在独立 goroutine 执行，消费循环不阻塞，通过回调通知完成状态后提交 offset

---

## 配置

### conf.RedisConfig 补充

```yaml
redis:
  addr: localhost:6379
  password: ""
  db: 0
  pool_size: 10
```

### conf.KafkaConfig 补充

```yaml
kafka:
  brokers:
    - localhost:9092
  max_retries: 3
```

---

## 初始化 & 生命周期

### 启动顺序（cmd/server.go）

```
conf.Load() → logger.Init() → redis.Init() → kafka.Init() → kafka.StartConsumers() → HTTP Server
```

### 优雅关停（SIGINT/SIGTERM）

```
HTTP Server.Shutdown() → kafka.StopConsumers() → kafka.Close() → redis.Close()
```

各组件独立 Init，`cmd/server.go` 分别调用。
