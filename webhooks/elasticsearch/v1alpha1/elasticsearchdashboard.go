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

	edapi "kubedb.dev/apimachinery/apis/elasticsearch/v1alpha1"
	"kubedb.dev/apimachinery/apis/kubedb"
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"
	amv "kubedb.dev/apimachinery/pkg/validator"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	coreutil "kmodules.xyz/client-go/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var forbiddenEnvVars = []string{
	ES_USER_ENV,
	ES_PASSWORD_ENV,
	ES_USER_KEY,
	ES_PASSWORD_KEY,
	OS_USER_KEY,
	OS_PASSWORD_KEY,
	DashboardServerHostKey,
	DashboardServerNameKey,
	DashboardServerPortKey,
	DashboardServerSSLCaKey,
	DashboardServerSSLCertKey,
	DashboardServerSSLKey,
	DashboardServerSSLEnabledKey,
	ElasticsearchSSLCaKey,
	ElasticsearchHostsKey,
	OpensearchHostsKey,
	OpensearchSSLCaKey,
}

// log is for logging in this package.
var edLog = logf.Log.WithName("elasticsearchelasticsearch-validation")

// SetupDashboardWebhookWithManager registers the webhook for Solr in the manager.
func SetupElasticsearchdashboardWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&edapi.ElasticsearchDashboard{}).
		WithValidator(&ElasticsearchdashboardCustomWebhook{mgr.GetClient()}).
		WithDefaulter(&ElasticsearchdashboardCustomWebhook{mgr.GetClient()}).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-elasticsearch-kubedb-com-v1alpha1-elasticsearchelasticsearch,mutating=true,failurePolicy=fail,sideEffects=None,groups=elasticsearch.kubedb.com,resources=elasticsearchelasticsearchs,verbs=create;update,versions=v1alpha1,name=melasticsearchelasticsearch.kb.io,admissionReviewVersions={v1,v1beta1}

// +kubebuilder:object:generate=false
type ElasticsearchdashboardCustomWebhook struct {
	DefaultClient client.Client
}

var _ webhook.CustomDefaulter = &ElasticsearchdashboardCustomWebhook{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (w *ElasticsearchdashboardCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	dashboard, ok := obj.(*edapi.ElasticsearchDashboard)
	if !ok {
		return fmt.Errorf("expected an Solr object but got %T", obj)
	}

	edLog.Info("default", "name", dashboard.Name)

	dashboard.SetDefaults(w.DefaultClient)
	return nil
}

// +kubebuilder:webhook:path=/validate-elasticsearch-kubedb-com-v1alpha1-elasticsearchelasticsearch,mutating=false,failurePolicy=fail,sideEffects=None,groups=elasticsearch.kubedb.com,resources=elasticsearchelasticsearchs,verbs=create;update;delete,versions=v1alpha1,name=velasticsearchelasticsearch.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomValidator = &ElasticsearchdashboardCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *ElasticsearchdashboardCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	ed, ok := obj.(*edapi.ElasticsearchDashboard)
	if !ok {
		return nil, fmt.Errorf("expected an dashboard object but got %T", obj)
	}

	edLog.Info("validate create", "name", ed.Name)
	return nil, w.Validate(ed)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *ElasticsearchdashboardCustomWebhook) ValidateUpdate(ctx context.Context, old, newObj runtime.Object) (admission.Warnings, error) {
	ed, ok := newObj.(*edapi.ElasticsearchDashboard)
	if !ok {
		return nil, fmt.Errorf("expected an dashboard object but got %T", ed)
	}
	// Skip validation, if UPDATE operation is called after deletion.
	// Case: Removing Finalizer
	if ed.DeletionTimestamp != nil {
		return nil, nil
	}
	return nil, w.Validate(ed)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (w *ElasticsearchdashboardCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	ed, ok := obj.(*edapi.ElasticsearchDashboard)
	if !ok {
		return nil, fmt.Errorf("expected an dashboard object but got %T", obj)
	}

	edLog.Info("validate delete", "name", ed.Name)

	var allErr field.ErrorList

	if ed.Spec.DeletionPolicy == dbapi.DeletionPolicyDoNotTerminate {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("deletionpolicy"), ed.Name,
			fmt.Sprintf("ElasticsearchDashboard %s/%s can't be deleted. Change .spec.deletionpolicy", ed.Namespace, ed.Name)))
	}

	if len(allErr) == 0 {
		return nil, nil
	}

	return nil, apierrors.NewInvalid(
		schema.GroupKind{Group: "elasticsearch.kubedb.com", Kind: "ElasticsearchDashboard"},
		ed.Name, allErr)
}

func (w *ElasticsearchdashboardCustomWebhook) Validate(ed *edapi.ElasticsearchDashboard) error {
	var allErr field.ErrorList

	// database ref is required
	if ed.Spec.DatabaseRef == nil {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("databaseref"), ed.Name,
			"spec.databaseref can't be empty"))
	}

	// validate if user provided replicas are non-negative
	// user may provide 0 replicas
	if *ed.Spec.Replicas < 0 {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("replicas"), ed.Name,
			fmt.Sprintf("spec.replicas %v invalid. Must be greater than zero", ed.Spec.Replicas)))
	}

	// env variables needs to be validated
	// so that variables provided in config secret
	// and credential env may not be overwritten
	container := coreutil.GetContainerByName(ed.Spec.PodTemplate.Spec.Containers, kubedb.ElasticsearchContainerName)
	if err := amv.ValidateEnvVar(container.Env, forbiddenEnvVars, edapi.ResourceKindElasticsearchDashboard); err != nil {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("podtemplate").Child("spec").Child("containers").Child("env"), ed.Name,
			"Invalid spec.podtemplate.spec.containers[i].env , avoid using the forbidden env variables"))
	}

	initContainer := coreutil.GetContainerByName(ed.Spec.PodTemplate.Spec.InitContainers, kubedb.ElasticsearchInitConfigMergerContainerName)
	if err := amv.ValidateEnvVar(initContainer.Env, forbiddenEnvVars, edapi.ResourceKindElasticsearchDashboard); err != nil {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("podtemplate").Child("spec").Child("initContainers").Child("env"), ed.Name,
			"Invalid spec.podtemplate.spec.initContainers[i].env , avoid using the forbidden env variables"))
	}

	if len(allErr) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{Group: "elasticsearch.kubedb.com", Kind: "ElasticsearchDashboard"}, ed.Name, allErr)
}
