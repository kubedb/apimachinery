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

package v1alpha2

import (
	"strings"
	"testing"

	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/ptr"
)

func hadrDB2(dbs ...olddbapi.DB2HADRDatabase) *olddbapi.DB2 {
	return &olddbapi.DB2{Spec: olddbapi.DB2Spec{
		Replicas: ptr.To(int32(2)),
		HADR:     &olddbapi.DB2HADRSpec{Databases: dbs, PeerWindowSeconds: 120},
	}}
}

func TestValidateDB2HADRDatabases(t *testing.T) {
	cases := []struct {
		name    string
		db      *olddbapi.DB2
		wantErr string
	}{
		{"two databases, defaulted ports", hadrDB2(olddbapi.DB2HADRDatabase{Name: "TESTDB"}, olddbapi.DB2HADRDatabase{Name: "SALES"}), ""},
		{"lower-case name is normalised", hadrDB2(olddbapi.DB2HADRDatabase{Name: "sales"}), ""},
		{"name too long", hadrDB2(olddbapi.DB2HADRDatabase{Name: "TOOLONGNAME"}), "letter followed by up to 7"},
		{"name starting with a digit", hadrDB2(olddbapi.DB2HADRDatabase{Name: "1DB"}), "letter followed by up to 7"},
		{"dollar sign", hadrDB2(olddbapi.DB2HADRDatabase{Name: "A$B"}), "letter followed by up to 7"},
		{"duplicate name, different case", hadrDB2(olddbapi.DB2HADRDatabase{Name: "SALES"}, olddbapi.DB2HADRDatabase{Name: "sales"}), "Duplicate"},
		{"duplicate port", hadrDB2(olddbapi.DB2HADRDatabase{Name: "A", Port: 55000}, olddbapi.DB2HADRDatabase{Name: "B", Port: 55000}), "needs its own port"},
		{"database port", hadrDB2(olddbapi.DB2HADRDatabase{Name: "A", Port: 50000}), "already used"},
		{"coordinator port", hadrDB2(olddbapi.DB2HADRDatabase{Name: "A", Port: 8080}), "already used"},
		{"port too low", hadrDB2(olddbapi.DB2HADRDatabase{Name: "A", Port: 80}), "between 1024 and 65535"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			errs := validateDB2HADRDatabases(c.db, field.NewPath("spec", "hadr", "databases"))
			got := errs.ToAggregate()
			switch {
			case c.wantErr == "" && got != nil:
				t.Fatalf("unexpected error: %v", got)
			case c.wantErr != "" && (got == nil || !strings.Contains(got.Error(), c.wantErr)):
				t.Fatalf("want error containing %q, got %v", c.wantErr, got)
			}
		})
	}
}

func TestValidateDB2HADRPortsImmutable(t *testing.T) {
	old := hadrDB2(olddbapi.DB2HADRDatabase{Name: "A", Port: 55000}, olddbapi.DB2HADRDatabase{Name: "B", Port: 55001})

	moved := hadrDB2(olddbapi.DB2HADRDatabase{Name: "A", Port: 55000}, olddbapi.DB2HADRDatabase{Name: "B", Port: 55005})
	if errs := validateDB2HADRPortsImmutable(old, moved); len(errs) != 1 || !strings.Contains(errs[0].Error(), "cannot change the HADR port of B") {
		t.Fatalf("moving B's port must be rejected, got %v", errs)
	}

	added := hadrDB2(olddbapi.DB2HADRDatabase{Name: "A", Port: 55000}, olddbapi.DB2HADRDatabase{Name: "B", Port: 55001}, olddbapi.DB2HADRDatabase{Name: "C"})
	if errs := validateDB2HADRPortsImmutable(old, added); len(errs) != 0 {
		t.Fatalf("adding a database must be allowed, got %v", errs)
	}

	removed := hadrDB2(olddbapi.DB2HADRDatabase{Name: "B", Port: 55001})
	if errs := validateDB2HADRPortsImmutable(old, removed); len(errs) != 0 {
		t.Fatalf("removing a database must be allowed, got %v", errs)
	}

	legacy := &olddbapi.DB2{Spec: olddbapi.DB2Spec{Replicas: ptr.To(int32(2)), HADR: &olddbapi.DB2HADRSpec{DatabaseName: "A"}}}
	if errs := validateDB2HADRPortsImmutable(legacy, hadrDB2(olddbapi.DB2HADRDatabase{Name: "A", Port: 55000})); len(errs) != 0 {
		t.Fatalf("migrating databaseName to databases must keep port 55000 valid, got %v", errs)
	}
}

func TestValidateDB2HADRPeerWindowOutlastsTheLease(t *testing.T) {
	cases := []struct {
		name       string
		syncMode   olddbapi.DB2HADRSyncMode
		peerWindow int32
		wantErr    bool
	}{
		{"NEARSYNC at the defaults (timeout 60)", olddbapi.DB2HADRSyncModeNearSync, 120, false},
		{"NEARSYNC just inside timeout + 60", olddbapi.DB2HADRSyncModeNearSync, 120, false},
		{"NEARSYNC below timeout + 60", olddbapi.DB2HADRSyncModeNearSync, 119, true},
		{"SYNC at zero", olddbapi.DB2HADRSyncModeSync, 0, true},
		{"empty sync mode means NEARSYNC", "", 30, true},
		{"peer window no longer than the timeout (the 120/120 trap)", olddbapi.DB2HADRSyncModeNearSync, 120, true},
		{"SUPERASYNC ignores the peer window", olddbapi.DB2HADRSyncModeSuperAsync, 0, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			db := hadrDB2(olddbapi.DB2HADRDatabase{Name: "TESTDB"})
			db.Spec.HADR.SyncMode = c.syncMode
			db.Spec.HADR.PeerWindowSeconds = c.peerWindow
			db.Spec.HADR.TimeoutSeconds = 60
			if strings.Contains(c.name, "120/120") {
				db.Spec.HADR.TimeoutSeconds = 120
			}
			var got []string
			for _, e := range validateDB2HADR(db) {
				if strings.Contains(e.Field, "peerWindowSeconds") {
					got = append(got, e.Error())
				}
			}
			if c.wantErr != (len(got) > 0) {
				t.Fatalf("wantErr=%v, got %v", c.wantErr, got)
			}
		})
	}
}
