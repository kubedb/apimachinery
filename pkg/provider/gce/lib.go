package gce

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/k8sdb/apimachinery/pkg/provider/extpoints"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	gcs "google.golang.org/api/storage/v1"
	kapi "k8s.io/kubernetes/pkg/api"
	k8serr "k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/api/unversioned"
	kstorage "k8s.io/kubernetes/pkg/apis/storage"
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
)

type biblio struct{}

func init() {
	extpoints.CloudProviders.Register(new(biblio), "gce")
}

const (
	SecretKey = "gce"
)

func (b *biblio) GetSecretMountPath() string {
	return "/var/credentials"
}

func (b *biblio) GetCredentialData(client clientset.Interface, secretName, namespace string) (map[string]string, error) {
	secret, err := client.Core().Secrets(namespace).Get(secretName)
	if err != nil {
		return nil, err
	}

	credData := secret.Data[SecretKey]
	data := make(map[string]string)
	err = json.Unmarshal(credData, &data)
	return data, err
}

func newGCPClient(cred map[string]string, scope ...string) (*http.Client, error) {
	credGCP, err := json.Marshal(cred)

	conf, err := google.JWTConfigFromJSON(credGCP, scope...)
	if err != nil {
		return nil, errors.New("failed to connect gce")
	}
	return conf.Client(oauth2.NoContext), nil
}

func (b *biblio) CheckBucketAccess(bucketName string, cloudCredential map[string]string, writeAccess bool) error {
	m, err := newGCPClient(cloudCredential, gcs.DevstorageFullControlScope)
	if err != nil {
		return err
	}
	client, err := gcs.New(m)
	if err != nil {
		return errors.New("failed to connect gce")
	}

	controls, err := client.BucketAccessControls.List(bucketName).Do()
	if err != nil {
		return errors.New("failed to connect gce")
	}

	for _, item := range controls.Items {
		if item.Bucket == bucketName {
			if writeAccess && (item.Role == "OWNER" || item.Role == "WRITER") {
				return nil
			}
			if !writeAccess && (item.Role == "OWNER" || item.Role == "WRITER" || item.Role == "READER") {
				return nil
			}
		}
	}
	return errors.New("Insufficient Permission of Credential")
}

func (b *biblio) GetStorageClassName(client clientset.Interface) (string, error) {
	defaultName := "pd-ssd"
	_, err := client.Storage().StorageClasses().Get(defaultName)
	if err != nil {
		if !k8serr.IsNotFound(err) {
			return "", err
		} else {
			storageClass := &kstorage.StorageClass{
				TypeMeta: unversioned.TypeMeta{
					Kind:       "StorageClass",
					APIVersion: "storage.k8s.io/v1beta1",
				},
				ObjectMeta: kapi.ObjectMeta{
					Name: defaultName,
				},
				Provisioner: "kubernetes.io/gce-pd",
			}
			if _, err := client.Storage().StorageClasses().Create(storageClass); err != nil {
				return "", err
			}
		}
	}
	return defaultName, nil
}
