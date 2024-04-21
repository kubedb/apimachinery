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

package v1alpha2

import (
	"context"
	"time"

	v1alpha2 "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	scheme "kubedb.dev/apimachinery/client/clientset/versioned/scheme"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// MSSQLsGetter has a method to return a MSSQLInterface.
// A group's client should implement this interface.
type MSSQLsGetter interface {
	MSSQLs(namespace string) MSSQLInterface
}

// MSSQLInterface has methods to work with MSSQL resources.
type MSSQLInterface interface {
	Create(ctx context.Context, mSSQL *v1alpha2.MSSQL, opts v1.CreateOptions) (*v1alpha2.MSSQL, error)
	Update(ctx context.Context, mSSQL *v1alpha2.MSSQL, opts v1.UpdateOptions) (*v1alpha2.MSSQL, error)
	UpdateStatus(ctx context.Context, mSSQL *v1alpha2.MSSQL, opts v1.UpdateOptions) (*v1alpha2.MSSQL, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha2.MSSQL, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha2.MSSQLList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha2.MSSQL, err error)
	MSSQLExpansion
}

// mSSQLs implements MSSQLInterface
type mSSQLs struct {
	client rest.Interface
	ns     string
}

// newMSSQLs returns a MSSQLs
func newMSSQLs(c *KubedbV1alpha2Client, namespace string) *mSSQLs {
	return &mSSQLs{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the mSSQL, and returns the corresponding mSSQL object, and an error if there is any.
func (c *mSSQLs) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha2.MSSQL, err error) {
	result = &v1alpha2.MSSQL{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("mssqls").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of MSSQLs that match those selectors.
func (c *mSSQLs) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha2.MSSQLList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha2.MSSQLList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("mssqls").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested mSSQLs.
func (c *mSSQLs) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("mssqls").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a mSSQL and creates it.  Returns the server's representation of the mSSQL, and an error, if there is any.
func (c *mSSQLs) Create(ctx context.Context, mSSQL *v1alpha2.MSSQL, opts v1.CreateOptions) (result *v1alpha2.MSSQL, err error) {
	result = &v1alpha2.MSSQL{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("mssqls").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(mSSQL).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a mSSQL and updates it. Returns the server's representation of the mSSQL, and an error, if there is any.
func (c *mSSQLs) Update(ctx context.Context, mSSQL *v1alpha2.MSSQL, opts v1.UpdateOptions) (result *v1alpha2.MSSQL, err error) {
	result = &v1alpha2.MSSQL{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("mssqls").
		Name(mSSQL.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(mSSQL).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *mSSQLs) UpdateStatus(ctx context.Context, mSSQL *v1alpha2.MSSQL, opts v1.UpdateOptions) (result *v1alpha2.MSSQL, err error) {
	result = &v1alpha2.MSSQL{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("mssqls").
		Name(mSSQL.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(mSSQL).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the mSSQL and deletes it. Returns an error if one occurs.
func (c *mSSQLs) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("mssqls").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *mSSQLs) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("mssqls").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched mSSQL.
func (c *mSSQLs) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha2.MSSQL, err error) {
	result = &v1alpha2.MSSQL{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("mssqls").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
