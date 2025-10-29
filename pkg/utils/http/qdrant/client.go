package qdrant

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type ClusterResponse struct {
	Result ClusterResult `json:"result"`
	Status string        `json:"status"`
	Time   float64       `json:"time"`
}

type ClusterResult struct {
	Status               string                     `json:"status"`
	PeerID               *int64                     `json:"peer_id,omitempty"`
	Peers                map[string]ClusterPeer     `json:"peers"`
	RaftInfo             *RaftInfo                  `json:"raft_info,omitempty"`
	ConsensusThreadState map[string]json.RawMessage `json:"consensus_thread_status,omitempty"`
}

type ClusterPeer struct {
	URI string `json:"uri"`
}

type RaftInfo struct {
	Term              int64  `json:"term"`
	Commit            int64  `json:"commit"`
	PendingOperations int64  `json:"pending_operations"`
	Leader            *int64 `json:"leader,omitempty"`
	Role              string `json:"role,omitempty"`
	IsVoter           bool   `json:"is_voter,omitempty"`
}

func GetClusterResponse(ctx context.Context, address string, apiKey string) (*ClusterResponse, error) {

	req, err := http.NewRequestWithContext(ctx, "GET", "http://"+address+"/cluster", nil)
	if err != nil {
		return nil, err
	}

	if apiKey != "" {
		req.Header.Set("api-key", apiKey)
	}

	req.Header.Set("Accept", "application/json")

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("bad status %d from %s (try checking API key or headers)", res.StatusCode, address)
	}

	var cr ClusterResponse
	if err := json.NewDecoder(res.Body).Decode(&cr); err != nil {
		return nil, fmt.Errorf("decode err: %w", err)
	}

	return &cr, nil
}

func GetClusterStatus(ctx context.Context, address string, apiKey string) (string, error) {
	cr, err := GetClusterResponse(ctx, address, apiKey)
	if err != nil {
		return "", err
	}
	return cr.Result.Status, nil
}

func GetReplicaCount(ctx context.Context, address string, apiKey string) (int, error) {
	cr, err := GetClusterResponse(ctx, address, apiKey)
	if err != nil {
		return 0, err
	}
	return len(cr.Result.Peers), nil
}
