package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"q-dev/http/common"
	"q-dev/pkg/logger"

	"moul.io/http2curl/v2"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	debug      bool
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: http.DefaultClient,
		debug:      true,
	}
}

func NewClientWithHTTP(baseURL string, httpClient *http.Client) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: httpClient,
		debug:      true,
	}
}

func (c *Client) SetDebug(on bool) {
	c.debug = on
}

// request 封装单次请求的参数
type request struct {
	method  string
	path    string
	query   url.Values
	headers http.Header
	body    any // 非 nil 时 JSON 序列化为 request body
}

func (c *Client) do(ctx context.Context, r *request) (*common.Response, error) {
	// 拼 URL + query
	u := c.baseURL + r.path
	if len(r.query) > 0 {
		u += "?" + r.query.Encode()
	}

	// body
	var bodyReader io.Reader
	if r.body != nil {
		data, err := json.Marshal(r.body)
		if err != nil {
			return nil, fmt.Errorf("sdk: marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, r.method, u, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("sdk: create request: %w", err)
	}

	// headers
	if r.body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, vs := range r.headers {
		for _, v := range vs {
			req.Header.Set(k, v)
		}
	}

	// debug curl
	if c.debug {
		if cmd, err := http2curl.GetCurlCommand(req); err == nil {
			logger.Debugf("sdk: %s", cmd)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sdk: do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("sdk: read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sdk: unexpected status %d: %s", resp.StatusCode, respBody)
	}

	var result common.Response
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("sdk: decode response: %w", err)
	}

	return &result, nil
}
