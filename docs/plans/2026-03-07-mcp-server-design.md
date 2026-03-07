# MCP Server 设计

## 目标

创建独立的 MCP Server，让 Claude 能自行调用后端接口、查看日志，辅助排查和修复问题。

## 范围

- 独立 CLI 命令 `q-dev mcp`，通过 stdio 通信
- 两个工具：`call_api`（调用 app 层方法）、`read_logs`（查看日志）
- 完全独立，不依赖 HTTP 服务运行

## 目录结构

```
backend/
├── cmd/
│   └── mcp.go              # 新增：mcp 子命令
├── pkg/
│   └── mcp/
│       └── server.go       # MCP Server 实现
└── conf/
    └── config.yaml         # 新增 log.file 配置项
```

## 配置

```yaml
log:
  file: logs/app.log   # 默认值，可配置
```

## MCP 工具

| 工具 | 参数 | 说明 |
|------|------|------|
| `call_api` | `action`, `params?` | 直接调用 app 层方法 |
| `read_logs` | `lines?` (默认 100) | 读取日志文件最后 N 行 |

### call_api 支持的 action

基于现有 hello_world 模块：
- `hello_world.list` — 列表查询
- `hello_world.get` — 根据 ID 查询
- `hello_world.create` — 创建
- `hello_world.update` — 更新
- `hello_world.delete` — 删除

后续新增模块时扩展 action 列表。

## 日志改造

修改 `pkg/logger/`，支持同时输出到 stdout 和文件：
- stdout：保持现有行为
- file：追加写入，按需配置

## 依赖

```go
require github.com/modelcontextprotocol/go-sdk latest
```

## 使用方式

```bash
# 启动 MCP Server
q-dev mcp

# Claude Code 配置
# .claude/settings.json
{
  "mcpServers": {
    "q-dev": {
      "command": "/path/to/q-dev",
      "args": ["mcp"]
    }
  }
}
```
