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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestValidatePostgresDCDRArchiver(t *testing.T) {
	pg := func(dcdr, archiver, initArchiver bool) *dbapi.Postgres {
		p := &dbapi.Postgres{}
		p.Spec.Distributed = true
		if dcdr {
			p.Annotations = map[string]string{dbapi.DCDREnabledAnnotation: "true"}
		}
		if archiver {
			p.Spec.Archiver = &dbapi.Archiver{}
		}
		if initArchiver {
			p.Spec.Init = &dbapi.InitSpec{Archiver: &dbapi.ArchiverRecovery{RecoveryTimestamp: metav1.Now()}}
		}
		return p
	}
	cases := []struct {
		name    string
		new     *dbapi.Postgres
		old     *dbapi.Postgres
		wantErr bool
	}{
		{"create dcdr no archiver", pg(true, false, false), nil, false},
		{"create dcdr archiver", pg(true, true, false), nil, true},
		{"create dcdr init archiver", pg(true, false, true), nil, true},
		{"create plain distributed archiver", pg(false, true, false), nil, false},
		{"update adds archiver", pg(true, true, false), pg(true, false, false), true},
		{"update adds init archiver", pg(true, false, true), pg(true, false, false), true},
		{"update enables dcdr with archiver", pg(true, true, false), pg(false, true, false), true},
		{"update keeps existing archiver", pg(true, true, false), pg(true, true, false), false},
		{"update keeps existing init archiver", pg(true, false, true), pg(true, false, true), false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := validatePostgresDCDRArchiver(c.new, c.old)
			if (err != nil) != c.wantErr {
				t.Fatalf("wantErr=%v, got %v", c.wantErr, err)
			}
		})
	}
}
