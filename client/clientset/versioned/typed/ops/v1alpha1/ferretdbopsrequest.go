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
	v1alpha1 "kubedb.dev/apimachinery/apis/ops/v1alpha1"
	scheme "kubedb.dev/apimachinery/client/clientset/versioned/scheme"
)

// FerretDBOpsRequestsGetter has a method to return a FerretDBOpsRequestInterface.
// A group's client should implement this interface.
type FerretDBOpsRequestsGetter interface {
	FerretDBOpsRequests(namespace string) FerretDBOpsRequestInterface
}

// FerretDBOpsRequestInterface has methods to work with FerretDBOpsRequest resources.
type FerretDBOpsRequestInterface interface {
	Create(ctx context.Context, ferretDBOpsRequest *v1alpha1.FerretDBOpsRequest, opts v1.CreateOptions) (*v1alpha1.FerretDBOpsRequest, error)
	Update(ctx context.Context, ferretDBOpsRequest *v1alpha1.FerretDBOpsRequest, opts v1.UpdateOptions) (*v1alpha1.FerretDBOpsRequest, error)
	UpdateStatus(ctx context.Context, ferretDBOpsRequest *v1alpha1.FerretDBOpsRequest, opts v1.UpdateOptions) (*v1alpha1.FerretDBOpsRequest, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.FerretDBOpsRequest, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.FerretDBOpsRequestList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.FerretDBOpsRequest, err error)
	FerretDBOpsRequestExpansion
}

// ferretDBOpsRequests implements FerretDBOpsRequestInterface
type ferretDBOpsRequests struct {
	client rest.Interface
	ns     string
}

// newFerretDBOpsRequests returns a FerretDBOpsRequests
func newFerretDBOpsRequests(c *OpsV1alpha1Client, namespace string) *ferretDBOpsRequests {
	return &ferretDBOpsRequests{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the ferretDBOpsRequest, and returns the corresponding ferretDBOpsRequest object, and an error if there is any.
func (c *ferretDBOpsRequests) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.FerretDBOpsRequest, err error) {
	result = &v1alpha1.FerretDBOpsRequest{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("ferretdbopsrequests").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of FerretDBOpsRequests that match those selectors.
func (c *ferretDBOpsRequests) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.FerretDBOpsRequestList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.FerretDBOpsRequestList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("ferretdbopsrequests").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested ferretDBOpsRequests.
func (c *ferretDBOpsRequests) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("ferretdbopsrequests").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a ferretDBOpsRequest and creates it.  Returns the server's representation of the ferretDBOpsRequest, and an error, if there is any.
func (c *ferretDBOpsRequests) Create(ctx context.Context, ferretDBOpsRequest *v1alpha1.FerretDBOpsRequest, opts v1.CreateOptions) (result *v1alpha1.FerretDBOpsRequest, err error) {
	result = &v1alpha1.FerretDBOpsRequest{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("ferretdbopsrequests").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(ferretDBOpsRequest).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a ferretDBOpsRequest and updates it. Returns the server's representation of the ferretDBOpsRequest, and an error, if there is any.
func (c *ferretDBOpsRequests) Update(ctx context.Context, ferretDBOpsRequest *v1alpha1.FerretDBOpsRequest, opts v1.UpdateOptions) (result *v1alpha1.FerretDBOpsRequest, err error) {
	result = &v1alpha1.FerretDBOpsRequest{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("ferretdbopsrequests").
		Name(ferretDBOpsRequest.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(ferretDBOpsRequest).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *ferretDBOpsRequests) UpdateStatus(ctx context.Context, ferretDBOpsRequest *v1alpha1.FerretDBOpsRequest, opts v1.UpdateOptions) (result *v1alpha1.FerretDBOpsRequest, err error) {
	result = &v1alpha1.FerretDBOpsRequest{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("ferretdbopsrequests").
		Name(ferretDBOpsRequest.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(ferretDBOpsRequest).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the ferretDBOpsRequest and deletes it. Returns an error if one occurs.
func (c *ferretDBOpsRequests) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("ferretdbopsrequests").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *ferretDBOpsRequests) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("ferretdbopsrequests").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched ferretDBOpsRequest.
func (c *ferretDBOpsRequests) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.FerretDBOpsRequest, err error) {
	result = &v1alpha1.FerretDBOpsRequest{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("ferretdbopsrequests").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
