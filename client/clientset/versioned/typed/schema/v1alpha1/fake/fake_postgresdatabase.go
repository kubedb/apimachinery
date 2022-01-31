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

	v1alpha1 "kubedb.dev/apimachinery/apis/schema/v1alpha1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakePostgresDatabases implements PostgresDatabaseInterface
type FakePostgresDatabases struct {
	Fake *FakeSchemaV1alpha1
	ns   string
}

var postgresdatabasesResource = schema.GroupVersionResource{Group: "schema.kubedb.com", Version: "v1alpha1", Resource: "postgresdatabases"}

var postgresdatabasesKind = schema.GroupVersionKind{Group: "schema.kubedb.com", Version: "v1alpha1", Kind: "PostgresDatabase"}

// Get takes name of the postgresDatabase, and returns the corresponding postgresDatabase object, and an error if there is any.
func (c *FakePostgresDatabases) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.PostgresDatabase, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(postgresdatabasesResource, c.ns, name), &v1alpha1.PostgresDatabase{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PostgresDatabase), err
}

// List takes label and field selectors, and returns the list of PostgresDatabases that match those selectors.
func (c *FakePostgresDatabases) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.PostgresDatabaseList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(postgresdatabasesResource, postgresdatabasesKind, c.ns, opts), &v1alpha1.PostgresDatabaseList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.PostgresDatabaseList{ListMeta: obj.(*v1alpha1.PostgresDatabaseList).ListMeta}
	for _, item := range obj.(*v1alpha1.PostgresDatabaseList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested postgresDatabases.
func (c *FakePostgresDatabases) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(postgresdatabasesResource, c.ns, opts))

}

// Create takes the representation of a postgresDatabase and creates it.  Returns the server's representation of the postgresDatabase, and an error, if there is any.
func (c *FakePostgresDatabases) Create(ctx context.Context, postgresDatabase *v1alpha1.PostgresDatabase, opts v1.CreateOptions) (result *v1alpha1.PostgresDatabase, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(postgresdatabasesResource, c.ns, postgresDatabase), &v1alpha1.PostgresDatabase{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PostgresDatabase), err
}

// Update takes the representation of a postgresDatabase and updates it. Returns the server's representation of the postgresDatabase, and an error, if there is any.
func (c *FakePostgresDatabases) Update(ctx context.Context, postgresDatabase *v1alpha1.PostgresDatabase, opts v1.UpdateOptions) (result *v1alpha1.PostgresDatabase, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(postgresdatabasesResource, c.ns, postgresDatabase), &v1alpha1.PostgresDatabase{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PostgresDatabase), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakePostgresDatabases) UpdateStatus(ctx context.Context, postgresDatabase *v1alpha1.PostgresDatabase, opts v1.UpdateOptions) (*v1alpha1.PostgresDatabase, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(postgresdatabasesResource, "status", c.ns, postgresDatabase), &v1alpha1.PostgresDatabase{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PostgresDatabase), err
}

// Delete takes name of the postgresDatabase and deletes it. Returns an error if one occurs.
func (c *FakePostgresDatabases) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(postgresdatabasesResource, c.ns, name), &v1alpha1.PostgresDatabase{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakePostgresDatabases) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(postgresdatabasesResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.PostgresDatabaseList{})
	return err
}

// Patch applies the patch and returns the patched postgresDatabase.
func (c *FakePostgresDatabases) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.PostgresDatabase, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(postgresdatabasesResource, c.ns, name, pt, data, subresources...), &v1alpha1.PostgresDatabase{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PostgresDatabase), err
}