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
	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"

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

// SetupSolrOpsRequestWebhookWithManager registers the webhook for SolrOpsRequest in the manager.
func SetupSolrOpsRequestWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&opsapi.SolrOpsRequest{}).
		WithValidator(&SolrOpsRequestCustomWebhook{mgr.GetClient()}).
		Complete()
}

type SolrOpsRequestCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var slLog = logf.Log.WithName("Solr-opsrequest")

var _ webhook.CustomValidator = &SolrOpsRequestCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *SolrOpsRequestCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	ops, ok := obj.(*opsapi.SolrOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an SolrOpsRequest object but got %T", obj)
	}
	slLog.Info("validate create", "name", ops.Name)
	return nil, w.validateCreateOrUpdate(ops)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *SolrOpsRequestCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	ops, ok := newObj.(*opsapi.SolrOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an SolrOpsRequest object but got %T", newObj)
	}
	slLog.Info("validate update", "name", ops.Name)

	oldOps, ok := oldObj.(*opsapi.SolrOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an SolrOpsRequest object but got %T", oldObj)
	}

	if err := validateSolrOpsRequest(ops, oldOps); err != nil {
		return nil, err
	}
	return nil, w.validateCreateOrUpdate(ops)
}

func (w *SolrOpsRequestCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func validateSolrOpsRequest(req *opsapi.SolrOpsRequest, oldReq *opsapi.SolrOpsRequest) error {
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

func (w *SolrOpsRequestCustomWebhook) validateCreateOrUpdate(req *opsapi.SolrOpsRequest) error {
	if validType, _ := arrays.Contains(opsapi.SolrOpsRequestTypeNames(), string(req.Spec.Type)); !validType {
		return field.Invalid(field.NewPath("spec").Child("type"), req.Name,
			fmt.Sprintf("defined OpsRequestType %s is not supported, supported types for Solr are %s", req.Spec.Type, strings.Join(opsapi.SolrOpsRequestTypeNames(), ", ")))
	}
	db, err := w.hasDatabaseRef(req)
	if err != nil {
		return err
	}

	var allErr field.ErrorList
	switch opsapi.SolrOpsRequestType(req.GetRequestType()) {
	case opsapi.SolrOpsRequestTypeRestart:

	case opsapi.SolrOpsRequestTypeHorizontalScaling:
		if err := w.validateSolrHorizontalScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("type").Child("HorizontalScaling"),
				req.Name,
				err.Error()))
		}
	case opsapi.SolrOpsRequestTypeReconfigure:

	case opsapi.SolrOpsRequestTypeUpdateVersion:
		if err := w.validateSolrUpdateVersionOpsRequest(db, req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("updateVersion"),
				req.Name,
				err.Error()))
		}
	case opsapi.SolrOpsRequestTypeVerticalScaling:
		if err := w.validateSolrVerticalScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("verticalScaling"),
				req.Name,
				err.Error()))
		}
	case opsapi.SolrOpsRequestTypeVolumeExpansion:
		if err := w.validateSolrVolumeExpansionOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("volumeExpansion"),
				req.Name,
				err.Error()))
		}
	case opsapi.SolrOpsRequestTypeReconfigureTLS:
		if err := w.validateSolrReconfigureTLSOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("type").Child("ReconfigureTLS"),
				req.Name,
				err.Error()))
		}
	case opsapi.SolrOpsRequestTypeRotateAuth:
		if err := w.validateSolrRotateAuthenticationOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("authentication"),
				req.Name,
				err.Error()))
		}
	}

	if len(allErr) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "solropsrequests.kubedb.com", Kind: "SolrOpsRequest"}, req.Name, allErr)
}

func (w *SolrOpsRequestCustomWebhook) hasDatabaseRef(req *opsapi.SolrOpsRequest) (*olddbapi.Solr, error) {
	db := &olddbapi.Solr{}
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      req.GetDBRefName(),
		Namespace: req.GetNamespace(),
	}, db)
	if err != nil {
		return nil, fmt.Errorf("spec.databaseRef %s/%s, is invalid or not found", req.GetNamespace(), req.GetDBRefName())
	}

	return db, nil
}

func (w *SolrOpsRequestCustomWebhook) validateSolrVerticalScalingOpsRequest(req *opsapi.SolrOpsRequest) error {
	verticalScalingSpec := req.Spec.VerticalScaling
	if verticalScalingSpec == nil {
		return errors.New("spec.verticalScaling nil not supported in VerticalScaling type")
	}

	if verticalScalingSpec.Node != nil && (verticalScalingSpec.Data != nil || verticalScalingSpec.Overseer != nil || verticalScalingSpec.Coordinator != nil) {
		return errors.New("spec.verticalScaling.Node && spec.verticalScaling.Topology both can't be non-empty at the same ops request")
	}

	return nil
}

func (w *SolrOpsRequestCustomWebhook) validateSolrHorizontalScalingOpsRequest(req *opsapi.SolrOpsRequest) error {
	horizontalScalingSpec := req.Spec.HorizontalScaling
	if horizontalScalingSpec == nil {
		return errors.New("spec.horizontalScaling nil not supported in VerticalScaling type")
	}

	if horizontalScalingSpec.Node != nil && (horizontalScalingSpec.Data != nil || horizontalScalingSpec.Overseer != nil || horizontalScalingSpec.Coordinator != nil) {
		return errors.New("spec.horizontalScalingSpec.Node && spec.horizontalScalingSpec.Topology both can't be non-empty at the same ops request")
	}

	return nil
}

func (w *SolrOpsRequestCustomWebhook) validateSolrVolumeExpansionOpsRequest(req *opsapi.SolrOpsRequest) error {
	volumeExpansionSpec := req.Spec.VolumeExpansion
	if volumeExpansionSpec == nil {
		return errors.New("spec.volumeExpansion nil not supported in VolumeExpansion type")
	}

	if volumeExpansionSpec.Node != nil && (volumeExpansionSpec.Data != nil || volumeExpansionSpec.Overseer != nil || volumeExpansionSpec.Coordinator != nil) {
		return errors.New("spec.volumeExpansion.Node && spec.volumeExpansion.Topology both can't be non-empty at the same ops request")
	}

	return nil
}

func (w *SolrOpsRequestCustomWebhook) validateSolrReconfigureTLSOpsRequest(req *opsapi.SolrOpsRequest) error {
	TLSSpec := req.Spec.TLS
	if TLSSpec == nil {
		return errors.New("spec.TLS nil not supported in ReconfigureTLS type")
	}

	configCount := 0
	if req.Spec.TLS.Remove {
		configCount++
	}
	if req.Spec.TLS.RotateCertificates {
		configCount++
	}
	if req.Spec.TLS.IssuerRef != nil || req.Spec.TLS.Certificates != nil {
		configCount++
	}

	if configCount == 0 {
		return errors.New("no reconfiguration is provided in TLS Spec")
	}

	if configCount > 1 {
		return errors.New("more than 1 field have assigned to spec.reconfigureTLS but at a time one is allowed to run one operation")
	}
	return nil
}

func (w *SolrOpsRequestCustomWebhook) validateSolrRotateAuthenticationOpsRequest(req *opsapi.SolrOpsRequest) error {
	authSpec := req.Spec.Authentication
	if authSpec != nil && authSpec.SecretRef != nil {
		if authSpec.SecretRef.Name == "" {
			return errors.New("spec.authentication.secretRef.name can not be empty")
		}
		err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
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

func (w *SolrOpsRequestCustomWebhook) validateSolrUpdateVersionOpsRequest(db *olddbapi.Solr, req *opsapi.SolrOpsRequest) error {
	updateVersionSpec := req.Spec.UpdateVersion
	if updateVersionSpec == nil {
		return errors.New("spec.updateVersion nil not supported in UpdateVersion type")
	}

	yes, err := IsUpgradable(w.DefaultClient, catalog.ResourceKindSolrVersion, db.Spec.Version, updateVersionSpec.TargetVersion)
	if err != nil {
		return err
	}
	if !yes {
		return fmt.Errorf("upgrade from version %v to %v is not supported", db.Spec.Version, req.Spec.UpdateVersion.TargetVersion)
	}

	return nil
}
