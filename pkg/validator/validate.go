package validator

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"

	"github.com/appscode/log"
	otx "github.com/appscode/osm/pkg/context"
	"github.com/graymeta/stow"
	"github.com/graymeta/stow/azure"
	gcs "github.com/graymeta/stow/google"
	"github.com/graymeta/stow/local"
	"github.com/graymeta/stow/s3"
	"github.com/graymeta/stow/swift"
	tapi "github.com/k8sdb/apimachinery/api"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

func ValidateStorageSpec(client clientset.Interface, spec *tapi.StorageSpec) (*tapi.StorageSpec, error) {
	if spec == nil {
		return nil, nil
	}

	if spec.Class == "" {
		return nil, fmt.Errorf(`Object 'Class' is missing in '%v'`, *spec)
	}

	if _, err := client.StorageV1().StorageClasses().Get(spec.Class, metav1.GetOptions{}); err != nil {
		if kerr.IsNotFound(err) {
			return nil, fmt.Errorf(`Spec.Storage.Class "%v" not found`, spec.Class)
		}
		return nil, err
	}

	if len(spec.AccessModes) == 0 {
		spec.AccessModes = []apiv1.PersistentVolumeAccessMode{
			apiv1.ReadWriteOnce,
		}
		log.Infof(`Using "%v" as AccessModes in "%v"`, apiv1.ReadWriteOnce, *spec)
	}

	if val, found := spec.Resources.Requests[apiv1.ResourceStorage]; found {
		if val.Value() <= 0 {
			return nil, errors.New("Invalid ResourceStorage request")
		}
	} else {
		return nil, errors.New("Missing ResourceStorage request")
	}

	return spec, nil
}

func ValidateBackupSchedule(spec *tapi.BackupScheduleSpec) error {
	if spec == nil {
		return nil
	}
	// CronExpression can't be empty
	if spec.CronExpression == "" {
		return errors.New("Invalid cron expression")
	}

	return ValidateSnapshotSpec(spec.SnapshotStorageSpec)
}

func ValidateSnapshotSpec(spec tapi.SnapshotStorageSpec) error {
	// BucketName can't be empty
	if spec.S3 == nil && spec.GCS == nil && spec.Azure == nil && spec.Swift == nil && spec.Local == nil {
		return errors.New("No storage provider is configured.")
	}

	// Need to provide Storage credential secret
	if spec.StorageSecretName == "" {
		return fmt.Errorf(`Object 'SecretName' is missing in '%v'`, spec.StorageSecretName)
	}
	return nil
}

func CheckBucketAccess(client clientset.Interface, spec tapi.SnapshotStorageSpec, namespace string) error {
	cfg, err := CreateOSMContext(client, spec, namespace)
	if err != nil {
		return err
	}
	loc, err := stow.Dial(cfg.Provider, cfg.Config)
	if err != nil {
		return err
	}
	container, err := loc.Container(cfg.Name)
	if err != nil {
		return err
	}
	r := bytes.NewReader([]byte("CheckBucketAccess"))
	item, err := container.Put(".kubedb", r, r.Size(), nil)
	if err != nil {
		return err
	}
	if err := container.RemoveItem(item.ID()); err != nil {
		return err
	}
	return nil
}

func CreateOSMContext(client clientset.Interface, spec tapi.SnapshotStorageSpec, namespace string) (*otx.Context, error) {
	secret, err := client.CoreV1().Secrets(namespace).Get(spec.StorageSecretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	nc := &otx.Context{
		Config: stow.ConfigMap{},
	}

	if spec.S3 != nil {
		nc.Provider = s3.Kind
		nc.Name = spec.S3.Bucket
		nc.Config[s3.ConfigAccessKeyID] = string(secret.Data[tapi.AWS_ACCESS_KEY_ID])
		nc.Config[s3.ConfigEndpoint] = spec.S3.Endpoint
		nc.Config[s3.ConfigRegion] = spec.S3.Region
		nc.Config[s3.ConfigSecretKey] = string(secret.Data[tapi.AWS_SECRET_ACCESS_KEY])
		if u, err := url.Parse(spec.S3.Endpoint); err == nil {
			nc.Config[s3.ConfigDisableSSL] = strconv.FormatBool(u.Scheme == "http")
		}
		return nc, nil
	} else if spec.GCS != nil {
		nc.Provider = gcs.Kind
		nc.Name = spec.GCS.Bucket
		nc.Config[gcs.ConfigProjectId] = string(secret.Data[tapi.GOOGLE_PROJECT_ID])
		nc.Config[gcs.ConfigJSON] = string(secret.Data[tapi.GOOGLE_SERVICE_ACCOUNT_JSON_KEY])
		return nc, nil
	} else if spec.Azure != nil {
		nc.Provider = azure.Kind
		nc.Name = spec.Azure.Container
		nc.Config[azure.ConfigAccount] = string(secret.Data[tapi.AZURE_ACCOUNT_NAME])
		nc.Config[azure.ConfigKey] = string(secret.Data[tapi.AZURE_ACCOUNT_KEY])
		return nc, nil
	} else if spec.Local != nil {
		nc.Provider = local.Kind
		nc.Name = "stash"
		nc.Config[local.ConfigKeyPath] = spec.Local.Path
		return nc, nil
	} else if spec.Swift != nil {
		nc.Provider = swift.Kind
		nc.Name = spec.Swift.Container
		nc.Config[swift.ConfigKey] = string(secret.Data[tapi.OS_PASSWORD])
		nc.Config[swift.ConfigTenantAuthURL] = string(secret.Data[tapi.OS_AUTH_URL])
		nc.Config[swift.ConfigTenantName] = string(secret.Data[tapi.OS_TENANT_NAME])
		nc.Config[swift.ConfigUsername] = string(secret.Data[tapi.OS_USERNAME])
		return nc, nil
	}
	return nil, errors.New("No storage provider is configured.")
}

func ValidateMonitorSpec(monitorSpec *tapi.MonitorSpec) error {
	specData, err := json.Marshal(monitorSpec)
	if err != nil {
		return err
	}

	if monitorSpec.Agent == "" {
		return fmt.Errorf(`Object 'Agent' is missing in '%v'`, string(specData))
	}
	if monitorSpec.Prometheus != nil {
		if monitorSpec.Agent != tapi.AgentCoreosPrometheus {
			return fmt.Errorf(`Invalid 'Agent' in '%v'`, string(specData))
		}
	}

	return nil
}

func ValidateSnapshot(client clientset.Interface, snapshot *tapi.Snapshot) error {
	snapshotSpec := snapshot.Spec.SnapshotStorageSpec
	if err := ValidateSnapshotSpec(snapshotSpec); err != nil {
		return err
	}

	if err := CheckBucketAccess(client, snapshot.Spec.SnapshotStorageSpec, snapshot.Namespace); err != nil {
		return err
	}
	return nil
}
