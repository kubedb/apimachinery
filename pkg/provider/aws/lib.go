package aws

import (
	"errors"

	"github.com/appscode/go/types"
	_aws "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	_s3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/k8sdb/apimachinery/pkg/provider/extpoints"
	kapi "k8s.io/kubernetes/pkg/api"
	k8serr "k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/api/unversioned"
	kstorage "k8s.io/kubernetes/pkg/apis/storage"
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
)

type biblio struct{}

func init() {
	extpoints.CloudProviders.Register(new(biblio), "aws")
}

const (
	AWSCredentialAccessKeyID     = "aws_access_key_id"
	AWSCredentialSecretAccessKey = "aws_secret_access_key"
	DefaultStorageClass          = "gp2"
)

func (b *biblio) GetSecretMountPath() string {
	return "/var/credentials/aws"
}

func (b *biblio) GetCredentialData(client clientset.Interface, secretName, namespace string) (map[string]string, error) {
	secret, err := client.Core().Secrets(namespace).Get(secretName)
	if err != nil {
		return nil, err
	}

	keyId := secret.Data[AWSCredentialAccessKeyID]
	accessKey := secret.Data[AWSCredentialSecretAccessKey]
	return map[string]string{
		AWSCredentialAccessKeyID:     string(keyId),
		AWSCredentialSecretAccessKey: string(accessKey),
	}, nil
}

func newSession(region string, cred map[string]string) *session.Session {
	id := cred[AWSCredentialAccessKeyID]
	secret := cred[AWSCredentialSecretAccessKey]
	return session.New(&_aws.Config{
		Region:      types.StringP(region),
		Credentials: credentials.NewStaticCredentials(id, secret, ""),
	})
}

func getAWSSession(bucketName string, cloudCredential map[string]string) (*session.Session, error) {
	bucketRegion := "us-west-2"
	newRegion := ""
	session := newSession(bucketRegion, cloudCredential)
	s3 := _s3.New(session)

	bucketLocationOutput, err := s3.GetBucketLocation(&_s3.GetBucketLocationInput{
		Bucket: types.StringP(bucketName),
	})
	if err != nil {
		return nil, err
	} else {
		if bucketLocationOutput.LocationConstraint != nil {
			newRegion = *bucketLocationOutput.LocationConstraint
		}
	}

	if newRegion != "" && newRegion != bucketRegion {
		session = newSession(bucketRegion, cloudCredential)
	}
	return session, nil
}

func (b *biblio) CheckBucketAccess(bucketName string, cloudCredential map[string]string, writeAccess bool) error {
	session, err := getAWSSession(bucketName, cloudCredential)
	if err != nil {
		return err
	}
	s3 := _s3.New(session)
	controls, err := s3.GetBucketAcl(&_s3.GetBucketAclInput{
		Bucket: types.StringP(bucketName),
	})
	if err != nil {
		return errors.New("failed to connect gce")
	}

	for _, grant := range controls.Grants {
		permission := _aws.StringValue(grant.Permission)
		if writeAccess && (permission == "FULL_CONTROL" || permission == "WRITE") {
			return nil
		}
		if !writeAccess && (permission == "FULL_CONTROL" || permission == "WRITE" || permission == "READ") {
			return nil
		}
	}
	return nil
}

func (b *biblio) GetStorageClassName(client clientset.Interface) (string, error) {
	defaultName := DefaultStorageClass
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
				Provisioner: "kubernetes.io/aws-ebs",
			}
			if _, err := client.Storage().StorageClasses().Create(storageClass); err != nil {
				return "", err
			}
		}
	}
	return defaultName, nil
}
