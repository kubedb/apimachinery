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

	archiverapi "kubedb.dev/apimachinery/apis/archiver/v1alpha1"
	catalogapi "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	"kubedb.dev/apimachinery/apis/kubedb"
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"
	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/apimachinery/pkg/double_optin"
	amv "kubedb.dev/apimachinery/pkg/validator"

	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/mergepatch"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/ptr"
	kmapi "kmodules.xyz/client-go/api/v1"
	meta_util "kmodules.xyz/client-go/meta"
	ofstv1 "kmodules.xyz/offshoot-api/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// SetupMariaDBWebhookWithManager registers the webhook for MariaDB in the manager.
func SetupMariaDBWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&dbapi.MariaDB{}).
		WithValidator(&MariaDBCustomWebhook{mgr.GetClient()}).
		WithDefaulter(&MariaDBCustomWebhook{mgr.GetClient()}).
		Complete()
}

type MariaDBCustomWebhook struct {
	DefaultClient client.Client
}

var (
	_          webhook.CustomDefaulter = &MariaDBCustomWebhook{}
	mariadbLog                         = logf.Log.WithName("mariadb-resource")
)

// setDefaultValues provides the defaulting that is performed in mutating stage of creating/updating a MySQL database
func (a *MariaDBCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	db := obj.(*dbapi.MariaDB)
	mariadbLog.Info("defaulting", "name", db.GetName())
	if db.Spec.Version == "" {
		return errors.New(`'spec.version' is missing`)
	}

	if db.Spec.WsrepSSTMethod == "" && db.IsCluster() {
		db.Spec.WsrepSSTMethod = dbapi.GaleraWsrepSSTMethodRsync
	}
	if db.Spec.Halted {
		if db.Spec.DeletionPolicy == dbapi.DeletionPolicyDoNotTerminate {
			return errors.New(`Can't halt, since termination policy is 'DoNotTerminate'`)
		}
		db.Spec.DeletionPolicy = dbapi.DeletionPolicyHalt
	}

	var mdVersion catalogapi.MariaDBVersion
	err := a.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name: db.Spec.Version,
	}, &mdVersion)
	if err != nil {
		return errors.Wrapf(err, "failed to get MariaDBVersion: %s", db.Spec.Version)
	}
	db.SetDefaults(&mdVersion)

	archiverList := &archiverapi.MariaDBArchiverList{}
	err = a.DefaultClient.List(context.TODO(), archiverList)
	if err != nil {
		return err
	}

	for _, archiver := range archiverList.Items {
		var archiverNs core.Namespace
		err = a.DefaultClient.Get(context.TODO(), types.NamespacedName{
			Name: archiver.Namespace,
		}, &archiverNs)
		if err != nil {
			return err
		}

		var dbNs core.Namespace
		err = a.DefaultClient.Get(context.TODO(), types.NamespacedName{
			Name: db.Namespace,
		}, &dbNs)
		if err != nil {
			return err
		}

		possible, err := double_optin.CheckIfDoubleOptInPossible(db.ObjectMeta, archiverNs.ObjectMeta, dbNs.ObjectMeta, archiver.Spec.Databases)
		if err != nil {
			return err
		}
		if possible {
			db.Spec.Archiver = &dbapi.Archiver{
				Ref: kmapi.ObjectReference{
					Namespace: archiver.Namespace,
					Name:      archiver.Name,
				},
			}
			break
		}
	}

	return nil
}

var _ webhook.CustomValidator = &MariaDBCustomWebhook{}

func (mv MariaDBCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	mariadb := obj.(*dbapi.MariaDB)
	mariadbLog.Info("validating", "name", mariadb.Name)

	return mv.validateMariaDB(ctx, obj)

}

func (mv MariaDBCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	oldMariaDB, ok := oldObj.(*dbapi.MariaDB)
	if !ok {
		return nil, fmt.Errorf("expected a MariaDB but got a %T", oldMariaDB)
	}
	mariadb, ok := newObj.(*dbapi.MariaDB)
	if !ok {
		return nil, fmt.Errorf("expected a MariaDB but got a %T", mariadb)
	}

	var mariadbVersion catalogapi.MariaDBVersion
	err := mv.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name: oldMariaDB.Spec.Version,
	}, &mariadbVersion)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get MariaDBVersion: %s", oldMariaDB.Spec.Version)
	}
	oldMariaDB.SetDefaults(&mariadbVersion)
	if oldMariaDB.Spec.AuthSecret == nil {
		oldMariaDB.Spec.AuthSecret = mariadb.Spec.AuthSecret
	}
	if err := mv.validateUpdate(mariadb, oldMariaDB); err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return mv.validateMariaDB(ctx, mariadb)
}

func (mv MariaDBCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	mariadb, ok := obj.(*dbapi.MariaDB)
	if !ok {
		return nil, fmt.Errorf("expected a MariaDB but got a %T", obj)
	}

	var pg dbapi.MariaDB
	err := mv.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      mariadb.Name,
		Namespace: mariadb.Namespace,
	}, &pg)
	if err != nil && !kerr.IsNotFound(err) {
		return nil, errors.Wrapf(err, "failed to get MariaDB: %s", mariadb.Name)
	} else if err == nil && pg.Spec.DeletionPolicy == dbapi.DeletionPolicyDoNotTerminate {
		return nil, fmt.Errorf(`mariadb "%v/%v" can't be terminated. To delete, change spec.deletionPolicy`, pg.Namespace, pg.Name)
	}
	return nil, nil
}

var mariaDBforbiddenEnvVars = []string{
	"MYSQL_ROOT_PASSWORD",
	"MYSQL_ALLOW_EMPTY_PASSWORD",
	"MYSQL_RANDOM_ROOT_PASSWORD",
	"MYSQL_ONETIME_PASSWORD",
}

// validateCluster checks whether the configurations for MariaDB Cluster are ok
func (mv MariaDBCustomWebhook) validateCluster(db *dbapi.MariaDB) error {
	if db.IsCluster() {
		clusterName := db.ClusterName()
		if len(clusterName) > kubedb.MariaDBMaxClusterNameLength {
			return errors.Errorf(`'spec.md.clusterName' "%s" shouldn't have more than %d characters'`,
				clusterName, kubedb.MariaDBMaxClusterNameLength)
		}
		if db.Spec.Init != nil && db.Spec.Init.Script != nil {
			return fmt.Errorf("`.spec.init.scriptSource` is not supported for cluster. For MariaDB cluster initialization see https://stash.run/docs/latest/addons/mariadb/guides/5.7/clusterd/")
		}
	}

	return nil
}

func validateMariaDBReplicationSpec(db *dbapi.MariaDB) error {
	if db.Spec.Topology.MaxScale.Replicas == nil || ptr.Deref(db.Spec.Replicas, 0) < 1 {
		return fmt.Errorf(`spec.topology.maxscale.replicas "%d" invalid. Value must be greater than zero`, ptr.Deref(db.Spec.Topology.MaxScale.Replicas, 0))
	}
	if db.Spec.Topology.MaxScale.StorageType != dbapi.StorageTypeDurable && db.Spec.Topology.MaxScale.StorageType != dbapi.StorageTypeEphemeral {
		return fmt.Errorf(`'mariadb.Spec.Topology.MaxScale.storageType' %s is invalid`, db.Spec.Topology.MaxScale.StorageType)
	}
	return nil
}

func (mv MariaDBCustomWebhook) validateEnvsForAllContainers(db *dbapi.MariaDB) error {
	var err error
	for _, container := range db.Spec.PodTemplate.Spec.Containers {
		if errC := amv.ValidateEnvVar(container.Env, mariaDBforbiddenEnvVars, dbapi.ResourceKindMariaDB); errC != nil {
			if err == nil {
				err = errC
			} else {
				err = errors.Wrap(err, errC.Error())
			}
		}
	}
	return err
}

func (mv MariaDBCustomWebhook) validateUpdate(obj, oldObj *dbapi.MariaDB) error {
	preconditions := meta_util.PreConditionSet{
		Set: sets.New[string](
			"spec.storageType",
			//"spec.authSecret",
			"spec.podTemplate.spec.nodeSelector",
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

var mariaDBreservedVolumes = []string{
	kubedb.MariaDBDataVolumeName,
	kubedb.MariaDBCustomConfigVolumeName,
	kubedb.MariaDBInitScriptVolumeName,
	kubedb.MariaDBRunScriptVolumeName,
	kubedb.MariaDBInitDBVolumeName,
}

func getTLSReservedVolumes() []string {
	var volumes []string
	volumes = append(volumes, kubedb.MariaDBServerTLSVolumeName)
	volumes = append(volumes, kubedb.MariaDBClientTLSVolumeName)
	volumes = append(volumes, kubedb.MariaDBExporterTLSVolumeName)
	volumes = append(volumes, kubedb.MariaDBMetricsExporterTLSVolumeName)
	return volumes
}

func (mv MariaDBCustomWebhook) validateVolumes(db *dbapi.MariaDB) error {
	return amv.ValidateVolumes(ofstv1.ConvertVolumes(db.Spec.PodTemplate.Spec.Volumes), append(mariaDBreservedVolumes, getTLSReservedVolumes()...))
}

var reservedVolumeMounts = []string{
	kubedb.MariaDBDataMountPath,
	kubedb.MariaDBCustomConfigMountPath,
	kubedb.MariaDBClusterCustomConfigMountPath,
	kubedb.MariaDBInitScriptVolumeMountPath,
	kubedb.MariaDBRunScriptVolumeMountPath,
	kubedb.MariaDBInitDBMountPath,
}

func getTLSReservedVolumeMounts(db *dbapi.MariaDB) []string {
	var volumes []string
	volumes = append(volumes, db.CertMountPath(dbapi.MariaDBServerCert))
	volumes = append(volumes, db.CertMountPath(dbapi.MariaDBClientCert))
	volumes = append(volumes, db.CertMountPath(dbapi.MariaDBExporterCert))
	volumes = append(volumes, kubedb.MariaDBMetricsExporterConfigPath)
	return volumes
}

func (mv MariaDBCustomWebhook) validateVolumeMounts(db *dbapi.MariaDB) error {
	return mv.validateVolumeMountsForAllContainers(db)
}

func (mv MariaDBCustomWebhook) validateVolumeMountsForAllContainers(db *dbapi.MariaDB) error {
	var err error
	for _, container := range db.Spec.PodTemplate.Spec.Containers {
		if errC := amv.ValidateMountPaths(container.VolumeMounts, append(reservedVolumeMounts, getTLSReservedVolumeMounts(db)...)); errC != nil {
			if err == nil {
				err = errC
			} else {
				err = errors.Wrap(err, errC.Error())
			}
		}
	}
	return err
}

func validateWsrepSSTMethod(db *dbapi.MariaDB) error {
	if !db.IsCluster() {
		return nil
	}
	if db.Spec.WsrepSSTMethod != dbapi.GaleraWsrepSSTMethodRsync && db.Spec.WsrepSSTMethod != dbapi.GaleraWsrepSSTMethodMariabackup {
		return errors.Errorf("wsrepSSTMethod %s not valid. Expected values: %s, %s.", db.Spec.WsrepSSTMethod, dbapi.GaleraWsrepSSTMethodRsync, dbapi.GaleraWsrepSSTMethodMariabackup)
	}
	return nil
}

func (mv MariaDBCustomWebhook) validateMariaDB(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	log := logf.FromContext(ctx)
	mariadb, ok := obj.(*dbapi.MariaDB)
	if !ok {
		return nil, fmt.Errorf("expected a MariaDB but got a %T", obj)
	}
	log.Info("Validating MariaDB", mariadb.Namespace, "/", mariadb.Name)
	if mariadb.Spec.Version == "" {
		return nil, errors.New(`'spec.version' is missing`)
	}
	var mariadbVersion catalogapi.MariaDBVersion
	err := mv.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name: mariadb.Spec.Version,
	}, &mariadbVersion)
	if err != nil {
		return nil, err
	}

	if mariadb.Spec.Replicas == nil || ptr.Deref(mariadb.Spec.Replicas, 0) < 1 {
		return nil, fmt.Errorf(`spec.replicas "%d" invalid. Value must be greater than zero`, ptr.Deref(mariadb.Spec.Replicas, 0))
	}

	if err := mv.validateEnvsForAllContainers(mariadb); err != nil {
		return nil, err
	}

	if mariadb.Spec.StorageType == "" {
		return nil, fmt.Errorf(`'spec.storageType' is missing`)
	}
	if mariadb.Spec.StorageType != dbapi.StorageTypeDurable && mariadb.Spec.StorageType != dbapi.StorageTypeEphemeral {
		return nil, fmt.Errorf(`'spec.storageType' %s is invalid`, mariadb.Spec.StorageType)
	}
	if err := amv.ValidateStorage(mv.DefaultClient, olddbapi.StorageType(mariadb.Spec.StorageType), mariadb.Spec.Storage); err != nil {
		return nil, err
	}

	err = mv.validateVolumes(mariadb)
	if err != nil {
		return nil, err
	}
	err = mv.validateVolumeMountsForAllContainers(mariadb)
	if err != nil {
		return nil, err
	}

	// if secret managed externally verify auth secret name is not empty

	if mariadb.Spec.DeletionPolicy == "" {
		return nil, fmt.Errorf(`'spec.deletionPolicy' is missing`)
	}

	if mariadb.Spec.StorageType == dbapi.StorageTypeEphemeral && mariadb.Spec.DeletionPolicy == dbapi.DeletionPolicyHalt {
		return nil, fmt.Errorf(`'spec.deletionPolicy: Halt' can not be used for 'Ephemeral' storage`)
	}

	monitorSpec := mariadb.Spec.Monitor
	if monitorSpec != nil {
		if err := amv.ValidateMonitorSpec(monitorSpec); err != nil {
			return nil, err
		}
	}

	if err = amv.ValidateHealth(&mariadb.Spec.HealthChecker); err != nil {
		return nil, err
	}
	if mariadb.IsMariaDBReplication() {
		if err := validateMariaDBReplicationSpec(mariadb); err != nil {
			return nil, err
		}
	}
	return nil, nil
}
