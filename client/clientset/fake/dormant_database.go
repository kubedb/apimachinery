package fake

import (
	aci "github.com/k8sdb/apimachinery/api"
	"k8s.io/kubernetes/pkg/api"
	schema "k8s.io/kubernetes/pkg/api/unversioned"
	testing "k8s.io/kubernetes/pkg/client/testing/core"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/watch"
)

type FakeDormantDatabase struct {
	Fake *testing.Fake
	ns   string
}

var dormantDatabaseResource = schema.GroupVersionResource{Group: "k8sdb.com", Version: "v1beta1", Resource: aci.ResourceTypeDormantDatabase}

// Get returns the DormantDatabase by name.
func (mock *FakeDormantDatabase) Get(name string) (*aci.DormantDatabase, error) {
	obj, err := mock.Fake.
		Invokes(testing.NewGetAction(dormantDatabaseResource, mock.ns, name), &aci.DormantDatabase{})

	if obj == nil {
		return nil, err
	}
	return obj.(*aci.DormantDatabase), err
}

// List returns the a of DormantDatabases.
func (mock *FakeDormantDatabase) List(opts api.ListOptions) (*aci.DormantDatabaseList, error) {
	obj, err := mock.Fake.
		Invokes(testing.NewListAction(dormantDatabaseResource, mock.ns, opts), &aci.DormantDatabase{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &aci.DormantDatabaseList{}
	for _, item := range obj.(*aci.DormantDatabaseList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Create creates a new DormantDatabase.
func (mock *FakeDormantDatabase) Create(svc *aci.DormantDatabase) (*aci.DormantDatabase, error) {
	obj, err := mock.Fake.
		Invokes(testing.NewCreateAction(dormantDatabaseResource, mock.ns, svc), &aci.DormantDatabase{})

	if obj == nil {
		return nil, err
	}
	return obj.(*aci.DormantDatabase), err
}

// Update updates a DormantDatabase.
func (mock *FakeDormantDatabase) Update(svc *aci.DormantDatabase) (*aci.DormantDatabase, error) {
	obj, err := mock.Fake.
		Invokes(testing.NewUpdateAction(dormantDatabaseResource, mock.ns, svc), &aci.DormantDatabase{})

	if obj == nil {
		return nil, err
	}
	return obj.(*aci.DormantDatabase), err
}

// Delete deletes a DormantDatabase by name.
func (mock *FakeDormantDatabase) Delete(name string) error {
	_, err := mock.Fake.
		Invokes(testing.NewDeleteAction(dormantDatabaseResource, mock.ns, name), &aci.DormantDatabase{})

	return err
}

func (mock *FakeDormantDatabase) UpdateStatus(srv *aci.DormantDatabase) (*aci.DormantDatabase, error) {
	obj, err := mock.Fake.
		Invokes(testing.NewUpdateSubresourceAction(dormantDatabaseResource, "status", mock.ns, srv), &aci.DormantDatabase{})

	if obj == nil {
		return nil, err
	}
	return obj.(*aci.DormantDatabase), err
}

func (mock *FakeDormantDatabase) Watch(opts api.ListOptions) (watch.Interface, error) {
	return mock.Fake.
		InvokesWatch(testing.NewWatchAction(dormantDatabaseResource, mock.ns, opts))
}
