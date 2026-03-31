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
	"crypto/tls"
	"fmt"
	"net/http"
	"time"
)

const (
	defaultHost             = "localhost"
	defaultPort             = 6333
	defaultKeepAliveTime    = 90
	defaultKeepAliveTimeout = 30
)

// Config holds the configuration for a Qdrant HTTP client.
type Config struct {
	Host             string
	Port             int
	APIKey           string
	UseTLS           bool
	TLSConfig        *tls.Config
	KeepAliveTime    int
	KeepAliveTimeout uint
}

func (c *Config) getBaseURL() string {
	host := c.Host
	if host == "" {
		host = defaultHost
	}

	port := c.Port
	if port == 0 {
		port = defaultPort
	}

	scheme := "http"
	if c.UseTLS {
		scheme = "https"
	}

	return fmt.Sprintf("%s://%s:%d", scheme, host, port)
}

func (c *Config) getHTTPClient() *http.Client {
	keepAliveTime := defaultKeepAliveTime
	if c.KeepAliveTime > 0 {
		keepAliveTime = c.KeepAliveTime
	}

	keepAliveTimeout := defaultKeepAliveTimeout
	if c.KeepAliveTimeout > 0 {
		keepAliveTimeout = int(c.KeepAliveTimeout)
	}

	transport := &http.Transport{
		IdleConnTimeout:     time.Duration(keepAliveTime) * time.Second,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
	}

	if c.UseTLS {
		tlsCfg := c.TLSConfig
		if tlsCfg == nil {
			tlsCfg = &tls.Config{}
		}
		transport.TLSClientConfig = tlsCfg
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(keepAliveTimeout) * time.Second,
	}

	return httpClient
}
