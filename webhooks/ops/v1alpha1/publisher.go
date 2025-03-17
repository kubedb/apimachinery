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

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
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

// SetupPublisherWebhookWithManager registers the webhook for Publisher in the manager.
func SetupPublisherWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&replapi.Publisher{}).
		WithValidator(&PublisherCustomWebhook{mgr.GetClient()}).
		Complete()
}

type PublisherCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var publisherLog = logf.Log.WithName("postgres-publisher")

var _ webhook.CustomValidator = &PublisherCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (in *PublisherCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	pub, ok := obj.(*replapi.Publisher)
	if !ok {
		return nil, fmt.Errorf("expected an Publisher object but got %T", pub)
	}
	publisherLog.Info("validate create", "name", pub.Name)
	return nil, in.validateCreateOrUpdate(pub)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *PublisherCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	newPub, ok := newObj.(*replapi.Publisher)
	if !ok {
		return nil, fmt.Errorf("expected an Publisher object but got %T", newObj)
	}
	publisherLog.Info("validate update", "name", newPub.Name)

	oldPub, ok := oldObj.(*replapi.Publisher)
	if !ok {
		return nil, fmt.Errorf("expected an Publisher object but got %T", oldObj)
	}

	if err := validatePubUpdate(newPub, oldPub); err != nil {
		return nil, err
	}
	return nil, in.validateCreateOrUpdate(newPub)
}

func (in *PublisherCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func validatePubUpdate(obj, oldObj *replapi.Publisher) error {
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

func (k *PublisherCustomWebhook) validateCreateOrUpdate(req *replapi.Publisher) error {
	pgVersion := &catalog.PostgresVersion{}
	pg := &dbapi.Postgres{}
	err := k.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      req.Spec.DatabaseRef.Name,
		Namespace: req.Namespace,
	}, pg)
	if err != nil {
		return err
	}
	err = k.DefaultClient.Get(context.TODO(), types.NamespacedName{
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
	if majorVersion <= 10 {
		err = k.validatePubForMajorVersion10(req, pgVersion)
		if err != nil {
			return err
		}
	}
	if majorVersion <= 12 {
		err = k.validatePubForMajorVersion12(req, pgVersion)
		if err != nil {
			return err
		}
	}

	checkParam := make(map[replapi.DMLOperation]bool)

	for _, value := range req.Spec.Parameters.Operations {
		found := checkParam[value]
		if found {
			return fmt.Errorf("redundant parameters to publish")
		}
		checkParam[value] = true
	}

	return nil
}

func (k *PublisherCustomWebhook) validatePubForMajorVersion10(pub *replapi.Publisher, dbVersion *catalog.PostgresVersion) error {
	if pub.Spec.Parameters != nil && pub.Spec.Parameters.Operations != nil {
		for _, value := range pub.Spec.Parameters.Operations {
			if value == replapi.DMLOpTruncate {
				msg := fmt.Sprintf("truncate in publish paramaters is not allowed in postgresVersion %s", dbVersion.Spec.Version)
				return errors.New(msg)
			}
		}
	}
	return nil
}

func (k *PublisherCustomWebhook) validatePubForMajorVersion12(pub *replapi.Publisher, dbVersion *catalog.PostgresVersion) error {
	if pub.Spec.Parameters != nil && pub.Spec.Parameters.PublishViaPartitionRoot != nil {
		msg := fmt.Sprintf("publish_via_partition_root is not allowed in postgresVersion %s", dbVersion.Spec.Version)
		return errors.New(msg)
	}
	return nil
}

func GetMajorPgVersion(postgresVersion *catalog.PostgresVersion) (uint64, error) {
	ver, err := semver.NewVersion(postgresVersion.Spec.Version)
	if err != nil {
		return 0, errors.Wrap(err, "Failed to get postgres major.")
	}
	return ver.Major(), nil
}
