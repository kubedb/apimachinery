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
	"unsafe"

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	"kubedb.dev/apimachinery/apis/kubedb"
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"
	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	amv "kubedb.dev/apimachinery/pkg/validator"

	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
	ofst "kmodules.xyz/offshoot-api/api/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// SetupNeo4jWebhookWithManager registers the webhook for Neo4j in the manager.
func SetupNeo4jWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&olddbapi.Neo4j{}).
		WithValidator(&Neo4jCustomWebhook{mgr.GetClient()}).
		WithDefaulter(&Neo4jCustomWebhook{mgr.GetClient()}).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-Neo4j-kubedb-com-v1alpha1-Neo4j,mutating=true,failurePolicy=fail,sideEffects=None,groups=kubedb.com,resources=Neo4js,verbs=create;update,versions=v1alpha1,name=mNeo4j.kb.io,admissionReviewVersions={v1,v1beta1}

// +kubebuilder:object:generate=false
type Neo4jCustomWebhook struct {
	DefaultClient client.Client
}

var _ webhook.CustomDefaulter = &Neo4jCustomWebhook{}

// log is for logging in this package.
var Neo4jlog = logf.Log.WithName("Neo4j-resource")

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (w *Neo4jCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	if isDeletionInProgress(obj) {
		return nil
	}
	db, ok := obj.(*olddbapi.Neo4j)
	if !ok {
		return fmt.Errorf("expected an Neo4j object but got %T", obj)
	}

	Neo4jlog.Info("default", "name", db.GetName())
	db.SetDefaults(w.DefaultClient)
	return nil
}

//+kubebuilder:webhook:path=/validate-Neo4j-kubedb-com-v1alpha1-Neo4j,mutating=false,failurePolicy=fail,sideEffects=None,groups=kubedb.com,resources=Neo4js,verbs=create;update,versions=v1alpha1,name=vNeo4j.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomValidator = &Neo4jCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *Neo4jCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	if isDeletionInProgress(obj) {
		return nil, nil
	}
	if err := amv.ValidateNameUniqueness(ctx, w.DefaultClient, obj); err != nil {
		return nil, err
	}
	db, ok := obj.(*olddbapi.Neo4j)
	if !ok {
		return nil, fmt.Errorf("expected an Neo4j object but got %T", obj)
	}
	Neo4jlog.Info("validate create", "name", db.GetName())
	return nil, w.ValidateCreateOrUpdate(db)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *Neo4jCustomWebhook) ValidateUpdate(ctx context.Context, old, newObj runtime.Object) (admission.Warnings, error) {
	if isDeletionInProgress(newObj) {
		return nil, nil
	}
	db, ok := newObj.(*olddbapi.Neo4j)
	if !ok {
		return nil, fmt.Errorf("expected an Neo4j object but got %T", newObj)
	}

	Neo4jlog.Info("validate update", "name", db.GetName())
	return nil, w.ValidateCreateOrUpdate(db)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (w *Neo4jCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	db, ok := obj.(*olddbapi.Neo4j)
	if !ok {
		return nil, fmt.Errorf("expected an Neo4j object but got %T", obj)
	}
	Neo4jlog.Info("validate delete", "name", db.GetName())

	var allErr field.ErrorList
	if db.Spec.DeletionPolicy == olddbapi.DeletionPolicyDoNotTerminate {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("deletionPolicy"),
			db.GetName(),
			"Can not delete as terminationPolicy is set to \"DoNotTerminate\""))
		return nil, apierrors.NewInvalid(schema.GroupKind{Group: "Neo4j.kubedb.com", Kind: "Neo4j"}, db.GetName(), allErr)
	}
	return nil, nil
}

func (w *Neo4jCustomWebhook) ValidateCreateOrUpdate(db *olddbapi.Neo4j) error {
	var allErr field.ErrorList

	if db.Spec.TLS != nil && db.Spec.TLS.IssuerRef == nil {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("tls").Child("issuerRef"),
			db.Name, "spec.tls.issuerRef' is missing"))
	}

	// number of replicas can not be 0 or less
	if db.Spec.Replicas != nil && *db.Spec.Replicas <= 0 {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("replicas"),
			db.GetName(),
			"number of replicas can not be 0 or less"))
	}

	if db.Spec.Version == "" {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("version"),
			db.GetName(),
			"spec.version' is missing"))
	} else {
		err := w.ValidateVersion(db)
		if err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("version"),
				db.GetName(),
				err.Error()))
		}
	}

	err := w.validateVolumes(db)
	if err != nil {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("podTemplate").Child("spec").Child("volumes"),
			db.GetName(),
			err.Error()))
	}

	err = w.validateVolumesMountPaths(&db.Spec.PodTemplate)
	if err != nil {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("podTemplate").Child("spec").Child("volumeMounts"),
			db.GetName(),
			err.Error()))
	}

	if db.Spec.StorageType == "" {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("storageType"),
			db.GetName(),
			"StorageType can not be empty"))
	} else {
		if db.Spec.StorageType != olddbapi.StorageTypeDurable && db.Spec.StorageType != olddbapi.StorageTypeEphemeral {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("storageType"),
				db.GetName(),
				"StorageType should be either durable or ephemeral"))
		}
	}

	if db.Spec.Monitor != nil {
		if db.Spec.Monitor.Agent == "" {
			return fmt.Errorf("monitor.agent is missing")
		}
		if !mona.IsKnownAgentType(db.Spec.Monitor.Agent) {
			return fmt.Errorf("monitor.agent '%v' is not known", db.Spec.Monitor.Agent)
		}
	}

	if db.Spec.Configuration != nil && db.Spec.Configuration.SecretName != "" {
		configSecret := &core.Secret{}
		err := w.DefaultClient.Get(context.TODO(), client.ObjectKey{
			Name:      db.Spec.Configuration.SecretName,
			Namespace: db.Namespace,
		}, configSecret)
		if err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("configuration").Child("secretName"),
				db.GetName(),
				"failed to get ConfigSecret"))
		}
	}

	// Validate that the git-sync clone root path does not collide with any reserved mount path.
	if err := amv.ValidateGitInitRootPath((*dbapi.InitSpec)(unsafe.Pointer(db.Spec.Init)), Neo4jReservedVolumeMountPaths); err != nil {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("init"), db.GetName(), err.Error()))
	}

	allErr = append(allErr, w.validateLogForwarder(db)...)

	if len(allErr) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "Neo4j.kubedb.com", Kind: "Neo4j"}, db.GetName(), allErr)
}

// validateLogForwarder validates the native log-forwarder feature. It is a no-op when the feature
// is absent, so an unconfigured Neo4j is unaffected.
func (w *Neo4jCustomWebhook) validateLogForwarder(db *olddbapi.Neo4j) field.ErrorList {
	lf := db.Spec.LogForwarder
	if lf == nil {
		return nil
	}
	var errs field.ErrorList
	lfPath := field.NewPath("spec").Child("logForwarder")

	// Only enforce the heavier rules when forwarding is actually enabled; a disabled forwarder may
	// legitimately retain a partial configuration and its state PVCs.
	enabled := lf.Enabled == nil || *lf.Enabled
	if !enabled {
		return errs
	}

	// Destination: exactly one of profile / exporterConfig (also enforced by CEL; repeated here for a
	// clear message and because CEL is unavailable on older clusters).
	dstPath := lfPath.Child("destination")
	hasProfile := lf.Destination.Profile != ""
	hasRaw := lf.Destination.ExporterConfig != ""
	if hasProfile == hasRaw {
		errs = append(errs, field.Invalid(dstPath, db.GetName(),
			"exactly one of destination.profile or destination.exporterConfig must be set"))
	}

	// State storage must be a filesystem, ReadWriteOnce volume with a positive capacity request.
	ssPath := lfPath.Child("stateStorage")
	modes := lf.StateStorage.AccessModes
	if len(modes) != 1 || modes[0] != core.ReadWriteOnce {
		errs = append(errs, field.Invalid(ssPath.Child("accessModes"), db.GetName(),
			"stateStorage must use exactly [ReadWriteOnce]"))
	}
	if lf.StateStorage.VolumeMode != nil && *lf.StateStorage.VolumeMode != core.PersistentVolumeFilesystem {
		errs = append(errs, field.Invalid(ssPath.Child("volumeMode"), db.GetName(),
			"stateStorage must use the Filesystem volumeMode"))
	}
	if qty, ok := lf.StateStorage.Resources.Requests[core.ResourceStorage]; !ok || qty.IsZero() {
		errs = append(errs, field.Invalid(ssPath.Child("resources").Child("requests").Child("storage"),
			db.GetName(), "stateStorage must request a positive storage capacity"))
	}

	// Sources must be advertised by the version's log capabilities.
	if len(lf.Sources) > 0 {
		version := catalog.Neo4jVersion{}
		if err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: db.Spec.Version}, &version); err == nil {
			supported := map[string]bool{}
			for _, c := range version.Spec.LogCapabilities {
				supported[c.Name] = true
			}
			// Only enforce when the version actually advertises capabilities; empty means unknown.
			if len(supported) > 0 {
				for i, s := range lf.Sources {
					if !supported[s.Name] {
						errs = append(errs, field.Invalid(lfPath.Child("sources").Index(i).Child("name"),
							s.Name, fmt.Sprintf("log source %q is not supported by Neo4jVersion %q", s.Name, db.Spec.Version)))
					}
				}
			}
		}
	}

	// Reject collisions with the reserved sidecar container name; native forwarding owns it.
	for i, c := range db.Spec.PodTemplate.Spec.Containers {
		if c.Name == kubedb.LogForwarderContainerName {
			errs = append(errs, field.Invalid(field.NewPath("spec").Child("podTemplate").Child("spec").Child("containers").Index(i).Child("name"),
				c.Name, "container name is reserved by native logForwarder; remove the manual sidecar or disable logForwarder"))
		}
	}

	return errs
}

func (w *Neo4jCustomWebhook) ValidateVersion(db *olddbapi.Neo4j) error {
	rmVersion := catalog.Neo4jVersion{}
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: db.Spec.Version}, &rmVersion)
	if err != nil {
		return errors.New("version not supported")
	}
	return nil
}

var Neo4jReservedVolumes = []string{
	kubedb.Neo4jVolumeData,
	kubedb.GitSecretVolume,
}

func (w *Neo4jCustomWebhook) validateVolumes(db *olddbapi.Neo4j) error {
	if db.Spec.PodTemplate.Spec.Volumes == nil {
		return nil
	}
	rsv := make([]string, len(Neo4jReservedVolumes))
	copy(rsv, Neo4jReservedVolumes)
	volumes := db.Spec.PodTemplate.Spec.Volumes
	for _, rv := range rsv {
		for _, ugv := range volumes {
			if ugv.Name == rv {
				return errors.New("Cannot use a reserve volume name: " + rv)
			}
		}
	}
	return nil
}

var Neo4jReservedVolumeMountPaths = []string{
	kubedb.Neo4jDataDir,
	kubedb.GitSecretMountPath,
}

func (w *Neo4jCustomWebhook) validateVolumesMountPaths(podTemplate *ofst.PodTemplateSpec) error {
	if podTemplate == nil {
		return nil
	}
	if podTemplate.Spec.Containers == nil {
		return nil
	}

	for _, rvmp := range Neo4jReservedVolumeMountPaths {
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

	if podTemplate.Spec.InitContainers == nil {
		return nil
	}

	for _, rvmp := range Neo4jReservedVolumeMountPaths {
		containerList := podTemplate.Spec.InitContainers
		for i := range containerList {
			mountPathList := containerList[i].VolumeMounts
			for j := range mountPathList {
				if mountPathList[j].MountPath == rvmp {
					return errors.New("Can't use a reserve volume mount path name: " + rvmp)
				}
			}
		}
	}

	return nil
}
