package fake

import (
	aci "github.com/k8sdb/apimachinery/api"
	"k8s.io/kubernetes/pkg/api"
	schema "k8s.io/kubernetes/pkg/api/unversioned"
	testing "k8s.io/kubernetes/pkg/client/testing/core"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/watch"
)

type FakeDeletedDatabase struct {
	Fake *testing.Fake
	ns   string
}

var deletedDatabaseResource = schema.GroupVersionResource{Group: "k8sdb.com", Version: "v1beta1", Resource: aci.ResourceTypeDeletedDatabase}

// Get returns the DeletedDatabase by name.
func (mock *FakeDeletedDatabase) Get(name string) (*aci.DeletedDatabase, error) {
	obj, err := mock.Fake.
		Invokes(testing.NewGetAction(deletedDatabaseResource, mock.ns, name), &aci.DeletedDatabase{})

	if obj == nil {
		return nil, err
	}
	return obj.(*aci.DeletedDatabase), err
}

// List returns the a of DeletedDatabases.
func (mock *FakeDeletedDatabase) List(opts api.ListOptions) (*aci.DeletedDatabaseList, error) {
	obj, err := mock.Fake.
		Invokes(testing.NewListAction(deletedDatabaseResource, mock.ns, opts), &aci.DeletedDatabase{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &aci.DeletedDatabaseList{}
	for _, item := range obj.(*aci.DeletedDatabaseList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Create creates a new DeletedDatabase.
func (mock *FakeDeletedDatabase) Create(svc *aci.DeletedDatabase) (*aci.DeletedDatabase, error) {
	obj, err := mock.Fake.
		Invokes(testing.NewCreateAction(deletedDatabaseResource, mock.ns, svc), &aci.DeletedDatabase{})

	if obj == nil {
		return nil, err
	}
	return obj.(*aci.DeletedDatabase), err
}

// Update updates a DeletedDatabase.
func (mock *FakeDeletedDatabase) Update(svc *aci.DeletedDatabase) (*aci.DeletedDatabase, error) {
	obj, err := mock.Fake.
		Invokes(testing.NewUpdateAction(deletedDatabaseResource, mock.ns, svc), &aci.DeletedDatabase{})

	if obj == nil {
		return nil, err
	}
	return obj.(*aci.DeletedDatabase), err
}

// Delete deletes a DeletedDatabase by name.
func (mock *FakeDeletedDatabase) Delete(name string) error {
	_, err := mock.Fake.
		Invokes(testing.NewDeleteAction(deletedDatabaseResource, mock.ns, name), &aci.DeletedDatabase{})

	return err
}

func (mock *FakeDeletedDatabase) UpdateStatus(srv *aci.DeletedDatabase) (*aci.DeletedDatabase, error) {
	obj, err := mock.Fake.
		Invokes(testing.NewUpdateSubresourceAction(deletedDatabaseResource, "status", mock.ns, srv), &aci.DeletedDatabase{})

	if obj == nil {
		return nil, err
	}
	return obj.(*aci.DeletedDatabase), err
}

func (mock *FakeDeletedDatabase) Watch(opts api.ListOptions) (watch.Interface, error) {
	return mock.Fake.
		InvokesWatch(testing.NewWatchAction(deletedDatabaseResource, mock.ns, opts))
}
