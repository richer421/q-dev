# 单元测试规范设计

## 概述

为 Q-DEV 脚手架设计单元测试规范，保持现有全局变量架构不变，通过替换全局变量实现 mock。以 HelloWorld 模块为示例，后续业务模块照此模式编写测试。

## 工具链

| 库 | 用途 |
|---|---|
| `github.com/stretchr/testify` | 断言（assert/require）+ test suite |
| `github.com/DATA-DOG/go-sqlmock` | Mock MySQL，拦截 database/sql 层，GORM 兼容 |
| `github.com/alicebob/miniredis/v2` | 内存 Redis，go-redis 原生兼容 |
| `net/http/httptest` | Go 标准库，测试 Gin handler |
| Kafka | 单元测试中不测，handler 逻辑提取为纯函数单独测试 |

## Makefile

```makefile
test:     go test ./... -v -count=1
cover:    go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out
```

---

## 各层测试规范

### HTTP 层（api handler）

- 用 `httptest.NewRecorder()` + `gin.CreateTestContext()` 构造请求
- 直接调用 handler 函数
- 测试前替换全局 infra 变量为 mock 实例
- 测试内容：参数校验、正常响应格式和状态码、错误场景

### MySQL 层（infra/mysql）

- 用 go-sqlmock 创建 mock `*sql.DB`，再用 `gorm.Open(mysql.New(mysql.Config{Conn: db}))` 包装
- 调用 `dao.SetDefault(mockGormDB)` 替换全局查询
- 测试内容：DAO 查询生成正确 SQL、结果映射、错误场景

### Redis 层（infra/redis）

- 用 miniredis 启动内存 Redis，`redis.NewClient()` 连接后替换全局 `Client`
- 测试内容：缓存读写、分布式锁加锁/释放

### App / Domain 层

- 替换全局 infra 变量后测试业务逻辑
- 纯业务逻辑（不依赖 infra）直接测试，无需 mock

### Kafka

- 单元测试中不测 Kafka（CGO 依赖）
- handler 函数逻辑提取为纯函数单独测试

---

## 测试目录结构

测试文件与被测代码同目录（Go 惯例）：

```
http/api/hello_world_test.go
http/common/response_test.go
infra/mysql/mysql_test.go
infra/redis/redis_test.go
```

### 测试辅助工具

```
pkg/testutil/
├── mysql.go    # NewMockDB() → (*gorm.DB, sqlmock.Sqlmock)
└── redis.go    # NewMockRedis() → (*redis.Client, *miniredis.Miniredis)
```

## 命名约定

- 测试函数：`Test<函数名>_<场景>`（如 `TestList_Success`、`TestList_NotFound`）
- 多场景优先用表格驱动测试
- 需要全局 mock 初始化时使用 TestMain，否则不强制

## 示例覆盖（HelloWorld 模块）

| 文件 | 测试内容 |
|------|---------|
| `http/api/hello_world_test.go` | List 和 Get handler 的请求/响应 |
| `http/common/response_test.go` | OK/Fail 函数的响应格式 |
