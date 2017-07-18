package clientset

import (
	aci "github.com/k8sdb/apimachinery/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"
)

type ElasticNamespacer interface {
	Elastics(namespace string) ElasticInterface
}

type ElasticInterface interface {
	List(opts metav1.ListOptions) (*aci.ElasticsearchList, error)
	Get(name string) (*aci.Elasticsearch, error)
	Create(elastic *aci.Elasticsearch) (*aci.Elasticsearch, error)
	Update(elastic *aci.Elasticsearch) (*aci.Elasticsearch, error)
	Delete(name string) error
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	UpdateStatus(elastic *aci.Elasticsearch) (*aci.Elasticsearch, error)
}

type ElasticImpl struct {
	r  rest.Interface
	ns string
}

var _ ElasticInterface = &ElasticImpl{}

func newElastic(c *ExtensionClient, namespace string) *ElasticImpl {
	return &ElasticImpl{c.restClient, namespace}
}

func (c *ElasticImpl) List(opts metav1.ListOptions) (result *aci.ElasticsearchList, err error) {
	result = &aci.ElasticsearchList{}
	err = c.r.Get().
		Namespace(c.ns).
		Resource(aci.ResourceTypeElastic).
		VersionedParams(&opts, ExtendedCodec).
		Do().
		Into(result)
	return
}

func (c *ElasticImpl) Get(name string) (result *aci.Elasticsearch, err error) {
	result = &aci.Elasticsearch{}
	err = c.r.Get().
		Namespace(c.ns).
		Resource(aci.ResourceTypeElastic).
		Name(name).
		Do().
		Into(result)
	return
}

func (c *ElasticImpl) Create(elastic *aci.Elasticsearch) (result *aci.Elasticsearch, err error) {
	result = &aci.Elasticsearch{}
	err = c.r.Post().
		Namespace(c.ns).
		Resource(aci.ResourceTypeElastic).
		Body(elastic).
		Do().
		Into(result)
	return
}

func (c *ElasticImpl) Update(elastic *aci.Elasticsearch) (result *aci.Elasticsearch, err error) {
	result = &aci.Elasticsearch{}
	err = c.r.Put().
		Namespace(c.ns).
		Resource(aci.ResourceTypeElastic).
		Name(elastic.Name).
		Body(elastic).
		Do().
		Into(result)
	return
}

func (c *ElasticImpl) Delete(name string) (err error) {
	return c.r.Delete().
		Namespace(c.ns).
		Resource(aci.ResourceTypeElastic).
		Name(name).
		Do().
		Error()
}

func (c *ElasticImpl) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return c.r.Get().
		Prefix("watch").
		Namespace(c.ns).
		Resource(aci.ResourceTypeElastic).
		VersionedParams(&opts, ExtendedCodec).
		Watch()
}

func (c *ElasticImpl) UpdateStatus(elastic *aci.Elasticsearch) (result *aci.Elasticsearch, err error) {
	result = &aci.Elasticsearch{}
	err = c.r.Put().
		Namespace(c.ns).
		Resource(aci.ResourceTypeElastic).
		Name(elastic.Name).
		SubResource("status").
		Body(elastic).
		Do().
		Into(result)
	return
}
