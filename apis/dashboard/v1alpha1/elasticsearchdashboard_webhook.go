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
	"errors"
	"fmt"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	amv "kubedb.dev/apimachinery/pkg/validator"

	"gomodules.xyz/pointer"
	"gomodules.xyz/x/arrays"
	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var forbiddenEnvVars = []string{
	"ELASTICSEARCH_USERNAME",
	"ELASTICSEARCH_PASSWORD",
	"server.name",
	"server.port",
	"server.host",
	"server.ssl.enabled",
	"server.ssl.certificate",
	"server.ssl.key",
	"server.ssl.certificateAuthorities",
	"elasticsearch.hosts",
	"elasticsearch.username",
	"elasticsearch.password",
	"elasticsearch.ssl.certificateAuthorities",
}

var allowedVersions = []string{
	"xpack-7.14.0",
}

// log is for logging in this package.
var elasticsearchdashboardlog = logf.Log.WithName("elasticsearchdashboard-resource")

func (r *ElasticsearchDashboard) SetupWebhookWithManager(mgr manager.Manager) error {
	return builder.WebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-dashboard-kubedb-com-v1alpha1-elasticsearchdashboard,mutating=true,failurePolicy=fail,sideEffects=None,groups=dashboard.kubedb.com,resources=elasticsearchdashboards,verbs=create;update,versions=v1alpha1,name=melasticsearchdashboard.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Defaulter = &ElasticsearchDashboard{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *ElasticsearchDashboard) Default() {
	elasticsearchdashboardlog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.

	if r.Spec.Replicas == nil {
		r.Spec.Replicas = pointer.Int32P(1)
	}

	if r.Spec.PodTemplate.Spec.Resources.Limits == nil {
		r.Spec.PodTemplate.Spec.Resources.Limits = core.ResourceList{
			"memory": resource.MustParse(ElasticsearchDashboardMemLimit),
		}
	}

	if r.Spec.PodTemplate.Spec.Resources.Requests == nil {
		r.Spec.PodTemplate.Spec.Resources.Requests = core.ResourceList{
			"cpu":    resource.MustParse(ElasticsearchDashboardCpuReq),
			"memory": resource.MustParse(ElasticsearchDashboardMemReq),
		}
	}

	if len(r.Spec.TerminationPolicy) == 0 {
		r.Spec.TerminationPolicy = api.TerminationPolicyDoNotTerminate
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:path=/validate-dashboard-kubedb-com-v1alpha1-elasticsearchdashboard,mutating=false,failurePolicy=fail,sideEffects=None,groups=dashboard.kubedb.com,resources=elasticsearchdashboards,verbs=create;update;delete,versions=v1alpha1,name=velasticsearchdashboard.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &ElasticsearchDashboard{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *ElasticsearchDashboard) ValidateCreate() error {
	elasticsearchdashboardlog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.

	err := r.Validate()
	return err
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *ElasticsearchDashboard) ValidateUpdate(old runtime.Object) error {
	elasticsearchdashboardlog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	// Skip validation, if UPDATE operation is called after deletion.
	// Case: Removing Finalizer
	if r.DeletionTimestamp != nil {
		return nil
	}
	err := r.Validate()
	return err
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *ElasticsearchDashboard) ValidateDelete() error {
	elasticsearchdashboardlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.

	var allErr field.ErrorList

	if r.Spec.TerminationPolicy == api.TerminationPolicyDoNotTerminate {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("terminationpolicy"), r.Name,
			fmt.Sprintf("ElasticsearchDashboard %s/%s can't be deleted. Change .spec.terminationpolicy", r.Namespace, r.Name)))
	}

	if len(allErr) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: "dashboard.kubedb.com", Kind: "ElasticsearchDashboard"},
		r.Name, allErr)
}

func (r *ElasticsearchDashboard) Validate() error {

	// TODO(user): fill in your validation logic upon object creation or update

	var allErr field.ErrorList

	// The resource name length is 63 character like all Kubernetes objects
	// (which must fit in a DNS subdomain). The ElasticsearchDashboard controller appends
	// suffixes(max 13 characters) to resources that it creates and watches.  Therefore, ElasticsearchDashboard
	// names must have length <= 63-13=50. If we don't validate this here,
	// then ElasticsearchDashboard resource creations will fail later.
	if len(r.Name) == validation.DNS1035LabelMaxLength-13 {
		allErr = append(allErr, field.Invalid(field.NewPath("Name"), r.Name,
			fmt.Sprintf("%v is too long. keep it within 50 characters", r.Name)))
	}

	//database ref is required
	if r.Spec.DatabaseRef == nil {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("databaseref"), r.Name,
			"spec.databaseref can't be empty"))
	}

	// validate if user provided replicas are non-negative
	// user may provide 0 replicas
	if *r.Spec.Replicas < 0 {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("replicas"), r.Name,
			fmt.Sprintf("spec.replicas %v invalid. Must be greater than zero", r.Spec.Replicas)))
	}

	// SSL can not be enabled if security is disabled
	if r.Spec.DisableSecurity && r.Spec.EnableSSL {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("disablesecurity", "enablessl"), r.Name,
			"to enable spec.enableSSL, spec.disableSecurity needs to be set to false"))
	}

	// env variables needs to be validated
	// so that variables provided in config secret
	// and credential env may not be overwritten
	if err := amv.ValidateEnvVar(r.Spec.PodTemplate.Spec.Env, forbiddenEnvVars, ResourceKindElasticsearchDashboard); err != nil {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("podtemplate").Child("spec").Child("env"), r.Name,
			"Invalid spec.podtemplate.spec.env , avoid using the forbidden env variables"))
	}

	if err := r.ValidateVersion(r.Spec.Version); err != nil {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("version"), r.Name,
			fmt.Sprintf("Invalid spec.version , use a valid version for "+ResourceKindElasticsearchDashboard)))
	}

	if len(allErr) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{Group: "dashboard.kubedb.com", Kind: "ElasticsearchDashboard"}, r.Name, allErr)
}

func (r *ElasticsearchDashboard) ValidateVersion(version string) error {

	present, _ := arrays.Contains(allowedVersions, version)
	if !present {
		return errors.New("invalid version")
	}
	return nil

}
