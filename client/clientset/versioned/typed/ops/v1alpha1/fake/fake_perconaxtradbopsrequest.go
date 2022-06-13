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
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
	v1alpha1 "kubedb.dev/apimachinery/apis/ops/v1alpha1"
)

// FakePerconaXtraDBOpsRequests implements PerconaXtraDBOpsRequestInterface
type FakePerconaXtraDBOpsRequests struct {
	Fake *FakeOpsV1alpha1
	ns   string
}

var perconaxtradbopsrequestsResource = schema.GroupVersionResource{Group: "ops.kubedb.com", Version: "v1alpha1", Resource: "perconaxtradbopsrequests"}

var perconaxtradbopsrequestsKind = schema.GroupVersionKind{Group: "ops.kubedb.com", Version: "v1alpha1", Kind: "PerconaXtraDBOpsRequest"}

// Get takes name of the perconaXtraDBOpsRequest, and returns the corresponding perconaXtraDBOpsRequest object, and an error if there is any.
func (c *FakePerconaXtraDBOpsRequests) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.PerconaXtraDBOpsRequest, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(perconaxtradbopsrequestsResource, c.ns, name), &v1alpha1.PerconaXtraDBOpsRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PerconaXtraDBOpsRequest), err
}

// List takes label and field selectors, and returns the list of PerconaXtraDBOpsRequests that match those selectors.
func (c *FakePerconaXtraDBOpsRequests) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.PerconaXtraDBOpsRequestList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(perconaxtradbopsrequestsResource, perconaxtradbopsrequestsKind, c.ns, opts), &v1alpha1.PerconaXtraDBOpsRequestList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.PerconaXtraDBOpsRequestList{ListMeta: obj.(*v1alpha1.PerconaXtraDBOpsRequestList).ListMeta}
	for _, item := range obj.(*v1alpha1.PerconaXtraDBOpsRequestList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested perconaXtraDBOpsRequests.
func (c *FakePerconaXtraDBOpsRequests) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(perconaxtradbopsrequestsResource, c.ns, opts))

}

// Create takes the representation of a perconaXtraDBOpsRequest and creates it.  Returns the server's representation of the perconaXtraDBOpsRequest, and an error, if there is any.
func (c *FakePerconaXtraDBOpsRequests) Create(ctx context.Context, perconaXtraDBOpsRequest *v1alpha1.PerconaXtraDBOpsRequest, opts v1.CreateOptions) (result *v1alpha1.PerconaXtraDBOpsRequest, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(perconaxtradbopsrequestsResource, c.ns, perconaXtraDBOpsRequest), &v1alpha1.PerconaXtraDBOpsRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PerconaXtraDBOpsRequest), err
}

// Update takes the representation of a perconaXtraDBOpsRequest and updates it. Returns the server's representation of the perconaXtraDBOpsRequest, and an error, if there is any.
func (c *FakePerconaXtraDBOpsRequests) Update(ctx context.Context, perconaXtraDBOpsRequest *v1alpha1.PerconaXtraDBOpsRequest, opts v1.UpdateOptions) (result *v1alpha1.PerconaXtraDBOpsRequest, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(perconaxtradbopsrequestsResource, c.ns, perconaXtraDBOpsRequest), &v1alpha1.PerconaXtraDBOpsRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PerconaXtraDBOpsRequest), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakePerconaXtraDBOpsRequests) UpdateStatus(ctx context.Context, perconaXtraDBOpsRequest *v1alpha1.PerconaXtraDBOpsRequest, opts v1.UpdateOptions) (*v1alpha1.PerconaXtraDBOpsRequest, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(perconaxtradbopsrequestsResource, "status", c.ns, perconaXtraDBOpsRequest), &v1alpha1.PerconaXtraDBOpsRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PerconaXtraDBOpsRequest), err
}

// Delete takes name of the perconaXtraDBOpsRequest and deletes it. Returns an error if one occurs.
func (c *FakePerconaXtraDBOpsRequests) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(perconaxtradbopsrequestsResource, c.ns, name, opts), &v1alpha1.PerconaXtraDBOpsRequest{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakePerconaXtraDBOpsRequests) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(perconaxtradbopsrequestsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.PerconaXtraDBOpsRequestList{})
	return err
}

// Patch applies the patch and returns the patched perconaXtraDBOpsRequest.
func (c *FakePerconaXtraDBOpsRequests) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.PerconaXtraDBOpsRequest, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(perconaxtradbopsrequestsResource, c.ns, name, pt, data, subresources...), &v1alpha1.PerconaXtraDBOpsRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PerconaXtraDBOpsRequest), err
}
