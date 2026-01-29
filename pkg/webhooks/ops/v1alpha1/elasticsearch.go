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
	"errors"
	"fmt"
	"strings"

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	"github.com/Masterminds/semver/v3"
	"gomodules.xyz/x/arrays"
	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/mergepatch"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
	meta_util "kmodules.xyz/client-go/meta"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// SetupElasticsearchOpsRequestWebhookWithManager registers the webhook for ElasticsearchOpsRequest in the manager.
func SetupElasticsearchOpsRequestWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&opsapi.ElasticsearchOpsRequest{}).
		WithValidator(&ElasticsearchOpsRequestCustomWebhook{mgr.GetClient()}).
		Complete()
}

type ElasticsearchOpsRequestCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var esLog = logf.Log.WithName("Elasticsearch-opsrequest")

var _ webhook.CustomValidator = &ElasticsearchOpsRequestCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *ElasticsearchOpsRequestCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	ops, ok := obj.(*opsapi.ElasticsearchOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an ElasticsearchOpsRequest object but got %T", obj)
	}
	esLog.Info("validate create", "name", ops.Name)
	return nil, w.validateCreateOrUpdate(ops)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *ElasticsearchOpsRequestCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	ops, ok := newObj.(*opsapi.ElasticsearchOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an ElasticsearchOpsRequest object but got %T", newObj)
	}
	esLog.Info("validate update", "name", ops.Name)

	oldOps, ok := oldObj.(*opsapi.ElasticsearchOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an ElasticsearchOpsRequest object but got %T", oldObj)
	}

	if err := validateElasticsearchOpsRequest(ops, oldOps); err != nil {
		return nil, err
	}

	if err := w.validateCreateOrUpdate(ops); err != nil {
		return nil, err
	}

	if isOpsReqCompleted(ops.Status.Phase) && !isOpsReqCompleted(oldOps.Status.Phase) { // just completed
		var db dbapi.Elasticsearch
		err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: ops.Spec.DatabaseRef.Name, Namespace: ops.Namespace}, &db)
		if err != nil {
			return nil, err
		}
		return nil, resumeDatabase(w.DefaultClient, &db)
	}
	return nil, nil
}

func (w *ElasticsearchOpsRequestCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func validateElasticsearchOpsRequest(obj, oldObj runtime.Object) error {
	preconditions := meta_util.PreConditionSet{Set: sets.New[string]("spec")}
	_, err := meta_util.CreateStrategicPatch(oldObj, obj, preconditions.PreconditionFunc()...)
	if err != nil {
		if mergepatch.IsPreconditionFailed(err) {
			return fmt.Errorf("%v.%v", err, preconditions.Error())
		}
		return err
	}
	return nil
}

func (w *ElasticsearchOpsRequestCustomWebhook) validateOpensearchVersionCompatibility(req *opsapi.ElasticsearchOpsRequest, db *dbapi.Elasticsearch) (bool, error) {
	if req.Spec.Type != opsapi.ElasticsearchOpsRequestTypeReconfigure && req.Spec.Type != opsapi.ElasticsearchOpsRequestTypeRotateAuth {
		return false, nil
	}
	esversion := &catalog.ElasticsearchVersion{}
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name: db.Spec.Version,
	}, esversion)
	if err != nil {
		return true, err
	}
	version, err := semver.NewVersion(esversion.Spec.Version)
	if err != nil {
		return true, err
	}

	if !db.Spec.EnableSSL && esversion.Spec.AuthPlugin == catalog.ElasticsearchAuthPluginOpenSearch && version.Major() > 1 {
		return true, nil
	}

	return false, nil
}

func (w *ElasticsearchOpsRequestCustomWebhook) validateCreateOrUpdate(req *opsapi.ElasticsearchOpsRequest) error {
	if validType, _ := arrays.Contains(opsapi.ElasticsearchOpsRequestTypeNames(), string(req.Spec.Type)); !validType {
		return field.Invalid(field.NewPath("spec").Child("type"), req.Name,
			fmt.Sprintf("defined OpsRequestType %s is not supported, supported types for Elasticsearch are %s", req.Spec.Type, strings.Join(opsapi.ElasticsearchOpsRequestTypeNames(), ", ")))
	}
	var allErr field.ErrorList
	db := &dbapi.Elasticsearch{}
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      req.GetDBRefName(),
		Namespace: req.GetNamespace(),
	}, db)
	if err != nil && !apierrors.IsNotFound(err) {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("databaseRef"), req.Name,
			fmt.Sprintf("referenced database %s/%s is not found", req.Namespace, req.Spec.DatabaseRef.Name)))
	}
	switch opsapi.ElasticsearchOpsRequestType(req.GetRequestType()) {
	case opsapi.ElasticsearchOpsRequestTypeUpdateVersion:
		if err := w.validateElasticsearchUpdateVersionOpsRequest(req, db); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("updateVersion"),
				req.Name,
				err.Error()))
		}
	case opsapi.ElasticsearchOpsRequestTypeReconfigure:
		if err := w.validateElasticsearchReconfigureOpsRequest(req, db); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("configuration"),
				req.Name,
				err.Error()))
		}
	case opsapi.ElasticsearchOpsRequestTypeRotateAuth:
		if err := w.validateElasticsearchRotateAuthenticationOpsRequest(req, db); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("authentication"),
				req.Name,
				err.Error()))
		}
	}

	if len(allErr) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "elasticsearchopsrequests.kubedb.com", Kind: "ElasticsearchOpsRequest"}, req.Name, allErr)
}

func (w *ElasticsearchOpsRequestCustomWebhook) validateElasticsearchUpdateVersionOpsRequest(req *opsapi.ElasticsearchOpsRequest, db *dbapi.Elasticsearch) error {
	updateVersionSpec := req.Spec.UpdateVersion
	if updateVersionSpec == nil {
		return errors.New("spec.updateVersion nil not supported in UpdateVersion type")
	}

	yes, err := IsUpgradable(w.DefaultClient, catalog.ResourceKindElasticsearchVersion, db.Spec.Version, updateVersionSpec.TargetVersion)
	if err != nil {
		return err
	}
	if !yes {
		return fmt.Errorf("upgrade from version %v to %v is not supported", db.Spec.Version, req.Spec.UpdateVersion.TargetVersion)
	}

	return nil
}

func (w *ElasticsearchOpsRequestCustomWebhook) validateElasticsearchRotateAuthenticationOpsRequest(req *opsapi.ElasticsearchOpsRequest, db *dbapi.Elasticsearch) error {
	issue, err := w.validateOpensearchVersionCompatibility(req, db)
	if err != nil {
		return err
	}
	if issue {
		return fmt.Errorf("opsrequest %s/%s is not compatible for version %s", req.Namespace, req.Name, db.Spec.Version)
	}
	authSpec := req.Spec.Authentication
	if authSpec != nil && authSpec.SecretRef != nil {
		if authSpec.SecretRef.Name == "" {
			return errors.New("spec.authentication.secretRef.name can not be empty")
		}
		err = w.DefaultClient.Get(context.TODO(), types.NamespacedName{
			Name:      authSpec.SecretRef.Name,
			Namespace: req.Namespace,
		}, &core.Secret{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return fmt.Errorf("referenced secret %s not found", authSpec.SecretRef.Name)
			}
			return err
		}
	}

	return nil
}

func (w *ElasticsearchOpsRequestCustomWebhook) validateElasticsearchReconfigureOpsRequest(req *opsapi.ElasticsearchOpsRequest, db *dbapi.Elasticsearch) error {
	issue, err := w.validateOpensearchVersionCompatibility(req, db)
	if err != nil {
		return err
	}
	if issue {
		return fmt.Errorf("opsrequest %s/%s is not compatible for version %s", req.Namespace, req.Name, db.Spec.Version)
	}
	configuration := req.Spec.Configuration
	if configuration == nil {
		return fmt.Errorf("configuration can not be empty for %s/%s", req.Namespace, req.Name)
	}

	return nil
}
