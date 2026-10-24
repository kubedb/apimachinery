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
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"
	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	amv "kubedb.dev/apimachinery/pkg/validator"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// SetupDurableWorkflowWebhookWithManager registers the webhook for DurableWorkflow in the manager.
func SetupDurableWorkflowWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&olddbapi.DurableWorkflow{}).
		WithValidator(amv.WithServiceTemplateAliasWarnings(&DurableWorkflowCustomWebhook{mgr.GetClient()}, dbapi.PrimaryServiceAlias, dbapi.StatsServiceAlias)).
		WithDefaulter(&DurableWorkflowCustomWebhook{mgr.GetClient()}).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-kubedb-com-v1alpha2-durableworkflow,mutating=true,failurePolicy=fail,sideEffects=None,groups=kubedb.com,resources=durableworkflows,verbs=create;update,versions=v1alpha2,name=mdurableworkflow.kb.io,admissionReviewVersions={v1,v1beta1}

// +kubebuilder:object:generate=false
type DurableWorkflowCustomWebhook struct {
	DefaultClient client.Client
}

var _ webhook.CustomDefaulter = &DurableWorkflowCustomWebhook{}

// log is for logging in this package.
var durableWorkflowLog = logf.Log.WithName("durableworkflow-resource")

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (w *DurableWorkflowCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	if isDeletionInProgress(obj) {
		return nil
	}
	db, ok := obj.(*olddbapi.DurableWorkflow)
	if !ok {
		return fmt.Errorf("expected a DurableWorkflow object but got %T", obj)
	}

	durableWorkflowLog.Info("default", "name", db.GetName())
	db.SetDefaults(w.DefaultClient)
	return nil
}

//+kubebuilder:webhook:path=/validate-kubedb-com-v1alpha2-durableworkflow,mutating=false,failurePolicy=fail,sideEffects=None,groups=kubedb.com,resources=durableworkflows,verbs=create;update,versions=v1alpha2,name=vdurableworkflow.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomValidator = &DurableWorkflowCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *DurableWorkflowCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	if isDeletionInProgress(obj) {
		return nil, nil
	}
	if err := amv.ValidateNameUniqueness(ctx, w.DefaultClient, obj); err != nil {
		return nil, err
	}
	db, ok := obj.(*olddbapi.DurableWorkflow)
	if !ok {
		return nil, fmt.Errorf("expected a DurableWorkflow object but got %T", obj)
	}
	durableWorkflowLog.Info("validate create", "name", db.GetName())
	return w.validate(db, nil)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *DurableWorkflowCustomWebhook) ValidateUpdate(ctx context.Context, old, newObj runtime.Object) (admission.Warnings, error) {
	if isDeletionInProgress(newObj) {
		return nil, nil
	}
	db, ok := newObj.(*olddbapi.DurableWorkflow)
	if !ok {
		return nil, fmt.Errorf("expected a DurableWorkflow object but got %T", newObj)
	}
	oldDB, ok := old.(*olddbapi.DurableWorkflow)
	if !ok {
		return nil, fmt.Errorf("expected a DurableWorkflow object but got %T", old)
	}

	durableWorkflowLog.Info("validate update", "name", db.GetName())
	return w.validate(db, oldDB)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (w *DurableWorkflowCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	db, ok := obj.(*olddbapi.DurableWorkflow)
	if !ok {
		return nil, fmt.Errorf("expected a DurableWorkflow object but got %T", obj)
	}
	durableWorkflowLog.Info("validate delete", "name", db.GetName())

	var allErr field.ErrorList
	if db.Spec.DeletionPolicy == olddbapi.DeletionPolicyDoNotTerminate {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("deletionPolicy"),
			db.GetName(),
			"Can not delete as terminationPolicy is set to \"DoNotTerminate\""))
		return nil, apierrors.NewInvalid(schema.GroupKind{Group: "kubedb.com", Kind: "DurableWorkflow"}, db.GetName(), allErr)
	}
	return nil, nil
}

func (w *DurableWorkflowCustomWebhook) validate(db, old *olddbapi.DurableWorkflow) (admission.Warnings, error) {
	var allErr field.ErrorList
	var warnings admission.Warnings
	spec := field.NewPath("spec")

	var version *catalog.DurableWorkflowVersion
	if db.Spec.Version == "" {
		allErr = append(allErr, field.Required(spec.Child("version"), "spec.version is missing"))
	} else {
		v, err := w.getVersion(db)
		if err != nil {
			allErr = append(allErr, field.Invalid(spec.Child("version"), db.Spec.Version, err.Error()))
		} else {
			version = v
		}
	}

	if old != nil && old.Spec.NumHistoryShards != nil && db.Spec.NumHistoryShards != nil &&
		*old.Spec.NumHistoryShards != *db.Spec.NumHistoryShards {
		allErr = append(allErr, field.Forbidden(spec.Child("numHistoryShards"), "numHistoryShards is immutable"))
	}
	if db.Spec.NumHistoryShards != nil && *db.Spec.NumHistoryShards < 1 {
		allErr = append(allErr, field.Invalid(spec.Child("numHistoryShards"), *db.Spec.NumHistoryShards, "must be at least 1"))
	}

	if t := db.Spec.Topology; t != nil {
		for role, node := range map[string]*olddbapi.DurableWorkflowNode{
			kubedb.DurableWorkflowRoleFrontend:         t.Frontend,
			kubedb.DurableWorkflowRoleHistory:          t.History,
			kubedb.DurableWorkflowRoleMatching:         t.Matching,
			kubedb.DurableWorkflowRoleWorker:           t.Worker,
			kubedb.DurableWorkflowRoleInternalFrontend: t.InternalFrontend,
		} {
			if node != nil && node.Replicas != nil && *node.Replicas < 1 {
				allErr = append(allErr, field.Invalid(spec.Child("topology").Child(role).Child("replicas"), *node.Replicas, "number of replicas can not be 0 or less"))
			}
		}
		if version != nil && t.History != nil && t.History.Replicas != nil && db.Spec.NumHistoryShards != nil &&
			*t.History.Replicas > *db.Spec.NumHistoryShards {
			warnings = append(warnings, fmt.Sprintf("history replicas (%d) exceed numHistoryShards (%d); the extra pods will own no shards",
				*t.History.Replicas, *db.Spec.NumHistoryShards))
		}
	}

	allErr = append(allErr, w.validatePersistence(db, version)...)

	if version != nil {
		if t := db.Spec.Topology; t != nil && t.InternalFrontend != nil && !version.IsTemporal() {
			allErr = append(allErr, field.Forbidden(spec.Child("topology").Child("internalFrontend"), "internalFrontend is only available for Temporal"))
		}
		if db.Spec.UI != nil && db.Spec.UI.Enabled && db.Spec.UI.Auth == olddbapi.DurableWorkflowUIAuthJWT && !version.IsCadence() {
			allErr = append(allErr, field.Forbidden(spec.Child("ui").Child("auth"), "UI JWT auth is only available for Cadence"))
		}
		if version.IsCadence() && db.Spec.UI != nil && db.Spec.UI.Enabled && db.IsFrontendClientAuthRequired(version) && db.IsTLSEnabled() {
			allErr = append(allErr, field.Forbidden(spec.Child("frontendClientAuth"),
				"the Cadence web UI can't present a client certificate; set frontendClientAuth to NotRequired or disable the UI"))
		}
		if version.IsTemporal() && db.Spec.Authorization != nil && db.Spec.Authorization.MaxTokenTTL != nil {
			allErr = append(allErr, field.Forbidden(spec.Child("authorization").Child("maxTokenTTL"), "maxTokenTTL is only available for Cadence"))
		}
	}

	if db.Spec.TLS != nil && db.Spec.TLS.IssuerRef == nil {
		allErr = append(allErr, field.Required(spec.Child("tls").Child("issuerRef"), "spec.tls.issuerRef is missing"))
	}

	if db.Spec.Monitor != nil {
		if db.Spec.Monitor.Agent == "" {
			allErr = append(allErr, field.Required(spec.Child("monitor").Child("agent"), "monitor.agent is missing"))
		} else if !mona.IsKnownAgentType(db.Spec.Monitor.Agent) {
			allErr = append(allErr, field.Invalid(spec.Child("monitor").Child("agent"), db.Spec.Monitor.Agent, "monitor.agent is not known"))
		}
	}

	if len(allErr) == 0 {
		return warnings, nil
	}
	return warnings, apierrors.NewInvalid(schema.GroupKind{Group: "kubedb.com", Kind: "DurableWorkflow"}, db.GetName(), allErr)
}

func (w *DurableWorkflowCustomWebhook) validatePersistence(db *olddbapi.DurableWorkflow, version *catalog.DurableWorkflowVersion) field.ErrorList {
	var allErr field.ErrorList
	p := field.NewPath("spec").Child("persistence")

	validateStore := func(path *field.Path, s olddbapi.DurableWorkflowStore, visibility bool) {
		if s.ObjectReference == nil || s.Name == "" {
			allErr = append(allErr, field.Required(path.Child("name"), "an AppBinding name is required"))
		}
		if !s.ShouldCreateDatabase() && (s.AuthSecret == nil || s.AuthSecret.Name == "") {
			allErr = append(allErr, field.Required(path.Child("authSecret"), "authSecret is required when createDatabase is false"))
		}
		if version != nil && s.Type != "" && !version.SupportsStore(string(s.Type), visibility) {
			role := "default"
			if visibility {
				role = "visibility"
			}
			allErr = append(allErr, field.Invalid(path.Child("type"), s.Type, fmt.Sprintf("%s is not supported as the %s store of %s %s",
				s.Type, role, version.Spec.Distribution, version.Spec.Version)))
		}
	}

	validateStore(p.Child("default"), db.Spec.Persistence.Default, false)
	if db.Spec.Persistence.Visibility != nil {
		validateStore(p.Child("visibility"), *db.Spec.Persistence.Visibility, true)
	}

	if version != nil {
		visibilityType := db.VisibilityStore().Type
		isSearch := visibilityType == olddbapi.DurableWorkflowStoreElasticsearch || visibilityType == olddbapi.DurableWorkflowStoreOpenSearch
		switch {
		case version.IsCadence() && isSearch && db.Spec.Persistence.KafkaRef == nil:
			allErr = append(allErr, field.Required(p.Child("kafkaRef"), "Cadence requires kafkaRef when the visibility store is Elasticsearch or OpenSearch"))
		case version.IsTemporal() && db.Spec.Persistence.KafkaRef != nil:
			allErr = append(allErr, field.Forbidden(p.Child("kafkaRef"), "kafkaRef is not used by Temporal"))
		case version.IsCadence() && !isSearch && db.Spec.Persistence.KafkaRef != nil:
			allErr = append(allErr, field.Forbidden(p.Child("kafkaRef"), "kafkaRef is only used with an Elasticsearch or OpenSearch visibility store"))
		}
	}
	return allErr
}

func (w *DurableWorkflowCustomWebhook) getVersion(db *olddbapi.DurableWorkflow) (*catalog.DurableWorkflowVersion, error) {
	v := &catalog.DurableWorkflowVersion{}
	if err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: db.Spec.Version}, v); err != nil {
		return nil, errors.New("version not supported")
	}
	return v, nil
}
