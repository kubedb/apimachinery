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
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// SetupConnectorWebhookWithManager registers the webhook for Kafka Connector in the manager.
func SetupConnectorWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&kafkapi.Connector{}).
		WithValidator(&ConnectorCustomWebhook{mgr.GetClient()}).
		WithDefaulter(&ConnectorCustomWebhook{mgr.GetClient()}).
		Complete()
}

type ConnectorCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var connectorlog = logf.Log.WithName("connector-resource")

var _ webhook.CustomDefaulter = &ConnectorCustomWebhook{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (k ConnectorCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	c, ok := obj.(*kafkapi.Connector)
	if !ok {
		return fmt.Errorf("expected an connector object but got %T", obj)
	}

	connectClusterLog.Info("default", "name", c.Name)
	c.Default()

	return nil
}

var _ webhook.CustomValidator = &ConnectorCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (k ConnectorCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	c, ok := obj.(*kafkapi.Connector)
	if !ok {
		return nil, fmt.Errorf("expected an connector object but got %T", obj)
	}
	connectClusterLog.Info("validate create", "name", c.Name)
	return nil, k.ValidateCreateOrUpdate(c)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (k ConnectorCustomWebhook) ValidateUpdate(ctx context.Context, old, newObj runtime.Object) (admission.Warnings, error) {
	c, ok := newObj.(*kafkapi.Connector)
	if !ok {
		return nil, fmt.Errorf("expected an connector object but got %T", newObj)
	}

	connectClusterLog.Info("validate update", "name", c.Name)
	return nil, k.ValidateCreateOrUpdate(c)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (k ConnectorCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	c, ok := obj.(*kafkapi.Connector)
	if !ok {
		return nil, fmt.Errorf("expected an connector object but got %T", obj)
	}

	connectorlog.Info("validate delete", "name", c.Name)

	var allErr field.ErrorList
	if c.Spec.DeletionPolicy == dbapi.DeletionPolicyDoNotTerminate {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("deletionPolicy"),
			c.Name,
			"Can not delete as deletionPolicy is set to \"DoNotTerminate\""))
		return nil, apierrors.NewInvalid(schema.GroupKind{Group: "kafka.kubedb.com", Kind: "Connector"}, c.Name, allErr)
	}
	return nil, nil
}

func (k ConnectorCustomWebhook) ValidateCreateOrUpdate(c *kafkapi.Connector) error {
	var allErr field.ErrorList
	if c.Spec.Configuration != nil && c.Spec.Configuration.SecretName != "" && c.Spec.ConfigSecret != nil {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("configuration").Child("secretName"),
			c.Name,
			"cannot use both configuration.secretName and configSecret, use configuration.secretName"))
	}

	if c.Spec.DeletionPolicy == dbapi.DeletionPolicyHalt {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("deletionPolicy"),
			c.Name,
			"DeletionPolicyHalt isn't supported for Connector"))
	}
	if len(allErr) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "kafka.kubedb.com", Kind: "ConnectCluster"}, c.Name, allErr)
}
