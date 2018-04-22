/*
Copyright 2018 The KubeDB Authors.

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

package v1alpha1

import (
	v1alpha1 "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	scheme "github.com/kubedb/apimachinery/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// EtcdsGetter has a method to return a EtcdInterface.
// A group's client should implement this interface.
type EtcdsGetter interface {
	Etcds(namespace string) EtcdInterface
}

// EtcdInterface has methods to work with Etcd resources.
type EtcdInterface interface {
	Create(*v1alpha1.Etcd) (*v1alpha1.Etcd, error)
	Update(*v1alpha1.Etcd) (*v1alpha1.Etcd, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.Etcd, error)
	List(opts v1.ListOptions) (*v1alpha1.EtcdList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Etcd, err error)
	EtcdExpansion
}

// etcds implements EtcdInterface
type etcds struct {
	client rest.Interface
	ns     string
}

// newEtcds returns a Etcds
func newEtcds(c *KubedbV1alpha1Client, namespace string) *etcds {
	return &etcds{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the etcd, and returns the corresponding etcd object, and an error if there is any.
func (c *etcds) Get(name string, options v1.GetOptions) (result *v1alpha1.Etcd, err error) {
	result = &v1alpha1.Etcd{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("etcds").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Etcds that match those selectors.
func (c *etcds) List(opts v1.ListOptions) (result *v1alpha1.EtcdList, err error) {
	result = &v1alpha1.EtcdList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("etcds").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested etcds.
func (c *etcds) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("etcds").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a etcd and creates it.  Returns the server's representation of the etcd, and an error, if there is any.
func (c *etcds) Create(etcd *v1alpha1.Etcd) (result *v1alpha1.Etcd, err error) {
	result = &v1alpha1.Etcd{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("etcds").
		Body(etcd).
		Do().
		Into(result)
	return
}

// Update takes the representation of a etcd and updates it. Returns the server's representation of the etcd, and an error, if there is any.
func (c *etcds) Update(etcd *v1alpha1.Etcd) (result *v1alpha1.Etcd, err error) {
	result = &v1alpha1.Etcd{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("etcds").
		Name(etcd.Name).
		Body(etcd).
		Do().
		Into(result)
	return
}

// Delete takes name of the etcd and deletes it. Returns an error if one occurs.
func (c *etcds) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("etcds").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *etcds) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("etcds").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched etcd.
func (c *etcds) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Etcd, err error) {
	result = &v1alpha1.Etcd{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("etcds").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
