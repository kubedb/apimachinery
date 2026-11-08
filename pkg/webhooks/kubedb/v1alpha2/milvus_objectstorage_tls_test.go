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
	"testing"

	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func TestMilvusValidateObjectStorageTLS(t *testing.T) {
	secretSel := &core.SecretKeySelector{LocalObjectReference: core.LocalObjectReference{Name: "ca"}, Key: "ca.crt"}
	cmSel := &core.ConfigMapKeySelector{LocalObjectReference: core.LocalObjectReference{Name: "openshift-service-ca.crt"}, Key: "service-ca.crt"}

	tests := []struct {
		name    string
		tls     *olddbapi.ObjectStorageTLS
		wantErr bool
	}{
		{"secret", &olddbapi.ObjectStorageTLS{CASecret: secretSel}, false},
		{"configmap", &olddbapi.ObjectStorageTLS{CAConfigMap: cmSel}, false},
		{"neither", &olddbapi.ObjectStorageTLS{}, true},
		{"both", &olddbapi.ObjectStorageTLS{CASecret: secretSel, CAConfigMap: cmSel}, true},
		{"secret without key", &olddbapi.ObjectStorageTLS{CASecret: &core.SecretKeySelector{LocalObjectReference: core.LocalObjectReference{Name: "ca"}}}, true},
		{"configmap without name", &olddbapi.ObjectStorageTLS{CAConfigMap: &core.ConfigMapKeySelector{Key: "k"}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := milvusValidateObjectStorageTLS(tt.tls, field.NewPath("spec", "objectStorage", "tls"))
			if (len(errs) > 0) != tt.wantErr {
				t.Fatalf("wantErr=%v, got %v", tt.wantErr, errs)
			}
		})
	}
}
