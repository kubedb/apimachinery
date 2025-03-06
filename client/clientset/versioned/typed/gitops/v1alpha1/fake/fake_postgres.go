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
	v1alpha1 "kubedb.dev/apimachinery/apis/gitops/v1alpha1"
)

// FakePostgreses implements PostgresInterface
type FakePostgreses struct {
	Fake *FakeGitopsV1alpha1
	ns   string
}

var postgresesResource = v1alpha1.SchemeGroupVersion.WithResource("postgreses")

var postgresesKind = v1alpha1.SchemeGroupVersion.WithKind("Postgres")

// Get takes name of the postgres, and returns the corresponding postgres object, and an error if there is any.
func (c *FakePostgreses) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.Postgres, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(postgresesResource, c.ns, name), &v1alpha1.Postgres{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Postgres), err
}

// List takes label and field selectors, and returns the list of Postgreses that match those selectors.
func (c *FakePostgreses) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.PostgresList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(postgresesResource, postgresesKind, c.ns, opts), &v1alpha1.PostgresList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.PostgresList{ListMeta: obj.(*v1alpha1.PostgresList).ListMeta}
	for _, item := range obj.(*v1alpha1.PostgresList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested postgreses.
func (c *FakePostgreses) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(postgresesResource, c.ns, opts))

}

// Create takes the representation of a postgres and creates it.  Returns the server's representation of the postgres, and an error, if there is any.
func (c *FakePostgreses) Create(ctx context.Context, postgres *v1alpha1.Postgres, opts v1.CreateOptions) (result *v1alpha1.Postgres, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(postgresesResource, c.ns, postgres), &v1alpha1.Postgres{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Postgres), err
}

// Update takes the representation of a postgres and updates it. Returns the server's representation of the postgres, and an error, if there is any.
func (c *FakePostgreses) Update(ctx context.Context, postgres *v1alpha1.Postgres, opts v1.UpdateOptions) (result *v1alpha1.Postgres, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(postgresesResource, c.ns, postgres), &v1alpha1.Postgres{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Postgres), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakePostgreses) UpdateStatus(ctx context.Context, postgres *v1alpha1.Postgres, opts v1.UpdateOptions) (*v1alpha1.Postgres, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(postgresesResource, "status", c.ns, postgres), &v1alpha1.Postgres{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Postgres), err
}

// Delete takes name of the postgres and deletes it. Returns an error if one occurs.
func (c *FakePostgreses) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(postgresesResource, c.ns, name, opts), &v1alpha1.Postgres{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakePostgreses) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(postgresesResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.PostgresList{})
	return err
}

// Patch applies the patch and returns the patched postgres.
func (c *FakePostgreses) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Postgres, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(postgresesResource, c.ns, name, pt, data, subresources...), &v1alpha1.Postgres{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Postgres), err
}
