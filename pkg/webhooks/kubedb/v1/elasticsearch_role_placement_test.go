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
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	psapi "kubeops.dev/petset/apis/apps/v1"
)

func esRolesDB(master, ingest, data int32) *dbapi.Elasticsearch {
	return &dbapi.Elasticsearch{
		ObjectMeta: meta.ObjectMeta{Name: "es-topo", Namespace: "demo"},
		Spec: dbapi.ElasticsearchSpec{Distributed: true, Topology: &dbapi.ElasticsearchClusterTopology{
			Master: dbapi.ElasticsearchNode{Replicas: pointer.Int32P(master)},
			Ingest: dbapi.ElasticsearchNode{Replicas: pointer.Int32P(ingest)},
			Data:   &dbapi.ElasticsearchNode{Replicas: pointer.Int32P(data)},
		}},
	}
}

var esGroupFP = map[string]any{"mode": "TwoDC", "failoverGroupRef": map[string]any{"name": "es-topo"}}

// rolePP is a DC-DR role policy: spoke-1 and spoke-2 Members with the given indices, hub Arbiter.
func rolePP(name string, fp map[string]any, s1, s2 []int32) esRolePlacement {
	rules := []psapi.DistributionRule{
		{ClusterName: "spoke-1", ReplicaIndices: s1, Role: psapi.DCRoleMember},
		{ClusterName: "spoke-2", ReplicaIndices: s2, Role: psapi.DCRoleMember},
		{ClusterName: "hub", ReplicaIndices: []int32{}, Role: psapi.DCRoleArbiter},
	}
	pp := &psapi.PlacementPolicy{ObjectMeta: meta.ObjectMeta{Name: name}}
	pp.Spec.ClusterSpreadConstraint = &psapi.ClusterSpreadConstraint{DistributionRules: rules}
	return esRolePlacement{pp: pp, failoverPolicy: fp}
}

func TestValidateElasticsearchRolePlacement(t *testing.T) {
	master := rolePP("es-topo-master", esGroupFP, []int32{0, 1, 2}, []int32{3, 4, 5})
	ingest := rolePP("es-topo-ingest", esGroupFP, []int32{0, 1}, []int32{2, 3})
	data := rolePP("es-topo-data", esGroupFP, []int32{0, 1}, []int32{2, 3})
	set := func(m, i, d esRolePlacement) map[string]esRolePlacement {
		return map[string]esRolePlacement{"master": m, "ingest": i, "data": d}
	}
	noArbiter := rolePP("es-topo-data", esGroupFP, []int32{0, 1}, []int32{2, 3})
	noArbiter.pp.Spec.ClusterSpreadConstraint.DistributionRules = noArbiter.pp.Spec.ClusterSpreadConstraint.DistributionRules[:2]
	cases := []struct {
		name  string
		db    *dbapi.Elasticsearch
		roles map[string]esRolePlacement
		err   string
	}{
		{"symmetric", esRolesDB(6, 4, 4), set(master, ingest, data), ""},
		{"uneven ingest", esRolesDB(6, 3, 4), set(master, rolePP("i", esGroupFP, []int32{0, 1}, []int32{2}), data), ""},
		{"wrong total", esRolesDB(3, 4, 4), set(master, ingest, data), "'spec.topology.master.replicas' is 3 but PlacementPolicy \"es-topo-master\" places 6"},
		{"role without its policy", esRolesDB(6, 4, 4), map[string]esRolePlacement{"master": master, "ingest": ingest}, "'spec.topology.data.podTemplate.spec.podPlacementPolicy' is required when 'spec.distributed' is true"},
		{"not a DC-DR policy", esRolesDB(6, 4, 4), set(master, ingest, rolePP("d", nil, []int32{0, 1}, []int32{2, 3})), "is not a DC-DR PlacementPolicy"},
		{"different failover group", esRolesDB(6, 4, 4), set(master, ingest, rolePP("d", map[string]any{"mode": "TwoDC", "failoverGroupRef": map[string]any{"name": "other"}}, []int32{0, 1}, []int32{2, 3})), "different failoverPolicy"},
		{"different Arbiters", esRolesDB(6, 4, 4), set(master, ingest, noArbiter), "they must match"},
		{"duplicate index", esRolesDB(6, 4, 4), set(master, ingest, rolePP("d", esGroupFP, []int32{0, 1}, []int32{1, 3})), "places index 1 on both"},
		{"combined", &dbapi.Elasticsearch{Spec: dbapi.ElasticsearchSpec{Distributed: true, Replicas: pointer.Int32P(6)}},
			map[string]esRolePlacement{"combined": rolePP("es-comb", map[string]any{"mode": "TwoDC"}, []int32{0, 1, 2}, []int32{3, 4, 5})}, ""},
		{"combined wrong total", &dbapi.Elasticsearch{Spec: dbapi.ElasticsearchSpec{Distributed: true, Replicas: pointer.Int32P(3)}},
			map[string]esRolePlacement{"combined": rolePP("es-comb", map[string]any{"mode": "TwoDC"}, []int32{0, 1, 2}, []int32{3, 4, 5})}, "'spec.replicas' is 3 but PlacementPolicy \"es-comb\" places 6"},
		{"no data node in a DC", esRolesDB(6, 4, 2), set(master, ingest, rolePP("d", esGroupFP, []int32{0, 1}, []int32{})), `data center "spoke-2" needs at least one data node`},
	}
	for _, tc := range cases {
		err := validateElasticsearchPlacement(tc.db, tc.roles)
		switch {
		case tc.err == "" && err != nil:
			t.Errorf("%s: unexpected error %v", tc.name, err)
		case tc.err != "" && (err == nil || !strings.Contains(err.Error(), tc.err)):
			t.Errorf("%s: want error containing %q, got %v", tc.name, tc.err, err)
		}
	}
}
