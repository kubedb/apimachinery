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

	autoscalingapi "kubedb.dev/apimachinery/apis/autoscaling/v1alpha1"
	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
)

func TestYugabyteDBAutoscalerTopology(t *testing.T) {
	plain := &olddbapi.YugabyteDB{}
	withRR := &olddbapi.YugabyteDB{Spec: olddbapi.YugabyteDBSpec{Topology: &olddbapi.YugabyteDBTopology{
		ReadReplica: &olddbapi.YugabyteDBReadReplica{YugabyteDBNode: olddbapi.YugabyteDBNode{}},
	}}}
	ephemeralMaster := &olddbapi.YugabyteDB{Spec: olddbapi.YugabyteDBSpec{Topology: &olddbapi.YugabyteDBTopology{
		Master: &olddbapi.YugabyteDBNode{StorageType: olddbapi.StorageTypeEphemeral},
	}}}
	storage := &autoscalingapi.StorageAutoscalerSpec{}
	compute := &autoscalingapi.ComputeAutoscalerSpec{}
	cases := []struct {
		name string
		db   *olddbapi.YugabyteDB
		spec autoscalingapi.YugabyteDBAutoscalerSpec
		want string
	}{
		{"tserver compute and storage", plain, autoscalingapi.YugabyteDBAutoscalerSpec{
			Compute: &autoscalingapi.YugabyteDBComputeAutoscalerSpec{TServer: compute},
			Storage: &autoscalingapi.YugabyteDBStorageAutoscalerSpec{TServer: storage},
		}, ""},
		{"read replica compute without read replicas", plain, autoscalingapi.YugabyteDBAutoscalerSpec{
			Compute: &autoscalingapi.YugabyteDBComputeAutoscalerSpec{ReadReplica: compute},
		}, "spec.compute.readReplica"},
		{"read replica storage without read replicas", plain, autoscalingapi.YugabyteDBAutoscalerSpec{
			Storage: &autoscalingapi.YugabyteDBStorageAutoscalerSpec{ReadReplica: storage},
		}, "spec.storage.readReplica"},
		{"read replica with read replicas", withRR, autoscalingapi.YugabyteDBAutoscalerSpec{
			Compute: &autoscalingapi.YugabyteDBComputeAutoscalerSpec{ReadReplica: compute},
			Storage: &autoscalingapi.YugabyteDBStorageAutoscalerSpec{ReadReplica: storage},
		}, ""},
		{"storage on ephemeral master", ephemeralMaster, autoscalingapi.YugabyteDBAutoscalerSpec{
			Storage: &autoscalingapi.YugabyteDBStorageAutoscalerSpec{Master: storage},
		}, "ephemeral storage"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := validateYugabyteDBAutoscalerTopology(&autoscalingapi.YugabyteDBAutoscaler{Spec: c.spec}, c.db)
			if c.want == "" && err != nil || c.want != "" && (err == nil || !strings.Contains(err.Error(), c.want)) {
				t.Fatalf("got %v, want %q", err, c.want)
			}
		})
	}
}
