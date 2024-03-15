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
	v1alpha1 "kubedb.dev/apimachinery/apis/postgres/v1alpha1"
	scheme "kubedb.dev/apimachinery/client/clientset/versioned/scheme"
)

// PublishersGetter has a method to return a PublisherInterface.
// A group's client should implement this interface.
type PublishersGetter interface {
	Publishers(namespace string) PublisherInterface
}

// PublisherInterface has methods to work with Publisher resources.
type PublisherInterface interface {
	Create(ctx context.Context, publisher *v1alpha1.Publisher, opts v1.CreateOptions) (*v1alpha1.Publisher, error)
	Update(ctx context.Context, publisher *v1alpha1.Publisher, opts v1.UpdateOptions) (*v1alpha1.Publisher, error)
	UpdateStatus(ctx context.Context, publisher *v1alpha1.Publisher, opts v1.UpdateOptions) (*v1alpha1.Publisher, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.Publisher, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.PublisherList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Publisher, err error)
	PublisherExpansion
}

// publishers implements PublisherInterface
type publishers struct {
	client rest.Interface
	ns     string
}

// newPublishers returns a Publishers
func newPublishers(c *PostgresV1alpha1Client, namespace string) *publishers {
	return &publishers{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the publisher, and returns the corresponding publisher object, and an error if there is any.
func (c *publishers) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.Publisher, err error) {
	result = &v1alpha1.Publisher{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("publishers").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Publishers that match those selectors.
func (c *publishers) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.PublisherList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.PublisherList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("publishers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested publishers.
func (c *publishers) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("publishers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a publisher and creates it.  Returns the server's representation of the publisher, and an error, if there is any.
func (c *publishers) Create(ctx context.Context, publisher *v1alpha1.Publisher, opts v1.CreateOptions) (result *v1alpha1.Publisher, err error) {
	result = &v1alpha1.Publisher{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("publishers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(publisher).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a publisher and updates it. Returns the server's representation of the publisher, and an error, if there is any.
func (c *publishers) Update(ctx context.Context, publisher *v1alpha1.Publisher, opts v1.UpdateOptions) (result *v1alpha1.Publisher, err error) {
	result = &v1alpha1.Publisher{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("publishers").
		Name(publisher.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(publisher).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *publishers) UpdateStatus(ctx context.Context, publisher *v1alpha1.Publisher, opts v1.UpdateOptions) (result *v1alpha1.Publisher, err error) {
	result = &v1alpha1.Publisher{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("publishers").
		Name(publisher.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(publisher).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the publisher and deletes it. Returns an error if one occurs.
func (c *publishers) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("publishers").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *publishers) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("publishers").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched publisher.
func (c *publishers) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Publisher, err error) {
	result = &v1alpha1.Publisher{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("publishers").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
