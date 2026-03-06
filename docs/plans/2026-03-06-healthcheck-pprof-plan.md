# 健康检查 & pprof 实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 为 Q-DEV 添加存活探针（/healthz）、就绪探针（/readyz）和 pprof 性能分析端点（/debug/pprof）。

**Architecture:** 在 `http/router/router.go` 中注册根路径端点，readyz 调用各 infra 包的连通性检查，pprof 使用 gin-contrib/pprof 库一行挂载。

**Tech Stack:** gin-contrib/pprof

**Design Doc:** `docs/plans/2026-03-06-healthcheck-pprof-design.md`

---

### Task 1: 安装 gin-contrib/pprof 依赖

**Step 1: 安装依赖**

```bash
cd backend
go get github.com/gin-contrib/pprof
```

**Step 2: 验证编译通过**

Run: `cd backend && go mod tidy && go build ./...`
Expected: 编译成功

**Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "feat(deps): add gin-contrib/pprof dependency"
```

---

### Task 2: 注册 healthz、readyz 和 pprof 路由

**Files:**
- Modify: `backend/http/router/router.go:1-11`

**Step 1: 修改 router.go，添加健康检查和 pprof**

将 `backend/http/router/router.go` 替换为：

```go
package router

import (
	"context"
	"net/http"
	"time"

	inframysql "q-dev/infra/mysql"
	infraredis "q-dev/infra/redis"
	infrakafka "q-dev/infra/kafka"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
)

func Register(r *gin.Engine) {
	// 存活探针
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 就绪探针
	r.GET("/readyz", readyz)

	// pprof
	pprof.Register(r)

	// 业务路由
	api := r.Group("/api")
	RegisterV1(api)
}

func readyz(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	checks := make(map[string]string)
	healthy := true

	// MySQL
	if inframysql.DB != nil {
		sqlDB, err := inframysql.DB.DB()
		if err != nil {
			checks["mysql"] = err.Error()
			healthy = false
		} else if err := sqlDB.PingContext(ctx); err != nil {
			checks["mysql"] = err.Error()
			healthy = false
		} else {
			checks["mysql"] = "ok"
		}
	}

	// Redis
	if infraredis.Client != nil {
		if err := infraredis.Client.Ping(ctx).Err(); err != nil {
			checks["redis"] = err.Error()
			healthy = false
		} else {
			checks["redis"] = "ok"
		}
	}

	// Kafka
	if infrakafka.Producer != nil {
		_, err := infrakafka.Producer.GetMetadata(nil, true, 3000)
		if err != nil {
			checks["kafka"] = err.Error()
			healthy = false
		} else {
			checks["kafka"] = "ok"
		}
	}

	status := http.StatusOK
	if !healthy {
		status = http.StatusServiceUnavailable
	}

	c.JSON(status, gin.H{
		"status": map[bool]string{true: "ok", false: "unavailable"}[healthy],
		"checks": checks,
	})
}
```

**Step 2: 验证编译通过**

Run: `cd backend && go build ./...`
Expected: 编译成功

**Step 3: Commit**

```bash
git add http/router/router.go
git commit -m "feat(http): add healthz, readyz probes and pprof endpoints"
```

---

### Task 3: 更新 knowledge 文档

**Files:**
- Modify: `backend/knowledge/capability.md`

**Step 1: 在 capability.md 中添加运维端点说明**

追加以下内容：

```markdown
## 运维端点

- `GET /healthz` — 存活探针，直接返回 200
- `GET /readyz` — 就绪探针，检查 MySQL/Redis/Kafka 连通性，全通 200，任一失败 503
- `/debug/pprof/*` — Go pprof 性能分析端点
```

**Step 2: Commit**

```bash
git add knowledge/capability.md
git commit -m "docs(knowledge): add healthz, readyz, pprof capability"
```
