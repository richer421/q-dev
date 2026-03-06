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
