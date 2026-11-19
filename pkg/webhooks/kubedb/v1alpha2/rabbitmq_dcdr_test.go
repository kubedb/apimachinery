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

package v1alpha2

import (
	"context"
	"testing"

	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	kmapi "kmodules.xyz/client-go/api/v1"
	psapi "kubeops.dev/petset/apis/apps/v1"
)

func dcdrRM() (*olddbapi.RabbitMQ, *psapi.PlacementPolicy) {
	db := &olddbapi.RabbitMQ{ObjectMeta: metav1.ObjectMeta{Name: "rmq", Namespace: "demo"}}
	db.Spec.Distributed = true
	db.Spec.Replicas = ptr.To[int32](6)
	db.Spec.PodTemplate.Spec.PodPlacementPolicy = &core.LocalObjectReference{Name: "rmq-dr"}
	pp := &psapi.PlacementPolicy{ObjectMeta: metav1.ObjectMeta{Name: "rmq-dr"}}
	pp.Spec.ClusterSpreadConstraint = &psapi.ClusterSpreadConstraint{
		DistributionRules: []psapi.DistributionRule{
			{ClusterName: "spoke-1", ReplicaIndices: []int32{0, 1, 2}, Role: psapi.DCRoleMember},
			{ClusterName: "spoke-2", ReplicaIndices: []int32{3, 4, 5}, Role: psapi.DCRoleMember},
			{ClusterName: "hub", Role: psapi.DCRoleArbiter},
		},
		FailoverPolicy: &psapi.FailoverPolicy{Mode: psapi.FailoverModeTwoDC},
	}
	return db, pp
}

func TestValidateDCDRRabbitMQ(t *testing.T) {
	db, pp := dcdrRM()
	if w, err := validateDCDRRabbitMQ(db, pp); err != nil || len(w) != 0 {
		t.Fatalf("valid 3+3 refused: %v %v", w, err)
	}

	db, pp = dcdrRM()
	pp.Spec.ClusterSpreadConstraint.FailoverPolicy = nil
	if _, err := validateDCDRRabbitMQ(db, pp); err == nil {
		t.Error("a policy without failoverPolicy must be refused")
	}

	db, pp = dcdrRM()
	pp.Spec.ClusterSpreadConstraint.DistributionRules[1].ReplicaIndices = []int32{3, 4}
	if _, err := validateDCDRRabbitMQ(db, pp); err == nil {
		t.Error("ordinal 5 of 6 unplaced must be refused")
	}

	db, pp = dcdrRM()
	pp.Spec.ClusterSpreadConstraint.DistributionRules[2].ReplicaIndices = []int32{6}
	if _, err := validateDCDRRabbitMQ(db, pp); err == nil {
		t.Error("nodes on the arbiter must be refused")
	}

	db, pp = dcdrRM()
	db.Spec.DisableSecurity = true
	if _, err := validateDCDRRabbitMQ(db, pp); err == nil {
		t.Error("disableSecurity must be refused")
	}

	db, pp = dcdrRM()
	db.Spec.TLS = &kmapi.TLSConfig{}
	if _, err := validateDCDRRabbitMQ(db, pp); err == nil {
		t.Error("tls must be refused for now")
	}

	db, pp = dcdrRM()
	db.Spec.StorageType = olddbapi.StorageTypeEphemeral
	if _, err := validateDCDRRabbitMQ(db, pp); err == nil {
		t.Error("ephemeral storage must be refused")
	}

	db, pp = dcdrRM()
	db.Name = "a-rabbitmq-name-that-is-far-too-long"
	if _, err := validateDCDRRabbitMQ(db, pp); err == nil {
		t.Error("a long name must be refused")
	}

	db, pp = dcdrRM()
	db.Spec.Replicas = ptr.To[int32](2)
	pp.Spec.ClusterSpreadConstraint.DistributionRules[0].ReplicaIndices = []int32{0}
	pp.Spec.ClusterSpreadConstraint.DistributionRules[1].ReplicaIndices = []int32{1}
	if w, err := validateDCDRRabbitMQ(db, pp); err != nil || len(w) == 0 {
		t.Errorf("1+1 must be allowed with a warning: %v %v", w, err)
	}

	db, pp = dcdrRM()
	pp.Spec.ClusterSpreadConstraint.DistributionRules = append(pp.Spec.ClusterSpreadConstraint.DistributionRules,
		psapi.DistributionRule{ClusterName: "spoke-3", ReplicaIndices: []int32{6}, Role: psapi.DCRoleMember})
	db.Spec.Replicas = ptr.To[int32](7)
	if _, err := validateDCDRRabbitMQ(db, pp); err == nil {
		t.Error("three member data centers must be refused")
	}
}

func TestValidateDistributedRabbitMQSkipsNonDistributed(t *testing.T) {
	db, _ := dcdrRM()
	db.Spec.Distributed = false
	w := RabbitMQCustomWebhook{}
	if _, err := w.validateDistributedRabbitMQ(context.TODO(), db); err != nil {
		t.Errorf("a non distributed RabbitMQ is not checked: %v", err)
	}
	db.Spec.Distributed = true
	db.Spec.PodTemplate.Spec.PodPlacementPolicy = nil
	if _, err := w.validateDistributedRabbitMQ(context.TODO(), db); err == nil {
		t.Error("a distributed RabbitMQ without a placement policy must be refused")
	}
}
