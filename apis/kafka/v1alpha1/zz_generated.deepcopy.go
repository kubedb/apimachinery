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
	corev1 "k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	v1 "kmodules.xyz/client-go/api/v1"
	apiv1 "kmodules.xyz/monitoring-agent-api/api/v1"
	kubedbv1 "kubedb.dev/apimachinery/apis/kubedb/v1"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ConnectCluster) DeepCopyInto(out *ConnectCluster) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ConnectCluster.
func (in *ConnectCluster) DeepCopy() *ConnectCluster {
	if in == nil {
		return nil
	}
	out := new(ConnectCluster)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ConnectCluster) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ConnectClusterApp) DeepCopyInto(out *ConnectClusterApp) {
	*out = *in
	if in.ConnectCluster != nil {
		in, out := &in.ConnectCluster, &out.ConnectCluster
		*out = new(ConnectCluster)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ConnectClusterApp.
func (in *ConnectClusterApp) DeepCopy() *ConnectClusterApp {
	if in == nil {
		return nil
	}
	out := new(ConnectClusterApp)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ConnectClusterList) DeepCopyInto(out *ConnectClusterList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ConnectCluster, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ConnectClusterList.
func (in *ConnectClusterList) DeepCopy() *ConnectClusterList {
	if in == nil {
		return nil
	}
	out := new(ConnectClusterList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ConnectClusterList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ConnectClusterSpec) DeepCopyInto(out *ConnectClusterSpec) {
	*out = *in
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int32)
		**out = **in
	}
	if in.KafkaRef != nil {
		in, out := &in.KafkaRef, &out.KafkaRef
		*out = new(v1.ObjectReference)
		**out = **in
	}
	if in.AuthSecret != nil {
		in, out := &in.AuthSecret, &out.AuthSecret
		*out = new(kubedbv1.SecretReference)
		**out = **in
	}
	if in.KeystoreCredSecret != nil {
		in, out := &in.KeystoreCredSecret, &out.KeystoreCredSecret
		*out = new(kubedbv1.SecretReference)
		**out = **in
	}
	if in.TLS != nil {
		in, out := &in.TLS, &out.TLS
		*out = new(v1.TLSConfig)
		(*in).DeepCopyInto(*out)
	}
	if in.ConnectorPlugins != nil {
		in, out := &in.ConnectorPlugins, &out.ConnectorPlugins
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.ConfigSecret != nil {
		in, out := &in.ConfigSecret, &out.ConfigSecret
		*out = new(corev1.LocalObjectReference)
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
	in.HealthChecker.DeepCopyInto(&out.HealthChecker)
	if in.Monitor != nil {
		in, out := &in.Monitor, &out.Monitor
		*out = new(apiv1.AgentSpec)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ConnectClusterSpec.
func (in *ConnectClusterSpec) DeepCopy() *ConnectClusterSpec {
	if in == nil {
		return nil
	}
	out := new(ConnectClusterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ConnectClusterStatus) DeepCopyInto(out *ConnectClusterStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ConnectClusterStatus.
func (in *ConnectClusterStatus) DeepCopy() *ConnectClusterStatus {
	if in == nil {
		return nil
	}
	out := new(ConnectClusterStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Connector) DeepCopyInto(out *Connector) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Connector.
func (in *Connector) DeepCopy() *Connector {
	if in == nil {
		return nil
	}
	out := new(Connector)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Connector) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ConnectorList) DeepCopyInto(out *ConnectorList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Connector, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ConnectorList.
func (in *ConnectorList) DeepCopy() *ConnectorList {
	if in == nil {
		return nil
	}
	out := new(ConnectorList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ConnectorList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ConnectorSpec) DeepCopyInto(out *ConnectorSpec) {
	*out = *in
	if in.ConnectClusterRef != nil {
		in, out := &in.ConnectClusterRef, &out.ConnectClusterRef
		*out = new(v1.ObjectReference)
		**out = **in
	}
	if in.ConfigSecret != nil {
		in, out := &in.ConfigSecret, &out.ConfigSecret
		*out = new(corev1.LocalObjectReference)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ConnectorSpec.
func (in *ConnectorSpec) DeepCopy() *ConnectorSpec {
	if in == nil {
		return nil
	}
	out := new(ConnectorSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ConnectorStatus) DeepCopyInto(out *ConnectorStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ConnectorStatus.
func (in *ConnectorStatus) DeepCopy() *ConnectorStatus {
	if in == nil {
		return nil
	}
	out := new(ConnectorStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RestProxy) DeepCopyInto(out *RestProxy) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RestProxy.
func (in *RestProxy) DeepCopy() *RestProxy {
	if in == nil {
		return nil
	}
	out := new(RestProxy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RestProxy) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RestProxyApp) DeepCopyInto(out *RestProxyApp) {
	*out = *in
	if in.RestProxy != nil {
		in, out := &in.RestProxy, &out.RestProxy
		*out = new(RestProxy)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RestProxyApp.
func (in *RestProxyApp) DeepCopy() *RestProxyApp {
	if in == nil {
		return nil
	}
	out := new(RestProxyApp)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RestProxyList) DeepCopyInto(out *RestProxyList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]RestProxy, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RestProxyList.
func (in *RestProxyList) DeepCopy() *RestProxyList {
	if in == nil {
		return nil
	}
	out := new(RestProxyList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RestProxyList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RestProxySpec) DeepCopyInto(out *RestProxySpec) {
	*out = *in
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int32)
		**out = **in
	}
	if in.KafkaRef != nil {
		in, out := &in.KafkaRef, &out.KafkaRef
		*out = new(v1.ObjectReference)
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
	in.HealthChecker.DeepCopyInto(&out.HealthChecker)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RestProxySpec.
func (in *RestProxySpec) DeepCopy() *RestProxySpec {
	if in == nil {
		return nil
	}
	out := new(RestProxySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RestProxyStatus) DeepCopyInto(out *RestProxyStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RestProxyStatus.
func (in *RestProxyStatus) DeepCopy() *RestProxyStatus {
	if in == nil {
		return nil
	}
	out := new(RestProxyStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SchemaRegistry) DeepCopyInto(out *SchemaRegistry) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SchemaRegistry.
func (in *SchemaRegistry) DeepCopy() *SchemaRegistry {
	if in == nil {
		return nil
	}
	out := new(SchemaRegistry)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SchemaRegistry) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SchemaRegistryApp) DeepCopyInto(out *SchemaRegistryApp) {
	*out = *in
	if in.SchemaRegistry != nil {
		in, out := &in.SchemaRegistry, &out.SchemaRegistry
		*out = new(SchemaRegistry)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SchemaRegistryApp.
func (in *SchemaRegistryApp) DeepCopy() *SchemaRegistryApp {
	if in == nil {
		return nil
	}
	out := new(SchemaRegistryApp)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SchemaRegistryList) DeepCopyInto(out *SchemaRegistryList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]SchemaRegistry, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SchemaRegistryList.
func (in *SchemaRegistryList) DeepCopy() *SchemaRegistryList {
	if in == nil {
		return nil
	}
	out := new(SchemaRegistryList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SchemaRegistryList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SchemaRegistrySpec) DeepCopyInto(out *SchemaRegistrySpec) {
	*out = *in
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int32)
		**out = **in
	}
	if in.KafkaRef != nil {
		in, out := &in.KafkaRef, &out.KafkaRef
		*out = new(v1.ObjectReference)
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
	in.HealthChecker.DeepCopyInto(&out.HealthChecker)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SchemaRegistrySpec.
func (in *SchemaRegistrySpec) DeepCopy() *SchemaRegistrySpec {
	if in == nil {
		return nil
	}
	out := new(SchemaRegistrySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SchemaRegistryStatus) DeepCopyInto(out *SchemaRegistryStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SchemaRegistryStatus.
func (in *SchemaRegistryStatus) DeepCopy() *SchemaRegistryStatus {
	if in == nil {
		return nil
	}
	out := new(SchemaRegistryStatus)
	in.DeepCopyInto(out)
	return out
}
