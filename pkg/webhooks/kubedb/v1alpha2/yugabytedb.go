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
	"errors"
	"fmt"
	"strings"

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	"kubedb.dev/apimachinery/apis/kubedb"
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"
	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	amv "kubedb.dev/apimachinery/pkg/validator"

	core "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	kmc "kmodules.xyz/client-go/conditions"
	psapi "kubeops.dev/petset/apis/apps/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// SetupYugabyteDBWebhookWithManager registers the webhook for YugabyteDB in the manager.
func SetupYugabyteDBWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&olddbapi.YugabyteDB{}).
		WithValidator(amv.WithServiceTemplateAliasWarnings(&YugabyteDBCustomWebhook{DefaultClient: mgr.GetClient()}, dbapi.PrimaryServiceAlias, dbapi.StatsServiceAlias)).
		WithDefaulter(&YugabyteDBCustomWebhook{DefaultClient: mgr.GetClient()}).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-kubedb-com-v1alpha2-yugabytedb,mutating=true,failurePolicy=fail,sideEffects=None,groups=kubedb.com,resources=yugabytedbs,verbs=create;update,versions=v1alpha2,name=myugabytedb.kb.io,admissionReviewVersions={v1,v1beta1}

// +kubebuilder:object:generate=false
type YugabyteDBCustomWebhook struct {
	DefaultClient client.Client
}

var (
	_ webhook.CustomDefaulter = &YugabyteDBCustomWebhook{}
	_ webhook.CustomValidator = &YugabyteDBCustomWebhook{}
)

// log is for logging in this package.
var yugabyteDBLog = logf.Log.WithName("yugabytedb-resource")

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type
func (w *YugabyteDBCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	if isDeletionInProgress(obj) {
		return nil
	}
	db, ok := obj.(*olddbapi.YugabyteDB)
	if !ok {
		return fmt.Errorf("expected a YugabyteDB object but got %T", obj)
	}
	yugabyteDBLog.Info("default", "name", db.GetName())

	// SetDefaults forces the master replicas to the replication factor. A different value set by the user
	// is a mistake and is rejected instead of being overwritten silently.
	if t := db.Spec.Topology; t != nil && t.Master != nil && t.Master.Replicas != nil && *t.Master.Replicas != db.ReplicationFactor() {
		return fmt.Errorf("spec.topology.master.replicas must equal spec.replicationFactor (%d), got %d", db.ReplicationFactor(), *t.Master.Replicas)
	}

	if db.Spec.Halted {
		if db.Spec.DeletionPolicy == olddbapi.DeletionPolicyDoNotTerminate {
			return errors.New(`can't halt, since deletion policy is 'DoNotTerminate'`)
		}
		db.Spec.DeletionPolicy = olddbapi.DeletionPolicyHalt
	}

	db.SetDefaults(w.DefaultClient)
	return nil
}

//+kubebuilder:webhook:path=/validate-kubedb-com-v1alpha2-yugabytedb,mutating=false,failurePolicy=fail,sideEffects=None,groups=kubedb.com,resources=yugabytedbs,verbs=create;update;delete,versions=v1alpha2,name=vyugabytedb.kb.io,admissionReviewVersions={v1,v1beta1}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type
func (w *YugabyteDBCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	if isDeletionInProgress(obj) {
		return nil, nil
	}
	if err := amv.ValidateNameUniqueness(ctx, w.DefaultClient, obj); err != nil {
		return nil, err
	}
	db, ok := obj.(*olddbapi.YugabyteDB)
	if !ok {
		return nil, fmt.Errorf("expected a YugabyteDB object but got %T", obj)
	}
	yugabyteDBLog.Info("validate create", "name", db.GetName())
	return w.validate(ctx, db, nil)
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type
func (w *YugabyteDBCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	if isDeletionInProgress(newObj) {
		return nil, nil
	}
	db, ok := newObj.(*olddbapi.YugabyteDB)
	if !ok {
		return nil, fmt.Errorf("expected a YugabyteDB object but got %T", newObj)
	}
	old, ok := oldObj.(*olddbapi.YugabyteDB)
	if !ok {
		return nil, fmt.Errorf("expected a YugabyteDB object but got %T", oldObj)
	}
	yugabyteDBLog.Info("validate update", "name", db.GetName())

	var allErr field.ErrorList
	allErr = append(allErr, validateYugabyteDBImmutable(old, db)...)
	if !isYugabyteDBOperatorRequest(ctx) {
		allErr = append(allErr, validateYugabyteDBOpsOwned(old, db)...)
	}
	if len(allErr) > 0 {
		return nil, yugabyteDBInvalid(db.Name, allErr)
	}
	return w.validate(ctx, db, old)
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type
func (w *YugabyteDBCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	db, ok := obj.(*olddbapi.YugabyteDB)
	if !ok {
		return nil, fmt.Errorf("expected a YugabyteDB object but got %T", obj)
	}
	yugabyteDBLog.Info("validate delete", "name", db.GetName())

	if db.Spec.DeletionPolicy == olddbapi.DeletionPolicyDoNotTerminate {
		allErr := field.ErrorList{field.Invalid(field.NewPath("spec").Child("deletionPolicy"),
			db.GetName(),
			"Can not delete as deletionPolicy is set to \"DoNotTerminate\"")}
		return nil, yugabyteDBInvalid(db.GetName(), allErr)
	}
	return nil, nil
}

func yugabyteDBInvalid(name string, allErr field.ErrorList) error {
	return apierrors.NewInvalid(schema.GroupKind{Group: kubedb.GroupName, Kind: olddbapi.ResourceKindYugabyteDB}, name, allErr)
}

// isYugabyteDBOperatorRequest reports whether the request comes from a service account, which is how the
// KubeDB provisioner and ops-manager write the spec. Users and CI systems that edit a YugabyteDB use their own
// identity. A request without admission information is not trusted.
func isYugabyteDBOperatorRequest(ctx context.Context) bool {
	req, err := admission.RequestFromContext(ctx)
	if err != nil {
		return false
	}
	return strings.HasPrefix(req.UserInfo.Username, "system:serviceaccount:")
}

func yugabyteDBStorageType(n *olddbapi.YugabyteDBNode) olddbapi.StorageType {
	if n.StorageType == "" {
		return olddbapi.StorageTypeDurable
	}
	return n.StorageType
}

// validateYugabyteDBImmutable rejects the changes that can never be made after creation.
func validateYugabyteDBImmutable(old, db *olddbapi.YugabyteDB) field.ErrorList {
	var allErr field.ErrorList
	spec := field.NewPath("spec")
	if old.ReplicationFactor() != db.ReplicationFactor() {
		allErr = append(allErr, field.Forbidden(spec.Child("replicationFactor"), "replicationFactor is immutable"))
	}
	if old.IsYSQLEnabled() != db.IsYSQLEnabled() {
		allErr = append(allErr, field.Forbidden(spec.Child("ysql", "enabled"), "ysql.enabled is immutable"))
	}
	if old.IsYCQLEnabled() != db.IsYCQLEnabled() {
		allErr = append(allErr, field.Forbidden(spec.Child("ycql", "enabled"), "ycql.enabled is immutable"))
	}
	if old.Spec.DisableSecurity != db.Spec.DisableSecurity {
		allErr = append(allErr, field.Forbidden(spec.Child("disableSecurity"), "disableSecurity is immutable"))
	}
	if old.IsReadReplicaEnabled() && db.IsReadReplicaEnabled() && old.ReadReplicaPlacementUUID() != db.ReadReplicaPlacementUUID() {
		allErr = append(allErr, field.Forbidden(spec.Child("topology", "readReplica", "placementUUID"), "placementUUID is immutable"))
	}
	for _, role := range db.Roles() {
		p := spec.Child("topology", string(role))
		if old.DataVolumes(role) != db.DataVolumes(role) {
			allErr = append(allErr, field.Forbidden(p.Child("dataVolumes"), "dataVolumes is immutable"))
		}
		if yugabyteDBStorageType(old.Node(role)) != yugabyteDBStorageType(db.Node(role)) {
			allErr = append(allErr, field.Forbidden(p.Child("storageType"), "storageType is immutable"))
		}
	}
	return allErr
}

func yugabyteDBStorageSize(n *olddbapi.YugabyteDBNode) string {
	if n.Storage == nil {
		return ""
	}
	q := n.Storage.Resources.Requests[core.ResourceStorage]
	return q.String()
}

func yugabyteDBContainerResources(db *olddbapi.YugabyteDB, role olddbapi.YugabyteDBRole) core.ResourceRequirements {
	for _, c := range db.Node(role).PodTemplate.Spec.Containers {
		if c.Name == db.ContainerName(role) {
			return c.Resources
		}
	}
	return core.ResourceRequirements{}
}

func yugabyteDBConfiguration(c *olddbapi.ConfigurationSpec) olddbapi.ConfigurationSpec {
	if c == nil {
		return olddbapi.ConfigurationSpec{}
	}
	out := *c
	if len(out.Inline) == 0 {
		out.Inline = nil
	}
	return out
}

// validateYugabyteDBOpsOwned rejects direct edits of the fields that an OpsRequest owns, once the database is
// provisioned. Before that, the user is still writing the initial spec.
func validateYugabyteDBOpsOwned(old, db *olddbapi.YugabyteDB) field.ErrorList {
	if !kmc.IsConditionTrue(old.Status.Conditions, kubedb.DatabaseProvisioned) {
		return nil
	}
	const hint = "is owned by a YugabyteDBOpsRequest and cannot be edited directly after provisioning"
	var allErr field.ErrorList
	spec := field.NewPath("spec")
	if old.Spec.Version != db.Spec.Version {
		allErr = append(allErr, field.Forbidden(spec.Child("version"), "version "+hint+" (UpdateVersion)"))
	}
	if old.Replicas(olddbapi.YugabyteDBRoleTServer) != db.Replicas(olddbapi.YugabyteDBRoleTServer) {
		allErr = append(allErr, field.Forbidden(spec.Child("topology", "tserver", "replicas"), "replicas "+hint+" (HorizontalScaling)"))
	}
	if old.IsReadReplicaEnabled() && db.IsReadReplicaEnabled() && old.Replicas(olddbapi.YugabyteDBRoleReadReplica) != db.Replicas(olddbapi.YugabyteDBRoleReadReplica) {
		allErr = append(allErr, field.Forbidden(spec.Child("topology", "readReplica", "replicas"), "replicas "+hint+" (HorizontalScaling)"))
	}
	for _, role := range db.Roles() {
		p := spec.Child("topology", string(role))
		if !apiequality.Semantic.DeepEqual(yugabyteDBContainerResources(old, role), yugabyteDBContainerResources(db, role)) {
			allErr = append(allErr, field.Forbidden(p.Child("podTemplate", "spec", "containers", "resources"), "resources "+hint+" (VerticalScaling)"))
		}
		if yugabyteDBStorageSize(old.Node(role)) != yugabyteDBStorageSize(db.Node(role)) {
			allErr = append(allErr, field.Forbidden(p.Child("storage", "resources", "requests", "storage"), "storage size "+hint+" (VolumeExpansion)"))
		}
	}
	oc, nc := yugabyteDBConfiguration(old.Spec.Configuration), yugabyteDBConfiguration(db.Spec.Configuration)
	if !apiequality.Semantic.DeepEqual(oc, nc) {
		allErr = append(allErr, field.Forbidden(spec.Child("configuration"), "configuration "+hint+" (Reconfigure)"))
	}
	if !apiequality.Semantic.DeepEqual(old.Spec.TLS, db.Spec.TLS) {
		allErr = append(allErr, field.Forbidden(spec.Child("tls"), "tls "+hint+" (ReconfigureTLS)"))
	}
	if old.IsConnectionManagerEnabled() != db.IsConnectionManagerEnabled() ||
		(old.IsConnectionManagerEnabled() && old.ConnectionManagerMaxClientConnections() != db.ConnectionManagerMaxClientConnections()) {
		allErr = append(allErr, field.Forbidden(spec.Child("ysql", "connectionManager"), "connectionManager "+hint+" (Reconfigure)"))
	}
	if old.GetAuthSecretName() != db.GetAuthSecretName() {
		allErr = append(allErr, field.Forbidden(spec.Child("authSecret"), "authSecret "+hint+" (RotateAuth)"))
	}
	return allErr
}

var yugabyteDBReplicationFactors = map[int32]bool{1: true, 3: true, 5: true, 7: true}

// validate checks a YugabyteDB. old is nil on create.
func (w *YugabyteDBCustomWebhook) validate(ctx context.Context, db, old *olddbapi.YugabyteDB) (admission.Warnings, error) {
	var allErr field.ErrorList
	var warnings admission.Warnings
	spec := field.NewPath("spec")
	rf := db.ReplicationFactor()

	if db.Spec.Version == "" {
		allErr = append(allErr, field.Required(spec.Child("version"), "spec.version is missing"))
	} else if old == nil || old.Spec.Version != db.Spec.Version {
		if err := w.validateVersion(ctx, db); err != nil {
			allErr = append(allErr, field.Invalid(spec.Child("version"), db.Spec.Version, err.Error()))
		}
	}

	if !yugabyteDBReplicationFactors[rf] {
		allErr = append(allErr, field.NotSupported(spec.Child("replicationFactor"), rf, []string{"1", "3", "5", "7"}))
	}

	if t := db.Spec.Topology; t != nil {
		if t.Master != nil && t.Master.Replicas != nil && *t.Master.Replicas != rf {
			allErr = append(allErr, field.Invalid(spec.Child("topology", "master", "replicas"), *t.Master.Replicas,
				fmt.Sprintf("master replicas must equal replicationFactor (%d)", rf)))
		}
		if t.TServer != nil && t.TServer.Replicas != nil && *t.TServer.Replicas < rf {
			allErr = append(allErr, field.Invalid(spec.Child("topology", "tserver", "replicas"), *t.TServer.Replicas,
				fmt.Sprintf("tserver replicas must be at least replicationFactor (%d)", rf)))
		}
	}

	if rr := db.Spec.Topology; rr != nil && rr.ReadReplica != nil {
		allErr = append(allErr, validateYugabyteDBReadReplica(db)...)
	}

	for _, role := range db.Roles() {
		p := spec.Child("topology", string(role))
		node := db.Node(role)

		switch yugabyteDBStorageType(node) {
		case olddbapi.StorageTypeDurable:
			if node.Storage == nil {
				allErr = append(allErr, field.Required(p.Child("storage"), "storage is required when storageType is Durable"))
			}
		case olddbapi.StorageTypeEphemeral:
		default:
			allErr = append(allErr, field.NotSupported(p.Child("storageType"), node.StorageType,
				[]string{string(olddbapi.StorageTypeDurable), string(olddbapi.StorageTypeEphemeral)}))
		}
		if node.DataVolumes != nil && (*node.DataVolumes < 1 || *node.DataVolumes > 8) {
			allErr = append(allErr, field.Invalid(p.Child("dataVolumes"), *node.DataVolumes, "dataVolumes must be between 1 and 8"))
		}

		if mem, ok := yugabyteDBContainerResources(db, role).Limits[core.ResourceMemory]; !ok || mem.IsZero() {
			allErr = append(allErr, field.Required(p.Child("podTemplate", "spec", "containers").Key(db.ContainerName(role)).Child("resources", "limits", "memory"),
				"limits.memory must be set, the memory limit of the process is derived from it"))
		}

		if ppp := node.PodTemplate.Spec.PodPlacementPolicy; ppp != nil {
			warns, err := w.validatePlacementPolicy(ctx, ppp.Name, rf)
			warnings = append(warnings, warns...)
			if err != nil {
				allErr = append(allErr, field.Invalid(p.Child("podTemplate", "spec", "podPlacementPolicy"), ppp.Name, err.Error()))
			}
		}
	}

	if db.Spec.DeletionPolicy == "" {
		allErr = append(allErr, field.Required(spec.Child("deletionPolicy"), "spec.deletionPolicy is missing"))
	}
	if db.Spec.Halted && db.Spec.DeletionPolicy == olddbapi.DeletionPolicyDoNotTerminate {
		allErr = append(allErr, field.Invalid(spec.Child("halted"), true, "can't halt, since deletionPolicy is DoNotTerminate"))
	}
	if db.Spec.Halted && db.Spec.DeletionPolicy != "" && db.Spec.DeletionPolicy != olddbapi.DeletionPolicyHalt && db.Spec.DeletionPolicy != olddbapi.DeletionPolicyDoNotTerminate {
		allErr = append(allErr, field.Invalid(spec.Child("halted"), true, "halted: true needs deletionPolicy Halt"))
	}
	if db.Spec.DeletionPolicy == olddbapi.DeletionPolicyHalt {
		for _, role := range db.Roles() {
			if yugabyteDBStorageType(db.Node(role)) == olddbapi.StorageTypeEphemeral {
				allErr = append(allErr, field.Invalid(spec.Child("deletionPolicy"), db.Spec.DeletionPolicy, "deletionPolicy Halt can not be used with Ephemeral storage"))
				break
			}
		}
	}

	if db.Spec.AuthSecret != nil && db.Spec.AuthSecret.ExternallyManaged && db.Spec.AuthSecret.Name == "" {
		allErr = append(allErr, field.Required(spec.Child("authSecret", "name"), `for an externallyManaged auth secret, spec.authSecret.name must be set`))
	}

	if db.Spec.Configuration != nil {
		allErr = append(allErr, w.validateConfiguration(ctx, db)...)
	}

	if tls := db.Spec.TLS; tls != nil {
		if tls.IssuerRef == nil {
			allErr = append(allErr, field.Required(spec.Child("tls", "issuerRef"), "spec.tls.issuerRef is missing"))
		}
		if tls.NodeToNode != nil && !*tls.NodeToNode && tls.ClientToNode != nil && !*tls.ClientToNode {
			allErr = append(allErr, field.Invalid(spec.Child("tls"), tls, "at least one of nodeToNode and clientToNode must be enabled"))
		}
	}

	if tls := db.Spec.TLS; tls != nil && tls.Webserver != nil && *tls.Webserver && tls.IssuerRef == nil {
		allErr = append(allErr, field.Invalid(spec.Child("tls", "webserver"), true, "webserver TLS needs spec.tls.issuerRef"))
	}
	if cm := db.Spec.YSQL; cm != nil && cm.ConnectionManager != nil {
		cp := spec.Child("ysql", "connectionManager")
		if cm.ConnectionManager.Enabled && !db.IsYSQLEnabled() {
			allErr = append(allErr, field.Invalid(cp.Child("enabled"), true, "the Connection Manager needs spec.ysql.enabled"))
		}
		if m := cm.ConnectionManager.MaxClientConnections; m != nil && *m < 1 {
			allErr = append(allErr, field.Invalid(cp.Child("maxClientConnections"), *m, "maxClientConnections must be at least 1"))
		}
	}

	allErr = append(allErr, validateYugabyteDBPITR(db)...)

	if pl := db.Spec.Placement; pl != nil {
		seen := map[string]bool{}
		for i, z := range pl.PreferredZones {
			if z == "" || seen[z] {
				allErr = append(allErr, field.Invalid(spec.Child("placement", "preferredZones").Index(i), z, "zones must be non-empty and unique"))
			}
			seen[z] = true
		}
	}

	if db.Spec.Monitor != nil {
		if err := amv.ValidateMonitorSpec(db.Spec.Monitor); err != nil {
			allErr = append(allErr, field.Invalid(spec.Child("monitor"), db.Name, err.Error()))
		}
	}
	if err := amv.ValidateHealth(&db.Spec.HealthChecker); err != nil {
		allErr = append(allErr, field.Invalid(spec.Child("healthChecker"), db.Name, err.Error()))
	}

	if len(allErr) > 0 {
		return warnings, yugabyteDBInvalid(db.Name, allErr)
	}
	return warnings, nil
}

// validateYugabyteDBPITR checks spec.pitr. YugabyteDB allows one snapshot schedule for each database.
func validateYugabyteDBPITR(db *olddbapi.YugabyteDB) field.ErrorList {
	var allErr field.ErrorList
	seen := map[string]bool{}
	for i, p := range db.Spec.PITR {
		path := field.NewPath("spec", "pitr").Index(i)
		if p.Database == "" {
			allErr = append(allErr, field.Required(path.Child("database"), "database is required"))
		} else if seen[p.Database] {
			allErr = append(allErr, field.Duplicate(path.Child("database"), p.Database))
		}
		seen[p.Database] = true
		switch p.GetKind() {
		case olddbapi.YugabyteDBPITRKindYSQL:
			if !db.IsYSQLEnabled() {
				allErr = append(allErr, field.Invalid(path.Child("kind"), p.Kind, "YSQL is not enabled"))
			}
		case olddbapi.YugabyteDBPITRKindYCQL:
			if !db.IsYCQLEnabled() {
				allErr = append(allErr, field.Invalid(path.Child("kind"), p.Kind, "YCQL is not enabled"))
			}
		default:
			allErr = append(allErr, field.NotSupported(path.Child("kind"), p.Kind, []string{string(olddbapi.YugabyteDBPITRKindYSQL), string(olddbapi.YugabyteDBPITRKindYCQL)}))
		}
		if p.IntervalMinutes < 1 {
			allErr = append(allErr, field.Invalid(path.Child("intervalMinutes"), p.IntervalMinutes, "intervalMinutes must be at least 1"))
		}
		if p.RetentionMinutes < p.IntervalMinutes || p.RetentionMinutes < 1 {
			allErr = append(allErr, field.Invalid(path.Child("retentionMinutes"), p.RetentionMinutes, "retentionMinutes must be at least intervalMinutes"))
		}
	}
	return allErr
}

// validateYugabyteDBReadReplica checks spec.topology.readReplica.
func validateYugabyteDBReadReplica(db *olddbapi.YugabyteDB) field.ErrorList {
	var allErr field.ErrorList
	p := field.NewPath("spec", "topology", "readReplica")
	rr := db.Spec.Topology.ReadReplica
	if rr.ReplicationFactor != nil && *rr.ReplicationFactor < 1 {
		allErr = append(allErr, field.Invalid(p.Child("replicationFactor"), *rr.ReplicationFactor, "the read replica replicationFactor must be at least 1"))
	}
	if rr.Replicas != nil {
		n, rf := *rr.Replicas, db.ReadReplicaReplicationFactor()
		switch {
		case n < 0:
			allErr = append(allErr, field.Invalid(p.Child("replicas"), n, "replicas must not be negative"))
		case n != 0 && n < rf && rf >= 1:
			allErr = append(allErr, field.Invalid(p.Child("replicas"), n,
				fmt.Sprintf("read replica replicas must be 0 or at least the read replica replicationFactor (%d)", rf)))
		}
	}
	if rr.PlacementUUID != "" {
		if errs := validation.IsDNS1123Label(rr.PlacementUUID); len(errs) > 0 {
			allErr = append(allErr, field.Invalid(p.Child("placementUUID"), rr.PlacementUUID, strings.Join(errs, "; ")))
		}
	}
	return allErr
}

func (w *YugabyteDBCustomWebhook) validateVersion(ctx context.Context, db *olddbapi.YugabyteDB) error {
	var v catalog.YugabyteDBVersion
	if err := w.DefaultClient.Get(ctx, types.NamespacedName{Name: db.Spec.Version}, &v); err != nil {
		return fmt.Errorf("version %q is not available in the catalog: %v", db.Spec.Version, err)
	}
	if v.Spec.Deprecated {
		return fmt.Errorf("version %q is deprecated", db.Spec.Version)
	}
	if err := v.ValidateSpecs(); err != nil {
		return err
	}
	return nil
}

// validatePlacementPolicy checks that the PlacementPolicy exists. A replicated database whose policy does not
// spread the pods across zones with DoNotSchedule only gets a warning, since the policy may be fixed later.
func (w *YugabyteDBCustomWebhook) validatePlacementPolicy(ctx context.Context, name string, rf int32) (admission.Warnings, error) {
	var pp psapi.PlacementPolicy
	if err := w.DefaultClient.Get(ctx, client.ObjectKey{Name: name}, &pp); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("PlacementPolicy %q not found", name)
		}
		return nil, err
	}
	if rf > 1 {
		z := pp.Spec.ZoneSpreadConstraint
		if z == nil || z.WhenUnsatisfiable != core.DoNotSchedule {
			return admission.Warnings{fmt.Sprintf("PlacementPolicy %q does not spread the pods across zones with whenUnsatisfiable: DoNotSchedule, a zone failure may take down more than one replica of a tablet", name)}, nil
		}
	}
	return nil, nil
}

func (w *YugabyteDBCustomWebhook) validateConfiguration(ctx context.Context, db *olddbapi.YugabyteDB) field.ErrorList {
	var allErr field.ErrorList
	p := field.NewPath("spec", "configuration")
	if name := db.Spec.Configuration.SecretName; name != "" {
		var secret core.Secret
		if err := w.DefaultClient.Get(ctx, client.ObjectKey{Name: name, Namespace: db.Namespace}, &secret); err != nil {
			allErr = append(allErr, field.Invalid(p.Child("secretName"), name, "failed to get the config secret: "+err.Error()))
		} else {
			data := make(map[string]string, len(secret.Data)+len(secret.StringData))
			for k, v := range secret.Data {
				data[k] = string(v)
			}
			for k, v := range secret.StringData {
				data[k] = v
			}
			for _, err := range olddbapi.ValidateYugabyteDBConfigData(data) {
				allErr = append(allErr, field.Invalid(p.Child("secretName"), name, err.Error()))
			}
		}
	}
	for _, err := range olddbapi.ValidateYugabyteDBConfigData(db.Spec.Configuration.Inline) {
		allErr = append(allErr, field.Invalid(p.Child("inline"), db.Name, err.Error()))
	}
	return allErr
}
