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

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMilvusUpdateVersionGPUError(t *testing.T) {
	gpuDB := &dbapi.Milvus{ObjectMeta: metav1.ObjectMeta{Name: "mv", Namespace: "demo"}}
	gpuDB.Spec.GPU = &dbapi.MilvusGPUSpec{Count: 1}
	cpuDB := &dbapi.Milvus{ObjectMeta: metav1.ObjectMeta{Name: "mv", Namespace: "demo"}}

	gpuBuild := &catalog.MilvusVersion{ObjectMeta: metav1.ObjectMeta{Name: "2.6.15-gpu"}}
	gpuBuild.Spec.DB.GPU = &catalog.MilvusVersionGPUSpec{Supported: true}
	cpuBuild := &catalog.MilvusVersion{ObjectMeta: metav1.ObjectMeta{Name: "3.0.0"}}

	tests := []struct {
		name    string
		db      *dbapi.Milvus
		target  *catalog.MilvusVersion
		wantErr bool
	}{
		{"gpu db to gpu build", gpuDB, gpuBuild, false},
		{"gpu db to cpu build", gpuDB, cpuBuild, true},
		{"cpu db to cpu build", cpuDB, cpuBuild, false},
		{"cpu db to gpu build", cpuDB, gpuBuild, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := milvusUpdateVersionGPUError(tt.db, tt.target)
			if (err != nil) != tt.wantErr {
				t.Fatalf("milvusUpdateVersionGPUError() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
