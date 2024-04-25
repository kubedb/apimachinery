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

// FakePgpoolVersions implements PgpoolVersionInterface
type FakePgpoolVersions struct {
	Fake *FakeCatalogV1alpha1
}

var pgpoolversionsResource = v1alpha1.SchemeGroupVersion.WithResource("pgpoolversions")

var pgpoolversionsKind = v1alpha1.SchemeGroupVersion.WithKind("PgpoolVersion")

// Get takes name of the pgpoolVersion, and returns the corresponding pgpoolVersion object, and an error if there is any.
func (c *FakePgpoolVersions) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.PgpoolVersion, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(pgpoolversionsResource, name), &v1alpha1.PgpoolVersion{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PgpoolVersion), err
}

// List takes label and field selectors, and returns the list of PgpoolVersions that match those selectors.
func (c *FakePgpoolVersions) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.PgpoolVersionList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(pgpoolversionsResource, pgpoolversionsKind, opts), &v1alpha1.PgpoolVersionList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.PgpoolVersionList{ListMeta: obj.(*v1alpha1.PgpoolVersionList).ListMeta}
	for _, item := range obj.(*v1alpha1.PgpoolVersionList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested pgpoolVersions.
func (c *FakePgpoolVersions) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(pgpoolversionsResource, opts))
}

// Create takes the representation of a pgpoolVersion and creates it.  Returns the server's representation of the pgpoolVersion, and an error, if there is any.
func (c *FakePgpoolVersions) Create(ctx context.Context, pgpoolVersion *v1alpha1.PgpoolVersion, opts v1.CreateOptions) (result *v1alpha1.PgpoolVersion, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(pgpoolversionsResource, pgpoolVersion), &v1alpha1.PgpoolVersion{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PgpoolVersion), err
}

// Update takes the representation of a pgpoolVersion and updates it. Returns the server's representation of the pgpoolVersion, and an error, if there is any.
func (c *FakePgpoolVersions) Update(ctx context.Context, pgpoolVersion *v1alpha1.PgpoolVersion, opts v1.UpdateOptions) (result *v1alpha1.PgpoolVersion, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(pgpoolversionsResource, pgpoolVersion), &v1alpha1.PgpoolVersion{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PgpoolVersion), err
}

// Delete takes name of the pgpoolVersion and deletes it. Returns an error if one occurs.
func (c *FakePgpoolVersions) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteActionWithOptions(pgpoolversionsResource, name, opts), &v1alpha1.PgpoolVersion{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakePgpoolVersions) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(pgpoolversionsResource, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.PgpoolVersionList{})
	return err
}

// Patch applies the patch and returns the patched pgpoolVersion.
func (c *FakePgpoolVersions) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.PgpoolVersion, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(pgpoolversionsResource, name, pt, data, subresources...), &v1alpha1.PgpoolVersion{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PgpoolVersion), err
}
