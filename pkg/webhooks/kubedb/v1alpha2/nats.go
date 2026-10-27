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
	"strings"

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	"kubedb.dev/apimachinery/apis/kubedb"
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"
	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	amv "kubedb.dev/apimachinery/pkg/validator"

	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ofstv1 "kmodules.xyz/offshoot-api/api/v1"
	psapi "kubeops.dev/petset/apis/apps/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// SetupNATSWebhookWithManager registers the webhook for NATS in the manager.
func SetupNATSWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&olddbapi.NATS{}).
		WithValidator(amv.WithServiceTemplateAliasWarnings(&NATSCustomWebhook{DefaultClient: mgr.GetClient(), StrictValidation: true}, dbapi.PrimaryServiceAlias, dbapi.StatsServiceAlias, dbapi.LeafNodesServiceAlias)).
		WithDefaulter(&NATSCustomWebhook{DefaultClient: mgr.GetClient(), StrictValidation: true}).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-kubedb-com-v1alpha2-nats,mutating=true,failurePolicy=fail,sideEffects=None,groups=kubedb.com,resources=natses,verbs=create;update,versions=v1alpha2,name=mnats.kb.io,admissionReviewVersions=v1
//+kubebuilder:webhook:path=/validate-kubedb-com-v1alpha2-nats,mutating=false,failurePolicy=fail,sideEffects=None,groups=kubedb.com,resources=natses,verbs=create;update;delete,versions=v1alpha2,name=vnats.kb.io,admissionReviewVersions=v1

// +kubebuilder:object:generate=false
type NATSCustomWebhook struct {
	DefaultClient    client.Client
	StrictValidation bool
}

var (
	_ webhook.CustomDefaulter = &NATSCustomWebhook{}
	_ webhook.CustomValidator = &NATSCustomWebhook{}
)

// log is for logging in this package.
var natsLog = logf.Log.WithName("nats-resource")

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type
func (w *NATSCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	if isDeletionInProgress(obj) {
		return nil
	}
	db, ok := obj.(*olddbapi.NATS)
	if !ok {
		return fmt.Errorf("expected a NATS object, got a %T", obj)
	}

	natsLog.Info("default", "name", db.Name)

	if db.Spec.Version == "" {
		return errors.New(`'spec.version' is missing`)
	}

	if db.Spec.Halted {
		if db.Spec.DeletionPolicy == olddbapi.DeletionPolicyDoNotTerminate {
			return errors.New(`can't halt, since deletion policy is 'DoNotTerminate'`)
		}
		db.Spec.DeletionPolicy = olddbapi.DeletionPolicyHalt
	}

	var natsVersion catalog.NATSVersion
	if err := w.DefaultClient.Get(ctx, types.NamespacedName{Name: db.Spec.Version}, &natsVersion); err != nil {
		return errors.Wrapf(err, "failed to get NATSVersion %q", db.Spec.Version)
	}

	db.SetDefaults(&natsVersion)
	return nil
}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type
func (w *NATSCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	if isDeletionInProgress(obj) {
		return nil, nil
	}
	if err := amv.ValidateNameUniqueness(ctx, w.DefaultClient, obj); err != nil {
		return nil, err
	}
	db, ok := obj.(*olddbapi.NATS)
	if !ok {
		return nil, fmt.Errorf("expected a NATS object, got a %T", obj)
	}

	natsLog.Info("validate create", "name", db.Name)

	warnings, allErr := w.ValidateCreateOrUpdate(ctx, db, true)
	if len(allErr) == 0 {
		return warnings, nil
	}
	return warnings, apierrors.NewInvalid(schema.GroupKind{Group: kubedb.GroupName, Kind: olddbapi.ResourceKindNATS}, db.Name, allErr)
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type
func (w *NATSCustomWebhook) ValidateUpdate(ctx context.Context, old, newObj runtime.Object) (admission.Warnings, error) {
	if isDeletionInProgress(newObj) {
		return nil, nil
	}
	db, ok := newObj.(*olddbapi.NATS)
	if !ok {
		return nil, fmt.Errorf("expected a NATS object, got a %T", newObj)
	}
	oldDB, ok := old.(*olddbapi.NATS)
	if !ok {
		return nil, fmt.Errorf("expected a NATS object, got a %T", old)
	}

	natsLog.Info("validate update", "name", db.Name)

	// A version that is deprecated only blocks a new choice of version, never an unrelated update.
	warnings, allErr := w.ValidateCreateOrUpdate(ctx, db, db.Spec.Version != oldDB.Spec.Version)

	if err := validateNATSUpdate(db, oldDB); err != nil {
		allErr = append(allErr, field.Invalid(field.NewPath("spec"), db.Name, err.Error()))
	}

	if len(allErr) == 0 {
		return warnings, nil
	}
	return warnings, apierrors.NewInvalid(schema.GroupKind{Group: kubedb.GroupName, Kind: olddbapi.ResourceKindNATS}, db.Name, allErr)
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type
func (w *NATSCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	db, ok := obj.(*olddbapi.NATS)
	if !ok {
		return nil, fmt.Errorf("expected a NATS object, got a %T", obj)
	}

	natsLog.Info("validate delete", "name", db.Name)

	var allErr field.ErrorList
	if db.Spec.DeletionPolicy == olddbapi.DeletionPolicyDoNotTerminate {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("deletionPolicy"),
			db.Name,
			`can not delete as deletionPolicy is set to "DoNotTerminate"`))
		return nil, apierrors.NewInvalid(schema.GroupKind{Group: kubedb.GroupName, Kind: olddbapi.ResourceKindNATS}, db.Name, allErr)
	}
	return nil, nil
}

// forbiddenNATSEnvVars are owned by the operator.
var forbiddenNATSEnvVars = []string{
	kubedb.NATSEnvServerName,
	kubedb.NATSEnvGoMemLimit,
}

var natsReservedVolumes = []string{
	kubedb.NATSDataVolumeName,
	kubedb.NATSConfigVolumeName,
	kubedb.NATSPidVolumeName,
	kubedb.NATSTLSVolumeName,
	kubedb.NATSPodConfigVolumeName,
	kubedb.NATSResolverVolumeName,
}

var natsReservedVolumeMountPaths = []string{
	kubedb.NATSDataDir,
	kubedb.NATSConfigDir,
	kubedb.NATSPidDir,
	kubedb.NATSTLSDir,
	kubedb.NATSPodConfigDir,
	kubedb.NATSResolverDir,
}

// ValidateCreateOrUpdate checks every rule of design section 4, "Webhook validation rules". It returns
// the warnings alongside the errors. checkDeprecated rejects a deprecated NATSVersion.
func (w *NATSCustomWebhook) ValidateCreateOrUpdate(ctx context.Context, db *olddbapi.NATS, checkDeprecated bool) (admission.Warnings, field.ErrorList) {
	var allErr field.ErrorList
	var warnings admission.Warnings
	specPath := field.NewPath("spec")

	// version exists in the catalog and is not deprecated
	if db.Spec.Version == "" {
		allErr = append(allErr, field.Required(specPath.Child("version"), "spec.version is missing"))
	} else if w.StrictValidation {
		if err := w.validateNATSVersion(ctx, db, checkDeprecated); err != nil {
			allErr = append(allErr, field.Invalid(specPath.Child("version"), db.Spec.Version, err.Error()))
		}
	}

	// replicas >= 1; with JetStream on: 1 or >= 3, 2 is rejected, even numbers >= 4 warn
	jsOn := db.IsJetStreamEnabled()
	if db.Spec.Replicas == nil {
		allErr = append(allErr, field.Required(specPath.Child("replicas"), "spec.replicas has to be defined"))
	} else {
		replicas := *db.Spec.Replicas
		switch {
		case replicas < 1:
			allErr = append(allErr, field.Invalid(specPath.Child("replicas"), replicas,
				"number of replicas can not be less than 1"))
		case jsOn && replicas == 2:
			allErr = append(allErr, field.Invalid(specPath.Child("replicas"), replicas,
				"with JetStream enabled spec.replicas must be 1 or at least 3, two servers have no quorum"))
		case jsOn && replicas >= 4 && replicas%2 == 0:
			warnings = append(warnings, fmt.Sprintf("spec.replicas %d is even: it tolerates no more failures than %d replicas", replicas, replicas-1))
		}
		if replicas > 1 && !db.Spec.Distributed {
			warnings = append(warnings, w.placementPolicyWarnings(ctx, db)...)
		}
	}

	// jetStream: ephemeral requires replicas == 1; durable needs a PVC request
	if db.Spec.JetStream != nil {
		js := db.Spec.JetStream
		if js.StorageType != "" && js.StorageType != olddbapi.StorageTypeDurable && js.StorageType != olddbapi.StorageTypeEphemeral {
			allErr = append(allErr, field.Invalid(specPath.Child("jetStream", "storageType"), js.StorageType,
				"spec.jetStream.storageType should be either Durable or Ephemeral"))
		}
		if jsOn {
			if js.StorageType == olddbapi.StorageTypeEphemeral && db.Spec.Replicas != nil && *db.Spec.Replicas != 1 {
				allErr = append(allErr, field.Invalid(specPath.Child("jetStream", "storageType"), js.StorageType,
					"Ephemeral JetStream storage is only allowed with spec.replicas == 1"))
			}
			if err := amv.ValidateStorage(w.DefaultClient, js.StorageType, js.Storage, "spec.jetStream.storage"); err != nil {
				allErr = append(allErr, field.Invalid(specPath.Child("jetStream", "storage"), db.Name, err.Error()))
			}
		}
	}

	// limits.memory is set; memoryStoreSize <= 50% of limits.memory
	memLimit := db.MemoryLimit()
	if memLimit == nil || memLimit.IsZero() {
		allErr = append(allErr, field.Required(specPath.Child("podTemplate", "spec", "containers").Key(kubedb.NATSContainerName).Child("resources", "limits", "memory"),
			"limits.memory of the nats container is required"))
	} else if db.MaxMemoryStoreBytes() > 0 && db.MaxMemoryStoreBytes()*100 > memLimit.Value()*kubedb.NATSMaxMemoryStorePercent {
		allErr = append(allErr, field.Invalid(specPath.Child("jetStream", "memoryStoreSize"), db.Spec.JetStream.MemoryStoreSize.String(),
			fmt.Sprintf("memoryStoreSize must not exceed %d%% of limits.memory (%s)", kubedb.NATSMaxMemoryStorePercent, memLimit.String())))
	}

	// lame-duck: gracePeriod < duration, duration >= 30s, terminationGracePeriodSeconds >= computed minimum
	allErr = append(allErr, olddbapi.ValidateNATSLameDuck(db.Spec.LameDuck, specPath.Child("lameDuck"))...)
	if t := db.Spec.PodTemplate.Spec.TerminationGracePeriodSeconds; t != nil && *t < db.MinTerminationGracePeriodSeconds() {
		allErr = append(allErr, field.Invalid(specPath.Child("podTemplate", "spec", "terminationGracePeriodSeconds"), *t,
			fmt.Sprintf("terminationGracePeriodSeconds must be at least %d (lameDuck.duration + 30s, minimum 60s)", db.MinTerminationGracePeriodSeconds())))
	}

	// typed server limits
	allErr = append(allErr, olddbapi.ValidateNATSServerConfig(db.Spec.ServerConfig, specPath.Child("serverConfig"))...)

	// operator-owned keys are rejected in configuration
	allErr = append(allErr, w.validateNATSConfiguration(ctx, db)...)

	// TLS
	if db.Spec.TLS != nil {
		if db.Spec.TLS.IssuerRef == nil {
			allErr = append(allErr, field.Required(specPath.Child("tls", "issuerRef"), "spec.tls.issuerRef is missing"))
		}
		allowedAliases := map[string]struct{}{
			string(olddbapi.NATSServerCert): {},
			string(olddbapi.NATSClientCert): {},
		}
		for i, cert := range db.Spec.TLS.Certificates {
			if _, ok := allowedAliases[cert.Alias]; !ok {
				allErr = append(allErr, field.Invalid(specPath.Child("tls", "certificates").Index(i).Child("alias"), cert.Alias,
					fmt.Sprintf("supported aliases are %q and %q", olddbapi.NATSServerCert, olddbapi.NATSClientCert)))
				continue
			}
			for j := 0; j < i; j++ {
				if db.Spec.TLS.Certificates[j].Alias == cert.Alias {
					allErr = append(allErr, field.Duplicate(specPath.Child("tls", "certificates").Index(i).Child("alias"), cert.Alias))
					break
				}
			}
		}
	} else {
		// clusterTLS (and the other channel switches) can only be set with tls
		for name, sw := range map[string]*bool{"clientTLS": db.Spec.ClientTLS, "clusterTLS": db.Spec.ClusterTLS, "monitorTLS": db.Spec.MonitorTLS, "gatewayTLS": db.Spec.GatewayTLS, "leafNodeTLS": db.Spec.LeafNodeTLS} {
			if sw != nil {
				allErr = append(allErr, field.Forbidden(specPath.Child(name), fmt.Sprintf("spec.%s can only be set together with spec.tls", name)))
			}
		}
	}

	// auth
	allErr = append(allErr, validateNATSAuth(db)...)

	// gateway, leafnodes, distributed and arbiter (design 15.1 to 15.4)
	mdWarnings, mdErr := w.validateNATSMultiDC(ctx, db)
	warnings = append(warnings, mdWarnings...)
	allErr = append(allErr, mdErr...)

	// halted: true sets deletionPolicy Halt, or is rejected if DoNotTerminate
	if db.Spec.Halted && db.Spec.DeletionPolicy == olddbapi.DeletionPolicyDoNotTerminate {
		allErr = append(allErr, field.Invalid(specPath.Child("halted"), db.Spec.Halted,
			`can't halt, since deletion policy is 'DoNotTerminate'`))
	}
	if db.Spec.DeletionPolicy == "" {
		allErr = append(allErr, field.Required(specPath.Child("deletionPolicy"), "spec.deletionPolicy is missing"))
	} else if db.IsEphemeral() && db.Spec.DeletionPolicy == olddbapi.DeletionPolicyHalt {
		allErr = append(allErr, field.Invalid(specPath.Child("deletionPolicy"), db.Spec.DeletionPolicy,
			`'spec.deletionPolicy: Halt' can not be used for 'Ephemeral' storage`))
	}

	// pod template
	if err := w.validateEnvsForAllContainers(db); err != nil {
		allErr = append(allErr, field.Invalid(specPath.Child("podTemplate", "spec", "containers", "env"), db.Name, err.Error()))
	}
	if err := amv.ValidateVolumes(ofstv1.ConvertVolumes(db.Spec.PodTemplate.Spec.Volumes), natsReservedVolumes); err != nil {
		allErr = append(allErr, field.Invalid(specPath.Child("podTemplate", "spec", "volumes"), db.Name, err.Error()))
	}
	if err := w.validateVolumeMountsForAllContainers(db); err != nil {
		allErr = append(allErr, field.Invalid(specPath.Child("podTemplate", "spec", "containers", "volumeMounts"), db.Name, err.Error()))
	}

	if db.Spec.Monitor != nil {
		if err := amv.ValidateMonitorSpec(db.Spec.Monitor); err != nil {
			allErr = append(allErr, field.Invalid(specPath.Child("monitor"), db.Name, err.Error()))
		}
	}
	if err := amv.ValidateHealth(&db.Spec.HealthChecker); err != nil {
		allErr = append(allErr, field.Invalid(specPath.Child("healthChecker"), db.Name, err.Error()))
	}

	return warnings, allErr
}

func validateNATSAuth(db *olddbapi.NATS) field.ErrorList {
	var allErr field.ErrorList
	path := field.NewPath("spec", "auth")
	if db.Spec.Auth == nil {
		return nil
	}
	switch db.Spec.Auth.Mode {
	case "", olddbapi.NATSAuthModeConfig:
		if db.Spec.Auth.Operator != nil {
			allErr = append(allErr, field.Forbidden(path.Child("operator"), "spec.auth.operator is only allowed with spec.auth.mode: Operator"))
		}
	case olddbapi.NATSAuthModeOperator:
		if db.Spec.DisableSecurity {
			allErr = append(allErr, field.Invalid(field.NewPath("spec", "disableSecurity"), db.Spec.DisableSecurity,
				"disableSecurity is not allowed with spec.auth.mode: Operator"))
		}
		if op := db.Spec.Auth.Operator; op != nil {
			if op.ExternalSecret != nil && op.ExternalSecret.Name == "" {
				allErr = append(allErr, field.Required(path.Child("operator", "externalSecret", "name"), "externalSecret.name can not be empty"))
			}
			if op.ResolverStorage != nil {
				if q, ok := op.ResolverStorage.Resources.Requests[core.ResourceStorage]; !ok || q.Value() <= 0 {
					allErr = append(allErr, field.Invalid(path.Child("operator", "resolverStorage"), db.Name, "missing or invalid ResourceStorage request"))
				}
			}
		}
	default:
		allErr = append(allErr, field.NotSupported(path.Child("mode"), db.Spec.Auth.Mode,
			[]string{string(olddbapi.NATSAuthModeConfig), string(olddbapi.NATSAuthModeOperator)}))
	}
	return allErr
}

// validateNATSConfiguration rejects unsupported keys and operator-owned settings in spec.configuration,
// both in inline and in the referenced Secret (when it already exists).
func (w *NATSCustomWebhook) validateNATSConfiguration(ctx context.Context, db *olddbapi.NATS) field.ErrorList {
	var allErr field.ErrorList
	cfg := db.Spec.Configuration
	if cfg == nil {
		return nil
	}
	path := field.NewPath("spec", "configuration")
	mode := db.AuthMode()

	for key, content := range cfg.Inline {
		if err := olddbapi.ValidateNATSUserConfig(key, content, mode); err != nil {
			allErr = append(allErr, field.Invalid(path.Child("inline").Key(key), key, err.Error()))
		}
	}

	if cfg.SecretName != "" && w.DefaultClient != nil {
		var sec core.Secret
		err := w.DefaultClient.Get(ctx, types.NamespacedName{Namespace: db.Namespace, Name: cfg.SecretName}, &sec)
		switch {
		case err == nil:
			for key, content := range sec.Data {
				if err := olddbapi.ValidateNATSUserConfig(key, string(content), mode); err != nil {
					allErr = append(allErr, field.Invalid(path.Child("secretName"), cfg.SecretName, err.Error()))
				}
			}
		case apierrors.IsNotFound(err):
			// the user creates the Secret; the renderer runs the same check
		default:
			allErr = append(allErr, field.InternalError(path.Child("secretName"), err))
		}
	}
	return allErr
}

// placementPolicyWarnings warns, but does not reject, when the referenced PlacementPolicy does not
// spread pods across zones with DoNotSchedule.
func (w *NATSCustomWebhook) placementPolicyWarnings(ctx context.Context, db *olddbapi.NATS) admission.Warnings {
	ref := db.Spec.PodTemplate.Spec.PodPlacementPolicy
	if ref == nil || w.DefaultClient == nil {
		return nil
	}
	var pp psapi.PlacementPolicy
	if err := w.DefaultClient.Get(ctx, client.ObjectKey{Name: ref.Name}, &pp); err != nil {
		return nil
	}
	if zc := pp.Spec.ZoneSpreadConstraint; zc == nil || zc.WhenUnsatisfiable != core.DoNotSchedule {
		return admission.Warnings{fmt.Sprintf("PlacementPolicy %q should have a zoneSpreadConstraint with whenUnsatisfiable: DoNotSchedule, otherwise two pods of a JetStream R3 group can land in one zone", ref.Name)}
	}
	return nil
}

// validateNATSUpdate rejects mutations of the create-time settings. Defaulting fills in the unset
// values, so an unset field and its default are compared as equal.
func validateNATSUpdate(obj, oldObj *olddbapi.NATS) error {
	if obj.IsJetStreamEnabled() != oldObj.IsJetStreamEnabled() {
		return fmt.Errorf("spec.jetStream.enabled is immutable")
	}
	if obj.IsJetStreamEnabled() && jetStreamStorageType(obj) != jetStreamStorageType(oldObj) {
		return fmt.Errorf("spec.jetStream.storageType is immutable")
	}
	if obj.AuthMode() != oldObj.AuthMode() {
		return fmt.Errorf("spec.auth.mode is immutable")
	}
	if obj.Spec.Distributed != oldObj.Spec.Distributed {
		return fmt.Errorf("spec.distributed is immutable")
	}
	if obj.Spec.DisableSecurity != oldObj.Spec.DisableSecurity {
		return fmt.Errorf("spec.disableSecurity is immutable")
	}
	return nil
}

func jetStreamStorageType(db *olddbapi.NATS) olddbapi.StorageType {
	if db.Spec.JetStream == nil || db.Spec.JetStream.StorageType == "" {
		return olddbapi.StorageTypeDurable
	}
	return db.Spec.JetStream.StorageType
}

func (w *NATSCustomWebhook) validateNATSVersion(ctx context.Context, db *olddbapi.NATS, checkDeprecated bool) error {
	natsVersion := catalog.NATSVersion{}
	if err := w.DefaultClient.Get(ctx, types.NamespacedName{Name: db.Spec.Version}, &natsVersion); err != nil {
		if apierrors.IsNotFound(err) {
			return fmt.Errorf("NATSVersion %q not found in the catalog", db.Spec.Version)
		}
		return err
	}
	if err := natsVersion.ValidateSpecs(); err != nil {
		return fmt.Errorf("nats %s/%s is using invalid NATSVersion %v. Skipped processing. reason: %v",
			db.Namespace, db.Name, natsVersion.Name, err)
	}
	if checkDeprecated && natsVersion.Spec.Deprecated {
		return fmt.Errorf("NATSVersion %q is deprecated, use a newer version", db.Spec.Version)
	}
	return nil
}

func (w *NATSCustomWebhook) validateEnvsForAllContainers(db *olddbapi.NATS) error {
	var msgs []string
	for _, container := range db.Spec.PodTemplate.Spec.Containers {
		if err := amv.ValidateEnvVar(container.Env, forbiddenNATSEnvVars, olddbapi.ResourceKindNATS); err != nil {
			msgs = append(msgs, err.Error())
		}
	}
	if len(msgs) > 0 {
		return errors.New(strings.Join(msgs, "; "))
	}
	return nil
}

func (w *NATSCustomWebhook) validateVolumeMountsForAllContainers(db *olddbapi.NATS) error {
	var msgs []string
	for _, container := range db.Spec.PodTemplate.Spec.Containers {
		if err := amv.ValidateMountPaths(container.VolumeMounts, natsReservedVolumeMountPaths); err != nil {
			msgs = append(msgs, err.Error())
		}
	}
	if len(msgs) > 0 {
		return errors.New(strings.Join(msgs, "; "))
	}
	return nil
}
