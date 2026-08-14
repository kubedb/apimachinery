/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package appbinding

import (
	"testing"

	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestOptionsEnsureAppBindingName(t *testing.T) {
	tests := []struct {
		name           string
		appBindingName string
		expectedName   string
	}{
		{
			name:         "default name",
			expectedName: "rabbit",
		},
		{
			name:           "overridden name",
			appBindingName: "rabbit-management",
			expectedName:   "rabbit-management",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &dbapi.RabbitMQ{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "rabbit",
					Namespace: "demo",
				},
			}

			got := (Options{
				DB:   db,
				Name: tt.appBindingName,
			}).appBindingName()
			if got != tt.expectedName {
				t.Fatalf("expected AppBinding name %q, got %q", tt.expectedName, got)
			}
		})
	}
}
