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

// Code generated by client-gen. DO NOT EDIT.

package scheme

import (
	archiverv1alpha1 "kubedb.dev/apimachinery/apis/archiver/v1alpha1"
	autoscalingv1alpha1 "kubedb.dev/apimachinery/apis/autoscaling/v1alpha1"
	catalogv1alpha1 "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	configv1alpha1 "kubedb.dev/apimachinery/apis/config/v1alpha1"
	elasticsearchv1alpha1 "kubedb.dev/apimachinery/apis/elasticsearch/v1alpha1"
	kafkav1alpha1 "kubedb.dev/apimachinery/apis/kafka/v1alpha1"
	kubedbv1 "kubedb.dev/apimachinery/apis/kubedb/v1"
	kubedbv1alpha1 "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	kubedbv1alpha2 "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	opsv1alpha1 "kubedb.dev/apimachinery/apis/ops/v1alpha1"
	postgresv1alpha1 "kubedb.dev/apimachinery/apis/postgres/v1alpha1"
	schemav1alpha1 "kubedb.dev/apimachinery/apis/schema/v1alpha1"
	uiv1alpha1 "kubedb.dev/apimachinery/apis/ui/v1alpha1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	serializer "k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

var Scheme = runtime.NewScheme()
var Codecs = serializer.NewCodecFactory(Scheme)
var ParameterCodec = runtime.NewParameterCodec(Scheme)
var localSchemeBuilder = runtime.SchemeBuilder{
	archiverv1alpha1.AddToScheme,
	autoscalingv1alpha1.AddToScheme,
	catalogv1alpha1.AddToScheme,
	configv1alpha1.AddToScheme,
	elasticsearchv1alpha1.AddToScheme,
	kafkav1alpha1.AddToScheme,
	kubedbv1alpha1.AddToScheme,
	kubedbv1alpha2.AddToScheme,
	kubedbv1.AddToScheme,
	opsv1alpha1.AddToScheme,
	postgresv1alpha1.AddToScheme,
	schemav1alpha1.AddToScheme,
	uiv1alpha1.AddToScheme,
}

// AddToScheme adds all types of this clientset into the given scheme. This allows composition
// of clientsets, like in:
//
//	import (
//	  "k8s.io/client-go/kubernetes"
//	  clientsetscheme "k8s.io/client-go/kubernetes/scheme"
//	  aggregatorclientsetscheme "k8s.io/kube-aggregator/pkg/client/clientset_generated/clientset/scheme"
//	)
//
//	kclientset, _ := kubernetes.NewForConfig(c)
//	_ = aggregatorclientsetscheme.AddToScheme(clientsetscheme.Scheme)
//
// After this, RawExtensions in Kubernetes types will serialize kube-aggregator types
// correctly.
var AddToScheme = localSchemeBuilder.AddToScheme

func init() {
	v1.AddToGroupVersion(Scheme, schema.GroupVersion{Version: "v1"})
	utilruntime.Must(AddToScheme(Scheme))
}
