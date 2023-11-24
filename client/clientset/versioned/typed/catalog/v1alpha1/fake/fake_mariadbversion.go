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

// FakeMariaDBVersions implements MariaDBVersionInterface
type FakeMariaDBVersions struct {
	Fake *FakeCatalogV1alpha1
}

var mariadbversionsResource = schema.GroupVersionResource{Group: "catalog.kubedb.com", Version: "v1alpha1", Resource: "mariadbversions"}

var mariadbversionsKind = schema.GroupVersionKind{Group: "catalog.kubedb.com", Version: "v1alpha1", Kind: "MariaDBVersion"}

// Get takes name of the mariaDBVersion, and returns the corresponding mariaDBVersion object, and an error if there is any.
func (c *FakeMariaDBVersions) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.MariaDBVersion, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(mariadbversionsResource, name), &v1alpha1.MariaDBVersion{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MariaDBVersion), err
}

// List takes label and field selectors, and returns the list of MariaDBVersions that match those selectors.
func (c *FakeMariaDBVersions) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.MariaDBVersionList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(mariadbversionsResource, mariadbversionsKind, opts), &v1alpha1.MariaDBVersionList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.MariaDBVersionList{ListMeta: obj.(*v1alpha1.MariaDBVersionList).ListMeta}
	for _, item := range obj.(*v1alpha1.MariaDBVersionList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested mariaDBVersions.
func (c *FakeMariaDBVersions) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(mariadbversionsResource, opts))
}

// Create takes the representation of a mariaDBVersion and creates it.  Returns the server's representation of the mariaDBVersion, and an error, if there is any.
func (c *FakeMariaDBVersions) Create(ctx context.Context, mariaDBVersion *v1alpha1.MariaDBVersion, opts v1.CreateOptions) (result *v1alpha1.MariaDBVersion, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(mariadbversionsResource, mariaDBVersion), &v1alpha1.MariaDBVersion{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MariaDBVersion), err
}

// Update takes the representation of a mariaDBVersion and updates it. Returns the server's representation of the mariaDBVersion, and an error, if there is any.
func (c *FakeMariaDBVersions) Update(ctx context.Context, mariaDBVersion *v1alpha1.MariaDBVersion, opts v1.UpdateOptions) (result *v1alpha1.MariaDBVersion, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(mariadbversionsResource, mariaDBVersion), &v1alpha1.MariaDBVersion{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MariaDBVersion), err
}

// Delete takes name of the mariaDBVersion and deletes it. Returns an error if one occurs.
func (c *FakeMariaDBVersions) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteActionWithOptions(mariadbversionsResource, name, opts), &v1alpha1.MariaDBVersion{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeMariaDBVersions) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(mariadbversionsResource, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.MariaDBVersionList{})
	return err
}

// Patch applies the patch and returns the patched mariaDBVersion.
func (c *FakeMariaDBVersions) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.MariaDBVersion, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(mariadbversionsResource, name, pt, data, subresources...), &v1alpha1.MariaDBVersion{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MariaDBVersion), err
}
