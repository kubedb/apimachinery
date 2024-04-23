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
	v1alpha2 "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
)

// FakeFerretDBs implements FerretDBInterface
type FakeFerretDBs struct {
	Fake *FakeKubedbV1alpha2
	ns   string
}

var ferretdbsResource = v1alpha2.SchemeGroupVersion.WithResource("ferretdbs")

var ferretdbsKind = v1alpha2.SchemeGroupVersion.WithKind("FerretDB")

// Get takes name of the ferretDB, and returns the corresponding ferretDB object, and an error if there is any.
func (c *FakeFerretDBs) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha2.FerretDB, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(ferretdbsResource, c.ns, name), &v1alpha2.FerretDB{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.FerretDB), err
}

// List takes label and field selectors, and returns the list of FerretDBs that match those selectors.
func (c *FakeFerretDBs) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha2.FerretDBList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(ferretdbsResource, ferretdbsKind, c.ns, opts), &v1alpha2.FerretDBList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha2.FerretDBList{ListMeta: obj.(*v1alpha2.FerretDBList).ListMeta}
	for _, item := range obj.(*v1alpha2.FerretDBList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested ferretDBs.
func (c *FakeFerretDBs) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(ferretdbsResource, c.ns, opts))

}

// Create takes the representation of a ferretDB and creates it.  Returns the server's representation of the ferretDB, and an error, if there is any.
func (c *FakeFerretDBs) Create(ctx context.Context, ferretDB *v1alpha2.FerretDB, opts v1.CreateOptions) (result *v1alpha2.FerretDB, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(ferretdbsResource, c.ns, ferretDB), &v1alpha2.FerretDB{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.FerretDB), err
}

// Update takes the representation of a ferretDB and updates it. Returns the server's representation of the ferretDB, and an error, if there is any.
func (c *FakeFerretDBs) Update(ctx context.Context, ferretDB *v1alpha2.FerretDB, opts v1.UpdateOptions) (result *v1alpha2.FerretDB, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(ferretdbsResource, c.ns, ferretDB), &v1alpha2.FerretDB{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.FerretDB), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeFerretDBs) UpdateStatus(ctx context.Context, ferretDB *v1alpha2.FerretDB, opts v1.UpdateOptions) (*v1alpha2.FerretDB, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(ferretdbsResource, "status", c.ns, ferretDB), &v1alpha2.FerretDB{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.FerretDB), err
}

// Delete takes name of the ferretDB and deletes it. Returns an error if one occurs.
func (c *FakeFerretDBs) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(ferretdbsResource, c.ns, name, opts), &v1alpha2.FerretDB{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeFerretDBs) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(ferretdbsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha2.FerretDBList{})
	return err
}

// Patch applies the patch and returns the patched ferretDB.
func (c *FakeFerretDBs) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha2.FerretDB, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(ferretdbsResource, c.ns, name, pt, data, subresources...), &v1alpha2.FerretDB{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.FerretDB), err
}
