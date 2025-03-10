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

	kafkapi "kubedb.dev/apimachinery/apis/kafka/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	coreutil "kmodules.xyz/client-go/core/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// SetupSchemaRegistryWebhookWithManager registers the webhook for Kafka SchemaRegistry in the manager.
func SetupSchemaRegistryWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&kafkapi.SchemaRegistry{}).
		WithValidator(&SchemaRegistryCustomWebhook{mgr.GetClient()}).
		WithDefaulter(&SchemaRegistryCustomWebhook{mgr.GetClient()}).
		Complete()
}

type SchemaRegistryCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var schemaregistrylog = logf.Log.WithName("schemaregistry-resource")

var _ webhook.CustomDefaulter = &SchemaRegistryCustomWebhook{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (k *SchemaRegistryCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	sr, ok := obj.(*kafkapi.SchemaRegistry)
	if !ok {
		return fmt.Errorf("expected an schema-registry object but got %T", obj)
	}
	schemaregistrylog.Info("default", "name", sr.Name)
	sr.SetDefaults()
	return nil
}

var _ webhook.CustomValidator = &SchemaRegistryCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (k *SchemaRegistryCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	sr, ok := obj.(*kafkapi.SchemaRegistry)
	if !ok {
		return nil, fmt.Errorf("expected an SchemaRegistry object but got %T", obj)
	}

	schemaregistrylog.Info("validate create", "name", sr.Name)
	allErr := k.ValidateCreateOrUpdate(sr)
	if len(allErr) == 0 {
		return nil, nil
	}
	return nil, apierrors.NewInvalid(schema.GroupKind{Group: "kafka.kubedb.com", Kind: "SchemaRegistry"}, sr.Name, allErr)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (k *SchemaRegistryCustomWebhook) ValidateUpdate(ctx context.Context, old, newObj runtime.Object) (admission.Warnings, error) {
	sr, ok := newObj.(*kafkapi.SchemaRegistry)
	if !ok {
		return nil, fmt.Errorf("expected an SchemaRegistry object but got %T", newObj)
	}

	schemaregistrylog.Info("validate update", "name", sr.Name)

	oldRegistry := old.(*kafkapi.SchemaRegistry)
	allErr := k.ValidateCreateOrUpdate(sr)

	if *oldRegistry.Spec.Replicas == 1 && *sr.Spec.Replicas > 1 {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("replicas"),
			sr.Name,
			"Cannot scale up from 1 to more than 1 in standalone mode"))
	}

	if len(allErr) == 0 {
		return nil, nil
	}
	return nil, apierrors.NewInvalid(schema.GroupKind{Group: "kafka.kubedb.com", Kind: "SchemaRegistry"}, sr.Name, allErr)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (k *SchemaRegistryCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	sr, ok := obj.(*kafkapi.SchemaRegistry)
	if !ok {
		return nil, fmt.Errorf("expected an SchemaRegistry object but got %T", obj)
	}

	schemaregistrylog.Info("validate delete", "name", sr.Name)

	var allErr field.ErrorList
	if sr.Spec.DeletionPolicy == dbapi.DeletionPolicyDoNotTerminate {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("deletionPolicy"),
			sr.Name,
			"Can not delete as deletionPolicy is set to \"DoNotTerminate\""))
		return nil, apierrors.NewInvalid(schema.GroupKind{Group: "kafka.kubedb.com", Kind: "SchemaRegistry"}, sr.Name, allErr)
	}
	return nil, nil
}

func (k *SchemaRegistryCustomWebhook) ValidateCreateOrUpdate(sr *kafkapi.SchemaRegistry) field.ErrorList {
	var allErr field.ErrorList

	if sr.Spec.DeletionPolicy == dbapi.DeletionPolicyHalt {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("deletionPolicy"),
			sr.Name,
			"DeletionPolicyHalt is not supported for SchemaRegistry"))
	}

	// number of replicas can not be 0 or less
	if sr.Spec.Replicas != nil && *sr.Spec.Replicas <= 0 {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("replicas"),
			sr.Name,
			"number of replicas can not be 0 or less"))
	}

	err := k.validateVersion(sr)
	if err != nil {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("version"),
			sr.Name,
			err.Error()))
	}

	err = k.validateVolumes(sr)
	if err != nil {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("podTemplate").Child("spec").Child("volumes"),
			sr.Name,
			err.Error()))
	}

	err = k.validateContainerVolumeMountPaths(sr)
	if err != nil {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("podTemplate").Child("spec").Child("containers").Child("volumeMounts"),
			sr.Name,
			err.Error()))
	}

	err = k.validateEnvVars()
	if err != nil {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("podTemplate").Child("spec").Child("containers").Child("envs"),
			sr.Name,
			err.Error()))
	}

	if len(allErr) == 0 {
		return nil
	}
	return allErr
}

func (k *SchemaRegistryCustomWebhook) validateEnvVars() error {
	return nil
}

func (k *SchemaRegistryCustomWebhook) validateVersion(sr *kafkapi.SchemaRegistry) error {
	ksrVersion := &catalog.SchemaRegistryVersion{}
	err := kafkapi.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: sr.Spec.Version}, ksrVersion)
	if err != nil {
		return errors.New("version not supported")
	}
	return nil
}

var schemaRegistryReservedVolumes = []string{
	kafkapi.KafkaClientCertVolumeName,
}

func (k *SchemaRegistryCustomWebhook) validateVolumes(sr *kafkapi.SchemaRegistry) error {
	if sr.Spec.PodTemplate.Spec.Volumes == nil {
		return nil
	}
	rsv := make([]string, len(schemaRegistryReservedVolumes))
	copy(rsv, schemaRegistryReservedVolumes)
	volumes := sr.Spec.PodTemplate.Spec.Volumes
	for _, rv := range rsv {
		for _, ugv := range volumes {
			if ugv.Name == rv {
				return errors.New("Cannot use a reserve volume name: " + rv)
			}
		}
	}
	return nil
}

var schemaRegistryReservedVolumeMountPaths = []string{
	kafkapi.KafkaClientCertDir,
}

func (k *SchemaRegistryCustomWebhook) validateContainerVolumeMountPaths(sr *kafkapi.SchemaRegistry) error {
	container := coreutil.GetContainerByName(sr.Spec.PodTemplate.Spec.Containers, kafkapi.SchemaRegistryContainerName)
	if container == nil {
		return errors.New("container not found")
	}
	rPaths := schemaRegistryReservedVolumeMountPaths
	volumeMountPaths := container.VolumeMounts
	for _, rvm := range rPaths {
		for _, ugv := range volumeMountPaths {
			if ugv.MountPath == rvm {
				return errors.New("Cannot use a reserve volume mount path: " + rvm)
			}
		}
	}
	return nil
}
