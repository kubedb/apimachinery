/*
Copyright AppsCode Inc. and Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package qdrant

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	client  *http.Client
	baseURL string
	apiKey  string
}

func NewClient(config *Config) (*Client, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if config.Host == "" {
		return nil, fmt.Errorf("host is required")
	}

	if config.Port == 0 {
		config.Port = 6333
	}

	if config.KeepAliveTimeout == 0 {
		config.KeepAliveTimeout = 30
	}

	return &Client{
		client:  config.getHTTPClient(),
		baseURL: config.getBaseURL(),
		apiKey:  config.APIKey,
	}, nil
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	if c.apiKey != "" {
		req.Header.Set("api-key", c.apiKey)
	}

	return c.client.Do(req)
}

func (c *Client) NewRequest(
	ctx context.Context,
	method string,
	path string,
	body io.Reader,
) (*http.Request, error) {
	url := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c *Client) Close() {
	if tr, ok := c.client.Transport.(*http.Transport); ok {
		tr.CloseIdleConnections()
	}
}
