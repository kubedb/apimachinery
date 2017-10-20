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

// FakeDormantDatabases implements DormantDatabaseInterface
type FakeDormantDatabases struct {
	Fake *FakeKubedb
	ns   string
}

var dormantdatabasesResource = schema.GroupVersionResource{Group: "kubedb.com", Version: "", Resource: "dormantdatabases"}

var dormantdatabasesKind = schema.GroupVersionKind{Group: "kubedb.com", Version: "", Kind: "DormantDatabase"}

// Get takes name of the dormantDatabase, and returns the corresponding dormantDatabase object, and an error if there is any.
func (c *FakeDormantDatabases) Get(name string, options v1.GetOptions) (result *kubedb.DormantDatabase, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(dormantdatabasesResource, c.ns, name), &kubedb.DormantDatabase{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kubedb.DormantDatabase), err
}

// List takes label and field selectors, and returns the list of DormantDatabases that match those selectors.
func (c *FakeDormantDatabases) List(opts v1.ListOptions) (result *kubedb.DormantDatabaseList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(dormantdatabasesResource, dormantdatabasesKind, c.ns, opts), &kubedb.DormantDatabaseList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &kubedb.DormantDatabaseList{}
	for _, item := range obj.(*kubedb.DormantDatabaseList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested dormantDatabases.
func (c *FakeDormantDatabases) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(dormantdatabasesResource, c.ns, opts))

}

// Create takes the representation of a dormantDatabase and creates it.  Returns the server's representation of the dormantDatabase, and an error, if there is any.
func (c *FakeDormantDatabases) Create(dormantDatabase *kubedb.DormantDatabase) (result *kubedb.DormantDatabase, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(dormantdatabasesResource, c.ns, dormantDatabase), &kubedb.DormantDatabase{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kubedb.DormantDatabase), err
}

// Update takes the representation of a dormantDatabase and updates it. Returns the server's representation of the dormantDatabase, and an error, if there is any.
func (c *FakeDormantDatabases) Update(dormantDatabase *kubedb.DormantDatabase) (result *kubedb.DormantDatabase, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(dormantdatabasesResource, c.ns, dormantDatabase), &kubedb.DormantDatabase{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kubedb.DormantDatabase), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeDormantDatabases) UpdateStatus(dormantDatabase *kubedb.DormantDatabase) (*kubedb.DormantDatabase, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(dormantdatabasesResource, "status", c.ns, dormantDatabase), &kubedb.DormantDatabase{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kubedb.DormantDatabase), err
}

// Delete takes name of the dormantDatabase and deletes it. Returns an error if one occurs.
func (c *FakeDormantDatabases) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(dormantdatabasesResource, c.ns, name), &kubedb.DormantDatabase{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeDormantDatabases) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(dormantdatabasesResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &kubedb.DormantDatabaseList{})
	return err
}

// Patch applies the patch and returns the patched dormantDatabase.
func (c *FakeDormantDatabases) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *kubedb.DormantDatabase, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(dormantdatabasesResource, c.ns, name, data, subresources...), &kubedb.DormantDatabase{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kubedb.DormantDatabase), err
}
