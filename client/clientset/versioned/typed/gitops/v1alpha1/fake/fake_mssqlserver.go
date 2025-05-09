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

	v1alpha1 "kubedb.dev/apimachinery/apis/gitops/v1alpha1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeMSSQLServers implements MSSQLServerInterface
type FakeMSSQLServers struct {
	Fake *FakeGitopsV1alpha1
	ns   string
}

var mssqlserversResource = v1alpha1.SchemeGroupVersion.WithResource("mssqlservers")

var mssqlserversKind = v1alpha1.SchemeGroupVersion.WithKind("MSSQLServer")

// Get takes name of the mSSQLServer, and returns the corresponding mSSQLServer object, and an error if there is any.
func (c *FakeMSSQLServers) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.MSSQLServer, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(mssqlserversResource, c.ns, name), &v1alpha1.MSSQLServer{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MSSQLServer), err
}

// List takes label and field selectors, and returns the list of MSSQLServers that match those selectors.
func (c *FakeMSSQLServers) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.MSSQLServerList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(mssqlserversResource, mssqlserversKind, c.ns, opts), &v1alpha1.MSSQLServerList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.MSSQLServerList{ListMeta: obj.(*v1alpha1.MSSQLServerList).ListMeta}
	for _, item := range obj.(*v1alpha1.MSSQLServerList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested mSSQLServers.
func (c *FakeMSSQLServers) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(mssqlserversResource, c.ns, opts))

}

// Create takes the representation of a mSSQLServer and creates it.  Returns the server's representation of the mSSQLServer, and an error, if there is any.
func (c *FakeMSSQLServers) Create(ctx context.Context, mSSQLServer *v1alpha1.MSSQLServer, opts v1.CreateOptions) (result *v1alpha1.MSSQLServer, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(mssqlserversResource, c.ns, mSSQLServer), &v1alpha1.MSSQLServer{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MSSQLServer), err
}

// Update takes the representation of a mSSQLServer and updates it. Returns the server's representation of the mSSQLServer, and an error, if there is any.
func (c *FakeMSSQLServers) Update(ctx context.Context, mSSQLServer *v1alpha1.MSSQLServer, opts v1.UpdateOptions) (result *v1alpha1.MSSQLServer, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(mssqlserversResource, c.ns, mSSQLServer), &v1alpha1.MSSQLServer{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MSSQLServer), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeMSSQLServers) UpdateStatus(ctx context.Context, mSSQLServer *v1alpha1.MSSQLServer, opts v1.UpdateOptions) (*v1alpha1.MSSQLServer, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(mssqlserversResource, "status", c.ns, mSSQLServer), &v1alpha1.MSSQLServer{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MSSQLServer), err
}

// Delete takes name of the mSSQLServer and deletes it. Returns an error if one occurs.
func (c *FakeMSSQLServers) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(mssqlserversResource, c.ns, name, opts), &v1alpha1.MSSQLServer{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeMSSQLServers) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(mssqlserversResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.MSSQLServerList{})
	return err
}

// Patch applies the patch and returns the patched mSSQLServer.
func (c *FakeMSSQLServers) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.MSSQLServer, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(mssqlserversResource, c.ns, name, pt, data, subresources...), &v1alpha1.MSSQLServer{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MSSQLServer), err
}
