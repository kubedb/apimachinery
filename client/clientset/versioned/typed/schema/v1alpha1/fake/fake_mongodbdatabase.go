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
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeMongoDBDatabases implements MongoDBDatabaseInterface
type FakeMongoDBDatabases struct {
	Fake *FakeSchemaV1alpha1
	ns   string
}

var mongodbdatabasesResource = v1alpha1.SchemeGroupVersion.WithResource("mongodbdatabases")

var mongodbdatabasesKind = v1alpha1.SchemeGroupVersion.WithKind("MongoDBDatabase")

// Get takes name of the mongoDBDatabase, and returns the corresponding mongoDBDatabase object, and an error if there is any.
func (c *FakeMongoDBDatabases) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.MongoDBDatabase, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(mongodbdatabasesResource, c.ns, name), &v1alpha1.MongoDBDatabase{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MongoDBDatabase), err
}

// List takes label and field selectors, and returns the list of MongoDBDatabases that match those selectors.
func (c *FakeMongoDBDatabases) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.MongoDBDatabaseList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(mongodbdatabasesResource, mongodbdatabasesKind, c.ns, opts), &v1alpha1.MongoDBDatabaseList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.MongoDBDatabaseList{ListMeta: obj.(*v1alpha1.MongoDBDatabaseList).ListMeta}
	for _, item := range obj.(*v1alpha1.MongoDBDatabaseList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested mongoDBDatabases.
func (c *FakeMongoDBDatabases) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(mongodbdatabasesResource, c.ns, opts))

}

// Create takes the representation of a mongoDBDatabase and creates it.  Returns the server's representation of the mongoDBDatabase, and an error, if there is any.
func (c *FakeMongoDBDatabases) Create(ctx context.Context, mongoDBDatabase *v1alpha1.MongoDBDatabase, opts v1.CreateOptions) (result *v1alpha1.MongoDBDatabase, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(mongodbdatabasesResource, c.ns, mongoDBDatabase), &v1alpha1.MongoDBDatabase{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MongoDBDatabase), err
}

// Update takes the representation of a mongoDBDatabase and updates it. Returns the server's representation of the mongoDBDatabase, and an error, if there is any.
func (c *FakeMongoDBDatabases) Update(ctx context.Context, mongoDBDatabase *v1alpha1.MongoDBDatabase, opts v1.UpdateOptions) (result *v1alpha1.MongoDBDatabase, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(mongodbdatabasesResource, c.ns, mongoDBDatabase), &v1alpha1.MongoDBDatabase{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MongoDBDatabase), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeMongoDBDatabases) UpdateStatus(ctx context.Context, mongoDBDatabase *v1alpha1.MongoDBDatabase, opts v1.UpdateOptions) (*v1alpha1.MongoDBDatabase, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(mongodbdatabasesResource, "status", c.ns, mongoDBDatabase), &v1alpha1.MongoDBDatabase{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MongoDBDatabase), err
}

// Delete takes name of the mongoDBDatabase and deletes it. Returns an error if one occurs.
func (c *FakeMongoDBDatabases) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(mongodbdatabasesResource, c.ns, name, opts), &v1alpha1.MongoDBDatabase{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeMongoDBDatabases) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(mongodbdatabasesResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.MongoDBDatabaseList{})
	return err
}

// Patch applies the patch and returns the patched mongoDBDatabase.
func (c *FakeMongoDBDatabases) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.MongoDBDatabase, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(mongodbdatabasesResource, c.ns, name, pt, data, subresources...), &v1alpha1.MongoDBDatabase{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MongoDBDatabase), err
}
