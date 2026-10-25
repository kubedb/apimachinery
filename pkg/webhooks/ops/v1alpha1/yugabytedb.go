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
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"
	secret_lib "kubedb.dev/apimachinery/pkg/secret"
	opsutil "kubedb.dev/apimachinery/pkg/webhooks/ops"

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
	vsecretapi "go.virtual-secrets.dev/apimachinery/apis/virtual/v1alpha1"
	"gomodules.xyz/x/arrays"
	core "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/mergepatch"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
	kmc "kmodules.xyz/client-go/conditions"
	meta_util "kmodules.xyz/client-go/meta"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// SetupYugabyteDBOpsRequestWebhookWithManager registers the webhook for YugabyteDBOpsRequest in the manager.
func SetupYugabyteDBOpsRequestWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&opsapi.YugabyteDBOpsRequest{}).
		WithValidator(&YugabyteDBOpsRequestCustomWebhook{mgr.GetClient()}).
		Complete()
}

type YugabyteDBOpsRequestCustomWebhook struct {
	DefaultClient client.Client
}

var yugabyteDBLog = logf.Log.WithName("yugabytedb-opsrequest")

var _ webhook.CustomValidator = &YugabyteDBOpsRequestCustomWebhook{}

func (w *YugabyteDBOpsRequestCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	req, ok := obj.(*opsapi.YugabyteDBOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected a YugabyteDBOpsRequest object but got %T", obj)
	}
	yugabyteDBLog.Info("validate create", "name", req.Name)
	return w.validateCreateOrUpdate(ctx, req)
}

func (w *YugabyteDBOpsRequestCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	req, ok := newObj.(*opsapi.YugabyteDBOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected a YugabyteDBOpsRequest object but got %T", newObj)
	}
	oldReq, ok := oldObj.(*opsapi.YugabyteDBOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected a YugabyteDBOpsRequest object but got %T", oldObj)
	}
	yugabyteDBLog.Info("validate update", "name", req.Name)

	if err := validateYugabyteDBOpsRequestImmutable(req, oldReq); err != nil {
		return nil, err
	}
	warnings, err := w.validateCreateOrUpdate(ctx, req)
	if err != nil {
		return warnings, err
	}
	if isOpsReqCompleted(req.Status.Phase) && !isOpsReqCompleted(oldReq.Status.Phase) { // just completed
		var db dbapi.YugabyteDB
		if err := w.DefaultClient.Get(ctx, types.NamespacedName{Name: req.Spec.DatabaseRef.Name, Namespace: req.Namespace}, &db); err != nil {
			return warnings, err
		}
		return warnings, resumeDatabase(w.DefaultClient, &db)
	}
	return warnings, nil
}

func (w *YugabyteDBOpsRequestCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	if _, ok := obj.(*opsapi.YugabyteDBOpsRequest); !ok {
		return nil, fmt.Errorf("expected a YugabyteDBOpsRequest object but got %T", obj)
	}
	return nil, nil
}

// validateYugabyteDBOpsRequestImmutable rejects any change of the spec.
func validateYugabyteDBOpsRequestImmutable(req, oldReq *opsapi.YugabyteDBOpsRequest) error {
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

func yugabyteDBOpsInvalid(name string, allErr field.ErrorList) error {
	return apierrors.NewInvalid(schema.GroupKind{Group: "yugabytedbopsrequests.kubedb.com", Kind: opsapi.ResourceKindYugabyteDBOpsRequest}, name, allErr)
}

// isPending reports whether the request has not started. The checks that compare the request with the current
// state of the database only make sense then, since a running request changes the database itself.
func yugabyteDBOpsPending(req *opsapi.YugabyteDBOpsRequest) bool {
	return req.Status.Phase == "" || req.Status.Phase == opsapi.OpsRequestPhasePending
}

func (w *YugabyteDBOpsRequestCustomWebhook) validateCreateOrUpdate(ctx context.Context, req *opsapi.YugabyteDBOpsRequest) (admission.Warnings, error) {
	if validType, _ := arrays.Contains(opsapi.YugabyteDBOpsRequestTypeNames(), string(req.Spec.Type)); !validType {
		return nil, yugabyteDBOpsInvalid(req.Name, field.ErrorList{field.Invalid(field.NewPath("spec").Child("type"), req.Name,
			fmt.Sprintf("defined OpsRequestType %s is not supported, supported types for YugabyteDB are %s", req.Spec.Type, strings.Join(opsapi.YugabyteDBOpsRequestTypeNames(), ", ")))})
	}

	db, err := w.hasDatabaseRef(ctx, req)
	if err != nil {
		return nil, err
	}

	spec := field.NewPath("spec")
	var allErr field.ErrorList
	var warnings admission.Warnings
	add := func(child string, err error) {
		if err != nil {
			allErr = append(allErr, field.Invalid(spec.Child(child), req.Name, err.Error()))
		}
	}
	if req.Spec.ConnectionManager != nil && req.Spec.Type != opsapi.YugabyteDBOpsRequestTypeReconfigure {
		add("connectionManager", fmt.Errorf("spec.connectionManager is only valid in a Reconfigure request, not in %s", req.Spec.Type))
	}
	switch opsapi.YugabyteDBOpsRequestType(req.GetRequestType()) {
	case opsapi.YugabyteDBOpsRequestTypeRestart:
	case opsapi.YugabyteDBOpsRequestTypeUpdateVersion:
		add("updateVersion", w.validateUpdateVersion(ctx, db, req))
		add("type", w.validateOneUpgradeAtATime(ctx, db, req))
	case opsapi.YugabyteDBOpsRequestTypeFinalizeUpgrade, opsapi.YugabyteDBOpsRequestTypeRollbackUpgrade:
		add("type", w.validateFinalizeOrRollback(ctx, db, req))
	case opsapi.YugabyteDBOpsRequestTypeHorizontalScaling:
		add("horizontalScaling", w.validateHorizontalScaling(db, req))
	case opsapi.YugabyteDBOpsRequestTypeVerticalScaling:
		add("verticalScaling", w.validateVerticalScaling(db, req))
	case opsapi.YugabyteDBOpsRequestTypeStorageMigration:
		warns, err := w.validateStorageMigration(ctx, db, req)
		warnings = append(warnings, warns...)
		add("migration", err)
	case opsapi.YugabyteDBOpsRequestTypeVolumeExpansion:
		add("volumeExpansion", w.validateVolumeExpansion(db, req))
	case opsapi.YugabyteDBOpsRequestTypeReconfigure:
		add("configuration", w.validateReconfigure(ctx, db, req))
	case opsapi.YugabyteDBOpsRequestTypeReconfigureTLS:
		add("tls", w.validateReconfigureTLS(db, req))
	case opsapi.YugabyteDBOpsRequestTypeRotateAuth:
		add("authentication", w.validateRotateAuth(ctx, db, req))
	}

	if len(allErr) == 0 {
		return warnings, nil
	}
	return warnings, yugabyteDBOpsInvalid(req.Name, allErr)
}

func (w *YugabyteDBOpsRequestCustomWebhook) hasDatabaseRef(ctx context.Context, req *opsapi.YugabyteDBOpsRequest) (*dbapi.YugabyteDB, error) {
	var db dbapi.YugabyteDB
	if err := w.DefaultClient.Get(ctx, types.NamespacedName{Name: req.GetDBRefName(), Namespace: req.GetNamespace()}, &db); err != nil {
		allErr := field.ErrorList{field.Invalid(field.NewPath("spec").Child("databaseRef"), req.GetDBRefName(), fmt.Sprintf("%s/%s is invalid or not found", req.GetNamespace(), req.GetDBRefName()))}
		return nil, yugabyteDBOpsInvalid(req.Name, allErr)
	}
	return &db, nil
}

func (w *YugabyteDBOpsRequestCustomWebhook) validateUpdateVersion(ctx context.Context, db *dbapi.YugabyteDB, req *opsapi.YugabyteDBOpsRequest) error {
	uv := req.Spec.UpdateVersion
	if uv == nil {
		return errors.New("spec.updateVersion nil not supported in UpdateVersion type")
	}
	if uv.TargetVersion == "" {
		return errors.New("spec.updateVersion.targetVersion is required")
	}
	switch uv.Finalize {
	case "", opsapi.YugabyteDBFinalizeModeAuto, opsapi.YugabyteDBFinalizeModeManual:
	default:
		return fmt.Errorf("spec.updateVersion.finalize %q is not valid, use %s or %s", uv.Finalize, opsapi.YugabyteDBFinalizeModeAuto, opsapi.YugabyteDBFinalizeModeManual)
	}
	if uv.SoakDuration != nil && uv.SoakDuration.Duration < 0 {
		return errors.New("spec.updateVersion.soakDuration must not be negative")
	}

	var target catalog.YugabyteDBVersion
	if err := w.DefaultClient.Get(ctx, types.NamespacedName{Name: uv.TargetVersion}, &target); err != nil {
		if apierrors.IsNotFound(err) {
			return fmt.Errorf("target version %q does not exist in the catalog", uv.TargetVersion)
		}
		return err
	}
	if target.Spec.Deprecated {
		return fmt.Errorf("target version %q is deprecated", uv.TargetVersion)
	}
	if !yugabyteDBOpsPending(req) {
		return nil
	}
	if db.Spec.Version == uv.TargetVersion {
		return fmt.Errorf("target version %q is the current version", uv.TargetVersion)
	}

	var current catalog.YugabyteDBVersion
	if err := w.DefaultClient.Get(ctx, types.NamespacedName{Name: db.Spec.Version}, &current); err != nil {
		return errors.Wrapf(err, "failed to get the current YugabyteDBVersion %q", db.Spec.Version)
	}
	if current.Spec.YSQLMajorVersion != target.Spec.YSQLMajorVersion {
		return fmt.Errorf("an update between YSQL major versions (%d to %d) is not supported", current.Spec.YSQLMajorVersion, target.Spec.YSQLMajorVersion)
	}
	ok, err := yugabyteDBUpgradeAllowed(&current, &target)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("upgrade from version %v to %v is not allowed by the updateConstraints of %v", db.Spec.Version, uv.TargetVersion, uv.TargetVersion)
	}
	return nil
}

// yugabyteDBIsUpgradeType reports the ops request types that move the version of the universe. Only one of them may
// be active on a database at a time.
func yugabyteDBIsUpgradeType(t opsapi.YugabyteDBOpsRequestType) bool {
	switch t {
	case opsapi.YugabyteDBOpsRequestTypeUpdateVersion, opsapi.YugabyteDBOpsRequestTypeFinalizeUpgrade, opsapi.YugabyteDBOpsRequestTypeRollbackUpgrade:
		return true
	}
	return false
}

// validateOneUpgradeAtATime rejects a new upgrade related request while another one of the same database has not
// completed. A FinalizeUpgrade or RollbackUpgrade may run next to the UpdateVersion request that left the
// upgrade pending, since that request stays active until the upgrade is completed. A pending upgrade of the
// database also blocks a new UpdateVersion.
func (w *YugabyteDBOpsRequestCustomWebhook) validateOneUpgradeAtATime(ctx context.Context, db *dbapi.YugabyteDB, req *opsapi.YugabyteDBOpsRequest) error {
	if !yugabyteDBOpsPending(req) {
		return nil
	}
	pendingFinalize := kmc.IsConditionTrue(db.Status.Conditions, kubedb.YugabyteDBConditionUpgradePendingFinalize)
	if req.Spec.Type == opsapi.YugabyteDBOpsRequestTypeUpdateVersion && pendingFinalize {
		return errors.New("an upgrade of this database is pending finalization, send a FinalizeUpgrade or RollbackUpgrade request first")
	}
	var list opsapi.YugabyteDBOpsRequestList
	if err := w.DefaultClient.List(ctx, &list, client.InNamespace(req.Namespace)); err != nil {
		return err
	}
	for i := range list.Items {
		o := &list.Items[i]
		if o.Name == req.Name || o.Spec.DatabaseRef.Name != req.Spec.DatabaseRef.Name || !yugabyteDBIsUpgradeType(o.Spec.Type) || isOpsReqCompleted(o.Status.Phase) {
			continue
		}
		if o.Spec.Type == opsapi.YugabyteDBOpsRequestTypeUpdateVersion && req.Spec.Type != opsapi.YugabyteDBOpsRequestTypeUpdateVersion && pendingFinalize {
			continue
		}
		return fmt.Errorf("the upgrade related request %s (%s) of this database has not completed, only one upgrade related request can run at a time", o.Name, o.Spec.Type)
	}
	return nil
}

// validateFinalizeOrRollback checks a FinalizeUpgrade or RollbackUpgrade request. Both are only valid while the
// database has the UpgradePendingFinalize condition, which an UpdateVersion with finalize: Manual sets.
func (w *YugabyteDBOpsRequestCustomWebhook) validateFinalizeOrRollback(ctx context.Context, db *dbapi.YugabyteDB, req *opsapi.YugabyteDBOpsRequest) error {
	if yugabyteDBOpsPending(req) && !kmc.IsConditionTrue(db.Status.Conditions, kubedb.YugabyteDBConditionUpgradePendingFinalize) {
		return fmt.Errorf("%s needs a pending upgrade: the database has no %s condition, run an UpdateVersion with spec.updateVersion.finalize: Manual first",
			req.Spec.Type, kubedb.YugabyteDBConditionUpgradePendingFinalize)
	}
	return w.validateOneUpgradeAtATime(ctx, db, req)
}

// yugabyteDBUpgradeAllowed evaluates the updateConstraints of the target version against the current version.
// As in the design, the allowlist of a YugabyteDBVersion lists the versions it can be reached from. An empty
// allowlist accepts every version that is not in the denylist.
func yugabyteDBUpgradeAllowed(current, target *catalog.YugabyteDBVersion) (bool, error) {
	cur, err := semver.NewVersion(current.Spec.Version)
	if err != nil {
		return false, fmt.Errorf("current version %q is not a semantic version: %w", current.Spec.Version, err)
	}
	uc := target.Spec.UpdateConstraints
	for _, d := range uc.Denylist {
		c, err := semver.NewConstraint(d)
		if err != nil {
			return false, fmt.Errorf("invalid denylist constraint %q of %s: %w", d, target.Name, err)
		}
		if c.Check(cur) {
			return false, nil
		}
	}
	if len(uc.Allowlist) == 0 {
		return true, nil
	}
	for _, a := range uc.Allowlist {
		c, err := semver.NewConstraint(a)
		if err != nil {
			return false, fmt.Errorf("invalid allowlist constraint %q of %s: %w", a, target.Name, err)
		}
		if c.Check(cur) {
			return true, nil
		}
	}
	return false, nil
}

func (w *YugabyteDBOpsRequestCustomWebhook) validateHorizontalScaling(db *dbapi.YugabyteDB, req *opsapi.YugabyteDBOpsRequest) error {
	hs := req.Spec.HorizontalScaling
	if hs == nil {
		return errors.New("spec.horizontalScaling nil not supported in HorizontalScaling type")
	}
	if hs.TServer == nil && hs.ReadReplica == nil {
		return errors.New("at least one of spec.horizontalScaling.tserver and spec.horizontalScaling.readReplica is required, only the tservers and the read replicas can be scaled")
	}
	if hs.TServer != nil {
		target := *hs.TServer
		rf := db.ReplicationFactor()
		if target < rf {
			return fmt.Errorf("spec.horizontalScaling.tserver (%d) must be at least the replication factor (%d)", target, rf)
		}
		if cur := db.Replicas(dbapi.YugabyteDBRoleTServer); yugabyteDBOpsPending(req) && cur == target && hs.ReadReplica == nil {
			return fmt.Errorf("target tserver replicas %d is same as current replicas %d", target, cur)
		}
	}
	if hs.ReadReplica != nil {
		target := *hs.ReadReplica
		if !db.IsReadReplicaEnabled() {
			return errors.New("spec.horizontalScaling.readReplica: the database has no spec.topology.readReplica")
		}
		if target < 0 {
			return fmt.Errorf("spec.horizontalScaling.readReplica (%d) must not be negative", target)
		}
		if rf := db.ReadReplicaReplicationFactor(); target != 0 && target < rf {
			return fmt.Errorf("spec.horizontalScaling.readReplica (%d) must be 0 or at least the read replica replicationFactor (%d)", target, rf)
		}
		if cur := db.Replicas(dbapi.YugabyteDBRoleReadReplica); yugabyteDBOpsPending(req) && cur == target && hs.TServer == nil {
			return fmt.Errorf("target read replica replicas %d is same as current replicas %d", target, cur)
		}
	}
	return nil
}

func validateYugabyteDBRoleResources(role string, pr *opsapi.PodResources) error {
	mem, ok := pr.Resources.Limits[core.ResourceMemory]
	if !ok || mem.IsZero() {
		return fmt.Errorf("spec.verticalScaling.%s.resources.limits.memory must be set", role)
	}
	for name, limit := range pr.Resources.Limits {
		if req, ok := pr.Resources.Requests[name]; ok && req.Cmp(limit) > 0 {
			return fmt.Errorf("spec.verticalScaling.%s.resources.requests.%s must not exceed the limit", role, name)
		}
	}
	return nil
}

func (w *YugabyteDBOpsRequestCustomWebhook) validateVerticalScaling(db *dbapi.YugabyteDB, req *opsapi.YugabyteDBOpsRequest) error {
	vs := req.Spec.VerticalScaling
	if vs == nil {
		return errors.New("spec.verticalScaling nil not supported in VerticalScaling type")
	}
	switch vs.Mode {
	case "", opsapi.VerticalScalingModeRestart, opsapi.VerticalScalingModeInPlace:
	default:
		return fmt.Errorf("spec.verticalScaling.mode %q is not valid, use %s or %s", vs.Mode, opsapi.VerticalScalingModeRestart, opsapi.VerticalScalingModeInPlace)
	}
	if vs.Master == nil && vs.TServer == nil && vs.ReadReplica == nil {
		return errors.New("at least one of spec.verticalScaling.master, spec.verticalScaling.tserver and spec.verticalScaling.readReplica must be set")
	}
	if vs.ReadReplica != nil {
		if !db.IsReadReplicaEnabled() {
			return errors.New("spec.verticalScaling.readReplica: the database has no spec.topology.readReplica")
		}
		if err := validateYugabyteDBRoleResources("readReplica", vs.ReadReplica); err != nil {
			return err
		}
	}
	if vs.Master != nil {
		if err := validateYugabyteDBRoleResources("master", vs.Master); err != nil {
			return err
		}
	}
	if vs.TServer != nil {
		if err := validateYugabyteDBRoleResources("tserver", vs.TServer); err != nil {
			return err
		}
	}
	if vs.Mode == opsapi.VerticalScalingModeInPlace && yugabyteDBOpsPending(req) {
		// Only the CPU can be resized in place. The memory limit of the processes is derived from the container
		// memory when they start (--memory_limit_hard_bytes and the PostgreSQL memory), so a new memory value
		// needs a restart.
		for role, pr := range map[dbapi.YugabyteDBRole]*opsapi.PodResources{
			dbapi.YugabyteDBRoleMaster: vs.Master, dbapi.YugabyteDBRoleTServer: vs.TServer, dbapi.YugabyteDBRoleReadReplica: vs.ReadReplica,
		} {
			if pr == nil {
				continue
			}
			cur := yugabyteDBContainerResources(db, role)
			for _, c := range []struct {
				kind     string
				old, new core.ResourceList
			}{{"limits", cur.Limits, pr.Resources.Limits}, {"requests", cur.Requests, pr.Resources.Requests}} {
				want, ok := c.new[core.ResourceMemory]
				if !ok {
					continue
				}
				have := c.old[core.ResourceMemory]
				if want.Cmp(have) != 0 {
					return fmt.Errorf("spec.verticalScaling.%s.resources.%s.memory (%s) differs from the current value (%s): only the CPU can be resized in place, use mode Restart to change the memory",
						role, c.kind, want.String(), have.String())
				}
			}
		}
	}
	return nil
}

func yugabyteDBContainerResources(db *dbapi.YugabyteDB, role dbapi.YugabyteDBRole) core.ResourceRequirements {
	for _, c := range db.Node(role).PodTemplate.Spec.Containers {
		if c.Name == db.ContainerName(role) {
			return c.Resources
		}
	}
	return core.ResourceRequirements{}
}

// validateStorageMigration checks a StorageMigration request. It warns when the tservers cannot absorb the data of
// the pod that is being migrated.
func (w *YugabyteDBOpsRequestCustomWebhook) validateStorageMigration(ctx context.Context, db *dbapi.YugabyteDB, req *opsapi.YugabyteDBOpsRequest) (admission.Warnings, error) {
	m := req.Spec.Migration
	if m == nil {
		return nil, errors.New("spec.migration nil not supported in StorageMigration type")
	}
	if m.StorageClassName == nil || *m.StorageClassName == "" {
		return nil, errors.New("spec.migration.storageClassName is required")
	}
	switch m.OldPVReclaimPolicy {
	case "", core.PersistentVolumeReclaimRetain, core.PersistentVolumeReclaimDelete:
	default:
		return nil, fmt.Errorf("spec.migration.oldPVReclaimPolicy %q is not valid, use Retain or Delete", m.OldPVReclaimPolicy)
	}
	if req.Spec.Timeout == nil {
		return nil, errors.New("spec.timeout is required for a StorageMigration, the operation moves the data of every pod, set it according to the size of the database")
	}
	for _, role := range db.Roles() {
		if db.Node(role).StorageType == dbapi.StorageTypeEphemeral {
			return nil, fmt.Errorf("the %s uses Ephemeral storage and has no volume to migrate", role)
		}
	}
	var newSC storagev1.StorageClass
	if err := w.DefaultClient.Get(ctx, types.NamespacedName{Name: *m.StorageClassName}, &newSC); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("storage class %q not found", *m.StorageClassName)
		}
		return nil, err
	}
	if !yugabyteDBOpsPending(req) {
		return nil, nil
	}
	same := true
	for _, role := range db.Roles() {
		if db.GetStorageClassName(role) != *m.StorageClassName {
			same = false
		}
	}
	if same {
		return nil, fmt.Errorf("every volume of the database already uses the storage class %q", *m.StorageClassName)
	}
	if cur := db.GetStorageClassName(dbapi.YugabyteDBRoleTServer); cur != "" {
		var oldSC storagev1.StorageClass
		if err := w.DefaultClient.Get(ctx, types.NamespacedName{Name: cur}, &oldSC); err == nil &&
			oldSC.VolumeBindingMode != nil && *oldSC.VolumeBindingMode == storagev1.VolumeBindingWaitForFirstConsumer &&
			(newSC.VolumeBindingMode == nil || *newSC.VolumeBindingMode != storagev1.VolumeBindingWaitForFirstConsumer) {
			return nil, fmt.Errorf("the volume binding mode of storage class %q must be WaitForFirstConsumer, as it is for %q", newSC.Name, cur)
		}
	}
	var warnings admission.Warnings
	if n, rf := db.Replicas(dbapi.YugabyteDBRoleTServer), db.ReplicationFactor(); n < rf+1 {
		warnings = append(warnings, fmt.Sprintf("the database has %d tservers, fewer than the replication factor plus one (%d): the data of a tserver cannot be moved off it before its volume is replaced, so the migration may not complete", n, rf+1))
	}
	return warnings, nil
}

func (w *YugabyteDBOpsRequestCustomWebhook) validateVolumeExpansion(db *dbapi.YugabyteDB, req *opsapi.YugabyteDBOpsRequest) error {
	ve := req.Spec.VolumeExpansion
	if ve == nil {
		return errors.New("spec.volumeExpansion nil not supported in VolumeExpansion type")
	}
	if ve.Mode != opsapi.VolumeExpansionModeOnline && ve.Mode != opsapi.VolumeExpansionModeOffline {
		return fmt.Errorf("spec.volumeExpansion.mode must be %s or %s", opsapi.VolumeExpansionModeOnline, opsapi.VolumeExpansionModeOffline)
	}
	if ve.Master == nil && ve.TServer == nil && ve.ReadReplica == nil {
		return errors.New("at least one of spec.volumeExpansion.master, spec.volumeExpansion.tserver and spec.volumeExpansion.readReplica must be set")
	}
	if ve.ReadReplica != nil {
		if !db.IsReadReplicaEnabled() {
			return errors.New("spec.volumeExpansion.readReplica: the database has no spec.topology.readReplica")
		}
		if err := w.validateRoleExpansion(db, dbapi.YugabyteDBRoleReadReplica, req); err != nil {
			return err
		}
	}
	if ve.Master != nil {
		if err := w.validateRoleExpansion(db, dbapi.YugabyteDBRoleMaster, req); err != nil {
			return err
		}
	}
	if ve.TServer != nil {
		if err := w.validateRoleExpansion(db, dbapi.YugabyteDBRoleTServer, req); err != nil {
			return err
		}
	}
	return nil
}

func (w *YugabyteDBOpsRequestCustomWebhook) validateRoleExpansion(db *dbapi.YugabyteDB, role dbapi.YugabyteDBRole, req *opsapi.YugabyteDBOpsRequest) error {
	node := db.Node(role)
	if node.StorageType == dbapi.StorageTypeEphemeral {
		return fmt.Errorf("the %s uses Ephemeral storage and cannot be expanded", role)
	}
	q := req.Spec.VolumeExpansion.Master
	switch role {
	case dbapi.YugabyteDBRoleTServer:
		q = req.Spec.VolumeExpansion.TServer
	case dbapi.YugabyteDBRoleReadReplica:
		q = req.Spec.VolumeExpansion.ReadReplica
	}
	return opsutil.ValidateStorageExpansion(node.Storage, q, req.Status.Phase, "YugabyteDB "+string(role))
}

func (w *YugabyteDBOpsRequestCustomWebhook) validateReconfigure(ctx context.Context, db *dbapi.YugabyteDB, req *opsapi.YugabyteDBOpsRequest) error {
	cfg := req.Spec.Configuration
	if cm := req.Spec.ConnectionManager; cm != nil {
		if err := w.validateConnectionManager(db, req); err != nil {
			return err
		}
		if cfg == nil {
			return nil
		}
	}
	if cfg == nil {
		return errors.New("spec.configuration nil not supported in Reconfigure type")
	}
	if !cfg.RemoveCustomConfig && cfg.ConfigSecret == nil && len(cfg.ApplyConfig) == 0 {
		return errors.New("at least one of `removeCustomConfig`, `configSecret` or `applyConfig` must be specified")
	}
	switch cfg.Restart {
	case "", opsapi.ReconfigureRestartAuto, opsapi.ReconfigureRestartTrue, opsapi.ReconfigureRestartFalse:
	default:
		return fmt.Errorf("spec.configuration.restart %q is not valid, use auto, true or false", cfg.Restart)
	}
	if errs := dbapi.ValidateYugabyteDBConfigData(cfg.ApplyConfig); len(errs) > 0 {
		return fmt.Errorf("spec.configuration.applyConfig: %v", errs[0])
	}
	if cfg.ConfigSecret != nil {
		if cfg.ConfigSecret.Name == "" {
			return errors.New("spec.configuration.configSecret.name can not be empty")
		}
		var secret core.Secret
		if err := w.DefaultClient.Get(ctx, types.NamespacedName{Name: cfg.ConfigSecret.Name, Namespace: req.Namespace}, &secret); err != nil {
			if apierrors.IsNotFound(err) {
				return fmt.Errorf("config secret %s/%s not found", req.Namespace, cfg.ConfigSecret.Name)
			}
			return err
		}
		data := make(map[string]string, len(secret.Data)+len(secret.StringData))
		for k, v := range secret.Data {
			data[k] = string(v)
		}
		for k, v := range secret.StringData {
			data[k] = v
		}
		if errs := dbapi.ValidateYugabyteDBConfigData(data); len(errs) > 0 {
			return fmt.Errorf("config secret %s: %v", cfg.ConfigSecret.Name, errs[0])
		}
	}
	return nil
}

func (w *YugabyteDBOpsRequestCustomWebhook) validateConnectionManager(db *dbapi.YugabyteDB, req *opsapi.YugabyteDBOpsRequest) error {
	cm := req.Spec.ConnectionManager
	if !db.IsYSQLEnabled() {
		return errors.New("spec.connectionManager: YSQL is not enabled on this database")
	}
	if cm.MaxClientConnections != nil && *cm.MaxClientConnections < 1 {
		return errors.New("spec.connectionManager.maxClientConnections must be at least 1")
	}
	if !yugabyteDBOpsPending(req) {
		return nil
	}
	cur := db.IsConnectionManagerEnabled()
	if cm.Enabled == cur && (!cm.Enabled || cm.MaxClientConnections == nil || *cm.MaxClientConnections == db.ConnectionManagerMaxClientConnections()) {
		return errors.New("spec.connectionManager is the same as the current setting")
	}
	return nil
}

func (w *YugabyteDBOpsRequestCustomWebhook) validateReconfigureTLS(db *dbapi.YugabyteDB, req *opsapi.YugabyteDBOpsRequest) error {
	tls := req.Spec.TLS
	if tls == nil {
		return errors.New("spec.tls nil not supported in ReconfigureTLS type")
	}

	change := tls.IssuerRef != nil || len(tls.Certificates) > 0 || tls.NodeToNode != nil || tls.ClientToNode != nil || tls.Webserver != nil
	n := 0
	for _, b := range []bool{tls.Remove, tls.RotateCertificates, change} {
		if b {
			n++
		}
	}
	if n == 0 {
		return errors.New("no reconfiguration is provided in TLS spec")
	}
	if n > 1 {
		return errors.New("more than 1 field have assigned to spec.tls but at a time one is allowed to run one operation")
	}
	if !yugabyteDBOpsPending(req) {
		return nil
	}

	hasTLS := db.Spec.TLS != nil
	if tls.Webserver != nil && *tls.Webserver && (tls.Remove || tls.RotateCertificates) {
		return errors.New("spec.tls.webserver can not be combined with remove or rotateCertificates")
	}
	switch {
	case tls.Remove:
		if !hasTLS {
			return errors.New("spec.tls.remove: the database has no TLS configured")
		}
	case tls.RotateCertificates:
		if !hasTLS {
			return errors.New("spec.tls.rotateCertificates: the database has no TLS configured")
		}
	case !hasTLS:
		// add TLS
		if tls.IssuerRef == nil {
			return errors.New("spec.tls.issuerRef is required to add TLS")
		}
		if tls.NodeToNode != nil && !*tls.NodeToNode && tls.ClientToNode != nil && !*tls.ClientToNode {
			return errors.New("at least one of nodeToNode and clientToNode must be enabled")
		}
	default:
		// issuer change, certificates update or channel toggle
		if tls.Webserver != nil && tls.IssuerRef == nil && len(tls.Certificates) == 0 && tls.NodeToNode == nil && tls.ClientToNode == nil && *tls.Webserver == db.IsWebserverTLS() {
			return errors.New("spec.tls.webserver is the same as the current setting")
		}
		if tls.IssuerRef != nil && db.Spec.TLS.IssuerRef != nil && apiequality.Semantic.DeepEqual(*tls.IssuerRef, *db.Spec.TLS.IssuerRef) &&
			len(tls.Certificates) == 0 && tls.NodeToNode == nil && tls.ClientToNode == nil {
			return errors.New("spec.tls.issuerRef is the same as the current issuer, use rotateCertificates to renew the certificates")
		}
	}
	return nil
}

func (w *YugabyteDBOpsRequestCustomWebhook) validateRotateAuth(ctx context.Context, db *dbapi.YugabyteDB, req *opsapi.YugabyteDBOpsRequest) error {
	if db.Spec.DisableSecurity {
		return errors.New("authentication is disabled on this database (spec.disableSecurity), there is nothing to rotate")
	}
	auth := req.Spec.Authentication
	if auth == nil || auth.SecretRef == nil {
		return nil
	}
	ref := auth.SecretRef
	if ref.Name == "" {
		return errors.New("spec.authentication.secretRef.name can not be empty")
	}
	if yugabyteDBOpsPending(req) && ref.Name == db.GetAuthSecretName() {
		return fmt.Errorf("spec.authentication.secretRef.name %q is the current auth secret, give a new secret or leave it empty to generate a password", ref.Name)
	}
	isVirtual := ref.APIGroup == vsecretapi.GroupName
	data, err := secret_lib.GetData(ctx, w.DefaultClient, req.Namespace, ref.Name, isVirtual)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return fmt.Errorf("referenced secret %s not found", ref.Name)
		}
		return err
	}
	if data == nil {
		return errors.New("spec.authentication.secretRef.name is a valid secret but it does not contain any data")
	}
	user, pass := data[core.BasicAuthUsernameKey], data[core.BasicAuthPasswordKey]
	if len(user) == 0 || len(pass) == 0 {
		return errors.New("spec.authentication.secretRef.name is a valid secret but it does not contain username or password")
	}
	if string(user) != kubedb.YugabyteDBAdminUser {
		return fmt.Errorf("the username in the secret must be %q, the database username cannot be changed", kubedb.YugabyteDBAdminUser)
	}
	return nil
}
