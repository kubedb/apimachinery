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
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeDruidVersions implements DruidVersionInterface
type FakeDruidVersions struct {
	Fake *FakeCatalogV1alpha1
}

var druidversionsResource = v1alpha1.SchemeGroupVersion.WithResource("druidversions")

var druidversionsKind = v1alpha1.SchemeGroupVersion.WithKind("DruidVersion")

// Get takes name of the druidVersion, and returns the corresponding druidVersion object, and an error if there is any.
func (c *FakeDruidVersions) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.DruidVersion, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(druidversionsResource, name), &v1alpha1.DruidVersion{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DruidVersion), err
}

// List takes label and field selectors, and returns the list of DruidVersions that match those selectors.
func (c *FakeDruidVersions) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.DruidVersionList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(druidversionsResource, druidversionsKind, opts), &v1alpha1.DruidVersionList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.DruidVersionList{ListMeta: obj.(*v1alpha1.DruidVersionList).ListMeta}
	for _, item := range obj.(*v1alpha1.DruidVersionList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested druidVersions.
func (c *FakeDruidVersions) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(druidversionsResource, opts))
}

// Create takes the representation of a druidVersion and creates it.  Returns the server's representation of the druidVersion, and an error, if there is any.
func (c *FakeDruidVersions) Create(ctx context.Context, druidVersion *v1alpha1.DruidVersion, opts v1.CreateOptions) (result *v1alpha1.DruidVersion, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(druidversionsResource, druidVersion), &v1alpha1.DruidVersion{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DruidVersion), err
}

// Update takes the representation of a druidVersion and updates it. Returns the server's representation of the druidVersion, and an error, if there is any.
func (c *FakeDruidVersions) Update(ctx context.Context, druidVersion *v1alpha1.DruidVersion, opts v1.UpdateOptions) (result *v1alpha1.DruidVersion, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(druidversionsResource, druidVersion), &v1alpha1.DruidVersion{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DruidVersion), err
}

// Delete takes name of the druidVersion and deletes it. Returns an error if one occurs.
func (c *FakeDruidVersions) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteActionWithOptions(druidversionsResource, name, opts), &v1alpha1.DruidVersion{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeDruidVersions) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(druidversionsResource, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.DruidVersionList{})
	return err
}

// Patch applies the patch and returns the patched druidVersion.
func (c *FakeDruidVersions) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.DruidVersion, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(druidversionsResource, name, pt, data, subresources...), &v1alpha1.DruidVersion{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DruidVersion), err
}
