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
	"context"

	v1alpha2 "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeElasticsearches implements ElasticsearchInterface
type FakeElasticsearches struct {
	Fake *FakeKubedbV1alpha2
	ns   string
}

var elasticsearchesResource = schema.GroupVersionResource{Group: "kubedb.com", Version: "v1alpha2", Resource: "elasticsearches"}

var elasticsearchesKind = schema.GroupVersionKind{Group: "kubedb.com", Version: "v1alpha2", Kind: "Elasticsearch"}

// Get takes name of the elasticsearch, and returns the corresponding elasticsearch object, and an error if there is any.
func (c *FakeElasticsearches) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha2.Elasticsearch, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(elasticsearchesResource, c.ns, name), &v1alpha2.Elasticsearch{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.Elasticsearch), err
}

// List takes label and field selectors, and returns the list of Elasticsearches that match those selectors.
func (c *FakeElasticsearches) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha2.ElasticsearchList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(elasticsearchesResource, elasticsearchesKind, c.ns, opts), &v1alpha2.ElasticsearchList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha2.ElasticsearchList{ListMeta: obj.(*v1alpha2.ElasticsearchList).ListMeta}
	for _, item := range obj.(*v1alpha2.ElasticsearchList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested elasticsearches.
func (c *FakeElasticsearches) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(elasticsearchesResource, c.ns, opts))

}

// Create takes the representation of a elasticsearch and creates it.  Returns the server's representation of the elasticsearch, and an error, if there is any.
func (c *FakeElasticsearches) Create(ctx context.Context, elasticsearch *v1alpha2.Elasticsearch, opts v1.CreateOptions) (result *v1alpha2.Elasticsearch, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(elasticsearchesResource, c.ns, elasticsearch), &v1alpha2.Elasticsearch{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.Elasticsearch), err
}

// Update takes the representation of a elasticsearch and updates it. Returns the server's representation of the elasticsearch, and an error, if there is any.
func (c *FakeElasticsearches) Update(ctx context.Context, elasticsearch *v1alpha2.Elasticsearch, opts v1.UpdateOptions) (result *v1alpha2.Elasticsearch, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(elasticsearchesResource, c.ns, elasticsearch), &v1alpha2.Elasticsearch{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.Elasticsearch), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeElasticsearches) UpdateStatus(ctx context.Context, elasticsearch *v1alpha2.Elasticsearch, opts v1.UpdateOptions) (*v1alpha2.Elasticsearch, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(elasticsearchesResource, "status", c.ns, elasticsearch), &v1alpha2.Elasticsearch{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.Elasticsearch), err
}

// Delete takes name of the elasticsearch and deletes it. Returns an error if one occurs.
func (c *FakeElasticsearches) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(elasticsearchesResource, c.ns, name, opts), &v1alpha2.Elasticsearch{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeElasticsearches) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(elasticsearchesResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha2.ElasticsearchList{})
	return err
}

// Patch applies the patch and returns the patched elasticsearch.
func (c *FakeElasticsearches) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha2.Elasticsearch, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(elasticsearchesResource, c.ns, name, pt, data, subresources...), &v1alpha2.Elasticsearch{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.Elasticsearch), err
}
