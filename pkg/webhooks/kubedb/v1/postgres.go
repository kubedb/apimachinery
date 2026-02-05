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
	"github.com/pkg/errors"
	"gomodules.xyz/pointer"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/mergepatch"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/ptr"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	ofst "kmodules.xyz/offshoot-api/api/v1"
	psapi "kubeops.dev/petset/apis/apps/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// SetupPostgresWebhookWithManager registers the webhook for Postgres in the manager.
func SetupPostgresWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&dbapi.Postgres{}).
		WithValidator(&PostgresCustomWebhook{DefaultClient: mgr.GetClient()}).
		WithDefaulter(&PostgresCustomWebhook{DefaultClient: mgr.GetClient()}).
		Complete()
}

type PostgresCustomWebhook struct {
	DefaultClient    client.Client
	StrictValidation bool
}

var pgLog = logf.Log.WithName("postgres-resource")

var _ webhook.CustomDefaulter = &PostgresCustomWebhook{}

func (wh *PostgresCustomWebhook) Default(_ context.Context, obj runtime.Object) error {
	db := obj.(*dbapi.Postgres)

	pgLog.Info("defaulting", "name", db.GetName())
	if db.Spec.Version == "" {
		return errors.New(`'spec.version' is missing`)
	}

	if db.Spec.Halted {
		if db.Spec.DeletionPolicy == dbapi.DeletionPolicyDoNotTerminate {
			return errors.New(`Can't halt, since termination policy is 'DoNotTerminate'`)
		}
		db.Spec.DeletionPolicy = dbapi.DeletionPolicyHalt
	}

	if db.Spec.Replicas == nil {
		db.Spec.Replicas = pointer.Int32P(1)
	}
	var postgresVersion catalogapi.PostgresVersion
	err := wh.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name: db.Spec.Version,
	}, &postgresVersion)
	if err != nil {
		return errors.Wrapf(err, "failed to get PostgresVersion: %s", db.Spec.Version)
	}

	db.SetDefaults(&postgresVersion)

	return nil
}

var _ webhook.CustomValidator = &PostgresCustomWebhook{}

var forbiddenPostgresEnvVars = []string{
	"POSTGRES_PASSWORD",
	"POSTGRES_USER",
}

var postgresReservedMountPaths = []string{
	kubedb.PostgresInitDir,
	kubedb.PostgresSharedMemoryDir,
	kubedb.PostgresDataDir,
	kubedb.PostgresCustomConfigDir,
	kubedb.PostgresRunScriptsDir,
	kubedb.PostgresRoleScriptsDir,
	kubedb.PostgresSharedScriptsDir,
	kubedb.PostgresSharedTlsVolumeMountPath,
}

var postgresReservedVolumes = []string{
	kubedb.PostgresInitVolumeName,
	kubedb.PostgresSharedMemoryVolumeName,
	kubedb.PostgresDataVolumeName,
	kubedb.PostgresCustomConfigVolumeName,
	kubedb.PostgresRunScriptsVolumeName,
	kubedb.PostgresRoleScriptsVolumeName,
	kubedb.PostgresSharedScriptsVolumeName,
	kubedb.PostgresSharedTlsVolumeName,
}

func (wh *PostgresCustomWebhook) validateEnvsForAllContainers(postgres *dbapi.Postgres) error {
	var err error
	for _, container := range postgres.Spec.PodTemplate.Spec.Containers {
		if errC := amv.ValidateEnvVar(container.Env, forbiddenPostgresEnvVars, dbapi.ResourceKindPostgres); errC != nil {
			if err == nil {
				err = errC
			} else {
				err = errors.Wrap(err, errC.Error())
			}
		}
	}
	return err
}

func (wh *PostgresCustomWebhook) validateVolumeMountsForAllContainers(postgres *dbapi.Postgres) error {
	var err error
	for _, container := range postgres.Spec.PodTemplate.Spec.Containers {
		if errC := amv.ValidateMountPaths(container.VolumeMounts, postgresReservedMountPaths); errC != nil {
			if err == nil {
				err = errC
			} else {
				err = errors.Wrap(err, errC.Error())
			}
		}
	}
	return err
}

func (wh *PostgresCustomWebhook) validateUpdate(obj, oldObj *dbapi.Postgres) error {
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

func (wh *PostgresCustomWebhook) validateSpecForDB(postgres *dbapi.Postgres, pgVersion *catalogapi.PostgresVersion) error {
	// need to set the UserID and GroupID
	container := core_util.GetContainerByName(postgres.Spec.PodTemplate.Spec.Containers, kubedb.PostgresContainerName)
	if container == nil {
		return fmt.Errorf("postgres container %s not found", kubedb.PostgresContainerName)
	}
	if pgVersion.Spec.SecurityContext.RunAsUser != nil &&
		container.SecurityContext != nil &&
		pointer.Int64(container.SecurityContext.RunAsUser) != pointer.Int64(pgVersion.Spec.SecurityContext.RunAsUser) &&
		!pgVersion.Spec.SecurityContext.RunAsAnyNonRoot {
		return fmt.Errorf("can't change ContainerSecurityContext's RunAsUser for this Postgres Version. It has to be the defualt UserID. The default UserID for this Postgres Version is %v but Container's security context UserID is %v", pointer.Int64(pgVersion.Spec.SecurityContext.RunAsUser), pointer.Int64(container.SecurityContext.RunAsUser))
	}
	if pgVersion.Spec.SecurityContext.RunAsUser != nil &&
		container.SecurityContext != nil &&
		pointer.Int64(container.SecurityContext.RunAsGroup) != pointer.Int64(pgVersion.Spec.SecurityContext.RunAsUser) &&
		!pgVersion.Spec.SecurityContext.RunAsAnyNonRoot {
		return fmt.Errorf("can't change ContainerSecurityContext's RunAsGroup for this Postgres Version. It has to be the defualt GroupID. The default GroupID for this Postgres Version is %v but Container's security context GroupID is %v", pointer.Int64(pgVersion.Spec.SecurityContext.RunAsUser), pointer.Int64(container.SecurityContext.RunAsGroup))
	}

	if (postgres.Spec.ClientAuthMode == dbapi.ClientAuthModeCert) &&
		(postgres.Spec.SSLMode == dbapi.PostgresSSLModeDisable) {
		return fmt.Errorf("can't have %v set to postgres.spec.sslMode when postgres.spec.ClientAuthMode is set to %v",
			postgres.Spec.SSLMode, postgres.Spec.ClientAuthMode)
	}
	if (postgres.Spec.TLS != nil) &&
		(postgres.Spec.SSLMode == dbapi.PostgresSSLModeDisable) {
		return fmt.Errorf("can't have %v set to postgres.spec.sslMode when postgres.spec.TLS is set ",
			postgres.Spec.SSLMode)
	}
	if (postgres.Spec.SSLMode != "" && postgres.Spec.SSLMode != dbapi.PostgresSSLModeDisable) && postgres.Spec.TLS == nil {
		return fmt.Errorf("can't have %v set to postgres.Spec.SSLMode when postgres.Spec.TLS is null",
			postgres.Spec.SSLMode)
	}
	if postgres.Spec.Replication != nil {
		version, _ := semver.NewVersion(pgVersion.Spec.Version)
		majorVersion := version.Major()
		if postgres.Spec.Replication.WALLimitPolicy != dbapi.WALKeepSegment && postgres.Spec.Replication.WALLimitPolicy != dbapi.WALKeepSize && postgres.Spec.Replication.WALLimitPolicy != dbapi.ReplicationSlot {
			return fmt.Errorf("can't have %v set to postgres.Spec.Replication.WALLimitPolicy, supported values are %s, %s, %s",
				postgres.Spec.Replication.WALLimitPolicy, dbapi.WALKeepSegment, dbapi.WALKeepSize, dbapi.ReplicationSlot)
		}
		if majorVersion <= uint64(12) && postgres.Spec.Replication.WALLimitPolicy != dbapi.WALKeepSegment {
			return fmt.Errorf("can't have %v set to postgres.Spec.Replication.WALLimitPolicy when major postgresversion is less than 13",
				postgres.Spec.Replication.WALLimitPolicy)
		}
		if majorVersion > uint64(12) && postgres.Spec.Replication.WALLimitPolicy == dbapi.WALKeepSegment {
			return fmt.Errorf("can't have %v set to postgres.Spec.Replication.WALLimitPolicy when major postgresversion is more than 12",
				postgres.Spec.Replication.WALLimitPolicy)
		}
		if majorVersion <= uint64(12) && postgres.Spec.Replication.WALLimitPolicy == dbapi.WALKeepSegment && ptr.Deref(postgres.Spec.Replication.WalKeepSegment, 0) < 0 {
			return fmt.Errorf("walKeepSegment can't be less than 0")
		}
		if majorVersion > uint64(12) && postgres.Spec.Replication.WALLimitPolicy == dbapi.WALKeepSize && ptr.Deref(postgres.Spec.Replication.WalKeepSizeInMegaBytes, 0) < 0 {
			return fmt.Errorf("walKeepSize can't be less than 0")
		}
		if majorVersion > uint64(12) && postgres.Spec.Replication.WALLimitPolicy == dbapi.ReplicationSlot && ptr.Deref(postgres.Spec.Replication.MaxSlotWALKeepSizeInMegaBytes, 0) < -1 {
			return fmt.Errorf("maxSlotWALKeepSize can't be less than -1")
		}
	}
	// validate leader election configs
	// ==============> start
	lec := postgres.Spec.LeaderElection
	if lec != nil {
		if lec.ElectionTick <= lec.HeartbeatTick {
			return fmt.Errorf("ElectionTick must be greater than HeartbeatTick")
		}
		if lec.ElectionTick < 1 {
			return fmt.Errorf("ElectionTick must be greater than zero")
		}
		if lec.HeartbeatTick < 1 {
			return fmt.Errorf("HeartbeatTick must be greater than zero")
		}
	}
	// end <==============
	return nil
}

func (wh *PostgresCustomWebhook) validateVolumes(db *dbapi.Postgres) error {
	if db.Spec.PodTemplate.Spec.Volumes == nil {
		return nil
	}
	rsv := postgresReservedVolumes
	if db.Spec.TLS != nil && db.Spec.TLS.Certificates != nil {
		for _, c := range db.Spec.TLS.Certificates {
			rsv = append(rsv, db.CertificateName(dbapi.PostgresCertificateAlias(c.Alias)))
		}
	}

	return amv.ValidateVolumes(ofst.ConvertVolumes(db.Spec.PodTemplate.Spec.Volumes), rsv)
}

func (wh *PostgresCustomWebhook) validateReadReplicas(postgres *dbapi.Postgres) error {
	if len(postgres.Spec.ReadReplicas) == 0 {
		return nil
	}
	for i := range postgres.Spec.ReadReplicas {
		rr := &postgres.Spec.ReadReplicas[i]
		if rr.Name == "" {
			return fmt.Errorf("readReplica name is missing for readReplica at index %d", i)
		}
		if rr.Name == kubedb.NodeTypeArbiter {
			return fmt.Errorf("readReplica name %q can't be used for readReplica at index %d, this name is reserved", kubedb.NodeTypeArbiter, i)
		}
		if rr.Replicas == nil || ptr.Deref(rr.Replicas, 0) < 1 {
			return fmt.Errorf(`spec.readReplicas[%d].replicas "%d" invalid. Value must be greater than zero`, i, ptr.Deref(rr.Replicas, 0))
		}
	}
	return nil
}

func (wh *PostgresCustomWebhook) validate(postgres *dbapi.Postgres) (admission.Warnings, error) {
	if postgres.Spec.Version == "" {
		return nil, errors.New(`'spec.version' is missing`)
	}
	var postgresVersion catalogapi.PostgresVersion
	err := wh.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name: postgres.Spec.Version,
	}, &postgresVersion)
	if err != nil {
		return nil, err
	}

	if postgres.Spec.Replicas == nil || ptr.Deref(postgres.Spec.Replicas, 0) < 1 {
		return nil, fmt.Errorf(`spec.replicas "%d" invalid. Value must be greater than zero`, ptr.Deref(postgres.Spec.Replicas, 0))
	}

	if err := wh.validateReadReplicas(postgres); err != nil {
		return nil, err
	}

	if err := wh.validateEnvsForAllContainers(postgres); err != nil {
		return nil, err
	}

	if postgres.Spec.StorageType == "" {
		return nil, fmt.Errorf(`'spec.storageType' is missing`)
	}
	if postgres.Spec.StorageType != dbapi.StorageTypeDurable && postgres.Spec.StorageType != dbapi.StorageTypeEphemeral {
		return nil, fmt.Errorf(`'spec.storageType' %s is invalid`, postgres.Spec.StorageType)
	}
	if err := amv.ValidateStorage(wh.DefaultClient, olddbapi.StorageType(postgres.Spec.StorageType), postgres.Spec.Storage); err != nil {
		return nil, err
	}

	if postgres.Spec.StandbyMode != nil {
		standByMode := *postgres.Spec.StandbyMode
		if standByMode != dbapi.HotPostgresStandbyMode &&
			standByMode != dbapi.WarmPostgresStandbyMode {
			return nil, fmt.Errorf(`spec.standbyMode "%s" invalid`, standByMode)
		}
	}

	if postgres.Spec.StreamingMode != nil {
		streamingMode := *postgres.Spec.StreamingMode
		// TODO: synchronous Streaming is unavailable due to lack of support
		if streamingMode != dbapi.AsynchronousPostgresStreamingMode &&
			streamingMode != dbapi.SynchronousPostgresStreamingMode {
			return nil, fmt.Errorf(`spec.streamingMode "%s" invalid`, streamingMode)
		}
	}

	if postgres.Spec.ClientAuthMode == dbapi.ClientAuthModeScram {
		if err := checkPgScramAuthMethodSupport(postgresVersion.Spec.Version); err != nil {
			return nil, err
		}
	}

	if postgres.Spec.Configuration != nil && len(postgres.Spec.Configuration.Inline) > 0 {
		if len(postgres.Spec.Configuration.Inline) > 1 {
			return nil, fmt.Errorf(`only one configuration source is allowed in spec.configuration.applyConfig and it should be %q`, kubedb.PostgresCustomConfigFile)
		}
		_, exists := postgres.Spec.Configuration.Inline[kubedb.PostgresCustomConfigFile]
		if !exists {
			return nil, fmt.Errorf(`invalid configuration source found in spec.configuration.applyConfig. only %q is allowed`, kubedb.PostgresCustomConfigFile)
		}
	}

	err = wh.validateVolumes(postgres)
	if err != nil {
		return nil, err
	}
	err = wh.validateVolumeMountsForAllContainers(postgres)
	if err != nil {
		return nil, err
	}

	err = wh.validateSpecForDB(postgres, &postgresVersion)
	if err != nil {
		return nil, err
	}

	// if secret managed externally verify auth secret name is not empty
	if postgres.Spec.AuthSecret != nil &&
		postgres.Spec.AuthSecret.ExternallyManaged &&
		postgres.Spec.AuthSecret.Name == "" {
		return nil, fmt.Errorf(`for externallyManaged auth secret, user must configure "spec.authSecret.name"`)
	}

	if wh.StrictValidation {

		// Check if postgresVersion is deprecated.
		// If deprecated, return error

		if postgresVersion.Spec.Deprecated {
			return nil, fmt.Errorf("postgres %s/%s is using deprecated version %v. Skipped processing",
				postgres.Namespace, postgres.Name, postgresVersion.Name)
		}

		if err := postgresVersion.ValidateSpecs(); err != nil {
			return nil, fmt.Errorf("postgres %s/%s is using invalid postgresVersion %v. Skipped processing. reason: %v", postgres.Namespace,
				postgres.Name, postgresVersion.Name, err)
		}
	}

	if postgres.Spec.DeletionPolicy == "" {
		return nil, fmt.Errorf(`'spec.deletionPolicy' is missing`)
	}

	if postgres.Spec.StorageType == dbapi.StorageTypeEphemeral && postgres.Spec.DeletionPolicy == dbapi.DeletionPolicyHalt {
		return nil, fmt.Errorf(`'spec.deletionPolicy: Halt' can not be used for 'Ephemeral' storage`)
	}

	monitorSpec := postgres.Spec.Monitor
	if monitorSpec != nil {
		if err := amv.ValidateMonitorSpec(monitorSpec); err != nil {
			return nil, err
		}
	}

	if err = amv.ValidateHealth(&postgres.Spec.HealthChecker); err != nil {
		return nil, err
	}
	if postgres.IsRemoteReplica() && postgresVersion.Spec.Version < "13" {
		return nil, fmt.Errorf("remote replica is not currently supported for version bellow 13")
	}

	if postgres.Spec.Distributed {
		if postgres.Spec.PodTemplate.Spec.PodPlacementPolicy == nil {
			return nil, fmt.Errorf(`'spec.podPlacementPolicy' is required for distributed postgres`)
		}
		pp := psapi.PlacementPolicy{}
		err = wh.DefaultClient.Get(context.Background(), client.ObjectKey{Name: postgres.Spec.PodTemplate.Spec.PodPlacementPolicy.Name}, &pp)
		if err != nil {
			return nil, err
		}
		if pp.Spec.ClusterSpreadConstraint == nil {
			return nil, fmt.Errorf(`'spec.clusterSpreadConstraint' is required in %v/%v for distributed postgres`, pp.Namespace, pp.Name)
		}
	}

	return nil, nil
}

func (wh *PostgresCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	postgres, ok := obj.(*dbapi.Postgres)
	if !ok {
		return nil, fmt.Errorf("expected a Postgres but got a %T", obj)
	}
	pgLog.Info("validating", "name", postgres.GetName())
	// Need a validation on readReplica Name, can't set arbiter as read replica name
	return wh.validate(postgres)
}

func (wh *PostgresCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	oldPostgres, ok := oldObj.(*dbapi.Postgres)
	if !ok {
		return nil, fmt.Errorf("expected a Postgres but got a %T", oldPostgres)
	}
	postgres, ok := newObj.(*dbapi.Postgres)
	if !ok {
		return nil, fmt.Errorf("expected a Postgres but got a %T", postgres)
	}

	var postgresVersion catalogapi.PostgresVersion
	err := wh.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name: oldPostgres.Spec.Version,
	}, &postgresVersion)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get PostgresVersion: %s", oldPostgres.Spec.Version)
	}
	oldPostgres.SetDefaults(&postgresVersion)
	if oldPostgres.Spec.AuthSecret == nil {
		oldPostgres.Spec.AuthSecret = postgres.Spec.AuthSecret
	}
	if err := wh.validateUpdate(postgres, oldPostgres); err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return wh.validate(postgres)
}

func (wh *PostgresCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	postgres, ok := obj.(*dbapi.Postgres)
	if !ok {
		return nil, fmt.Errorf("expected a Postgres but got a %T", obj)
	}

	var pg dbapi.Postgres
	err := wh.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      postgres.Name,
		Namespace: postgres.Namespace,
	}, &pg)
	if err != nil && !kerr.IsNotFound(err) {
		return nil, errors.Wrapf(err, "failed to get Postgres: %s", postgres.Name)
	} else if err == nil && pg.Spec.DeletionPolicy == dbapi.DeletionPolicyDoNotTerminate {
		return nil, fmt.Errorf(`postgres "%v/%v" can't be terminated. To delete, change spec.deletionPolicy`, pg.Namespace, pg.Name)
	}
	return nil, nil
}

func checkPgScramAuthMethodSupport(v string) error {
	pgVersion, err := semver.NewVersion(v)
	if err != nil {
		return err
	}
	if pgVersion.Major() < 11 {
		return fmt.Errorf("scram auth method is available only for 11 or higher Versions")
	}
	return nil
}
