package sdk

import (
	"context"
	"fmt"
	"net/http"

	"q-dev/http/common"
)

func (c *Client) HelloWorldList(ctx context.Context) (*common.Response, error) {
	return c.do(ctx, &request{
		method: http.MethodGet,
		path:   "/api/v1/hello-world",
	})
}

func (c *Client) HelloWorldGet(ctx context.Context, id string) (*common.Response, error) {
	return c.do(ctx, &request{
		method: http.MethodGet,
		path:   fmt.Sprintf("/api/v1/hello-world/%s", id),
	})
}

func (c *Client) HelloWorldCreate(ctx context.Context, body any) (*common.Response, error) {
	return c.do(ctx, &request{
		method: http.MethodPost,
		path:   "/api/v1/hello-world",
		body:   body,
	})
}

func (c *Client) HelloWorldUpdate(ctx context.Context, id string, body any) (*common.Response, error) {
	return c.do(ctx, &request{
		method: http.MethodPut,
		path:   fmt.Sprintf("/api/v1/hello-world/%s", id),
		body:   body,
	})
}

func (c *Client) HelloWorldDelete(ctx context.Context, id string) (*common.Response, error) {
	return c.do(ctx, &request{
		method: http.MethodDelete,
		path:   fmt.Sprintf("/api/v1/hello-world/%s", id),
	})
}
