package fake

import (
	aci "github.com/k8sdb/apimachinery/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/testing"
)

type FakeElasticsearch struct {
	Fake *testing.Fake
	ns   string
}

var resourceElasticsearch = aci.V1alpha1SchemeGroupVersion.WithResource(aci.ResourceTypeElasticsearch)
var kindElasticsearch = aci.V1alpha1SchemeGroupVersion.WithKind(aci.ResourceKindElasticsearch)

// Get returns the Elasticsearch by name.
func (mock *FakeElasticsearch) Get(name string) (*aci.Elasticsearch, error) {
	obj, err := mock.Fake.
		Invokes(testing.NewGetAction(resourceElasticsearch, mock.ns, name), &aci.Elasticsearch{})

	if obj == nil {
		return nil, err
	}
	return obj.(*aci.Elasticsearch), err
}

// List returns a list of Elastics.
func (mock *FakeElasticsearch) List(opts metav1.ListOptions) (*aci.ElasticsearchList, error) {
	obj, err := mock.Fake.
		Invokes(testing.NewListAction(resourceElasticsearch, kindElasticsearch, mock.ns, opts), &aci.Elasticsearch{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &aci.ElasticsearchList{}
	for _, item := range obj.(*aci.ElasticsearchList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Create creates a new Elasticsearch.
func (mock *FakeElasticsearch) Create(svc *aci.Elasticsearch) (*aci.Elasticsearch, error) {
	obj, err := mock.Fake.
		Invokes(testing.NewCreateAction(resourceElasticsearch, mock.ns, svc), &aci.Elasticsearch{})

	if obj == nil {
		return nil, err
	}
	return obj.(*aci.Elasticsearch), err
}

// Update updates a Elasticsearch.
func (mock *FakeElasticsearch) Update(svc *aci.Elasticsearch) (*aci.Elasticsearch, error) {
	obj, err := mock.Fake.
		Invokes(testing.NewUpdateAction(resourceElasticsearch, mock.ns, svc), &aci.Elasticsearch{})

	if obj == nil {
		return nil, err
	}
	return obj.(*aci.Elasticsearch), err
}

// Delete deletes a Elasticsearch by name.
func (mock *FakeElasticsearch) Delete(name string) error {
	_, err := mock.Fake.
		Invokes(testing.NewDeleteAction(resourceElasticsearch, mock.ns, name), &aci.Elasticsearch{})

	return err
}

func (mock *FakeElasticsearch) UpdateStatus(srv *aci.Elasticsearch) (*aci.Elasticsearch, error) {
	obj, err := mock.Fake.
		Invokes(testing.NewUpdateSubresourceAction(resourceElasticsearch, "status", mock.ns, srv), &aci.Elasticsearch{})

	if obj == nil {
		return nil, err
	}
	return obj.(*aci.Elasticsearch), err
}

func (mock *FakeElasticsearch) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return mock.Fake.
		InvokesWatch(testing.NewWatchAction(resourceElasticsearch, mock.ns, opts))
}

func (mock *FakeElasticsearch) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (*aci.Elasticsearch, error) {
	obj, err := mock.Fake.
		Invokes(testing.NewPatchSubresourceAction(resourceElasticsearch, mock.ns, name, data, subresources...), &aci.Elasticsearch{})

	if obj == nil {
		return nil, err
	}
	return obj.(*aci.Elasticsearch), err
}
