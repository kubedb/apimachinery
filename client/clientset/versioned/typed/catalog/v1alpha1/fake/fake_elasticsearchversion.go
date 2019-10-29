/*
Copyright The KubeDB Authors.

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
	v1alpha1 "kubedb.dev/apimachinery/apis/catalog/v1alpha1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeElasticsearchVersions implements ElasticsearchVersionInterface
type FakeElasticsearchVersions struct {
	Fake *FakeCatalogV1alpha1
}

var elasticsearchversionsResource = schema.GroupVersionResource{Group: "catalog.kubedb.com", Version: "v1alpha1", Resource: "elasticsearchversions"}

var elasticsearchversionsKind = schema.GroupVersionKind{Group: "catalog.kubedb.com", Version: "v1alpha1", Kind: "ElasticsearchVersion"}

// Get takes name of the elasticsearchVersion, and returns the corresponding elasticsearchVersion object, and an error if there is any.
func (c *FakeElasticsearchVersions) Get(name string, options v1.GetOptions) (result *v1alpha1.ElasticsearchVersion, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(elasticsearchversionsResource, name), &v1alpha1.ElasticsearchVersion{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ElasticsearchVersion), err
}

// List takes label and field selectors, and returns the list of ElasticsearchVersions that match those selectors.
func (c *FakeElasticsearchVersions) List(opts v1.ListOptions) (result *v1alpha1.ElasticsearchVersionList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(elasticsearchversionsResource, elasticsearchversionsKind, opts), &v1alpha1.ElasticsearchVersionList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.ElasticsearchVersionList{ListMeta: obj.(*v1alpha1.ElasticsearchVersionList).ListMeta}
	for _, item := range obj.(*v1alpha1.ElasticsearchVersionList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested elasticsearchVersions.
func (c *FakeElasticsearchVersions) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(elasticsearchversionsResource, opts))
}

// Create takes the representation of a elasticsearchVersion and creates it.  Returns the server's representation of the elasticsearchVersion, and an error, if there is any.
func (c *FakeElasticsearchVersions) Create(elasticsearchVersion *v1alpha1.ElasticsearchVersion) (result *v1alpha1.ElasticsearchVersion, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(elasticsearchversionsResource, elasticsearchVersion), &v1alpha1.ElasticsearchVersion{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ElasticsearchVersion), err
}

// Update takes the representation of a elasticsearchVersion and updates it. Returns the server's representation of the elasticsearchVersion, and an error, if there is any.
func (c *FakeElasticsearchVersions) Update(elasticsearchVersion *v1alpha1.ElasticsearchVersion) (result *v1alpha1.ElasticsearchVersion, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(elasticsearchversionsResource, elasticsearchVersion), &v1alpha1.ElasticsearchVersion{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ElasticsearchVersion), err
}

// Delete takes name of the elasticsearchVersion and deletes it. Returns an error if one occurs.
func (c *FakeElasticsearchVersions) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteAction(elasticsearchversionsResource, name), &v1alpha1.ElasticsearchVersion{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeElasticsearchVersions) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(elasticsearchversionsResource, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.ElasticsearchVersionList{})
	return err
}

// Patch applies the patch and returns the patched elasticsearchVersion.
func (c *FakeElasticsearchVersions) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.ElasticsearchVersion, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(elasticsearchversionsResource, name, pt, data, subresources...), &v1alpha1.ElasticsearchVersion{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ElasticsearchVersion), err
}
