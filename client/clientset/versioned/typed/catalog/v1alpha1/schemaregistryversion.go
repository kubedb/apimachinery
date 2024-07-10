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
	v1alpha1 "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	scheme "kubedb.dev/apimachinery/client/clientset/versioned/scheme"
)

// SchemaRegistryVersionsGetter has a method to return a SchemaRegistryVersionInterface.
// A group's client should implement this interface.
type SchemaRegistryVersionsGetter interface {
	SchemaRegistryVersions() SchemaRegistryVersionInterface
}

// SchemaRegistryVersionInterface has methods to work with SchemaRegistryVersion resources.
type SchemaRegistryVersionInterface interface {
	Create(ctx context.Context, schemaRegistryVersion *v1alpha1.SchemaRegistryVersion, opts v1.CreateOptions) (*v1alpha1.SchemaRegistryVersion, error)
	Update(ctx context.Context, schemaRegistryVersion *v1alpha1.SchemaRegistryVersion, opts v1.UpdateOptions) (*v1alpha1.SchemaRegistryVersion, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.SchemaRegistryVersion, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.SchemaRegistryVersionList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.SchemaRegistryVersion, err error)
	SchemaRegistryVersionExpansion
}

// schemaRegistryVersions implements SchemaRegistryVersionInterface
type schemaRegistryVersions struct {
	client rest.Interface
}

// newSchemaRegistryVersions returns a SchemaRegistryVersions
func newSchemaRegistryVersions(c *CatalogV1alpha1Client) *schemaRegistryVersions {
	return &schemaRegistryVersions{
		client: c.RESTClient(),
	}
}

// Get takes name of the schemaRegistryVersion, and returns the corresponding schemaRegistryVersion object, and an error if there is any.
func (c *schemaRegistryVersions) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.SchemaRegistryVersion, err error) {
	result = &v1alpha1.SchemaRegistryVersion{}
	err = c.client.Get().
		Resource("schemaregistryversions").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of SchemaRegistryVersions that match those selectors.
func (c *schemaRegistryVersions) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.SchemaRegistryVersionList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.SchemaRegistryVersionList{}
	err = c.client.Get().
		Resource("schemaregistryversions").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested schemaRegistryVersions.
func (c *schemaRegistryVersions) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Resource("schemaregistryversions").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a schemaRegistryVersion and creates it.  Returns the server's representation of the schemaRegistryVersion, and an error, if there is any.
func (c *schemaRegistryVersions) Create(ctx context.Context, schemaRegistryVersion *v1alpha1.SchemaRegistryVersion, opts v1.CreateOptions) (result *v1alpha1.SchemaRegistryVersion, err error) {
	result = &v1alpha1.SchemaRegistryVersion{}
	err = c.client.Post().
		Resource("schemaregistryversions").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(schemaRegistryVersion).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a schemaRegistryVersion and updates it. Returns the server's representation of the schemaRegistryVersion, and an error, if there is any.
func (c *schemaRegistryVersions) Update(ctx context.Context, schemaRegistryVersion *v1alpha1.SchemaRegistryVersion, opts v1.UpdateOptions) (result *v1alpha1.SchemaRegistryVersion, err error) {
	result = &v1alpha1.SchemaRegistryVersion{}
	err = c.client.Put().
		Resource("schemaregistryversions").
		Name(schemaRegistryVersion.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(schemaRegistryVersion).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the schemaRegistryVersion and deletes it. Returns an error if one occurs.
func (c *schemaRegistryVersions) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Resource("schemaregistryversions").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *schemaRegistryVersions) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Resource("schemaregistryversions").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched schemaRegistryVersion.
func (c *schemaRegistryVersions) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.SchemaRegistryVersion, err error) {
	result = &v1alpha1.SchemaRegistryVersion{}
	err = c.client.Patch(pt).
		Resource("schemaregistryversions").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
