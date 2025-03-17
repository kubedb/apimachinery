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

	catalogapi "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	kerr "k8s.io/apimachinery/pkg/api/errors"
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

// SetupKafkaOpsRequestWebhookWithManager registers the webhook for KafkaOpsRequest in the manager.
func SetupKafkaOpsRequestWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&opsapi.KafkaOpsRequest{}).
		WithValidator(&KafkaOpsRequestCustomWebhook{mgr.GetClient()}).
		Complete()
}

type KafkaOpsRequestCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var kafkaLog = logf.Log.WithName("kafka-opsrequest")

var _ webhook.CustomValidator = &KafkaOpsRequestCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (in *KafkaOpsRequestCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	ops, ok := obj.(*opsapi.KafkaOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an KafkaOpsRequest object but got %T", obj)
	}
	kafkaLog.Info("validate create", "name", ops.Name)
	return nil, in.validateCreateOrUpdate(ops)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *KafkaOpsRequestCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	ops, ok := newObj.(*opsapi.KafkaOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an KafkaOpsRequest object but got %T", newObj)
	}
	kafkaLog.Info("validate update", "name", ops.Name)

	oldOps, ok := oldObj.(*opsapi.KafkaOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an KafkaOpsRequest object but got %T", oldObj)
	}

	if err := validateKafkaOpsRequest(ops, oldOps); err != nil {
		return nil, err
	}
	return nil, in.validateCreateOrUpdate(ops)
}

func (in *KafkaOpsRequestCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	_, ok := obj.(*opsapi.KafkaOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an KafkaOpsRequest object but got %T", obj)
	}
	return nil, nil
}

func validateKafkaOpsRequest(req *opsapi.KafkaOpsRequest, oldReq *opsapi.KafkaOpsRequest) error {
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

func (k *KafkaOpsRequestCustomWebhook) validateCreateOrUpdate(req *opsapi.KafkaOpsRequest) error {
	var allErr field.ErrorList
	switch req.GetRequestType().(opsapi.KafkaOpsRequestType) {
	case opsapi.KafkaOpsRequestTypeRestart:
		if _, err := k.hasDatabaseRef(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("restart"),
				req.Name,
				err.Error()))
		}
	case opsapi.KafkaOpsRequestTypeUpdateVersion:
		if err := k.validateKafkaUpdateVersionOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("updateVersion"),
				req.Name,
				err.Error()))
		}
	case opsapi.KafkaOpsRequestTypeHorizontalScaling:
		if err := k.validateKafkaHorizontalScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("horizontalScaling"),
				req.Name,
				err.Error()))
		}
	case opsapi.KafkaOpsRequestTypeVerticalScaling:
		if err := k.validateKafkaVerticalScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("verticalScaling"),
				req.Name,
				err.Error()))
		}
	case opsapi.KafkaOpsRequestTypeVolumeExpansion:
		if err := k.validateKafkaVolumeExpansionOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("volumeExpansion"),
				req.Name,
				err.Error()))
		}
	case opsapi.KafkaOpsRequestTypeReconfigure:
		if err := k.validateKafkaReconfigurationOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("configuration"),
				req.Name,
				err.Error()))
		}
	case opsapi.KafkaOpsRequestTypeReconfigureTLS:
		if err := k.validateKafkaReconfigurationTLSOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("tls"),
				req.Name,
				err.Error()))
		}
	case opsapi.KafkaOpsRequestTypeRotateAuth:
		if err := k.validateKafkaRotateAuthenticationOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("authentication"),
				req.Name,
				err.Error()))
		}

	default:
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("type"), req.Name,
			fmt.Sprintf("defined OpsRequestType %s is not supported, supported types for kafka are %s", req.Spec.Type, strings.Join(opsapi.KafkaOpsRequestTypeNames(), ", "))))
	}
	if len(allErr) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "kafkaopsrequests.kubedb.com", Kind: "KafkaOpsRequest"}, req.Name, allErr)
}

func (k *KafkaOpsRequestCustomWebhook) hasDatabaseRef(req *opsapi.KafkaOpsRequest) (*dbapi.Kafka, error) {
	kafka := dbapi.Kafka{}
	if err := k.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      req.GetDBRefName(),
		Namespace: req.GetNamespace(),
	}, &kafka); err != nil {
		return nil, errors.New(fmt.Sprintf("spec.databaseRef %s/%s, is invalid or not found", req.GetNamespace(), req.GetDBRefName()))
	}
	return &kafka, nil
}

func (k *KafkaOpsRequestCustomWebhook) validateKafkaUpdateVersionOpsRequest(req *opsapi.KafkaOpsRequest) error {
	updateVersionSpec := req.Spec.UpdateVersion
	if updateVersionSpec == nil {
		return errors.New("spec.updateVersion nil not supported in UpdateVersion type")
	}

	if _, err := k.hasDatabaseRef(req); err != nil {
		return err
	}

	nextKafkaVersion := catalogapi.KafkaVersion{}
	err := k.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      req.Spec.UpdateVersion.TargetVersion,
		Namespace: req.GetNamespace(),
	}, &nextKafkaVersion)
	if err != nil {
		return errors.New(fmt.Sprintf("spec.updateVersion.targetVersion - %s, is not supported!", req.Spec.UpdateVersion.TargetVersion))
	}
	// check if nextKafkaVersion is deprecated.if deprecated, return error
	if nextKafkaVersion.Spec.Deprecated {
		return errors.New(fmt.Sprintf("spec.updateVersion.targetVersion - %s, is depricated!", req.Spec.UpdateVersion.TargetVersion))
	}
	return nil
}

func (k *KafkaOpsRequestCustomWebhook) validateKafkaHorizontalScalingOpsRequest(req *opsapi.KafkaOpsRequest) error {
	horizontalScalingSpec := req.Spec.HorizontalScaling
	if horizontalScalingSpec == nil {
		return errors.New("spec.horizontalScaling nil not supported in HorizontalScaling type")
	}
	kafka, err := k.hasDatabaseRef(req)
	if err != nil {
		return err
	}

	if horizontalScalingSpec.Topology != nil && horizontalScalingSpec.Node != nil {
		return errors.New("spec.horizontalScaling.Node && spec.horizontalScaling.Topology both can't be non-empty at the same ops request")
	}
	if kafka.Spec.Topology != nil && horizontalScalingSpec.Topology == nil {
		return errors.New("spec.horizontalScaling.topology can not be empty as reference database mode is topology")
	}
	if kafka.Spec.Topology == nil && horizontalScalingSpec.Node == nil {
		return errors.New("spec.horizontalScaling.node can not be empty as reference database mode is combined")
	}

	if horizontalScalingSpec.Topology != nil {
		if horizontalScalingSpec.Topology.Broker != nil && *horizontalScalingSpec.Topology.Broker <= 0 {
			return errors.New("spec.horizontalScaling.topology.broker must be positive")
		}
		if horizontalScalingSpec.Topology.Controller != nil && *horizontalScalingSpec.Topology.Controller <= 0 {
			return errors.New("spec.horizontalScaling.topology.controller must be positive")
		}
	} else {
		if *horizontalScalingSpec.Node <= 0 {
			return errors.New("spec.horizontalScaling.node must be positive")
		}
	}

	return nil
}

func (k *KafkaOpsRequestCustomWebhook) validateKafkaVerticalScalingOpsRequest(req *opsapi.KafkaOpsRequest) error {
	verticalScalingSpec := req.Spec.VerticalScaling
	if verticalScalingSpec == nil {
		return errors.New("spec.verticalScaling nil not supported in VerticalScaling type")
	}
	kafka, err := k.hasDatabaseRef(req)
	if err != nil {
		return err
	}
	if (verticalScalingSpec.Broker != nil || verticalScalingSpec.Controller != nil) && verticalScalingSpec.Node != nil {
		return errors.New("spec.verticalScaling.Node && spec.verticalScaling.Topology both can't be non-empty at the same ops request")
	}
	if kafka.Spec.Topology != nil && verticalScalingSpec.Broker == nil && verticalScalingSpec.Controller == nil {
		return errors.New("spec.verticalScaling.topology can not be empty as reference database mode is topology")
	}
	if kafka.Spec.Topology == nil && verticalScalingSpec.Node == nil {
		return errors.New("spec.verticalScaling.node can not be empty as reference database mode is combined")
	}

	return nil
}

func (k *KafkaOpsRequestCustomWebhook) validateKafkaVolumeExpansionOpsRequest(req *opsapi.KafkaOpsRequest) error {
	volumeExpansionSpec := req.Spec.VolumeExpansion
	if volumeExpansionSpec == nil {
		return errors.New("spec.volumeExpansion nil not supported in VolumeExpansion type")
	}
	kafka, err := k.hasDatabaseRef(req)
	if err != nil {
		return err
	}
	if (volumeExpansionSpec.Broker != nil || volumeExpansionSpec.Controller != nil) && volumeExpansionSpec.Node != nil {
		return errors.New("spec.volumeExpansion.Node && spec.volumeExpansion.Topology both can't be non-empty at the same ops request")
	}
	if kafka.Spec.Topology != nil && volumeExpansionSpec.Broker == nil && volumeExpansionSpec.Controller == nil {
		return errors.New("spec.volumeExpansion.topology can not be empty as reference database mode is topology")
	}
	if kafka.Spec.Topology == nil && volumeExpansionSpec.Node == nil {
		return errors.New("spec.volumeExpansion.node can not be empty as reference database mode is combined")
	}

	return nil
}

func (k *KafkaOpsRequestCustomWebhook) validateKafkaReconfigurationOpsRequest(req *opsapi.KafkaOpsRequest) error {
	configurationSpec := req.Spec.Configuration
	if configurationSpec == nil {
		return errors.New("spec.configuration nil not supported in Reconfigure type")
	}
	if _, err := k.hasDatabaseRef(req); err != nil {
		return err
	}

	if configurationSpec.RemoveCustomConfig && (configurationSpec.ConfigSecret != nil || len(configurationSpec.ApplyConfig) != 0) {
		return errors.New("at a time one configuration is allowed to run one operation(`RemoveCustomConfig` or `ConfigSecret with or without ApplyConfig`) to reconfigure")
	}
	return nil
}

func (k *KafkaOpsRequestCustomWebhook) validateKafkaReconfigurationTLSOpsRequest(req *opsapi.KafkaOpsRequest) error {
	TLSSpec := req.Spec.TLS
	if TLSSpec == nil {
		return errors.New("spec.TLS nil not supported in ReconfigureTLS type")
	}
	if _, err := k.hasDatabaseRef(req); err != nil {
		return err
	}
	configCount := 0
	if req.Spec.TLS.Remove {
		configCount++
	}
	if req.Spec.TLS.RotateCertificates {
		configCount++
	}
	if req.Spec.TLS.TLSConfig.IssuerRef != nil || req.Spec.TLS.TLSConfig.Certificates != nil {
		configCount++
	}

	if configCount == 0 {
		return errors.New("No reconfiguration is provided in TLS Spec.")
	}

	if configCount > 1 {
		return errors.New("more than 1 field have assigned to spec.reconfigureTLS but at a time one is allowed to run one operation")
	}
	return nil
}

func (k *KafkaOpsRequestCustomWebhook) validateKafkaRotateAuthenticationOpsRequest(req *opsapi.KafkaOpsRequest) error {
	authSpec := req.Spec.Authentication
	if _, err := k.hasDatabaseRef(req); err != nil {
		return err
	}
	if authSpec != nil && authSpec.SecretRef != nil {
		if authSpec.SecretRef.Name == "" {
			return errors.New("spec.authentication.secretRef.name can not be empty")
		}
		err := k.DefaultClient.Get(context.TODO(), types.NamespacedName{
			Name:      authSpec.SecretRef.Name,
			Namespace: req.Namespace,
		}, &core.Secret{})
		if err != nil {
			if kerr.IsNotFound(err) {
				return fmt.Errorf("referenced secret %s not found", authSpec.SecretRef.Name)
			}
			return err
		}
	}

	return nil
}
