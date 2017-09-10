package controller

import (
	"time"

	tcs "github.com/k8sdb/apimachinery/client/internalclientset/typed/kubedb/internalversion"
	clientset "k8s.io/client-go/kubernetes"
)

type Controller struct {
	// Kubernetes client
	Client clientset.Interface
	// ThirdPartyExtension client
	ExtClient tcs.KubedbInterface
}

const (
	sleepDuration = time.Second * 10
)
