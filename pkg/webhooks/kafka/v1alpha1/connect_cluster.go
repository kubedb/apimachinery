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
	"strings"

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	kafkapi "kubedb.dev/apimachinery/apis/kafka/v1alpha1"
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/klog/v2"
	coreutil "kmodules.xyz/client-go/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// SetupConnectClusterWebhookWithManager registers the webhook for Kafka ConnectCluster in the manager.
func SetupConnectClusterWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&kafkapi.ConnectCluster{}).
		WithValidator(&ConnectClusterCustomWebhook{mgr.GetClient()}).
		WithDefaulter(&ConnectClusterCustomWebhook{mgr.GetClient()}).
		Complete()
}

type ConnectClusterCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var connectClusterLog = logf.Log.WithName("connectCluster-resource")

var _ webhook.CustomDefaulter = &ConnectClusterCustomWebhook{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (k *ConnectClusterCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	c, ok := obj.(*kafkapi.ConnectCluster)
	if !ok {
		return fmt.Errorf("expected an connect-cluster object but got %T", obj)
	}

	connectClusterLog.Info("default", "name", c.Name)
	c.SetDefaults(k.DefaultClient)
	return nil
}

var _ webhook.CustomValidator = &ConnectorCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (k *ConnectClusterCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	c, ok := obj.(*kafkapi.ConnectCluster)
	if !ok {
		return nil, fmt.Errorf("expected an connect-cluster object but got %T", obj)
	}

	connectClusterLog.Info("validate create", "name", c.Name)
	allErr := k.ValidateCreateOrUpdate(c)
	if len(allErr) == 0 {
		return nil, nil
	}
	return nil, apierrors.NewInvalid(schema.GroupKind{Group: "kafka.kubedb.com", Kind: "ConnectCluster"}, c.Name, allErr)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (k *ConnectClusterCustomWebhook) ValidateUpdate(ctx context.Context, old, newObj runtime.Object) (admission.Warnings, error) {
	c, ok := newObj.(*kafkapi.ConnectCluster)
	if !ok {
		return nil, fmt.Errorf("expected an connect-cluster object but got %T", newObj)
	}

	connectClusterLog.Info("validate update", "name", c.Name)

	oldConnect := old.(*kafkapi.ConnectCluster)
	allErr := k.ValidateCreateOrUpdate(c)

	if *oldConnect.Spec.Replicas == 1 && *c.Spec.Replicas > 1 {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("replicas"),
			c.Name,
			"Cannot scale up from 1 to more than 1 in standalone mode"))
	}

	if len(allErr) == 0 {
		return nil, nil
	}
	return nil, apierrors.NewInvalid(schema.GroupKind{Group: "kafka.kubedb.com", Kind: "ConnectCluster"}, c.Name, allErr)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (k *ConnectClusterCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	c, ok := obj.(*kafkapi.ConnectCluster)
	if !ok {
		return nil, fmt.Errorf("expected an connect-cluster object but got %T", obj)
	}

	connectClusterLog.Info("validate delete", "name", c.Name)

	var allErr field.ErrorList
	if c.Spec.DeletionPolicy == dbapi.DeletionPolicyDoNotTerminate {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("deletionPolicy"),
			c.Name,
			"Can not delete as deletionPolicy is set to \"DoNotTerminate\""))
		return nil, apierrors.NewInvalid(schema.GroupKind{Group: "kafka.kubedb.com", Kind: "ConnectCluster"}, c.Name, allErr)
	}
	return nil, nil
}

func (k *ConnectClusterCustomWebhook) ValidateCreateOrUpdate(c *kafkapi.ConnectCluster) field.ErrorList {
	var allErr field.ErrorList
	if c.Spec.EnableSSL {
		if c.Spec.TLS == nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("enableSSL"),
				c.Name,
				".spec.tls can't be nil, if .spec.enableSSL is true"))
		}
	} else {
		if c.Spec.TLS != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("enableSSL"),
				c.Name,
				".spec.tls must be nil, if .spec.enableSSL is disabled"))
		}
	}

	if c.Spec.DeletionPolicy == dbapi.DeletionPolicyHalt {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("deletionPolicy"),
			c.Name,
			"DeletionPolicyHalt is not supported for ConnectCluster"))
	}

	// number of replicas can not be 0 or less
	if c.Spec.Replicas != nil && *c.Spec.Replicas <= 0 {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("replicas"),
			c.Name,
			"number of replicas can not be 0 or less"))
	}

	if c.Spec.Configuration != nil && c.Spec.Configuration.SecretName != "" && c.Spec.ConfigSecret != nil {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("configuration").Child("secretName"),
			c.Name,
			"cannot use both configuration.secretName and configSecret, use configuration.secretName"))
	}

	err := k.validateVersion(c)
	if err != nil {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("version"),
			c.Name,
			err.Error()))
	}

	err = k.validateVolumes(c)
	if err != nil {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("podTemplate").Child("spec").Child("volumes"),
			c.Name,
			err.Error()))
	}

	err = k.validateContainerVolumeMountPaths(c)
	if err != nil {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("podTemplate").Child("spec").Child("containers").Child("volumeMounts"),
			c.Name,
			err.Error()))
	}

	err = k.validateInitContainerVolumeMountPaths(c)
	if err != nil {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("podTemplate").Child("spec").Child("initContainers").Child("volumeMounts"),
			c.Name,
			err.Error()))
	}

	err = k.validateEnvVars(c)
	if err != nil {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("podTemplate").Child("spec").Child("envs"),
			c.Name,
			err.Error()))
	}

	if len(allErr) == 0 {
		return nil
	}
	return allErr
}

func (k *ConnectClusterCustomWebhook) validateEnvVars(connect *kafkapi.ConnectCluster) error {
	container := coreutil.GetContainerByName(connect.Spec.PodTemplate.Spec.Containers, kafkapi.ConnectClusterContainerName)
	if container == nil {
		return errors.New("container not found")
	}
	env := coreutil.GetEnvByName(container.Env, kafkapi.ConnectClusterModeEnv)
	if env != nil {
		if *connect.Spec.Replicas > 1 && env.Value == string(kafkapi.ConnectClusterNodeRoleStandalone) {
			return errors.New("can't use standalone mode as env, if replicas is more than 1")
		}
	}
	return nil
}

func (k *ConnectClusterCustomWebhook) validateVersion(connect *kafkapi.ConnectCluster) error {
	kccVersion := &catalog.KafkaVersion{}
	err := k.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: connect.Spec.Version}, kccVersion)
	if err != nil {
		return errors.New("version not supported")
	}
	return nil
}

var reservedVolumes = []string{
	kafkapi.ConnectClusterOperatorVolumeConfig,
	kafkapi.ConnectClusterCustomVolumeConfig,
	kafkapi.ConnectorPluginsVolumeName,
	kafkapi.ConnectClusterAuthSecretVolumeName,
	kafkapi.ConnectClusterOffsetFileDirName,
	kafkapi.KafkaClientCertVolumeName,
	kafkapi.ConnectClusterServerCertsVolumeName,
}

func (k *ConnectClusterCustomWebhook) validateVolumes(connect *kafkapi.ConnectCluster) error {
	if connect.Spec.PodTemplate.Spec.Volumes == nil {
		return nil
	}
	rsv := make([]string, len(reservedVolumes))
	copy(rsv, reservedVolumes)
	volumes := connect.Spec.PodTemplate.Spec.Volumes
	for _, rv := range rsv {
		for _, ugv := range volumes {
			if ugv.Name == rv {
				return errors.New("Cannot use a reserve volume name: " + rv)
			}
		}
	}
	return nil
}

var reservedVolumeMountPaths = []string{
	kafkapi.ConnectClusterOperatorConfigPath,
	kafkapi.ConnectorPluginsVolumeDir,
	kafkapi.ConnectClusterAuthSecretVolumePath,
	kafkapi.ConnectClusterOffsetFileDir,
	kafkapi.ConnectClusterCustomConfigPath,
	kafkapi.KafkaClientCertDir,
	kafkapi.ConnectClusterServerCertVolumeDir,
}

func (k *ConnectClusterCustomWebhook) validateContainerVolumeMountPaths(connect *kafkapi.ConnectCluster) error {
	container := coreutil.GetContainerByName(connect.Spec.PodTemplate.Spec.Containers, kafkapi.ConnectClusterContainerName)
	if container == nil {
		return errors.New("container not found")
	}
	rPaths := reservedVolumeMountPaths
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

func (k *ConnectClusterCustomWebhook) validateInitContainerVolumeMountPaths(connect *kafkapi.ConnectCluster) error {
	for _, name := range connect.Spec.ConnectorPlugins {
		connectorVersion := &catalog.KafkaConnectorVersion{}
		err := k.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: name}, connectorVersion)
		if err != nil {
			klog.Errorf("can't get the kafka connector version object %s for %s \n", err.Error(), name)
			return errors.New("no connector version found for " + name)
		}
		initContainer := coreutil.GetContainerByName(connect.Spec.PodTemplate.Spec.InitContainers, strings.ToLower(connectorVersion.Spec.Type))
		if initContainer == nil {
			return errors.New("init container not found for " + strings.ToLower(connectorVersion.Spec.Type))
		}
		volumeMount := coreutil.GetVolumeMountByName(initContainer.VolumeMounts, kafkapi.ConnectorPluginsVolumeName)
		if volumeMount != nil && volumeMount.MountPath == kafkapi.ConnectorPluginsVolumeDir {
			return errors.New("Cannot use a reserve volume mount path: " + kafkapi.ConnectorPluginsVolumeDir)
		}
	}
	return nil
}
