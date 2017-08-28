package fake

import (
	aci "github.com/k8sdb/apimachinery/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/testing"
)

type FakeSnapshot struct {
	Fake *testing.Fake
	ns   string
}

var resourceSnapshot = aci.V1alpha1SchemeGroupVersion.WithResource(aci.ResourceTypeSnapshot)
var kindSnapshot = aci.V1alpha1SchemeGroupVersion.WithKind(aci.ResourceKindSnapshot)

// Get returns the Snapshot by name.
func (mock *FakeSnapshot) Get(name string) (*aci.Snapshot, error) {
	obj, err := mock.Fake.
		Invokes(testing.NewGetAction(resourceSnapshot, mock.ns, name), &aci.Snapshot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*aci.Snapshot), err
}

// List returns the a of Snapshots.
func (mock *FakeSnapshot) List(opts metav1.ListOptions) (*aci.SnapshotList, error) {
	obj, err := mock.Fake.
		Invokes(testing.NewListAction(resourceSnapshot, kindSnapshot, mock.ns, opts), &aci.Snapshot{})

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
		Invokes(testing.NewCreateAction(resourceSnapshot, mock.ns, svc), &aci.Snapshot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*aci.Snapshot), err
}

// Update updates a Snapshot.
func (mock *FakeSnapshot) Update(svc *aci.Snapshot) (*aci.Snapshot, error) {
	obj, err := mock.Fake.
		Invokes(testing.NewUpdateAction(resourceSnapshot, mock.ns, svc), &aci.Snapshot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*aci.Snapshot), err
}

// Delete deletes a Snapshot by name.
func (mock *FakeSnapshot) Delete(name string) error {
	_, err := mock.Fake.
		Invokes(testing.NewDeleteAction(resourceSnapshot, mock.ns, name), &aci.Snapshot{})

	return err
}

func (mock *FakeSnapshot) UpdateStatus(srv *aci.Snapshot) (*aci.Snapshot, error) {
	obj, err := mock.Fake.
		Invokes(testing.NewUpdateSubresourceAction(resourceSnapshot, "status", mock.ns, srv), &aci.Snapshot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*aci.Snapshot), err
}

func (mock *FakeSnapshot) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return mock.Fake.
		InvokesWatch(testing.NewWatchAction(resourceSnapshot, mock.ns, opts))
}

func (mock *FakeSnapshot) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (*aci.Snapshot, error) {
	obj, err := mock.Fake.
		Invokes(testing.NewPatchSubresourceAction(resourceSnapshot, mock.ns, name, data, subresources...), &aci.Snapshot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*aci.Snapshot), err
}
