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
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeRabbitmqs implements RabbitmqInterface
type FakeRabbitmqs struct {
	Fake *FakeKubedbV1alpha2
	ns   string
}

var rabbitmqsResource = v1alpha2.SchemeGroupVersion.WithResource("rabbitmqs")

var rabbitmqsKind = v1alpha2.SchemeGroupVersion.WithKind("Rabbitmq")

// Get takes name of the rabbitmq, and returns the corresponding rabbitmq object, and an error if there is any.
func (c *FakeRabbitmqs) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha2.Rabbitmq, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(rabbitmqsResource, c.ns, name), &v1alpha2.Rabbitmq{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.Rabbitmq), err
}

// List takes label and field selectors, and returns the list of Rabbitmqs that match those selectors.
func (c *FakeRabbitmqs) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha2.RabbitmqList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(rabbitmqsResource, rabbitmqsKind, c.ns, opts), &v1alpha2.RabbitmqList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha2.RabbitmqList{ListMeta: obj.(*v1alpha2.RabbitmqList).ListMeta}
	for _, item := range obj.(*v1alpha2.RabbitmqList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested rabbitmqs.
func (c *FakeRabbitmqs) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(rabbitmqsResource, c.ns, opts))

}

// Create takes the representation of a rabbitmq and creates it.  Returns the server's representation of the rabbitmq, and an error, if there is any.
func (c *FakeRabbitmqs) Create(ctx context.Context, rabbitmq *v1alpha2.Rabbitmq, opts v1.CreateOptions) (result *v1alpha2.Rabbitmq, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(rabbitmqsResource, c.ns, rabbitmq), &v1alpha2.Rabbitmq{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.Rabbitmq), err
}

// Update takes the representation of a rabbitmq and updates it. Returns the server's representation of the rabbitmq, and an error, if there is any.
func (c *FakeRabbitmqs) Update(ctx context.Context, rabbitmq *v1alpha2.Rabbitmq, opts v1.UpdateOptions) (result *v1alpha2.Rabbitmq, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(rabbitmqsResource, c.ns, rabbitmq), &v1alpha2.Rabbitmq{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.Rabbitmq), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeRabbitmqs) UpdateStatus(ctx context.Context, rabbitmq *v1alpha2.Rabbitmq, opts v1.UpdateOptions) (*v1alpha2.Rabbitmq, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(rabbitmqsResource, "status", c.ns, rabbitmq), &v1alpha2.Rabbitmq{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.Rabbitmq), err
}

// Delete takes name of the rabbitmq and deletes it. Returns an error if one occurs.
func (c *FakeRabbitmqs) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(rabbitmqsResource, c.ns, name, opts), &v1alpha2.Rabbitmq{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeRabbitmqs) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(rabbitmqsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha2.RabbitmqList{})
	return err
}

// Patch applies the patch and returns the patched rabbitmq.
func (c *FakeRabbitmqs) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha2.Rabbitmq, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(rabbitmqsResource, c.ns, name, pt, data, subresources...), &v1alpha2.Rabbitmq{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.Rabbitmq), err
}
