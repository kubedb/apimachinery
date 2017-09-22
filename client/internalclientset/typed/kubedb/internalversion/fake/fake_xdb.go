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

// FakeXdbs implements XdbInterface
type FakeXdbs struct {
	Fake *FakeKubedb
	ns   string
}

var xdbsResource = schema.GroupVersionResource{Group: "kubedb.com", Version: "", Resource: "xdbs"}

var xdbsKind = schema.GroupVersionKind{Group: "kubedb.com", Version: "", Kind: "Xdb"}

// Get takes name of the xdb, and returns the corresponding xdb object, and an error if there is any.
func (c *FakeXdbs) Get(name string, options v1.GetOptions) (result *kubedb.Xdb, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(xdbsResource, c.ns, name), &kubedb.Xdb{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kubedb.Xdb), err
}

// List takes label and field selectors, and returns the list of Xdbs that match those selectors.
func (c *FakeXdbs) List(opts v1.ListOptions) (result *kubedb.XdbList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(xdbsResource, xdbsKind, c.ns, opts), &kubedb.XdbList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &kubedb.XdbList{}
	for _, item := range obj.(*kubedb.XdbList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested xdbs.
func (c *FakeXdbs) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(xdbsResource, c.ns, opts))

}

// Create takes the representation of a xdb and creates it.  Returns the server's representation of the xdb, and an error, if there is any.
func (c *FakeXdbs) Create(xdb *kubedb.Xdb) (result *kubedb.Xdb, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(xdbsResource, c.ns, xdb), &kubedb.Xdb{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kubedb.Xdb), err
}

// Update takes the representation of a xdb and updates it. Returns the server's representation of the xdb, and an error, if there is any.
func (c *FakeXdbs) Update(xdb *kubedb.Xdb) (result *kubedb.Xdb, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(xdbsResource, c.ns, xdb), &kubedb.Xdb{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kubedb.Xdb), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeXdbs) UpdateStatus(xdb *kubedb.Xdb) (*kubedb.Xdb, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(xdbsResource, "status", c.ns, xdb), &kubedb.Xdb{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kubedb.Xdb), err
}

// Delete takes name of the xdb and deletes it. Returns an error if one occurs.
func (c *FakeXdbs) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(xdbsResource, c.ns, name), &kubedb.Xdb{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeXdbs) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(xdbsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &kubedb.XdbList{})
	return err
}

// Patch applies the patch and returns the patched xdb.
func (c *FakeXdbs) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *kubedb.Xdb, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(xdbsResource, c.ns, name, data, subresources...), &kubedb.Xdb{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kubedb.Xdb), err
}
