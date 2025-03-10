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

package utils

import (
	"context"
	"os"

	core "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ReadinessGateType = "kubedb.com/conversion"
)

func UpdateReadinessGateCondition(ctx context.Context, kc client.Client) error {
	namespace := os.Getenv("POD_NAMESPACE")
	name := os.Getenv("POD_NAME")
	var pod core.Pod
	err := kc.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, &pod)
	if err != nil {
		return err
	}

	foundCondition := false
	for i := range pod.Status.Conditions {
		if pod.Status.Conditions[i].Type == ReadinessGateType {
			pod.Status.Conditions[i].Status = core.ConditionTrue
			foundCondition = true
			break
		}
	}

	if !foundCondition { // Add a new condition if not found
		pod.Status.Conditions = append(pod.Status.Conditions, core.PodCondition{
			Type:   ReadinessGateType,
			Status: core.ConditionTrue,
		})
	}

	err = kc.Status().Update(context.TODO(), &pod)
	if err != nil {
		return err
	}

	klog.Infoln("Successfully updated the readiness gate condition to True")
	return nil
}
