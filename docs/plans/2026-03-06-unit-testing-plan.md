# 单元测试规范 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 为 Q-DEV 搭建单元测试基础设施，建立各层测试规范，以 HelloWorld 模块为示例。

**Architecture:** 使用 testify 断言 + go-sqlmock mock MySQL + miniredis mock Redis。测试辅助工具放在 `pkg/testutil/`，测试文件与被测代码同目录。不改动现有架构，通过替换全局变量实现 mock。

**Tech Stack:** testify, go-sqlmock, miniredis/v2, httptest

**Design Doc:** `docs/plans/2026-03-06-unit-testing-design.md`

---

### Task 1: 安装测试依赖

**Step 1: 安装 testify、go-sqlmock、miniredis**

```bash
cd backend
go get github.com/stretchr/testify
go get github.com/DATA-DOG/go-sqlmock
go get github.com/alicebob/miniredis/v2
```

**Step 2: 验证**

Run: `cd backend && go mod tidy && go build ./...`
Expected: 编译成功

**Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "feat(deps): add testify, go-sqlmock, miniredis for unit testing"
```

---

### Task 2: 创建 MySQL 测试辅助工具

**Files:**
- Create: `backend/pkg/testutil/mysql.go`

**Step 1: 创建 testutil/mysql.go**

```go
package testutil

import (
	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// NewMockDB 创建一个 mock 的 GORM DB 和 sqlmock 实例。
// 调用者通过 sqlmock 设置期望，通过 *gorm.DB 执行查询。
func NewMockDB() (*gorm.DB, sqlmock.Sqlmock, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, err
	}

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, nil, err
	}

	return gormDB, mock, nil
}
```

**Step 2: 验证编译**

Run: `cd backend && go build ./pkg/testutil/...`
Expected: 编译成功

**Step 3: Commit**

```bash
git add pkg/testutil/mysql.go
git commit -m "feat(testutil): add NewMockDB helper for GORM + sqlmock"
```

---

### Task 3: 创建 Redis 测试辅助工具

**Files:**
- Create: `backend/pkg/testutil/redis.go`

**Step 1: 创建 testutil/redis.go**

```go
package testutil

import (
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// NewMockRedis 启动一个内存 Redis 并返回 go-redis Client。
// 测试结束后调用 miniredis.Close() 释放资源。
func NewMockRedis() (*redis.Client, *miniredis.Miniredis, error) {
	mr, err := miniredis.Run()
	if err != nil {
		return nil, nil, err
	}

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return client, mr, nil
}
```

**Step 2: 验证编译**

Run: `cd backend && go build ./pkg/testutil/...`
Expected: 编译成功

**Step 3: Commit**

```bash
git add pkg/testutil/redis.go
git commit -m "feat(testutil): add NewMockRedis helper for miniredis"
```

---

### Task 4: 添加 Makefile test/cover 目标

**Files:**
- Modify: `Makefile:25-27`

**Step 1: 在 Makefile 的 `lint` 目标后添加测试目标**

在 `lint` 目标后（`# ---------- Docker ----------` 行之前）添加：

```makefile
# ---------- 测试 ----------

test:
	cd $(BUILD_DIR) && go test ./... -v -count=1

cover:
	cd $(BUILD_DIR) && go test ./... -coverprofile=coverage.out -count=1
	cd $(BUILD_DIR) && go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: $(BUILD_DIR)/coverage.html"
```

同时在 `.PHONY` 行追加 `test cover`。

**Step 2: 验证**

Run: `make test`
Expected: 输出 `ok` 或 `no test files`（目前还没有测试文件，不应报错）

**Step 3: Commit**

```bash
git add Makefile
git commit -m "feat(makefile): add test and cover targets"
```

---

### Task 5: 编写 common/response 测试

**Files:**
- Create: `backend/http/common/response_test.go`

**Step 1: 创建 response_test.go**

```go
package common

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestOK_WithData(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	OK(c, gin.H{"key": "value"})

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "ok", resp.Message)
	assert.NotNil(t, resp.Data)
}

func TestOK_NilData(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	OK(c, nil)

	assert.Equal(t, http.StatusOK, w.Code)

	var raw map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &raw)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), raw["code"])
	_, hasData := raw["data"]
	assert.False(t, hasData) // omitempty: data 字段不应出现
}

func TestFail(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Fail(c, errors.New("something went wrong"))

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, -1, resp.Code)
	assert.Equal(t, "something went wrong", resp.Message)
}
```

**Step 2: 运行测试**

Run: `cd backend && go test ./http/common/... -v`
Expected: 3 个测试全部 PASS

**Step 3: Commit**

```bash
git add http/common/response_test.go
git commit -m "test(common): add unit tests for OK and Fail response helpers"
```

---

### Task 6: 编写 HTTP handler 测试（HelloWorld）

**Files:**
- Create: `backend/http/api/hello_world_test.go`

**说明：** HelloWorldAPI 依赖 `app.AppService`，而 AppService 内部依赖 domain.Service → dao 全局变量。测试时通过 `testutil.NewMockDB()` 创建 mock DB，再调用 `dao.SetDefault()` 替换全局 DAO，使 handler 整条链路走 mock。

**Step 1: 创建 hello_world_test.go**

```go
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"q-dev/http/common"
	"q-dev/infra/mysql/dao"
	"q-dev/pkg/testutil"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupMockDB(t *testing.T) sqlmock.Sqlmock {
	t.Helper()
	gormDB, mock, err := testutil.NewMockDB()
	require.NoError(t, err)
	dao.SetDefault(gormDB)
	return mock
}

func TestHelloWorldAPI_List_Success(t *testing.T) {
	mock := setupMockDB(t)

	// mock count query
	mock.ExpectQuery("SELECT count").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// mock select query
	mock.ExpectQuery("SELECT \\*").
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "created_at", "updated_at"}).
			AddRow(1, "test", "desc", "2026-01-01 00:00:00", "2026-01-01 00:00:00"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/hello-world?page=1&page_size=10", nil)

	h := NewHelloWorldAPI()
	h.List(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp common.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHelloWorldAPI_List_MissingParams(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/hello-world", nil) // 缺少 page, page_size

	h := NewHelloWorldAPI()
	h.List(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp common.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, -1, resp.Code) // 参数校验失败
}

func TestHelloWorldAPI_Get_Success(t *testing.T) {
	mock := setupMockDB(t)

	mock.ExpectQuery("SELECT \\*").
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "created_at", "updated_at"}).
			AddRow(1, "test", "desc", "2026-01-01 00:00:00", "2026-01-01 00:00:00"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/hello-world/1", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	h := NewHelloWorldAPI()
	h.Get(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp common.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHelloWorldAPI_Get_InvalidID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/hello-world/abc", nil)
	c.Params = gin.Params{{Key: "id", Value: "abc"}}

	h := NewHelloWorldAPI()
	h.Get(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp common.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, -1, resp.Code) // 解析 ID 失败
}
```

**Step 2: 运行测试**

Run: `cd backend && go test ./http/api/... -v`
Expected: 4 个测试全部 PASS

**Step 3: Commit**

```bash
git add http/api/hello_world_test.go
git commit -m "test(api): add unit tests for HelloWorld API handlers"
```

---

### Task 7: 运行全量测试 & 验证

**Step 1: 运行全量测试**

Run: `make test`
Expected: 所有测试 PASS，无错误

**Step 2: 生成覆盖率报告**

Run: `make cover`
Expected: 生成 `backend/coverage.html`

**Step 3: 检查覆盖率**

Run: `cd backend && go test ./http/common/... -cover && go test ./http/api/... -cover`
Expected: 显示 common 和 api 包的覆盖率百分比

**Step 4: 将 coverage 输出文件加入 .gitignore**

在 `backend/.gitignore`（如果不存在则创建）中添加：

```
coverage.out
coverage.html
```

**Step 5: Commit**

```bash
git add .gitignore
git commit -m "chore: add coverage output files to gitignore"
```
