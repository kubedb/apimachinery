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

// FakeProxySQLs implements ProxySQLInterface
type FakeProxySQLs struct {
	Fake *FakeGitopsV1alpha1
	ns   string
}

var proxysqlsResource = v1alpha1.SchemeGroupVersion.WithResource("proxysqls")

var proxysqlsKind = v1alpha1.SchemeGroupVersion.WithKind("ProxySQL")

// Get takes name of the proxySQL, and returns the corresponding proxySQL object, and an error if there is any.
func (c *FakeProxySQLs) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.ProxySQL, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(proxysqlsResource, c.ns, name), &v1alpha1.ProxySQL{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ProxySQL), err
}

// List takes label and field selectors, and returns the list of ProxySQLs that match those selectors.
func (c *FakeProxySQLs) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.ProxySQLList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(proxysqlsResource, proxysqlsKind, c.ns, opts), &v1alpha1.ProxySQLList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.ProxySQLList{ListMeta: obj.(*v1alpha1.ProxySQLList).ListMeta}
	for _, item := range obj.(*v1alpha1.ProxySQLList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested proxySQLs.
func (c *FakeProxySQLs) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(proxysqlsResource, c.ns, opts))

}

// Create takes the representation of a proxySQL and creates it.  Returns the server's representation of the proxySQL, and an error, if there is any.
func (c *FakeProxySQLs) Create(ctx context.Context, proxySQL *v1alpha1.ProxySQL, opts v1.CreateOptions) (result *v1alpha1.ProxySQL, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(proxysqlsResource, c.ns, proxySQL), &v1alpha1.ProxySQL{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ProxySQL), err
}

// Update takes the representation of a proxySQL and updates it. Returns the server's representation of the proxySQL, and an error, if there is any.
func (c *FakeProxySQLs) Update(ctx context.Context, proxySQL *v1alpha1.ProxySQL, opts v1.UpdateOptions) (result *v1alpha1.ProxySQL, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(proxysqlsResource, c.ns, proxySQL), &v1alpha1.ProxySQL{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ProxySQL), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeProxySQLs) UpdateStatus(ctx context.Context, proxySQL *v1alpha1.ProxySQL, opts v1.UpdateOptions) (*v1alpha1.ProxySQL, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(proxysqlsResource, "status", c.ns, proxySQL), &v1alpha1.ProxySQL{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ProxySQL), err
}

// Delete takes name of the proxySQL and deletes it. Returns an error if one occurs.
func (c *FakeProxySQLs) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(proxysqlsResource, c.ns, name, opts), &v1alpha1.ProxySQL{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeProxySQLs) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(proxysqlsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.ProxySQLList{})
	return err
}

// Patch applies the patch and returns the patched proxySQL.
func (c *FakeProxySQLs) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.ProxySQL, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(proxysqlsResource, c.ns, name, pt, data, subresources...), &v1alpha1.ProxySQL{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ProxySQL), err
}
