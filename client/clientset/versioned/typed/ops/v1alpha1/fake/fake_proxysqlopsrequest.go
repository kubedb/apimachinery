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
	v1alpha1 "kubedb.dev/apimachinery/apis/ops/v1alpha1"
)

// FakeProxySQLOpsRequests implements ProxySQLOpsRequestInterface
type FakeProxySQLOpsRequests struct {
	Fake *FakeOpsV1alpha1
	ns   string
}

var proxysqlopsrequestsResource = v1alpha1.SchemeGroupVersion.WithResource("proxysqlopsrequests")

var proxysqlopsrequestsKind = v1alpha1.SchemeGroupVersion.WithKind("ProxySQLOpsRequest")

// Get takes name of the proxySQLOpsRequest, and returns the corresponding proxySQLOpsRequest object, and an error if there is any.
func (c *FakeProxySQLOpsRequests) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.ProxySQLOpsRequest, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(proxysqlopsrequestsResource, c.ns, name), &v1alpha1.ProxySQLOpsRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ProxySQLOpsRequest), err
}

// List takes label and field selectors, and returns the list of ProxySQLOpsRequests that match those selectors.
func (c *FakeProxySQLOpsRequests) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.ProxySQLOpsRequestList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(proxysqlopsrequestsResource, proxysqlopsrequestsKind, c.ns, opts), &v1alpha1.ProxySQLOpsRequestList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.ProxySQLOpsRequestList{ListMeta: obj.(*v1alpha1.ProxySQLOpsRequestList).ListMeta}
	for _, item := range obj.(*v1alpha1.ProxySQLOpsRequestList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested proxySQLOpsRequests.
func (c *FakeProxySQLOpsRequests) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(proxysqlopsrequestsResource, c.ns, opts))

}

// Create takes the representation of a proxySQLOpsRequest and creates it.  Returns the server's representation of the proxySQLOpsRequest, and an error, if there is any.
func (c *FakeProxySQLOpsRequests) Create(ctx context.Context, proxySQLOpsRequest *v1alpha1.ProxySQLOpsRequest, opts v1.CreateOptions) (result *v1alpha1.ProxySQLOpsRequest, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(proxysqlopsrequestsResource, c.ns, proxySQLOpsRequest), &v1alpha1.ProxySQLOpsRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ProxySQLOpsRequest), err
}

// Update takes the representation of a proxySQLOpsRequest and updates it. Returns the server's representation of the proxySQLOpsRequest, and an error, if there is any.
func (c *FakeProxySQLOpsRequests) Update(ctx context.Context, proxySQLOpsRequest *v1alpha1.ProxySQLOpsRequest, opts v1.UpdateOptions) (result *v1alpha1.ProxySQLOpsRequest, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(proxysqlopsrequestsResource, c.ns, proxySQLOpsRequest), &v1alpha1.ProxySQLOpsRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ProxySQLOpsRequest), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeProxySQLOpsRequests) UpdateStatus(ctx context.Context, proxySQLOpsRequest *v1alpha1.ProxySQLOpsRequest, opts v1.UpdateOptions) (*v1alpha1.ProxySQLOpsRequest, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(proxysqlopsrequestsResource, "status", c.ns, proxySQLOpsRequest), &v1alpha1.ProxySQLOpsRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ProxySQLOpsRequest), err
}

// Delete takes name of the proxySQLOpsRequest and deletes it. Returns an error if one occurs.
func (c *FakeProxySQLOpsRequests) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(proxysqlopsrequestsResource, c.ns, name, opts), &v1alpha1.ProxySQLOpsRequest{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeProxySQLOpsRequests) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(proxysqlopsrequestsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.ProxySQLOpsRequestList{})
	return err
}

// Patch applies the patch and returns the patched proxySQLOpsRequest.
func (c *FakeProxySQLOpsRequests) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.ProxySQLOpsRequest, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(proxysqlopsrequestsResource, c.ns, name, pt, data, subresources...), &v1alpha1.ProxySQLOpsRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ProxySQLOpsRequest), err
}
