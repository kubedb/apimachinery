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
	v1alpha1 "kubedb.dev/apimachinery/apis/postgres/v1alpha1"
)

// FakeSubscribers implements SubscriberInterface
type FakeSubscribers struct {
	Fake *FakePostgresV1alpha1
	ns   string
}

var subscribersResource = v1alpha1.SchemeGroupVersion.WithResource("subscribers")

var subscribersKind = v1alpha1.SchemeGroupVersion.WithKind("Subscriber")

// Get takes name of the subscriber, and returns the corresponding subscriber object, and an error if there is any.
func (c *FakeSubscribers) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.Subscriber, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(subscribersResource, c.ns, name), &v1alpha1.Subscriber{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Subscriber), err
}

// List takes label and field selectors, and returns the list of Subscribers that match those selectors.
func (c *FakeSubscribers) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.SubscriberList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(subscribersResource, subscribersKind, c.ns, opts), &v1alpha1.SubscriberList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.SubscriberList{ListMeta: obj.(*v1alpha1.SubscriberList).ListMeta}
	for _, item := range obj.(*v1alpha1.SubscriberList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested subscribers.
func (c *FakeSubscribers) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(subscribersResource, c.ns, opts))

}

// Create takes the representation of a subscriber and creates it.  Returns the server's representation of the subscriber, and an error, if there is any.
func (c *FakeSubscribers) Create(ctx context.Context, subscriber *v1alpha1.Subscriber, opts v1.CreateOptions) (result *v1alpha1.Subscriber, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(subscribersResource, c.ns, subscriber), &v1alpha1.Subscriber{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Subscriber), err
}

// Update takes the representation of a subscriber and updates it. Returns the server's representation of the subscriber, and an error, if there is any.
func (c *FakeSubscribers) Update(ctx context.Context, subscriber *v1alpha1.Subscriber, opts v1.UpdateOptions) (result *v1alpha1.Subscriber, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(subscribersResource, c.ns, subscriber), &v1alpha1.Subscriber{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Subscriber), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeSubscribers) UpdateStatus(ctx context.Context, subscriber *v1alpha1.Subscriber, opts v1.UpdateOptions) (*v1alpha1.Subscriber, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(subscribersResource, "status", c.ns, subscriber), &v1alpha1.Subscriber{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Subscriber), err
}

// Delete takes name of the subscriber and deletes it. Returns an error if one occurs.
func (c *FakeSubscribers) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(subscribersResource, c.ns, name, opts), &v1alpha1.Subscriber{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeSubscribers) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(subscribersResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.SubscriberList{})
	return err
}

// Patch applies the patch and returns the patched subscriber.
func (c *FakeSubscribers) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Subscriber, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(subscribersResource, c.ns, name, pt, data, subresources...), &v1alpha1.Subscriber{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Subscriber), err
}
