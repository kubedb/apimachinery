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

// Client is an HTTP client for the Qdrant API.
// Client is an HTTP client for the Qdrant API.
type Client struct {
	client  *http.Client
	baseURL string
	apiKey  string
}

// NewClient creates a new Qdrant HTTP client from the given config.
func NewClient(config *Config) (*Client, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
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

// Do executes an HTTP request, adding the API key header if configured.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	if c.apiKey != "" {
		req.Header.Set("api-key", c.apiKey)
	}

	return c.client.Do(req)
}

// NewRequest creates a new HTTP request bound to the client's base URL.
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

// Close closes idle connections in the underlying HTTP client.
func (c *Client) Close() {
	if tr, ok := c.client.Transport.(*http.Transport); ok {
		tr.CloseIdleConnections()
	}
}
