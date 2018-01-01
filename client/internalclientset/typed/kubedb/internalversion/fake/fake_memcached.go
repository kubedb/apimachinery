/*
Copyright 2018 The KubeDB Authors.

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
	kubedb "github.com/kubedb/apimachinery/apis/kubedb"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeMemcacheds implements MemcachedInterface
type FakeMemcacheds struct {
	Fake *FakeKubedb
	ns   string
}

var memcachedsResource = schema.GroupVersionResource{Group: "kubedb.com", Version: "", Resource: "memcacheds"}

var memcachedsKind = schema.GroupVersionKind{Group: "kubedb.com", Version: "", Kind: "Memcached"}

// Get takes name of the memcached, and returns the corresponding memcached object, and an error if there is any.
func (c *FakeMemcacheds) Get(name string, options v1.GetOptions) (result *kubedb.Memcached, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(memcachedsResource, c.ns, name), &kubedb.Memcached{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kubedb.Memcached), err
}

// List takes label and field selectors, and returns the list of Memcacheds that match those selectors.
func (c *FakeMemcacheds) List(opts v1.ListOptions) (result *kubedb.MemcachedList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(memcachedsResource, memcachedsKind, c.ns, opts), &kubedb.MemcachedList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &kubedb.MemcachedList{}
	for _, item := range obj.(*kubedb.MemcachedList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested memcacheds.
func (c *FakeMemcacheds) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(memcachedsResource, c.ns, opts))

}

// Create takes the representation of a memcached and creates it.  Returns the server's representation of the memcached, and an error, if there is any.
func (c *FakeMemcacheds) Create(memcached *kubedb.Memcached) (result *kubedb.Memcached, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(memcachedsResource, c.ns, memcached), &kubedb.Memcached{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kubedb.Memcached), err
}

// Update takes the representation of a memcached and updates it. Returns the server's representation of the memcached, and an error, if there is any.
func (c *FakeMemcacheds) Update(memcached *kubedb.Memcached) (result *kubedb.Memcached, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(memcachedsResource, c.ns, memcached), &kubedb.Memcached{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kubedb.Memcached), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeMemcacheds) UpdateStatus(memcached *kubedb.Memcached) (*kubedb.Memcached, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(memcachedsResource, "status", c.ns, memcached), &kubedb.Memcached{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kubedb.Memcached), err
}

// Delete takes name of the memcached and deletes it. Returns an error if one occurs.
func (c *FakeMemcacheds) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(memcachedsResource, c.ns, name), &kubedb.Memcached{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeMemcacheds) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(memcachedsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &kubedb.MemcachedList{})
	return err
}

// Patch applies the patch and returns the patched memcached.
func (c *FakeMemcacheds) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *kubedb.Memcached, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(memcachedsResource, c.ns, name, data, subresources...), &kubedb.Memcached{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kubedb.Memcached), err
}
