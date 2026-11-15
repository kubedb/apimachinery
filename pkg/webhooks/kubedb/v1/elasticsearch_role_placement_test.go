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

func esPolicy(name string, fp *psapi.FailoverPolicy, rules ...psapi.DistributionRule) *psapi.PlacementPolicy {
	return &psapi.PlacementPolicy{
		ObjectMeta: meta.ObjectMeta{Name: name},
		Spec:       psapi.PlacementPolicySpec{ClusterSpreadConstraint: &psapi.ClusterSpreadConstraint{DistributionRules: rules, FailoverPolicy: fp}},
	}
}

func dcRule(dc string, idx ...int32) psapi.DistributionRule {
	return psapi.DistributionRule{ClusterName: dc, ReplicaIndices: idx}
}

func TestValidateElasticsearchRolePlacement(t *testing.T) {
	top := esPolicy("es-topo", &psapi.FailoverPolicy{Mode: psapi.FailoverModeTwoDC},
		psapi.DistributionRule{ClusterName: "spoke-1", ReplicaIndices: []int32{0, 1, 2}, Role: psapi.DCRoleMember},
		psapi.DistributionRule{ClusterName: "spoke-2", ReplicaIndices: []int32{3, 4, 5}, Role: psapi.DCRoleMember},
		psapi.DistributionRule{ClusterName: "hub", ReplicaIndices: []int32{}, Role: psapi.DCRoleArbiter})
	master := esPolicy("es-topo-master", nil, dcRule("spoke-1", 0, 1, 2), dcRule("spoke-2", 3, 4, 5))
	ingest := esPolicy("es-topo-ingest", nil, dcRule("spoke-1", 0, 1), dcRule("spoke-2", 2, 3))
	data := esPolicy("es-topo-data", nil, dcRule("spoke-1", 0, 1), dcRule("spoke-2", 2, 3))
	roles := func(m, i, d *psapi.PlacementPolicy) map[string]*psapi.PlacementPolicy {
		out := map[string]*psapi.PlacementPolicy{"master": m, "ingest": i}
		if d != nil {
			out["data"] = d
		}
		return out
	}
	cases := []struct {
		name  string
		db    *dbapi.Elasticsearch
		roles map[string]*psapi.PlacementPolicy
		err   string
	}{
		{"symmetric", esRolesDB(6, 4, 4), roles(master, ingest, data), ""},
		{"uneven ingest", esRolesDB(6, 3, 4), roles(master, esPolicy("i", nil, dcRule("spoke-1", 0, 1), dcRule("spoke-2", 2)), data), ""},
		{"wrong total", esRolesDB(3, 4, 4), roles(master, ingest, data), "'spec.topology.master.replicas' is 3 but PlacementPolicy \"es-topo-master\" places 6"},
		{"role without its own policy", esRolesDB(6, 4, 4), roles(master, ingest, nil), "'spec.topology.data.podTemplate.spec.podPlacementPolicy' does not name a PlacementPolicy with a clusterSpreadConstraint"},
		{"not a Member", esRolesDB(6, 4, 4), roles(master, ingest, esPolicy("d", nil, dcRule("spoke-1", 0, 1), dcRule("hub", 2, 3))), `places nodes on "hub", which is not a Member`},
		{"role policy with failoverPolicy", esRolesDB(6, 4, 4), roles(master, ingest, esPolicy("d", &psapi.FailoverPolicy{}, dcRule("spoke-1", 0, 1), dcRule("spoke-2", 2, 3))), "must not set failoverPolicy"},
		{"duplicate index", esRolesDB(6, 4, 4), roles(master, ingest, esPolicy("d", nil, dcRule("spoke-1", 0, 1), dcRule("spoke-2", 1, 3))), "places index 1 on both"},
		{"no master in a DC", esRolesDB(3, 4, 4), roles(esPolicy("m", nil, dcRule("spoke-1", 0, 1, 2)), ingest, data), `data center "spoke-2" needs at least one master node`},
		{"no data in a DC", esRolesDB(6, 4, 2), roles(master, ingest, esPolicy("d", nil, dcRule("spoke-1", 0, 1))), `data center "spoke-2" needs at least one data node`},
	}
	for _, tc := range cases {
		err := validateElasticsearchRolePlacement(tc.db, top, tc.roles)
		switch {
		case tc.err == "" && err != nil:
			t.Errorf("%s: unexpected error %v", tc.name, err)
		case tc.err != "" && (err == nil || !strings.Contains(err.Error(), tc.err)):
			t.Errorf("%s: want error containing %q, got %v", tc.name, tc.err, err)
		}
	}
}
