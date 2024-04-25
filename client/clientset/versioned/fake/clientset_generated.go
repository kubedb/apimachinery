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

package fake

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/discovery"
	fakediscovery "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/testing"
	clientset "kubedb.dev/apimachinery/client/clientset/versioned"
	archiverv1alpha1 "kubedb.dev/apimachinery/client/clientset/versioned/typed/archiver/v1alpha1"
	fakearchiverv1alpha1 "kubedb.dev/apimachinery/client/clientset/versioned/typed/archiver/v1alpha1/fake"
	autoscalingv1alpha1 "kubedb.dev/apimachinery/client/clientset/versioned/typed/autoscaling/v1alpha1"
	fakeautoscalingv1alpha1 "kubedb.dev/apimachinery/client/clientset/versioned/typed/autoscaling/v1alpha1/fake"
	catalogv1alpha1 "kubedb.dev/apimachinery/client/clientset/versioned/typed/catalog/v1alpha1"
	fakecatalogv1alpha1 "kubedb.dev/apimachinery/client/clientset/versioned/typed/catalog/v1alpha1/fake"
	configv1alpha1 "kubedb.dev/apimachinery/client/clientset/versioned/typed/config/v1alpha1"
	fakeconfigv1alpha1 "kubedb.dev/apimachinery/client/clientset/versioned/typed/config/v1alpha1/fake"
	elasticsearchv1alpha1 "kubedb.dev/apimachinery/client/clientset/versioned/typed/elasticsearch/v1alpha1"
	fakeelasticsearchv1alpha1 "kubedb.dev/apimachinery/client/clientset/versioned/typed/elasticsearch/v1alpha1/fake"
	kafkav1alpha1 "kubedb.dev/apimachinery/client/clientset/versioned/typed/kafka/v1alpha1"
	fakekafkav1alpha1 "kubedb.dev/apimachinery/client/clientset/versioned/typed/kafka/v1alpha1/fake"
	kubedbv1alpha1 "kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1"
	fakekubedbv1alpha1 "kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/fake"
	kubedbv1alpha2 "kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha2"
	fakekubedbv1alpha2 "kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha2/fake"
	kubedbv1alpha3 "kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha3"
	fakekubedbv1alpha3 "kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha3/fake"
	opsv1alpha1 "kubedb.dev/apimachinery/client/clientset/versioned/typed/ops/v1alpha1"
	fakeopsv1alpha1 "kubedb.dev/apimachinery/client/clientset/versioned/typed/ops/v1alpha1/fake"
	postgresv1alpha1 "kubedb.dev/apimachinery/client/clientset/versioned/typed/postgres/v1alpha1"
	fakepostgresv1alpha1 "kubedb.dev/apimachinery/client/clientset/versioned/typed/postgres/v1alpha1/fake"
	schemav1alpha1 "kubedb.dev/apimachinery/client/clientset/versioned/typed/schema/v1alpha1"
	fakeschemav1alpha1 "kubedb.dev/apimachinery/client/clientset/versioned/typed/schema/v1alpha1/fake"
	uiv1alpha1 "kubedb.dev/apimachinery/client/clientset/versioned/typed/ui/v1alpha1"
	fakeuiv1alpha1 "kubedb.dev/apimachinery/client/clientset/versioned/typed/ui/v1alpha1/fake"
)

// NewSimpleClientset returns a clientset that will respond with the provided objects.
// It's backed by a very simple object tracker that processes creates, updates and deletions as-is,
// without applying any validations and/or defaults. It shouldn't be considered a replacement
// for a real clientset and is mostly useful in simple unit tests.
func NewSimpleClientset(objects ...runtime.Object) *Clientset {
	o := testing.NewObjectTracker(scheme, codecs.UniversalDecoder())
	for _, obj := range objects {
		if err := o.Add(obj); err != nil {
			panic(err)
		}
	}

	cs := &Clientset{tracker: o}
	cs.discovery = &fakediscovery.FakeDiscovery{Fake: &cs.Fake}
	cs.AddReactor("*", "*", testing.ObjectReaction(o))
	cs.AddWatchReactor("*", func(action testing.Action) (handled bool, ret watch.Interface, err error) {
		gvr := action.GetResource()
		ns := action.GetNamespace()
		watch, err := o.Watch(gvr, ns)
		if err != nil {
			return false, nil, err
		}
		return true, watch, nil
	})

	return cs
}

// Clientset implements clientset.Interface. Meant to be embedded into a
// struct to get a default implementation. This makes faking out just the method
// you want to test easier.
type Clientset struct {
	testing.Fake
	discovery *fakediscovery.FakeDiscovery
	tracker   testing.ObjectTracker
}

func (c *Clientset) Discovery() discovery.DiscoveryInterface {
	return c.discovery
}

func (c *Clientset) Tracker() testing.ObjectTracker {
	return c.tracker
}

var (
	_ clientset.Interface = &Clientset{}
	_ testing.FakeClient  = &Clientset{}
)

// ArchiverV1alpha1 retrieves the ArchiverV1alpha1Client
func (c *Clientset) ArchiverV1alpha1() archiverv1alpha1.ArchiverV1alpha1Interface {
	return &fakearchiverv1alpha1.FakeArchiverV1alpha1{Fake: &c.Fake}
}

// AutoscalingV1alpha1 retrieves the AutoscalingV1alpha1Client
func (c *Clientset) AutoscalingV1alpha1() autoscalingv1alpha1.AutoscalingV1alpha1Interface {
	return &fakeautoscalingv1alpha1.FakeAutoscalingV1alpha1{Fake: &c.Fake}
}

// CatalogV1alpha1 retrieves the CatalogV1alpha1Client
func (c *Clientset) CatalogV1alpha1() catalogv1alpha1.CatalogV1alpha1Interface {
	return &fakecatalogv1alpha1.FakeCatalogV1alpha1{Fake: &c.Fake}
}

// ConfigV1alpha1 retrieves the ConfigV1alpha1Client
func (c *Clientset) ConfigV1alpha1() configv1alpha1.ConfigV1alpha1Interface {
	return &fakeconfigv1alpha1.FakeConfigV1alpha1{Fake: &c.Fake}
}

// ElasticsearchV1alpha1 retrieves the ElasticsearchV1alpha1Client
func (c *Clientset) ElasticsearchV1alpha1() elasticsearchv1alpha1.ElasticsearchV1alpha1Interface {
	return &fakeelasticsearchv1alpha1.FakeElasticsearchV1alpha1{Fake: &c.Fake}
}

// KafkaV1alpha1 retrieves the KafkaV1alpha1Client
func (c *Clientset) KafkaV1alpha1() kafkav1alpha1.KafkaV1alpha1Interface {
	return &fakekafkav1alpha1.FakeKafkaV1alpha1{Fake: &c.Fake}
}

// KubedbV1alpha1 retrieves the KubedbV1alpha1Client
func (c *Clientset) KubedbV1alpha1() kubedbv1alpha1.KubedbV1alpha1Interface {
	return &fakekubedbv1alpha1.FakeKubedbV1alpha1{Fake: &c.Fake}
}

// KubedbV1alpha2 retrieves the KubedbV1alpha2Client
func (c *Clientset) KubedbV1alpha2() kubedbv1alpha2.KubedbV1alpha2Interface {
	return &fakekubedbv1alpha2.FakeKubedbV1alpha2{Fake: &c.Fake}
}

// KubedbV1alpha3 retrieves the KubedbV1alpha3Client
func (c *Clientset) KubedbV1alpha3() kubedbv1alpha3.KubedbV1alpha3Interface {
	return &fakekubedbv1alpha3.FakeKubedbV1alpha3{Fake: &c.Fake}
}

// OpsV1alpha1 retrieves the OpsV1alpha1Client
func (c *Clientset) OpsV1alpha1() opsv1alpha1.OpsV1alpha1Interface {
	return &fakeopsv1alpha1.FakeOpsV1alpha1{Fake: &c.Fake}
}

// PostgresV1alpha1 retrieves the PostgresV1alpha1Client
func (c *Clientset) PostgresV1alpha1() postgresv1alpha1.PostgresV1alpha1Interface {
	return &fakepostgresv1alpha1.FakePostgresV1alpha1{Fake: &c.Fake}
}

// SchemaV1alpha1 retrieves the SchemaV1alpha1Client
func (c *Clientset) SchemaV1alpha1() schemav1alpha1.SchemaV1alpha1Interface {
	return &fakeschemav1alpha1.FakeSchemaV1alpha1{Fake: &c.Fake}
}

// UiV1alpha1 retrieves the UiV1alpha1Client
func (c *Clientset) UiV1alpha1() uiv1alpha1.UiV1alpha1Interface {
	return &fakeuiv1alpha1.FakeUiV1alpha1{Fake: &c.Fake}
}
