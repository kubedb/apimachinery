//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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

// Code generated by deepcopy-gen. DO NOT EDIT.

package v1alpha1

import (
	apiv1 "kmodules.xyz/client-go/api/v1"
	offshootapiapiv1 "kmodules.xyz/offshoot-api/api/v1"
	shared "kmodules.xyz/resource-metadata/apis/shared"
	api "kmodules.xyz/resource-metrics/api"

	v1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	version "k8s.io/apimachinery/pkg/version"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ContainerView) DeepCopyInto(out *ContainerView) {
	*out = *in
	if in.Command != nil {
		in, out := &in.Command, &out.Command
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Args != nil {
		in, out := &in.Args, &out.Args
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Ports != nil {
		in, out := &in.Ports, &out.Ports
		*out = make([]v1.ContainerPort, len(*in))
		copy(*out, *in)
	}
	if in.EnvFrom != nil {
		in, out := &in.EnvFrom, &out.EnvFrom
		*out = make([]v1.EnvFromSource, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Env != nil {
		in, out := &in.Env, &out.Env
		*out = make([]v1.EnvVar, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.Resources.DeepCopyInto(&out.Resources)
	if in.VolumeMounts != nil {
		in, out := &in.VolumeMounts, &out.VolumeMounts
		*out = make([]v1.VolumeMount, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.VolumeDevices != nil {
		in, out := &in.VolumeDevices, &out.VolumeDevices
		*out = make([]v1.VolumeDevice, len(*in))
		copy(*out, *in)
	}
	if in.LivenessProbe != nil {
		in, out := &in.LivenessProbe, &out.LivenessProbe
		*out = new(v1.Probe)
		(*in).DeepCopyInto(*out)
	}
	if in.ReadinessProbe != nil {
		in, out := &in.ReadinessProbe, &out.ReadinessProbe
		*out = new(v1.Probe)
		(*in).DeepCopyInto(*out)
	}
	if in.StartupProbe != nil {
		in, out := &in.StartupProbe, &out.StartupProbe
		*out = new(v1.Probe)
		(*in).DeepCopyInto(*out)
	}
	if in.Lifecycle != nil {
		in, out := &in.Lifecycle, &out.Lifecycle
		*out = new(v1.Lifecycle)
		(*in).DeepCopyInto(*out)
	}
	if in.SecurityContext != nil {
		in, out := &in.SecurityContext, &out.SecurityContext
		*out = new(v1.SecurityContext)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ContainerView.
func (in *ContainerView) DeepCopy() *ContainerView {
	if in == nil {
		return nil
	}
	out := new(ContainerView)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ControlPlaneInfo) DeepCopyInto(out *ControlPlaneInfo) {
	*out = *in
	if in.DNSNames != nil {
		in, out := &in.DNSNames, &out.DNSNames
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.EmailAddresses != nil {
		in, out := &in.EmailAddresses, &out.EmailAddresses
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.IPAddresses != nil {
		in, out := &in.IPAddresses, &out.IPAddresses
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.URIs != nil {
		in, out := &in.URIs, &out.URIs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	in.NotBefore.DeepCopyInto(&out.NotBefore)
	in.NotAfter.DeepCopyInto(&out.NotAfter)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ControlPlaneInfo.
func (in *ControlPlaneInfo) DeepCopy() *ControlPlaneInfo {
	if in == nil {
		return nil
	}
	out := new(ControlPlaneInfo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExecServiceFacilitator) DeepCopyInto(out *ExecServiceFacilitator) {
	*out = *in
	out.Ref = in.Ref
	if in.Command != nil {
		in, out := &in.Command, &out.Command
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExecServiceFacilitator.
func (in *ExecServiceFacilitator) DeepCopy() *ExecServiceFacilitator {
	if in == nil {
		return nil
	}
	out := new(ExecServiceFacilitator)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GenericResource) DeepCopyInto(out *GenericResource) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	if in.Status != nil {
		in, out := &in.Status, &out.Status
		*out = new(runtime.RawExtension)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GenericResource.
func (in *GenericResource) DeepCopy() *GenericResource {
	if in == nil {
		return nil
	}
	out := new(GenericResource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *GenericResource) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GenericResourceList) DeepCopyInto(out *GenericResourceList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]GenericResource, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GenericResourceList.
func (in *GenericResourceList) DeepCopy() *GenericResourceList {
	if in == nil {
		return nil
	}
	out := new(GenericResourceList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *GenericResourceList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GenericResourceService) DeepCopyInto(out *GenericResourceService) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	if in.Status != nil {
		in, out := &in.Status, &out.Status
		*out = new(runtime.RawExtension)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GenericResourceService.
func (in *GenericResourceService) DeepCopy() *GenericResourceService {
	if in == nil {
		return nil
	}
	out := new(GenericResourceService)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *GenericResourceService) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GenericResourceServiceFacilitator) DeepCopyInto(out *GenericResourceServiceFacilitator) {
	*out = *in
	if in.Resource != nil {
		in, out := &in.Resource, &out.Resource
		*out = new(apiv1.ResourceID)
		**out = **in
	}
	if in.Refs != nil {
		in, out := &in.Refs, &out.Refs
		*out = make([]apiv1.ObjectReference, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GenericResourceServiceFacilitator.
func (in *GenericResourceServiceFacilitator) DeepCopy() *GenericResourceServiceFacilitator {
	if in == nil {
		return nil
	}
	out := new(GenericResourceServiceFacilitator)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GenericResourceServiceFacilities) DeepCopyInto(out *GenericResourceServiceFacilities) {
	*out = *in
	in.Exposed.DeepCopyInto(&out.Exposed)
	in.TLS.DeepCopyInto(&out.TLS)
	in.Backup.DeepCopyInto(&out.Backup)
	in.Monitoring.DeepCopyInto(&out.Monitoring)
	if in.Exec != nil {
		in, out := &in.Exec, &out.Exec
		*out = make([]ExecServiceFacilitator, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Gateway != nil {
		in, out := &in.Gateway, &out.Gateway
		*out = new(offshootapiapiv1.Gateway)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GenericResourceServiceFacilities.
func (in *GenericResourceServiceFacilities) DeepCopy() *GenericResourceServiceFacilities {
	if in == nil {
		return nil
	}
	out := new(GenericResourceServiceFacilities)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GenericResourceServiceList) DeepCopyInto(out *GenericResourceServiceList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]GenericResourceService, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GenericResourceServiceList.
func (in *GenericResourceServiceList) DeepCopy() *GenericResourceServiceList {
	if in == nil {
		return nil
	}
	out := new(GenericResourceServiceList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *GenericResourceServiceList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GenericResourceServiceSpec) DeepCopyInto(out *GenericResourceServiceSpec) {
	*out = *in
	out.Cluster = in.Cluster
	out.APIType = in.APIType
	in.Facilities.DeepCopyInto(&out.Facilities)
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GenericResourceServiceSpec.
func (in *GenericResourceServiceSpec) DeepCopy() *GenericResourceServiceSpec {
	if in == nil {
		return nil
	}
	out := new(GenericResourceServiceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GenericResourceServiceStatus) DeepCopyInto(out *GenericResourceServiceStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GenericResourceServiceStatus.
func (in *GenericResourceServiceStatus) DeepCopy() *GenericResourceServiceStatus {
	if in == nil {
		return nil
	}
	out := new(GenericResourceServiceStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GenericResourceSpec) DeepCopyInto(out *GenericResourceSpec) {
	*out = *in
	out.Cluster = in.Cluster
	out.APIType = in.APIType
	if in.RoleReplicas != nil {
		in, out := &in.RoleReplicas, &out.RoleReplicas
		*out = make(api.ReplicaList, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	in.TotalResource.DeepCopyInto(&out.TotalResource)
	in.AppResource.DeepCopyInto(&out.AppResource)
	if in.RoleResourceLimits != nil {
		in, out := &in.RoleResourceLimits, &out.RoleResourceLimits
		*out = make(map[api.PodRole]v1.ResourceList, len(*in))
		for key, val := range *in {
			var outVal map[v1.ResourceName]resource.Quantity
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = make(v1.ResourceList, len(*in))
				for key, val := range *in {
					(*out)[key] = val.DeepCopy()
				}
			}
			(*out)[key] = outVal
		}
	}
	if in.RoleResourceRequests != nil {
		in, out := &in.RoleResourceRequests, &out.RoleResourceRequests
		*out = make(map[api.PodRole]v1.ResourceList, len(*in))
		for key, val := range *in {
			var outVal map[v1.ResourceName]resource.Quantity
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = make(v1.ResourceList, len(*in))
				for key, val := range *in {
					(*out)[key] = val.DeepCopy()
				}
			}
			(*out)[key] = outVal
		}
	}
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GenericResourceSpec.
func (in *GenericResourceSpec) DeepCopy() *GenericResourceSpec {
	if in == nil {
		return nil
	}
	out := new(GenericResourceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GenericResourceStatus) DeepCopyInto(out *GenericResourceStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GenericResourceStatus.
func (in *GenericResourceStatus) DeepCopy() *GenericResourceStatus {
	if in == nil {
		return nil
	}
	out := new(GenericResourceStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KubernetesInfo) DeepCopyInto(out *KubernetesInfo) {
	*out = *in
	if in.Version != nil {
		in, out := &in.Version, &out.Version
		*out = new(version.Info)
		**out = **in
	}
	if in.ControlPlane != nil {
		in, out := &in.ControlPlane, &out.ControlPlane
		*out = new(ControlPlaneInfo)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KubernetesInfo.
func (in *KubernetesInfo) DeepCopy() *KubernetesInfo {
	if in == nil {
		return nil
	}
	out := new(KubernetesInfo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PodView) DeepCopyInto(out *PodView) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PodView.
func (in *PodView) DeepCopy() *PodView {
	if in == nil {
		return nil
	}
	out := new(PodView)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PodView) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PodViewList) DeepCopyInto(out *PodViewList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]PodView, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PodViewList.
func (in *PodViewList) DeepCopy() *PodViewList {
	if in == nil {
		return nil
	}
	out := new(PodViewList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PodViewList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PodViewSpec) DeepCopyInto(out *PodViewSpec) {
	*out = *in
	in.Resources.DeepCopyInto(&out.Resources)
	if in.Containers != nil {
		in, out := &in.Containers, &out.Containers
		*out = make([]ContainerView, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PodViewSpec.
func (in *PodViewSpec) DeepCopy() *PodViewSpec {
	if in == nil {
		return nil
	}
	out := new(PodViewSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Project) DeepCopyInto(out *Project) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Project.
func (in *Project) DeepCopy() *Project {
	if in == nil {
		return nil
	}
	out := new(Project)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Project) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ProjectList) DeepCopyInto(out *ProjectList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Project, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ProjectList.
func (in *ProjectList) DeepCopy() *ProjectList {
	if in == nil {
		return nil
	}
	out := new(ProjectList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ProjectList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ProjectMonitoring) DeepCopyInto(out *ProjectMonitoring) {
	*out = *in
	if in.PrometheusRef != nil {
		in, out := &in.PrometheusRef, &out.PrometheusRef
		*out = new(apiv1.ObjectReference)
		**out = **in
	}
	if in.AlertmanagerRef != nil {
		in, out := &in.AlertmanagerRef, &out.AlertmanagerRef
		*out = new(apiv1.ObjectReference)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ProjectMonitoring.
func (in *ProjectMonitoring) DeepCopy() *ProjectMonitoring {
	if in == nil {
		return nil
	}
	out := new(ProjectMonitoring)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ProjectSpec) DeepCopyInto(out *ProjectSpec) {
	*out = *in
	if in.Namespaces != nil {
		in, out := &in.Namespaces, &out.Namespaces
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.NamespaceSelector != nil {
		in, out := &in.NamespaceSelector, &out.NamespaceSelector
		*out = new(metav1.LabelSelector)
		(*in).DeepCopyInto(*out)
	}
	if in.Monitoring != nil {
		in, out := &in.Monitoring, &out.Monitoring
		*out = new(ProjectMonitoring)
		(*in).DeepCopyInto(*out)
	}
	if in.Presets != nil {
		in, out := &in.Presets, &out.Presets
		*out = make([]shared.SourceLocator, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ProjectSpec.
func (in *ProjectSpec) DeepCopy() *ProjectSpec {
	if in == nil {
		return nil
	}
	out := new(ProjectSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceSummary) DeepCopyInto(out *ResourceSummary) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceSummary.
func (in *ResourceSummary) DeepCopy() *ResourceSummary {
	if in == nil {
		return nil
	}
	out := new(ResourceSummary)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ResourceSummary) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceSummaryList) DeepCopyInto(out *ResourceSummaryList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ResourceSummary, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceSummaryList.
func (in *ResourceSummaryList) DeepCopy() *ResourceSummaryList {
	if in == nil {
		return nil
	}
	out := new(ResourceSummaryList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ResourceSummaryList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceSummarySpec) DeepCopyInto(out *ResourceSummarySpec) {
	*out = *in
	out.Cluster = in.Cluster
	out.APIType = in.APIType
	in.TotalResource.DeepCopyInto(&out.TotalResource)
	in.AppResource.DeepCopyInto(&out.AppResource)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceSummarySpec.
func (in *ResourceSummarySpec) DeepCopy() *ResourceSummarySpec {
	if in == nil {
		return nil
	}
	out := new(ResourceSummarySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceView) DeepCopyInto(out *ResourceView) {
	*out = *in
	if in.Limits != nil {
		in, out := &in.Limits, &out.Limits
		*out = make(v1.ResourceList, len(*in))
		for key, val := range *in {
			(*out)[key] = val.DeepCopy()
		}
	}
	if in.Requests != nil {
		in, out := &in.Requests, &out.Requests
		*out = make(v1.ResourceList, len(*in))
		for key, val := range *in {
			(*out)[key] = val.DeepCopy()
		}
	}
	if in.Usage != nil {
		in, out := &in.Usage, &out.Usage
		*out = make(v1.ResourceList, len(*in))
		for key, val := range *in {
			(*out)[key] = val.DeepCopy()
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceView.
func (in *ResourceView) DeepCopy() *ResourceView {
	if in == nil {
		return nil
	}
	out := new(ResourceView)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Service) DeepCopyInto(out *Service) {
	*out = *in
	out.ResourceLocator = in.ResourceLocator
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Service.
func (in *Service) DeepCopy() *Service {
	if in == nil {
		return nil
	}
	out := new(Service)
	in.DeepCopyInto(out)
	return out
}
