/*
Copyright 2017 The KubeDB Authors.

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

package internalversion

import (
	kubedb "github.com/k8sdb/apimachinery/apis/kubedb"
	scheme "github.com/k8sdb/apimachinery/client/internalclientset/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// XdbsGetter has a method to return a XdbInterface.
// A group's client should implement this interface.
type XdbsGetter interface {
	Xdbs(namespace string) XdbInterface
}

// XdbInterface has methods to work with Xdb resources.
type XdbInterface interface {
	Create(*kubedb.Xdb) (*kubedb.Xdb, error)
	Update(*kubedb.Xdb) (*kubedb.Xdb, error)
	UpdateStatus(*kubedb.Xdb) (*kubedb.Xdb, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*kubedb.Xdb, error)
	List(opts v1.ListOptions) (*kubedb.XdbList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *kubedb.Xdb, err error)
	XdbExpansion
}

// xdbs implements XdbInterface
type xdbs struct {
	client rest.Interface
	ns     string
}

// newXdbs returns a Xdbs
func newXdbs(c *KubedbClient, namespace string) *xdbs {
	return &xdbs{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the xdb, and returns the corresponding xdb object, and an error if there is any.
func (c *xdbs) Get(name string, options v1.GetOptions) (result *kubedb.Xdb, err error) {
	result = &kubedb.Xdb{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("xdbs").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Xdbs that match those selectors.
func (c *xdbs) List(opts v1.ListOptions) (result *kubedb.XdbList, err error) {
	result = &kubedb.XdbList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("xdbs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested xdbs.
func (c *xdbs) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("xdbs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a xdb and creates it.  Returns the server's representation of the xdb, and an error, if there is any.
func (c *xdbs) Create(xdb *kubedb.Xdb) (result *kubedb.Xdb, err error) {
	result = &kubedb.Xdb{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("xdbs").
		Body(xdb).
		Do().
		Into(result)
	return
}

// Update takes the representation of a xdb and updates it. Returns the server's representation of the xdb, and an error, if there is any.
func (c *xdbs) Update(xdb *kubedb.Xdb) (result *kubedb.Xdb, err error) {
	result = &kubedb.Xdb{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("xdbs").
		Name(xdb.Name).
		Body(xdb).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *xdbs) UpdateStatus(xdb *kubedb.Xdb) (result *kubedb.Xdb, err error) {
	result = &kubedb.Xdb{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("xdbs").
		Name(xdb.Name).
		SubResource("status").
		Body(xdb).
		Do().
		Into(result)
	return
}

// Delete takes name of the xdb and deletes it. Returns an error if one occurs.
func (c *xdbs) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("xdbs").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *xdbs) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("xdbs").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched xdb.
func (c *xdbs) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *kubedb.Xdb, err error) {
	result = &kubedb.Xdb{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("xdbs").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
