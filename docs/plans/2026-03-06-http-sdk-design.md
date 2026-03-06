# HTTP SDK 设计文档

## 背景

微服务之间需要快速调用本服务暴露的 HTTP API。在 http 层新增 sdk 目录，为每个模块提供类型安全的 Go 客户端，内部服务 import 后即可调用。

## 目录结构

```
http/sdk/
├── client.go              # SDK 客户端（baseURL + http.Client 封装）
└── hello_world.go         # hello_world 模块的 SDK 方法
```

每个业务模块一个文件，与 `http/api/` 一一对应。

## 核心 API

```go
// client.go
type Client struct {
    baseURL    string
    httpClient *http.Client
}

func NewClient(baseURL string) *Client

// hello_world.go
func (c *Client) HelloWorldList(ctx context.Context) (*common.Response, error)
func (c *Client) HelloWorldGet(ctx context.Context, id string) (*common.Response, error)
```

## 使用方式

```go
cli := sdk.NewClient("http://q-dev-service:8080")
resp, err := cli.HelloWorldList(ctx)
```

## 设计要点

- 复用 `common.Response` 做返回值解析，与服务端统一
- 每个方法对应一个 HTTP 接口，URL/Method/参数与 router 注册一致
- ctx 透传，支持超时和取消
- 错误处理：网络错误返回 error，业务错误通过 Response.Code 判断
- Client 内部封装通用的 do() 方法处理请求发送和响应解析

## 不做的事

- 不做重试、熔断、服务发现（后续按需加）
- 不做多语言 SDK
- 仅供内部 Go 微服务使用

## 依赖

- 仅标准库 net/http + encoding/json，零外部依赖
