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

// FakeDruidOpsRequests implements DruidOpsRequestInterface
type FakeDruidOpsRequests struct {
	Fake *FakeOpsV1alpha1
	ns   string
}

var druidopsrequestsResource = v1alpha1.SchemeGroupVersion.WithResource("druidopsrequests")

var druidopsrequestsKind = v1alpha1.SchemeGroupVersion.WithKind("DruidOpsRequest")

// Get takes name of the druidOpsRequest, and returns the corresponding druidOpsRequest object, and an error if there is any.
func (c *FakeDruidOpsRequests) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.DruidOpsRequest, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(druidopsrequestsResource, c.ns, name), &v1alpha1.DruidOpsRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DruidOpsRequest), err
}

// List takes label and field selectors, and returns the list of DruidOpsRequests that match those selectors.
func (c *FakeDruidOpsRequests) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.DruidOpsRequestList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(druidopsrequestsResource, druidopsrequestsKind, c.ns, opts), &v1alpha1.DruidOpsRequestList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.DruidOpsRequestList{ListMeta: obj.(*v1alpha1.DruidOpsRequestList).ListMeta}
	for _, item := range obj.(*v1alpha1.DruidOpsRequestList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested druidOpsRequests.
func (c *FakeDruidOpsRequests) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(druidopsrequestsResource, c.ns, opts))

}

// Create takes the representation of a druidOpsRequest and creates it.  Returns the server's representation of the druidOpsRequest, and an error, if there is any.
func (c *FakeDruidOpsRequests) Create(ctx context.Context, druidOpsRequest *v1alpha1.DruidOpsRequest, opts v1.CreateOptions) (result *v1alpha1.DruidOpsRequest, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(druidopsrequestsResource, c.ns, druidOpsRequest), &v1alpha1.DruidOpsRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DruidOpsRequest), err
}

// Update takes the representation of a druidOpsRequest and updates it. Returns the server's representation of the druidOpsRequest, and an error, if there is any.
func (c *FakeDruidOpsRequests) Update(ctx context.Context, druidOpsRequest *v1alpha1.DruidOpsRequest, opts v1.UpdateOptions) (result *v1alpha1.DruidOpsRequest, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(druidopsrequestsResource, c.ns, druidOpsRequest), &v1alpha1.DruidOpsRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DruidOpsRequest), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeDruidOpsRequests) UpdateStatus(ctx context.Context, druidOpsRequest *v1alpha1.DruidOpsRequest, opts v1.UpdateOptions) (*v1alpha1.DruidOpsRequest, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(druidopsrequestsResource, "status", c.ns, druidOpsRequest), &v1alpha1.DruidOpsRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DruidOpsRequest), err
}

// Delete takes name of the druidOpsRequest and deletes it. Returns an error if one occurs.
func (c *FakeDruidOpsRequests) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(druidopsrequestsResource, c.ns, name, opts), &v1alpha1.DruidOpsRequest{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeDruidOpsRequests) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(druidopsrequestsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.DruidOpsRequestList{})
	return err
}

// Patch applies the patch and returns the patched druidOpsRequest.
func (c *FakeDruidOpsRequests) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.DruidOpsRequest, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(druidopsrequestsResource, c.ns, name, pt, data, subresources...), &v1alpha1.DruidOpsRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DruidOpsRequest), err
}
