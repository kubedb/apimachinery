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

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"
)

func makePostgresVersionWithImage(name, baseOS, image string) *catalog.PostgresVersion {
	pv := makePostgresVersion(name, baseOS)
	pv.Spec.DB.Image = image
	return pv
}

func TestValidatePostgresUpdateVersionOpsRequest_CapabilitiesPreserved(t *testing.T) {
	phase := opsapi.OpsRequestPhaseProgressing

	const (
		extAlpine   = "appscode-images/postgres-enterprise:16.10-alpine-ext"
		plainAlpine = "appscode-images/postgres-enterprise:16.10-alpine"
	)

	tests := []struct {
		name          string
		currentVer    string
		targetVer     string
		versions      map[string]*catalog.PostgresVersion
		wantErrSubstr string
	}{
		{
			name:       "extension image to extension image - allowed",
			currentVer: "16.10-ext",
			targetVer:  "16.15-ext",
			versions: map[string]*catalog.PostgresVersion{
				"16.10-ext": makePostgresVersionWithImage("16.10-ext", "alpine", extAlpine),
				"16.15-ext": makePostgresVersionWithImage("16.15-ext", "alpine", "appscode-images/postgres-enterprise:16.15-alpine-ext"),
			},
		},
		{
			name:       "plain image to plain image - allowed",
			currentVer: "16.10",
			targetVer:  "16.15",
			versions: map[string]*catalog.PostgresVersion{
				"16.10": makePostgresVersionWithImage("16.10", "alpine", plainAlpine),
				"16.15": makePostgresVersionWithImage("16.15", "alpine", "appscode-images/postgres-enterprise:16.15-alpine"),
			},
		},
		{
			name:       "plain image to extension image - allowed (adds extensions)",
			currentVer: "16.10",
			targetVer:  "16.10-ext",
			versions: map[string]*catalog.PostgresVersion{
				"16.10":     makePostgresVersionWithImage("16.10", "alpine", plainAlpine),
				"16.10-ext": makePostgresVersionWithImage("16.10-ext", "alpine", extAlpine),
			},
		},
		{
			name:          "extension image to plain image of same version - blocked",
			currentVer:    "16.10-ext",
			targetVer:     "16.10",
			wantErrSubstr: "bundles PostgreSQL extensions but",
			versions: map[string]*catalog.PostgresVersion{
				"16.10-ext": makePostgresVersionWithImage("16.10-ext", "alpine", extAlpine),
				"16.10":     makePostgresVersionWithImage("16.10", "alpine", plainAlpine),
			},
		},
		{
			name:          "extension image to plain image of newer version - blocked",
			currentVer:    "16.10-ext",
			targetVer:     "16.15",
			wantErrSubstr: "bundles PostgreSQL extensions but",
			versions: map[string]*catalog.PostgresVersion{
				"16.10-ext": makePostgresVersionWithImage("16.10-ext", "alpine", extAlpine),
				"16.15":     makePostgresVersionWithImage("16.15", "alpine", "appscode-images/postgres-enterprise:16.15-alpine"),
			},
		},
		{
			name:          "digest-pinned extension image is still recognised - blocked",
			currentVer:    "16.10-ext",
			targetVer:     "16.10",
			wantErrSubstr: "bundles PostgreSQL extensions but",
			versions: map[string]*catalog.PostgresVersion{
				"16.10-ext": makePostgresVersionWithImage("16.10-ext", "alpine", extAlpine+"@sha256:abc123"),
				"16.10":     makePostgresVersionWithImage("16.10", "alpine", plainAlpine),
			},
		},
		{
			name:          "pg_tde source to target without pg_tde - blocked",
			currentVer:    "17.9-percona",
			targetVer:     "17.9-official",
			wantErrSubstr: "supports pg_tde but",
			versions: map[string]*catalog.PostgresVersion{
				"17.9-percona":  withTDE(makePostgresVersionWithImage("17.9-percona", "bookworm", "appscode-images/postgres:17.9-percona")),
				"17.9-official": makePostgresVersionWithImage("17.9-official", "bookworm", "appscode-images/postgres:17.9-bookworm"),
			},
		},
		{
			name:       "pg_tde source to target with pg_tde - allowed",
			currentVer: "17.9-percona",
			targetVer:  "17.10-percona",
			versions: map[string]*catalog.PostgresVersion{
				"17.9-percona":  withTDE(makePostgresVersionWithImage("17.9-percona", "bookworm", "appscode-images/postgres:17.9-percona")),
				"17.10-percona": withTDE(makePostgresVersionWithImage("17.10-percona", "bookworm", "appscode-images/postgres:17.10-percona")),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wh := &PostgresOpsRequestCustomWebhook{
				DefaultClient: &postgresVersionClient{versions: tt.versions},
			}
			db := makePostgresDB(tt.currentVer)
			req := makeUpdateVersionOpsReq(tt.targetVer, phase)

			err := wh.validatePostgresUpdateVersionOpsRequest(db, req)
			if tt.wantErrSubstr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q but got nil", tt.wantErrSubstr)
				}
				if !strings.Contains(err.Error(), tt.wantErrSubstr) {
					t.Fatalf("expected error containing %q but got: %v", tt.wantErrSubstr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error but got: %v", err)
			}
		})
	}
}

func withTDE(pv *catalog.PostgresVersion) *catalog.PostgresVersion {
	pv.Spec.TDE = &catalog.PostgresVersionTDE{}
	return pv
}

func withExtensions(pv *catalog.PostgresVersion, exts ...string) *catalog.PostgresVersion {
	pv.Spec.DB.Extensions = exts
	return pv
}

func TestValidatePostgresUpdateVersionOpsRequest_DeclaredExtensions(t *testing.T) {
	phase := opsapi.OpsRequestPhaseProgressing

	tests := []struct {
		name          string
		currentVer    string
		targetVer     string
		versions      map[string]*catalog.PostgresVersion
		wantErrSubstr string
	}{
		{
			name:       "target declares a superset of the source extensions - allowed",
			currentVer: "src",
			targetVer:  "dst",
			versions: map[string]*catalog.PostgresVersion{
				"src": withExtensions(makePostgresVersionWithImage("src", "alpine", "img:16"), "postgis", "pg_cron"),
				"dst": withExtensions(makePostgresVersionWithImage("dst", "alpine", "img:17"), "postgis", "pg_cron", "vector"),
			},
		},
		{
			name:          "target is missing one source extension - blocked and named",
			currentVer:    "src",
			targetVer:     "dst",
			wantErrSubstr: "[pg_cron]",
			versions: map[string]*catalog.PostgresVersion{
				"src": withExtensions(makePostgresVersionWithImage("src", "alpine", "img:16"), "postgis", "pg_cron"),
				"dst": withExtensions(makePostgresVersionWithImage("dst", "alpine", "img:17"), "postgis"),
			},
		},
		{
			name:          "extension image onto plain image that declares nothing - blocked (timescaledb labelled Official)",
			currentVer:    "ts",
			targetVer:     "plain",
			wantErrSubstr: "[timescaledb]",
			versions: map[string]*catalog.PostgresVersion{
				"ts":    withExtensions(makePostgresVersionWithImage("ts", "alpine", "timescale/timescaledb:2.14.2-pg15-oss"), "timescaledb"),
				"plain": makePostgresVersionWithImage("plain", "alpine", "appscode-images/postgres:16.4-alpine"),
			},
		},
		{
			name:          "declared list wins over image tag - plain-looking tag but extensions declared",
			currentVer:    "src",
			targetVer:     "dst",
			wantErrSubstr: "[postgis]",
			versions: map[string]*catalog.PostgresVersion{
				"src": withExtensions(makePostgresVersionWithImage("src", "alpine", "img:16"), "postgis"),
				"dst": makePostgresVersionWithImage("dst", "alpine", "img:17"),
			},
		},
		{
			name:       "source declares nothing and target declares extensions - allowed",
			currentVer: "plain",
			targetVer:  "ext",
			versions: map[string]*catalog.PostgresVersion{
				"plain": makePostgresVersionWithImage("plain", "alpine", "img:16"),
				"ext":   withExtensions(makePostgresVersionWithImage("ext", "alpine", "img:17"), "postgis"),
			},
		},
		{
			name:          "fallback: -ext tag source onto plain target, nothing declared - blocked",
			currentVer:    "src",
			targetVer:     "dst",
			wantErrSubstr: "bundles PostgreSQL extensions but",
			versions: map[string]*catalog.PostgresVersion{
				"src": makePostgresVersionWithImage("src", "alpine", "img:16-alpine-ext"),
				"dst": makePostgresVersionWithImage("dst", "alpine", "img:17-alpine"),
			},
		},
		{
			name:       "fallback: -ext tag source onto target that declares extensions - allowed",
			currentVer: "src",
			targetVer:  "dst",
			versions: map[string]*catalog.PostgresVersion{
				"src": makePostgresVersionWithImage("src", "alpine", "img:16-alpine-ext"),
				"dst": withExtensions(makePostgresVersionWithImage("dst", "alpine", "img:17-alpine"), "postgis"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wh := &PostgresOpsRequestCustomWebhook{DefaultClient: &postgresVersionClient{versions: tt.versions}}
			err := wh.validatePostgresUpdateVersionOpsRequest(makePostgresDB(tt.currentVer), makeUpdateVersionOpsReq(tt.targetVer, phase))
			if tt.wantErrSubstr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q but got nil", tt.wantErrSubstr)
				}
				if !strings.Contains(err.Error(), tt.wantErrSubstr) {
					t.Fatalf("expected error containing %q but got: %v", tt.wantErrSubstr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error but got: %v", err)
			}
		})
	}
}
