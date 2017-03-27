package fake

import (
	"github.com/k8sdb/apimachinery/client/clientset"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apimachinery/registered"
	testing "k8s.io/kubernetes/pkg/client/testing/core"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/watch"
	rest "k8s.io/kubernetes/pkg/client/restclient"
)

type FakeExtensionClient struct {
	*testing.Fake
}

func NewFakeExtensionClient(objects ...runtime.Object) *FakeExtensionClient {
	o := testing.NewObjectTracker(api.Scheme, api.Codecs.UniversalDecoder())
	for _, obj := range objects {
		if obj.GetObjectKind().GroupVersionKind().Group == "k8sdb.com" {
			if err := o.Add(obj); err != nil {
				panic(err)
			}
		}
	}

	fakePtr := testing.Fake{}
	fakePtr.AddReactor("*", "*", testing.ObjectReaction(o, registered.RESTMapper()))

	fakePtr.AddWatchReactor("*", testing.DefaultWatchReactor(watch.NewFake(), nil))

	return &FakeExtensionClient{&fakePtr}
}

func (m *FakeExtensionClient) DatabaseSnapshot(ns string) client.DatabaseSnapshotInterface {
	return &FakeDatabaseSnapshot{m.Fake, ns}
}

func (m *FakeExtensionClient) DeletedDatabase(ns string) client.DeletedDatabaseInterface {
	return &FakeDeletedDatabase{m.Fake, ns}
}

func (m *FakeExtensionClient) Elastic(ns string) client.ElasticInterface {
	return &FakeElastic{m.Fake, ns}
}

func (m *FakeExtensionClient) Postgres(ns string) client.PostgresInterface {
	return &FakePostgres{m.Fake, ns}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeExtensionClient) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}