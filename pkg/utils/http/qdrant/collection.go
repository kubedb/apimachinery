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

// ListCollections returns a list of all existing collections.
func (c *Client) ListCollections(ctx context.Context) (*ListCollectionsResponse, error) {
	path := "/collections"

	req, err := c.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var response ListCollectionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &response, nil
}

// GetCollectionDetails retrieves parameters from the specified collection.
func (c *Client) GetCollectionDetails(ctx context.Context, collectionName string) (*GetCollectionDetailsResponse, error) {
	path := fmt.Sprintf("/collections/%s", collectionName)

	req, err := c.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var response GetCollectionDetailsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &response, nil
}

// CollectionExists checks if a collection exists
func (c *Client) CollectionExists(ctx context.Context, collectionName string) (bool, error) {
	path := fmt.Sprintf("/collections/%s", collectionName)

	req, err := c.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return false, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.Do(req)
	if err != nil {
		return false, fmt.Errorf("executing request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	}

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	return false, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(bodyBytes))
}

// CreateCollection creates a new collection.
func (c *Client) CreateCollection(ctx context.Context, collectionName string, req *CreateCollectionRequest) (*CreateCollectionResponse, error) {
	path := fmt.Sprintf("/collections/%s", collectionName)

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

	var response CreateCollectionResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &response, nil
}
