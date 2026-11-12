/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

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

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func dcdrMariaDB(mode MariaDBMode, annotated bool) MariaDB {
	db := MariaDB{
		ObjectMeta: metav1.ObjectMeta{Name: "md", Namespace: "demo"},
		Spec: MariaDBSpec{
			Distributed: true,
			Topology:    &MariaDBTopology{Mode: ptr.To(mode)},
		},
	}
	db.Spec.PodTemplate.Spec.PodPlacementPolicy = &core.LocalObjectReference{Name: "md"}
	if annotated {
		db.Annotations = map[string]string{DCDREnabledAnnotation: "true"}
	}
	return db
}

func TestMariaDBIsDCDRDistributed(t *testing.T) {
	if !dcdrMariaDB(MariaDBModeGaleraCluster, true).IsDCDRDistributed() {
		t.Error("distributed + annotation must be DC-DR")
	}
	// without the annotation, distributed keeps meaning one Galera cluster stretched over spokes
	if dcdrMariaDB(MariaDBModeGaleraCluster, false).IsDCDRDistributed() {
		t.Error("distributed without the annotation must not be DC-DR")
	}
	db := dcdrMariaDB(MariaDBModeGaleraCluster, true)
	db.Spec.Distributed = false
	if db.IsDCDRDistributed() {
		t.Error("annotation without distributed must not be DC-DR")
	}
}

func TestMariaDBValidateDCDR(t *testing.T) {
	ok := dcdrMariaDB(MariaDBModeGaleraCluster, true)
	cases := []struct {
		name     string
		db       MariaDB
		dcs      []string
		replicas []int32
		wantErr  string
	}{
		{"valid 3+3", ok, []string{"spoke-1", "spoke-2"}, []int32{3, 3}, ""},
		{"valid 3+5", ok, []string{"spoke-1", "spoke-2"}, []int32{3, 5}, ""},
		{"not DC-DR: nothing checked", dcdrMariaDB(MariaDBModeGaleraCluster, false), nil, nil, ""},
		{"replication topology", dcdrMariaDB(MariaDBModeReplication, true), []string{"a", "b"}, []int32{3, 3}, "GaleraCluster"},
		{"even node count", ok, []string{"spoke-1", "spoke-2"}, []int32{3, 4}, "odd number"},
		{"too small", ok, []string{"spoke-1", "spoke-2"}, []int32{1, 3}, "odd number"},
		{"one data center", ok, []string{"spoke-1"}, []int32{3}, "at least 2"},
		{"duplicate data center", ok, []string{"spoke-1", "spoke-1"}, []int32{3, 3}, "twice"},
	}
	for _, tc := range cases {
		err := tc.db.ValidateDCDR(tc.dcs, tc.replicas)
		if tc.wantErr == "" && err != nil {
			t.Errorf("%s: unexpected error %v", tc.name, err)
		}
		if tc.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tc.wantErr)) {
			t.Errorf("%s: got %v, want an error containing %q", tc.name, err, tc.wantErr)
		}
	}
	noPP := ok
	noPP.Spec.PodTemplate.Spec.PodPlacementPolicy = nil
	if err := noPP.ValidateDCDR([]string{"a", "b"}, []int32{3, 3}); err == nil {
		t.Error("missing podPlacementPolicy must be rejected")
	}
}

func TestMariaDBPerDCIDs(t *testing.T) {
	if MariaDBServerIDForDC(0) == MariaDBServerIDForDC(1) || MariaDBGTIDDomainIDForDC(0) == MariaDBGTIDDomainIDForDC(1) {
		t.Error("server_id and GTID domain id must differ between data centers")
	}
	db := dcdrMariaDB(MariaDBModeGaleraCluster, true)
	if db.PetSetNameForDC("spoke-1") != "md-spoke-1" || db.GoverningServiceNameForDC("spoke-1") != "md-spoke-1-pods" {
		t.Errorf("names: %s %s", db.PetSetNameForDC("spoke-1"), db.GoverningServiceNameForDC("spoke-1"))
	}
}
