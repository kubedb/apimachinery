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

// FakeRedisSentinelOpsRequests implements RedisSentinelOpsRequestInterface
type FakeRedisSentinelOpsRequests struct {
	Fake *FakeOpsV1alpha1
	ns   string
}

var redissentinelopsrequestsResource = v1alpha1.SchemeGroupVersion.WithResource("redissentinelopsrequests")

var redissentinelopsrequestsKind = v1alpha1.SchemeGroupVersion.WithKind("RedisSentinelOpsRequest")

// Get takes name of the redisSentinelOpsRequest, and returns the corresponding redisSentinelOpsRequest object, and an error if there is any.
func (c *FakeRedisSentinelOpsRequests) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.RedisSentinelOpsRequest, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(redissentinelopsrequestsResource, c.ns, name), &v1alpha1.RedisSentinelOpsRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.RedisSentinelOpsRequest), err
}

// List takes label and field selectors, and returns the list of RedisSentinelOpsRequests that match those selectors.
func (c *FakeRedisSentinelOpsRequests) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.RedisSentinelOpsRequestList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(redissentinelopsrequestsResource, redissentinelopsrequestsKind, c.ns, opts), &v1alpha1.RedisSentinelOpsRequestList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.RedisSentinelOpsRequestList{ListMeta: obj.(*v1alpha1.RedisSentinelOpsRequestList).ListMeta}
	for _, item := range obj.(*v1alpha1.RedisSentinelOpsRequestList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested redisSentinelOpsRequests.
func (c *FakeRedisSentinelOpsRequests) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(redissentinelopsrequestsResource, c.ns, opts))

}

// Create takes the representation of a redisSentinelOpsRequest and creates it.  Returns the server's representation of the redisSentinelOpsRequest, and an error, if there is any.
func (c *FakeRedisSentinelOpsRequests) Create(ctx context.Context, redisSentinelOpsRequest *v1alpha1.RedisSentinelOpsRequest, opts v1.CreateOptions) (result *v1alpha1.RedisSentinelOpsRequest, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(redissentinelopsrequestsResource, c.ns, redisSentinelOpsRequest), &v1alpha1.RedisSentinelOpsRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.RedisSentinelOpsRequest), err
}

// Update takes the representation of a redisSentinelOpsRequest and updates it. Returns the server's representation of the redisSentinelOpsRequest, and an error, if there is any.
func (c *FakeRedisSentinelOpsRequests) Update(ctx context.Context, redisSentinelOpsRequest *v1alpha1.RedisSentinelOpsRequest, opts v1.UpdateOptions) (result *v1alpha1.RedisSentinelOpsRequest, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(redissentinelopsrequestsResource, c.ns, redisSentinelOpsRequest), &v1alpha1.RedisSentinelOpsRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.RedisSentinelOpsRequest), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeRedisSentinelOpsRequests) UpdateStatus(ctx context.Context, redisSentinelOpsRequest *v1alpha1.RedisSentinelOpsRequest, opts v1.UpdateOptions) (*v1alpha1.RedisSentinelOpsRequest, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(redissentinelopsrequestsResource, "status", c.ns, redisSentinelOpsRequest), &v1alpha1.RedisSentinelOpsRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.RedisSentinelOpsRequest), err
}

// Delete takes name of the redisSentinelOpsRequest and deletes it. Returns an error if one occurs.
func (c *FakeRedisSentinelOpsRequests) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(redissentinelopsrequestsResource, c.ns, name, opts), &v1alpha1.RedisSentinelOpsRequest{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeRedisSentinelOpsRequests) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(redissentinelopsrequestsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.RedisSentinelOpsRequestList{})
	return err
}

// Patch applies the patch and returns the patched redisSentinelOpsRequest.
func (c *FakeRedisSentinelOpsRequests) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.RedisSentinelOpsRequest, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(redissentinelopsrequestsResource, c.ns, name, pt, data, subresources...), &v1alpha1.RedisSentinelOpsRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.RedisSentinelOpsRequest), err
}
