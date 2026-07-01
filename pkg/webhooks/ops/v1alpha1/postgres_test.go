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
	"context"
	"strings"
	"testing"

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// postgresVersionClient is a minimal client.Client mock for PostgresVersion lookups.
// All methods except Get panic — tests must set phase to Progressing to skip IsUpgradable
// (which also calls List), ensuring only Get is exercised.
type postgresVersionClient struct {
	client.Client
	versions map[string]*catalog.PostgresVersion
}

func (m *postgresVersionClient) Get(_ context.Context, key types.NamespacedName, obj client.Object, _ ...client.GetOption) error {
	pv, ok := obj.(*catalog.PostgresVersion)
	if !ok {
		return nil
	}
	v, found := m.versions[key.Name]
	if !found {
		return apierrors.NewNotFound(schema.GroupResource{Group: catalog.SchemeGroupVersion.Group, Resource: "postgresversions"}, key.Name)
	}
	*pv = *v
	return nil
}

func makePostgresVersion(name, baseOS string) *catalog.PostgresVersion {
	pv := &catalog.PostgresVersion{}
	pv.Name = name
	pv.Spec.DB.BaseOS = baseOS
	pv.Spec.Version = name // simplified; real field would be e.g. "16.8"
	return pv
}

func makePostgresDB(currentVersion string) *dbapi.Postgres {
	db := &dbapi.Postgres{}
	db.Spec.Version = currentVersion
	return db
}

func makeUpdateVersionOpsReq(targetVersion string, phase opsapi.OpsRequestPhase) *opsapi.PostgresOpsRequest {
	return &opsapi.PostgresOpsRequest{
		Spec: opsapi.PostgresOpsRequestSpec{
			Type: opsapi.PostgresOpsRequestTypeUpdateVersion,
			UpdateVersion: &opsapi.PostgresUpdateVersionSpec{
				TargetVersion: targetVersion,
			},
		},
		Status: opsapi.OpsRequestStatus{
			Phase: phase,
		},
	}
}

func TestValidatePostgresUpdateVersionOpsRequest_BaseOSCheck(t *testing.T) {
	// Use Progressing phase throughout to bypass the IsUpgradable semver check
	// (which would also need List). We're testing only the baseOS guard added
	// to validatePostgresUpdateVersionOpsRequest.
	phase := opsapi.OpsRequestPhaseProgressing

	tests := []struct {
		name          string
		currentVer    string
		targetVer     string
		versions      map[string]*catalog.PostgresVersion
		wantErrSubstr string // empty means no error expected
	}{
		{
			name:       "same baseOS alpine to alpine - allowed",
			currentVer: "16.8",
			targetVer:  "17.8",
			versions: map[string]*catalog.PostgresVersion{
				"16.8": makePostgresVersion("16.8", "alpine"),
				"17.8": makePostgresVersion("17.8", "alpine"),
			},
		},
		{
			name:       "same baseOS bookworm to bookworm - allowed",
			currentVer: "16.8-bookworm",
			targetVer:  "17.8-bookworm",
			versions: map[string]*catalog.PostgresVersion{
				"16.8-bookworm": makePostgresVersion("16.8-bookworm", "bookworm"),
				"17.8-bookworm": makePostgresVersion("17.8-bookworm", "bookworm"),
			},
		},
		{
			name:          "alpine to bookworm - blocked",
			currentVer:    "16.8",
			targetVer:     "17.8-bookworm",
			wantErrSubstr: "upgrading between different base OS variants is not allowed",
			versions: map[string]*catalog.PostgresVersion{
				"16.8":          makePostgresVersion("16.8", "alpine"),
				"17.8-bookworm": makePostgresVersion("17.8-bookworm", "bookworm"),
			},
		},
		{
			name:          "bookworm to alpine - blocked",
			currentVer:    "16.8-bookworm",
			targetVer:     "17.8",
			wantErrSubstr: "upgrading between different base OS variants is not allowed",
			versions: map[string]*catalog.PostgresVersion{
				"16.8-bookworm": makePostgresVersion("16.8-bookworm", "bookworm"),
				"17.8":          makePostgresVersion("17.8", "alpine"),
			},
		},
		{
			name:          "same version different baseOS (lateral switch) - blocked",
			currentVer:    "16.8",
			targetVer:     "16.8-bookworm",
			wantErrSubstr: "upgrading between different base OS variants is not allowed",
			versions: map[string]*catalog.PostgresVersion{
				"16.8":          makePostgresVersion("16.8", "alpine"),
				"16.8-bookworm": makePostgresVersion("16.8-bookworm", "bookworm"),
			},
		},
		{
			name:       "empty baseOS on source - check skipped (graceful fallback)",
			currentVer: "16.8",
			targetVer:  "17.8-bookworm",
			versions: map[string]*catalog.PostgresVersion{
				"16.8":          makePostgresVersion("16.8", ""),
				"17.8-bookworm": makePostgresVersion("17.8-bookworm", "bookworm"),
			},
		},
		{
			name:       "empty baseOS on target - check skipped (graceful fallback)",
			currentVer: "16.8",
			targetVer:  "17.8",
			versions: map[string]*catalog.PostgresVersion{
				"16.8": makePostgresVersion("16.8", "alpine"),
				"17.8": makePostgresVersion("17.8", ""),
			},
		},
		{
			name:          "target version not found - error",
			currentVer:    "16.8",
			targetVer:     "99.0",
			wantErrSubstr: "not found",
			versions: map[string]*catalog.PostgresVersion{
				"16.8": makePostgresVersion("16.8", "alpine"),
			},
		},
		{
			name:          "source version not found - error",
			currentVer:    "99.0",
			targetVer:     "17.8",
			wantErrSubstr: "not found",
			versions: map[string]*catalog.PostgresVersion{
				"17.8": makePostgresVersion("17.8", "alpine"),
			},
		},
		{
			name:          "nil updateVersion spec - error",
			currentVer:    "16.8",
			targetVer:     "",
			wantErrSubstr: "nil not supported",
			versions:      map[string]*catalog.PostgresVersion{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wh := &PostgresOpsRequestCustomWebhook{
				DefaultClient: &postgresVersionClient{versions: tt.versions},
			}
			db := makePostgresDB(tt.currentVer)

			var req *opsapi.PostgresOpsRequest
			if tt.targetVer == "" {
				// test nil spec case
				req = &opsapi.PostgresOpsRequest{Status: opsapi.OpsRequestStatus{Phase: phase}}
			} else {
				req = makeUpdateVersionOpsReq(tt.targetVer, phase)
			}

			err := wh.validatePostgresUpdateVersionOpsRequest(db, req)
			if tt.wantErrSubstr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q but got nil", tt.wantErrSubstr)
				}
				if !strings.Contains(err.Error(), tt.wantErrSubstr) {
					t.Fatalf("expected error containing %q but got: %v", tt.wantErrSubstr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error but got: %v", err)
				}
			}
		})
	}
}

