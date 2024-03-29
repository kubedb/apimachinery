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
	json "encoding/json"
	"fmt"
	"time"

	v1alpha1 "kmodules.xyz/custom-resources/apis/metrics/v1alpha1"
	metricsv1alpha1 "kmodules.xyz/custom-resources/client/applyconfiguration/metrics/v1alpha1"
	scheme "kmodules.xyz/custom-resources/client/clientset/versioned/scheme"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// MetricsConfigurationsGetter has a method to return a MetricsConfigurationInterface.
// A group's client should implement this interface.
type MetricsConfigurationsGetter interface {
	MetricsConfigurations() MetricsConfigurationInterface
}

// MetricsConfigurationInterface has methods to work with MetricsConfiguration resources.
type MetricsConfigurationInterface interface {
	Create(ctx context.Context, metricsConfiguration *v1alpha1.MetricsConfiguration, opts v1.CreateOptions) (*v1alpha1.MetricsConfiguration, error)
	Update(ctx context.Context, metricsConfiguration *v1alpha1.MetricsConfiguration, opts v1.UpdateOptions) (*v1alpha1.MetricsConfiguration, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.MetricsConfiguration, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.MetricsConfigurationList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.MetricsConfiguration, err error)
	Apply(ctx context.Context, metricsConfiguration *metricsv1alpha1.MetricsConfigurationApplyConfiguration, opts v1.ApplyOptions) (result *v1alpha1.MetricsConfiguration, err error)
	MetricsConfigurationExpansion
}

// metricsConfigurations implements MetricsConfigurationInterface
type metricsConfigurations struct {
	client rest.Interface
}

// newMetricsConfigurations returns a MetricsConfigurations
func newMetricsConfigurations(c *MetricsV1alpha1Client) *metricsConfigurations {
	return &metricsConfigurations{
		client: c.RESTClient(),
	}
}

// Get takes name of the metricsConfiguration, and returns the corresponding metricsConfiguration object, and an error if there is any.
func (c *metricsConfigurations) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.MetricsConfiguration, err error) {
	result = &v1alpha1.MetricsConfiguration{}
	err = c.client.Get().
		Resource("metricsconfigurations").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of MetricsConfigurations that match those selectors.
func (c *metricsConfigurations) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.MetricsConfigurationList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.MetricsConfigurationList{}
	err = c.client.Get().
		Resource("metricsconfigurations").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested metricsConfigurations.
func (c *metricsConfigurations) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Resource("metricsconfigurations").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a metricsConfiguration and creates it.  Returns the server's representation of the metricsConfiguration, and an error, if there is any.
func (c *metricsConfigurations) Create(ctx context.Context, metricsConfiguration *v1alpha1.MetricsConfiguration, opts v1.CreateOptions) (result *v1alpha1.MetricsConfiguration, err error) {
	result = &v1alpha1.MetricsConfiguration{}
	err = c.client.Post().
		Resource("metricsconfigurations").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(metricsConfiguration).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a metricsConfiguration and updates it. Returns the server's representation of the metricsConfiguration, and an error, if there is any.
func (c *metricsConfigurations) Update(ctx context.Context, metricsConfiguration *v1alpha1.MetricsConfiguration, opts v1.UpdateOptions) (result *v1alpha1.MetricsConfiguration, err error) {
	result = &v1alpha1.MetricsConfiguration{}
	err = c.client.Put().
		Resource("metricsconfigurations").
		Name(metricsConfiguration.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(metricsConfiguration).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the metricsConfiguration and deletes it. Returns an error if one occurs.
func (c *metricsConfigurations) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Resource("metricsconfigurations").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *metricsConfigurations) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Resource("metricsconfigurations").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched metricsConfiguration.
func (c *metricsConfigurations) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.MetricsConfiguration, err error) {
	result = &v1alpha1.MetricsConfiguration{}
	err = c.client.Patch(pt).
		Resource("metricsconfigurations").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}

// Apply takes the given apply declarative configuration, applies it and returns the applied metricsConfiguration.
func (c *metricsConfigurations) Apply(ctx context.Context, metricsConfiguration *metricsv1alpha1.MetricsConfigurationApplyConfiguration, opts v1.ApplyOptions) (result *v1alpha1.MetricsConfiguration, err error) {
	if metricsConfiguration == nil {
		return nil, fmt.Errorf("metricsConfiguration provided to Apply must not be nil")
	}
	patchOpts := opts.ToPatchOptions()
	data, err := json.Marshal(metricsConfiguration)
	if err != nil {
		return nil, err
	}
	name := metricsConfiguration.Name
	if name == nil {
		return nil, fmt.Errorf("metricsConfiguration.Name must be provided to Apply")
	}
	result = &v1alpha1.MetricsConfiguration{}
	err = c.client.Patch(types.ApplyPatchType).
		Resource("metricsconfigurations").
		Name(*name).
		VersionedParams(&patchOpts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
