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

	v1alpha1 "kubedb.dev/apimachinery/apis/autoscaling/v1alpha1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeDruidAutoscalers implements DruidAutoscalerInterface
type FakeDruidAutoscalers struct {
	Fake *FakeAutoscalingV1alpha1
	ns   string
}

var druidautoscalersResource = v1alpha1.SchemeGroupVersion.WithResource("druidautoscalers")

var druidautoscalersKind = v1alpha1.SchemeGroupVersion.WithKind("DruidAutoscaler")

// Get takes name of the druidAutoscaler, and returns the corresponding druidAutoscaler object, and an error if there is any.
func (c *FakeDruidAutoscalers) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.DruidAutoscaler, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(druidautoscalersResource, c.ns, name), &v1alpha1.DruidAutoscaler{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DruidAutoscaler), err
}

// List takes label and field selectors, and returns the list of DruidAutoscalers that match those selectors.
func (c *FakeDruidAutoscalers) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.DruidAutoscalerList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(druidautoscalersResource, druidautoscalersKind, c.ns, opts), &v1alpha1.DruidAutoscalerList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.DruidAutoscalerList{ListMeta: obj.(*v1alpha1.DruidAutoscalerList).ListMeta}
	for _, item := range obj.(*v1alpha1.DruidAutoscalerList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested druidAutoscalers.
func (c *FakeDruidAutoscalers) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(druidautoscalersResource, c.ns, opts))

}

// Create takes the representation of a druidAutoscaler and creates it.  Returns the server's representation of the druidAutoscaler, and an error, if there is any.
func (c *FakeDruidAutoscalers) Create(ctx context.Context, druidAutoscaler *v1alpha1.DruidAutoscaler, opts v1.CreateOptions) (result *v1alpha1.DruidAutoscaler, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(druidautoscalersResource, c.ns, druidAutoscaler), &v1alpha1.DruidAutoscaler{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DruidAutoscaler), err
}

// Update takes the representation of a druidAutoscaler and updates it. Returns the server's representation of the druidAutoscaler, and an error, if there is any.
func (c *FakeDruidAutoscalers) Update(ctx context.Context, druidAutoscaler *v1alpha1.DruidAutoscaler, opts v1.UpdateOptions) (result *v1alpha1.DruidAutoscaler, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(druidautoscalersResource, c.ns, druidAutoscaler), &v1alpha1.DruidAutoscaler{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DruidAutoscaler), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeDruidAutoscalers) UpdateStatus(ctx context.Context, druidAutoscaler *v1alpha1.DruidAutoscaler, opts v1.UpdateOptions) (*v1alpha1.DruidAutoscaler, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(druidautoscalersResource, "status", c.ns, druidAutoscaler), &v1alpha1.DruidAutoscaler{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DruidAutoscaler), err
}

// Delete takes name of the druidAutoscaler and deletes it. Returns an error if one occurs.
func (c *FakeDruidAutoscalers) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(druidautoscalersResource, c.ns, name, opts), &v1alpha1.DruidAutoscaler{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeDruidAutoscalers) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(druidautoscalersResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.DruidAutoscalerList{})
	return err
}

// Patch applies the patch and returns the patched druidAutoscaler.
func (c *FakeDruidAutoscalers) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.DruidAutoscaler, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(druidautoscalersResource, c.ns, name, pt, data, subresources...), &v1alpha1.DruidAutoscaler{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DruidAutoscaler), err
}
