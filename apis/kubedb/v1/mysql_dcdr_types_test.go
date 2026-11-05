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
	"testing"
	"time"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func dcdrMySQL() MySQL {
	m := MySQL{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "my",
			Namespace:   "demo",
			Annotations: map[string]string{DCDREnabledAnnotation: "true"},
		},
		Spec: MySQLSpec{
			Distributed: true,
			Topology: &MySQLTopology{
				Mode: ptr.To(MySQLModeGroupReplication),
			},
		},
	}
	m.Spec.PodTemplate.Spec.PodPlacementPolicy = &core.LocalObjectReference{Name: "my-dcdr"}
	return m
}

func TestMySQLIsDCDRDistributed(t *testing.T) {
	m := dcdrMySQL()
	if !m.IsDCDRDistributed() {
		t.Fatal("expected DC-DR distributed")
	}
	m.Annotations = nil
	if m.IsDCDRDistributed() {
		t.Fatal("annotation missing must not enable DC-DR")
	}
	m = dcdrMySQL()
	m.Spec.Distributed = false
	if m.IsDCDRDistributed() {
		t.Fatal("non distributed must not enable DC-DR")
	}
}

func TestMySQLDCNames(t *testing.T) {
	m := dcdrMySQL()
	if got, want := m.PetSetNameForDC("spoke-1"), m.OffshootName()+"-spoke-1"; got != want {
		t.Fatalf("petset name %q, want %q", got, want)
	}
	if got, want := m.GoverningServiceNameForDC("spoke-2"), m.OffshootName()+"-spoke-2-pods"; got != want {
		t.Fatalf("governing service %q, want %q", got, want)
	}
	if m.PetSetNameForDC("spoke-1") == m.PetSetNameForDC("spoke-2") {
		t.Fatal("two DCs must not share a PetSet name")
	}
	if ServerIDOffsetForDC(0) == ServerIDOffsetForDC(1) {
		t.Fatal("server_id offsets of two DCs must differ")
	}
}

func TestMySQLValidateDCDR(t *testing.T) {
	two := []string{"spoke-1", "spoke-2"}
	tests := []struct {
		name    string
		mutate  func(*MySQL)
		dcs     []string
		wantErr bool
	}{
		{"valid", func(m *MySQL) {}, two, false},
		{"not DC-DR is skipped", func(m *MySQL) { m.Annotations = nil; m.Spec.Topology = nil }, nil, false},
		{"not group replication", func(m *MySQL) { m.Spec.Topology.Mode = ptr.To(MySQLModeSemiSync) }, two, true},
		{"no topology", func(m *MySQL) { m.Spec.Topology = nil }, two, true},
		{"multi primary", func(m *MySQL) {
			m.Spec.Topology.Group = &MySQLGroupSpec{Mode: ptr.To(MySQLGroupModeMultiPrimary)}
		}, two, true},
		{"no placement policy", func(m *MySQL) { m.Spec.PodTemplate.Spec.PodPlacementPolicy = nil }, two, true},
		{"one data center", func(m *MySQL) {}, []string{"spoke-1"}, true},
		{"duplicate data center", func(m *MySQL) {}, []string{"spoke-1", "spoke-1"}, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := dcdrMySQL()
			tc.mutate(&m)
			err := m.ValidateDCDR(tc.dcs)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err=%v, wantErr=%v", err, tc.wantErr)
			}
		})
	}
}

func TestMySQLWritablePrimaryConfirmed(t *testing.T) {
	now := time.Now()
	fresh := metav1.NewTime(now.Add(-time.Minute))
	stale := metav1.NewTime(now.Add(-2 * MySQLWritableObservationTTL))
	tests := []struct {
		name string
		dc   *MySQLDCStatus
		want bool
	}{
		{"nil", nil, false},
		{"fresh writable", &MySQLDCStatus{Primary: "my-spoke-1-0", Writable: true, WritableObservedAt: &fresh}, true},
		{"never probed", &MySQLDCStatus{Primary: "my-spoke-1-0", Writable: true}, false},
		{"stale", &MySQLDCStatus{Primary: "my-spoke-1-0", Writable: true, WritableObservedAt: &stale}, false},
		{"no primary", &MySQLDCStatus{Writable: true, WritableObservedAt: &fresh}, false},
		{"read only", &MySQLDCStatus{Primary: "my-spoke-1-0", WritableObservedAt: &fresh}, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.dc.WritablePrimaryConfirmed(now); got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}
