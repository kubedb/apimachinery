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

package lib

import (
	"context"
	"encoding/json"

	core "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apitypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

// InPlaceResizePod requests an in-place resize of the running pod's containers through
// the pods/resize subresource. targets maps a container name to its desired resources.
// It does not wait for the resize to be enacted; use IsResizeSettled/IsResizeInfeasible
// against a freshly fetched pod for that.
func InPlaceResizePod(kc kubernetes.Interface, namespace, podName string, targets map[string]core.ResourceRequirements) error {
	type containerPatch struct {
		Name      string                    `json:"name"`
		Resources core.ResourceRequirements `json:"resources"`
	}
	containers := make([]containerPatch, 0, len(targets))
	for name, res := range targets {
		containers = append(containers, containerPatch{Name: name, Resources: res})
	}
	patch := map[string]any{
		"spec": map[string]any{
			"containers": containers,
		},
	}
	data, err := json.Marshal(patch)
	if err != nil {
		return err
	}
	_, err = kc.CoreV1().Pods(namespace).Patch(
		context.TODO(), podName, apitypes.StrategicMergePatchType, data, metav1.PatchOptions{}, "resize")
	return err
}

// ResizeTargets returns the desired resources keyed by container name for the named
// containers found in the given container list (typically a PetSet/StatefulSet template).
// Names that are not present in the list are skipped.
func ResizeTargets(containers []core.Container, names ...string) map[string]core.ResourceRequirements {
	want := make(map[string]struct{}, len(names))
	for _, n := range names {
		want[n] = struct{}{}
	}
	res := make(map[string]core.ResourceRequirements)
	for _, ctr := range containers {
		if _, ok := want[ctr.Name]; ok {
			res[ctr.Name] = ctr.Resources
		}
	}
	return res
}

// IsResizeInfeasible reports whether the kubelet rejected the requested resize as
// impossible (PodResizePending with reason Infeasible), returning the condition message.
func IsResizeInfeasible(pod *core.Pod) (string, bool) {
	for _, cond := range pod.Status.Conditions {
		if cond.Type == core.PodResizePending && cond.Reason == core.PodReasonInfeasible {
			return cond.Message, true
		}
	}
	return "", false
}

// IsResizeSettled reports whether no resize is pending or in progress and every target
// container's node-enacted resources (status.containerStatuses[].resources) equal the
// desired resources.
func IsResizeSettled(pod *core.Pod, targets map[string]core.ResourceRequirements) bool {
	for _, cond := range pod.Status.Conditions {
		if cond.Type == core.PodResizePending || cond.Type == core.PodResizeInProgress {
			return false
		}
	}
	for name, want := range targets {
		cs := findContainerStatus(pod.Status.ContainerStatuses, name)
		if cs == nil || cs.Resources == nil || !resourcesEqual(want, *cs.Resources) {
			return false
		}
	}
	return true
}

// PodResourcesMatch reports whether the pod spec already carries the desired resources
// for every target container (i.e. the resize has already been requested).
func PodResourcesMatch(pod *core.Pod, targets map[string]core.ResourceRequirements) bool {
	for name, want := range targets {
		ctr := findContainer(pod.Spec.Containers, name)
		if ctr == nil || !resourcesEqual(want, ctr.Resources) {
			return false
		}
	}
	return true
}

func resourcesEqual(a, b core.ResourceRequirements) bool {
	return apiequality.Semantic.DeepEqual(a.Requests, b.Requests) &&
		apiequality.Semantic.DeepEqual(a.Limits, b.Limits)
}

func findContainer(containers []core.Container, name string) *core.Container {
	for i := range containers {
		if containers[i].Name == name {
			return &containers[i]
		}
	}
	return nil
}

func findContainerStatus(statuses []core.ContainerStatus, name string) *core.ContainerStatus {
	for i := range statuses {
		if statuses[i].Name == name {
			return &statuses[i]
		}
	}
	return nil
}
