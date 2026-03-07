# MCP Server Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Create standalone MCP Server for Claude to call APIs and read logs.

**Architecture:** New `mcp` subcommand starts stdio-based MCP server. Tools call app layer directly (no HTTP). Log file already supported by existing logger.

**Tech Stack:** Go 1.25 + github.com/modelcontextprotocol/go-sdk v1.4.0

---

## Task 1: Add MCP SDK Dependency

**Files:**
- Modify: `backend/go.mod`

**Step 1: Add dependency**

```bash
cd backend && go get github.com/modelcontextprotocol/go-sdk@v1.4.0
```

**Step 2: Verify**

Run: `cd backend && go mod tidy`
Expected: No errors

**Step 3: Commit**

```bash
git add backend/go.mod backend/go.sum
git commit -m "chore: add MCP SDK dependency"
```

---

## Task 2: Create MCP Server Package

**Files:**
- Create: `backend/pkg/mcp/server.go`

**Step 1: Create directory and file**

```bash
mkdir -p backend/pkg/mcp
```

**Step 2: Write MCP server implementation**

```go
package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"q-dev/app/hello_world"
	"q-dev/app/hello_world/vo"
	"q-dev/conf"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Server struct {
	helloWorld *hello_world.AppService
}

func NewServer() *Server {
	return &Server{
		helloWorld: hello_world.NewAppService(),
	}
}

func (s *Server) Run() error {
	// Create MCP server
	srv := mcp.NewServer("q-dev", "1.0.0", nil)

	// Register tools
	s.registerTools(srv)

	// Run stdio transport
	return srv.Run(context.Background())
}

func (s *Server) registerTools(srv *mcp.Server) {
	// Tool: call_api
	srv.AddTool(mcp.Tool{
		Name:        "call_api",
		Description: "Call Q-DEV API directly",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"action": map[string]any{
					"type":        "string",
					"description": "Action to call: hello_world.list, hello_world.get, hello_world.create, hello_world.update, hello_world.delete",
				},
				"params": map[string]any{
					"type":        "object",
					"description": "Parameters for the action",
				},
			},
			Required: []string{"action"},
		},
	}, s.handleCallAPI)

	// Tool: read_logs
	srv.AddTool(mcp.Tool{
		Name:        "read_logs",
		Description: "Read last N lines from log file",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"lines": map[string]any{
					"type":        "integer",
					"description": "Number of lines to read (default 100)",
				},
			},
		},
	}, s.handleReadLogs)
}

func (s *Server) handleCallAPI(ctx context.Context, args map[string]any) (*mcp.CallToolResult, error) {
	action, _ := args["action"].(string)
	params, _ := args["params"].(map[string]any)

	var result any
	var err error

	switch action {
	case "hello_world.list":
		req := &vo.ListReq{Page: 1, PageSize: 10}
		if p, ok := params["page"].(float64); ok {
			req.Page = int(p)
		}
		if p, ok := params["page_size"].(float64); ok {
			req.PageSize = int(p)
		}
		result, err = s.helloWorld.List(ctx, req)

	case "hello_world.get":
		id := uint(params["id"].(float64))
		result, err = s.helloWorld.Get(ctx, id)

	case "hello_world.create":
		req := &vo.CreateReq{
			Title:       params["title"].(string),
			Description: params["description"].(string),
		}
		result, err = s.helloWorld.Create(ctx, req)

	case "hello_world.update":
		id := uint(params["id"].(float64))
		req := &vo.UpdateReq{}
		if v, ok := params["title"].(string); ok {
			req.Title = &v
		}
		if v, ok := params["description"].(string); ok {
			req.Description = &v
		}
		err = s.helloWorld.Update(ctx, id, req)

	case "hello_world.delete":
		id := uint(params["id"].(float64))
		err = s.helloWorld.Delete(ctx, id)

	default:
		return mcp.NewToolResultError(fmt.Sprintf("unknown action: %s", action)), nil
	}

	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func (s *Server) handleReadLogs(ctx context.Context, args map[string]any) (*mcp.CallToolResult, error) {
	logPath := conf.C.Log.File.Path
	if logPath == "" {
		logPath = "logs/app.log"
	}

	lines := 100
	if l, ok := args["lines"].(float64); ok {
		lines = int(l)
	}

	content, err := readLastLines(logPath, lines)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to read log: %v", err)), nil
	}

	return mcp.NewToolResultText(content), nil
}

func readLastLines(path string, n int) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) > n {
			lines = lines[1:]
		}
	}

	return strings.Join(lines, "\n"), scanner.Err()
}
```

**Step 3: Verify compilation**

Run: `cd backend && go build ./pkg/mcp/...`
Expected: No errors

**Step 4: Commit**

```bash
git add backend/pkg/mcp/
git commit -m "feat(mcp): add MCP server package with call_api and read_logs tools"
```

---

## Task 3: Add MCP CLI Command

**Files:**
- Create: `backend/cmd/mcp.go`

**Step 1: Create mcp command**

```go
package cmd

import (
	"context"
	"fmt"
	"os"

	"q-dev/infra/mysql"
	"q-dev/pkg/logger"
	"q-dev/pkg/mcp"

	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start MCP server for Claude integration",
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize MySQL (required for app layer)
		if err := mysql.Init(); err != nil {
			fmt.Fprintf(os.Stderr, "mysql init failed: %v\n", err)
			os.Exit(1)
		}

		logger.Infof("Starting MCP server...")
		srv := mcp.NewServer()
		if err := srv.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}
```

**Step 2: Verify compilation**

Run: `cd backend && go build -o bin/q-dev .`
Expected: Binary created successfully

**Step 3: Commit**

```bash
git add backend/cmd/mcp.go
git commit -m "feat(cmd): add mcp subcommand"
```

---

## Task 4: Update Configuration Default

**Files:**
- Modify: `backend/conf/config.yaml`

**Step 1: Enable file logging by default**

Change line 43 from `enabled: false` to `enabled: true`:

```yaml
log:
  level: "info"
  format: "console"
  file:
    enabled: true
    path: "logs/app.log"
    max_size: 100
    max_age: 30
    compress: true
```

**Step 2: Commit**

```bash
git add backend/conf/config.yaml
git commit -m "chore: enable file logging by default for MCP read_logs"
```

---

## Task 5: Update Knowledge Docs

**Files:**
- Modify: `backend/knowledge/capability.md`

**Step 1: Add MCP capability**

Add to capability.md:

```markdown
### MCP Server

独立 MCP Server，供 Claude 调用：

- `q-dev mcp` — 启动 MCP Server（stdio 模式）
- 工具：
  - `call_api` — 调用 API（action: hello_world.list/get/create/update/delete）
  - `read_logs` — 读取日志文件
```

**Step 2: Commit**

```bash
git add backend/knowledge/capability.md
git commit -m "docs: add MCP Server capability to knowledge"
```

---

## Summary

| Task | Description |
|------|-------------|
| 1 | Add MCP SDK dependency |
| 2 | Create MCP server package |
| 3 | Add mcp CLI command |
| 4 | Enable file logging |
| 5 | Update knowledge docs |

**Usage:**
```bash
./bin/q-dev mcp  # Start MCP server
```
