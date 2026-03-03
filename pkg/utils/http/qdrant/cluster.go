package qdrant

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"kubedb.dev/apimachinery/pkg/utils/http/qdrant/types"
)

// GetClusterInfo retrieves information about the cluster
func (c *Client) GetClusterInfo(ctx context.Context) (*types.GetClusterInfoResponse, error) {
	path := "/cluster"

	req, err := c.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var response types.GetClusterInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &response, nil
}

// GetCollectionClusterInfo retrieves cluster information for a specific collection
func (c *Client) GetCollectionClusterInfo(ctx context.Context, collectionName string) (*types.GetCollectionClusterInfoResponse, error) {
	path := fmt.Sprintf("/collections/%s/cluster", collectionName)

	req, err := c.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var response types.GetCollectionClusterInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &response, nil
}
