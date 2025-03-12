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

// FakeZooKeepers implements ZooKeeperInterface
type FakeZooKeepers struct {
	Fake *FakeGitopsV1alpha1
	ns   string
}

var zookeepersResource = v1alpha1.SchemeGroupVersion.WithResource("zookeepers")

var zookeepersKind = v1alpha1.SchemeGroupVersion.WithKind("ZooKeeper")

// Get takes name of the zooKeeper, and returns the corresponding zooKeeper object, and an error if there is any.
func (c *FakeZooKeepers) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.ZooKeeper, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(zookeepersResource, c.ns, name), &v1alpha1.ZooKeeper{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ZooKeeper), err
}

// List takes label and field selectors, and returns the list of ZooKeepers that match those selectors.
func (c *FakeZooKeepers) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.ZooKeeperList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(zookeepersResource, zookeepersKind, c.ns, opts), &v1alpha1.ZooKeeperList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.ZooKeeperList{ListMeta: obj.(*v1alpha1.ZooKeeperList).ListMeta}
	for _, item := range obj.(*v1alpha1.ZooKeeperList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested zooKeepers.
func (c *FakeZooKeepers) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(zookeepersResource, c.ns, opts))

}

// Create takes the representation of a zooKeeper and creates it.  Returns the server's representation of the zooKeeper, and an error, if there is any.
func (c *FakeZooKeepers) Create(ctx context.Context, zooKeeper *v1alpha1.ZooKeeper, opts v1.CreateOptions) (result *v1alpha1.ZooKeeper, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(zookeepersResource, c.ns, zooKeeper), &v1alpha1.ZooKeeper{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ZooKeeper), err
}

// Update takes the representation of a zooKeeper and updates it. Returns the server's representation of the zooKeeper, and an error, if there is any.
func (c *FakeZooKeepers) Update(ctx context.Context, zooKeeper *v1alpha1.ZooKeeper, opts v1.UpdateOptions) (result *v1alpha1.ZooKeeper, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(zookeepersResource, c.ns, zooKeeper), &v1alpha1.ZooKeeper{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ZooKeeper), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeZooKeepers) UpdateStatus(ctx context.Context, zooKeeper *v1alpha1.ZooKeeper, opts v1.UpdateOptions) (*v1alpha1.ZooKeeper, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(zookeepersResource, "status", c.ns, zooKeeper), &v1alpha1.ZooKeeper{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ZooKeeper), err
}

// Delete takes name of the zooKeeper and deletes it. Returns an error if one occurs.
func (c *FakeZooKeepers) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(zookeepersResource, c.ns, name, opts), &v1alpha1.ZooKeeper{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeZooKeepers) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(zookeepersResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.ZooKeeperList{})
	return err
}

// Patch applies the patch and returns the patched zooKeeper.
func (c *FakeZooKeepers) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.ZooKeeper, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(zookeepersResource, c.ns, name, pt, data, subresources...), &v1alpha1.ZooKeeper{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ZooKeeper), err
}
