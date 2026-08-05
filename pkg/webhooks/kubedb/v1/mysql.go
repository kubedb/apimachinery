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

	catalogapi "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	"kubedb.dev/apimachinery/apis/kubedb"
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"
	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	amv "kubedb.dev/apimachinery/pkg/validator"

	"github.com/Masterminds/semver/v3"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/mergepatch"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/ptr"
	meta_util "kmodules.xyz/client-go/meta"
	ofstv1 "kmodules.xyz/offshoot-api/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// SetupMySQLWebhookWithManager registers the webhook for MySQL in the manager.
func SetupMySQLWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&dbapi.MySQL{}).
		WithValidator(&MySQLCustomWebhook{DefaultClient: mgr.GetClient()}).
		WithDefaulter(&MySQLCustomWebhook{DefaultClient: mgr.GetClient()}).
		Complete()
}

type MySQLCustomWebhook struct {
	DefaultClient    client.Client
	StrictValidation bool
}

var mysqlLog = logf.Log.WithName("mysql-resource")

var _ webhook.CustomDefaulter = &MySQLCustomWebhook{}

func (w MySQLCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	db := obj.(*dbapi.MySQL)
	mysqlLog.Info("defaulting", "name", db.GetName())

	if db.Spec.Version == "" {
		return errors.New(`'spec.version' is missing`)
	}

	if db.Spec.Halted {
		if db.Spec.DeletionPolicy == dbapi.DeletionPolicyDoNotTerminate {
			return errors.New(`Can't halt, since termination policy is 'DoNotTerminate'`)
		}
		db.Spec.DeletionPolicy = dbapi.DeletionPolicyHalt
	}

	var myVersion catalogapi.MySQLVersion
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name: db.Spec.Version,
	}, &myVersion)
	if err != nil {
		return err
	}

	if err = db.SetDefaults(&myVersion); err != nil {
		return err
	}

	return nil
}

var _ webhook.CustomValidator = &MySQLCustomWebhook{}

func (w MySQLCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (warnings admission.Warnings, err error) {
	mysql := obj.(*dbapi.MySQL)
	err = w.ValidateMySQL(mysql)
	mysqlLog.Info("validating", "name", mysql.Name)
	return admission.Warnings{}, err
}

func (w MySQLCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (warnings admission.Warnings, err error) {
	oldMySQL, ok := oldObj.(*dbapi.MySQL)
	if !ok {
		return nil, fmt.Errorf("expected a MySQL but got a %T", oldMySQL)
	}
	mysql, ok := newObj.(*dbapi.MySQL)
	if !ok {
		return nil, fmt.Errorf("expected a MySQL but got a %T", mysql)
	}

	var mysqlVersion catalogapi.MySQLVersion
	err = w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name: oldMySQL.Spec.Version,
	}, &mysqlVersion)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get mysqlversion: %s", oldMySQL.Spec.Version)
	}

	if err = oldMySQL.SetDefaults(&mysqlVersion); err != nil {
		return nil, err
	}

	if oldMySQL.Spec.AuthSecret == nil {
		oldMySQL.Spec.AuthSecret = mysql.Spec.AuthSecret
	}
	if err := validateUpdate(mysql, oldMySQL); err != nil {
		return nil, fmt.Errorf("%v", err)
	}

	err = w.ValidateMySQL(mysql)
	return nil, err
}

func (w MySQLCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (warnings admission.Warnings, err error) {
	mysql, ok := obj.(*dbapi.MySQL)
	if !ok {
		return nil, fmt.Errorf("expected a MySQL but got a %T", obj)
	}

	var my dbapi.MySQL
	err = w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      mysql.Name,
		Namespace: mysql.Namespace,
	}, &my)
	if err != nil && !kerr.IsNotFound(err) {
		return nil, errors.Wrapf(err, "failed to get MySQL: %s", mysql.Name)
	} else if err == nil && my.Spec.DeletionPolicy == dbapi.DeletionPolicyDoNotTerminate {
		return nil, fmt.Errorf(`mysql "%v/%v" can't be terminated. To delete, change spec.deletionPolicy`, my.Namespace, my.Name)
	}
	return nil, nil
}

var forbiddenMySQLEnvVars = []string{
	"MYSQL_ROOT_PASSWORD",
	"MYSQL_ALLOW_EMPTY_PASSWORD",
	"MYSQL_RANDOM_ROOT_PASSWORD",
	"MYSQL_ONETIME_PASSWORD",
}

func validateGroupReplicas(replicas int32) error {
	if replicas < 3 || replicas > kubedb.MySQLMaxGroupMembers {
		return fmt.Errorf("accepted value of 'spec.replicas' for group replication is in range [3, %d]",
			kubedb.MySQLMaxGroupMembers)
	}
	return nil
}

func validateMySQLGroup(replicas int32, version string, group dbapi.MySQLGroupSpec) error {
	if err := validateGroupReplicas(replicas); err != nil {
		return err
	}

	// validate group name whether it is a valid uuid
	if _, err := uuid.Parse(group.Name); err != nil {
		return errors.Wrapf(err, "invalid group name is set")
	}

	if group.Mode != nil && string(*group.Mode) == "Multi-Primary" {
		refVersion := semver.MustParse("8.4.2")
		curVersion := semver.MustParse(version)

		if curVersion.LessThan(refVersion) {
			return fmt.Errorf("mysql group support multi primary mode starting from version %s and above", refVersion)
		}
	}

	return nil
}

// ValidateMySQL checks if the object satisfies all the requirements.
// It is not method of Interface, because it is referenced from controller package too.
func (w MySQLCustomWebhook) ValidateMySQL(mysql *dbapi.MySQL) error {
	if mysql.Spec.Version == "" {
		return errors.New(`'spec.version' is missing`)
	}

	var mysqlVersion catalogapi.MySQLVersion
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: mysql.Spec.Version}, &mysqlVersion)
	if err != nil {
		return err
	}

	if mysql.Spec.Replicas == nil {
		return fmt.Errorf(`spec.replicas "%d" invalid. Value must be greater than 0, but for group replication this value shouldn't be more than %d'`,
			ptr.Deref(mysql.Spec.Replicas, 0), kubedb.MySQLMaxGroupMembers)
	}

	if mysql.Spec.Topology != nil {
		if mysql.Spec.Topology.Mode == nil {
			return errors.New("a valid 'spec.topology.mode' must be set for MySQL clustering")
		}

		if mysql.IsInnoDBCluster() && mysqlVersion.Spec.Router.Image == "" {
			return errors.Errorf("InnoDBCluster mode is not supported for MySQL version %s", mysqlVersion.Name)
		}

		// validation for group configuration is performed only when
		// 'spec.topology.mode' is set to "GroupReplication"
		if *mysql.Spec.Topology.Mode == dbapi.MySQLModeGroupReplication {
			// if spec.topology.mode is "GroupReplication", spec.topology.group is set to default during mutating
			if mysql.Spec.Init == nil || mysql.Spec.Init.Archiver == nil || mysql.Spec.Init.Initialized {
				if err := validateMySQLGroup(ptr.Deref(mysql.Spec.Replicas, 0), mysql.Spec.Version, *mysql.Spec.Topology.Group); err != nil {
					return err
				}
			}
		}
	}

	if err := validateEnvsForAllContainers(mysql); err != nil {
		return err
	}

	if mysql.Spec.StorageType == "" {
		return fmt.Errorf(`'spec.storageType' is missing`)
	}

	if mysql.Spec.StorageType != dbapi.StorageTypeDurable && mysql.Spec.StorageType != dbapi.StorageTypeEphemeral {
		return fmt.Errorf(`'spec.storageType' %s is invalid`, mysql.Spec.StorageType)
	}

	if err := amv.ValidateStorage(w.DefaultClient, olddbapi.StorageType(mysql.Spec.StorageType), mysql.Spec.Storage); err != nil {
		return err
	}

	if w.StrictValidation {

		// Check if mysqlVersion is deprecated.
		// If deprecated, return error
		if mysqlVersion.Spec.Deprecated {
			return fmt.Errorf("mysql %s/%s is using deprecated version %v. Skipped processing", mysql.Namespace, mysql.Name, mysqlVersion.Name)
		}
		if err := mysqlVersion.ValidateSpecs(); err != nil {
			return fmt.Errorf("mysql %s/%s is using invalid mysqlVersion %v. Skipped processing. reason: %v", mysql.Namespace,
				mysql.Name, mysqlVersion.Name, err)
		}
	}

	// if secret managed externally verify auth secret name is not empty
	if mysql.Spec.AuthSecret != nil && mysql.Spec.AuthSecret.ExternallyManaged && mysql.Spec.AuthSecret.Name == "" {
		return fmt.Errorf("for externallyManaged auth secret, user need to provide \"mysql.Spec.AuthSecret.Name\"")
	}

	if mysql.Spec.DeletionPolicy == "" {
		return fmt.Errorf(`'spec.deletionPolicy' is missing`)
	}

	if mysql.Spec.StorageType == dbapi.StorageTypeEphemeral && mysql.Spec.DeletionPolicy == dbapi.DeletionPolicyHalt {
		return fmt.Errorf(`'spec.deletionPolicy: Halt' can not be used for 'Ephemeral' storage`)
	}

	monitorSpec := mysql.Spec.Monitor
	if monitorSpec != nil {
		if err := amv.ValidateMonitorSpec(monitorSpec); err != nil {
			return err
		}
	}

	if mysql.Spec.Version[0] == '5' {
		dnsName := mysql.Name + "." + mysql.Name + "-pods." + mysql.Namespace + ".svc"
		if len(dnsName) > 60 {
			return fmt.Errorf("MySQL 5.*.* does not support dns name longer than 60 characters. 'name.name-pods.namespace.svc' should not exceed 60 characters")
		}
	}

	if err := amv.ValidateHealth(&mysql.Spec.HealthChecker); err != nil {
		return err
	}

	if err := validateVolumes(mysql); err != nil {
		return err
	}

	if err := validateVolumeMountsForAllContainers(mysql); err != nil {
		return err
	}

	return nil
}

func validateEnvsForAllContainers(mysql *dbapi.MySQL) error {
	var err error
	for _, container := range mysql.Spec.PodTemplate.Spec.Containers {
		if errC := amv.ValidateEnvVar(container.Env, forbiddenMySQLEnvVars, dbapi.ResourceKindMySQL); errC != nil {
			if err == nil {
				err = errC
			} else {
				err = errors.Wrap(err, errC.Error())
			}
		}
	}
	return err
}

func validateVolumes(db *dbapi.MySQL) error {
	if db.Spec.PodTemplate.Spec.Volumes == nil {
		return nil
	}
	rsv := mySqlReservedVolumes
	if db.Spec.TLS != nil && db.Spec.TLS.Certificates != nil {
		for _, c := range db.Spec.TLS.Certificates {
			rsv = append(rsv, db.CertificateName(dbapi.MySQLCertificateAlias(c.Alias)))
		}
	}

	return amv.ValidateVolumes(ofstv1.ConvertVolumes(db.Spec.PodTemplate.Spec.Volumes), rsv)
}

func validateVolumeMountsForAllContainers(mysql *dbapi.MySQL) error {
	var err error
	for _, container := range mysql.Spec.PodTemplate.Spec.Containers {
		if errC := amv.ValidateMountPaths(container.VolumeMounts, mySqlReservedVolumeMountPaths); errC != nil {
			if err == nil {
				err = errC
			} else {
				err = errors.Wrap(err, errC.Error())
			}
		}
	}
	return err
}

// reserved volume and volumes mounts for mysql
var mySqlReservedVolumes = []string{
	kubedb.MySQLVolumeNameTemp,
	kubedb.MySQLVolumeNameData,
	kubedb.MySQLVolumeNameInitScript,
	kubedb.MySQLVolumeNameUserInitScript,
	kubedb.MySQLVolumeNameTLS,
	kubedb.MySQLVolumeNameExporterTLS,
	kubedb.MySQLVolumeNameCustomConfig,
	kubedb.MySQLVolumeNameSourceCA,
}

var mySqlReservedVolumeMountPaths = []string{
	kubedb.MySQLVolumeMountPathTemp,
	kubedb.MySQLVolumeMountPathData,
	kubedb.MySQLVolumeMountPathInitScript,
	kubedb.MySQLVolumeMountPathUserInitScript,
	kubedb.MySQLVolumeMountPathTLS,
	kubedb.MySQLVolumeMountPathExporterTLS,
	kubedb.MySQLVolumeMountPathCustomConfig,
	kubedb.MySQLVolumeMountPathSourceCA,
}

func validateUpdate(obj, oldObj *dbapi.MySQL) error {
	preconditions := meta_util.PreConditionSet{
		Set: sets.New[string](
			"spec.storageType",
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
