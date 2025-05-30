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

	v1alpha1 "kubedb.dev/apimachinery/apis/gitops/v1alpha1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeMariaDBs implements MariaDBInterface
type FakeMariaDBs struct {
	Fake *FakeGitopsV1alpha1
	ns   string
}

var mariadbsResource = v1alpha1.SchemeGroupVersion.WithResource("mariadbs")

var mariadbsKind = v1alpha1.SchemeGroupVersion.WithKind("MariaDB")

// Get takes name of the mariaDB, and returns the corresponding mariaDB object, and an error if there is any.
func (c *FakeMariaDBs) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.MariaDB, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(mariadbsResource, c.ns, name), &v1alpha1.MariaDB{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MariaDB), err
}

// List takes label and field selectors, and returns the list of MariaDBs that match those selectors.
func (c *FakeMariaDBs) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.MariaDBList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(mariadbsResource, mariadbsKind, c.ns, opts), &v1alpha1.MariaDBList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.MariaDBList{ListMeta: obj.(*v1alpha1.MariaDBList).ListMeta}
	for _, item := range obj.(*v1alpha1.MariaDBList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested mariaDBs.
func (c *FakeMariaDBs) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(mariadbsResource, c.ns, opts))

}

// Create takes the representation of a mariaDB and creates it.  Returns the server's representation of the mariaDB, and an error, if there is any.
func (c *FakeMariaDBs) Create(ctx context.Context, mariaDB *v1alpha1.MariaDB, opts v1.CreateOptions) (result *v1alpha1.MariaDB, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(mariadbsResource, c.ns, mariaDB), &v1alpha1.MariaDB{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MariaDB), err
}

// Update takes the representation of a mariaDB and updates it. Returns the server's representation of the mariaDB, and an error, if there is any.
func (c *FakeMariaDBs) Update(ctx context.Context, mariaDB *v1alpha1.MariaDB, opts v1.UpdateOptions) (result *v1alpha1.MariaDB, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(mariadbsResource, c.ns, mariaDB), &v1alpha1.MariaDB{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MariaDB), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeMariaDBs) UpdateStatus(ctx context.Context, mariaDB *v1alpha1.MariaDB, opts v1.UpdateOptions) (*v1alpha1.MariaDB, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(mariadbsResource, "status", c.ns, mariaDB), &v1alpha1.MariaDB{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MariaDB), err
}

// Delete takes name of the mariaDB and deletes it. Returns an error if one occurs.
func (c *FakeMariaDBs) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(mariadbsResource, c.ns, name, opts), &v1alpha1.MariaDB{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeMariaDBs) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(mariadbsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.MariaDBList{})
	return err
}

// Patch applies the patch and returns the patched mariaDB.
func (c *FakeMariaDBs) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.MariaDB, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(mariadbsResource, c.ns, name, pt, data, subresources...), &v1alpha1.MariaDB{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MariaDB), err
}
