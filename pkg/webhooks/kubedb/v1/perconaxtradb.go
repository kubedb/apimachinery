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

// SetupPerconaXtraDBWebhookWithManager registers the webhook for PerconaXtraDB in the manager.
func SetupPerconaXtraDBWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&dbapi.PerconaXtraDB{}).
		WithValidator(&PerconaXtraDBCustomWebhook{mgr.GetClient()}).
		WithDefaulter(&PerconaXtraDBCustomWebhook{mgr.GetClient()}).
		Complete()
}

type PerconaXtraDBCustomWebhook struct {
	DefaultClient client.Client
}

var _ webhook.CustomDefaulter = &PerconaXtraDBCustomWebhook{}

func (w PerconaXtraDBCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	log := logf.FromContext(ctx)
	db := obj.(*dbapi.PerconaXtraDB)
	log.Info("defaulting PerconaXtraDB")
	if db.Spec.Version == "" {
		return errors.New(`'spec.version' is missing`)
	}

	if db.Spec.Halted {
		if db.Spec.DeletionPolicy == dbapi.DeletionPolicyDoNotTerminate {
			return errors.New(`Can't halt, since termination policy is 'DoNotTerminate'`)
		}
		db.Spec.DeletionPolicy = dbapi.DeletionPolicyHalt
	}

	var pxVersion catalogapi.PerconaXtraDBVersion
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name: db.Spec.Version,
	}, &pxVersion)
	if err != nil {
		return errors.Wrapf(err, "failed to get PerconaXtraDBVersion: %s", db.Spec.Version)
	}
	db.SetDefaults(&pxVersion)

	return nil
}

var _ webhook.CustomValidator = &PerconaXtraDBCustomWebhook{}

func (w PerconaXtraDBCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (warnings admission.Warnings, err error) {
	return w.validate(ctx, obj.(*dbapi.PerconaXtraDB))
}

func (w PerconaXtraDBCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	oldPerconaXtraDB, ok := oldObj.(*dbapi.PerconaXtraDB)
	if !ok {
		return nil, fmt.Errorf("expected a PerconaXtraDB but got a %T", oldPerconaXtraDB)
	}
	perconaxtradb, ok := newObj.(*dbapi.PerconaXtraDB)
	if !ok {
		return nil, fmt.Errorf("expected a PerconaXtraDB but got a %T", perconaxtradb)
	}

	var perconaxtradbVersion catalogapi.PerconaXtraDBVersion
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name: oldPerconaXtraDB.Spec.Version,
	}, &perconaxtradbVersion)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get PerconaXtraDBVersion: %s", oldPerconaXtraDB.Spec.Version)
	}
	oldPerconaXtraDB.SetDefaults(&perconaxtradbVersion)
	if oldPerconaXtraDB.Spec.AuthSecret == nil {
		oldPerconaXtraDB.Spec.AuthSecret = perconaxtradb.Spec.AuthSecret
	}
	if err := validateXtraDBUpdate(perconaxtradb, oldPerconaXtraDB); err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return w.validate(ctx, perconaxtradb)
}

func (w PerconaXtraDBCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (warnings admission.Warnings, err error) {
	perconaxtradb, ok := obj.(*dbapi.PerconaXtraDB)
	if !ok {
		return nil, fmt.Errorf("expected a PerconaXtraDB but got a %T", obj)
	}

	var pg dbapi.PerconaXtraDB
	err = w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      perconaxtradb.Name,
		Namespace: perconaxtradb.Namespace,
	}, &pg)
	if err != nil && !kerr.IsNotFound(err) {
		return nil, errors.Wrapf(err, "failed to get PerconaXtraDB: %s", perconaxtradb.Name)
	} else if err == nil && pg.Spec.DeletionPolicy == dbapi.DeletionPolicyDoNotTerminate {
		return nil, fmt.Errorf(`perconaxtradb "%v/%v" can't be terminated. To delete, change spec.terminationPolicy`, pg.Namespace, pg.Name)
	}
	return nil, nil
}

// ValidatePerconaXtraDB checks if the object satisfies all the requirements.
// It is not method of Interface, because it is referenced from controller package too.
func ValidatePerconaXtraDB(kc client.Client, db *dbapi.PerconaXtraDB, strictValidation bool) error {
	if db.Spec.Version == "" {
		return errors.New(`'spec.version' is missing`)
	}

	if db.Spec.Replicas == nil {
		return fmt.Errorf(`'spec.replicas' is missing`)
	}

	var pxVersion catalogapi.PerconaXtraDBVersion
	err := kc.Get(context.TODO(), types.NamespacedName{Name: db.Spec.Version}, &pxVersion)
	if err != nil {
		return err
	}

	if *db.Spec.Replicas < kubedb.PerconaXtraDBDefaultClusterSize {
		return fmt.Errorf(`'spec.replicas' "%v" invalid. Value must be atleast %d for xtradb cluster`,
			db.Spec.Replicas, kubedb.PerconaXtraDBDefaultClusterSize)
	}

	if err := validateCluster(db); err != nil {
		return err
	}

	if err := validateXtraDBEnvsForAllContainers(db); err != nil {
		return err
	}

	if db.Spec.StorageType == "" {
		return fmt.Errorf(`'spec.storageType' is missing`)
	}
	if db.Spec.StorageType != dbapi.StorageTypeDurable && db.Spec.StorageType != dbapi.StorageTypeEphemeral {
		return fmt.Errorf(`'spec.storageType' %s is invalid`, db.Spec.StorageType)
	}
	if err := amv.ValidateStorage(kc, olddbapi.StorageType(db.Spec.StorageType), db.Spec.Storage); err != nil {
		return err
	}

	// if secret managed externally verify auth secret name is not empty
	if db.Spec.AuthSecret != nil && db.Spec.AuthSecret.ExternallyManaged && db.Spec.AuthSecret.Name == "" {
		return fmt.Errorf("for externallyManaged auth secret, user need to provide \"spec.authSecret.name\"")
	}
	if strictValidation {
		// Check if percona-xtradb Version is deprecated.
		// If deprecated, return error
		if pxVersion.Spec.Deprecated {
			return fmt.Errorf("percona-xtradb %s/%s is using deprecated version %v. Skipped processing", db.Namespace, db.Name, pxVersion.Name)
		}

		if err := pxVersion.ValidateSpecs(); err != nil {
			return fmt.Errorf("perconaXtraDBVersion %s/%s is using invalid perconaXtraDBVersion %v. Skipped processing. reason: %v", pxVersion.Namespace,
				pxVersion.Name, pxVersion.Name, err)
		}
	}

	if db.Spec.DeletionPolicy == "" {
		return fmt.Errorf(`'spec.terminationPolicy' is missing`)
	}

	if db.Spec.StorageType == dbapi.StorageTypeEphemeral && db.Spec.DeletionPolicy == dbapi.DeletionPolicyHalt {
		return fmt.Errorf(`'spec.terminationPolicy: Halt' can not be used for 'Ephemeral' storage`)
	}

	monitorSpec := db.Spec.Monitor
	if monitorSpec != nil {
		if err := amv.ValidateMonitorSpec(monitorSpec); err != nil {
			return err
		}
	}

	if err := amv.ValidateHealth(&db.Spec.HealthChecker); err != nil {
		return err
	}

	if err := validateXtraDBVolumes(db); err != nil {
		return err
	}

	if err := validateXtraDBVolumeMountsForAllContainers(db); err != nil {
		return err
	}

	return nil
}

func (w PerconaXtraDBCustomWebhook) validate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	log := logf.FromContext(ctx)
	perconaxtradb, ok := obj.(*dbapi.PerconaXtraDB)
	if !ok {
		return nil, fmt.Errorf("expected a PerconaXtraDB but got a %T", obj)
	}
	log.Info("Validating PerconaXtraDB", perconaxtradb.Namespace, "/", perconaxtradb.Name)
	if perconaxtradb.Spec.Version == "" {
		return nil, errors.New(`'spec.version' is missing`)
	}
	var perconaxtradbVersion catalogapi.PerconaXtraDBVersion
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name: perconaxtradb.Spec.Version,
	}, &perconaxtradbVersion)
	if err != nil {
		return nil, err
	}

	if perconaxtradb.Spec.Replicas == nil || ptr.Deref(perconaxtradb.Spec.Replicas, 0) < 1 {
		return nil, fmt.Errorf(`spec.replicas "%d" invalid. Value must be greater than zero`, ptr.Deref(perconaxtradb.Spec.Replicas, 0))
	}

	if err := validateXtraDBEnvsForAllContainers(perconaxtradb); err != nil {
		return nil, err
	}

	if perconaxtradb.Spec.StorageType == "" {
		return nil, fmt.Errorf(`'spec.storageType' is missing`)
	}
	if perconaxtradb.Spec.StorageType != dbapi.StorageTypeDurable && perconaxtradb.Spec.StorageType != dbapi.StorageTypeEphemeral {
		return nil, fmt.Errorf(`'spec.storageType' %s is invalid`, perconaxtradb.Spec.StorageType)
	}
	if err := amv.ValidateStorage(w.DefaultClient, olddbapi.StorageType(perconaxtradb.Spec.StorageType), perconaxtradb.Spec.Storage); err != nil {
		return nil, err
	}

	err = validateXtraDBVolumes(perconaxtradb)
	if err != nil {
		return nil, err
	}
	err = validateXtraDBEnvsForAllContainers(perconaxtradb)
	if err != nil {
		return nil, err
	}

	// if secret managed externally verify auth secret name is not empty

	if perconaxtradb.Spec.DeletionPolicy == "" {
		return nil, fmt.Errorf(`'spec.terminationPolicy' is missing`)
	}

	if perconaxtradb.Spec.StorageType == dbapi.StorageTypeEphemeral && perconaxtradb.Spec.DeletionPolicy == dbapi.DeletionPolicyHalt {
		return nil, fmt.Errorf(`'spec.terminationPolicy: Halt' can not be used for 'Ephemeral' storage`)
	}

	monitorSpec := perconaxtradb.Spec.Monitor
	if monitorSpec != nil {
		if err := amv.ValidateMonitorSpec(monitorSpec); err != nil {
			return nil, err
		}
	}

	if err = amv.ValidateHealth(&perconaxtradb.Spec.HealthChecker); err != nil {
		return nil, err
	}

	return nil, nil
}

// validateCluster checks whether the configurations for PerconaXtraDB Cluster are ok
func validateCluster(db *dbapi.PerconaXtraDB) error {
	clusterName := db.ClusterName()
	if len(clusterName) > kubedb.PerconaXtraDBMaxClusterNameLength {
		return errors.Errorf(`'spec.px.clusterName' "%s" shouldn't have more than %d characters'`,
			clusterName, kubedb.PerconaXtraDBMaxClusterNameLength)
	}
	if db.Spec.Init != nil && db.Spec.Init.Script != nil {
		return fmt.Errorf("`.spec.init.scriptSource` is not supported for cluster. For PerconaXtraDB cluster initialization see https://stash.run/docs/latest/addons/percona-xtradb/guides/5.7/clusterd/")
	}

	return nil
}

var forbiddenXtraDBEnvVars = []string{
	"MYSQL_ROOT_PASSWORD",
	"MYSQL_ALLOW_EMPTY_PASSWORD",
	"MYSQL_RANDOM_ROOT_PASSWORD",
	"MYSQL_ONETIME_PASSWORD",
}

func validateXtraDBEnvsForAllContainers(db *dbapi.PerconaXtraDB) error {
	var err error
	for _, container := range db.Spec.PodTemplate.Spec.Containers {
		if errC := amv.ValidateEnvVar(container.Env, forbiddenXtraDBEnvVars, dbapi.ResourceKindPerconaXtraDB); errC != nil {
			if err == nil {
				err = errC
			} else {
				err = errors.Wrap(err, errC.Error())
			}
		}
	}
	return err
}

var reservedXtraDBVolumes = []string{
	kubedb.PerconaXtraDBDataVolumeName,
	kubedb.PerconaXtraDBCustomConfigVolumeName,
	kubedb.PerconaXtraDBInitScriptVolumeName,
	kubedb.PerconaXtraDBRunScriptVolumeName,
}

func getXtraDBTLSReservedVolumes() []string {
	var volumes []string
	volumes = append(volumes, kubedb.PerconaXtraDBServerTLSVolumeName)
	volumes = append(volumes, kubedb.PerconaXtraDBClientTLSVolumeName)
	volumes = append(volumes, kubedb.PerconaXtraDBExporterTLSVolumeName)
	volumes = append(volumes, kubedb.PerconaXtraDBMetricsExporterTLSVolumeName)
	return volumes
}

func validateXtraDBVolumes(db *dbapi.PerconaXtraDB) error {
	return amv.ValidateVolumes(ofstv1.ConvertVolumes(db.Spec.PodTemplate.Spec.Volumes), append(reservedXtraDBVolumes, getXtraDBTLSReservedVolumes()...))
}

func validateXtraDBVolumeMountsForAllContainers(db *dbapi.PerconaXtraDB) error {
	var err error
	for _, container := range db.Spec.PodTemplate.Spec.Containers {
		if errC := amv.ValidateMountPaths(container.VolumeMounts, append(reservedXtraDBVolumeMounts, getXtraDBTLSReservedVolumeMounts(db)...)); errC != nil {
			if err == nil {
				err = errC
			} else {
				err = errors.Wrap(err, errC.Error())
			}
		}
	}
	return err
}

var reservedXtraDBVolumeMounts = []string{
	kubedb.PerconaXtraDBDataMountPath,
	kubedb.PerconaXtraDBClusterCustomConfigMountPath,
	kubedb.PerconaXtraDBInitScriptVolumeMountPath,
	kubedb.PerconaXtraDBRunScriptVolumeMountPath,
}

func getXtraDBTLSReservedVolumeMounts(db *dbapi.PerconaXtraDB) []string {
	var volumes []string
	volumes = append(volumes, db.CertMountPath(dbapi.PerconaXtraDBServerCert))
	volumes = append(volumes, db.CertMountPath(dbapi.PerconaXtraDBClientCert))
	volumes = append(volumes, db.CertMountPath(dbapi.PerconaXtraDBExporterCert))
	volumes = append(volumes, kubedb.PerconaXtraDBMetricsExporterConfigPath)
	return volumes
}

func validateXtraDBUpdate(obj, oldObj *dbapi.PerconaXtraDB) error {
	preconditions := meta_util.PreConditionSet{
		Set: sets.New[string](
			"spec.storageType",
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
