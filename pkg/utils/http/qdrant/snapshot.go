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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

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

func toReader(b []byte) io.Reader {
	return &byteReader{data: b}
}

// CreateFullSnapshot creates a snapshot of the entire Qdrant instance.
func (c *Client) CreateFullSnapshot(ctx context.Context) (*CreateSnapshotResponse, error) {
	path := "/snapshots"

	req, err := c.NewRequest(ctx, http.MethodPost, path, nil)
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

	var response CreateSnapshotResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &response, nil
}

// ListFullSnapshots lists all snapshots of the entire Qdrant instance.
func (c *Client) ListFullSnapshots(ctx context.Context) (*ListSnapshotsResponse, error) {
	path := "/snapshots"

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

	var response ListSnapshotsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &response, nil
}

// DeleteFullSnapshot deletes a full snapshot by name.
func (c *Client) DeleteFullSnapshot(ctx context.Context, snapshotName string) (*DeleteSnapshotResponse, error) {
	path := fmt.Sprintf("/snapshots/%s", snapshotName)

	req, err := c.NewRequest(ctx, http.MethodDelete, path, nil)
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

	var response DeleteSnapshotResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &response, nil
}

// RecoverFullSnapshot recovers the entire Qdrant instance from a snapshot at the given location.
func (c *Client) RecoverFullSnapshot(ctx context.Context, location string) (*RecoverSnapshotResponse, error) {
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
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var response RecoverSnapshotResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &response, nil
}

// CreateCollectionSnapshot creates a snapshot of a specific collection.
func (c *Client) CreateCollectionSnapshot(ctx context.Context, collectionName string) (*CreateSnapshotResponse, error) {
	path := fmt.Sprintf("/collections/%s/snapshots", collectionName)

	req, err := c.NewRequest(ctx, http.MethodPost, path, nil)
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

	var response CreateSnapshotResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &response, nil
}

// ListCollectionSnapshots lists all snapshots of a specific collection.
func (c *Client) ListCollectionSnapshots(ctx context.Context, collectionName string) (*ListSnapshotsResponse, error) {
	path := fmt.Sprintf("/collections/%s/snapshots", collectionName)

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

	var response ListSnapshotsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &response, nil
}

// DeleteCollectionSnapshot deletes a specific snapshot of a collection.
func (c *Client) DeleteCollectionSnapshot(ctx context.Context, collectionName string, snapshotName string) (*DeleteSnapshotResponse, error) {
	path := fmt.Sprintf("/collections/%s/snapshots/%s", collectionName, snapshotName)

	req, err := c.NewRequest(ctx, http.MethodDelete, path, nil)
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

	var response DeleteSnapshotResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &response, nil
}

// RecoverCollectionSnapshot uploads and restores a collection from a snapshot file
func (c *Client) RecoverCollectionSnapshot(
	ctx context.Context,
	collectionName string,
	snapshotPath string,
) (*RecoverSnapshotResponse, error) {
	// Endpoint
	urlPath := fmt.Sprintf("/collections/%s/snapshots/upload", collectionName)

	// Open snapshot file
	file, err := os.Open(snapshotPath)
	if err != nil {
		return nil, fmt.Errorf("opening snapshot file: %w", err)
	}
	defer func() { _ = file.Close() }()

	// Prepare multipart body
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("snapshot", filepath.Base(snapshotPath))
	if err != nil {
		return nil, fmt.Errorf("creating form file: %w", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return nil, fmt.Errorf("copying file data: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("closing multipart writer: %w", err)
	}

	// Create request
	req, err := c.NewRequest(ctx, http.MethodPost, urlPath, body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Execute request
	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Handle non-200 response
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Decode response
	var response RecoverSnapshotResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &response, nil
}

// DownloadCollectionSnapshot downloads a specific collection snapshot.
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
		_ = resp.Body.Close()
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return resp.Body, nil
}

// CreateShardSnapshot creates a snapshot of a specific shard.
func (c *Client) CreateShardSnapshot(ctx context.Context, collectionName string, shardID string) (*CreateSnapshotResponse, error) {
	path := fmt.Sprintf("/collections/%s/shards/%s/snapshots", collectionName, shardID)

	req, err := c.NewRequest(ctx, http.MethodPost, path, nil)
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

	var response CreateSnapshotResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &response, nil
}

// ListShardSnapshots lists all snapshots of a specific shard.
// ListShardSnapshots lists all snapshots of a specific shard.
func (c *Client) ListShardSnapshots(ctx context.Context, collectionName string, shardID string) (*ListSnapshotsResponse, error) {
	path := fmt.Sprintf("/collections/%s/shards/%s/snapshots", collectionName, shardID)

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

	var response ListSnapshotsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &response, nil
}

// DeleteShardSnapshot deletes a specific snapshot of a shard.
// DeleteShardSnapshot deletes a specific snapshot of a shard.
func (c *Client) DeleteShardSnapshot(ctx context.Context, collectionName string, shardID string, snapshotName string) (*DeleteSnapshotResponse, error) {
	path := fmt.Sprintf("/collections/%s/shards/%s/snapshots/%s", collectionName, shardID, snapshotName)

	req, err := c.NewRequest(ctx, http.MethodDelete, path, nil)
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

	var response DeleteSnapshotResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &response, nil
}

// RecoverShardSnapshot recovers a specific shard from a snapshot at the given location.
func (c *Client) RecoverShardSnapshot(ctx context.Context, collectionName string, shardID string, snapshotName string, location string) (*RecoverSnapshotResponse, error) {
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
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var response RecoverSnapshotResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &response, nil
}

// DownloadShardSnapshot downloads a specific shard snapshot.
func (c *Client) DownloadShardSnapshot(ctx context.Context, collectionName string, shardID string, snapshotName string) (io.ReadCloser, error) {
	path := fmt.Sprintf("/collections/%s/shards/%s/snapshots/%s", collectionName, shardID, snapshotName)

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
		_ = resp.Body.Close()
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return resp.Body, nil
}
