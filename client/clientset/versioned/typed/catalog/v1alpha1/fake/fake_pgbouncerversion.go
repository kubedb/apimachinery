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

// FakePgBouncerVersions implements PgBouncerVersionInterface
type FakePgBouncerVersions struct {
	Fake *FakeCatalogV1alpha1
}

var pgbouncerversionsResource = v1alpha1.SchemeGroupVersion.WithResource("pgbouncerversions")

var pgbouncerversionsKind = v1alpha1.SchemeGroupVersion.WithKind("PgBouncerVersion")

// Get takes name of the pgBouncerVersion, and returns the corresponding pgBouncerVersion object, and an error if there is any.
func (c *FakePgBouncerVersions) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.PgBouncerVersion, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(pgbouncerversionsResource, name), &v1alpha1.PgBouncerVersion{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PgBouncerVersion), err
}

// List takes label and field selectors, and returns the list of PgBouncerVersions that match those selectors.
func (c *FakePgBouncerVersions) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.PgBouncerVersionList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(pgbouncerversionsResource, pgbouncerversionsKind, opts), &v1alpha1.PgBouncerVersionList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.PgBouncerVersionList{ListMeta: obj.(*v1alpha1.PgBouncerVersionList).ListMeta}
	for _, item := range obj.(*v1alpha1.PgBouncerVersionList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested pgBouncerVersions.
func (c *FakePgBouncerVersions) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(pgbouncerversionsResource, opts))
}

// Create takes the representation of a pgBouncerVersion and creates it.  Returns the server's representation of the pgBouncerVersion, and an error, if there is any.
func (c *FakePgBouncerVersions) Create(ctx context.Context, pgBouncerVersion *v1alpha1.PgBouncerVersion, opts v1.CreateOptions) (result *v1alpha1.PgBouncerVersion, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(pgbouncerversionsResource, pgBouncerVersion), &v1alpha1.PgBouncerVersion{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PgBouncerVersion), err
}

// Update takes the representation of a pgBouncerVersion and updates it. Returns the server's representation of the pgBouncerVersion, and an error, if there is any.
func (c *FakePgBouncerVersions) Update(ctx context.Context, pgBouncerVersion *v1alpha1.PgBouncerVersion, opts v1.UpdateOptions) (result *v1alpha1.PgBouncerVersion, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(pgbouncerversionsResource, pgBouncerVersion), &v1alpha1.PgBouncerVersion{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PgBouncerVersion), err
}

// Delete takes name of the pgBouncerVersion and deletes it. Returns an error if one occurs.
func (c *FakePgBouncerVersions) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteActionWithOptions(pgbouncerversionsResource, name, opts), &v1alpha1.PgBouncerVersion{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakePgBouncerVersions) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(pgbouncerversionsResource, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.PgBouncerVersionList{})
	return err
}

// Patch applies the patch and returns the patched pgBouncerVersion.
func (c *FakePgBouncerVersions) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.PgBouncerVersion, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(pgbouncerversionsResource, name, pt, data, subresources...), &v1alpha1.PgBouncerVersion{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PgBouncerVersion), err
}
