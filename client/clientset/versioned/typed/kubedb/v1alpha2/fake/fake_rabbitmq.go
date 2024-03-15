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

// FakeRabbitMQs implements RabbitMQInterface
type FakeRabbitMQs struct {
	Fake *FakeKubedbV1alpha2
	ns   string
}

var rabbitmqsResource = v1alpha2.SchemeGroupVersion.WithResource("rabbitmqs")

var rabbitmqsKind = v1alpha2.SchemeGroupVersion.WithKind("RabbitMQ")

// Get takes name of the rabbitMQ, and returns the corresponding rabbitMQ object, and an error if there is any.
func (c *FakeRabbitMQs) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha2.RabbitMQ, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(rabbitmqsResource, c.ns, name), &v1alpha2.RabbitMQ{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.RabbitMQ), err
}

// List takes label and field selectors, and returns the list of RabbitMQs that match those selectors.
func (c *FakeRabbitMQs) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha2.RabbitMQList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(rabbitmqsResource, rabbitmqsKind, c.ns, opts), &v1alpha2.RabbitMQList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha2.RabbitMQList{ListMeta: obj.(*v1alpha2.RabbitMQList).ListMeta}
	for _, item := range obj.(*v1alpha2.RabbitMQList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested rabbitMQs.
func (c *FakeRabbitMQs) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(rabbitmqsResource, c.ns, opts))

}

// Create takes the representation of a rabbitMQ and creates it.  Returns the server's representation of the rabbitMQ, and an error, if there is any.
func (c *FakeRabbitMQs) Create(ctx context.Context, rabbitMQ *v1alpha2.RabbitMQ, opts v1.CreateOptions) (result *v1alpha2.RabbitMQ, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(rabbitmqsResource, c.ns, rabbitMQ), &v1alpha2.RabbitMQ{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.RabbitMQ), err
}

// Update takes the representation of a rabbitMQ and updates it. Returns the server's representation of the rabbitMQ, and an error, if there is any.
func (c *FakeRabbitMQs) Update(ctx context.Context, rabbitMQ *v1alpha2.RabbitMQ, opts v1.UpdateOptions) (result *v1alpha2.RabbitMQ, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(rabbitmqsResource, c.ns, rabbitMQ), &v1alpha2.RabbitMQ{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.RabbitMQ), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeRabbitMQs) UpdateStatus(ctx context.Context, rabbitMQ *v1alpha2.RabbitMQ, opts v1.UpdateOptions) (*v1alpha2.RabbitMQ, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(rabbitmqsResource, "status", c.ns, rabbitMQ), &v1alpha2.RabbitMQ{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.RabbitMQ), err
}

// Delete takes name of the rabbitMQ and deletes it. Returns an error if one occurs.
func (c *FakeRabbitMQs) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(rabbitmqsResource, c.ns, name, opts), &v1alpha2.RabbitMQ{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeRabbitMQs) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(rabbitmqsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha2.RabbitMQList{})
	return err
}

// Patch applies the patch and returns the patched rabbitMQ.
func (c *FakeRabbitMQs) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha2.RabbitMQ, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(rabbitmqsResource, c.ns, name, pt, data, subresources...), &v1alpha2.RabbitMQ{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.RabbitMQ), err
}
