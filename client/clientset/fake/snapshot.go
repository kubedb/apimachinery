package fake

import (
	aci "github.com/k8sdb/apimachinery/api"
	"k8s.io/kubernetes/pkg/api"
	schema "k8s.io/kubernetes/pkg/api/unversioned"
	testing "k8s.io/kubernetes/pkg/client/testing/core"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/watch"
)

type FakeSnapshot struct {
	Fake *testing.Fake
	ns   string
}

var snapshotResource = schema.GroupVersionResource{Group: "kubedb.com", Version: "v1alpha1", Resource: aci.ResourceTypeSnapshot}

// Get returns the Snapshot by name.
func (mock *FakeSnapshot) Get(name string) (*aci.Snapshot, error) {
	obj, err := mock.Fake.
		Invokes(testing.NewGetAction(snapshotResource, mock.ns, name), &aci.Snapshot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*aci.Snapshot), err
}

// List returns the a of Snapshots.
func (mock *FakeSnapshot) List(opts api.ListOptions) (*aci.SnapshotList, error) {
	obj, err := mock.Fake.
		Invokes(testing.NewListAction(snapshotResource, mock.ns, opts), &aci.Snapshot{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &aci.SnapshotList{}
	for _, item := range obj.(*aci.SnapshotList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Create creates a new Snapshot.
func (mock *FakeSnapshot) Create(svc *aci.Snapshot) (*aci.Snapshot, error) {
	obj, err := mock.Fake.
		Invokes(testing.NewCreateAction(snapshotResource, mock.ns, svc), &aci.Snapshot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*aci.Snapshot), err
}

// Update updates a Snapshot.
func (mock *FakeSnapshot) Update(svc *aci.Snapshot) (*aci.Snapshot, error) {
	obj, err := mock.Fake.
		Invokes(testing.NewUpdateAction(snapshotResource, mock.ns, svc), &aci.Snapshot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*aci.Snapshot), err
}

// Delete deletes a Snapshot by name.
func (mock *FakeSnapshot) Delete(name string) error {
	_, err := mock.Fake.
		Invokes(testing.NewDeleteAction(snapshotResource, mock.ns, name), &aci.Snapshot{})

	return err
}

func (mock *FakeSnapshot) UpdateStatus(srv *aci.Snapshot) (*aci.Snapshot, error) {
	obj, err := mock.Fake.
		Invokes(testing.NewUpdateSubresourceAction(snapshotResource, "status", mock.ns, srv), &aci.Snapshot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*aci.Snapshot), err
}

func (mock *FakeSnapshot) Watch(opts api.ListOptions) (watch.Interface, error) {
	return mock.Fake.
		InvokesWatch(testing.NewWatchAction(snapshotResource, mock.ns, opts))
}
