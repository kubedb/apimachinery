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
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"
	replapi "kubedb.dev/apimachinery/apis/postgres/v1alpha1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/mergepatch"
	"k8s.io/apimachinery/pkg/util/sets"
	meta_util "kmodules.xyz/client-go/meta"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// SetupSubscriberWebhookWithManager registers the webhook for Subscriber in the manager.
func SetupSubscriberWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&replapi.Subscriber{}).
		WithValidator(&SubscriberCustomWebhook{mgr.GetClient()}).
		Complete()
}

type SubscriberCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var subscriberLog = logf.Log.WithName("postgres-subscriber")

var _ webhook.CustomValidator = &SubscriberCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *SubscriberCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	sub, ok := obj.(*replapi.Subscriber)
	if !ok {
		return nil, fmt.Errorf("expected an Subscriber object but got %T", sub)
	}
	subscriberLog.Info("validate create", "name", sub.Name)
	return nil, w.validateCreateOrUpdate(sub)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *SubscriberCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	newSub, ok := newObj.(*replapi.Subscriber)
	if !ok {
		return nil, fmt.Errorf("expected an Subscriber object but got %T", newObj)
	}
	subscriberLog.Info("validate update", "name", newSub.Name)

	oldSub, ok := oldObj.(*replapi.Subscriber)
	if !ok {
		return nil, fmt.Errorf("expected an Subscriber object but got %T", oldObj)
	}

	if err := validateSubUpdate(newSub, oldSub); err != nil {
		return nil, err
	}
	if err := w.validateAlterParams(newSub, oldSub); err != nil {
		return nil, err
	}
	return nil, w.validateCreateOrUpdate(newSub)
}

func (w *SubscriberCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func validateSubUpdate(obj, oldObj *replapi.Subscriber) error {
	preconditions := meta_util.PreConditionSet{Set: sets.New[string]("spec.name", "Spec.DatabaseRef", "spec.databaseName")}
	_, err := meta_util.CreateStrategicPatch(oldObj, obj, preconditions.PreconditionFunc()...)
	if err != nil {
		if mergepatch.IsPreconditionFailed(err) {
			return fmt.Errorf("%v.%v", err, preconditions.Error())
		}
		return err
	}
	return nil
}

func (w *SubscriberCustomWebhook) validateCreateOrUpdate(req *replapi.Subscriber) error {
	pg := &dbapi.Postgres{}
	pgVersion := &catalog.PostgresVersion{}
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      req.Spec.DatabaseRef.Name,
		Namespace: req.Namespace,
	}, pg)
	if err != nil {
		return err
	}
	err = w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name: pg.Spec.Version,
	}, pgVersion)
	if err != nil {
		return err
	}
	majorVersion, err := GetMajorPgVersion(pgVersion)
	if err != nil {
		return err
	}
	if majorVersion <= 9 {
		return fmt.Errorf("logical replication is not allowed in postgresVersion %s", pgVersion.Spec.Version)
	}

	if majorVersion <= 13 {
		err = w.validateSubForMajorVersion13(req, pgVersion)
		if err != nil {
			return err
		}
	}

	// nolint:staticcheck
	if (req.Spec.Parameters.Connect != nil) && !(*req.Spec.Parameters.Connect &&
		(req.Spec.Parameters.CreateSlot == nil || *req.Spec.Parameters.CreateSlot) &&
		(req.Spec.Parameters.Enabled == nil || *req.Spec.Parameters.Enabled) &&
		(req.Spec.Parameters.CopyData == nil || *req.Spec.Parameters.CopyData)) {
		return fmt.Errorf("%s %s %s should be true while connect is true", "copyData", "createSlot", "enabled")
	}

	if req.Spec.Publisher.Managed == nil && req.Spec.Publisher.External == nil {
		return fmt.Errorf("need to specified the publisher")
	}

	return nil
}

func (w *SubscriberCustomWebhook) validateSubForMajorVersion13(sub *replapi.Subscriber, dbVersion *catalog.PostgresVersion) error {
	if sub.Spec.Parameters != nil {
		if sub.Spec.Parameters.Streaming != nil {
			return fmt.Errorf("streaming in parameters is not allowed in postgresVersion %s", dbVersion.Spec.Version)
		}
		if sub.Spec.Parameters.Binary != nil {
			return fmt.Errorf("binary in parameters is not allowed in postgresVersion %s", dbVersion.Spec.Version)
		}
	}
	return nil
}

func (w *SubscriberCustomWebhook) validateAlterParams(sub *replapi.Subscriber, oldSub *replapi.Subscriber) error {
	if (oldSub.Spec.Parameters.CopyData != nil) &&
		(sub.Spec.Parameters.CopyData != nil) &&
		(*oldSub.Spec.Parameters.CopyData != *sub.Spec.Parameters.CopyData) {
		return fmt.Errorf("can't change  spec.parameters.%s in %s/%s", "copyData", sub.Name, sub.Namespace)
	}

	if (oldSub.Spec.Parameters.CreateSlot != nil) &&
		(sub.Spec.Parameters.CreateSlot != nil) &&
		(*oldSub.Spec.Parameters.CreateSlot != *sub.Spec.Parameters.CreateSlot) {
		return fmt.Errorf("can't change  spec.parameters.%s in %s/%s", "createSlot", sub.Name, sub.Namespace)
	}

	if (oldSub.Spec.Parameters.Enabled != nil) &&
		(sub.Spec.Parameters.Enabled != nil) &&
		(*oldSub.Spec.Parameters.Enabled != *sub.Spec.Parameters.Enabled) {
		return fmt.Errorf("can't change  spec.parameters.%s in %s/%s", "enabled", sub.Name, sub.Namespace)
	}

	if (oldSub.Spec.Parameters.Connect != nil) &&
		(sub.Spec.Parameters.Connect != nil) &&
		(*oldSub.Spec.Parameters.Connect != *sub.Spec.Parameters.Connect) {
		return fmt.Errorf("can't change  spec.parameters.%s in %s/%s", "connect", sub.Name, sub.Namespace)
	}
	return nil
}
