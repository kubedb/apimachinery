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

	v1alpha1 "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeFerretDBOpsRequests implements FerretDBOpsRequestInterface
type FakeFerretDBOpsRequests struct {
	Fake *FakeOpsV1alpha1
	ns   string
}

var ferretdbopsrequestsResource = v1alpha1.SchemeGroupVersion.WithResource("ferretdbopsrequests")

var ferretdbopsrequestsKind = v1alpha1.SchemeGroupVersion.WithKind("FerretDBOpsRequest")

// Get takes name of the ferretDBOpsRequest, and returns the corresponding ferretDBOpsRequest object, and an error if there is any.
func (c *FakeFerretDBOpsRequests) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.FerretDBOpsRequest, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(ferretdbopsrequestsResource, c.ns, name), &v1alpha1.FerretDBOpsRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.FerretDBOpsRequest), err
}

// List takes label and field selectors, and returns the list of FerretDBOpsRequests that match those selectors.
func (c *FakeFerretDBOpsRequests) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.FerretDBOpsRequestList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(ferretdbopsrequestsResource, ferretdbopsrequestsKind, c.ns, opts), &v1alpha1.FerretDBOpsRequestList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.FerretDBOpsRequestList{ListMeta: obj.(*v1alpha1.FerretDBOpsRequestList).ListMeta}
	for _, item := range obj.(*v1alpha1.FerretDBOpsRequestList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested ferretDBOpsRequests.
func (c *FakeFerretDBOpsRequests) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(ferretdbopsrequestsResource, c.ns, opts))

}

// Create takes the representation of a ferretDBOpsRequest and creates it.  Returns the server's representation of the ferretDBOpsRequest, and an error, if there is any.
func (c *FakeFerretDBOpsRequests) Create(ctx context.Context, ferretDBOpsRequest *v1alpha1.FerretDBOpsRequest, opts v1.CreateOptions) (result *v1alpha1.FerretDBOpsRequest, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(ferretdbopsrequestsResource, c.ns, ferretDBOpsRequest), &v1alpha1.FerretDBOpsRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.FerretDBOpsRequest), err
}

// Update takes the representation of a ferretDBOpsRequest and updates it. Returns the server's representation of the ferretDBOpsRequest, and an error, if there is any.
func (c *FakeFerretDBOpsRequests) Update(ctx context.Context, ferretDBOpsRequest *v1alpha1.FerretDBOpsRequest, opts v1.UpdateOptions) (result *v1alpha1.FerretDBOpsRequest, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(ferretdbopsrequestsResource, c.ns, ferretDBOpsRequest), &v1alpha1.FerretDBOpsRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.FerretDBOpsRequest), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeFerretDBOpsRequests) UpdateStatus(ctx context.Context, ferretDBOpsRequest *v1alpha1.FerretDBOpsRequest, opts v1.UpdateOptions) (*v1alpha1.FerretDBOpsRequest, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(ferretdbopsrequestsResource, "status", c.ns, ferretDBOpsRequest), &v1alpha1.FerretDBOpsRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.FerretDBOpsRequest), err
}

// Delete takes name of the ferretDBOpsRequest and deletes it. Returns an error if one occurs.
func (c *FakeFerretDBOpsRequests) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(ferretdbopsrequestsResource, c.ns, name, opts), &v1alpha1.FerretDBOpsRequest{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeFerretDBOpsRequests) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(ferretdbopsrequestsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.FerretDBOpsRequestList{})
	return err
}

// Patch applies the patch and returns the patched ferretDBOpsRequest.
func (c *FakeFerretDBOpsRequests) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.FerretDBOpsRequest, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(ferretdbopsrequestsResource, c.ns, name, pt, data, subresources...), &v1alpha1.FerretDBOpsRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.FerretDBOpsRequest), err
}
