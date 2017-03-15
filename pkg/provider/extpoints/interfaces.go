package extpoints

import (
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
)

type CloudProvider interface {
	GetSecretMountPath() string
	GetCredentialData(clientset.Interface, string, string) (map[string]string, error)
	CheckBucketAccess(string, map[string]string, bool) error
	GetStorageClassName(clientset.Interface) (string, error)
}
