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

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
	v1alpha2 "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	scheme "kubedb.dev/apimachinery/client/clientset/versioned/scheme"
)

// DruidsGetter has a method to return a DruidInterface.
// A group's client should implement this interface.
type DruidsGetter interface {
	Druids(namespace string) DruidInterface
}

// DruidInterface has methods to work with Druid resources.
type DruidInterface interface {
	Create(ctx context.Context, druid *v1alpha2.Druid, opts v1.CreateOptions) (*v1alpha2.Druid, error)
	Update(ctx context.Context, druid *v1alpha2.Druid, opts v1.UpdateOptions) (*v1alpha2.Druid, error)
	UpdateStatus(ctx context.Context, druid *v1alpha2.Druid, opts v1.UpdateOptions) (*v1alpha2.Druid, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha2.Druid, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha2.DruidList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha2.Druid, err error)
	DruidExpansion
}

// druids implements DruidInterface
type druids struct {
	client rest.Interface
	ns     string
}

// newDruids returns a Druids
func newDruids(c *KubedbV1alpha2Client, namespace string) *druids {
	return &druids{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the druid, and returns the corresponding druid object, and an error if there is any.
func (c *druids) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha2.Druid, err error) {
	result = &v1alpha2.Druid{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("druids").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Druids that match those selectors.
func (c *druids) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha2.DruidList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha2.DruidList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("druids").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested druids.
func (c *druids) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("druids").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a druid and creates it.  Returns the server's representation of the druid, and an error, if there is any.
func (c *druids) Create(ctx context.Context, druid *v1alpha2.Druid, opts v1.CreateOptions) (result *v1alpha2.Druid, err error) {
	result = &v1alpha2.Druid{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("druids").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(druid).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a druid and updates it. Returns the server's representation of the druid, and an error, if there is any.
func (c *druids) Update(ctx context.Context, druid *v1alpha2.Druid, opts v1.UpdateOptions) (result *v1alpha2.Druid, err error) {
	result = &v1alpha2.Druid{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("druids").
		Name(druid.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(druid).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *druids) UpdateStatus(ctx context.Context, druid *v1alpha2.Druid, opts v1.UpdateOptions) (result *v1alpha2.Druid, err error) {
	result = &v1alpha2.Druid{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("druids").
		Name(druid.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(druid).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the druid and deletes it. Returns an error if one occurs.
func (c *druids) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("druids").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *druids) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("druids").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched druid.
func (c *druids) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha2.Druid, err error) {
	result = &v1alpha2.Druid{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("druids").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
