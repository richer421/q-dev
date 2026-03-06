# Kafka & Redis Infra 层实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 为 Q-DEV 实现 Redis（缓存 + 分布式锁）和 Kafka（Producer + Consumer 注册模式 + 死信队列）infra 层。

**Architecture:** 薄封装 + 全局变量，与现有 MySQL DAO 风格一致。各组件独立 Init/Close，在 cmd/server.go 中按顺序初始化和优雅关停。

**Tech Stack:** go-redis/v9, go-redsync/v4, confluent-kafka-go/v2

**Design Doc:** `docs/plans/2026-03-06-kafka-redis-infra-design.md`

---

### Task 1: 扩展配置结构体

**Files:**
- Modify: `backend/conf/conf.go:45-53`
- Modify: `backend/conf/config.yaml:20-27`

**Step 1: 给 RedisConfig 添加 PoolSize 字段**

在 `backend/conf/conf.go` 的 `RedisConfig` struct 中添加：

```go
type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	PoolSize int    `yaml:"pool_size"`
}
```

**Step 2: 给 KafkaConfig 添加 MaxRetries 字段**

```go
type KafkaConfig struct {
	Brokers    []string `yaml:"brokers"`
	MaxRetries int      `yaml:"max_retries"`
}
```

**Step 3: 更新 config.yaml 添加新字段默认值**

```yaml
redis:
  addr: "127.0.0.1:6379"
  password: ""
  db: 0
  pool_size: 10

kafka:
  brokers:
    - "127.0.0.1:9092"
  max_retries: 3
```

**Step 4: 验证编译通过**

Run: `cd backend && go build ./...`
Expected: 编译成功，无错误

**Step 5: Commit**

```bash
git add conf/conf.go conf/config.yaml
git commit -m "feat(conf): add pool_size for redis and max_retries for kafka config"
```

---

### Task 2: 安装依赖

**Step 1: 安装 go-redis、redsync、confluent-kafka-go**

```bash
cd backend
go get github.com/redis/go-redis/v9
go get github.com/go-redsync/redsync/v4
go get github.com/go-redsync/redsync/v4/redis/goredis/v9
go get github.com/confluentinc/confluent-kafka-go/v2/kafka
```

**Step 2: 验证 go.mod 已更新**

Run: `cd backend && go mod tidy && go build ./...`
Expected: 编译成功

**Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "feat(deps): add go-redis, redsync, confluent-kafka-go dependencies"
```

---

### Task 3: 实现 Redis infra 层

**Files:**
- Create: `backend/infra/redis/redis.go`

**Step 1: 创建 redis.go**

```go
package redis

import (
	"context"
	"time"

	"q-dev/conf"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
)

var (
	Client *redis.Client
	RS     *redsync.Redsync
)

func Init(cfg conf.RedisConfig) error {
	poolSize := cfg.PoolSize
	if poolSize <= 0 {
		poolSize = 10
	}

	Client = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: poolSize,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := Client.Ping(ctx).Err(); err != nil {
		return err
	}

	pool := goredis.NewPool(Client)
	RS = redsync.New(pool)

	return nil
}

func Close() error {
	if Client != nil {
		return Client.Close()
	}
	return nil
}
```

**Step 2: 验证编译通过**

Run: `cd backend && go build ./...`
Expected: 编译成功

**Step 3: Commit**

```bash
git add infra/redis/redis.go
git commit -m "feat(infra): implement redis client with go-redis and redsync"
```

---

### Task 4: 实现 Kafka infra 层 — Producer

**Files:**
- Create: `backend/infra/kafka/kafka.go`

**Step 1: 创建 kafka.go，先实现 Producer 部分**

```go
package kafka

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"q-dev/conf"
	"q-dev/pkg/logger"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

var (
	Producer   *kafka.Producer
	brokers    string
	maxRetries int
)

func Init(cfg conf.KafkaConfig) error {
	brokers = strings.Join(cfg.Brokers, ",")
	maxRetries = cfg.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}

	var err error
	Producer, err = kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": brokers,
	})
	if err != nil {
		return fmt.Errorf("create kafka producer: %w", err)
	}

	// 异步处理 delivery report
	go func() {
		for e := range Producer.Events() {
			if m, ok := e.(*kafka.Message); ok && m.TopicPartition.Error != nil {
				logger.Errorf("kafka delivery failed: %s, topic: %s",
					m.TopicPartition.Error, *m.TopicPartition.Topic)
			}
		}
	}()

	return nil
}

func Produce(topic string, key, value []byte) error {
	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          value,
	}
	if key != nil {
		msg.Key = key
	}
	return Producer.Produce(msg, nil)
}

func Close() {
	stopConsumers()
	if Producer != nil {
		Producer.Flush(5000)
		Producer.Close()
	}
}
```

**Step 2: 验证编译通过**

Run: `cd backend && go build ./...`
Expected: 可能因缺少 stopConsumers 报错，属正常，Task 5 补齐

---

### Task 5: 实现 Kafka infra 层 — Consumer 注册 + 消费循环

**Files:**
- Modify: `backend/infra/kafka/kafka.go`（在已有文件末尾追加）

**Step 1: 在 kafka.go 中追加 Consumer 注册与消费逻辑**

在 kafka.go 中追加以下代码：

```go
// HandleFunc 消费处理函数签名
type HandleFunc func(msg *kafka.Message) error

// ConsumerOption 消费者选项
type ConsumerOption func(*consumerConfig)

type consumerConfig struct {
	async      bool
	maxRetries int
}

type registration struct {
	topic   string
	groupID string
	handler HandleFunc
	config  consumerConfig
}

var (
	registrations []registration
	consumers     []*kafka.Consumer
	cancelFunc    context.CancelFunc
	wg            sync.WaitGroup
)

// WithAsync 标记为异步消费模式
func WithAsync() ConsumerOption {
	return func(c *consumerConfig) {
		c.async = true
	}
}

// WithMaxRetries 设置最大重试次数，覆盖全局配置
func WithMaxRetries(n int) ConsumerOption {
	return func(c *consumerConfig) {
		c.maxRetries = n
	}
}

// Register 注册消费函数，在 StartConsumers 之前调用
func Register(topic, groupID string, handler HandleFunc, opts ...ConsumerOption) {
	cfg := consumerConfig{maxRetries: maxRetries}
	for _, opt := range opts {
		opt(&cfg)
	}
	registrations = append(registrations, registration{
		topic:   topic,
		groupID: groupID,
		handler: handler,
		config:  cfg,
	})
}

// StartConsumers 启动所有已注册的消费者
func StartConsumers(ctx context.Context) error {
	var consumeCtx context.Context
	consumeCtx, cancelFunc = context.WithCancel(ctx)

	for _, reg := range registrations {
		c, err := kafka.NewConsumer(&kafka.ConfigMap{
			"bootstrap.servers":  brokers,
			"group.id":           reg.groupID,
			"auto.offset.reset":  "earliest",
			"enable.auto.commit": false,
		})
		if err != nil {
			return fmt.Errorf("create consumer for topic %s: %w", reg.topic, err)
		}

		if err := c.Subscribe(reg.topic, nil); err != nil {
			c.Close()
			return fmt.Errorf("subscribe topic %s: %w", reg.topic, err)
		}

		consumers = append(consumers, c)

		wg.Add(1)
		if reg.config.async {
			go runAsyncConsumer(consumeCtx, c, reg)
		} else {
			go runSyncConsumer(consumeCtx, c, reg)
		}
	}

	return nil
}

func runSyncConsumer(ctx context.Context, c *kafka.Consumer, reg registration) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			ev := c.Poll(100)
			if ev == nil {
				continue
			}
			msg, ok := ev.(*kafka.Message)
			if !ok {
				continue
			}
			handleWithRetry(c, msg, reg)
		}
	}
}

func runAsyncConsumer(ctx context.Context, c *kafka.Consumer, reg registration) {
	defer wg.Done()
	sem := make(chan struct{}, 64) // 控制并发上限
	var asyncWg sync.WaitGroup

	for {
		select {
		case <-ctx.Done():
			asyncWg.Wait()
			return
		default:
			ev := c.Poll(100)
			if ev == nil {
				continue
			}
			msg, ok := ev.(*kafka.Message)
			if !ok {
				continue
			}

			sem <- struct{}{}
			asyncWg.Add(1)
			go func(m *kafka.Message) {
				defer func() {
					<-sem
					asyncWg.Done()
				}()
				handleWithRetry(c, m, reg)
			}(msg)
		}
	}
}

func handleWithRetry(c *kafka.Consumer, msg *kafka.Message, reg registration) {
	var err error
	for i := 0; i <= reg.config.maxRetries; i++ {
		err = reg.handler(msg)
		if err == nil {
			if _, commitErr := c.CommitMessage(msg); commitErr != nil {
				logger.Errorf("kafka commit offset failed: %s, topic: %s", commitErr, reg.topic)
			}
			return
		}
		logger.Warnf("kafka consume retry %d/%d failed: %s, topic: %s",
			i+1, reg.config.maxRetries, err, reg.topic)
	}

	// 重试耗尽，发送到死信队列
	dlqTopic := reg.topic + ".DLQ"
	dlqErr := Produce(dlqTopic, msg.Key, msg.Value)
	if dlqErr != nil {
		logger.Errorf("kafka send to DLQ failed: %s, topic: %s, original error: %s",
			dlqErr, dlqTopic, err)
	} else {
		logger.Errorf("kafka message sent to DLQ: %s, original error: %s", dlqTopic, err)
	}

	// 提交 offset，继续消费
	if _, commitErr := c.CommitMessage(msg); commitErr != nil {
		logger.Errorf("kafka commit offset after DLQ failed: %s, topic: %s", commitErr, reg.topic)
	}
}

func stopConsumers() {
	if cancelFunc != nil {
		cancelFunc()
	}
	wg.Wait()
	for _, c := range consumers {
		c.Close()
	}
	consumers = nil
}

// StopConsumers 优雅停止所有消费者
func StopConsumers() {
	stopConsumers()
}
```

**Step 2: 验证编译通过**

Run: `cd backend && go build ./...`
Expected: 编译成功

**Step 3: Commit**

```bash
git add infra/kafka/kafka.go
git commit -m "feat(infra): implement kafka producer, consumer registration, retry and DLQ"
```

---

### Task 6: 集成到 cmd/server.go

**Files:**
- Modify: `backend/cmd/server.go`

**Step 1: 在 server.go 中添加 Redis 和 Kafka 的初始化与关停**

将 `cmd/server.go` 的 `Run` 函数修改为：

```go
Run: func(cmd *cobra.Command, args []string) {
    defer logger.Sync()

    // 初始化 Redis
    if err := infraredis.Init(conf.C.Redis); err != nil {
        logger.Fatalf("redis init: %s", err)
    }
    defer infraredis.Close()

    // 初始化 Kafka
    if err := infrakafka.Init(conf.C.Kafka); err != nil {
        logger.Fatalf("kafka init: %s", err)
    }
    defer infrakafka.Close()

    // 启动 Kafka 消费者
    consumeCtx, consumeCancel := context.WithCancel(context.Background())
    defer consumeCancel()
    if err := infrakafka.StartConsumers(consumeCtx); err != nil {
        logger.Fatalf("kafka start consumers: %s", err)
    }

    srv := &http.Server{
        Addr:    fmt.Sprintf(":%d", conf.C.Server.Port),
        Handler: apphttp.NewServer(),
    }

    go func() {
        logger.Infof("server starting on %s", srv.Addr)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logger.Fatalf("listen: %s", err)
        }
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    logger.Infof("shutting down server...")
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        logger.Fatalf("server forced to shutdown: %s", err)
    }

    // 优雅停止消费者（defer 中的 Close 会处理 Producer）
    infrakafka.StopConsumers()

    logger.Infof("server exited")
},
```

import 中添加：

```go
infrakafka "q-dev/infra/kafka"
infraredis "q-dev/infra/redis"
```

**Step 2: 验证编译通过**

Run: `cd backend && go build ./...`
Expected: 编译成功

**Step 3: Commit**

```bash
git add cmd/server.go
git commit -m "feat(cmd): integrate redis and kafka initialization into server startup"
```

---

### Task 7: 更新 knowledge 文档

**Files:**
- Modify: `backend/knowledge/capability.md`
- Modify: `backend/knowledge/abstraction.md`

**Step 1: 在 capability.md 中添加 Redis 和 Kafka 能力描述**

添加内容：

```markdown
## Redis

- 通用缓存：通过 `infraredis.Client` 直接使用 go-redis 原生 API
- 分布式锁：通过 `infraredis.RS.NewMutex("lock-key")` 创建分布式互斥锁

## Kafka

- 生产者：`infrakafka.Produce(topic, key, value)` 异步发送消息
- 消费者注册：`infrakafka.Register(topic, groupID, handler, opts...)` 注册消费函数
- 同步/异步消费：默认同步，`WithAsync()` 启用异步消费
- 失败重试 + 死信队列：重试 N 次失败后发送到 `<topic>.DLQ`
```

**Step 2: 在 abstraction.md 中添加 infra 初始化抽象说明**

添加内容：

```markdown
## Infra 初始化

各基础设施组件独立 Init/Close，在 cmd/server.go 中按顺序调用：
- 启动：`redis.Init() → kafka.Init() → kafka.StartConsumers()`
- 关停：`kafka.StopConsumers() → kafka.Close() → redis.Close()`
```

**Step 3: Commit**

```bash
git add knowledge/capability.md knowledge/abstraction.md
git commit -m "docs(knowledge): add redis and kafka capability and abstraction docs"
```
