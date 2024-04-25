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
	v1alpha3 "kubedb.dev/apimachinery/apis/kubedb/v1alpha3"
)

// FakePgBouncers implements PgBouncerInterface
type FakePgBouncers struct {
	Fake *FakeKubedbV1alpha3
	ns   string
}

var pgbouncersResource = v1alpha3.SchemeGroupVersion.WithResource("pgbouncers")

var pgbouncersKind = v1alpha3.SchemeGroupVersion.WithKind("PgBouncer")

// Get takes name of the pgBouncer, and returns the corresponding pgBouncer object, and an error if there is any.
func (c *FakePgBouncers) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha3.PgBouncer, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(pgbouncersResource, c.ns, name), &v1alpha3.PgBouncer{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha3.PgBouncer), err
}

// List takes label and field selectors, and returns the list of PgBouncers that match those selectors.
func (c *FakePgBouncers) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha3.PgBouncerList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(pgbouncersResource, pgbouncersKind, c.ns, opts), &v1alpha3.PgBouncerList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha3.PgBouncerList{ListMeta: obj.(*v1alpha3.PgBouncerList).ListMeta}
	for _, item := range obj.(*v1alpha3.PgBouncerList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested pgBouncers.
func (c *FakePgBouncers) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(pgbouncersResource, c.ns, opts))

}

// Create takes the representation of a pgBouncer and creates it.  Returns the server's representation of the pgBouncer, and an error, if there is any.
func (c *FakePgBouncers) Create(ctx context.Context, pgBouncer *v1alpha3.PgBouncer, opts v1.CreateOptions) (result *v1alpha3.PgBouncer, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(pgbouncersResource, c.ns, pgBouncer), &v1alpha3.PgBouncer{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha3.PgBouncer), err
}

// Update takes the representation of a pgBouncer and updates it. Returns the server's representation of the pgBouncer, and an error, if there is any.
func (c *FakePgBouncers) Update(ctx context.Context, pgBouncer *v1alpha3.PgBouncer, opts v1.UpdateOptions) (result *v1alpha3.PgBouncer, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(pgbouncersResource, c.ns, pgBouncer), &v1alpha3.PgBouncer{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha3.PgBouncer), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakePgBouncers) UpdateStatus(ctx context.Context, pgBouncer *v1alpha3.PgBouncer, opts v1.UpdateOptions) (*v1alpha3.PgBouncer, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(pgbouncersResource, "status", c.ns, pgBouncer), &v1alpha3.PgBouncer{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha3.PgBouncer), err
}

// Delete takes name of the pgBouncer and deletes it. Returns an error if one occurs.
func (c *FakePgBouncers) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(pgbouncersResource, c.ns, name, opts), &v1alpha3.PgBouncer{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakePgBouncers) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(pgbouncersResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha3.PgBouncerList{})
	return err
}

// Patch applies the patch and returns the patched pgBouncer.
func (c *FakePgBouncers) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha3.PgBouncer, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(pgbouncersResource, c.ns, name, pt, data, subresources...), &v1alpha3.PgBouncer{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha3.PgBouncer), err
}
