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
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
)

func IsReplicasReady(stsList []*appsv1.StatefulSet) (bool, string) {
	for _, sts := range stsList {
		rep := int32(1)
		if sts.Spec.Replicas != nil {
			rep = *sts.Spec.Replicas
		}

		if rep > sts.Status.ReadyReplicas {
			return false, fmt.Sprintf("All desired replicas are not ready. For StatefulSet: %s/%s Desired replicas: %d, Ready replicas: %d.", sts.Namespace, sts.Name, rep, sts.Status.ReadyReplicas)
		}
	}

	return true, fmt.Sprint("All desired replicas are ready.")
}
