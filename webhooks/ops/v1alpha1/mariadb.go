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
func (in *MariaDBOpsRequestCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	ops, ok := obj.(*opsapi.MariaDBOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an MariaDBOpsRequest object but got %T", obj)
	}
	mdLog.Info("validate create", "name", ops.Name)
	return nil, in.validateCreateOrUpdate(ops)
}

// ValidateUpdate implements webhooin.Validator so a webhook will be registered for the type
func (in *MariaDBOpsRequestCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
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
	return nil, in.validateCreateOrUpdate(ops)
}

func (in *MariaDBOpsRequestCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (in *MariaDBOpsRequestCustomWebhook) validateCreateOrUpdate(req *opsapi.MariaDBOpsRequest) error {
	var allErr field.ErrorList
	switch req.GetRequestType().(opsapi.MariaDBOpsRequestType) {
	case opsapi.MariaDBOpsRequestTypeRestart:
		if err := in.hasDatabaseRef(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("restart"),
				req.Name,
				err.Error()))
		}
	case opsapi.MariaDBOpsRequestTypeVerticalScaling:
		if err := in.validateMariaDBScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("verticalScaling"),
				req.Name,
				err.Error()))
		}
	case opsapi.MariaDBOpsRequestTypeHorizontalScaling:
		if err := in.validateMariaDBScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("horizontalScaling"),
				req.Name,
				err.Error()))
		}
	case opsapi.MariaDBOpsRequestTypeReconfigure:
		if err := in.validateMariaDBReconfigurationOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("configuration"),
				req.Name,
				err.Error()))
		}
	case opsapi.MariaDBOpsRequestTypeUpdateVersion:
		if err := in.validateMariaDBUpgradeOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("updateVersion"),
				req.Name,
				err.Error()))
		}
	case opsapi.MariaDBOpsRequestTypeReconfigureTLS:
		if err := in.validateMariaDBReconfigurationTLSOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("tls"),
				req.Name,
				err.Error()))
		}
	case opsapi.MariaDBOpsRequestTypeVolumeExpansion:
		if err := in.validateMariaDBVolumeExpansionOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("volumeExpansion"),
				req.Name,
				err.Error()))
		}
	default:
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("type"), req.Name,
			fmt.Sprintf("defined OpsRequestType %s is not supported, supported types for MariaDB are %s", req.Spec.Type, strings.Join(opsapi.MariaDBOpsRequestTypeNames(), ", "))))
	}
	if len(allErr) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "MariaDBopsrequests.kubedb.com", Kind: "MariaDBOpsRequest"}, req.Name, allErr)
}

func (in *MariaDBOpsRequestCustomWebhook) hasDatabaseRef(req *opsapi.MariaDBOpsRequest) error {
	md := dbapi.MariaDB{}
	if err := in.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      req.GetDBRefName(),
		Namespace: req.GetNamespace(),
	}, &md); err != nil {
		return errors.New(fmt.Sprintf("spec.databaseRef %s/%s, is invalid or not found", req.GetNamespace(), req.GetDBRefName()))
	}
	return nil
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

func (in *MariaDBOpsRequestCustomWebhook) validateMariaDBUpgradeOpsRequest(req *opsapi.MariaDBOpsRequest) error {
	// right now, kubeDB support the following mariadb version: 10.4.17 and 10.5.8
	if req.Spec.UpdateVersion == nil {
		return errors.New("spec.Upgrade is nil")
	}
	db := &dbapi.MariaDB{}
	err := in.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: req.GetDBRefName(), Namespace: req.GetNamespace()}, db)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to get mariadb: %s/%s", req.Namespace, req.Spec.DatabaseRef.Name))
	}
	mdNextVersion := &catalog.MariaDBVersion{}
	err = in.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: req.Spec.UpdateVersion.TargetVersion}, mdNextVersion)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to get mariadbVersion: %s", req.Spec.UpdateVersion.TargetVersion))
	}
	// check if mdNextVersion is deprecated.if deprecated, return error
	if mdNextVersion.Spec.Deprecated {
		return fmt.Errorf("mariadb target version %s/%s is deprecated. Skipped processing", db.Namespace, mdNextVersion.Name)
	}
	return nil
}

func (in *MariaDBOpsRequestCustomWebhook) validateMariaDBScalingOpsRequest(req *opsapi.MariaDBOpsRequest) error {
	if req.Spec.Type == opsapi.MariaDBOpsRequestTypeHorizontalScaling {
		if req.Spec.HorizontalScaling == nil {
			return errors.New("`spec.Scale.HorizontalScaling` field is nil")
		}

		if err := in.ensureMariaDBGroupReplication(req); err != nil {
			return err
		}

		if *req.Spec.HorizontalScaling.Member <= int32(2) || *req.Spec.HorizontalScaling.Member > int32(9) {
			return errors.New("Group size can not be less than 3 or greater than 9, range: [3,9]")
		}
		return nil
	}

	if req.Spec.VerticalScaling == nil {
		return errors.New("`spec.Scale.Vertical` field is empty")
	}

	return nil
}

func (in *MariaDBOpsRequestCustomWebhook) ensureMariaDBGroupReplication(req *opsapi.MariaDBOpsRequest) error {
	db := &dbapi.MariaDB{}
	err := in.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: req.GetDBRefName(), Namespace: req.GetNamespace()}, db)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to get mariadb: %s/%s", req.Namespace, req.Spec.DatabaseRef.Name))
	}

	if !db.IsCluster() {
		return errors.New("OpsRequest haven't pointed to a MariaDB Cluster, Horizontal scaling applicable only for MariaDB Cluster")
	}
	return nil
}

func (in *MariaDBOpsRequestCustomWebhook) validateMariaDBVolumeExpansionOpsRequest(req *opsapi.MariaDBOpsRequest) error {
	if req.Spec.VolumeExpansion == nil || req.Spec.VolumeExpansion.MariaDB == nil {
		return errors.New("`.Spec.VolumeExpansion` field is nil")
	}
	db := &dbapi.MariaDB{}
	err := in.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: req.GetDBRefName(), Namespace: req.GetNamespace()}, db)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to get mariadb: %s/%s", req.Namespace, req.Spec.DatabaseRef.Name))
	}

	cur, ok := db.Spec.Storage.Resources.Requests[core.ResourceStorage]
	if !ok {
		return errors.Wrap(err, "failed to parse current storage size")
	}

	if cur.Cmp(*req.Spec.VolumeExpansion.MariaDB) >= 0 {
		return fmt.Errorf("Desired storage size must be greater than current storage. Current storage: %v", cur.String())
	}

	return nil
}

func (in *MariaDBOpsRequestCustomWebhook) validateMariaDBReconfigurationOpsRequest(req *opsapi.MariaDBOpsRequest) error {
	if req.Spec.Configuration == nil || (!req.Spec.Configuration.RemoveCustomConfig && len(req.Spec.Configuration.ApplyConfig) == 0 && req.Spec.Configuration.ConfigSecret == nil) {
		return errors.New("`.Spec.Configuration` field is nil/not assigned properly")
	}

	assign := 0
	if req.Spec.Configuration.RemoveCustomConfig {
		assign++
	}
	if req.Spec.Configuration.ApplyConfig != nil {
		assign++
	}
	if req.Spec.Configuration.ConfigSecret != nil {
		assign++
	}
	if assign > 1 {
		return errors.New("more than 1 field have assigned to reconfigure your database but at a time you you are allowed to run one operation(`RemoveCustomConfig`, `ApplyConfig` or `ConfigSecret`) to reconfigure")
	}

	return nil
}

func (in *MariaDBOpsRequestCustomWebhook) validateMariaDBReconfigurationTLSOpsRequest(req *opsapi.MariaDBOpsRequest) error {
	db := &dbapi.MariaDB{}
	err := in.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: req.GetDBRefName(), Namespace: req.GetNamespace()}, db)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to get mariadb: %s/%s", req.Namespace, req.Spec.DatabaseRef.Name))
	}
	dbVersion := &catalog.MariaDBVersion{}
	err = in.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: db.Spec.Version}, dbVersion)
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
	if req.Spec.TLS.TLSConfig.IssuerRef != nil || req.Spec.TLS.TLSConfig.Certificates != nil || req.Spec.TLS.RequireSSL != nil {
		configCount++
	}

	if configCount == 0 {
		return errors.New("No of incomplete reconfiguration is provided in TLS Spec.")
	}

	if configCount > 1 {
		return errors.New("more than 1 field have assigned to reconfigureTLS to your database but at a time you you are allowed to run one operation")
	}
	return nil
}
