// Copyright 2015 The etcd Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package raft

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"sync"

	"go.etcd.io/etcd/raft/v3/raftpb"
	"go.etcd.io/etcd/server/v3/etcdserver/api/snap"
	"k8s.io/klog/v2"
)

// Kvstore a key-value store backed by raft
type Kvstore struct {
	proposeC    chan<- string // channel for proposing updates
	mu          sync.RWMutex
	kvStore     map[string]string // current committed key-value pairs
	snapshotter *snap.Snapshotter
}

type kv struct {
	Key string
	Val string
}

func NewKVStore(snapshotter *snap.Snapshotter, proposeC chan<- string, commitC <-chan *Commit, errorC <-chan error) (*Kvstore, error) {
	s := &Kvstore{proposeC: proposeC, kvStore: make(map[string]string), snapshotter: snapshotter}
	snapshot, err := s.loadSnapshot()
	if err != nil {
		return nil, err
	}
	if snapshot != nil {
		klog.Infoln(fmt.Sprintf("loading snapshot at term %d and index %d", snapshot.Metadata.Term, snapshot.Metadata.Index))
		if err := s.recoverFromSnapshot(snapshot.Data); err != nil {
			return nil, err
		}
	}
	// read commits from raft into kvStore map until error
	go s.readCommits(commitC, errorC)
	return s, nil
}

func (s *Kvstore) Lookup(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.kvStore[key]
	return v, ok
}

func (s *Kvstore) Propose(k string, v string) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(kv{k, v}); err != nil {
		klog.Fatalln(err)
	}
	s.proposeC <- buf.String()
}

func (s *Kvstore) readCommits(commitC <-chan *Commit, errorC <-chan error) {
	for commit := range commitC {
		if commit == nil {
			// signaled to load snapshot
			snapshot, err := s.loadSnapshot()
			if err != nil {
				panic(err)
			}
			if snapshot != nil {
				klog.Infoln(fmt.Sprintf("loading snapshot at term %d and index %d", snapshot.Metadata.Term, snapshot.Metadata.Index))
				if err := s.recoverFromSnapshot(snapshot.Data); err != nil {
					panic(err)
				}
			}
			continue
		}

		for _, data := range commit.Data {
			var dataKv kv
			dec := gob.NewDecoder(bytes.NewBufferString(data))
			if err := dec.Decode(&dataKv); err != nil {
				klog.Fatalf("raftexample: could not decode message (%v)", err)
			}
			s.mu.Lock()
			s.kvStore[dataKv.Key] = dataKv.Val
			s.mu.Unlock()
		}
		close(commit.ApplyDoneC)
	}
	if err, ok := <-errorC; ok {
		klog.Fatalln(err)
	}
}

func (s *Kvstore) GetSnapshot() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return json.Marshal(s.kvStore)
}

func (s *Kvstore) loadSnapshot() (*raftpb.Snapshot, error) {
	snapshot, err := s.snapshotter.Load()
	if err == snap.ErrNoSnapshot {
		return nil, nil
	}
	return snapshot, err
}

func (s *Kvstore) recoverFromSnapshot(snapshot []byte) error {
	var store map[string]string
	if err := json.Unmarshal(snapshot, &store); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.kvStore = store
	return nil
}
