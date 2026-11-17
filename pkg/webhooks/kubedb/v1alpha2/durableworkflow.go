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
	"net/url"
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

	allErr = append(allErr, w.validatePersistence(db, old, version)...)

	if version != nil {
		if t := db.Spec.Topology; t != nil && t.InternalFrontend != nil && !version.IsTemporal() {
			allErr = append(allErr, field.Forbidden(spec.Child("topology").Child("internalFrontend"), "internalFrontend is only available for Temporal"))
		}
		if db.Spec.UI != nil && db.Spec.UI.Enabled && db.Spec.UI.Auth == olddbapi.DurableWorkflowUIAuthJWT && !version.IsCadence() {
			allErr = append(allErr, field.Forbidden(spec.Child("ui").Child("auth"), "UI JWT auth is only available for Cadence"))
		}
		if version.IsCadence() && db.IsFrontendClientAuthRequired(version) && db.IsTLSEnabled() {
			allErr = append(allErr, field.Forbidden(spec.Child("frontendClientAuth"),
				"the Cadence web UI and the readiness probe can't present a client certificate; set frontendClientAuth to NotRequired"))
		}
		if version.IsCadence() && db.IsTLSEnabled() && !version.Spec.PublicClientTLS {
			allErr = append(allErr, field.Forbidden(spec.Child("tls"),
				fmt.Sprintf("%s doesn't support TLS: Cadence's worker service can't reach a TLS frontend without publicClient.tls (set publicClientTLS on a DurableWorkflowVersion built with it)", version.Name)))
		}
		allErr = append(allErr, validateAuthorization(db, version)...)
	}
	archErrs, archWarnings := validateArchival(db, version)
	allErr = append(allErr, archErrs...)
	warnings = append(warnings, archWarnings...)

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

func (w *DurableWorkflowCustomWebhook) validatePersistence(db, old *olddbapi.DurableWorkflow, version *catalog.DurableWorkflowVersion) field.ErrorList {
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

	if db.Spec.Persistence.Managed != nil {
		allErr = append(allErr, w.validateManaged(db, old, version)...)
	} else {
		if old != nil && old.Spec.Persistence.Managed != nil {
			allErr = append(allErr, field.Forbidden(p.Child("managed"), "managed can't be removed: the database it created is owned by the DurableWorkflow"))
		}
		validateStore(p.Child("default"), db.Spec.Persistence.Default, false)
	}
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

// validateManaged checks spec.persistence.managed, the database the operator creates itself.
func (w *DurableWorkflowCustomWebhook) validateManaged(db, old *olddbapi.DurableWorkflow, version *catalog.DurableWorkflowVersion) field.ErrorList {
	var allErr field.ErrorList
	p := field.NewPath("spec").Child("persistence")
	m := db.Spec.Persistence.Managed
	mp := p.Child("managed")

	if d := db.Spec.Persistence.Default; d.ObjectReference != nil && (d.Name != "" || d.Namespace != "") {
		allErr = append(allErr, field.Forbidden(p.Child("default"), "default.name and managed are exclusive: the managed database is "+db.ManagedStoreName()))
	}
	if db.Spec.Persistence.Visibility != nil {
		allErr = append(allErr, field.Forbidden(p.Child("visibility"), "a managed database holds the visibility store as well; visibility can't be set with managed"))
	}
	if d := db.Spec.Persistence.Default; !d.ShouldCreateDatabase() {
		allErr = append(allErr, field.Forbidden(p.Child("default").Child("createDatabase"), "the operator always creates the database and the user in a managed database"))
	}

	engine := m.ManagedEngine()
	switch engine {
	case olddbapi.DurableWorkflowManagedPostgres, olddbapi.DurableWorkflowManagedMySQL:
		if t := db.Spec.Persistence.Default.Type; t != "" && t != olddbapi.DurableWorkflowStoreType(engine) {
			allErr = append(allErr, field.Invalid(p.Child("default").Child("type"), t, fmt.Sprintf("must be %s, the engine of the managed database", engine)))
		}
		if version != nil && !version.SupportsStore(string(engine), false) {
			allErr = append(allErr, field.Invalid(mp.Child("engine"), engine, fmt.Sprintf("%s is not supported as the default store of %s %s",
				engine, version.Spec.Distribution, version.Spec.Version)))
		}
	default:
		allErr = append(allErr, field.NotSupported(mp.Child("engine"), engine, []string{
			string(olddbapi.DurableWorkflowManagedPostgres), string(olddbapi.DurableWorkflowManagedMySQL),
		}))
	}

	if m.Version == "" {
		allErr = append(allErr, field.Required(mp.Child("version"), "no version could be picked from the catalog; set one"))
	} else if engine == olddbapi.DurableWorkflowManagedPostgres || engine == olddbapi.DurableWorkflowManagedMySQL {
		var obj client.Object = &catalog.PostgresVersion{}
		if engine == olddbapi.DurableWorkflowManagedMySQL {
			obj = &catalog.MySQLVersion{}
		}
		if err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: m.Version}, obj); err != nil {
			allErr = append(allErr, field.Invalid(mp.Child("version"), m.Version, fmt.Sprintf("%sVersion not found", engine)))
		}
	}

	switch r := m.ManagedReplicas(); {
	case r < 1:
		allErr = append(allErr, field.Invalid(mp.Child("replicas"), r, "number of replicas can not be 0 or less"))
	case engine == olddbapi.DurableWorkflowManagedMySQL && (r == 2 || r > 9):
		allErr = append(allErr, field.Invalid(mp.Child("replicas"), r, "MySQL group replication needs 3 to 9 replicas; use 1 for a single server"))
	}
	if _, ok := m.Storage.Resources.Requests[core.ResourceStorage]; !ok {
		allErr = append(allErr, field.Required(mp.Child("storage").Child("resources").Child("requests").Key("storage"), "the size of the volume is required"))
	}

	if old != nil {
		if o := old.Spec.Persistence.Managed; o == nil {
			allErr = append(allErr, field.Forbidden(mp, "managed can't be added to an existing DurableWorkflow: it already uses the databases it references"))
		} else {
			// The version is filled in by the defaulter when it was left out.
			if o.Version == "" {
				o = o.DeepCopy()
				o.Version = m.Version
			}
			if !apiequality.Semantic.DeepEqual(o, m) {
				allErr = append(allErr, field.Forbidden(mp, fmt.Sprintf(
					"managed is immutable: change the database %s with its own ops requests (UpdateVersion, VerticalScaling, VolumeExpansion, HorizontalScaling, Reconfigure)", db.ManagedStoreName(),
				)))
			}
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

func validateAuthorization(db *olddbapi.DurableWorkflow, version *catalog.DurableWorkflowVersion) field.ErrorList {
	a := db.Spec.Authorization
	if a == nil {
		return nil
	}
	var errs field.ErrorList
	path := field.NewPath("spec").Child("authorization")
	if a.JWKSURL == "" && a.PublicKeySecret == nil {
		errs = append(errs, field.Required(path, "one of jwksURL and publicKeySecret is required"))
	}
	if a.JWKSURL != "" && a.PublicKeySecret != nil {
		errs = append(errs, field.Invalid(path.Child("publicKeySecret"), a.PublicKeySecret.Name, "jwksURL and publicKeySecret are exclusive"))
	}
	if version.IsTemporal() {
		if a.MaxTokenTTL != nil {
			errs = append(errs, field.Forbidden(path.Child("maxTokenTTL"), "maxTokenTTL is only available for Cadence"))
		}
		if a.GroupsAttributePath != "" || a.AdminAttributePath != "" {
			errs = append(errs, field.Forbidden(path, "groupsAttributePath and adminAttributePath are only available for Cadence"))
		}
		// Without an internal frontend the system worker would call the authorized frontend with no token.
		if db.Spec.Topology == nil || db.Spec.Topology.InternalFrontend == nil {
			errs = append(errs, field.Required(field.NewPath("spec").Child("topology").Child("internalFrontend"),
				"Temporal authorization needs an internalFrontend, which the system worker uses instead of the authorized frontend"))
		}
		return errs
	}
	if a.PublicKeySecret != nil {
		errs = append(errs, field.Forbidden(path.Child("publicKeySecret"), "Cadence verifies external tokens with a JWKS URL; its static key is the one the cluster signs internal tokens with"))
	}
	if a.PermissionsClaimName != "" || a.PermissionsRegex != "" || a.Audience != "" {
		errs = append(errs, field.Forbidden(path, "permissionsClaimName, permissionsRegex and audience are only available for Temporal"))
	}
	if a.MaxTokenTTL != nil && *a.MaxTokenTTL <= 0 {
		errs = append(errs, field.Invalid(path.Child("maxTokenTTL"), *a.MaxTokenTTL, "must be greater than 0"))
	}
	return errs
}

func validateArchival(db *olddbapi.DurableWorkflow, version *catalog.DurableWorkflowVersion) (field.ErrorList, admission.Warnings) {
	a := db.Spec.Archival
	if a == nil {
		return nil, nil
	}
	var errs field.ErrorList
	var warnings admission.Warnings
	root := field.NewPath("spec").Child("archival")
	targets := map[string]*olddbapi.DurableWorkflowArchivalTarget{"history": a.History, "visibility": a.Visibility}

	var dirs []string
	s3Secret := ""
	for _, name := range []string{"history", "visibility"} {
		t := targets[name]
		if t == nil {
			continue
		}
		p := root.Child(name)
		want := olddbapi.ArchivalURIScheme(t.Provider)
		if want == "" {
			errs = append(errs, field.NotSupported(p.Child("provider"), t.Provider,
				[]string{olddbapi.DurableWorkflowArchivalProviderFilestore, olddbapi.DurableWorkflowArchivalProviderS3, olddbapi.DurableWorkflowArchivalProviderGCS}))
			continue
		}
		u, err := url.Parse(t.URI)
		switch {
		case t.URI == "":
			errs = append(errs, field.Required(p.Child("uri"), "uri is required"))
		case err != nil || u.Scheme != want:
			errs = append(errs, field.Invalid(p.Child("uri"), t.URI, fmt.Sprintf("the %s provider needs a %s:// URI", t.Provider, want)))
		case t.Provider != olddbapi.DurableWorkflowArchivalProviderFilestore && u.Host == "":
			errs = append(errs, field.Invalid(p.Child("uri"), t.URI, "the bucket is missing, use "+want+"://bucket/prefix"))
		}

		switch t.Provider {
		case olddbapi.DurableWorkflowArchivalProviderFilestore:
			if t.ConfigSecret != nil {
				errs = append(errs, field.Forbidden(p.Child("configSecret"), "the filestore provider has no credentials"))
			}
			if dir, err := olddbapi.ArchivalFilestoreDir(t.URI); err != nil {
				if t.URI != "" && (u != nil && u.Scheme == "file") {
					errs = append(errs, field.Invalid(p.Child("uri"), t.URI, err.Error()))
				}
			} else {
				dirs = append(dirs, dir)
			}
			warnings = append(warnings, fmt.Sprintf("spec.archival.%s: the filestore archive is an emptyDir of each pod; it is lost when the pod is, and not shared between pods. Use s3 or gcs for anything but tests", name))
		case olddbapi.DurableWorkflowArchivalProviderS3, olddbapi.DurableWorkflowArchivalProviderGCS:
			if t.ConfigSecret == nil || t.ConfigSecret.Name == "" {
				errs = append(errs, field.Required(p.Child("configSecret"), fmt.Sprintf("the %s provider needs a configSecret with its credentials", t.Provider)))
			} else if t.Provider == olddbapi.DurableWorkflowArchivalProviderS3 {
				if s3Secret != "" && s3Secret != t.ConfigSecret.Name {
					errs = append(errs, field.Invalid(p.Child("configSecret"), t.ConfigSecret.Name,
						"history and visibility must use the same s3 secret: the credentials are env variables of the whole process"))
				}
				s3Secret = t.ConfigSecret.Name
			}
			if t.Provider == olddbapi.DurableWorkflowArchivalProviderGCS && version != nil && version.IsTemporal() {
				warnings = append(warnings, fmt.Sprintf("spec.archival.%s: the Temporal server image must be built with the temporal_with_gstorage tag, or it refuses to start", name))
			}
		}
	}
	// two filestore paths may be equal (one mount), but not nested: the mounts would shadow each other
	if len(dirs) == 2 && dirs[0] != dirs[1] &&
		(strings.HasPrefix(dirs[0], dirs[1]+"/") || strings.HasPrefix(dirs[1], dirs[0]+"/")) {
		errs = append(errs, field.Invalid(root.Child("visibility").Child("uri"), a.Visibility.URI, "filestore paths of history and visibility can't be nested"))
	}
	return errs, warnings
}
