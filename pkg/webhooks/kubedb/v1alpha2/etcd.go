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

package v1alpha2

import (
	"context"
	"fmt"

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	"kubedb.dev/apimachinery/apis/kubedb"
	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	amv "kubedb.dev/apimachinery/pkg/validator"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/mergepatch"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
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

// SetupEtcdWebhookWithManager registers the webhook for Etcd in the manager.
func SetupEtcdWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&olddbapi.Etcd{}).
		WithValidator(&EtcdCustomWebhook{DefaultClient: mgr.GetClient(), StrictValidation: true}).
		WithDefaulter(&EtcdCustomWebhook{DefaultClient: mgr.GetClient(), StrictValidation: true}).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-kubedb-com-v1alpha2-etcd,mutating=true,failurePolicy=fail,sideEffects=None,groups=kubedb.com,resources=etcds,verbs=create;update,versions=v1alpha2,name=metcd.kb.io,admissionReviewVersions=v1
//+kubebuilder:webhook:path=/validate-kubedb-com-v1alpha2-etcd,mutating=false,failurePolicy=fail,sideEffects=None,groups=kubedb.com,resources=etcds,verbs=create;update;delete,versions=v1alpha2,name=vetcd.kb.io,admissionReviewVersions=v1

// +kubebuilder:object:generate=false
type EtcdCustomWebhook struct {
	DefaultClient    client.Client
	StrictValidation bool
}

var (
	_ webhook.CustomDefaulter = &EtcdCustomWebhook{}
	_ webhook.CustomValidator = &EtcdCustomWebhook{}
)

// log is for logging in this package.
var etcdLog = logf.Log.WithName("etcd-resource")

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type
func (w *EtcdCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	if isDeletionInProgress(obj) {
		return nil
	}
	db, ok := obj.(*olddbapi.Etcd)
	if !ok {
		return fmt.Errorf("expected an Etcd object, got a %T", obj)
	}

	etcdLog.Info("default", "name", db.Name)

	if db.Spec.Version == "" {
		return errors.New(`'spec.version' is missing`)
	}

	if db.Spec.Halted {
		if db.Spec.DeletionPolicy == olddbapi.DeletionPolicyDoNotTerminate {
			return errors.New(`can't halt, since deletion policy is 'DoNotTerminate'`)
		}
		db.Spec.DeletionPolicy = olddbapi.DeletionPolicyHalt
	}

	// An odd member count keeps the Raft quorum optimal.
	if db.Spec.Replicas == nil {
		db.Spec.Replicas = ptr.To(int32(3))
	}

	var etcdVersion catalog.EtcdVersion
	if err := w.DefaultClient.Get(ctx, types.NamespacedName{Name: db.Spec.Version}, &etcdVersion); err != nil {
		return errors.Wrapf(err, "failed to get EtcdVersion %q", db.Spec.Version)
	}

	db.SetDefaults(&etcdVersion)
	return nil
}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type
func (w *EtcdCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	if isDeletionInProgress(obj) {
		return nil, nil
	}
	db, ok := obj.(*olddbapi.Etcd)
	if !ok {
		return nil, fmt.Errorf("expected an Etcd object, got a %T", obj)
	}

	etcdLog.Info("validate create", "name", db.Name)

	allErr := w.ValidateCreateOrUpdate(ctx, db)
	if len(allErr) == 0 {
		return nil, nil
	}
	return nil, apierrors.NewInvalid(schema.GroupKind{Group: kubedb.GroupName, Kind: olddbapi.ResourceKindEtcd}, db.Name, allErr)
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type
func (w *EtcdCustomWebhook) ValidateUpdate(ctx context.Context, old, newObj runtime.Object) (admission.Warnings, error) {
	if isDeletionInProgress(newObj) {
		return nil, nil
	}
	db, ok := newObj.(*olddbapi.Etcd)
	if !ok {
		return nil, fmt.Errorf("expected an Etcd object, got a %T", newObj)
	}
	oldDB, ok := old.(*olddbapi.Etcd)
	if !ok {
		return nil, fmt.Errorf("expected an Etcd object, got a %T", old)
	}

	etcdLog.Info("validate update", "name", db.Name)

	allErr := w.ValidateCreateOrUpdate(ctx, db)

	if err := validateEtcdUpdate(db, oldDB); err != nil {
		allErr = append(allErr, field.Invalid(field.NewPath("spec"), db.Name, err.Error()))
	}

	if len(allErr) == 0 {
		return nil, nil
	}
	return nil, apierrors.NewInvalid(schema.GroupKind{Group: kubedb.GroupName, Kind: olddbapi.ResourceKindEtcd}, db.Name, allErr)
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type
func (w *EtcdCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	db, ok := obj.(*olddbapi.Etcd)
	if !ok {
		return nil, fmt.Errorf("expected an Etcd object, got a %T", obj)
	}

	etcdLog.Info("validate delete", "name", db.Name)

	var allErr field.ErrorList
	if db.Spec.DeletionPolicy == olddbapi.DeletionPolicyDoNotTerminate {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("deletionPolicy"),
			db.Name,
			`can not delete as deletionPolicy is set to "DoNotTerminate"`))
		return nil, apierrors.NewInvalid(schema.GroupKind{Group: kubedb.GroupName, Kind: olddbapi.ResourceKindEtcd}, db.Name, allErr)
	}
	return nil, nil
}

// forbiddenEtcdEnvVars are owned by the operator: they are derived from the PetSet
// ordinal, the governing service and the current membership state of the cluster.
// Letting a user override any of them would desynchronize the member from the Raft
// cluster it is supposed to join.
var forbiddenEtcdEnvVars = []string{
	kubedb.EtcdEnvName,
	kubedb.EtcdEnvDataDir,
	kubedb.EtcdEnvInitialCluster,
	kubedb.EtcdEnvInitialClusterState,
	kubedb.EtcdEnvInitialAdvertisePeerURLs,
	kubedb.EtcdEnvListenPeerURLs,
	kubedb.EtcdEnvListenClientURLs,
	kubedb.EtcdEnvAdvertiseClientURLs,
}

var etcdReservedVolumes = []string{
	kubedb.EtcdDataVolumeName,
	kubedb.EtcdConfigVolumeName,
	kubedb.EtcdCustomConfigVolumeName,
	kubedb.EtcdInitScriptVolumeName,
	kubedb.EtcdServerTLSVolumeName,
	kubedb.EtcdClientTLSVolumeName,
	kubedb.EtcdPeerTLSVolumeName,
	kubedb.EtcdExporterTLSVolumeName,
}

var etcdReservedVolumeMountPaths = []string{
	kubedb.EtcdDataDir,
	kubedb.EtcdConfigDir,
	kubedb.EtcdCustomConfigDir,
	kubedb.EtcdInitScriptDir,
	kubedb.EtcdServerTLSMountPath,
	kubedb.EtcdClientTLSMountPath,
	kubedb.EtcdPeerTLSMountPath,
	kubedb.EtcdExporterTLSMountPath,
}

func (w *EtcdCustomWebhook) ValidateCreateOrUpdate(ctx context.Context, db *olddbapi.Etcd) field.ErrorList {
	var allErr field.ErrorList

	if db.Spec.Version == "" {
		allErr = append(allErr, field.Required(field.NewPath("spec").Child("version"), "spec.version is missing"))
	} else if w.StrictValidation {
		if err := w.validateEtcdVersion(ctx, db); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("version"), db.Spec.Version, err.Error()))
		}
	}

	// etcd forms a Raft quorum, so an odd replica count is what production
	// deployments should use. An even count is not rejected -- it is a legitimate
	// transient state during scaling -- it just buys no extra fault tolerance.
	if db.Spec.Replicas == nil {
		allErr = append(allErr, field.Required(field.NewPath("spec").Child("replicas"), "spec.replicas has to be defined"))
	} else if *db.Spec.Replicas < 1 {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("replicas"), *db.Spec.Replicas,
			"number of replicas can not be less than 1"))
	} else if *db.Spec.Replicas%2 == 0 {
		etcdLog.Info("even replica count buys no extra fault tolerance for a Raft quorum; an odd number is recommended",
			"name", db.Name, "replicas", *db.Spec.Replicas)
	}

	if err := w.validateEnvsForAllContainers(db); err != nil {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("podTemplate").Child("spec").Child("containers").Child("env"),
			db.Name, err.Error()))
	}

	if err := amv.ValidateVolumes(ofstv1.ConvertVolumes(db.Spec.PodTemplate.Spec.Volumes), etcdReservedVolumes); err != nil {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("podTemplate").Child("spec").Child("volumes"),
			db.Name, err.Error()))
	}

	if err := w.validateVolumeMountsForAllContainers(db); err != nil {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("podTemplate").Child("spec").Child("containers").Child("volumeMounts"),
			db.Name, err.Error()))
	}

	if db.Spec.StorageType == "" {
		allErr = append(allErr, field.Required(field.NewPath("spec").Child("storageType"), "spec.storageType can not be empty"))
	} else if db.Spec.StorageType != olddbapi.StorageTypeDurable && db.Spec.StorageType != olddbapi.StorageTypeEphemeral {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("storageType"), db.Spec.StorageType,
			"spec.storageType should be either Durable or Ephemeral"))
	}

	if err := amv.ValidateStorage(w.DefaultClient, db.Spec.StorageType, db.Spec.Storage); err != nil {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("storage"), db.Name, err.Error()))
	}

	if db.Spec.Monitor != nil {
		if err := amv.ValidateMonitorSpec(db.Spec.Monitor); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("monitor"), db.Name, err.Error()))
		}
	}

	if err := amv.ValidateHealth(&db.Spec.HealthChecker); err != nil {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("healthChecker"), db.Name, err.Error()))
	}

	if db.Spec.DeletionPolicy == "" {
		allErr = append(allErr, field.Required(field.NewPath("spec").Child("deletionPolicy"), "spec.deletionPolicy is missing"))
	} else if db.Spec.StorageType == olddbapi.StorageTypeEphemeral &&
		db.Spec.DeletionPolicy == olddbapi.DeletionPolicyHalt {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("deletionPolicy"), db.Spec.DeletionPolicy,
			`'spec.deletionPolicy: Halt' can not be used for 'Ephemeral' storage`))
	}

	if db.Spec.TLS != nil {
		if db.Spec.TLS.IssuerRef == nil {
			allErr = append(allErr, field.Required(field.NewPath("spec").Child("tls").Child("issuerRef"),
				"spec.tls.issuerRef is missing"))
		}

		allowedAliases := map[string]struct{}{
			string(olddbapi.EtcdServerCert):          {},
			string(olddbapi.EtcdClientCert):          {},
			string(olddbapi.EtcdPeerCert):            {},
			string(olddbapi.EtcdMetricsExporterCert): {},
		}
		for i, cert := range db.Spec.TLS.Certificates {
			if _, ok := allowedAliases[cert.Alias]; !ok {
				allErr = append(allErr, field.Invalid(
					field.NewPath("spec").Child("tls").Child("certificates").Index(i).Child("alias"),
					cert.Alias,
					fmt.Sprintf("supported aliases are %q, %q, %q and %q",
						olddbapi.EtcdServerCert, olddbapi.EtcdClientCert, olddbapi.EtcdPeerCert, olddbapi.EtcdMetricsExporterCert),
				))
				continue
			}
			if _, seen := kmapi.GetCertificate(db.Spec.TLS.Certificates[:i], cert.Alias); seen != nil {
				allErr = append(allErr, field.Duplicate(
					field.NewPath("spec").Child("tls").Child("certificates").Index(i).Child("alias"), cert.Alias,
				))
			}
		}
	}

	if len(allErr) == 0 {
		return nil
	}
	return allErr
}

// validateEtcdUpdate rejects mutations of the create-time settings. spec.storageType
// selects the volume shape of the PetSet, and spec.init only runs once, so neither
// can be flipped after the fact.
func validateEtcdUpdate(obj, oldObj *olddbapi.Etcd) error {
	preconditions := meta_util.PreConditionSet{
		Set: sets.New[string]("spec.storageType"),
	}
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

func (w *EtcdCustomWebhook) validateEtcdVersion(ctx context.Context, db *olddbapi.Etcd) error {
	etcdVersion := catalog.EtcdVersion{}
	if err := w.DefaultClient.Get(ctx, types.NamespacedName{Name: db.Spec.Version}, &etcdVersion); err != nil {
		return err
	}
	if err := etcdVersion.ValidateSpecs(); err != nil {
		return fmt.Errorf("etcd %s/%s is using invalid EtcdVersion %v. Skipped processing. reason: %v",
			db.Namespace, db.Name, etcdVersion.Name, err)
	}
	return nil
}

func (w *EtcdCustomWebhook) validateEnvsForAllContainers(db *olddbapi.Etcd) error {
	var err error
	for _, container := range db.Spec.PodTemplate.Spec.Containers {
		if errC := amv.ValidateEnvVar(container.Env, forbiddenEtcdEnvVars, olddbapi.ResourceKindEtcd); errC != nil {
			if err == nil {
				err = errC
			} else {
				err = errors.Wrap(err, errC.Error())
			}
		}
	}
	return err
}

func (w *EtcdCustomWebhook) validateVolumeMountsForAllContainers(db *olddbapi.Etcd) error {
	var err error
	for _, container := range db.Spec.PodTemplate.Spec.Containers {
		if errC := amv.ValidateMountPaths(container.VolumeMounts, etcdReservedVolumeMountPaths); errC != nil {
			if err == nil {
				err = errC
			} else {
				err = errors.Wrap(err, errC.Error())
			}
		}
	}
	return err
}
