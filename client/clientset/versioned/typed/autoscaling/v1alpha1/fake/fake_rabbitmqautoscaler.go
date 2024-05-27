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
	v1alpha1 "kubedb.dev/apimachinery/apis/autoscaling/v1alpha1"
)

// FakeRabbitMQAutoscalers implements RabbitMQAutoscalerInterface
type FakeRabbitMQAutoscalers struct {
	Fake *FakeAutoscalingV1alpha1
	ns   string
}

var rabbitmqautoscalersResource = v1alpha1.SchemeGroupVersion.WithResource("rabbitmqautoscalers")

var rabbitmqautoscalersKind = v1alpha1.SchemeGroupVersion.WithKind("RabbitMQAutoscaler")

// Get takes name of the rabbitMQAutoscaler, and returns the corresponding rabbitMQAutoscaler object, and an error if there is any.
func (c *FakeRabbitMQAutoscalers) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.RabbitMQAutoscaler, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(rabbitmqautoscalersResource, c.ns, name), &v1alpha1.RabbitMQAutoscaler{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.RabbitMQAutoscaler), err
}

// List takes label and field selectors, and returns the list of RabbitMQAutoscalers that match those selectors.
func (c *FakeRabbitMQAutoscalers) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.RabbitMQAutoscalerList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(rabbitmqautoscalersResource, rabbitmqautoscalersKind, c.ns, opts), &v1alpha1.RabbitMQAutoscalerList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.RabbitMQAutoscalerList{ListMeta: obj.(*v1alpha1.RabbitMQAutoscalerList).ListMeta}
	for _, item := range obj.(*v1alpha1.RabbitMQAutoscalerList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested rabbitMQAutoscalers.
func (c *FakeRabbitMQAutoscalers) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(rabbitmqautoscalersResource, c.ns, opts))

}

// Create takes the representation of a rabbitMQAutoscaler and creates it.  Returns the server's representation of the rabbitMQAutoscaler, and an error, if there is any.
func (c *FakeRabbitMQAutoscalers) Create(ctx context.Context, rabbitMQAutoscaler *v1alpha1.RabbitMQAutoscaler, opts v1.CreateOptions) (result *v1alpha1.RabbitMQAutoscaler, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(rabbitmqautoscalersResource, c.ns, rabbitMQAutoscaler), &v1alpha1.RabbitMQAutoscaler{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.RabbitMQAutoscaler), err
}

// Update takes the representation of a rabbitMQAutoscaler and updates it. Returns the server's representation of the rabbitMQAutoscaler, and an error, if there is any.
func (c *FakeRabbitMQAutoscalers) Update(ctx context.Context, rabbitMQAutoscaler *v1alpha1.RabbitMQAutoscaler, opts v1.UpdateOptions) (result *v1alpha1.RabbitMQAutoscaler, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(rabbitmqautoscalersResource, c.ns, rabbitMQAutoscaler), &v1alpha1.RabbitMQAutoscaler{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.RabbitMQAutoscaler), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeRabbitMQAutoscalers) UpdateStatus(ctx context.Context, rabbitMQAutoscaler *v1alpha1.RabbitMQAutoscaler, opts v1.UpdateOptions) (*v1alpha1.RabbitMQAutoscaler, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(rabbitmqautoscalersResource, "status", c.ns, rabbitMQAutoscaler), &v1alpha1.RabbitMQAutoscaler{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.RabbitMQAutoscaler), err
}

// Delete takes name of the rabbitMQAutoscaler and deletes it. Returns an error if one occurs.
func (c *FakeRabbitMQAutoscalers) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(rabbitmqautoscalersResource, c.ns, name, opts), &v1alpha1.RabbitMQAutoscaler{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeRabbitMQAutoscalers) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(rabbitmqautoscalersResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.RabbitMQAutoscalerList{})
	return err
}

// Patch applies the patch and returns the patched rabbitMQAutoscaler.
func (c *FakeRabbitMQAutoscalers) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.RabbitMQAutoscaler, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(rabbitmqautoscalersResource, c.ns, name, pt, data, subresources...), &v1alpha1.RabbitMQAutoscaler{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.RabbitMQAutoscaler), err
}
