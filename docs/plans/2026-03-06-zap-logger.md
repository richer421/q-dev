# Zap Logger Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a zap-based structured logger with file rotation to the Q-DEV backend scaffold.

**Architecture:** A `pkg/logger` package wraps zap's SugaredLogger behind global format functions. stdout is always on; file output via lumberjack is config-toggled. Gin's default Logger and Recovery middleware are replaced with custom ones that route through the same logger.

**Tech Stack:** go.uber.org/zap, gopkg.in/natefinch/lumberjack.v2

**Design doc:** `docs/plans/2026-03-06-zap-logger-design.md`

---

### Task 1: Add dependencies

**Step 1: Install zap and lumberjack**

Run from `backend/`:

```bash
go get go.uber.org/zap
go get gopkg.in/natefinch/lumberjack.v2
```

**Step 2: Verify go.mod**

Run: `grep -E "zap|lumberjack" go.mod`
Expected: both dependencies listed.

**Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "deps: add zap and lumberjack"
```

---

### Task 2: Add log config structs and YAML

**Files:**
- Modify: `backend/conf/conf.go:12-17` (Config struct) and append new types
- Modify: `backend/conf/config.yaml` (add log section)

**Step 1: Add LogConfig and LogFileConfig to conf.go**

In `conf.go`, add `Log` field to `Config` struct and the two new types:

```go
type Config struct {
	Server ServerConfig `yaml:"server"`
	MySQL  MySQLConfig  `yaml:"mysql"`
	Redis  RedisConfig  `yaml:"redis"`
	Kafka  KafkaConfig  `yaml:"kafka"`
	Log    LogConfig    `yaml:"log"`
}
```

Append after `KafkaConfig`:

```go
type LogConfig struct {
	Level  string        `yaml:"level"`
	Format string        `yaml:"format"`
	File   LogFileConfig `yaml:"file"`
}

type LogFileConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Path     string `yaml:"path"`
	MaxSize  int    `yaml:"max_size"`
	MaxAge   int    `yaml:"max_age"`
	Compress bool   `yaml:"compress"`
}
```

**Step 2: Add log section to config.yaml**

Append to `backend/conf/config.yaml`:

```yaml
log:
  level: "info"
  format: "console"
  file:
    enabled: false
    path: "logs/app.log"
    max_size: 100
    max_age: 30
    compress: true
```

**Step 3: Verify it compiles**

Run from `backend/`: `go build ./...`
Expected: no errors.

**Step 4: Commit**

```bash
git add conf/conf.go conf/config.yaml
git commit -m "feat(conf): add log config structs and yaml defaults"
```

---

### Task 3: Implement pkg/logger

**Files:**
- Create: `backend/pkg/logger/logger.go`

**Step 1: Create the logger package**

Create `backend/pkg/logger/logger.go` with the following content:

```go
package logger

import (
	"os"

	"q-dev/conf"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var sugar *zap.SugaredLogger

func Init(cfg conf.LogConfig) {
	// 1. Parse level
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}

	// 2. Build encoder
	var encoder zapcore.Encoder
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "ts"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder

	switch cfg.Format {
	case "json":
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	default:
		encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	}

	// 3. Build write syncers — stdout is always on
	cores := []zapcore.Core{
		zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level),
	}

	// 4. Optionally add file syncer with lumberjack rotation
	if cfg.File.Enabled {
		fileEncoder := zapcore.NewJSONEncoder(encoderCfg) // file always JSON
		fileSyncer := zapcore.AddSync(&lumberjack.Logger{
			Filename: cfg.File.Path,
			MaxSize:  cfg.File.MaxSize,
			MaxAge:   cfg.File.MaxAge,
			Compress: cfg.File.Compress,
		})
		cores = append(cores, zapcore.NewCore(fileEncoder, fileSyncer, level))
	}

	// 5. Build logger
	core := zapcore.NewTee(cores...)
	l := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	sugar = l.Sugar()
}

func Sync() {
	if sugar != nil {
		_ = sugar.Sync()
	}
}

func Debugf(template string, args ...interface{}) { sugar.Debugf(template, args...) }
func Infof(template string, args ...interface{})  { sugar.Infof(template, args...) }
func Warnf(template string, args ...interface{})  { sugar.Warnf(template, args...) }
func Errorf(template string, args ...interface{}) { sugar.Errorf(template, args...) }
func Fatalf(template string, args ...interface{}) { sugar.Fatalf(template, args...) }
```

**Step 2: Verify it compiles**

Run from `backend/`: `go build ./pkg/logger/...`
Expected: no errors.

**Step 3: Commit**

```bash
git add pkg/logger/logger.go
git commit -m "feat(logger): implement zap logger with lumberjack file rotation"
```

---

### Task 4: Implement Gin middleware

**Files:**
- Create: `backend/http/middleware/logger.go`
- Create: `backend/http/middleware/recovery.go`

**Step 1: Create request logger middleware**

Create `backend/http/middleware/logger.go`:

```go
package middleware

import (
	"time"

	"q-dev/pkg/logger"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		if raw != "" {
			path = path + "?" + raw
		}

		logger.Infof("%s %s %d %s %s",
			c.Request.Method,
			path,
			c.Writer.Status(),
			time.Since(start),
			c.ClientIP(),
		)
	}
}
```

**Step 2: Create recovery middleware**

Create `backend/http/middleware/recovery.go`:

```go
package middleware

import (
	"net/http"
	"runtime/debug"

	"q-dev/pkg/logger"

	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Errorf("panic recovered: %v\n%s", err, debug.Stack())
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}
```

**Step 3: Verify it compiles**

Run from `backend/`: `go build ./http/middleware/...`
Expected: no errors.

**Step 4: Commit**

```bash
git add http/middleware/logger.go http/middleware/recovery.go
git commit -m "feat(middleware): add zap-based logger and recovery middleware"
```

---

### Task 5: Wire everything together

**Files:**
- Modify: `backend/http/server.go:1-16` (replace gin defaults)
- Modify: `backend/cmd/root.go:16-18` (init logger after config load)
- Modify: `backend/cmd/server.go:1-52` (replace stdlib log)

**Step 1: Update http/server.go**

Replace the full file content with:

```go
package http

import (
	"q-dev/http/middleware"
	"q-dev/http/router"

	"github.com/gin-gonic/gin"
)

func NewServer() *gin.Engine {
	r := gin.New()
	r.Use(middleware.Logger(), middleware.Recovery())

	router.Register(r)

	return r
}
```

**Step 2: Update cmd/root.go**

Replace the full file content with:

```go
package cmd

import (
	"log"

	"q-dev/conf"
	"q-dev/pkg/logger"

	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "q-dev",
	Short: "Q-Dev 后端服务",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := conf.Load(cfgFile); err != nil {
			return err
		}
		logger.Init(conf.C.Log)
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "conf/config.yaml", "配置文件路径")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
```

Note: `Execute()` keeps stdlib `log.Fatal` because logger is not yet initialized at that point.

**Step 3: Update cmd/server.go**

Replace the full file content with:

```go
package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"q-dev/conf"
	apphttp "q-dev/http"
	"q-dev/pkg/logger"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(serverCmd)
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "启动 HTTP 服务",
	Run: func(cmd *cobra.Command, args []string) {
		defer logger.Sync()

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
		logger.Infof("server exited")
	},
}
```

**Step 4: Verify it compiles**

Run from `backend/`: `go build ./...`
Expected: no errors.

**Step 5: Commit**

```bash
git add http/server.go cmd/root.go cmd/server.go
git commit -m "feat: wire zap logger into server startup and gin middleware"
```

---

### Task 6: Smoke test

**Step 1: Run the server**

Run from `backend/`:

```bash
go run main.go server
```

Expected: server starts with zap-formatted log output (not stdlib log format). You should see a line like:

```
2026-03-06T... INFO server starting on :8080
```

**Step 2: Send a test request**

```bash
curl http://localhost:8080/api/v1/hello-world
```

Expected: request log line appears in console via zap, like:

```
2026-03-06T... INFO GET /api/v1/hello-world 200 1.234ms 127.0.0.1
```

**Step 3: Stop the server with Ctrl+C**

Expected: shutdown logs appear via zap:

```
2026-03-06T... INFO shutting down server...
2026-03-06T... INFO server exited
```
