package fake

import (
	aci "github.com/k8sdb/apimachinery/api"
	"k8s.io/kubernetes/pkg/api"
	schema "k8s.io/kubernetes/pkg/api/unversioned"
	testing "k8s.io/kubernetes/pkg/client/testing/core"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/watch"
)

type FakeDatabaseSnapshot struct {
	Fake *testing.Fake
	ns   string
}

var databaseSnapshotResource = schema.GroupVersionResource{Group: "k8sdb.com", Version: "v1beta1", Resource: "databasesnapshots"}

// Get returns the DatabaseSnapshot by name.
func (mock *FakeDatabaseSnapshot) Get(name string) (*aci.DatabaseSnapshot, error) {
	obj, err := mock.Fake.
		Invokes(testing.NewGetAction(databaseSnapshotResource, mock.ns, name), &aci.DatabaseSnapshot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*aci.DatabaseSnapshot), err
}

// List returns the a of DatabaseSnapshots.
func (mock *FakeDatabaseSnapshot) List(opts api.ListOptions) (*aci.DatabaseSnapshotList, error) {
	obj, err := mock.Fake.
		Invokes(testing.NewListAction(databaseSnapshotResource, mock.ns, opts), &aci.DatabaseSnapshot{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &aci.DatabaseSnapshotList{}
	for _, item := range obj.(*aci.DatabaseSnapshotList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Create creates a new DatabaseSnapshot.
func (mock *FakeDatabaseSnapshot) Create(svc *aci.DatabaseSnapshot) (*aci.DatabaseSnapshot, error) {
	obj, err := mock.Fake.
		Invokes(testing.NewCreateAction(databaseSnapshotResource, mock.ns, svc), &aci.DatabaseSnapshot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*aci.DatabaseSnapshot), err
}

// Update updates a DatabaseSnapshot.
func (mock *FakeDatabaseSnapshot) Update(svc *aci.DatabaseSnapshot) (*aci.DatabaseSnapshot, error) {
	obj, err := mock.Fake.
		Invokes(testing.NewUpdateAction(databaseSnapshotResource, mock.ns, svc), &aci.DatabaseSnapshot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*aci.DatabaseSnapshot), err
}

// Delete deletes a DatabaseSnapshot by name.
func (mock *FakeDatabaseSnapshot) Delete(name string) error {
	_, err := mock.Fake.
		Invokes(testing.NewDeleteAction(databaseSnapshotResource, mock.ns, name), &aci.DatabaseSnapshot{})

	return err
}

func (mock *FakeDatabaseSnapshot) UpdateStatus(srv *aci.DatabaseSnapshot) (*aci.DatabaseSnapshot, error) {
	obj, err := mock.Fake.
		Invokes(testing.NewUpdateSubresourceAction(databaseSnapshotResource, "status", mock.ns, srv), &aci.DatabaseSnapshot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*aci.DatabaseSnapshot), err
}

func (mock *FakeDatabaseSnapshot) Watch(opts api.ListOptions) (watch.Interface, error) {
	return mock.Fake.
		InvokesWatch(testing.NewWatchAction(databaseSnapshotResource, mock.ns, opts))
}
