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

package v1

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
	v1 "kubedb.dev/apimachinery/apis/kubedb/v1"
	scheme "kubedb.dev/apimachinery/client/clientset/versioned/scheme"
)

// MySQLsGetter has a method to return a MySQLInterface.
// A group's client should implement this interface.
type MySQLsGetter interface {
	MySQLs(namespace string) MySQLInterface
}

// MySQLInterface has methods to work with MySQL resources.
type MySQLInterface interface {
	Create(ctx context.Context, mySQL *v1.MySQL, opts metav1.CreateOptions) (*v1.MySQL, error)
	Update(ctx context.Context, mySQL *v1.MySQL, opts metav1.UpdateOptions) (*v1.MySQL, error)
	UpdateStatus(ctx context.Context, mySQL *v1.MySQL, opts metav1.UpdateOptions) (*v1.MySQL, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.MySQL, error)
	List(ctx context.Context, opts metav1.ListOptions) (*v1.MySQLList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.MySQL, err error)
	MySQLExpansion
}

// mySQLs implements MySQLInterface
type mySQLs struct {
	client rest.Interface
	ns     string
}

// newMySQLs returns a MySQLs
func newMySQLs(c *KubedbV1Client, namespace string) *mySQLs {
	return &mySQLs{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the mySQL, and returns the corresponding mySQL object, and an error if there is any.
func (c *mySQLs) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.MySQL, err error) {
	result = &v1.MySQL{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("mysqls").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of MySQLs that match those selectors.
func (c *mySQLs) List(ctx context.Context, opts metav1.ListOptions) (result *v1.MySQLList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1.MySQLList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("mysqls").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested mySQLs.
func (c *mySQLs) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("mysqls").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a mySQL and creates it.  Returns the server's representation of the mySQL, and an error, if there is any.
func (c *mySQLs) Create(ctx context.Context, mySQL *v1.MySQL, opts metav1.CreateOptions) (result *v1.MySQL, err error) {
	result = &v1.MySQL{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("mysqls").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(mySQL).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a mySQL and updates it. Returns the server's representation of the mySQL, and an error, if there is any.
func (c *mySQLs) Update(ctx context.Context, mySQL *v1.MySQL, opts metav1.UpdateOptions) (result *v1.MySQL, err error) {
	result = &v1.MySQL{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("mysqls").
		Name(mySQL.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(mySQL).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *mySQLs) UpdateStatus(ctx context.Context, mySQL *v1.MySQL, opts metav1.UpdateOptions) (result *v1.MySQL, err error) {
	result = &v1.MySQL{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("mysqls").
		Name(mySQL.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(mySQL).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the mySQL and deletes it. Returns an error if one occurs.
func (c *mySQLs) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("mysqls").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *mySQLs) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("mysqls").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched mySQL.
func (c *mySQLs) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.MySQL, err error) {
	result = &v1.MySQL{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("mysqls").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
