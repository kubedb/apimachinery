package qdrant

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"kubedb.dev/apimachinery/pkg/utils/http/qdrant/types"
)

// CreateFullSnapshot creates a new full storage snapshot
func (c *Client) CreateFullSnapshot(ctx context.Context) (*types.CreateSnapshotResponse, error) {
	path := "/snapshots"

	req, err := c.NewRequest(ctx, http.MethodPost, path, nil)
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

	var response types.CreateSnapshotResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &response, nil
}

// ListFullSnapshots lists all full storage snapshots
func (c *Client) ListFullSnapshots(ctx context.Context) (*types.ListSnapshotsResponse, error) {
	path := "/snapshots"

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

	var response types.ListSnapshotsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &response, nil
}

// DeleteFullSnapshot deletes a specific full storage snapshot
func (c *Client) DeleteFullSnapshot(ctx context.Context, snapshotName string) (*types.DeleteSnapshotResponse, error) {
	path := fmt.Sprintf("/snapshots/%s", snapshotName)

	req, err := c.NewRequest(ctx, http.MethodDelete, path, nil)
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

	var response types.DeleteSnapshotResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &response, nil
}

// RecoverFullSnapshot Recovers a full storage snapshot from a specified location
func (c *Client) RecoverFullSnapshot(ctx context.Context, location string) (*types.RecoverSnapshotResponse, error) {
	path := "/snapshots/upload"

	requestBody := map[string]string{
		"location": location,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("marshaling request body: %w", err)
	}

	req, err := c.NewRequest(ctx, http.MethodPut, path, toReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var response types.RecoverSnapshotResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &response, nil
}

// CreateCollectionSnapshot creates a new snapshot for a collection
func (c *Client) CreateCollectionSnapshot(ctx context.Context, collectionName string) (*types.CreateSnapshotResponse, error) {
	path := fmt.Sprintf("/collections/%s/snapshots", collectionName)

	req, err := c.NewRequest(ctx, http.MethodPost, path, nil)
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

	var response types.CreateSnapshotResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &response, nil
}

// ListCollectionSnapshots lists all snapshots for a collection
func (c *Client) ListCollectionSnapshots(ctx context.Context, collectionName string) (*types.ListSnapshotsResponse, error) {
	path := fmt.Sprintf("/collections/%s/snapshots", collectionName)

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

	var response types.ListSnapshotsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &response, nil
}

// DeleteCollectionSnapshot deletes a specific snapshot for a collection
func (c *Client) DeleteCollectionSnapshot(ctx context.Context, collectionName string, snapshotName string) (*types.DeleteSnapshotResponse, error) {
	path := fmt.Sprintf("/collections/%s/snapshots/%s", collectionName, snapshotName)

	req, err := c.NewRequest(ctx, http.MethodDelete, path, nil)
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

	var response types.DeleteSnapshotResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &response, nil
}

// RecoverCollectionSnapshot Recovers a collection snapshot from a specified location
func (c *Client) RecoverCollectionSnapshot(ctx context.Context, collectionName string, location string) (*types.RecoverSnapshotResponse, error) {
	path := fmt.Sprintf("/collections/%s/snapshots/upload", collectionName)

	requestBody := map[string]string{
		"location": location,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("marshaling request body: %w", err)
	}

	req, err := c.NewRequest(ctx, http.MethodPut, path, toReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var response types.RecoverSnapshotResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &response, nil
}

// DownloadCollectionSnapshot downloads a collection snapshot file
// Returns the response body containing the snapshot file data
// The caller is responsible for closing the returned reader
func (c *Client) DownloadCollectionSnapshot(ctx context.Context, collectionName string, snapshotName string) (io.ReadCloser, error) {
	path := fmt.Sprintf("/collections/%s/snapshots/%s", collectionName, snapshotName)

	req, err := c.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return resp.Body, nil
}

// CreateShardSnapshot creates a new snapshot for a specific shard
func (c *Client) CreateShardSnapshot(ctx context.Context, collectionName string, shardID string) (*types.CreateSnapshotResponse, error) {
	path := fmt.Sprintf("/collections/%s/shards/%s/snapshots", collectionName, shardID)

	req, err := c.NewRequest(ctx, http.MethodPost, path, nil)
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

	var response types.CreateSnapshotResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &response, nil
}

// ListShardSnapshots lists all snapshots for a specific shard
func (c *Client) ListShardSnapshots(ctx context.Context, collectionName string, shardID string) (*types.ListSnapshotsResponse, error) {
	path := fmt.Sprintf("/collections/%s/shards/%s/snapshots", collectionName, shardID)

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

	var response types.ListSnapshotsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &response, nil
}

// DeleteShardSnapshot deletes a specific snapshot for a shard
func (c *Client) DeleteShardSnapshot(ctx context.Context, collectionName string, shardID string, snapshotName string) (*types.DeleteSnapshotResponse, error) {
	path := fmt.Sprintf("/collections/%s/shards/%s/snapshots/%s", collectionName, shardID, snapshotName)

	req, err := c.NewRequest(ctx, http.MethodDelete, path, nil)
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

	var response types.DeleteSnapshotResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &response, nil
}

// RecoverShardSnapshot Recovers a shard snapshot from a specified location
// The location can be a URL or a file path depending on the Qdrant configuration
func (c *Client) RecoverShardSnapshot(ctx context.Context, collectionName string, shardID string, snapshotName string, location string) (*types.RecoverSnapshotResponse, error) {
	path := fmt.Sprintf("/collections/%s/shards/%s/snapshots/upload", collectionName, shardID)

	requestBody := map[string]string{
		"location": location,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("marshaling request body: %w", err)
	}

	req, err := c.NewRequest(ctx, http.MethodPut, path, toReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var response types.RecoverSnapshotResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &response, nil
}

// DownloadShardSnapshot downloads a shard snapshot file
// Returns the response body containing the snapshot file data
// The caller is responsible for closing the returned reader
func (c *Client) DownloadShardSnapshot(ctx context.Context, collectionName string, shardID int, snapshotName string) (io.ReadCloser, error) {
	path := fmt.Sprintf("/collections/%s/shards/%d/snapshots/%s", collectionName, shardID, snapshotName)

	req, err := c.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return resp.Body, nil
}

// toReader converts a byte slice to an io.Reader
func toReader(b []byte) io.Reader {
	return &byteReader{data: b}
}

// byteReader implements io.Reader for a byte slice
type byteReader struct {
	data []byte
	pos  int
}

func (r *byteReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}
