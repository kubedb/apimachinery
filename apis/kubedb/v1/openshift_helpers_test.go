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

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	"kubedb.dev/apimachinery/apis/kubedb"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	core_util "kmodules.xyz/client-go/core/v1"
	ofstv2 "kmodules.xyz/offshoot-api/api/v2"
)

func nsWith(annotations map[string]string) *core.Namespace {
	return &core.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "demo", Annotations: annotations}}
}

func TestGetNamespaceSecurityIDs(t *testing.T) {
	cases := []struct {
		name    string
		ns      *core.Namespace
		want    *NamespaceSecurityIDs
		wantErr bool
	}{
		{name: "nil namespace", ns: nil, want: nil},
		{name: "no annotations", ns: nsWith(nil), want: nil},
		{
			name: "only supplemental groups is ignored",
			ns:   nsWith(map[string]string{OpenShiftSCCSupplementalGroupsAnnotation: "2000/10"}),
			want: nil,
		},
		{
			name: "uid-range only: fsGroup falls back to uid",
			ns:   nsWith(map[string]string{OpenShiftSCCUIDRangeAnnotation: "1000750000/10000"}),
			want: &NamespaceSecurityIDs{RunAsUser: 1000750000, RunAsGroup: 1000750000, FSGroup: 1000750000},
		},
		{
			name: "split: fsGroup from supplemental groups",
			ns: nsWith(map[string]string{
				OpenShiftSCCUIDRangeAnnotation:           "1000750000/10000",
				OpenShiftSCCSupplementalGroupsAnnotation: "1000760000/10000",
			}),
			want: &NamespaceSecurityIDs{RunAsUser: 1000750000, RunAsGroup: 1000750000, FSGroup: 1000760000},
		},
		{
			name: "start-end form and multiple blocks",
			ns: nsWith(map[string]string{
				OpenShiftSCCUIDRangeAnnotation:           "1000750000-1000759999",
				OpenShiftSCCSupplementalGroupsAnnotation: " 5000/10, 6000/10",
			}),
			want: &NamespaceSecurityIDs{RunAsUser: 1000750000, RunAsGroup: 1000750000, FSGroup: 5000},
		},
		{name: "garbage", ns: nsWith(map[string]string{OpenShiftSCCUIDRangeAnnotation: "abc"}), wantErr: true},
		{name: "root start", ns: nsWith(map[string]string{OpenShiftSCCUIDRangeAnnotation: "0/10000"}), wantErr: true},
		{name: "zero size", ns: nsWith(map[string]string{OpenShiftSCCUIDRangeAnnotation: "1000/0"}), wantErr: true},
		{name: "end before start", ns: nsWith(map[string]string{OpenShiftSCCUIDRangeAnnotation: "2000-1000"}), wantErr: true},
		{
			name: "bad supplemental groups",
			ns: nsWith(map[string]string{
				OpenShiftSCCUIDRangeAnnotation:           "1000/10",
				OpenShiftSCCSupplementalGroupsAnnotation: "x/10",
			}),
			wantErr: true,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := GetNamespaceSecurityIDs(c.ns)
			if (err != nil) != c.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, c.wantErr)
			}
			if c.wantErr {
				return
			}
			if (got == nil) != (c.want == nil) || (got != nil && *got != *c.want) {
				t.Fatalf("got %+v, want %+v", got, c.want)
			}
		})
	}
}

func testPgVersion() *catalog.PostgresVersion {
	return &catalog.PostgresVersion{Spec: catalog.PostgresVersionSpec{
		Version:         "17.2",
		SecurityContext: catalog.PostgresSecurityContext{RunAsUser: ptr.To(int64(70))},
	}}
}

type securityIDsSeen struct {
	podUser, podGroup, podFS, pgUser, pgGroup, coordUser, initUser int64
}

func seen(p *Postgres) securityIDsSeen {
	pod := p.Spec.PodTemplate.Spec.SecurityContext
	pg := core_util.GetContainerByName(p.Spec.PodTemplate.Spec.Containers, kubedb.PostgresContainerName).SecurityContext
	coord := core_util.GetContainerByName(p.Spec.PodTemplate.Spec.Containers, kubedb.PostgresCoordinatorContainerName).SecurityContext
	init := core_util.GetContainerByName(p.Spec.PodTemplate.Spec.InitContainers, kubedb.PostgresInitContainerName).SecurityContext
	return securityIDsSeen{
		*pod.RunAsUser, *pod.RunAsGroup, *pod.FSGroup,
		*pg.RunAsUser, *pg.RunAsGroup, *coord.RunAsUser, *init.RunAsUser,
	}
}

func TestSetDefaultsWithSecurityIDs(t *testing.T) {
	nsIDs := &NamespaceSecurityIDs{RunAsUser: 1001880000, RunAsGroup: 1001880000, FSGroup: 1001890000}

	t.Run("no namespace IDs keeps the catalog UID", func(t *testing.T) {
		p := &Postgres{Spec: PostgresSpec{Replicas: ptr.To(int32(3))}}
		p.SetDefaults(testPgVersion())
		want := securityIDsSeen{70, 70, 70, 70, 70, 70, 70}
		if got := seen(p); got != want {
			t.Fatalf("got %+v, want %+v", got, want)
		}
	})

	t.Run("namespace IDs replace the catalog UID", func(t *testing.T) {
		p := &Postgres{Spec: PostgresSpec{Replicas: ptr.To(int32(3))}}
		p.SetDefaultsWithSecurityIDs(testPgVersion(), nsIDs)
		want := securityIDsSeen{1001880000, 1001880000, 1001890000, 1001880000, 1001880000, 1001880000, 1001880000}
		if got := seen(p); got != want {
			t.Fatalf("got %+v, want %+v", got, want)
		}
	})

	t.Run("user values win over namespace IDs", func(t *testing.T) {
		p := &Postgres{Spec: PostgresSpec{
			Replicas: ptr.To(int32(3)),
			PodTemplate: ofstv2.PodTemplateSpec{Spec: ofstv2.PodSpec{
				SecurityContext: &core.PodSecurityContext{FSGroup: ptr.To(int64(4242))},
				Containers: []core.Container{{
					Name:            kubedb.PostgresContainerName,
					SecurityContext: &core.SecurityContext{RunAsUser: ptr.To(int64(4343))},
				}},
			}},
		}}
		p.SetDefaultsWithSecurityIDs(testPgVersion(), nsIDs)
		want := securityIDsSeen{1001880000, 1001880000, 4242, 4343, 1001880000, 1001880000, 1001880000}
		if got := seen(p); got != want {
			t.Fatalf("got %+v, want %+v", got, want)
		}
	})
}
