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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ybVersion(name string, ysql int32, allow, deny []string) *YugabyteDBVersion {
	return &YugabyteDBVersion{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: YugabyteDBVersionSpec{
			Version:           name,
			YSQLMajorVersion:  ysql,
			UpdateConstraints: UpdateConstraints{Allowlist: allow, Denylist: deny},
		},
	}
}

func TestYugabyteDBVersionCheckUpdateTo(t *testing.T) {
	cases := []struct {
		name    string
		cur     *YugabyteDBVersion
		target  *YugabyteDBVersion
		wantErr string
	}{
		{"allowed by constraint", ybVersion("2025.1.4", 15, []string{"> 2025.1.4"}, nil), ybVersion("2025.2.8", 15, nil, nil), ""},
		{"allowed by name", ybVersion("2025.1.4", 15, []string{"2025.2.8"}, nil), ybVersion("2025.2.8", 15, nil, nil), ""},
		{"empty allowlist", ybVersion("2025.1.4", 15, nil, nil), ybVersion("2025.2.8", 15, nil, nil), ""},
		{"same version", ybVersion("2025.2.8", 15, nil, nil), ybVersion("2025.2.8", 15, nil, nil), "is the current version"},
		{"downgrade, empty allowlist", ybVersion("2025.2.8", 15, nil, nil), ybVersion("2025.1.4", 15, nil, nil), "downgrades are not supported"},
		// the constraints of an old catalog read the other way round: they must not open a downgrade
		{"downgrade, old-style allowlist", ybVersion("2025.2.8", 15, []string{">= 2025.1.4, < 2025.2.8"}, nil), ybVersion("2025.1.4", 15, nil, nil), "downgrades are not supported"},
		{"denied", ybVersion("2025.1.4", 15, nil, []string{"2025.2.8"}), ybVersion("2025.2.8", 15, nil, nil), "denylist"},
		{"not in allowlist", ybVersion("2025.1.4", 15, []string{"> 2026.1.0"}, nil), ybVersion("2025.2.8", 15, nil, nil), "allowlist [> 2026.1.0]"},
		{"ysql major", ybVersion("2024.2.4", 11, nil, nil), ybVersion("2025.2.8", 15, nil, nil), "YSQL major"},
		{"invalid constraint", ybVersion("2025.1.4", 15, []string{"nope!"}, nil), ybVersion("2025.2.8", 15, nil, nil), "invalid allowlist entry"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.cur.CheckUpdateTo(c.target)
			switch {
			case c.wantErr == "" && err != nil:
				t.Fatalf("unexpected error: %v", err)
			case c.wantErr != "" && (err == nil || !strings.Contains(err.Error(), c.wantErr)):
				t.Fatalf("want error containing %q, got %v", c.wantErr, err)
			}
		})
	}
}
