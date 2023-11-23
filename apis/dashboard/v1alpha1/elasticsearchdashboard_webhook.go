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
	"fmt"

	"kubedb.dev/apimachinery/apis"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	amv "kubedb.dev/apimachinery/pkg/validator"

	"github.com/pkg/errors"
	"gomodules.xyz/pointer"
	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	kmapi "kmodules.xyz/client-go/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
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
var edLog = logf.Log.WithName("elasticsearchdashboard-validation")

func (ed *ElasticsearchDashboard) SetupWebhookWithManager(mgr manager.Manager) error {
	return builder.WebhookManagedBy(mgr).
		For(ed).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-dashboard-kubedb-com-v1alpha1-elasticsearchdashboard,mutating=true,failurePolicy=fail,sideEffects=None,groups=dashboard.kubedb.com,resources=elasticsearchdashboards,verbs=create;update,versions=v1alpha1,name=melasticsearchdashboard.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Defaulter = &ElasticsearchDashboard{}

func (ed *ElasticsearchDashboard) setDefaultContainerSecurityContext() {
	containerSecurityContext := &core.SecurityContext{}
	if ed.Spec.PodTemplate.Spec.ContainerSecurityContext != nil {
		containerSecurityContext = ed.Spec.PodTemplate.Spec.ContainerSecurityContext
	}
	if containerSecurityContext.AllowPrivilegeEscalation == nil {
		containerSecurityContext.AllowPrivilegeEscalation = pointer.BoolP(false)
	}
	if containerSecurityContext.RunAsNonRoot == nil {
		containerSecurityContext.RunAsNonRoot = pointer.BoolP(true)
	}
	if containerSecurityContext.RunAsUser == nil {
		containerSecurityContext.RunAsUser = pointer.Int64P(1000)
	}
	if containerSecurityContext.RunAsGroup == nil {
		containerSecurityContext.RunAsGroup = pointer.Int64P(1000)
	}
	capabilities := &core.Capabilities{}
	if containerSecurityContext.Capabilities != nil {
		capabilities = containerSecurityContext.Capabilities
	}
	if len(capabilities.Drop) == 0 {
		capabilities.Drop = []core.Capability{"ALL"}
	}
	containerSecurityContext.Capabilities = capabilities
	seccomProfile := &core.SeccompProfile{}
	if containerSecurityContext.SeccompProfile != nil {
		seccomProfile = containerSecurityContext.SeccompProfile
	}
	if seccomProfile.Type == "" {
		seccomProfile.Type = core.SeccompProfileTypeRuntimeDefault
	}
	containerSecurityContext.SeccompProfile = seccomProfile
	ed.Spec.PodTemplate.Spec.ContainerSecurityContext = containerSecurityContext
}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (ed *ElasticsearchDashboard) Default() {
	if ed.Spec.Replicas == nil {
		ed.Spec.Replicas = pointer.Int32P(1)
		edLog.Info(".Spec.Replicas have been set to default")
	}

	apis.SetDefaultResourceLimits(&ed.Spec.PodTemplate.Spec.Resources, DashboardsDefaultResources)
	edLog.Info(".PodTemplate.Spec.Resources have been set to default")

	if len(ed.Spec.TerminationPolicy) == 0 {
		ed.Spec.TerminationPolicy = api.TerminationPolicyWipeOut
		edLog.Info(".Spec.TerminationPolicy have been set to TerminationPolicyWipeOut")
	}

	ed.setDefaultContainerSecurityContext()

	if ed.Spec.EnableSSL {
		if ed.Spec.TLS == nil {
			ed.Spec.TLS = &kmapi.TLSConfig{}
		}
		if ed.Spec.TLS.IssuerRef == nil {
			ed.Spec.TLS.Certificates = kmapi.SetMissingSpecForCertificate(ed.Spec.TLS.Certificates, kmapi.CertificateSpec{
				Alias:      string(ElasticsearchDashboardCACert),
				SecretName: ed.DefaultCertificateSecretName(ElasticsearchDashboardCACert),
			})
		}
		ed.Spec.TLS.Certificates = kmapi.SetMissingSpecForCertificate(ed.Spec.TLS.Certificates, kmapi.CertificateSpec{
			Alias:      string(ElasticsearchDashboardServerCert),
			SecretName: ed.DefaultCertificateSecretName(ElasticsearchDashboardServerCert),
		})
	}
}

// +kubebuilder:webhook:path=/validate-dashboard-kubedb-com-v1alpha1-elasticsearchdashboard,mutating=false,failurePolicy=fail,sideEffects=None,groups=dashboard.kubedb.com,resources=elasticsearchdashboards,verbs=create;update;delete,versions=v1alpha1,name=velasticsearchdashboard.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &ElasticsearchDashboard{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (ed *ElasticsearchDashboard) ValidateCreate() error {
	edLog.Info("validate create", "name", ed.Name)
	err := ed.Validate()
	return err
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (ed *ElasticsearchDashboard) ValidateUpdate(old runtime.Object) error {
	// Skip validation, if UPDATE operation is called after deletion.
	// Case: Removing Finalizer
	if ed.DeletionTimestamp != nil {
		return nil
	}
	err := ed.Validate()
	return err
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (ed *ElasticsearchDashboard) ValidateDelete() error {
	edLog.Info("validate delete", "name", ed.Name)

	var allErr field.ErrorList

	if ed.Spec.TerminationPolicy == api.TerminationPolicyDoNotTerminate {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("terminationpolicy"), ed.Name,
			fmt.Sprintf("ElasticsearchDashboard %s/%s can't be deleted. Change .spec.terminationpolicy", ed.Namespace, ed.Name)))
	}

	if len(allErr) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: "dashboard.kubedb.com", Kind: "ElasticsearchDashboard"},
		ed.Name, allErr)
}

func validateContainerCapabilities(ed *ElasticsearchDashboard) error {
	capabilities := ed.Spec.PodTemplate.Spec.ContainerSecurityContext.Capabilities
	if len(capabilities.Add) > 0 {
		return errors.Errorf("Can't add user provided capabilities")
	}

	if len(capabilities.Drop) != 1 {
		return errors.Errorf("Can't drop more than one capability")
	}

	if capabilities.Drop[0] != "ALL" {
		return errors.Errorf("Have to drop all capabilities")
	}
	return nil
}

func validateContainerUserAndGroup(ed *ElasticsearchDashboard) error {
	if ed.Spec.PodTemplate.Spec.ContainerSecurityContext.RunAsGroup != nil && *ed.Spec.PodTemplate.Spec.ContainerSecurityContext.RunAsGroup != 1000 {
		return errors.Errorf("elasticsearch should run as a group elasticsearch (ID=1000)")
	}

	if ed.Spec.PodTemplate.Spec.ContainerSecurityContext.RunAsUser != nil && *ed.Spec.PodTemplate.Spec.ContainerSecurityContext.RunAsUser != 1000 {
		return errors.Errorf("elasticsearch should run as a user elasticsearch (ID=1000)")
	}

	return nil
}

func validateAllowedPriviledgeEscalation(ed *ElasticsearchDashboard) error {
	if ed.Spec.PodTemplate.Spec.ContainerSecurityContext.AllowPrivilegeEscalation != nil && *ed.Spec.PodTemplate.Spec.ContainerSecurityContext.AllowPrivilegeEscalation {
		return errors.Errorf("AllowedPrivilegeEscalation can't be true")
	}
	return nil
}

func validateRunAsNonRoot(ed *ElasticsearchDashboard) error {
	if ed.Spec.PodTemplate.Spec.ContainerSecurityContext.RunAsNonRoot != nil && !*ed.Spec.PodTemplate.Spec.ContainerSecurityContext.RunAsNonRoot {
		return errors.Errorf("RunAsNonRoot can't be false")
	}
	return nil
}

func validateSeccomprofile(ed *ElasticsearchDashboard) error {
	if ed.Spec.PodTemplate.Spec.ContainerSecurityContext.SeccompProfile.Type != "" && ed.Spec.PodTemplate.Spec.ContainerSecurityContext.SeccompProfile.Type != core.SeccompProfileTypeRuntimeDefault {
		return errors.Errorf("Seccomprofile type must be RuntimeDefault")
	}
	return nil
}

func (ed *ElasticsearchDashboard) Validate() error {
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
	if err := amv.ValidateEnvVar(ed.Spec.PodTemplate.Spec.Env, forbiddenEnvVars, ResourceKindElasticsearchDashboard); err != nil {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("podtemplate").Child("spec").Child("env"), ed.Name,
			"Invalid spec.podtemplate.spec.env , avoid using the forbidden env variables"))
	}

	if len(allErr) == 0 {
		return nil
	}

	err := validateContainerUserAndGroup(ed)
	if err != nil {
		return err
	}

	err = validateContainerCapabilities(ed)
	if err != nil {
		return err
	}

	err = validateAllowedPriviledgeEscalation(ed)
	if err != nil {
		return nil
	}

	err = validateRunAsNonRoot(ed)
	if err != nil {
		return err
	}

	err = validateSeccomprofile(ed)
	if err != nil {
		return err
	}

	return apierrors.NewInvalid(schema.GroupKind{Group: "dashboard.kubedb.com", Kind: "ElasticsearchDashboard"}, ed.Name, allErr)
}
