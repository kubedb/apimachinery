/*
Copyright AppsCode Inc. and Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"context"
	"fmt"
	"math"
	"reflect"

	catalogapi "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	"kubedb.dev/apimachinery/apis/kubedb"
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"
	"kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	amv "kubedb.dev/apimachinery/pkg/validator"

	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/mergepatch"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
	meta_util "kmodules.xyz/client-go/meta"
	ofst "kmodules.xyz/offshoot-api/api/v1"
	ofstv2 "kmodules.xyz/offshoot-api/api/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/yaml"
)

// SetupMongoDBWebhookWithManager registers the webhook for MongoDB in the manager.
func SetupMongoDBWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&dbapi.MongoDB{}).
		WithValidator(&MongoDBCustomWebhook{DefaultClient: mgr.GetClient()}).
		WithDefaulter(&MongoDBCustomWebhook{DefaultClient: mgr.GetClient()}).
		Complete()
}

type MongoDBCustomWebhook struct {
	DefaultClient    client.Client
	StrictValidation bool
}

var _ webhook.CustomDefaulter = &MongoDBCustomWebhook{}

func (w MongoDBCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	log := logf.FromContext(ctx)
	log.Info("defaulting MongoDB")
	db := obj.(*dbapi.MongoDB)

	if db.Spec.Version == "" {
		return errors.New(`'spec.version' is missing`)
	}

	if db.Spec.Halted {
		if db.Spec.DeletionPolicy == dbapi.DeletionPolicyDoNotTerminate {
			return errors.New(`Can't halt, since deletion policy is 'DoNotTerminate'`)
		}
		db.Spec.DeletionPolicy = dbapi.DeletionPolicyHalt
	}

	var mongodbVersion catalogapi.MongoDBVersion
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name: db.Spec.Version,
	}, &mongodbVersion)
	if err != nil {
		return errors.Wrapf(err, "failed to get MongoDBVersion: %s", db.Spec.Version)
	}

	db.SetDefaults(&mongodbVersion)

	return nil
}

var _ webhook.CustomValidator = &MongoDBCustomWebhook{}

func (w MongoDBCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (warnings admission.Warnings, err error) {
	log := logf.FromContext(ctx)
	log.Info("creating MongoDB")
	return nil, w.ValidateMongoDB(obj.(*dbapi.MongoDB))
}

func (w MongoDBCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (warnings admission.Warnings, err error) {
	log := logf.FromContext(ctx)
	log.Info("updating MongoDB")
	oldMongoDB, ok := oldObj.(*dbapi.MongoDB)
	if !ok {
		return nil, fmt.Errorf("expected a MongoDB but got a %T", oldMongoDB)
	}
	mongodb, ok := newObj.(*dbapi.MongoDB)
	if !ok {
		return nil, fmt.Errorf("expected a MongoDB but got a %T", mongodb)
	}

	var mongodbVersion catalogapi.MongoDBVersion
	err = w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name: oldMongoDB.Spec.Version,
	}, &mongodbVersion)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get MongodbVersion: %s", oldMongoDB.Spec.Version)
	}
	oldMongoDB.SetDefaults(&mongodbVersion)
	if oldMongoDB.Spec.AuthSecret == nil {
		oldMongoDB.Spec.AuthSecret = mongodb.Spec.AuthSecret
	}
	if err := w.validateUpdate(mongodb, oldMongoDB); err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return nil, w.ValidateMongoDB(mongodb)
}

func (w MongoDBCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (warnings admission.Warnings, err error) {
	log := logf.FromContext(ctx)
	log.Info("deleting MongoDB")
	mongodb, ok := obj.(*dbapi.MongoDB)
	if !ok {
		return nil, fmt.Errorf("expected a Mongodb but got a %T", obj)
	}

	var mg dbapi.MongoDB
	err = w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      mongodb.Name,
		Namespace: mongodb.Namespace,
	}, &mg)
	if err != nil && !kerr.IsNotFound(err) {
		return nil, errors.Wrapf(err, "failed to get Mongodb: %s", mongodb.Name)
	} else if err == nil && mg.Spec.DeletionPolicy == dbapi.DeletionPolicyDoNotTerminate {
		return nil, fmt.Errorf(`mongodb "%v/%v" can't be terminated. To delete, change spec.deletionPolicy`, mg.Namespace, mg.Name)
	}
	return nil, nil
}

var forbiddenMongoDBEnvVars = []string{
	"MONGO_INITDB_ROOT_USERNAME",
	"MONGO_INITDB_ROOT_PASSWORD",
}

// ValidateMongoDB checks if the object satisfies all the requirements.
// It is not method of Interface, because it is referenced from controller package too.
func (w MongoDBCustomWebhook) ValidateMongoDB(db *dbapi.MongoDB) error {
	if err := checkVersion(db, w.DefaultClient); err != nil {
		return err
	}
	if err := checkInvalidFieldsAndReplicaCounts(db); err != nil {
		return err
	}
	if err := checkEnvs(db); err != nil {
		return err
	}

	if err := checkStorageStuffs(w.DefaultClient, db); err != nil {
		return err
	}
	if err := checkSSLStuffs(db); err != nil {
		return err
	}

	if db.Spec.AuthSecret != nil && db.Spec.AuthSecret.ExternallyManaged && db.Spec.AuthSecret.Name == "" {
		return fmt.Errorf(`for externallyManaged auth secret, user must configure "spec.authSecret.name"`)
	}
	if err := strictValidations(w.DefaultClient, db, w.StrictValidation); err != nil {
		return err
	}

	// Deletion Policy
	if db.Spec.DeletionPolicy == "" {
		return fmt.Errorf(`'spec.deletionPolicy' is missing`)
	}

	if db.Spec.StorageType == dbapi.StorageTypeEphemeral && db.Spec.DeletionPolicy == dbapi.DeletionPolicyHalt {
		return fmt.Errorf(`'spec.deletionPolicy: Halt' can not be used for 'Ephemeral' storage`)
	}

	// Monitoring
	monitorSpec := db.Spec.Monitor
	if monitorSpec != nil {
		if err := amv.ValidateMonitorSpec(monitorSpec); err != nil {
			return err
		}
	}

	if err := validateVolumesAndMounts(db); err != nil {
		return err
	}
	return amv.ValidateHealth(&db.Spec.HealthChecker)
}

func (w MongoDBCustomWebhook) validateUpdate(obj, oldObj *dbapi.MongoDB) error {
	preconditions := meta_util.PreConditionSet{
		Set: sets.New[string](
			"spec.storageType",
			"spec.replicaSet.name",
			"spec.shardTopology.*.prefix",
		),
	}

	// Once the database has been initialized, don't let update the "spec.init" section
	if oldObj.Spec.Init != nil && oldObj.Spec.Init.Initialized {
		preconditions.Insert("spec.init")
	}

	_, err := meta_util.CreateStrategicPatch(oldObj, obj, preconditions.PreconditionFunc()...)
	if err != nil {
		if mergepatch.IsPreconditionFailed(err) {
			return fmt.Errorf("%v.%v", err, preconditions.Error())
		}
		return err
	}
	return nil
}

func checkVersion(db *dbapi.MongoDB, client client.Client) error {
	if db.Spec.Version == "" {
		return errors.New(`'spec.version' is missing`)
	}
	var mongodbVersion catalogapi.MongoDBVersion
	err := client.Get(context.TODO(), types.NamespacedName{
		Name: db.Spec.Version,
	}, &mongodbVersion)
	if err != nil {
		return err
	}
	return err
}

func checkInvalidFieldsAndReplicaCounts(db *dbapi.MongoDB) error {
	top := db.Spec.ShardTopology
	if top != nil {
		if db.Spec.Replicas != nil {
			return fmt.Errorf(`doesn't support 'spec.replicas' when spec.shardTopology is set`)
		}
		if db.Spec.PodTemplate != nil {
			return fmt.Errorf(`doesn't support 'spec.podTemplate' when spec.shardTopology is set`)
		}
		if db.Spec.ConfigSecret != nil {
			return fmt.Errorf(`doesn't support 'spec.configSecret' when spec.shardTopology is set`)
		}

		// Validate Topology Replicas values
		if top.Shard.Shards < 1 {
			return fmt.Errorf(`spec.shardTopology.shard.shards %v invalid. Must be greater than zero when spec.shardTopology is set`, top.Shard.Shards)
		}
		if top.Shard.Replicas <= 0 {
			return fmt.Errorf(`spec.shardTopology.shard.replicas %v invalid. Must be greater than zero when spec.shardTopology is set`, top.Shard.Replicas)
		}
		if top.ConfigServer.Replicas <= 0 {
			return fmt.Errorf(`spec.shardTopology.configServer.replicas %v invalid. Must be greater than zero when spec.shardTopology is set`, top.ConfigServer.Replicas)
		}
		if top.Mongos.Replicas < 1 {
			return fmt.Errorf(`spec.shardTopology.mongos.replicas %v invalid. Must be greater than zero when spec.shardTopology is set`, top.Mongos.Replicas)
		}

		if db.Spec.StorageEngine == dbapi.StorageEngineInMemory {
			if top.Shard.Replicas < 3 {
				return fmt.Errorf(`spec.shardTopology.shard.replicas %v invalid. Must be 3 or more when storageEngine is set to inMemory`, top.Shard.Replicas)
			}
			if top.ConfigServer.Replicas < 3 {
				return fmt.Errorf(`spec.shardTopology.configServer.replicas %v invalid. Must be 3 or more when storageEngine is set to inMemory`, top.ConfigServer.Replicas)
			}
		}
	} else {
		if db.Spec.Replicas == nil || ptr.Deref(db.Spec.Replicas, 0) < 1 {
			return fmt.Errorf(`spec.replicas "%d" invalid. Must be greater than zero in non-shardTopology`, ptr.Deref(db.Spec.Replicas, 0))
		}

		if db.Spec.Replicas == nil || (db.Spec.ReplicaSet == nil && ptr.Deref(db.Spec.Replicas, 0) != 1) {
			return fmt.Errorf(`spec.replicas "%d" invalid for 'MongoDB Standalone' instance. Value must be one`, ptr.Deref(db.Spec.Replicas, 0))
		}

		if db.Spec.StorageEngine == dbapi.StorageEngineInMemory {
			if ptr.Deref(db.Spec.Replicas, 0) < 3 {
				return fmt.Errorf(`spec.replicas %d invalid. Must be 3 or more when storageEngine is set to inMemory`, ptr.Deref(db.Spec.Replicas, 0))
			}
		}
	}
	// arbiter & hidden not supported for standAlone db
	if db.Spec.ShardTopology == nil && db.Spec.ReplicaSet == nil {
		if db.Spec.Arbiter != nil {
			return fmt.Errorf(`spec.arbiter "%v" is invalid for Standalone MongoDB. `, ptr.Deref(db.Spec.Arbiter, dbapi.MongoArbiterNode{}))
		}
		if db.Spec.Hidden != nil {
			return fmt.Errorf(`spec.hidden "%v" is invalid for Standalone MongoDB. `, ptr.Deref(db.Spec.Hidden, dbapi.MongoHiddenNode{}))
		}
	}
	if db.Spec.Hidden != nil && db.Spec.Hidden.Replicas <= 0 {
		return fmt.Errorf("spec.Hidden.Replicas %d is invalid. Must be 1 or more", db.Spec.Hidden.Replicas)
	}
	return nil
}

func checkEnvs(db *dbapi.MongoDB) error {
	top := db.Spec.ShardTopology
	if top != nil {
		if err := validateEnvsForAllMongoDBContainers(top.Shard.PodTemplate); err != nil {
			return err
		}
		if err := validateEnvsForAllMongoDBContainers(top.ConfigServer.PodTemplate); err != nil {
			return err
		}
		if err := validateEnvsForAllMongoDBContainers(top.Mongos.PodTemplate); err != nil {
			return err
		}

	} else {
		if db.Spec.PodTemplate != nil {
			if err := validateEnvsForAllMongoDBContainers(db.Spec.PodTemplate); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateEnvsForAllMongoDBContainers(podTemplate *ofstv2.PodTemplateSpec) error {
	var err error
	for _, container := range podTemplate.Spec.Containers {
		if errC := amv.ValidateEnvVar(container.Env, forbiddenMongoDBEnvVars, dbapi.ResourceKindMongoDB); errC != nil {
			if err == nil {
				err = errC
			} else {
				err = errors.Wrap(err, errC.Error())
			}
		}
	}
	return err
}

func checkStorageStuffs(client client.Client, db *dbapi.MongoDB) error {
	if db.Spec.StorageType == "" {
		return fmt.Errorf(`'spec.storageType' is missing`)
	}
	top := db.Spec.ShardTopology
	storageType := v1alpha2.StorageType(db.Spec.StorageType)
	// Validate storage for ClusterTopology or non-ClusterTopology
	if top != nil {
		if db.Spec.Storage != nil {
			return fmt.Errorf("doesn't support 'spec.storage' when spec.shardTopology is set")
		}
		if err := amv.IsStorageTypeCompatibleWithSpec(storageType, top.Shard.Storage, top.Shard.EphemeralStorage); err != nil {
			return err
		}
		if err := amv.ValidateStorage(client, storageType, top.Shard.Storage, "spec.shardTopology.shard.storage"); err != nil {
			return err
		}
		if err := amv.IsStorageTypeCompatibleWithSpec(storageType, top.ConfigServer.Storage, top.ConfigServer.EphemeralStorage); err != nil {
			return err
		}
		if err := amv.ValidateStorage(client, storageType, top.ConfigServer.Storage, "spec.shardTopology.configServer.storage"); err != nil {
			return err
		}
	} else {
		if err := amv.IsStorageTypeCompatibleWithSpec(storageType, db.Spec.Storage, db.Spec.EphemeralStorage); err != nil {
			return err
		}

		if err := amv.ValidateStorage(client, storageType, db.Spec.Storage); err != nil {
			return err
		}
	}

	if db.Spec.Hidden != nil {
		// Hidden node doesn't have ephemeral storage, so always considering the type as `Durable`.
		if err := amv.ValidateStorage(client, v1alpha2.StorageType(dbapi.StorageTypeDurable), &db.Spec.Hidden.Storage, "spec.hidden.storage"); err != nil {
			return err
		}
	}
	return checkForInMemoryStorage(client, db)
}

func calcFromPodMemReq(memReq int64) float64 {
	// max(50% of (mem - 1GB), 256MB)
	return math.Max(float64(memReq-1<<30)*0.5, 256*(1<<20))
}

func getInMemoryStorage(client client.Client, dbNamespace string, name string, memReq int64) (float64, error) {
	convertFloat := func(v any) float64 {
		switch val := v.(type) {
		case int32:
			return float64(val)
		case float64:
			return val
		}
		klog.Errorf("failed to convert %v to float64", reflect.TypeOf(v))
		return -1
	}

	var sec core.Secret
	err := client.Get(context.TODO(), types.NamespacedName{
		Name:      name,
		Namespace: dbNamespace,
	}, &sec)
	if err != nil {
		return -1, err
	}

	conf := make(map[string]any)
	mongod, ok := sec.Data[kubedb.MongoDBCustomConfigFile]
	if !ok {
		return calcFromPodMemReq(memReq), nil
	}
	err = yaml.Unmarshal(mongod, conf)
	if err != nil {
		return -1, err
	}
	storage, ok := conf["storage"]
	if !ok {
		return calcFromPodMemReq(memReq), nil
	}
	inMem, ok := storage.(map[any]any)["inMemory"]
	if !ok {
		return calcFromPodMemReq(memReq), nil
	}
	engine, ok := inMem.(map[any]any)["engineConfig"]
	if !ok {
		return calcFromPodMemReq(memReq), nil
	}
	size, ok := engine.(map[any]any)["inMemorySizeGB"]
	if !ok {
		return calcFromPodMemReq(memReq), nil
	}
	// set by user
	return convertFloat(size) * (1 << 30), nil
}

func checkForInMemoryStorage(client client.Client, db *dbapi.MongoDB) error {
	if db.Spec.StorageEngine != dbapi.StorageEngineInMemory {
		return nil
	}

	checkLogic := func(inMem float64, memReq int64) error {
		if int64(inMem) > memReq {
			return errors.New("inMemorySizeGB has to be less or equal than the total Memory request")
		}
		return nil
	}
	checkOk := func(name string, memReq int64) error {
		inMem, err := getInMemoryStorage(client, db.Namespace, name, memReq)
		if err != nil {
			return err
		}
		return checkLogic(inMem, memReq)
	}

	top := db.Spec.ShardTopology
	if top != nil {
		for _, container := range top.Shard.PodTemplate.Spec.Containers {
			memReq := container.Resources.Requests.Memory().Value()
			if top.Shard.ConfigSecret != nil {
				err := checkOk(top.Shard.ConfigSecret.Name, memReq)
				if err != nil {
					return err
				}
			}
			if err := checkLogic(calcFromPodMemReq(memReq), memReq); err != nil {
				return err
			}
		}
	} else {
		// ReplicaSet or Standalone
		for _, container := range db.Spec.PodTemplate.Spec.Containers {
			memReq := container.Resources.Requests.Memory().Value()
			if db.Spec.ConfigSecret != nil {
				err := checkOk(db.Spec.ConfigSecret.Name, memReq)
				if err != nil {
					return err
				}
			}
			if err := checkLogic(calcFromPodMemReq(memReq), memReq); err != nil {
				return err
			}
		}
	}
	return nil
}

func checkSSLStuffs(db *dbapi.MongoDB) error {
	if (db.Spec.ClusterAuthMode == dbapi.ClusterAuthModeX509 || db.Spec.ClusterAuthMode == dbapi.ClusterAuthModeSendX509) &&
		(db.Spec.SSLMode == dbapi.SSLModeDisabled || db.Spec.SSLMode == dbapi.SSLModeAllowSSL) {
		return fmt.Errorf("can't have %v set to mongodb.spec.sslMode when mongodb.spec.clusterAuthMode is set to %v",
			db.Spec.SSLMode, db.Spec.ClusterAuthMode)
	}

	if db.Spec.ClusterAuthMode == dbapi.ClusterAuthModeSendKeyFile && db.Spec.SSLMode == dbapi.SSLModeDisabled {
		return fmt.Errorf("can't have %v set to mongodb.spec.sslMode when mongodb.spec.clusterAuthMode is set to %v",
			db.Spec.SSLMode, db.Spec.ClusterAuthMode)
	}
	return nil
}

func validateSecrets(client client.Client, db *dbapi.MongoDB) error {
	secrets := make([]string, 0)
	if db.Spec.KeyFileSecret != nil {
		secrets = append(secrets, db.Spec.KeyFileSecret.Name)
	}
	top := db.Spec.ShardTopology
	if top != nil {
		if top.Mongos.ConfigSecret != nil {
			secrets = append(secrets, top.Mongos.ConfigSecret.Name)
		}
		if top.Shard.ConfigSecret != nil {
			secrets = append(secrets, top.Shard.ConfigSecret.Name)
		}
		if top.ConfigServer.ConfigSecret != nil {
			secrets = append(secrets, top.ConfigServer.ConfigSecret.Name)
		}
	} else if db.Spec.ConfigSecret != nil {
		secrets = append(secrets, db.Spec.ConfigSecret.Name)
	}

	if db.Spec.Arbiter != nil && db.Spec.Arbiter.ConfigSecret != nil {
		secrets = append(secrets, db.Spec.Arbiter.ConfigSecret.Name)
	}
	if db.Spec.Hidden != nil && db.Spec.Hidden.ConfigSecret != nil {
		secrets = append(secrets, db.Spec.Hidden.ConfigSecret.Name)
	}
	return amv.CheckSecretsExist(client, secrets, db.Namespace)
}

func strictValidations(client client.Client, db *dbapi.MongoDB, on bool) error {
	if !on {
		return nil
	}
	if err := validateSecrets(client, db); err != nil {
		return err
	}

	// Check if mongodbVersion is deprecated.
	// If deprecated, return error
	var mongodbVersion catalogapi.MongoDBVersion
	err := client.Get(context.TODO(), types.NamespacedName{
		Name: db.Spec.Version,
	}, &mongodbVersion)
	if err != nil {
		return err
	}
	if mongodbVersion.Spec.Deprecated {
		return fmt.Errorf("mongoDB %s/%s is using deprecated version %v. Skipped processing",
			db.Namespace, db.Name, mongodbVersion.Name)
	}

	if err := mongodbVersion.ValidateSpecs(); err != nil {
		return fmt.Errorf("mongodb %s/%s is using invalid mongodbVersion %v. Skipped processing. reason: %v", db.Namespace,
			db.Name, mongodbVersion.Name, err)
	}
	return nil
}

var reservedMongoDBVolumes = []string{
	kubedb.MongoDBDataDirectoryName,
	kubedb.MongoDBConfigDirectoryName,
	kubedb.MongoDBInitScriptDirectoryName,
	dbapi.MongoCertDirectory,
	kubedb.MongoDBClientCertDirectoryName,
	kubedb.MongoDBServerCertDirectoryName,
	kubedb.MongoDBInitialKeyDirectoryName,
	kubedb.MongoDBInitialConfigDirectoryName,
	kubedb.MongoDBInitialDirectoryName,
	kubedb.MongoDBWorkDirectoryName,
}

var reservedMongoDBVolumeMounts = []string{
	kubedb.MongoDBDataDirectoryPath,
	kubedb.MongoDBConfigDirectoryPath,
	kubedb.MongoDBInitScriptDirectoryPath,
	kubedb.MongoDBCertDirectoryName,
	kubedb.MongoDBClientCertDirectoryPath,
	kubedb.MongoDBServerCertDirectoryPath,
	kubedb.MongoDBInitialKeyDirectoryPath,
	kubedb.MongoDBInitialConfigDirectoryPath,
	kubedb.MongoDBInitialDirectoryPath,
	kubedb.MongoDBWorkDirectoryPath,
}

func validateVolumesAndMounts(db *dbapi.MongoDB) error {
	volumes := make([]core.Volume, 0)
	mounts := make([]core.VolumeMount, 0)
	top := db.Spec.ShardTopology
	if top != nil {
		// shard
		volumes = append(volumes, ofst.ConvertVolumes(top.Shard.PodTemplate.Spec.Volumes)...)
		mounts = append(mounts, getVolumesMountsForAllContainers(top.Shard.PodTemplate)...)
		// configServer
		volumes = append(volumes, ofst.ConvertVolumes(top.ConfigServer.PodTemplate.Spec.Volumes)...)
		mounts = append(mounts, getVolumesMountsForAllContainers(top.ConfigServer.PodTemplate)...)
		// mongos
		volumes = append(volumes, ofst.ConvertVolumes(top.Mongos.PodTemplate.Spec.Volumes)...)
		mounts = append(mounts, getVolumesMountsForAllContainers(top.Mongos.PodTemplate)...)
	} else if db.Spec.PodTemplate != nil {
		volumes = append(volumes, ofst.ConvertVolumes(db.Spec.PodTemplate.Spec.Volumes)...)
		mounts = append(mounts, getVolumesMountsForAllContainers(db.Spec.PodTemplate)...)
	}
	if db.Spec.Arbiter != nil {
		volumes = append(volumes, ofst.ConvertVolumes(db.Spec.Arbiter.PodTemplate.Spec.Volumes)...)
		mounts = append(mounts, getVolumesMountsForAllContainers(db.Spec.Arbiter.PodTemplate)...)
	}
	if db.Spec.Hidden != nil {
		volumes = append(volumes, ofst.ConvertVolumes(db.Spec.Hidden.PodTemplate.Spec.Volumes)...)
		mounts = append(mounts, getVolumesMountsForAllContainers(db.Spec.Hidden.PodTemplate)...)
	}

	if err := amv.ValidateVolumes(volumes, reservedMongoDBVolumes); err != nil {
		return err
	}
	return amv.ValidateMountPaths(mounts, reservedMongoDBVolumeMounts)
}

func getVolumesMountsForAllContainers(podTemplate *ofstv2.PodTemplateSpec) []core.VolumeMount {
	mounts := make([]core.VolumeMount, 0)
	for _, container := range podTemplate.Spec.Containers {
		mounts = append(mounts, container.VolumeMounts...)
	}
	return mounts
}
