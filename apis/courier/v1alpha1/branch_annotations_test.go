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

	"kubedb.dev/apimachinery/apis/kubedb"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func obj(ann map[string]string) metav1.Object {
	return &metav1.ObjectMeta{Annotations: ann}
}

func TestIsBranched(t *testing.T) {
	cases := []struct {
		name string
		obj  metav1.Object
		want bool
	}{
		{"nil object", nil, false},
		{"nil annotations", &metav1.ObjectMeta{}, false},
		{"no branched annotation", obj(map[string]string{"other": "x"}), false},
		{"branched present", obj(map[string]string{kubedb.BranchedFromAnnotation: `{"source":"demo/prod-pg"}`}), true},
		{"branched present empty value", obj(map[string]string{kubedb.BranchedFromAnnotation: ""}), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsBranched(tc.obj); got != tc.want {
				t.Fatalf("IsBranched=%v, want %v", got, tc.want)
			}
		})
	}
}

func TestParseBranchedFrom(t *testing.T) {
	t.Run("absent", func(t *testing.T) {
		bf, ok, err := ParseBranchedFrom(obj(nil))
		if ok || err != nil || bf != (BranchedFrom{}) {
			t.Fatalf("absent: got bf=%+v ok=%v err=%v", bf, ok, err)
		}
	})
	t.Run("valid", func(t *testing.T) {
		bf, ok, err := ParseBranchedFrom(obj(map[string]string{
			kubedb.BranchedFromAnnotation: `{"cluster":"prod-east","source":"demo/prod-pg"}`,
		}))
		if err != nil || !ok {
			t.Fatalf("valid: ok=%v err=%v", ok, err)
		}
		if bf.Cluster != "prod-east" || bf.Source != "demo/prod-pg" {
			t.Fatalf("valid: bf=%+v", bf)
		}
	})
	t.Run("present but malformed", func(t *testing.T) {
		_, ok, err := ParseBranchedFrom(obj(map[string]string{kubedb.BranchedFromAnnotation: "not-json"}))
		if !ok || err == nil {
			t.Fatalf("malformed: want ok=true err!=nil, got ok=%v err=%v", ok, err)
		}
	})
}
