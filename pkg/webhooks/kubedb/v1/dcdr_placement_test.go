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
	"testing"

	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	psapi "kubeops.dev/petset/apis/apps/v1"
)

// A PlacementPolicy in the failoverGroupRef shape decodes with an empty trigger into the vendored
// petset types. It must not block a DC-DR MariaDB (lab 2026-11-15: every update was refused with
// "trigger.scope must be Global or Group").
func TestValidateDCDRMariaDBFailoverGroupRefShape(t *testing.T) {
	db := &dbapi.MariaDB{ObjectMeta: metav1.ObjectMeta{Name: "md", Namespace: "demo", Annotations: map[string]string{dbapi.DCDREnabledAnnotation: "true"}}}
	db.Spec.Distributed = true
	db.Spec.Topology = &dbapi.MariaDBTopology{Mode: ptrMode(dbapi.MariaDBModeGaleraCluster)}
	db.Spec.PodTemplate.Spec.PodPlacementPolicy = &core.LocalObjectReference{Name: "md"}
	pp := &psapi.PlacementPolicy{ObjectMeta: metav1.ObjectMeta{Name: "md"}}
	pp.Spec.ClusterSpreadConstraint = &psapi.ClusterSpreadConstraint{
		DistributionRules: []psapi.DistributionRule{
			{ClusterName: "spoke-1", ReplicaIndices: []int32{0, 1, 2}, Role: psapi.DCRoleMember},
			{ClusterName: "spoke-2", ReplicaIndices: []int32{3, 4, 5}, Role: psapi.DCRoleMember},
			{ClusterName: "hub", Role: psapi.DCRoleArbiter},
		},
		FailoverPolicy: &psapi.FailoverPolicy{Mode: psapi.FailoverModeTwoDC},
	}
	if err := validateDCDRMariaDB(db, pp); err != nil {
		t.Fatalf("failoverGroupRef-shaped policy refused: %v", err)
	}
	pp.Spec.ClusterSpreadConstraint.FailoverPolicy = nil
	if err := validateDCDRMariaDB(db, pp); err == nil {
		t.Error("a policy without failoverPolicy must still be refused")
	}
}

func ptrMode(m dbapi.MariaDBMode) *dbapi.MariaDBMode { return &m }
