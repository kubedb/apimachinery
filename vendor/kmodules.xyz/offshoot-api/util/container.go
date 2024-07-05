package util

import (
	core_util "kmodules.xyz/client-go/core/v1"
	ofstv2 "kmodules.xyz/offshoot-api/api/v2"

	core "k8s.io/api/core/v1"
)

// EnsureContainerExists ensures that given container either exits by default or
// it will create the container, then insert it to the podTemplate and return a pointer of that container
func EnsureContainerExists(podTemplate *ofstv2.PodTemplateSpec, containerName string) *core.Container {
	container := core_util.GetContainerByName(podTemplate.Spec.Containers, containerName)
	if container == nil {
		container = &core.Container{
			Name: containerName,
		}
	}
	podTemplate.Spec.Containers = core_util.UpsertContainer(podTemplate.Spec.Containers, *container)
	return core_util.GetContainerByName(podTemplate.Spec.Containers, containerName)
}

// EnsureInitContainerExists ensures that given initContainer either exits by default or
// it will create the initContainer, then insert it to the podTemplate and return a pointer of that initContainer
func EnsureInitContainerExists(podTemplate *ofstv2.PodTemplateSpec, containerName string) *core.Container {
	container := core_util.GetContainerByName(podTemplate.Spec.InitContainers, containerName)
	if container == nil {
		container = &core.Container{
			Name: containerName,
		}
	}
	podTemplate.Spec.InitContainers = core_util.UpsertContainer(podTemplate.Spec.InitContainers, *container)
	return core_util.GetContainerByName(podTemplate.Spec.InitContainers, containerName)
}
