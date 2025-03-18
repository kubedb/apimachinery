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

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	kafkapi "kubedb.dev/apimachinery/apis/kafka/v1alpha1"
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	coreutil "kmodules.xyz/client-go/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// SetupRestProxyWebhookWithManager registers the webhook for Kafka RestProxy in the manager.
func SetupRestProxyWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&kafkapi.RestProxy{}).
		WithValidator(&RestProxyCustomWebhook{mgr.GetClient()}).
		WithDefaulter(&RestProxyCustomWebhook{mgr.GetClient()}).
		Complete()
}

type RestProxyCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var restproxylog = logf.Log.WithName("restproxy-resource")

var _ webhook.CustomDefaulter = &RestProxyCustomWebhook{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (k *RestProxyCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	rp, ok := obj.(*kafkapi.RestProxy)
	if !ok {
		return fmt.Errorf("expected an RestProxy object but got %T", obj)
	}
	restproxylog.Info("default", "name", rp.Name)
	rp.SetDefaults(k.DefaultClient)
	return nil
}

var _ webhook.CustomValidator = &RestProxyCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (k *RestProxyCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	rp, ok := obj.(*kafkapi.RestProxy)
	if !ok {
		return nil, fmt.Errorf("expected an RestProxy object but got %T", obj)
	}

	restproxylog.Info("validate create", "name", rp.Name)
	allErr := k.ValidateCreateOrUpdate(rp)
	if len(allErr) == 0 {
		return nil, nil
	}
	return nil, apierrors.NewInvalid(schema.GroupKind{Group: "kafka.kubedb.com", Kind: "RestProxy"}, rp.Name, allErr)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (k *RestProxyCustomWebhook) ValidateUpdate(ctx context.Context, old, newObj runtime.Object) (admission.Warnings, error) {
	rp, ok := newObj.(*kafkapi.RestProxy)
	if !ok {
		return nil, fmt.Errorf("expected an RestProxy object but got %T", newObj)
	}

	restproxylog.Info("validate update", "name", rp.Name)
	allErr := k.ValidateCreateOrUpdate(rp)
	if len(allErr) == 0 {
		return nil, nil
	}
	return nil, apierrors.NewInvalid(schema.GroupKind{Group: "kafka.kubedb.com", Kind: "RestProxy"}, rp.Name, allErr)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (k *RestProxyCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	rp, ok := obj.(*kafkapi.RestProxy)
	if !ok {
		return nil, fmt.Errorf("expected an RestProxy object but got %T", obj)
	}

	restproxylog.Info("validate delete", "name", rp.Name)

	var allErr field.ErrorList
	if rp.Spec.DeletionPolicy == dbapi.DeletionPolicyDoNotTerminate {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("deletionPolicy"),
			rp.Name,
			"Can not delete as deletionPolicy is set to \"DoNotTerminate\""))
		return nil, apierrors.NewInvalid(schema.GroupKind{Group: "kafka.kubedb.com", Kind: "RestProxy"}, rp.Name, allErr)
	}
	return nil, nil
}

func (k *RestProxyCustomWebhook) ValidateCreateOrUpdate(rp *kafkapi.RestProxy) field.ErrorList {
	var allErr field.ErrorList

	err := k.validateVersion(rp)
	if err != nil {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("version"),
			rp.Name,
			err.Error()))
		return allErr
	}

	if rp.Spec.SchemaRegistryRef != nil {
		if rp.Spec.SchemaRegistryRef.InternallyManaged && rp.Spec.SchemaRegistryRef.ObjectReference != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("schemaRegistryRef").Child("objectReference"),
				rp.Name,
				"ObjectReference should be nil when InternallyManaged is true"))
		}
		if !rp.Spec.SchemaRegistryRef.InternallyManaged && rp.Spec.SchemaRegistryRef.ObjectReference == nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("schemaRegistryRef").Child("objectReference"),
				rp.Name,
				"ObjectReference should not be nil when InternallyManaged is false"))
		}
	}

	if rp.Spec.DeletionPolicy == dbapi.DeletionPolicyHalt {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("deletionPolicy"),
			rp.Name,
			"DeletionPolicyHalt is not supported for RestProxy"))
	}

	// number of replicas can not be 0 or less
	if rp.Spec.Replicas != nil && *rp.Spec.Replicas <= 0 {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("replicas"),
			rp.Name,
			"number of replicas can not be 0 or less"))
	}

	err = k.validateVolumes(rp)
	if err != nil {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("podTemplate").Child("spec").Child("volumes"),
			rp.Name,
			err.Error()))
	}

	err = k.validateContainerVolumeMountPaths(rp)
	if err != nil {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("podTemplate").Child("spec").Child("containers").Child("volumeMounts"),
			rp.Name,
			err.Error()))
	}

	if len(allErr) == 0 {
		return nil
	}
	return allErr
}

func (k *RestProxyCustomWebhook) validateVersion(rp *kafkapi.RestProxy) error {
	ksrVersion := &catalog.SchemaRegistryVersion{}
	err := k.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: rp.Spec.Version}, ksrVersion)
	if err != nil {
		return errors.New("version not supported")
	}
	if ksrVersion.Spec.Distribution != catalog.SchemaRegistryDistroAiven {
		return errors.New(fmt.Sprintf("Distribution %s is not supported, only supported distribution is Aiven", ksrVersion.Spec.Distribution))
	}
	return nil
}

var restProxyReservedVolumes = []string{
	kafkapi.KafkaClientCertVolumeName,
	kafkapi.RestProxyOperatorVolumeConfig,
}

func (k *RestProxyCustomWebhook) validateVolumes(rp *kafkapi.RestProxy) error {
	if rp.Spec.PodTemplate.Spec.Volumes == nil {
		return nil
	}
	rsv := make([]string, len(restProxyReservedVolumes))
	copy(rsv, restProxyReservedVolumes)
	volumes := rp.Spec.PodTemplate.Spec.Volumes
	for _, rv := range rsv {
		for _, ugv := range volumes {
			if ugv.Name == rv {
				return errors.New("Cannot use a reserve volume name: " + rv)
			}
		}
	}
	return nil
}

var restProxyReservedVolumeMountPaths = []string{
	kafkapi.KafkaClientCertDir,
	kafkapi.RestProxyOperatorVolumeConfig,
}

func (k *RestProxyCustomWebhook) validateContainerVolumeMountPaths(rp *kafkapi.RestProxy) error {
	container := coreutil.GetContainerByName(rp.Spec.PodTemplate.Spec.Containers, kafkapi.RestProxyContainerName)
	if container == nil {
		return errors.New("container not found")
	}
	rPaths := restProxyReservedVolumeMountPaths
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
