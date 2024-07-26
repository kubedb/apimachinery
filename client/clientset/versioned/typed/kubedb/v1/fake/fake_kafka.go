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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
	v1 "kubedb.dev/apimachinery/apis/kubedb/v1"
)

// FakeKafkas implements KafkaInterface
type FakeKafkas struct {
	Fake *FakeKubedbV1
	ns   string
}

var kafkasResource = v1.SchemeGroupVersion.WithResource("kafkas")

var kafkasKind = v1.SchemeGroupVersion.WithKind("Kafka")

// Get takes name of the kafka, and returns the corresponding kafka object, and an error if there is any.
func (c *FakeKafkas) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.Kafka, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(kafkasResource, c.ns, name), &v1.Kafka{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.Kafka), err
}

// List takes label and field selectors, and returns the list of Kafkas that match those selectors.
func (c *FakeKafkas) List(ctx context.Context, opts metav1.ListOptions) (result *v1.KafkaList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(kafkasResource, kafkasKind, c.ns, opts), &v1.KafkaList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1.KafkaList{ListMeta: obj.(*v1.KafkaList).ListMeta}
	for _, item := range obj.(*v1.KafkaList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested kafkas.
func (c *FakeKafkas) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(kafkasResource, c.ns, opts))

}

// Create takes the representation of a kafka and creates it.  Returns the server's representation of the kafka, and an error, if there is any.
func (c *FakeKafkas) Create(ctx context.Context, kafka *v1.Kafka, opts metav1.CreateOptions) (result *v1.Kafka, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(kafkasResource, c.ns, kafka), &v1.Kafka{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.Kafka), err
}

// Update takes the representation of a kafka and updates it. Returns the server's representation of the kafka, and an error, if there is any.
func (c *FakeKafkas) Update(ctx context.Context, kafka *v1.Kafka, opts metav1.UpdateOptions) (result *v1.Kafka, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(kafkasResource, c.ns, kafka), &v1.Kafka{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.Kafka), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeKafkas) UpdateStatus(ctx context.Context, kafka *v1.Kafka, opts metav1.UpdateOptions) (*v1.Kafka, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(kafkasResource, "status", c.ns, kafka), &v1.Kafka{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.Kafka), err
}

// Delete takes name of the kafka and deletes it. Returns an error if one occurs.
func (c *FakeKafkas) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(kafkasResource, c.ns, name, opts), &v1.Kafka{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeKafkas) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(kafkasResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1.KafkaList{})
	return err
}

// Patch applies the patch and returns the patched kafka.
func (c *FakeKafkas) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.Kafka, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(kafkasResource, c.ns, name, pt, data, subresources...), &v1.Kafka{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.Kafka), err
}
