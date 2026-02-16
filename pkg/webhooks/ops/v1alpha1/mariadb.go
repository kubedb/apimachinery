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
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
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

// SetupMariaDBOpsRequestWebhookWithManager registers the webhook for MariaDBOpsRequest in the manager.
func SetupMariaDBOpsRequestWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&opsapi.MariaDBOpsRequest{}).
		WithValidator(&MariaDBOpsRequestCustomWebhook{mgr.GetClient()}).
		Complete()
}

type MariaDBOpsRequestCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var mdLog = logf.Log.WithName("mariadb-opsrequest")

var _ webhook.CustomValidator = &MariaDBOpsRequestCustomWebhook{}

// ValidateCreate implements webhooin.Validator so a webhook will be registered for the type
func (w *MariaDBOpsRequestCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	ops, ok := obj.(*opsapi.MariaDBOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an MariaDBOpsRequest object but got %T", obj)
	}
	mdLog.Info("validate create", "name", ops.Name)
	return nil, w.validateCreateOrUpdate(ops)
}

// ValidateUpdate implements webhooin.Validator so a webhook will be registered for the type
func (w *MariaDBOpsRequestCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	ops, ok := newObj.(*opsapi.MariaDBOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an MariaDBOpsRequest object but got %T", newObj)
	}
	mdLog.Info("validate update", "name", ops.Name)

	oldOps, ok := oldObj.(*opsapi.MariaDBOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an MariaDBOpsRequest object but got %T", oldObj)
	}

	if err := validateMariaDBOpsRequest(ops, oldOps); err != nil {
		return nil, err
	}

	if err := w.validateCreateOrUpdate(ops); err != nil {
		return nil, err
	}

	if isOpsReqCompleted(ops.Status.Phase) && !isOpsReqCompleted(oldOps.Status.Phase) { // just completed
		var db dbapi.MariaDB
		err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: ops.Spec.DatabaseRef.Name, Namespace: ops.Namespace}, &db)
		if err != nil {
			return nil, err
		}
		return nil, resumeDatabase(w.DefaultClient, &db)
	}
	return nil, nil
}

func (w *MariaDBOpsRequestCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (w *MariaDBOpsRequestCustomWebhook) validateCreateOrUpdate(req *opsapi.MariaDBOpsRequest) error {
	if validType, _ := arrays.Contains(opsapi.MariaDBOpsRequestTypeNames(), string(req.Spec.Type)); !validType {
		return field.Invalid(field.NewPath("spec").Child("type"), req.Name,
			fmt.Sprintf("defined OpsRequestType %s is not supported, supported types for MariaDB are %s", req.Spec.Type, strings.Join(opsapi.MariaDBOpsRequestTypeNames(), ", ")))
	}

	db, err := w.hasDatabaseRef(req)
	if err != nil {
		return field.Invalid(field.NewPath("spec").Child("databaseRef"), req.Name, err.Error())
	}

	var allErr field.ErrorList
	switch opsapi.MariaDBOpsRequestType(req.GetRequestType()) {
	case opsapi.MariaDBOpsRequestTypeRestart:
	case opsapi.MariaDBOpsRequestTypeVerticalScaling:
		if err := w.validateMariaDBScalingOpsRequest(db, req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("verticalScaling"),
				req.Name,
				err.Error()))
		}
	case opsapi.MariaDBOpsRequestTypeHorizontalScaling:
		if err := w.validateMariaDBScalingOpsRequest(db, req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("horizontalScaling"),
				req.Name,
				err.Error()))
		}
	case opsapi.MariaDBOpsRequestTypeReconfigure:
		if err := w.validateMariaDBReconfigurationOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("configuration"),
				req.Name,
				err.Error()))
		}
	case opsapi.MariaDBOpsRequestTypeUpdateVersion:
		if err := w.validateMariaDBUpdateVersionOpsRequest(db, req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("updateVersion"),
				req.Name,
				err.Error()))
		}
	case opsapi.MariaDBOpsRequestTypeReconfigureTLS:
		if err := w.validateMariaDBReconfigurationTLSOpsRequest(db, req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("tls"),
				req.Name,
				err.Error()))
		}
	case opsapi.MariaDBOpsRequestTypeVolumeExpansion:
		if err := w.validateMariaDBVolumeExpansionOpsRequest(db, req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("volumeExpansion"),
				req.Name,
				err.Error()))
		}
	case opsapi.MariaDBOpsRequestTypeRotateAuth:
		if err := w.validateMariaDBRotateAuthenticationOpsRequest(db, req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("authentication"),
				req.Name,
				err.Error()))
		}

	}

	if len(allErr) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "MariaDBopsrequests.kubedb.com", Kind: "MariaDBOpsRequest"}, req.Name, allErr)
}

func (w *MariaDBOpsRequestCustomWebhook) hasDatabaseRef(req *opsapi.MariaDBOpsRequest) (*dbapi.MariaDB, error) {
	md := &dbapi.MariaDB{}
	if err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      req.GetDBRefName(),
		Namespace: req.GetNamespace(),
	}, md); err != nil {
		return nil, errors.New(fmt.Sprintf("spec.databaseRef %s/%s, is invalid or not found", req.GetNamespace(), req.GetDBRefName()))
	}
	return md, nil
}

func validateMariaDBOpsRequest(obj, oldObj runtime.Object) error {
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

func (w *MariaDBOpsRequestCustomWebhook) validateMariaDBUpdateVersionOpsRequest(db *dbapi.MariaDB, req *opsapi.MariaDBOpsRequest) error {
	updateVersionSpec := req.Spec.UpdateVersion
	if updateVersionSpec == nil {
		return errors.New("spec.updateVersion nil not supported in UpdateVersion type")
	}

	yes, err := IsUpgradable(w.DefaultClient, catalog.ResourceKindMariaDBVersion, db.Spec.Version, updateVersionSpec.TargetVersion)
	if err != nil {
		return err
	}
	if !yes {
		return fmt.Errorf("upgrade from version %v to %v is not supported", db.Spec.Version, req.Spec.UpdateVersion.TargetVersion)
	}

	return nil
}

func (w *MariaDBOpsRequestCustomWebhook) validateMariaDBScalingOpsRequest(db *dbapi.MariaDB, req *opsapi.MariaDBOpsRequest) error {
	if req.Spec.Type == opsapi.MariaDBOpsRequestTypeHorizontalScaling {
		if req.Spec.HorizontalScaling == nil {
			return errors.New("`spec.Scale.HorizontalScaling` field is nil")
		}

		if err := w.ensureMariaDBCluster(db); err != nil {
			return err
		}

		if *req.Spec.HorizontalScaling.Member <= int32(2) || *req.Spec.HorizontalScaling.Member > int32(9) {
			return errors.New("Group size can not be less than 3 or greater than 9, range: [3,9]")
		}
		return nil
	}
	if req.Spec.Type == opsapi.MariaDBOpsRequestTypeVerticalScaling {
		if req.Spec.VerticalScaling == nil {
			return errors.New("`spec.Scale.Vertical` field is empty")
		}
	}

	return nil
}

func (w *MariaDBOpsRequestCustomWebhook) ensureMariaDBCluster(db *dbapi.MariaDB) error {
	if !db.IsCluster() {
		return errors.New("OpsRequest haven't pointed to a MariaDB Cluster, Horizontal scaling applicable only for MariaDB Cluster")
	}

	return nil
}

func (w *MariaDBOpsRequestCustomWebhook) validateMariaDBVolumeExpansionOpsRequest(db *dbapi.MariaDB, req *opsapi.MariaDBOpsRequest) error {
	if req.Spec.VolumeExpansion == nil || (req.Spec.VolumeExpansion.MariaDB == nil && req.Spec.VolumeExpansion.MaxScale == nil) {
		return errors.New("`.Spec.VolumeExpansion` field is nil")
	}

	if req.Spec.VolumeExpansion.MariaDB != nil && req.Spec.VolumeExpansion.MaxScale != nil {
		return errors.New("Volume expansion for both server at a time is not allowed")
	}

	if req.Spec.VolumeExpansion.MariaDB != nil {
		cur, ok := db.Spec.Storage.Resources.Requests[core.ResourceStorage]
		if !ok {
			return errors.New("failed to parse mariadb storage size")
		}

		if cur.Cmp(*req.Spec.VolumeExpansion.MariaDB) >= 0 && (req.Status.Phase == opsapi.OpsRequestPhasePending || req.Status.Phase == "") {
			return fmt.Errorf("desired storage size must be greater than current storage. Current storage: %v", cur.String())
		}
	}

	if req.Spec.VolumeExpansion.MaxScale != nil {
		if !db.IsMariaDBReplication() {
			return errors.New("Topology is not mariadb replication")
		}
		cur, ok := db.Spec.Topology.MaxScale.Storage.Resources.Requests[core.ResourceStorage]
		if !ok {
			return errors.New("failed to parse maxscale storage size")
		}
		if (req.Status.Phase == opsapi.OpsRequestPhasePending || req.Status.Phase == "") && cur.Cmp(*req.Spec.VolumeExpansion.MaxScale) >= 0 {
			return fmt.Errorf("desired storage size must be greater than current storage. Current storage: %v", cur.String())
		}
	}

	return nil
}

func (w *MariaDBOpsRequestCustomWebhook) validateMariaDBReconfigurationOpsRequest(req *opsapi.MariaDBOpsRequest) error {
	if req.Spec.Configuration == nil {
		return errors.New("`.Spec.Configuration` field is nil")
	}

	if !req.Spec.Configuration.RemoveCustomConfig && len(req.Spec.Configuration.ApplyConfig) == 0 && req.Spec.Configuration.ConfigSecret == nil {
		return errors.New("at least one of `RemoveCustomConfig`, `ConfigSecret`, or `ApplyConfig` must be specified")
	}

	return nil
}

func (w *MariaDBOpsRequestCustomWebhook) validateMariaDBReconfigurationTLSOpsRequest(db *dbapi.MariaDB, req *opsapi.MariaDBOpsRequest) error {
	dbVersion := &catalog.MariaDBVersion{}
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: db.Spec.Version}, dbVersion)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to get mariadbversion: %s/%s", req.Namespace, req.Spec.DatabaseRef.Name))
	}

	curVersion, err := semver.NewVersion(dbVersion.Spec.Version)
	if err != nil {
		return fmt.Errorf(`unable to parse spec.version`)
	}
	supportedVersion, err := semver.NewVersion("10.5.2")
	if err != nil {
		return fmt.Errorf(`unable to parse spec.version`)
	}

	if req.Spec.TLS.RequireSSL != nil && *req.Spec.TLS.RequireSSL && curVersion.LessThan(supportedVersion) {
		return fmt.Errorf(`requireSSL is not supported for the MariaDDB Versions lower than 10.5.2`)
	}

	if req.Spec.TLS == nil {
		return errors.New("TLS Spec is empty")
	}
	configCount := 0
	if req.Spec.TLS.Remove {
		configCount++
	}
	if req.Spec.TLS.RotateCertificates {
		configCount++
	}
	if req.Spec.TLS.IssuerRef != nil || req.Spec.TLS.Certificates != nil || req.Spec.TLS.RequireSSL != nil {
		configCount++
	}

	if configCount == 0 {
		return errors.New("no of incomplete reconfiguration is provided in TLS spec")
	}

	if configCount > 1 {
		return errors.New("more than 1 field have assigned to reconfigureTLS to your database but at a time you you are allowed to run one operation")
	}
	return nil
}

func (w *MariaDBOpsRequestCustomWebhook) validateMariaDBRotateAuthenticationOpsRequest(db *dbapi.MariaDB, req *opsapi.MariaDBOpsRequest) error {
	authSpec := req.Spec.Authentication
	if authSpec != nil && authSpec.SecretRef != nil {
		if authSpec.SecretRef.Name == "" {
			return errors.New("spec.authentication.secretRef.name can not be empty")
		}
		var newAuthSecret, oldAuthSecret core.Secret
		err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
			Name:      authSpec.SecretRef.Name,
			Namespace: req.Namespace,
		}, &newAuthSecret)
		if err != nil {
			if apierrors.IsNotFound(err) {
				return errors.Wrap(err, fmt.Sprintf("referenced secret %s/%s not found", req.Namespace, authSpec.SecretRef.Name))
			}
			return err
		}

		err = w.DefaultClient.Get(context.TODO(), types.NamespacedName{
			Name:      db.GetAuthSecretName(),
			Namespace: db.GetNamespace(),
		}, &oldAuthSecret)
		if err != nil {
			return err
		}

		if string(oldAuthSecret.Data[core.BasicAuthUsernameKey]) != string(newAuthSecret.Data[core.BasicAuthUsernameKey]) {
			return errors.New("database username cannot be changed")
		}
	}
	return nil
}
