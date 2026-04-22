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
	"testing"

	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func newMSSQLServerDB(storageSize string) *olddbapi.MSSQLServer {
	q := resource.MustParse(storageSize)
	return &olddbapi.MSSQLServer{
		Spec: olddbapi.MSSQLServerSpec{
			Storage: &core.PersistentVolumeClaimSpec{
				Resources: core.VolumeResourceRequirements{
					Requests: core.ResourceList{
						core.ResourceStorage: q,
					},
				},
			},
		},
	}
}

func newMSSQLServerVolumeExpansionOpsReq(desiredSize string, phase opsapi.OpsRequestPhase) *opsapi.MSSQLServerOpsRequest {
	q := resource.MustParse(desiredSize)
	return &opsapi.MSSQLServerOpsRequest{
		Spec: opsapi.MSSQLServerOpsRequestSpec{
			Type: opsapi.MSSQLServerOpsRequestTypeVolumeExpansion,
			VolumeExpansion: &opsapi.MSSQLServerVolumeExpansionSpec{
				MSSQLServer: &q,
			},
		},
		Status: opsapi.OpsRequestStatus{
			Phase: phase,
		},
	}
}

func TestValidateMSSQLServerVolumeExpansionOpsRequest(t *testing.T) {
	webhook := &MSSQLServerOpsRequestCustomWebhook{}

	tests := []struct {
		name      string
		db        *olddbapi.MSSQLServer
		req       *opsapi.MSSQLServerOpsRequest
		wantError bool
	}{
		{
			name:      "nil volumeExpansion spec",
			db:        newMSSQLServerDB("10Gi"),
			req:       &opsapi.MSSQLServerOpsRequest{},
			wantError: true,
		},
		{
			name: "nil MSSQLServer field",
			db:   newMSSQLServerDB("10Gi"),
			req: &opsapi.MSSQLServerOpsRequest{
				Spec: opsapi.MSSQLServerOpsRequestSpec{
					VolumeExpansion: &opsapi.MSSQLServerVolumeExpansionSpec{},
				},
			},
			wantError: true,
		},
		{
			name:      "desired less than current (pending) - should fail",
			db:        newMSSQLServerDB("10Gi"),
			req:       newMSSQLServerVolumeExpansionOpsReq("5Gi", opsapi.OpsRequestPhasePending),
			wantError: true,
		},
		{
			name:      "desired equal to current (pending) - should fail",
			db:        newMSSQLServerDB("10Gi"),
			req:       newMSSQLServerVolumeExpansionOpsReq("10Gi", opsapi.OpsRequestPhasePending),
			wantError: true,
		},
		{
			name:      "desired less than current (empty phase) - should fail",
			db:        newMSSQLServerDB("10Gi"),
			req:       newMSSQLServerVolumeExpansionOpsReq("5Gi", ""),
			wantError: true,
		},
		{
			name:      "desired equal to current (empty phase) - should fail",
			db:        newMSSQLServerDB("10Gi"),
			req:       newMSSQLServerVolumeExpansionOpsReq("10Gi", ""),
			wantError: true,
		},
		{
			name:      "desired greater than current (pending) - should succeed",
			db:        newMSSQLServerDB("10Gi"),
			req:       newMSSQLServerVolumeExpansionOpsReq("20Gi", opsapi.OpsRequestPhasePending),
			wantError: false,
		},
		{
			name:      "desired greater than current (empty phase) - should succeed",
			db:        newMSSQLServerDB("10Gi"),
			req:       newMSSQLServerVolumeExpansionOpsReq("20Gi", ""),
			wantError: false,
		},
		{
			name:      "desired == current but progressing phase - should succeed",
			db:        newMSSQLServerDB("10Gi"),
			req:       newMSSQLServerVolumeExpansionOpsReq("10Gi", opsapi.OpsRequestPhaseProgressing),
			wantError: false,
		},
		{
			name:      "nil storage on DB - should fail",
			db:        &olddbapi.MSSQLServer{},
			req:       newMSSQLServerVolumeExpansionOpsReq("20Gi", opsapi.OpsRequestPhasePending),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := webhook.validateMSSQLServerVolumeExpansionOpsRequest(tt.db, tt.req)
			if tt.wantError && err == nil {
				t.Errorf("expected error but got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("expected no error but got: %v", err)
			}
		})
	}
}
