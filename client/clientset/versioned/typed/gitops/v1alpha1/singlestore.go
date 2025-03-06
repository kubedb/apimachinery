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

package v1alpha1

import (
	"context"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
	v1alpha1 "kubedb.dev/apimachinery/apis/gitops/v1alpha1"
	scheme "kubedb.dev/apimachinery/client/clientset/versioned/scheme"
)

// SinglestoresGetter has a method to return a SinglestoreInterface.
// A group's client should implement this interface.
type SinglestoresGetter interface {
	Singlestores(namespace string) SinglestoreInterface
}

// SinglestoreInterface has methods to work with Singlestore resources.
type SinglestoreInterface interface {
	Create(ctx context.Context, singlestore *v1alpha1.Singlestore, opts v1.CreateOptions) (*v1alpha1.Singlestore, error)
	Update(ctx context.Context, singlestore *v1alpha1.Singlestore, opts v1.UpdateOptions) (*v1alpha1.Singlestore, error)
	UpdateStatus(ctx context.Context, singlestore *v1alpha1.Singlestore, opts v1.UpdateOptions) (*v1alpha1.Singlestore, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.Singlestore, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.SinglestoreList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Singlestore, err error)
	SinglestoreExpansion
}

// singlestores implements SinglestoreInterface
type singlestores struct {
	client rest.Interface
	ns     string
}

// newSinglestores returns a Singlestores
func newSinglestores(c *GitopsV1alpha1Client, namespace string) *singlestores {
	return &singlestores{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the singlestore, and returns the corresponding singlestore object, and an error if there is any.
func (c *singlestores) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.Singlestore, err error) {
	result = &v1alpha1.Singlestore{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("singlestores").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Singlestores that match those selectors.
func (c *singlestores) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.SinglestoreList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.SinglestoreList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("singlestores").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested singlestores.
func (c *singlestores) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("singlestores").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a singlestore and creates it.  Returns the server's representation of the singlestore, and an error, if there is any.
func (c *singlestores) Create(ctx context.Context, singlestore *v1alpha1.Singlestore, opts v1.CreateOptions) (result *v1alpha1.Singlestore, err error) {
	result = &v1alpha1.Singlestore{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("singlestores").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(singlestore).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a singlestore and updates it. Returns the server's representation of the singlestore, and an error, if there is any.
func (c *singlestores) Update(ctx context.Context, singlestore *v1alpha1.Singlestore, opts v1.UpdateOptions) (result *v1alpha1.Singlestore, err error) {
	result = &v1alpha1.Singlestore{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("singlestores").
		Name(singlestore.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(singlestore).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *singlestores) UpdateStatus(ctx context.Context, singlestore *v1alpha1.Singlestore, opts v1.UpdateOptions) (result *v1alpha1.Singlestore, err error) {
	result = &v1alpha1.Singlestore{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("singlestores").
		Name(singlestore.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(singlestore).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the singlestore and deletes it. Returns an error if one occurs.
func (c *singlestores) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("singlestores").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *singlestores) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("singlestores").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched singlestore.
func (c *singlestores) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Singlestore, err error) {
	result = &v1alpha1.Singlestore{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("singlestores").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
