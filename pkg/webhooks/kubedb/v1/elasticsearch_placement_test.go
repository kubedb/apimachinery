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

package v1

import (
	"strings"
	"testing"

	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"

	"gomodules.xyz/pointer"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	psapi "kubeops.dev/petset/apis/apps/v1"
)

func esTopoDB(master, ingest, data int32) *dbapi.Elasticsearch {
	return &dbapi.Elasticsearch{
		ObjectMeta: meta.ObjectMeta{Name: "es-topo", Namespace: "demo"},
		Spec: dbapi.ElasticsearchSpec{Distributed: true, Topology: &dbapi.ElasticsearchClusterTopology{
			Master: dbapi.ElasticsearchNode{Replicas: pointer.Int32P(master)},
			Ingest: dbapi.ElasticsearchNode{Replicas: pointer.Int32P(ingest)},
			Data:   &dbapi.ElasticsearchNode{Replicas: pointer.Int32P(data)},
		}},
	}
}

func esPlacement(rules ...psapi.DistributionRule) *psapi.PlacementPolicy {
	rules = append(rules, psapi.DistributionRule{ClusterName: "hub", Role: psapi.DCRoleArbiter})
	return &psapi.PlacementPolicy{
		ObjectMeta: meta.ObjectMeta{Name: "es-topo"},
		Spec: psapi.PlacementPolicySpec{ClusterSpreadConstraint: &psapi.ClusterSpreadConstraint{
			DistributionRules: rules,
			FailoverPolicy:    &psapi.FailoverPolicy{Mode: psapi.FailoverModeTwoDC, FailoverGroupRef: &core.LocalObjectReference{Name: "es-topo"}},
		}},
	}
}

func member(dc string, t map[string][]int32) psapi.DistributionRule {
	return psapi.DistributionRule{ClusterName: dc, Role: psapi.DCRoleMember, Topology: t}
}

func TestValidateElasticsearchTopologyPlacement(t *testing.T) {
	spoke1 := member("spoke-1", map[string][]int32{"master": {0, 1, 2}, "ingest": {0, 1}, "data": {0, 1}})
	spoke2 := member("spoke-2", map[string][]int32{"master": {3, 4, 5}, "ingest": {2, 3}, "data": {2, 3}})
	combined := esPlacement(
		psapi.DistributionRule{ClusterName: "spoke-1", ReplicaIndices: []int32{0, 1, 2}},
		psapi.DistributionRule{ClusterName: "spoke-2", ReplicaIndices: []int32{3, 4, 5}})
	cases := []struct {
		name string
		db   *dbapi.Elasticsearch
		pp   *psapi.PlacementPolicy
		err  string
	}{
		{"symmetric", esTopoDB(6, 4, 4), esPlacement(spoke1, spoke2), ""},
		{"uneven", esTopoDB(6, 3, 3), esPlacement(spoke1,
			member("spoke-2", map[string][]int32{"master": {3, 4, 5}, "ingest": {2}, "data": {2}})), ""},
		{"legacy policy without topology", esTopoDB(3, 2, 2), combined, ""},
		{"combined database", &dbapi.Elasticsearch{Spec: dbapi.ElasticsearchSpec{Distributed: true, Replicas: pointer.Int32P(6)}}, combined, ""},
		{"wrong total", esTopoDB(3, 4, 4), esPlacement(spoke1, spoke2), "'spec.topology.master.replicas' is 3 but PlacementPolicy \"es-topo\" places 6"},
		{"topology policy, combined database", &dbapi.Elasticsearch{Spec: dbapi.ElasticsearchSpec{Distributed: true}}, esPlacement(spoke1, spoke2), "'spec.topology' is not set"},
		{"role not configured", esTopoDB(6, 4, 4), esPlacement(spoke1,
			member("spoke-2", map[string][]int32{"master": {3, 4, 5}, "ingest": {2, 3}, "data": {2, 3}, "ml": {0}})), `places role "ml"`},
		{"no master in a DC", esTopoDB(3, 4, 4), esPlacement(spoke1,
			member("spoke-2", map[string][]int32{"ingest": {2, 3}, "data": {2, 3}})), `"spoke-2" needs at least one master node`},
		{"no data in a DC", esTopoDB(6, 4, 2), esPlacement(spoke1,
			member("spoke-2", map[string][]int32{"master": {3, 4, 5}, "ingest": {2, 3}})), `"spoke-2" needs at least one data node`},
		{"replicaIndices with topology", esTopoDB(6, 4, 4), esPlacement(spoke1,
			psapi.DistributionRule{ClusterName: "spoke-2", ReplicaIndices: []int32{3}, Topology: spoke2.Topology}), "sets both replicaIndices and topology"},
	}
	for _, tc := range cases {
		err := validateElasticsearchTopologyPlacement(tc.db, tc.pp)
		switch {
		case tc.err == "" && err != nil:
			t.Errorf("%s: unexpected error %v", tc.name, err)
		case tc.err != "" && (err == nil || !strings.Contains(err.Error(), tc.err)):
			t.Errorf("%s: want error containing %q, got %v", tc.name, tc.err, err)
		}
	}
}
