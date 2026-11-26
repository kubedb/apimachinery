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

package v1alpha1

import (
	"strings"
	"testing"

	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMySQLDCDROpsRefusal(t *testing.T) {
	dcdr := &dbapi.MySQL{
		ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{dbapi.DCDREnabledAnnotation: "true"}},
		Spec:       dbapi.MySQLSpec{Distributed: true},
	}
	plain := &dbapi.MySQL{}
	req := func(t MySQLOpsRequestType, mut ...func(*MySQLOpsRequestSpec)) *MySQLOpsRequest {
		r := &MySQLOpsRequest{Spec: MySQLOpsRequestSpec{Type: t}}
		for _, m := range mut {
			m(&r.Spec)
		}
		return r
	}
	tests := []struct {
		name    string
		db      *dbapi.MySQL
		req     *MySQLOpsRequest
		refused bool
		substr  string
	}{
		{"restart allowed", dcdr, req(MySQLOpsRequestTypeRestart), false, ""},
		{"horizontal scaling allowed", dcdr, req(MySQLOpsRequestTypeHorizontalScaling), false, ""},
		{"not dc-dr: anything goes", plain, req(MySQLOpsRequestTypeStorageMigration), false, ""},
		{"distributed without the annotation is not dc-dr", &dbapi.MySQL{Spec: dbapi.MySQLSpec{Distributed: true}}, req(MySQLOpsRequestTypeArchiverRestore), false, ""},
		{"storage migration never", dcdr, req(MySQLOpsRequestTypeStorageMigration), true, "spoke clusters"},
		{"archiver restore never", dcdr, req(MySQLOpsRequestTypeArchiverRestore), true, "new DC-DR MySQL"},
		{"replication mode transformation never", dcdr, req(MySQLOpsRequestTypeReplicationModeTransformation), true, "single-primary"},
		{"vertical scaling in place", dcdr, req(MySQLOpsRequestTypeVerticalScaling, func(s *MySQLOpsRequestSpec) {
			s.VerticalScaling = &MySQLVerticalScalingSpec{Mode: VerticalScalingModeInPlace}
		}), true, "mode Restart"},
		{"volume expansion offline", dcdr, req(MySQLOpsRequestTypeVolumeExpansion, func(s *MySQLOpsRequestSpec) {
			s.VolumeExpansion = &MySQLVolumeExpansionSpec{Mode: VolumeExpansionModeOffline}
		}), true, "mode Online"},
		{"vertical scaling not yet", dcdr, req(MySQLOpsRequestTypeVerticalScaling), true, "not supported for a DC-DR MySQL yet"},
		{"rotate auth not yet", dcdr, req(MySQLOpsRequestTypeRotateAuth), true, "supported: HorizontalScaling, Restart"},
		{"nil request", dcdr, nil, false, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := MySQLDCDROpsRefusal(tc.db, tc.req)
			if (got != "") != tc.refused {
				t.Fatalf("refused=%v, want %v (%q)", got != "", tc.refused, got)
			}
			if tc.substr != "" && !strings.Contains(got, tc.substr) {
				t.Fatalf("%q does not mention %q", got, tc.substr)
			}
		})
	}
}
