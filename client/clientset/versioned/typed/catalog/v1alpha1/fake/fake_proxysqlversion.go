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

	v1alpha1 "kubedb.dev/apimachinery/apis/catalog/v1alpha1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeProxySQLVersions implements ProxySQLVersionInterface
type FakeProxySQLVersions struct {
	Fake *FakeCatalogV1alpha1
}

var proxysqlversionsResource = schema.GroupVersionResource{Group: "catalog.kubedb.com", Version: "v1alpha1", Resource: "proxysqlversions"}

var proxysqlversionsKind = schema.GroupVersionKind{Group: "catalog.kubedb.com", Version: "v1alpha1", Kind: "ProxySQLVersion"}

// Get takes name of the proxySQLVersion, and returns the corresponding proxySQLVersion object, and an error if there is any.
func (c *FakeProxySQLVersions) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.ProxySQLVersion, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(proxysqlversionsResource, name), &v1alpha1.ProxySQLVersion{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ProxySQLVersion), err
}

// List takes label and field selectors, and returns the list of ProxySQLVersions that match those selectors.
func (c *FakeProxySQLVersions) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.ProxySQLVersionList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(proxysqlversionsResource, proxysqlversionsKind, opts), &v1alpha1.ProxySQLVersionList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.ProxySQLVersionList{ListMeta: obj.(*v1alpha1.ProxySQLVersionList).ListMeta}
	for _, item := range obj.(*v1alpha1.ProxySQLVersionList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested proxySQLVersions.
func (c *FakeProxySQLVersions) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(proxysqlversionsResource, opts))
}

// Create takes the representation of a proxySQLVersion and creates it.  Returns the server's representation of the proxySQLVersion, and an error, if there is any.
func (c *FakeProxySQLVersions) Create(ctx context.Context, proxySQLVersion *v1alpha1.ProxySQLVersion, opts v1.CreateOptions) (result *v1alpha1.ProxySQLVersion, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(proxysqlversionsResource, proxySQLVersion), &v1alpha1.ProxySQLVersion{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ProxySQLVersion), err
}

// Update takes the representation of a proxySQLVersion and updates it. Returns the server's representation of the proxySQLVersion, and an error, if there is any.
func (c *FakeProxySQLVersions) Update(ctx context.Context, proxySQLVersion *v1alpha1.ProxySQLVersion, opts v1.UpdateOptions) (result *v1alpha1.ProxySQLVersion, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(proxysqlversionsResource, proxySQLVersion), &v1alpha1.ProxySQLVersion{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ProxySQLVersion), err
}

// Delete takes name of the proxySQLVersion and deletes it. Returns an error if one occurs.
func (c *FakeProxySQLVersions) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteActionWithOptions(proxysqlversionsResource, name, opts), &v1alpha1.ProxySQLVersion{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeProxySQLVersions) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(proxysqlversionsResource, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.ProxySQLVersionList{})
	return err
}

// Patch applies the patch and returns the patched proxySQLVersion.
func (c *FakeProxySQLVersions) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.ProxySQLVersion, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(proxysqlversionsResource, name, pt, data, subresources...), &v1alpha1.ProxySQLVersion{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ProxySQLVersion), err
}
