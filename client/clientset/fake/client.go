package fake

import (
	"flag"

	"github.com/appscode/log"
	_ "github.com/k8sdb/apimachinery/api/install"
	"k8s.io/kubernetes/pkg/client/clientset_generated/release_1_5/fake"
	"k8s.io/kubernetes/pkg/runtime"
)

type ClientSets struct {
	*fake.Clientset
	ACExtensionClient *FakeExtensionClient
}

func NewFakeClient(objects ...runtime.Object) *ClientSets {
	return &ClientSets{
		Clientset:         fake.NewSimpleClientset(objects...),
		ACExtensionClient: NewFakeExtensionClient(objects...),
	}
}
