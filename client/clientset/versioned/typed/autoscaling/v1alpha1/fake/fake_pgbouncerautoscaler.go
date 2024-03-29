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

// FakePgBouncerAutoscalers implements PgBouncerAutoscalerInterface
type FakePgBouncerAutoscalers struct {
	Fake *FakeAutoscalingV1alpha1
	ns   string
}

var pgbouncerautoscalersResource = v1alpha1.SchemeGroupVersion.WithResource("pgbouncerautoscalers")

var pgbouncerautoscalersKind = v1alpha1.SchemeGroupVersion.WithKind("PgBouncerAutoscaler")

// Get takes name of the pgBouncerAutoscaler, and returns the corresponding pgBouncerAutoscaler object, and an error if there is any.
func (c *FakePgBouncerAutoscalers) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.PgBouncerAutoscaler, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(pgbouncerautoscalersResource, c.ns, name), &v1alpha1.PgBouncerAutoscaler{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PgBouncerAutoscaler), err
}

// List takes label and field selectors, and returns the list of PgBouncerAutoscalers that match those selectors.
func (c *FakePgBouncerAutoscalers) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.PgBouncerAutoscalerList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(pgbouncerautoscalersResource, pgbouncerautoscalersKind, c.ns, opts), &v1alpha1.PgBouncerAutoscalerList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.PgBouncerAutoscalerList{ListMeta: obj.(*v1alpha1.PgBouncerAutoscalerList).ListMeta}
	for _, item := range obj.(*v1alpha1.PgBouncerAutoscalerList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested pgBouncerAutoscalers.
func (c *FakePgBouncerAutoscalers) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(pgbouncerautoscalersResource, c.ns, opts))

}

// Create takes the representation of a pgBouncerAutoscaler and creates it.  Returns the server's representation of the pgBouncerAutoscaler, and an error, if there is any.
func (c *FakePgBouncerAutoscalers) Create(ctx context.Context, pgBouncerAutoscaler *v1alpha1.PgBouncerAutoscaler, opts v1.CreateOptions) (result *v1alpha1.PgBouncerAutoscaler, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(pgbouncerautoscalersResource, c.ns, pgBouncerAutoscaler), &v1alpha1.PgBouncerAutoscaler{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PgBouncerAutoscaler), err
}

// Update takes the representation of a pgBouncerAutoscaler and updates it. Returns the server's representation of the pgBouncerAutoscaler, and an error, if there is any.
func (c *FakePgBouncerAutoscalers) Update(ctx context.Context, pgBouncerAutoscaler *v1alpha1.PgBouncerAutoscaler, opts v1.UpdateOptions) (result *v1alpha1.PgBouncerAutoscaler, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(pgbouncerautoscalersResource, c.ns, pgBouncerAutoscaler), &v1alpha1.PgBouncerAutoscaler{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PgBouncerAutoscaler), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakePgBouncerAutoscalers) UpdateStatus(ctx context.Context, pgBouncerAutoscaler *v1alpha1.PgBouncerAutoscaler, opts v1.UpdateOptions) (*v1alpha1.PgBouncerAutoscaler, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(pgbouncerautoscalersResource, "status", c.ns, pgBouncerAutoscaler), &v1alpha1.PgBouncerAutoscaler{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PgBouncerAutoscaler), err
}

// Delete takes name of the pgBouncerAutoscaler and deletes it. Returns an error if one occurs.
func (c *FakePgBouncerAutoscalers) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(pgbouncerautoscalersResource, c.ns, name, opts), &v1alpha1.PgBouncerAutoscaler{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakePgBouncerAutoscalers) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(pgbouncerautoscalersResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.PgBouncerAutoscalerList{})
	return err
}

// Patch applies the patch and returns the patched pgBouncerAutoscaler.
func (c *FakePgBouncerAutoscalers) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.PgBouncerAutoscaler, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(pgbouncerautoscalersResource, c.ns, name, pt, data, subresources...), &v1alpha1.PgBouncerAutoscaler{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PgBouncerAutoscaler), err
}
