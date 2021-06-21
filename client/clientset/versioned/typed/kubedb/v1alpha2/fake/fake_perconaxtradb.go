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

// FakePerconaXtraDBs implements PerconaXtraDBInterface
type FakePerconaXtraDBs struct {
	Fake *FakeKubedbV1alpha2
	ns   string
}

var perconaxtradbsResource = schema.GroupVersionResource{Group: "kubedb.com", Version: "v1alpha2", Resource: "perconaxtradbs"}

var perconaxtradbsKind = schema.GroupVersionKind{Group: "kubedb.com", Version: "v1alpha2", Kind: "PerconaXtraDB"}

// Get takes name of the perconaXtraDB, and returns the corresponding perconaXtraDB object, and an error if there is any.
func (c *FakePerconaXtraDBs) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha2.PerconaXtraDB, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(perconaxtradbsResource, c.ns, name), &v1alpha2.PerconaXtraDB{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.PerconaXtraDB), err
}

// List takes label and field selectors, and returns the list of PerconaXtraDBs that match those selectors.
func (c *FakePerconaXtraDBs) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha2.PerconaXtraDBList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(perconaxtradbsResource, perconaxtradbsKind, c.ns, opts), &v1alpha2.PerconaXtraDBList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha2.PerconaXtraDBList{ListMeta: obj.(*v1alpha2.PerconaXtraDBList).ListMeta}
	for _, item := range obj.(*v1alpha2.PerconaXtraDBList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested perconaXtraDBs.
func (c *FakePerconaXtraDBs) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(perconaxtradbsResource, c.ns, opts))

}

// Create takes the representation of a perconaXtraDB and creates it.  Returns the server's representation of the perconaXtraDB, and an error, if there is any.
func (c *FakePerconaXtraDBs) Create(ctx context.Context, perconaXtraDB *v1alpha2.PerconaXtraDB, opts v1.CreateOptions) (result *v1alpha2.PerconaXtraDB, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(perconaxtradbsResource, c.ns, perconaXtraDB), &v1alpha2.PerconaXtraDB{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.PerconaXtraDB), err
}

// Update takes the representation of a perconaXtraDB and updates it. Returns the server's representation of the perconaXtraDB, and an error, if there is any.
func (c *FakePerconaXtraDBs) Update(ctx context.Context, perconaXtraDB *v1alpha2.PerconaXtraDB, opts v1.UpdateOptions) (result *v1alpha2.PerconaXtraDB, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(perconaxtradbsResource, c.ns, perconaXtraDB), &v1alpha2.PerconaXtraDB{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.PerconaXtraDB), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakePerconaXtraDBs) UpdateStatus(ctx context.Context, perconaXtraDB *v1alpha2.PerconaXtraDB, opts v1.UpdateOptions) (*v1alpha2.PerconaXtraDB, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(perconaxtradbsResource, "status", c.ns, perconaXtraDB), &v1alpha2.PerconaXtraDB{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.PerconaXtraDB), err
}

// Delete takes name of the perconaXtraDB and deletes it. Returns an error if one occurs.
func (c *FakePerconaXtraDBs) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(perconaxtradbsResource, c.ns, name), &v1alpha2.PerconaXtraDB{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakePerconaXtraDBs) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(perconaxtradbsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha2.PerconaXtraDBList{})
	return err
}

// Patch applies the patch and returns the patched perconaXtraDB.
func (c *FakePerconaXtraDBs) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha2.PerconaXtraDB, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(perconaxtradbsResource, c.ns, name, pt, data, subresources...), &v1alpha2.PerconaXtraDB{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.PerconaXtraDB), err
}
