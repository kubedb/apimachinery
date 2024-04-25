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

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
	v1alpha1 "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
)

// FakeSolrVersions implements SolrVersionInterface
type FakeSolrVersions struct {
	Fake *FakeCatalogV1alpha1
}

var solrversionsResource = v1alpha1.SchemeGroupVersion.WithResource("solrversions")

var solrversionsKind = v1alpha1.SchemeGroupVersion.WithKind("SolrVersion")

// Get takes name of the solrVersion, and returns the corresponding solrVersion object, and an error if there is any.
func (c *FakeSolrVersions) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.SolrVersion, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(solrversionsResource, name), &v1alpha1.SolrVersion{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.SolrVersion), err
}

// List takes label and field selectors, and returns the list of SolrVersions that match those selectors.
func (c *FakeSolrVersions) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.SolrVersionList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(solrversionsResource, solrversionsKind, opts), &v1alpha1.SolrVersionList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.SolrVersionList{ListMeta: obj.(*v1alpha1.SolrVersionList).ListMeta}
	for _, item := range obj.(*v1alpha1.SolrVersionList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested solrVersions.
func (c *FakeSolrVersions) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(solrversionsResource, opts))
}

// Create takes the representation of a solrVersion and creates it.  Returns the server's representation of the solrVersion, and an error, if there is any.
func (c *FakeSolrVersions) Create(ctx context.Context, solrVersion *v1alpha1.SolrVersion, opts v1.CreateOptions) (result *v1alpha1.SolrVersion, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(solrversionsResource, solrVersion), &v1alpha1.SolrVersion{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.SolrVersion), err
}

// Update takes the representation of a solrVersion and updates it. Returns the server's representation of the solrVersion, and an error, if there is any.
func (c *FakeSolrVersions) Update(ctx context.Context, solrVersion *v1alpha1.SolrVersion, opts v1.UpdateOptions) (result *v1alpha1.SolrVersion, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(solrversionsResource, solrVersion), &v1alpha1.SolrVersion{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.SolrVersion), err
}

// Delete takes name of the solrVersion and deletes it. Returns an error if one occurs.
func (c *FakeSolrVersions) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteActionWithOptions(solrversionsResource, name, opts), &v1alpha1.SolrVersion{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeSolrVersions) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(solrversionsResource, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.SolrVersionList{})
	return err
}

// Patch applies the patch and returns the patched solrVersion.
func (c *FakeSolrVersions) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.SolrVersion, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(solrversionsResource, name, pt, data, subresources...), &v1alpha1.SolrVersion{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.SolrVersion), err
}
