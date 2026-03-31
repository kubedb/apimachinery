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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Upsert upserts points into a collection.
func (c *Client) Upsert(ctx context.Context, collectionName string, req *UpsertPointsRequest) (*UpsertPointsResponse, error) {
	path := fmt.Sprintf("/collections/%s/points", collectionName)

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request body: %w", err)
	}

	httpReq, err := c.NewRequest(ctx, http.MethodPut, path, toReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var response UpsertPointsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &response, nil
}

// Get retrieves points from a collection by their IDs.
func (c *Client) Get(ctx context.Context, collectionName string, req *GetPointsRequest) (*GetPointsResponse, error) {
	path := fmt.Sprintf("/collections/%s/points", collectionName)

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request body: %w", err)
	}

	httpReq, err := c.NewRequest(ctx, http.MethodPost, path, toReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var response GetPointsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &response, nil
}
