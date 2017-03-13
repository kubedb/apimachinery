package controller

import (
	"time"

	tcs "github.com/k8sdb/apimachinery/client/clientset"
	rest "k8s.io/kubernetes/pkg/client/restclient"
)

type Controller struct {
	Client tcs.ExtensionInterface
	// sync time to sync the list.
	SyncPeriod time.Duration
}

func New(c *rest.Config) *Controller {
	return &Controller{
		Client:     tcs.NewExtensionsForConfigOrDie(c),
		SyncPeriod: time.Minute * 2,
	}
}

// Blocks caller. Intended to be called as a Go routine.
func (w *Controller) RunAndHold() {

}
