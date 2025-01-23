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
	v1 "k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	apiv1 "kmodules.xyz/client-go/api/v1"
	kubedbv1 "kubedb.dev/apimachinery/apis/kubedb/v1"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElasticsearchDashboard) DeepCopyInto(out *ElasticsearchDashboard) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElasticsearchDashboard.
func (in *ElasticsearchDashboard) DeepCopy() *ElasticsearchDashboard {
	if in == nil {
		return nil
	}
	out := new(ElasticsearchDashboard)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ElasticsearchDashboard) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElasticsearchDashboardList) DeepCopyInto(out *ElasticsearchDashboardList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ElasticsearchDashboard, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElasticsearchDashboardList.
func (in *ElasticsearchDashboardList) DeepCopy() *ElasticsearchDashboardList {
	if in == nil {
		return nil
	}
	out := new(ElasticsearchDashboardList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ElasticsearchDashboardList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElasticsearchDashboardSpec) DeepCopyInto(out *ElasticsearchDashboardSpec) {
	*out = *in
	if in.DatabaseRef != nil {
		in, out := &in.DatabaseRef, &out.DatabaseRef
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int32)
		**out = **in
	}
	if in.AuthSecret != nil {
		in, out := &in.AuthSecret, &out.AuthSecret
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
	if in.ConfigSecret != nil {
		in, out := &in.ConfigSecret, &out.ConfigSecret
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
	in.PodTemplate.DeepCopyInto(&out.PodTemplate)
	if in.ServiceTemplates != nil {
		in, out := &in.ServiceTemplates, &out.ServiceTemplates
		*out = make([]kubedbv1.NamedServiceTemplateSpec, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.TLS != nil {
		in, out := &in.TLS, &out.TLS
		*out = new(apiv1.TLSConfig)
		(*in).DeepCopyInto(*out)
	}
	in.HealthChecker.DeepCopyInto(&out.HealthChecker)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElasticsearchDashboardSpec.
func (in *ElasticsearchDashboardSpec) DeepCopy() *ElasticsearchDashboardSpec {
	if in == nil {
		return nil
	}
	out := new(ElasticsearchDashboardSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElasticsearchDashboardStatus) DeepCopyInto(out *ElasticsearchDashboardStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]apiv1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElasticsearchDashboardStatus.
func (in *ElasticsearchDashboardStatus) DeepCopy() *ElasticsearchDashboardStatus {
	if in == nil {
		return nil
	}
	out := new(ElasticsearchDashboardStatus)
	in.DeepCopyInto(out)
	return out
}
