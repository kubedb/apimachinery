/*
Copyright 2017 The KubeDB Authors.

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

package fake

import (
	kubedb "github.com/k8sdb/apimachinery/apis/kubedb"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeMongoDBs implements MongoDBInterface
type FakeMongoDBs struct {
	Fake *FakeKubedb
	ns   string
}

var mongodbsResource = schema.GroupVersionResource{Group: "kubedb.com", Version: "", Resource: "mongodbs"}

var mongodbsKind = schema.GroupVersionKind{Group: "kubedb.com", Version: "", Kind: "MongoDB"}

// Get takes name of the mongoDB, and returns the corresponding mongoDB object, and an error if there is any.
func (c *FakeMongoDBs) Get(name string, options v1.GetOptions) (result *kubedb.MongoDB, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(mongodbsResource, c.ns, name), &kubedb.MongoDB{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kubedb.MongoDB), err
}

// List takes label and field selectors, and returns the list of MongoDBs that match those selectors.
func (c *FakeMongoDBs) List(opts v1.ListOptions) (result *kubedb.MongoDBList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(mongodbsResource, mongodbsKind, c.ns, opts), &kubedb.MongoDBList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &kubedb.MongoDBList{}
	for _, item := range obj.(*kubedb.MongoDBList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested mongoDBs.
func (c *FakeMongoDBs) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(mongodbsResource, c.ns, opts))

}

// Create takes the representation of a mongoDB and creates it.  Returns the server's representation of the mongoDB, and an error, if there is any.
func (c *FakeMongoDBs) Create(mongoDB *kubedb.MongoDB) (result *kubedb.MongoDB, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(mongodbsResource, c.ns, mongoDB), &kubedb.MongoDB{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kubedb.MongoDB), err
}

// Update takes the representation of a mongoDB and updates it. Returns the server's representation of the mongoDB, and an error, if there is any.
func (c *FakeMongoDBs) Update(mongoDB *kubedb.MongoDB) (result *kubedb.MongoDB, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(mongodbsResource, c.ns, mongoDB), &kubedb.MongoDB{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kubedb.MongoDB), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeMongoDBs) UpdateStatus(mongoDB *kubedb.MongoDB) (*kubedb.MongoDB, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(mongodbsResource, "status", c.ns, mongoDB), &kubedb.MongoDB{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kubedb.MongoDB), err
}

// Delete takes name of the mongoDB and deletes it. Returns an error if one occurs.
func (c *FakeMongoDBs) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(mongodbsResource, c.ns, name), &kubedb.MongoDB{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeMongoDBs) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(mongodbsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &kubedb.MongoDBList{})
	return err
}

// Patch applies the patch and returns the patched mongoDB.
func (c *FakeMongoDBs) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *kubedb.MongoDB, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(mongodbsResource, c.ns, name, data, subresources...), &kubedb.MongoDB{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kubedb.MongoDB), err
}
