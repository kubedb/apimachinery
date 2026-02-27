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
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"go.etcd.io/etcd/raft/v3/raftpb"
	"k8s.io/klog/v2"
)

// AuthChecker is a function that checks authentication for HTTP requests
type AuthChecker func(r *http.Request) error

// Handler for a http based key-value store backed by raft
type httpKVAPI struct {
	raftNode            *RaftNode
	petSetName          string
	namespace           string
	store               *Kvstore
	confChangeC         chan<- raftpb.ConfChange
	TransferLeadershipC chan<- TransferLeadershipConfig
	authChecker         AuthChecker
}

func (h *httpKVAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer func() {
			if cerr := r.Body.Close(); cerr != nil {
				klog.Warningf("failed to close request body: %v", cerr)
			}
		}()
	}
	if h.authChecker == nil {
		w.WriteHeader(http.StatusUnauthorized)
		encodeErr := json.NewEncoder(w).Encode(fmt.Errorf("auth checker not configured"))
		if encodeErr != nil {
			klog.Errorf("%s", encodeErr.Error())
		}
		klog.Error("auth checker is not configured")
		return
	}
	authErr := h.authChecker(r)
	if authErr != nil {
		w.WriteHeader(http.StatusUnauthorized)

		encodeErr := json.NewEncoder(w).Encode(authErr)
		if encodeErr != nil {
			klog.Errorf("%s", encodeErr.Error())
		}
		klog.Error("user or password mismatched. Error:", authErr)
		return
	}
	subUrl := r.URL.Path
	subUrl = strings.TrimSpace(subUrl)
	subUrl = strings.Trim(subUrl, "/")

	switch subUrl {
	case "make-learner":
		switch r.Method {
		case http.MethodPost:
			var node *NodeInfo
			err := json.NewDecoder(r.Body).Decode(&node)
			if err != nil {
				klog.Infoln("Failed to read on POST. Error:", err)
				http.Error(w, "failed to decode request body", http.StatusBadRequest)
				return

			}
			if node == nil || node.NodeId == nil {
				http.Error(w, "Failed on POST. parsed request body's nodeID nil", http.StatusBadRequest)
				return
			}
			cc := raftpb.ConfChange{
				Type:    raftpb.ConfChangeAddLearnerNode,
				NodeID:  uint64(*node.NodeId),
				Context: []byte(strconv.Itoa(*node.NodeId)),
			}
			h.confChangeC <- cc

			// As above, optimistic that raft will apply the conf change
			w.WriteHeader(http.StatusOK)
			_, err = w.Write([]byte("the requested node is going to be a learner"))
			if err != nil {
				http.Error(w, "Failed on POST", http.StatusBadRequest)
			}
		default:
			w.Header().Add("Allow", http.MethodPost)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case "add-node":
		switch r.Method {
		case http.MethodPost:
			var node *NodeInfo

			err := json.NewDecoder(r.Body).Decode(&node)
			if err != nil {
				http.Error(w, "Failed on POST", http.StatusBadRequest)
				klog.Infoln("Failed to read on POST. Error:", err)
				return

			}

			if node == nil || node.NodeId == nil || node.Url == nil {
				klog.Infoln("error node id or url can't be empty")
				http.Error(w, "Failed on POST", http.StatusBadRequest)
				return
			}

			cc := raftpb.ConfChange{
				Type:    raftpb.ConfChangeAddNode,
				NodeID:  uint64(*node.NodeId),
				Context: []byte(*node.Url),
			}
			h.confChangeC <- cc
			w.WriteHeader(http.StatusOK)
			// As above, optimistic that raft will apply the conf change
			_, err = w.Write([]byte("the requested node is adding to the cluster"))
			if err != nil {
				http.Error(w, "Failed on POST", http.StatusBadRequest)
			}
		default:
			w.Header().Add("Allow", http.MethodPost)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case "get-nodes":
		switch r.Method {
		case http.MethodGet:
			confState := h.raftNode.ConfState()
			nodes := confState.Voters
			learners := confState.Learners
			nodes = append(nodes, learners...)
			// As above, optimistic that raft will apply the leader transfer
			w.WriteHeader(http.StatusOK)
			err := json.NewEncoder(w).Encode(nodes)
			if err != nil {
				http.Error(w, "Failed on POST", http.StatusBadRequest)
			}
		default:
			w.Header().Add("Allow", http.MethodGet)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case "remove-node":
		switch r.Method {
		case http.MethodDelete:
			var node *NodeInfo
			err := json.NewDecoder(r.Body).Decode(&node)
			if err != nil {
				http.Error(w, "Failed on DELETE", http.StatusBadRequest)
				klog.Infoln("Failed  on DELETE. Error:", err)
				return
			}
			if node == nil || node.NodeId == nil {

				http.Error(w, "Failed on POST. node can't be empty", http.StatusBadRequest)
				return
			}
			cc := raftpb.ConfChange{
				Type:   raftpb.ConfChangeRemoveNode,
				NodeID: uint64(*node.NodeId),
			}
			h.confChangeC <- cc

			// As above, optimistic that raft will apply the conf change
			w.WriteHeader(http.StatusOK)
			_, err = w.Write([]byte("the requested node is removing from the cluster"))
			if err != nil {
				http.Error(w, "Failed on DELETE", http.StatusBadRequest)
			}
		default:
			w.Header().Add("Allow", http.MethodDelete)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case "transfer":
		switch r.Method {
		case http.MethodPost:
			var transfer TransferLeadershipConfig
			err := json.NewDecoder(r.Body).Decode(&transfer)
			if err != nil {
				http.Error(w, "Failed on POST", http.StatusBadRequest)
				klog.Infoln("Failed to read on POST. Error:", err)
				return
			}

			if transfer.Transferee == nil {
				http.Error(w, "Failed on transfer . transferee can't be empty", http.StatusBadRequest)
				klog.Infoln("Failed to read on POST. Error: ", err)
				return
			}

			h.TransferLeadershipC <- transfer
			// As above, optimistic that raft will apply the leader transfer
			w.WriteHeader(http.StatusOK)
			_, err = w.Write([]byte("leader role is  transferring to the requested node"))
			if err != nil {
				http.Error(w, "Failed on POST", http.StatusBadRequest)
			}
		default:
			w.Header().Add("Allow", http.MethodPost)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case "set":
		switch r.Method {
		case http.MethodPost:
			var keyValue KeyValue
			err := json.NewDecoder(r.Body).Decode(&keyValue)
			if err != nil {
				http.Error(w, "Failed on POST", http.StatusBadRequest)
				klog.Infoln("Failed to read on POST. Error:", err)
				return
			}

			if keyValue.Key == nil || keyValue.Value == nil {
				http.Error(w, "key or value can't be empty", http.StatusBadRequest)
				klog.Infoln("Failed to read on POST. Error: ", err)
				return
			}
			h.store.Propose(*keyValue.Key, *keyValue.Value)
			// As above, optimistic that raft will apply the set key = value
			w.WriteHeader(http.StatusOK)
			_, err = fmt.Fprintf(w, "proposed key saved. %s:%s", *keyValue.Key, *keyValue.Value)
			if err != nil {
				http.Error(w, "Failed on POST", http.StatusBadRequest)
			}
		default:
			w.Header().Add("Allow", http.MethodPost)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case "get":
		switch r.Method {
		case http.MethodGet:
			var keyValue KeyValue
			err := json.NewDecoder(r.Body).Decode(&keyValue)
			if err != nil {
				http.Error(w, "Failed on GET", http.StatusBadRequest)
				klog.Infoln("Failed to read on GET. Error:", err)
				return
			}

			if keyValue.Key == nil {
				http.Error(w, "key can't be empty", http.StatusBadRequest)
				klog.Infoln("Failed to read on GET. Error: ", err)
				return
			}
			value, ok := h.store.Lookup(*keyValue.Key)
			var respKeyValue KeyValue

			if ok {
				respKeyValue = KeyValue{
					Key:   keyValue.Key,
					Value: &value,
				}
				w.WriteHeader(http.StatusOK)
				err := json.NewEncoder(w).Encode(&respKeyValue)
				if err != nil {
					http.Error(w, "can't convert key-value into json", http.StatusInternalServerError)
				}
			} else {
				// As above, optimistic that raft will apply the leader transfer
				w.WriteHeader(http.StatusNotFound)
				_, err = w.Write([]byte("key not found"))
				if err != nil {
					http.Error(w, "Failed on POST", http.StatusBadRequest)
				}
			}
		default:
			w.Header().Add("Allow", http.MethodGet)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case "current-primary":
		switch r.Method {
		case http.MethodGet:
			primary := h.raftNode.Node.Status().Lead
			// As above, optimistic that raft will apply the leader transfer
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(strconv.FormatUint(primary, 10)))
			if err != nil {
				http.Error(w, "Failed on POST", http.StatusBadRequest)
			}
		default:
			w.Header().Add("Allow", http.MethodGet)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	default:
		http.Error(w, "not found", http.StatusNotFound)
	}
}

// HTTPKVAPIConfig holds configuration for the HTTP KV API server
type HTTPKVAPIConfig struct {
	Kvstore             *Kvstore
	RaftNode            *RaftNode
	PetSetName          string
	Namespace           string
	Port                int
	ConfChangeC         chan<- raftpb.ConfChange
	TransferLeadershipC chan<- TransferLeadershipConfig
	ErrorC              <-chan error
	AuthChecker         AuthChecker
}

// ServeHttpKVAPI starts a key-value server with a GET/PUT API and listens.
func ServeHttpKVAPI(cfg HTTPKVAPIConfig) {
	srv := http.Server{
		Addr: "0.0.0.0:" + strconv.Itoa(cfg.Port),
		Handler: &httpKVAPI{
			raftNode:            cfg.RaftNode,
			petSetName:          cfg.PetSetName,
			namespace:           cfg.Namespace,
			store:               cfg.Kvstore,
			confChangeC:         cfg.ConfChangeC,
			TransferLeadershipC: cfg.TransferLeadershipC,
			authChecker:         cfg.AuthChecker,
		},
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			klog.Fatalln(err)
		}
	}()
}
