/*
￼Copyright AppsCode Inc. and Contributors
￼
￼Licensed under the Apache License, Version 2.0 (the "License");
￼you may not use this file except in compliance with the License.
￼You may obtain a copy of the License at
￼
￼    http://www.apache.org/licenses/LICENSE-2.0
￼
￼Unless required by applicable law or agreed to in writing, software
￼distributed under the License is distributed on an "AS IS" BASIS,
￼WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
￼See the License for the specific language governing permissions and
￼limitations under the License.
￼*/

package v1alpha1

import (
	"context"
	"fmt"
	"strings"

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	"github.com/pkg/errors"
	"gomodules.xyz/x/arrays"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// SetupSinglestoreOpsRequestWebhookWithManager registers the webhook for SinglestoreOpsRequest in the manager.
func SetupSinglestoreOpsRequestWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&opsapi.SinglestoreOpsRequest{}).
		WithValidator(&SinglestoreOpsRequestCustomWebhook{mgr.GetClient()}).
		Complete()
}

type SinglestoreOpsRequestCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var sdbLog = logf.Log.WithName("singlestore-opsrequest")

var _ webhook.CustomValidator = &SinglestoreOpsRequestCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *SinglestoreOpsRequestCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	ops, ok := obj.(*opsapi.SinglestoreOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an SinglestoreOpsRequest object but got %T", obj)
	}
	sdbLog.Info("validate create", "name", ops.Name)
	return nil, w.validateCreateOrUpdate(ops)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *SinglestoreOpsRequestCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	ops, ok := newObj.(*opsapi.SinglestoreOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an SinglestoreOpsRequest object but got %T", newObj)
	}
	sdbLog.Info("validate update", "name", ops.Name)

	oldOps, ok := oldObj.(*opsapi.SinglestoreOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an SinglestoreOpsRequest object but got %T", oldObj)
	}

	if err := validateSinglestoreOpsRequest(ops, oldOps); err != nil {
		return nil, err
	}
	return nil, w.validateCreateOrUpdate(ops)
}

func (w *SinglestoreOpsRequestCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func validateSinglestoreOpsRequest(req *opsapi.SinglestoreOpsRequest, oldReq *opsapi.SinglestoreOpsRequest) error {
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

func (s *SinglestoreOpsRequestCustomWebhook) validateCreateOrUpdate(req *opsapi.SinglestoreOpsRequest) error {
	if validType, _ := arrays.Contains(opsapi.SinglestoreOpsRequestTypeNames(), string(req.Spec.Type)); !validType {
		return field.Invalid(field.NewPath("spec").Child("type"), req.Name,
			fmt.Sprintf("defined OpsRequestType %s is not supported, supported types for Singlestore are %s", req.Spec.Type, strings.Join(opsapi.SinglestoreOpsRequestTypeNames(), ", ")))
	}

	var allErr field.ErrorList
	var db dbapi.Singlestore
	pp := &dbapi.Singlestore{ObjectMeta: metav1.ObjectMeta{Name: req.Spec.DatabaseRef.Name, Namespace: req.Namespace}}
	err := s.DefaultClient.Get(context.TODO(), client.ObjectKeyFromObject(pp), pp)
	if err != nil && apierrors.IsNotFound(err) {
		return fmt.Errorf("referenced database %s/%s is not found", req.Namespace, req.Spec.DatabaseRef.Name)
	}

	switch req.GetRequestType().(opsapi.SinglestoreOpsRequestType) {
	case opsapi.SinglestoreOpsRequestTypeRestart:
		if _, err := s.hasDatabaseRef(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("restart"),
				req.Name,
				err.Error()))
		}
	case opsapi.SinglestoreOpsRequestTypeVerticalScaling:
		if err := s.validateSinglestoreVerticalScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("verticalScaling"),
				req.Name,
				err.Error()))
		}
	case opsapi.SinglestoreOpsRequestTypeVolumeExpansion:
		if err := s.validateSinglestoreVolumeExpansionOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("volumeExpansion"),
				req.Name,
				err.Error()))
		}
	case opsapi.SinglestoreOpsRequestTypeReconfigure:
		if err := s.validateSinglestoreReconfigurationOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("configuration"),
				req.Name,
				err.Error()))
		}
	case opsapi.SinglestoreOpsRequestTypeReconfigureTLS:
		if err := s.validateSinglestoreReconfigurationTLSOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("tls"),
				req.Name,
				err.Error()))
		}
	case opsapi.SinglestoreOpsRequestTypeHorizontalScaling:
		if err := s.validateSinglestoreHorizontalScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("horizontalScaling"),
				req.Name,
				err.Error()))
		}
	case opsapi.SinglestoreOpsRequestTypeUpdateVersion:
		if err := s.validateSinglestoreUpdateVersionOpsRequest(&db, req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("updateVersion"),
				req.Name,
				err.Error()))
		}
	}

	if len(allErr) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "Singlestoreopsrequests.kubedb.com", Kind: "SinglestoreOpsRequest"}, req.Name, allErr)
}

func (s *SinglestoreOpsRequestCustomWebhook) hasDatabaseRef(req *opsapi.SinglestoreOpsRequest) (*dbapi.Singlestore, error) {
	sdb := dbapi.Singlestore{}
	if err := s.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      req.GetDBRefName(),
		Namespace: req.GetNamespace(),
	}, &sdb); err != nil {
		return &sdb, errors.New(fmt.Sprintf("spec.databaseRef %s/%s, is invalid or not found", req.GetNamespace(), req.GetDBRefName()))
	}
	return &sdb, nil
}

func (s *SinglestoreOpsRequestCustomWebhook) validateSinglestoreVerticalScalingOpsRequest(req *opsapi.SinglestoreOpsRequest) error {
	verticalScalingSpec := req.Spec.VerticalScaling
	if verticalScalingSpec == nil {
		return errors.New("spec.verticalScaling nil not supported in VerticalScaling type")
	}
	sdb, err := s.hasDatabaseRef(req)
	if err != nil {
		return err
	}
	if verticalScalingSpec.Node != nil && (sdb.Spec.Topology != nil || verticalScalingSpec.Aggregator != nil || verticalScalingSpec.Leaf != nil || verticalScalingSpec.Coordinator != nil) {
		return errors.New("spec.verticalScaling.node only allowed in standalone mode and other field not allowed in standalone mode")
	}
	if sdb.Spec.Topology != nil && verticalScalingSpec.Aggregator == nil && verticalScalingSpec.Leaf == nil && verticalScalingSpec.Coordinator == nil {
		return errors.New("In clustering mode you have to mention one of them except spec.verticalScaling.Node")
	}
	if sdb.Spec.Topology == nil && verticalScalingSpec.Node == nil {
		return errors.New("In standalone mod spec.verticalScaling.node can't be empty")
	}

	return nil
}

func (s *SinglestoreOpsRequestCustomWebhook) validateSinglestoreVolumeExpansionOpsRequest(req *opsapi.SinglestoreOpsRequest) error {
	volumeExpansionSpec := req.Spec.VolumeExpansion
	if volumeExpansionSpec == nil {
		return errors.New("spec.volumeExpansion nil not supported in VolumeExpansion type")
	}
	sdb, err := s.hasDatabaseRef(req)
	if err != nil {
		return err
	}
	if (volumeExpansionSpec.Aggregator != nil || volumeExpansionSpec.Leaf != nil) && volumeExpansionSpec.Node != nil {
		return errors.New("spec.volumeExpansion.Node && spec.volumeExpansion.Topology both can't be non-empty at the same ops request")
	}
	if sdb.Spec.Topology != nil && volumeExpansionSpec.Aggregator == nil && volumeExpansionSpec.Leaf == nil {
		return errors.New("spec.volumeExpansion.topology can not be empty as reference database mode is clustering")
	}
	if sdb.Spec.Topology == nil && volumeExpansionSpec.Node == nil {
		return errors.New("spec.volumeExpansion.node can not be empty as reference database mode is standalone")
	}

	return nil
}

func (s *SinglestoreOpsRequestCustomWebhook) validateSinglestoreReconfigurationOpsRequest(req *opsapi.SinglestoreOpsRequest) error {
	configurationSpec := req.Spec.Configuration
	if configurationSpec == nil {
		return errors.New("spec.configuration nil not supported in Reconfigure type")
	}
	sdb, err := s.hasDatabaseRef(req)
	if err != nil {
		return err
	}
	if (configurationSpec.Aggregator != nil || configurationSpec.Leaf != nil) && configurationSpec.Node != nil {
		return errors.New("spec.configuration.Node && spec.configuration.Topology both can't be non-empty at the same ops request")
	}
	if sdb.Spec.Topology != nil && configurationSpec.Aggregator == nil && configurationSpec.Leaf == nil {
		return errors.New("spec.configuration.topology can not be empty as reference database mode is clustering")
	}
	if sdb.Spec.Topology == nil && configurationSpec.Node == nil {
		return errors.New("spec.configuration.node can not be empty as reference database mode is standalone")
	}
	return nil
}

func (s *SinglestoreOpsRequestCustomWebhook) validateSinglestoreReconfigurationTLSOpsRequest(req *opsapi.SinglestoreOpsRequest) error {
	TLSSpec := req.Spec.TLS
	if TLSSpec == nil {
		return errors.New("spec.TLS nil not supported in ReconfigureTLS type")
	}
	sdb, err := s.hasDatabaseRef(req)
	if err != nil {
		return err
	}
	if sdb.Spec.Topology == nil {
		return errors.New("standalone mode not supported in ReconfigureTLS type")
	}
	if sdb.Spec.TLS == nil && (req.Spec.TLS.Remove || req.Spec.TLS.RotateCertificates) {
		return errors.New("can't apply remove and rotate certificates operation when db.Spec.TLS is nil.")
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
		return errors.New("no reconfiguration is provided in TLS spec")
	}

	if configCount > 1 {
		return errors.New("more than 1 field have assigned to spec.reconfigureTLS but at a time one is allowed to run one operation")
	}
	return nil
}

func (s *SinglestoreOpsRequestCustomWebhook) validateSinglestoreHorizontalScalingOpsRequest(req *opsapi.SinglestoreOpsRequest) error {
	horizontalScalingSpec := req.Spec.HorizontalScaling
	if horizontalScalingSpec == nil {
		return errors.New("spec.horizontalScaling nil not supported in HorizontalScaling type")
	}
	sdb, err := s.hasDatabaseRef(req)
	if err != nil {
		return err
	}
	if sdb.Spec.Topology == nil {
		return errors.New("standalone mode not supported in HorizontalScaling type")
	}
	if horizontalScalingSpec.Aggregator != nil && *horizontalScalingSpec.Aggregator <= 0 {
		return errors.New("spec.horizontalScaling.aggregator must be positive")
	}
	if horizontalScalingSpec.Leaf != nil && *horizontalScalingSpec.Leaf <= 0 {
		return errors.New("spec.horizontalScaling.leaf must be positive")
	}
	return nil
}

func (s *SinglestoreOpsRequestCustomWebhook) validateSinglestoreUpdateVersionOpsRequest(db *dbapi.Singlestore, req *opsapi.SinglestoreOpsRequest) error {
	// right now, kubeDB support the following singlestore version: 8.1.32, 8.5.7
	updateVersionSpec := req.Spec.UpdateVersion
	if updateVersionSpec == nil {
		return errors.New("`spec.updateVersion` nil not supported in UpdateVersion type")
	}

	yes, err := IsUpgradable(s.DefaultClient, catalog.ResourceKindSinlestoreVersion, db.Spec.Version, updateVersionSpec.TargetVersion)
	if err != nil {
		return err
	}
	if !yes {
		return fmt.Errorf("upgrade from version %v to %v is not supported", db.Spec.Version, req.Spec.UpdateVersion.TargetVersion)
	}

	if _, err := s.hasDatabaseRef(req); err != nil {
		return err
	}

	nextSinglestoreVersion := catalog.SinglestoreVersion{}
	err = s.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      req.Spec.UpdateVersion.TargetVersion,
		Namespace: req.GetNamespace(),
	}, &nextSinglestoreVersion)
	if err != nil {
		return errors.New(fmt.Sprintf("spec.updateVersion.targetVersion - %s, is not supported!", req.Spec.UpdateVersion.TargetVersion))
	}
	// check if nextKafkaVersion is deprecated.if deprecated, return error
	if nextSinglestoreVersion.Spec.Deprecated {
		return errors.New(fmt.Sprintf("spec.updateVersion.targetVersion - %s, is depricated!", req.Spec.UpdateVersion.TargetVersion))
	}
	return nil
}
