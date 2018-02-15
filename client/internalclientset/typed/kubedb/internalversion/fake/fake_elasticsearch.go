/*
Copyright 2018 The KubeDB Authors.

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

package fake

import (
	kubedb "github.com/kubedb/apimachinery/apis/kubedb"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeElasticsearches implements ElasticsearchInterface
type FakeElasticsearches struct {
	Fake *FakeKubedb
	ns   string
}

var elasticsearchesResource = schema.GroupVersionResource{Group: "kubedb.com", Version: "", Resource: "elasticsearches"}

var elasticsearchesKind = schema.GroupVersionKind{Group: "kubedb.com", Version: "", Kind: "Elasticsearch"}

// Get takes name of the elasticsearch, and returns the corresponding elasticsearch object, and an error if there is any.
func (c *FakeElasticsearches) Get(name string, options v1.GetOptions) (result *kubedb.Elasticsearch, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(elasticsearchesResource, c.ns, name), &kubedb.Elasticsearch{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kubedb.Elasticsearch), err
}

// List takes label and field selectors, and returns the list of Elasticsearches that match those selectors.
func (c *FakeElasticsearches) List(opts v1.ListOptions) (result *kubedb.ElasticsearchList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(elasticsearchesResource, elasticsearchesKind, c.ns, opts), &kubedb.ElasticsearchList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &kubedb.ElasticsearchList{}
	for _, item := range obj.(*kubedb.ElasticsearchList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested elasticsearches.
func (c *FakeElasticsearches) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(elasticsearchesResource, c.ns, opts))

}

// Create takes the representation of a elasticsearch and creates it.  Returns the server's representation of the elasticsearch, and an error, if there is any.
func (c *FakeElasticsearches) Create(elasticsearch *kubedb.Elasticsearch) (result *kubedb.Elasticsearch, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(elasticsearchesResource, c.ns, elasticsearch), &kubedb.Elasticsearch{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kubedb.Elasticsearch), err
}

// Update takes the representation of a elasticsearch and updates it. Returns the server's representation of the elasticsearch, and an error, if there is any.
func (c *FakeElasticsearches) Update(elasticsearch *kubedb.Elasticsearch) (result *kubedb.Elasticsearch, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(elasticsearchesResource, c.ns, elasticsearch), &kubedb.Elasticsearch{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kubedb.Elasticsearch), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeElasticsearches) UpdateStatus(elasticsearch *kubedb.Elasticsearch) (*kubedb.Elasticsearch, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(elasticsearchesResource, "status", c.ns, elasticsearch), &kubedb.Elasticsearch{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kubedb.Elasticsearch), err
}

// Delete takes name of the elasticsearch and deletes it. Returns an error if one occurs.
func (c *FakeElasticsearches) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(elasticsearchesResource, c.ns, name), &kubedb.Elasticsearch{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeElasticsearches) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(elasticsearchesResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &kubedb.ElasticsearchList{})
	return err
}

// Patch applies the patch and returns the patched elasticsearch.
func (c *FakeElasticsearches) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *kubedb.Elasticsearch, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(elasticsearchesResource, c.ns, name, data, subresources...), &kubedb.Elasticsearch{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kubedb.Elasticsearch), err
}
