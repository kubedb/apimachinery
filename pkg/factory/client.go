package factory

import (
	kubedbscheme "kubedb.dev/apimachinery/client/clientset/versioned/scheme"

	cmscheme "github.com/cert-manager/cert-manager/pkg/client/clientset/versioned/scheme"
	promscheme "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/scheme"
	crdscheme "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/scheme"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	crscheme "kmodules.xyz/custom-resources/client/clientset/versioned/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	stashscheme "stash.appscode.dev/apimachinery/client/clientset/versioned/scheme"
)

func NewUncachedClient(cfg *rest.Config) (client.Client, error) {
	mapper, err := apiutil.NewDynamicRESTMapper(cfg)
	if err != nil {
		return nil, err
	}

	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := kubedbscheme.AddToScheme(scheme); err != nil {
		return nil, err
	}
	// cert-manager
	if err := cmscheme.AddToScheme(scheme); err != nil {
		return nil, err
	}
	// stash
	if err := stashscheme.AddToScheme(scheme); err != nil {
		return nil, err
	}
	// crd
	if err := crdscheme.AddToScheme(scheme); err != nil {
		return nil, err
	}
	// prometheus
	if err := promscheme.AddToScheme(scheme); err != nil {
		return nil, err
	}
	// appcatalog
	if err := crscheme.AddToScheme(scheme); err != nil {
		return nil, err
	}
	// https://github.com/kmodules/custom-resources/blob/master/apis/appcatalog/install/install.go
	if err := kubedbscheme.AddToScheme(scheme); err != nil {
		return nil, err
	}

	return client.New(cfg, client.Options{
		Scheme: scheme,
		Mapper: mapper,
		//Opts: client.WarningHandlerOptions{
		//	SuppressWarnings:   false,
		//	AllowDuplicateLogs: false,
		//},
	})
}
