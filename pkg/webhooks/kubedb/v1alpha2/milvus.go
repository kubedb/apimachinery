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

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	"kubedb.dev/apimachinery/apis/kubedb"
	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	amv "kubedb.dev/apimachinery/pkg/validator"

	"gomodules.xyz/x/arrays"
	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ofstv2 "kmodules.xyz/offshoot-api/api/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// SetupMilvusServerWebhookWithManager registers the webhook for Milvus in the manager.
func SetupMilvusWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&olddbapi.Milvus{}).
		WithValidator(&MilvusCustomWebhook{mgr.GetClient()}).
		WithDefaulter(&MilvusCustomWebhook{mgr.GetClient()}).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-kubedb-com-v1alpha2-milvus,mutating=true,failurePolicy=fail,sideEffects=None,groups=kubedb.com,resources=milvuses,verbs=create;update,versions=v1alpha2,name=milvus.kb.io,admissionReviewVersions=v1

// +kubebuilder:object:generate=false
type MilvusCustomWebhook struct {
	DefaultClient client.Client
}

var _ webhook.CustomDefaulter = &MilvusCustomWebhook{}

// log is for logging in this package.
var milvuslog = logf.Log.WithName("milvus-resource")

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (m *MilvusCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	if isDeletionInProgress(obj) {
		return nil
	}
	db, ok := obj.(*olddbapi.Milvus)
	if !ok {
		return fmt.Errorf("expected a Milvus object, got a %T", obj)
	}

	milvuslog.Info("default", "name", db.Name)

	db.SetDefaults(m.DefaultClient)
	return nil
}

var _ webhook.CustomValidator = &MilvusCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (m *MilvusCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	if isDeletionInProgress(obj) {
		return nil, nil
	}
	if err := amv.ValidateNameUniqueness(ctx, m.DefaultClient, obj); err != nil {
		return nil, err
	}
	db, ok := obj.(*olddbapi.Milvus)
	if !ok {
		return nil, fmt.Errorf("expected a Milvus object, got a %T", obj)
	}

	milvuslog.Info("validate create", "name", db.Name)

	warnings, allErr := m.ValidateCreateOrUpdate(db)
	if len(allErr) == 0 {
		return warnings, nil
	}
	return warnings, apierrors.NewInvalid(schema.GroupKind{Group: kubedb.GroupName, Kind: olddbapi.ResourceKindMilvus}, db.Name, allErr)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (m *MilvusCustomWebhook) ValidateUpdate(ctx context.Context, old, newObj runtime.Object) (admission.Warnings, error) {
	if isDeletionInProgress(newObj) {
		return nil, nil
	}
	db, ok := newObj.(*olddbapi.Milvus)
	if !ok {
		return nil, fmt.Errorf("expected a Milvus object, got a %T", newObj)
	}

	milvuslog.Info("validate update", "name", db.Name)

	warnings, allErr := m.ValidateCreateOrUpdate(db)
	if len(allErr) == 0 {
		return warnings, nil
	}

	return warnings, apierrors.NewInvalid(schema.GroupKind{Group: kubedb.GroupName, Kind: olddbapi.ResourceKindMilvus}, db.Name, allErr)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (m *MilvusCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	db, ok := obj.(*olddbapi.Milvus)
	if !ok {
		return nil, fmt.Errorf("expected a Milvus object, got a %T", obj)
	}

	milvuslog.Info("validate delete", "name", db.Name)

	var allErr field.ErrorList
	if db.Spec.DeletionPolicy == olddbapi.DeletionPolicyDoNotTerminate {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("terminationPolicy"),
			db.Name,
			"Can not delete as terminationPolicy is set to \"DoNotTerminate\""))
		return nil, apierrors.NewInvalid(schema.GroupKind{Group: kubedb.GroupName, Kind: olddbapi.ResourceKindMilvus}, db.Name, allErr)
	}
	return nil, nil
}

func (m *MilvusCustomWebhook) ValidateCreateOrUpdate(db *olddbapi.Milvus) (admission.Warnings, field.ErrorList) {
	var allErr field.ErrorList
	var warnings admission.Warnings

	milvusVersion, err := m.milvusValidateVersion(db)
	if err != nil {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("version"),
			db.Name,
			err.Error()))
	} else if err := milvusValidateGPU(db, milvusVersion); err != nil {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("gpu"),
			db.Name,
			err.Error()))
	}

	warnings = append(warnings, milvusWarnSRIOVTopology(db)...)

	if db.Spec.PodTemplate != nil {
		if err = ValidateMilvusEnvVar(getMilvusContainerEnvs(db), forbiddenMilvusEnvVars, db.ResourceKind()); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("podTemplate"),
				db.Name,
				err.Error()))
		}
	}

	err = milvusValidateVolumes(db.Spec.PodTemplate)
	if err != nil {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("podTemplate").Child("spec").Child("volumes"),
			db.Name,
			err.Error()))
	}

	err = milvusValidateVolumesMountPaths(db.Spec.PodTemplate)
	if err != nil {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("podTemplate").Child("spec").Child("containers"),
			db.Name,
			err.Error()))
	}

	if db.Spec.StorageType == "" {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("storageType"),
			db.Name,
			"StorageType can not be empty"))
	} else {
		if db.Spec.StorageType != olddbapi.StorageTypeDurable && db.Spec.StorageType != olddbapi.StorageTypeEphemeral {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("storageType"),
				db.Name,
				"StorageType should be either durable or ephemeral"))
		}
	}

	if monitorSpec := db.Spec.Monitor; monitorSpec != nil {
		if err := amv.ValidateMonitorSpec(monitorSpec); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("monitor"), db.Name, err.Error()))
		}
	}

	if db.Spec.TLS != nil {
		if db.Spec.TLS.IssuerRef == nil {
			allErr = append(allErr, field.Invalid(
				field.NewPath("spec").Child("tls").Child("issuerRef"),
				db.Name,
				"spec.tls.issuerRef is required when TLS is configured",
			))
		}

		if db.Spec.TLS.Internal != nil && db.Spec.TLS.Internal.Mode == olddbapi.TLSModeMTLS {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("tls").Child("internal").Child("mode"),
				db.Name, "spec.tls.internal.mTLS is not allowed"))
		}
	}

	if meta := db.Spec.MetaStorage; meta != nil {
		if meta.ExternallyManaged {
			if meta.TLS != nil {
				allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("metaStorage").Child("tls"),
					db.Name, "spec.metaStorage.tls is only used when metaStorage is not externally managed"))
			}
			if meta.AuthSecret != nil {
				allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("metaStorage").Child("authSecret"),
					db.Name, "spec.metaStorage.authSecret is only used when metaStorage is not externally managed"))
			}
		} else if meta.TLS != nil && meta.TLS.IssuerRef == nil {
			allErr = append(allErr, field.Invalid(
				field.NewPath("spec").Child("metaStorage").Child("tls").Child("issuerRef"),
				db.Name,
				"spec.metaStorage.tls.issuerRef is required when metaStorage.tls is configured",
			))
		}
	}

	if len(allErr) == 0 {
		return warnings, nil
	}

	return warnings, allErr
}

// reserved volume and volumes mounts for milvus
var milvusReservedVolumes = []string{
	kubedb.MilvusVolumeNameData,
	kubedb.MilvusConfigVolName,
	kubedb.MilvusConfigFileName,
}

var milvusReservedVolumesMountPaths = []string{
	kubedb.MilvusDataDir,
	kubedb.MilvusConfigVolDir,
}

func (m *MilvusCustomWebhook) milvusValidateVersion(db *olddbapi.Milvus) (*catalog.MilvusVersion, error) {
	var milvusVersion catalog.MilvusVersion

	err := m.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name: db.Spec.Version,
	}, &milvusVersion)
	if err != nil {
		return nil, err
	}
	return &milvusVersion, nil
}

// milvusValidateGPU rejects a Milvus CR that requests a GPU (via the typed
// spec.gpu/spec.topology.distributed.<role>.gpu field, or a hand-written
// nvidia.com/gpu resource request/limit on any podTemplate container)
// against a MilvusVersion that isn't declared GPU-capable. Runs on both
// CREATE and UPDATE, so downgrading spec.version out from under a live GPU
// request is caught too.
func milvusValidateGPU(db *olddbapi.Milvus, milvusVersion *catalog.MilvusVersion) error {
	if !db.RequestsGPU() {
		return nil
	}
	if milvusVersion.Spec.DB.GPU != nil && milvusVersion.Spec.DB.GPU.Supported {
		return nil
	}
	return fmt.Errorf(
		"a GPU resource is requested (spec.gpu, a distributed role's .gpu, or a hand-written "+
			"nvidia.com/gpu podTemplate resource) but MilvusVersion %q does not declare "+
			"spec.db.gpu.supported: true; use a GPU-capable MilvusVersion instead",
		db.Spec.Version)
}

// milvusWarnSRIOVTopology warns, rather than rejects, if spec.network.sriov
// is set on some but not all five Distributed roles. Every Distributed role
// dials, or is dialed by, at least one other role directly by its
// etcd-advertised address (never through a Service), so a role left off
// network.sriov may become unreachable from, or unable to reach, whichever
// roles do have it once their advertise-IP patch applies. This can't be a
// hard rejection: the webhook has no way to confirm the customer's actual
// NAD/subnet is (or isn't) routable from a role without network.sriov.
func milvusWarnSRIOVTopology(db *olddbapi.Milvus) admission.Warnings {
	if !db.IsDistributed() {
		return nil
	}
	withSRIOV := db.DistributedNodeRolesWithSRIOV()
	if len(withSRIOV) == 0 || len(withSRIOV) == 5 {
		return nil
	}

	allRoles := []olddbapi.MilvusNodeRoleType{
		olddbapi.MilvusNodeRoleMixCoord, olddbapi.MilvusNodeRoleDataNode, olddbapi.MilvusNodeRoleProxy,
		olddbapi.MilvusNodeRoleQueryNode, olddbapi.MilvusNodeRoleStreamingNode,
	}
	withSet := make(map[olddbapi.MilvusNodeRoleType]bool, len(withSRIOV))
	for _, r := range withSRIOV {
		withSet[r] = true
	}
	var missing []string
	for _, r := range allRoles {
		if !withSet[r] {
			missing = append(missing, string(r))
		}
	}

	return admission.Warnings{fmt.Sprintf(
		"spec.network.sriov is set on some Distributed roles but not %v; every role dials, or is "+
			"dialed by, at least one other role directly by its advertised address, so these roles may "+
			"be unable to reach, or be reached by, the roles that do have it once the advertise-IP patch "+
			"applies. Set spec.network.sriov on all five Distributed roles (mixcoord, datanode, proxy, "+
			"querynode, streamingnode), or none.", missing,
	)}
}

func milvusValidateVolumes(podTemplate *ofstv2.PodTemplateSpec) error {
	if podTemplate == nil {
		return nil
	}
	if podTemplate.Spec.Volumes == nil {
		return nil
	}

	for _, rv := range milvusReservedVolumes {
		for _, ugv := range podTemplate.Spec.Volumes {
			if ugv.Name == rv {
				return errors.New("Can't use a reserved volume name: " + rv)
			}
		}
	}

	return nil
}

func milvusValidateVolumesMountPaths(podTemplate *ofstv2.PodTemplateSpec) error {
	if podTemplate == nil {
		return nil
	}

	if podTemplate.Spec.Containers != nil {
		// Check container volume mounts
		for _, rvmp := range milvusReservedVolumesMountPaths {
			containerList := podTemplate.Spec.Containers
			for i := range containerList {
				mountPathList := containerList[i].VolumeMounts
				for j := range mountPathList {
					if mountPathList[j].MountPath == rvmp {
						return errors.New("Can't use a reserve volume mount path name: " + rvmp)
					}
				}
			}
		}
	}

	return nil
}

var forbiddenMilvusEnvVars = []string{
	kubedb.MinioAddressName,
	kubedb.MinioAddressKey,
	kubedb.MinioAccessKeyName,
	kubedb.MinioAccessKey,
	kubedb.MinioSecretKeyName,
	kubedb.MinioSecretKey,
	kubedb.MinioBucketName,
	kubedb.MinioBucketKey,
	kubedb.MinioPortKey,
	kubedb.MinioPortName,
	kubedb.EtcdEndpointsName,
}

func getMilvusContainerEnvs(db *olddbapi.Milvus) []core.EnvVar {
	for _, container := range db.Spec.PodTemplate.Spec.Containers {
		if container.Name == kubedb.MilvusContainerName {
			return container.Env
		}
	}
	return []core.EnvVar{}
}

func ValidateMilvusEnvVar(envs []core.EnvVar, forbiddenEnvs []string, resourceType string) error {
	for _, env := range envs {
		present, _ := arrays.Contains(forbiddenEnvs, env.Name)
		if present {
			return fmt.Errorf("environment variable %s is forbidden to use in %s spec", env.Name, resourceType)
		}
	}
	return nil
}
