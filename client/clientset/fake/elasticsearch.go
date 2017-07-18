package fake

import (
	aci "github.com/k8sdb/apimachinery/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/testing"
)

type FakeElasticsearch struct {
	Fake *testing.Fake
	ns   string
}

var elasticResource = schema.GroupVersionResource{Group: "kubedb.com", Version: "v1alpha1", Resource: aci.ResourceTypeElasticsearch}

// Get returns the Elasticsearch by name.
func (mock *FakeElasticsearch) Get(name string) (*aci.Elasticsearch, error) {
	obj, err := mock.Fake.
		Invokes(testing.NewGetAction(elasticResource, mock.ns, name), &aci.Elasticsearch{})

	if obj == nil {
		return nil, err
	}
	return obj.(*aci.Elasticsearch), err
}

// List returns a list of Elastics.
func (mock *FakeElasticsearch) List(opts metav1.ListOptions) (*aci.ElasticsearchList, error) {
	obj, err := mock.Fake.
		Invokes(testing.NewListAction(elasticResource, mock.ns, opts), &aci.Elasticsearch{})

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
		Invokes(testing.NewCreateAction(elasticResource, mock.ns, svc), &aci.Elasticsearch{})

	if obj == nil {
		return nil, err
	}
	return obj.(*aci.Elasticsearch), err
}

// Update updates a Elasticsearch.
func (mock *FakeElasticsearch) Update(svc *aci.Elasticsearch) (*aci.Elasticsearch, error) {
	obj, err := mock.Fake.
		Invokes(testing.NewUpdateAction(elasticResource, mock.ns, svc), &aci.Elasticsearch{})

	if obj == nil {
		return nil, err
	}
	return obj.(*aci.Elasticsearch), err
}

// Delete deletes a Elasticsearch by name.
func (mock *FakeElasticsearch) Delete(name string) error {
	_, err := mock.Fake.
		Invokes(testing.NewDeleteAction(elasticResource, mock.ns, name), &aci.Elasticsearch{})

	return err
}

func (mock *FakeElasticsearch) UpdateStatus(srv *aci.Elasticsearch) (*aci.Elasticsearch, error) {
	obj, err := mock.Fake.
		Invokes(testing.NewUpdateSubresourceAction(elasticResource, "status", mock.ns, srv), &aci.Elasticsearch{})

	if obj == nil {
		return nil, err
	}
	return obj.(*aci.Elasticsearch), err
}

func (mock *FakeElasticsearch) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return mock.Fake.
		InvokesWatch(testing.NewWatchAction(elasticResource, mock.ns, opts))
}
