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

package raft

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"kubedb.dev/apimachinery/apis/kubedb"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	hanadbraft "kubedb.dev/apimachinery/pkg/utils/raft/hanadb"
)

const leaderAPIRequestTimeout = 3 * time.Second

type raftNodeInfo struct {
	NodeID *int    `json:"id" protobuf:"varint,1,opt,name=id"`
	URL    *string `json:"url,omitempty" protobuf:"bytes,2,opt,name=url"`
}

// GetCurrentLeaderID queries raft leader id from a coordinator pod.
func GetCurrentLeaderID(db *api.HanaDB, podName string, user, pass string) (uint64, error) {
	dnsName := hanadbraft.GetGoverningServiceDNSName(podName, db)
	url := "http://" + dnsName + ":" + strconv.Itoa(kubedb.HanaDBCoordinatorClientPort) + "/current-primary"

	defaultLead := uint64(0)
	client := &http.Client{Timeout: leaderAPIRequestTimeout}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return defaultLead, err
	}

	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	req.SetBasicAuth(user, pass)
	resp, err := client.Do(req)
	if err != nil {
		return defaultLead, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return defaultLead, err
	}

	response := strings.TrimSpace(string(bodyText))
	podID, err := strconv.ParseUint(response, 10, 64)
	if err != nil {
		return defaultLead, err
	}
	if podID == 0 {
		return 0, fmt.Errorf("leader is not elected yet")
	}

	return podID, nil
}

// AddNodeToRaft requests raft membership add via coordinator /add-node endpoint.
func AddNodeToRaft(db *api.HanaDB, primaryPodName, podName string, nodeID int, user, pass string) (string, error) {
	primaryDNSName := hanadbraft.GetGoverningServiceDNSName(primaryPodName, db)
	primaryURL := "http://" + primaryDNSName + ":" + strconv.Itoa(kubedb.HanaDBCoordinatorClientPort) + "/add-node"

	dnsName := hanadbraft.GetGoverningServiceDNSName(podName, db)
	url := "http://" + dnsName + ":" + strconv.Itoa(kubedb.HanaDBCoordinatorPort)
	node := &raftNodeInfo{
		NodeID: &nodeID,
		URL:    &url,
	}

	return doRaftMembershipChange(http.MethodPost, primaryURL, node, user, pass, "add new node")
}

// RemoveNodeFromRaft requests raft membership remove via coordinator /remove-node endpoint.
func RemoveNodeFromRaft(db *api.HanaDB, primaryPodName string, nodeID int, user, pass string) (string, error) {
	primaryDNSName := hanadbraft.GetGoverningServiceDNSName(primaryPodName, db)
	primaryURL := "http://" + primaryDNSName + ":" + strconv.Itoa(kubedb.HanaDBCoordinatorClientPort) + "/remove-node"

	node := &raftNodeInfo{
		NodeID: &nodeID,
	}

	return doRaftMembershipChange(http.MethodDelete, primaryURL, node, user, pass, "remove node")
}

func doRaftMembershipChange(method, endpoint string, node *raftNodeInfo, user, pass, action string) (string, error) {
	requestByte, err := json.Marshal(node)
	if err != nil {
		return "", err
	}
	requestBody := bytes.NewReader(requestByte)

	httpClient := &http.Client{Timeout: leaderAPIRequestTimeout}
	req, err := http.NewRequest(method, endpoint, requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to %s: %w", action, err)
	}
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	req.SetBasicAuth(user, pass)

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to %s: %w", action, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to %s: %w", action, err)
	}
	return string(bodyText), nil
}

func GetRaftLeaderIDWithRetries(db *api.HanaDB, dbPodName, user, pass string, maxTries int, retryDelay time.Duration) (int, error) {
	var lastErr error
	for tries := 1; tries <= maxTries; tries++ {
		currentLeaderID, err := GetCurrentLeaderID(db, dbPodName, user, pass)
		if err == nil {
			return int(currentLeaderID), nil
		}
		lastErr = fmt.Errorf("failed on getting current leader: %w", err)
		time.Sleep(retryDelay)
	}
	return 0, fmt.Errorf("failed to get leader of raft cluster: %w", lastErr)
}

func GetRaftPrimaryNode(db *api.HanaDB, replicas int, user, pass string, maxTries int, retryDelay time.Duration) (int, error) {
	var lastErr error
	for rep := 0; rep < replicas; rep++ {
		podName := fmt.Sprintf("%s-%v", db.OffshootName(), rep)
		primaryPodID, err := GetRaftLeaderIDWithRetries(db, podName, user, pass, maxTries, retryDelay)
		if err == nil {
			return primaryPodID, nil
		}
		lastErr = err
	}
	return 0, lastErr
}

func AddRaftNodeWithRetries(db *api.HanaDB, primaryPodName, podName string, nodeID int, user, pass string, maxTries int, retryDelay time.Duration) error {
	var lastErr error
	for tries := 0; tries <= maxTries; tries++ {
		_, err := AddNodeToRaft(db, primaryPodName, podName, nodeID, user, pass)
		if err == nil {
			return nil
		}
		lastErr = err
		time.Sleep(retryDelay)
	}
	return fmt.Errorf("failed to add nodeId = %v to the raft: %w", nodeID, lastErr)
}

func RemoveRaftNodeWithRetries(db *api.HanaDB, primaryPodName string, nodeID int, user, pass string, maxTries int, retryDelay time.Duration) error {
	var lastErr error
	for tries := 0; tries <= maxTries; tries++ {
		_, err := RemoveNodeFromRaft(db, primaryPodName, nodeID, user, pass)
		if err == nil {
			return nil
		}
		lastErr = err
		time.Sleep(retryDelay)
	}
	return fmt.Errorf("failed to remove nodeId = %v from the raft: %w", nodeID, lastErr)
}
