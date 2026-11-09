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
	"context"
	"errors"
	"testing"

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	"kubedb.dev/apimachinery/apis/kubedb"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	core_util "kmodules.xyz/client-go/core/v1"
	ofstv2 "kmodules.xyz/offshoot-api/api/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

func defaultPostgres(p *Postgres, pgVersion *catalog.PostgresVersion, nsIDs *NamespaceSecurityIDs) {
	SetDefaultsWithNamespaceSecurityIDs(&p.Spec, nsIDs, func() { p.SetDefaults(pgVersion) })
}

func TestSetDefaultsWithNamespaceSecurityIDs(t *testing.T) {
	nsIDs := &NamespaceSecurityIDs{RunAsUser: 1001880000, RunAsGroup: 1001880000, FSGroup: 1001890000}

	t.Run("no namespace IDs keeps the catalog UID", func(t *testing.T) {
		p := &Postgres{Spec: PostgresSpec{Replicas: ptr.To(int32(3))}}
		defaultPostgres(p, testPgVersion(), nil)
		want := securityIDsSeen{70, 70, 70, 70, 70, 70, 70}
		if got := seen(p); got != want {
			t.Fatalf("got %+v, want %+v", got, want)
		}
	})

	t.Run("namespace IDs replace the catalog UID", func(t *testing.T) {
		pgVersion := testPgVersion()
		p := &Postgres{Spec: PostgresSpec{Replicas: ptr.To(int32(3))}}
		defaultPostgres(p, pgVersion, nsIDs)
		want := securityIDsSeen{1001880000, 1001880000, 1001890000, 1001880000, 1001880000, 1001880000, 1001880000}
		if got := seen(p); got != want {
			t.Fatalf("got %+v, want %+v", got, want)
		}
		if got := *pgVersion.Spec.SecurityContext.RunAsUser; got != 70 {
			t.Fatalf("catalog object was modified: runAsUser = %d", got)
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
		defaultPostgres(p, testPgVersion(), nsIDs)
		want := securityIDsSeen{1001880000, 1001880000, 4242, 4343, 1001880000, 1001880000, 1001880000}
		if got := seen(p); got != want {
			t.Fatalf("got %+v, want %+v", got, want)
		}
	})
}

func TestSetDefaultsWithNamespaceSecurityIDsGeneric(t *testing.T) {
	type sidecarSpec struct {
		PodTemplate ofstv2.PodTemplateSpec
		Exporter    *core.SecurityContext
		Raw         []byte
	}
	type spec struct {
		Nodes []sidecarSpec
	}
	catalogUID := ptr.To(int64(999))
	nsIDs := &NamespaceSecurityIDs{RunAsUser: 5000, RunAsGroup: 5000, FSGroup: 6000}

	s := &spec{Nodes: []sidecarSpec{{
		PodTemplate: ofstv2.PodTemplateSpec{Spec: ofstv2.PodSpec{
			Containers: []core.Container{{Name: "user", SecurityContext: &core.SecurityContext{RunAsGroup: ptr.To(int64(7))}}},
		}},
		Raw: []byte("x"),
	}}}
	setDefaults := func() {
		n := &s.Nodes[0]
		n.PodTemplate.Spec.SecurityContext = &core.PodSecurityContext{FSGroup: catalogUID}
		n.PodTemplate.Spec.InitContainers = []core.Container{
			{Name: "sysctl", SecurityContext: &core.SecurityContext{Privileged: ptr.To(true), RunAsUser: catalogUID}},
			{Name: "root", SecurityContext: &core.SecurityContext{RunAsUser: ptr.To(int64(0))}},
		}
		// a new container inserted before the user's one must not shift the user's values
		n.PodTemplate.Spec.Containers = append([]core.Container{{Name: "db", SecurityContext: &core.SecurityContext{RunAsUser: catalogUID, RunAsGroup: catalogUID}}}, n.PodTemplate.Spec.Containers...)
		n.PodTemplate.Spec.Containers[1].SecurityContext.RunAsUser = catalogUID
		n.Exporter = &core.SecurityContext{RunAsUser: catalogUID}
	}
	SetDefaultsWithNamespaceSecurityIDs(s, nsIDs, setDefaults)

	n := s.Nodes[0]
	check := func(what string, got *int64, want int64) {
		t.Helper()
		if got == nil || *got != want {
			t.Errorf("%s = %v, want %d", what, got, want)
		}
	}
	check("pod fsGroup", n.PodTemplate.Spec.SecurityContext.FSGroup, 6000)
	check("privileged init runAsUser", n.PodTemplate.Spec.InitContainers[0].SecurityContext.RunAsUser, 999)
	check("root init runAsUser", n.PodTemplate.Spec.InitContainers[1].SecurityContext.RunAsUser, 0)
	check("db runAsUser", n.PodTemplate.Spec.Containers[0].SecurityContext.RunAsUser, 5000)
	check("db runAsGroup", n.PodTemplate.Spec.Containers[0].SecurityContext.RunAsGroup, 5000)
	check("user runAsUser", n.PodTemplate.Spec.Containers[1].SecurityContext.RunAsUser, 5000)
	check("user runAsGroup", n.PodTemplate.Spec.Containers[1].SecurityContext.RunAsGroup, 7)
	check("exporter runAsUser", n.Exporter.RunAsUser, 5000)
	if n.Exporter.RunAsGroup != nil {
		t.Errorf("exporter runAsGroup = %d, want unset", *n.Exporter.RunAsGroup)
	}
	check("catalog UID", catalogUID, 999)
}

// namespaceReader serves one Namespace and counts the reads. Any other call panics on the nil
// embedded Reader.
type namespaceReader struct {
	client.Reader
	ns    *core.Namespace
	reads int
}

func (r *namespaceReader) Get(_ context.Context, key client.ObjectKey, obj client.Object, _ ...client.GetOption) error {
	r.reads++
	ns, ok := obj.(*core.Namespace)
	if !ok || key.Name != r.ns.Name {
		return errors.New("unexpected Get")
	}
	*ns = *r.ns.DeepCopy()
	return nil
}

func TestSetDefaultsInNamespaceForVersion(t *testing.T) {
	ns := nsWith(map[string]string{
		OpenShiftSCCUIDRangeAnnotation:           "1000650000/10000",
		OpenShiftSCCSupplementalGroupsAnnotation: "1000650000/10000",
	})

	t.Run("version pins the UID: catalog kept, namespace not read", func(t *testing.T) {
		kc := &namespaceReader{ns: ns}
		p := &Postgres{Spec: PostgresSpec{Replicas: ptr.To(int32(3))}}
		err := SetDefaultsInNamespaceForVersion(context.TODO(), kc, "demo", false, &p.Spec, func() error { p.SetDefaults(testPgVersion()); return nil })
		if err != nil {
			t.Fatal(err)
		}
		if want := (securityIDsSeen{70, 70, 70, 70, 70, 70, 70}); seen(p) != want {
			t.Fatalf("got %+v, want %+v", seen(p), want)
		}
		if kc.reads != 0 {
			t.Fatalf("namespace read %d times, want 0", kc.reads)
		}
	})

	t.Run("any non-root UID allowed: namespace IDs applied", func(t *testing.T) {
		kc := &namespaceReader{ns: ns}
		p := &Postgres{Spec: PostgresSpec{Replicas: ptr.To(int32(3))}}
		err := SetDefaultsInNamespaceForVersion(context.TODO(), kc, "demo", true, &p.Spec, func() error { p.SetDefaults(testPgVersion()); return nil })
		if err != nil {
			t.Fatal(err)
		}
		id := int64(1000650000)
		if want := (securityIDsSeen{id, id, id, id, id, id, id}); seen(p) != want {
			t.Fatalf("got %+v, want %+v", seen(p), want)
		}
	})

	t.Run("setDefaults error is returned", func(t *testing.T) {
		boom := errors.New("boom")
		for _, anyUID := range []bool{false, true} {
			err := SetDefaultsInNamespaceForVersion(context.TODO(), &namespaceReader{ns: ns}, "demo", anyUID, &PostgresSpec{}, func() error { return boom })
			if !errors.Is(err, boom) {
				t.Fatalf("runAsAnyNonRoot=%v: err = %v, want %v", anyUID, err, boom)
			}
		}
	})
}
