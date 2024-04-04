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

// FakeMongoDBOpsRequests implements MongoDBOpsRequestInterface
type FakeMongoDBOpsRequests struct {
	Fake *FakeOpsV1alpha1
	ns   string
}

var mongodbopsrequestsResource = v1alpha1.SchemeGroupVersion.WithResource("mongodbopsrequests")

var mongodbopsrequestsKind = v1alpha1.SchemeGroupVersion.WithKind("MongoDBOpsRequest")

// Get takes name of the mongoDBOpsRequest, and returns the corresponding mongoDBOpsRequest object, and an error if there is any.
func (c *FakeMongoDBOpsRequests) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.MongoDBOpsRequest, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(mongodbopsrequestsResource, c.ns, name), &v1alpha1.MongoDBOpsRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MongoDBOpsRequest), err
}

// List takes label and field selectors, and returns the list of MongoDBOpsRequests that match those selectors.
func (c *FakeMongoDBOpsRequests) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.MongoDBOpsRequestList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(mongodbopsrequestsResource, mongodbopsrequestsKind, c.ns, opts), &v1alpha1.MongoDBOpsRequestList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.MongoDBOpsRequestList{ListMeta: obj.(*v1alpha1.MongoDBOpsRequestList).ListMeta}
	for _, item := range obj.(*v1alpha1.MongoDBOpsRequestList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested mongoDBOpsRequests.
func (c *FakeMongoDBOpsRequests) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(mongodbopsrequestsResource, c.ns, opts))

}

// Create takes the representation of a mongoDBOpsRequest and creates it.  Returns the server's representation of the mongoDBOpsRequest, and an error, if there is any.
func (c *FakeMongoDBOpsRequests) Create(ctx context.Context, mongoDBOpsRequest *v1alpha1.MongoDBOpsRequest, opts v1.CreateOptions) (result *v1alpha1.MongoDBOpsRequest, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(mongodbopsrequestsResource, c.ns, mongoDBOpsRequest), &v1alpha1.MongoDBOpsRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MongoDBOpsRequest), err
}

// Update takes the representation of a mongoDBOpsRequest and updates it. Returns the server's representation of the mongoDBOpsRequest, and an error, if there is any.
func (c *FakeMongoDBOpsRequests) Update(ctx context.Context, mongoDBOpsRequest *v1alpha1.MongoDBOpsRequest, opts v1.UpdateOptions) (result *v1alpha1.MongoDBOpsRequest, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(mongodbopsrequestsResource, c.ns, mongoDBOpsRequest), &v1alpha1.MongoDBOpsRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MongoDBOpsRequest), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeMongoDBOpsRequests) UpdateStatus(ctx context.Context, mongoDBOpsRequest *v1alpha1.MongoDBOpsRequest, opts v1.UpdateOptions) (*v1alpha1.MongoDBOpsRequest, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(mongodbopsrequestsResource, "status", c.ns, mongoDBOpsRequest), &v1alpha1.MongoDBOpsRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MongoDBOpsRequest), err
}

// Delete takes name of the mongoDBOpsRequest and deletes it. Returns an error if one occurs.
func (c *FakeMongoDBOpsRequests) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(mongodbopsrequestsResource, c.ns, name, opts), &v1alpha1.MongoDBOpsRequest{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeMongoDBOpsRequests) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(mongodbopsrequestsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.MongoDBOpsRequestList{})
	return err
}

// Patch applies the patch and returns the patched mongoDBOpsRequest.
func (c *FakeMongoDBOpsRequests) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.MongoDBOpsRequest, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(mongodbopsrequestsResource, c.ns, name, pt, data, subresources...), &v1alpha1.MongoDBOpsRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MongoDBOpsRequest), err
}
