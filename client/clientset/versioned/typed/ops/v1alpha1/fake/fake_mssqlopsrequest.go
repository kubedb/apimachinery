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

// FakeMsSQLOpsRequests implements MsSQLOpsRequestInterface
type FakeMsSQLOpsRequests struct {
	Fake *FakeOpsV1alpha1
	ns   string
}

var mssqlopsrequestsResource = v1alpha1.SchemeGroupVersion.WithResource("mssqlopsrequests")

var mssqlopsrequestsKind = v1alpha1.SchemeGroupVersion.WithKind("MsSQLOpsRequest")

// Get takes name of the msSQLOpsRequest, and returns the corresponding msSQLOpsRequest object, and an error if there is any.
func (c *FakeMsSQLOpsRequests) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.MsSQLOpsRequest, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(mssqlopsrequestsResource, c.ns, name), &v1alpha1.MsSQLOpsRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MsSQLOpsRequest), err
}

// List takes label and field selectors, and returns the list of MsSQLOpsRequests that match those selectors.
func (c *FakeMsSQLOpsRequests) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.MsSQLOpsRequestList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(mssqlopsrequestsResource, mssqlopsrequestsKind, c.ns, opts), &v1alpha1.MsSQLOpsRequestList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.MsSQLOpsRequestList{ListMeta: obj.(*v1alpha1.MsSQLOpsRequestList).ListMeta}
	for _, item := range obj.(*v1alpha1.MsSQLOpsRequestList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested msSQLOpsRequests.
func (c *FakeMsSQLOpsRequests) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(mssqlopsrequestsResource, c.ns, opts))

}

// Create takes the representation of a msSQLOpsRequest and creates it.  Returns the server's representation of the msSQLOpsRequest, and an error, if there is any.
func (c *FakeMsSQLOpsRequests) Create(ctx context.Context, msSQLOpsRequest *v1alpha1.MsSQLOpsRequest, opts v1.CreateOptions) (result *v1alpha1.MsSQLOpsRequest, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(mssqlopsrequestsResource, c.ns, msSQLOpsRequest), &v1alpha1.MsSQLOpsRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MsSQLOpsRequest), err
}

// Update takes the representation of a msSQLOpsRequest and updates it. Returns the server's representation of the msSQLOpsRequest, and an error, if there is any.
func (c *FakeMsSQLOpsRequests) Update(ctx context.Context, msSQLOpsRequest *v1alpha1.MsSQLOpsRequest, opts v1.UpdateOptions) (result *v1alpha1.MsSQLOpsRequest, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(mssqlopsrequestsResource, c.ns, msSQLOpsRequest), &v1alpha1.MsSQLOpsRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MsSQLOpsRequest), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeMsSQLOpsRequests) UpdateStatus(ctx context.Context, msSQLOpsRequest *v1alpha1.MsSQLOpsRequest, opts v1.UpdateOptions) (*v1alpha1.MsSQLOpsRequest, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(mssqlopsrequestsResource, "status", c.ns, msSQLOpsRequest), &v1alpha1.MsSQLOpsRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MsSQLOpsRequest), err
}

// Delete takes name of the msSQLOpsRequest and deletes it. Returns an error if one occurs.
func (c *FakeMsSQLOpsRequests) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(mssqlopsrequestsResource, c.ns, name, opts), &v1alpha1.MsSQLOpsRequest{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeMsSQLOpsRequests) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(mssqlopsrequestsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.MsSQLOpsRequestList{})
	return err
}

// Patch applies the patch and returns the patched msSQLOpsRequest.
func (c *FakeMsSQLOpsRequests) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.MsSQLOpsRequest, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(mssqlopsrequestsResource, c.ns, name, pt, data, subresources...), &v1alpha1.MsSQLOpsRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MsSQLOpsRequest), err
}
