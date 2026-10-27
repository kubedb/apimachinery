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

package v1alpha1

import (
	"context"
	"fmt"
	"strings"

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	"kubedb.dev/apimachinery/apis/kubedb"
	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"
	secret_lib "kubedb.dev/apimachinery/pkg/secret"
	opsutil "kubedb.dev/apimachinery/pkg/webhooks/ops"

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
	vsecretapi "go.virtual-secrets.dev/apimachinery/apis/virtual/v1alpha1"
	core "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/mergepatch"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
	meta_util "kmodules.xyz/client-go/meta"
	psapi "kubeops.dev/petset/apis/apps/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func SetupNATSOpsRequestWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&opsapi.NATSOpsRequest{}).
		WithValidator(&NATSOpsRequestCustomWebhook{mgr.GetClient()}).
		Complete()
}

type NATSOpsRequestCustomWebhook struct {
	DefaultClient client.Client
}

var natsOpsLog = logf.Log.WithName("nats-opsrequest")

var _ webhook.CustomValidator = &NATSOpsRequestCustomWebhook{}

func (w *NATSOpsRequestCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	req, ok := obj.(*opsapi.NATSOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected a NATSOpsRequest object but got %T", obj)
	}
	natsOpsLog.Info("validate create", "name", req.Name)
	return nil, w.validateCreateOrUpdate(req)
}

func (w *NATSOpsRequestCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	req, ok := newObj.(*opsapi.NATSOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected a NATSOpsRequest object but got %T", newObj)
	}
	oldReq, ok := oldObj.(*opsapi.NATSOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected a NATSOpsRequest object but got %T", oldObj)
	}
	natsOpsLog.Info("validate update", "name", req.Name)

	if err := validateNATSOpsRequest(req, oldReq); err != nil {
		return nil, err
	}
	if isOpsReqCompleted(req.Status.Phase) && !isOpsReqCompleted(oldReq.Status.Phase) { // just completed
		var db olddbapi.NATS
		err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: req.Spec.DatabaseRef.Name, Namespace: req.Namespace}, &db)
		if err != nil {
			return nil, err
		}
		return nil, resumeDatabase(w.DefaultClient, &db)
	}
	return nil, nil
}

func (w *NATSOpsRequestCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func validateNATSOpsRequest(req *opsapi.NATSOpsRequest, oldReq *opsapi.NATSOpsRequest) error {
	preconditions := meta_util.PreConditionSet{Set: sets.New[string]("spec")}
	_, err := meta_util.CreateStrategicPatch(oldReq, req, preconditions.PreconditionFunc()...)
	if err != nil {
		if mergepatch.IsPreconditionFailed(err) {
			return fmt.Errorf("%v.%v", err, preconditions.Error())
		}
		return err
	}
	return nil
}

func (w *NATSOpsRequestCustomWebhook) validateCreateOrUpdate(req *opsapi.NATSOpsRequest) error {
	if validType := req.Spec.Type.IsValid(); !validType {
		return field.Invalid(field.NewPath("spec").Child("type"), req.Name,
			fmt.Sprintf("defined OpsRequestType %s is not supported, supported types for NATS are %s", req.Spec.Type, strings.Join(opsapi.NATSOpsRequestTypeNames(), ", ")))
	}

	db, err := w.hasDatabaseRef(req)
	if err != nil {
		return field.Invalid(field.NewPath("spec").Child("databaseRef"), req.Name, err.Error())
	}

	var allErr field.ErrorList
	switch req.Spec.Type {
	case opsapi.NATSOpsRequestTypeUpdateVersion:
		if err := w.validateNATSUpdateVersionOpsRequest(db, req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("updateVersion"), req.Name, err.Error()))
		}
	case opsapi.NATSOpsRequestTypeRestart:
	case opsapi.NATSOpsRequestTypeHorizontalScaling:
		if err := w.validateNATSHorizontalScalingOpsRequest(db, req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("horizontalScaling"), req.Name, err.Error()))
		}
	case opsapi.NATSOpsRequestTypeVerticalScaling:
		if err := w.validateNATSVerticalScalingOpsRequest(db, req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("verticalScaling"), req.Name, err.Error()))
		}
	case opsapi.NATSOpsRequestTypeVolumeExpansion:
		if err := w.validateNATSVolumeExpansionOpsRequest(db, req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("volumeExpansion"), req.Name, err.Error()))
		}
	case opsapi.NATSOpsRequestTypeReconfigure:
		if err := w.validateNATSReconfigureOpsRequest(db, req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("configuration"), req.Name, err.Error()))
		}
	case opsapi.NATSOpsRequestTypeReconfigureTLS:
		if err := w.validateNATSReconfigureTLSOpsRequest(db, req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("tls"), req.Name, err.Error()))
		}
	case opsapi.NATSOpsRequestTypeRotateAuth:
		if err := w.validateNATSRotateAuthenticationOpsRequest(db, req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("authentication"), req.Name, err.Error()))
		}
	case opsapi.NATSOpsRequestTypeStorageMigration:
		if err := w.validateNATSStorageMigrationOpsRequest(db, req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("migration"), req.Name, err.Error()))
		}
	}

	if len(allErr) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "natsopsrequests.ops.kubedb.com", Kind: opsapi.ResourceKindNATSOpsRequest}, req.Name, allErr)
}

func (w *NATSOpsRequestCustomWebhook) hasDatabaseRef(req *opsapi.NATSOpsRequest) (*olddbapi.NATS, error) {
	db := &olddbapi.NATS{}
	if err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      req.Spec.DatabaseRef.Name,
		Namespace: req.Namespace,
	}, db); err != nil {
		return nil, fmt.Errorf("spec.databaseRef %s/%s is invalid or not found", req.Namespace, req.Spec.DatabaseRef.Name)
	}
	return db, nil
}

// validateNATSUpdateVersionOpsRequest: the target exists and is not deprecated, the move is allowed by
// the current version's updateConstraints (one minor at a time), and it is not a downgrade, because a
// JetStream store written by a newer server may not be readable by an older one.
func (w *NATSOpsRequestCustomWebhook) validateNATSUpdateVersionOpsRequest(db *olddbapi.NATS, req *opsapi.NATSOpsRequest) error {
	updateVersionSpec := req.Spec.UpdateVersion
	if updateVersionSpec == nil {
		return errors.New("spec.updateVersion is nil, not supported in UpdateVersion type")
	}
	if updateVersionSpec.TargetVersion == "" {
		return errors.New("spec.updateVersion.targetVersion can not be empty")
	}

	var target catalog.NATSVersion
	if err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: updateVersionSpec.TargetVersion}, &target); err != nil {
		if apierrors.IsNotFound(err) {
			return fmt.Errorf("NATSVersion %q not found in the catalog", updateVersionSpec.TargetVersion)
		}
		return err
	}
	if target.Spec.Deprecated {
		return fmt.Errorf("NATSVersion %q is deprecated", updateVersionSpec.TargetVersion)
	}
	if err := target.ValidateSpecs(); err != nil {
		return err
	}

	cur, err := semver.NewVersion(db.Spec.Version)
	if err != nil {
		return fmt.Errorf("current version %q is not a semantic version: %w", db.Spec.Version, err)
	}
	tgt, err := semver.NewVersion(target.Spec.Version)
	if err != nil {
		return fmt.Errorf("target version %q is not a semantic version: %w", target.Spec.Version, err)
	}
	if !tgt.GreaterThan(cur) {
		return fmt.Errorf("downgrade or no-op from version %v to %v is not supported", db.Spec.Version, target.Spec.Version)
	}

	yes, err := IsUpgradable(w.DefaultClient, catalog.ResourceKindNATSVersion, db.Spec.Version, updateVersionSpec.TargetVersion)
	if err != nil {
		return err
	}
	if !yes {
		return fmt.Errorf("upgrade from version %v to %v is not supported", db.Spec.Version, updateVersionSpec.TargetVersion)
	}
	return nil
}

// validateNATSHorizontalScalingOpsRequest: with JetStream on the size stays 1 or >= 3, an ephemeral
// store stays at 1, and a cluster of 3 or more never shrinks below 3. Whether a stream has more
// replicas than the target size needs /jsz and is checked by the controller.
func (w *NATSOpsRequestCustomWebhook) validateNATSHorizontalScalingOpsRequest(db *olddbapi.NATS, req *opsapi.NATSOpsRequest) error {
	horizontalScalingSpec := req.Spec.HorizontalScaling
	if horizontalScalingSpec == nil {
		return errors.New("spec.horizontalScaling is nil, not supported in HorizontalScaling type")
	}
	if db.Spec.Distributed {
		return w.validateNATSDistributedScaling(db, horizontalScalingSpec)
	}
	if len(horizontalScalingSpec.DataCenters) > 0 {
		return errors.New("spec.horizontalScaling.dataCenters is only allowed for a distributed NATS")
	}
	if horizontalScalingSpec.Replicas == nil {
		return errors.New("spec.horizontalScaling.replicas can not be empty")
	}
	target := *horizontalScalingSpec.Replicas
	if target < 1 {
		return errors.New("spec.horizontalScaling.replicas must be at least 1")
	}
	if !db.IsJetStreamEnabled() {
		return nil
	}
	if target == 2 {
		return errors.New("with JetStream enabled spec.horizontalScaling.replicas must be 1 or at least 3, two servers have no quorum")
	}
	if db.IsEphemeral() && target != 1 {
		return errors.New("an Ephemeral JetStream store only allows 1 replica")
	}
	if db.Spec.Replicas != nil && *db.Spec.Replicas >= 3 && target < 3 {
		return fmt.Errorf("scaling a JetStream cluster below 3 replicas is not supported, current replicas: %d", *db.Spec.Replicas)
	}
	return nil
}

// validateNATSDistributedScaling: a distributed NATS scales per DC. Every entry names a distributionRule
// of the PlacementPolicy; a Member DC keeps the JetStream rule (1 or at least 3), the Arbiter DC at least 1.
func (w *NATSOpsRequestCustomWebhook) validateNATSDistributedScaling(db *olddbapi.NATS, hs *opsapi.NATSHorizontalScalingSpec) error {
	if hs.Replicas != nil {
		return errors.New("spec.horizontalScaling.replicas is not allowed for a distributed NATS, use spec.horizontalScaling.dataCenters")
	}
	if len(hs.DataCenters) == 0 {
		return errors.New("spec.horizontalScaling.dataCenters can not be empty for a distributed NATS")
	}
	ref := db.Spec.PodTemplate.Spec.PodPlacementPolicy
	if ref == nil {
		return errors.New("the distributed NATS has no spec.podTemplate.spec.podPlacementPolicy")
	}
	var pp psapi.PlacementPolicy
	if err := w.DefaultClient.Get(context.TODO(), client.ObjectKey{Name: ref.Name}, &pp); err != nil {
		return fmt.Errorf("failed to get PlacementPolicy %q: %w", ref.Name, err)
	}
	if pp.Spec.ClusterSpreadConstraint == nil {
		return fmt.Errorf("PlacementPolicy %q has no clusterSpreadConstraint", ref.Name)
	}
	roles := map[string]psapi.DCRole{}
	for _, rule := range pp.Spec.ClusterSpreadConstraint.DistributionRules {
		role := rule.Role
		if role == "" {
			role = psapi.DCRoleMember
		}
		roles[rule.ClusterName] = role
	}
	seen := map[string]struct{}{}
	for i, dc := range hs.DataCenters {
		role, ok := roles[dc.ClusterName]
		if !ok {
			return fmt.Errorf("spec.horizontalScaling.dataCenters[%d]: cluster %q is not in PlacementPolicy %q", i, dc.ClusterName, ref.Name)
		}
		if _, dup := seen[dc.ClusterName]; dup {
			return fmt.Errorf("spec.horizontalScaling.dataCenters[%d]: cluster %q is listed twice", i, dc.ClusterName)
		}
		seen[dc.ClusterName] = struct{}{}
		if dc.Replicas < 1 {
			return fmt.Errorf("spec.horizontalScaling.dataCenters[%d].replicas must be at least 1", i)
		}
		if role == psapi.DCRoleMember && db.IsJetStreamEnabled() {
			if dc.Replicas == 2 {
				return fmt.Errorf("spec.horizontalScaling.dataCenters[%d].replicas: with JetStream a DC must have 1 or at least 3 servers", i)
			}
			if db.IsEphemeral() && dc.Replicas != 1 {
				return errors.New("an Ephemeral JetStream store only allows 1 replica per DC")
			}
		}
	}
	return nil
}

// validateNATSVerticalScalingOpsRequest: limits.memory stays set and at least twice max_memory_store.
func (w *NATSOpsRequestCustomWebhook) validateNATSVerticalScalingOpsRequest(db *olddbapi.NATS, req *opsapi.NATSOpsRequest) error {
	verticalScalingSpec := req.Spec.VerticalScaling
	if verticalScalingSpec == nil {
		return errors.New("spec.verticalScaling is nil, not supported in VerticalScaling type")
	}
	if verticalScalingSpec.NATS == nil && verticalScalingSpec.Reloader == nil && verticalScalingSpec.Exporter == nil {
		return errors.New("at least one of spec.verticalScaling.nats, spec.verticalScaling.reloader or spec.verticalScaling.exporter must be specified")
	}

	if n := verticalScalingSpec.NATS; n != nil {
		limit, ok := n.Resources.Limits[core.ResourceMemory]
		if !ok || limit.IsZero() {
			return errors.New("spec.verticalScaling.nats.resources.limits.memory is required")
		}
		if req, ok := n.Resources.Requests[core.ResourceMemory]; ok && req.Cmp(limit) > 0 {
			return errors.New("spec.verticalScaling.nats.resources.requests.memory can not exceed the limit")
		}
		if store := db.MaxMemoryStoreBytes(); store > 0 && limit.Value() < store*2 {
			return fmt.Errorf("limits.memory (%s) must be at least twice the JetStream memoryStoreSize (%s)",
				limit.String(), db.Spec.JetStream.MemoryStoreSize.String())
		}
	}
	return nil
}

func (w *NATSOpsRequestCustomWebhook) validateNATSVolumeExpansionOpsRequest(db *olddbapi.NATS, req *opsapi.NATSOpsRequest) error {
	volumeExpansionSpec := req.Spec.VolumeExpansion
	if volumeExpansionSpec == nil {
		return errors.New("spec.volumeExpansion is nil, not supported in VolumeExpansion type")
	}
	if volumeExpansionSpec.NATS == nil {
		return errors.New("spec.volumeExpansion.nats can not be empty")
	}
	if volumeExpansionSpec.Mode == "" {
		return errors.New("spec.volumeExpansion.mode is required")
	}
	if !db.IsJetStreamEnabled() || db.IsEphemeral() {
		return errors.New("volume expansion needs a durable JetStream file store")
	}

	return opsutil.ValidateStorageExpansion(db.Spec.JetStream.Storage, volumeExpansionSpec.NATS, req.Status.Phase, "NATS")
}

func (w *NATSOpsRequestCustomWebhook) validateNATSStorageMigrationOpsRequest(db *olddbapi.NATS, req *opsapi.NATSOpsRequest) error {
	migrationSpec := req.Spec.Migration
	if migrationSpec == nil {
		return errors.New("spec.migration is required for StorageMigration type")
	}
	if migrationSpec.StorageClassName == nil {
		return errors.New("spec.migration.storageClassName is required")
	}
	if req.Spec.Timeout == nil {
		return errors.New("spec.timeout is required for Storage Migration ops request, adjust timeout according to the size of your database")
	}
	if !db.IsJetStreamEnabled() || db.IsEphemeral() {
		return errors.New("storage migration needs a durable JetStream file store")
	}
	// A pod's PVC is replaced with an empty one and the streams catch up from the other replicas, so a
	// cluster without replicas would lose its only copy.
	if db.Spec.Replicas == nil || *db.Spec.Replicas < 3 {
		return errors.New("storage migration re-syncs each pod from its peers and needs at least 3 replicas, otherwise the only copy of the data would be lost")
	}
	if db.Spec.JetStream.Storage == nil || db.Spec.JetStream.Storage.StorageClassName == nil {
		return nil
	}

	var newStorage, oldStorage storagev1.StorageClass
	if err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: *migrationSpec.StorageClassName}, &newStorage); err != nil {
		if apierrors.IsNotFound(err) {
			return fmt.Errorf("storage class %s not found: %w", *migrationSpec.StorageClassName, err)
		}
		return err
	}
	if err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: *db.Spec.JetStream.Storage.StorageClassName}, &oldStorage); err != nil {
		return err
	}
	if oldStorage.VolumeBindingMode != nil && *oldStorage.VolumeBindingMode == storagev1.VolumeBindingWaitForFirstConsumer {
		if newStorage.VolumeBindingMode == nil || *newStorage.VolumeBindingMode != storagev1.VolumeBindingWaitForFirstConsumer {
			return fmt.Errorf("volume binding mode should be WaitForFirstConsumer for %s storageClass", newStorage.Name)
		}
	}
	return nil
}

func (w *NATSOpsRequestCustomWebhook) validateNATSRotateAuthenticationOpsRequest(db *olddbapi.NATS, req *opsapi.NATSOpsRequest) error {
	if db.Spec.DisableSecurity {
		return errors.New("there is no credential to rotate, the NATS database has spec.disableSecurity: true")
	}
	authSpec := req.Spec.Authentication
	if authSpec == nil || authSpec.SecretRef == nil {
		return nil
	}
	if db.IsOperatorMode() {
		return errors.New("spec.authentication.secretRef is not supported with spec.auth.mode: Operator, credentials are re-issued from the account signing key")
	}
	if authSpec.SecretRef.Name == "" {
		return errors.New("spec.authentication.secretRef.name can not be empty")
	}

	isVirtual := authSpec.SecretRef.APIGroup == vsecretapi.GroupName
	newData, err := secret_lib.GetData(context.TODO(), w.DefaultClient, req.Namespace, authSpec.SecretRef.Name, isVirtual)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return errors.Errorf("referenced secret %s/%s not found", req.Namespace, authSpec.SecretRef.Name)
		}
		return err
	}
	if password := newData[core.BasicAuthPasswordKey]; len(password) == 0 {
		return errors.Errorf("referenced secret %s/%s must contain a non-empty %q", req.Namespace, authSpec.SecretRef.Name, core.BasicAuthPasswordKey)
	}
	if username := newData[core.BasicAuthUsernameKey]; len(username) > 0 && string(username) != kubedb.NATSUserAdmin {
		return errors.Errorf("username in referenced secret %s/%s must be %q", req.Namespace, authSpec.SecretRef.Name, kubedb.NATSUserAdmin)
	}
	return nil
}

// validateNATSReconfigureOpsRequest applies the same operator-owned-key rule as the database webhook to
// applyConfig, and the typed-field rules to serverConfig and lameDuck.
func (w *NATSOpsRequestCustomWebhook) validateNATSReconfigureOpsRequest(db *olddbapi.NATS, req *opsapi.NATSOpsRequest) error {
	cfg := req.Spec.Configuration
	if cfg == nil {
		return errors.New("spec.configuration is nil, not supported in Reconfigure type")
	}
	if cfg.ConfigSecret == nil && len(cfg.ApplyConfig) == 0 && !cfg.RemoveCustomConfig && cfg.ServerConfig == nil && cfg.LameDuck == nil {
		return errors.New("spec.configuration has nothing to apply, set configSecret, applyConfig, removeCustomConfig, serverConfig or lameDuck")
	}
	if cfg.ConfigSecret != nil && cfg.ConfigSecret.Name == "" {
		return errors.New("spec.configuration.configSecret.name can not be empty")
	}
	for key, content := range cfg.ApplyConfig {
		if err := olddbapi.ValidateNATSUserConfig(key, content, db.AuthMode()); err != nil {
			return fmt.Errorf("spec.configuration.applyConfig[%s]: %w", key, err)
		}
	}
	if errs := olddbapi.ValidateNATSServerConfig(cfg.ServerConfig, field.NewPath("spec", "configuration", "serverConfig")); len(errs) > 0 {
		return errs.ToAggregate()
	}
	if errs := olddbapi.ValidateNATSLameDuck(cfg.LameDuck, field.NewPath("spec", "configuration", "lameDuck")); len(errs) > 0 {
		return errs.ToAggregate()
	}
	return nil
}

func (w *NATSOpsRequestCustomWebhook) validateNATSReconfigureTLSOpsRequest(db *olddbapi.NATS, req *opsapi.NATSOpsRequest) error {
	tls := req.Spec.TLS
	if tls == nil {
		return errors.New("spec.tls is nil, not supported in ReconfigureTLS type")
	}

	certUpdateRequested := tls.IssuerRef != nil || len(tls.Certificates) > 0

	opCount := 0
	if tls.Remove {
		opCount++
	}
	if tls.RotateCertificates {
		opCount++
	}
	if certUpdateRequested {
		opCount++
	}
	if opCount == 0 {
		return errors.New("at least one of Remove, RotateCertificates, IssuerRef, or Certificates must be specified")
	}
	if opCount > 1 {
		return errors.New("only one TLS reconfiguration operation is allowed at a time")
	}

	for _, cert := range tls.Certificates {
		if cert.Alias != string(olddbapi.NATSServerCert) && cert.Alias != string(olddbapi.NATSClientCert) {
			return fmt.Errorf("unsupported certificate alias %q, supported aliases are %q and %q", cert.Alias, olddbapi.NATSServerCert, olddbapi.NATSClientCert)
		}
	}

	if tls.Remove {
		if db.Spec.TLS == nil {
			return errors.New("TLS is not enabled on the NATS database, nothing to remove")
		}
		return nil
	}

	if tls.RotateCertificates {
		if db.Spec.TLS == nil || db.Spec.TLS.IssuerRef == nil {
			return errors.New("rotateCertificates requires TLS to already be enabled with issuerRef on NATS")
		}
		return nil
	}

	if certUpdateRequested && tls.IssuerRef == nil && (db.Spec.TLS == nil || db.Spec.TLS.IssuerRef == nil) {
		return errors.New("tls.issuerRef is required for NATS ReconfigureTLS")
	}
	return nil
}
