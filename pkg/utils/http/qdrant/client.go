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
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"kubedb.dev/apimachinery/apis/kubedb"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

func GetClusterResponse(
	ctx context.Context,
	db *api.Qdrant,
	kc client.Client,
	address string,
) (*ClusterResponse, error) {
	scheme := "http"
	var transport *http.Transport

	if db.Spec.TLS != nil {
		scheme = "https"

		caSecret := &corev1.Secret{}
		err := kc.Get(ctx, types.NamespacedName{
			Name:      db.GetCertSecretName(api.QdrantServerCert),
			Namespace: db.Namespace,
		}, caSecret)
		if err != nil {
			return nil, fmt.Errorf("failed to get server cert secret: %w", err)
		}

		caPool := x509.NewCertPool()
		if !caPool.AppendCertsFromPEM(caSecret.Data["ca.crt"]) {
			return nil, fmt.Errorf("failed to append server CA cert")
		}

		tlsConfig := &tls.Config{
			RootCAs: caPool,
		}

		if db.Spec.TLS.ClientHTTPTLS != nil && *db.Spec.TLS.ClientHTTPTLS {
			clientSecret := &corev1.Secret{}
			err := kc.Get(ctx, types.NamespacedName{
				Name:      db.GetCertSecretName(api.QdrantClientCert),
				Namespace: db.Namespace,
			}, clientSecret)
			if err != nil {
				return nil, fmt.Errorf("failed to get client cert secret: %w", err)
			}

			cert, err := tls.X509KeyPair(
				clientSecret.Data["client.crt"],
				clientSecret.Data["client.key"],
			)
			if err != nil {
				return nil, fmt.Errorf("invalid client cert/key: %w", err)
			}

			tlsConfig.Certificates = []tls.Certificate{cert}
		}

		transport = &http.Transport{
			TLSClientConfig: tlsConfig,
		}
	}

	apiKey := db.GetAPIKey(ctx, kc)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s://%s/cluster", scheme, address),
		nil,
	)
	if err != nil {
		return nil, err
	}

	if apiKey != "" {
		req.Header.Set("api-key", apiKey)
	}
	req.Header.Set("Accept", "application/json")

	client := &http.Client{
		Timeout:   5 * time.Second,
		Transport: transport,
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"bad status %d from %s",
			res.StatusCode,
			address,
		)
	}

	var cr ClusterResponse
	if err := json.NewDecoder(res.Body).Decode(&cr); err != nil {
		return nil, fmt.Errorf("decode err: %w", err)
	}

	return &cr, nil
}

func GetClusterStatus(ctx context.Context, db *api.Qdrant, kc client.Client) (string, int, []string, error) {
	address := db.ServiceDNS() + ":" + strconv.Itoa(kubedb.QdrantHTTPPort)

	cr, err := GetClusterResponse(ctx, db, kc, address)
	if err != nil {
		return "", 0, nil, err
	}

	replicaCount := len(cr.Result.Peers)
	roles := []string{}

	for i := 0; i < replicaCount; i++ {
		podAddress := db.GetPodAddress(i)
		cr, err = GetClusterResponse(ctx, db, kc, podAddress)
		if err != nil {
			return "", 0, nil, err
		}
		roles = append(roles, cr.Result.RaftInfo.Role)
	}

	return cr.Result.Status, replicaCount, roles, nil
}
