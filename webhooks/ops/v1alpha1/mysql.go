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

// SetupMySQLOpsRequestWebhookWithManager registers the webhook for MySQLOpsRequest in the manager.
func SetupMySQLOpsRequestWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&opsapi.MySQLOpsRequest{}).
		WithValidator(&MySQLOpsRequestCustomWebhook{mgr.GetClient()}).
		Complete()
}

type MySQLOpsRequestCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var myLog = logf.Log.WithName("mysql-opsrequest")

var _ webhook.CustomValidator = &MySQLOpsRequestCustomWebhook{}

// ValidateCreate implements webhooin.Validator so a webhook will be registered for the type
func (in *MySQLOpsRequestCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	ops, ok := obj.(*opsapi.MySQLOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an MySQLOpsRequest object but got %T", obj)
	}
	myLog.Info("validate create", "name", ops.Name)
	return nil, in.validateCreateOrUpdate(ops)
}

// ValidateUpdate implements webhooin.Validator so a webhook will be registered for the type
func (in *MySQLOpsRequestCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	ops, ok := newObj.(*opsapi.MySQLOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an MySQLOpsRequest object but got %T", newObj)
	}
	myLog.Info("validate update", "name", ops.Name)

	oldOps, ok := oldObj.(*opsapi.MySQLOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an MySQLOpsRequest object but got %T", oldObj)
	}

	if err := in.validateMySQLOpsRequest(ops, oldOps); err != nil {
		return nil, err
	}
	return nil, in.validateCreateOrUpdate(ops)
}

func (in *MySQLOpsRequestCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (in *MySQLOpsRequestCustomWebhook) validateCreateOrUpdate(req *opsapi.MySQLOpsRequest) error {
	var allErr field.ErrorList
	switch req.GetRequestType().(opsapi.MySQLOpsRequestType) {
	case opsapi.MySQLOpsRequestTypeRestart:
		if err := in.hasDatabaseRef(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("restart"),
				req.Name,
				err.Error()))
		}
	case opsapi.MySQLOpsRequestTypeVerticalScaling:
		if err := in.validateMySQLScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("verticalScaling"),
				req.Name,
				err.Error()))
		}
	case opsapi.MySQLOpsRequestTypeHorizontalScaling:
		if err := in.validateMySQLScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("horizontalScaling"),
				req.Name,
				err.Error()))
		}
	case opsapi.MySQLOpsRequestTypeReconfigure:
		if err := in.validateMySQLReconfigurationOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("configuration"),
				req.Name,
				err.Error()))
		}
	case opsapi.MySQLOpsRequestTypeUpdateVersion:
		if err := in.validateMySQLUpgradeOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("updateVersion"),
				req.Name,
				err.Error()))
		}
	case opsapi.MySQLOpsRequestTypeReconfigureTLS:
		if err := in.validateMySQLReconfigurationTLSOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("tls"),
				req.Name,
				err.Error()))
		}
	case opsapi.MySQLOpsRequestTypeVolumeExpansion:
		if err := in.validateMySQLVolumeExpansionOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("volumeExpansion"),
				req.Name,
				err.Error()))
		}
	case opsapi.MySQLOpsRequestTypeReplicationModeTransformation:
		if err := in.validateMySQLReplicationModeTransformation(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("replicationModeTransformation"),
				req.Name,
				err.Error()))
		}

	default:
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("type"), req.Name,
			fmt.Sprintf("defined OpsRequestType %s is not supported, supported types for MySQL are %s", req.Spec.Type, strings.Join(opsapi.MySQLOpsRequestTypeNames(), ", "))))
	}
	if len(allErr) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "MySQLopsrequests.kubedb.com", Kind: "MySQLOpsRequest"}, req.Name, allErr)
}

func (in *MySQLOpsRequestCustomWebhook) hasDatabaseRef(req *opsapi.MySQLOpsRequest) error {
	md := dbapi.MySQL{}
	if err := in.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      req.GetDBRefName(),
		Namespace: req.GetNamespace(),
	}, &md); err != nil {
		return errors.New(fmt.Sprintf("spec.databaseRef %s/%s, is invalid or not found", req.GetNamespace(), req.GetDBRefName()))
	}
	return nil
}

func (in *MySQLOpsRequestCustomWebhook) validateMySQLUpgradeOpsRequest(req *opsapi.MySQLOpsRequest) error {
	// right now, kubeDB support the following mysql version: 5.7.25, 5.7.29, 5.7.31, 8.0.3, 8.0.14, 8.0.18, 8.0.20 and 8.0.21
	updateVersionSpec := req.Spec.UpdateVersion
	if updateVersionSpec == nil {
		return errors.New("spec.Upgrade & spec.UpdateVersion both nil not supported")
	}
	db := &dbapi.MySQL{}
	err := in.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: req.Spec.DatabaseRef.Name}, db)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to get mysql: %s/%s", req.Namespace, req.Spec.DatabaseRef.Name))
	}
	myCurVersion := &catalog.MySQLVersion{}
	err = in.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: db.Spec.Version}, myCurVersion)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to get mysqlVersion: %s", updateVersionSpec.TargetVersion))
	}
	myNextVersion := &catalog.MySQLVersion{}
	err = in.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: updateVersionSpec.TargetVersion}, myNextVersion)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to get mysqlVersion: %s", updateVersionSpec.TargetVersion))
	}
	// check if myNextVersion is deprecated.if deprecated, return error
	if myNextVersion.Spec.Deprecated {
		return fmt.Errorf("mysql target version %s/%s is deprecated. Skipped processing", db.Namespace, myNextVersion.Name)
	}

	// check whitelist to determine which version will be supported for upgrading modification
	if len(myCurVersion.Spec.UpdateConstraints.Allowlist.Standalone) > 0 || len(myCurVersion.Spec.UpdateConstraints.Allowlist.GroupReplication) > 0 {
		v, err := semver.NewVersion(myNextVersion.Spec.Version)
		if err != nil {
			return errors.Wrap(err, "unbale to parse the given myNext version")
		}
		// check group replication/standalone constraint
		if db.Spec.Topology != nil && db.Spec.Topology.Mode != nil {
			for _, constraints := range myCurVersion.Spec.UpdateConstraints.Allowlist.GroupReplication {
				con, err := semver.NewConstraint(constraints)
				if err != nil {
					return errors.Wrap(err, "unable to parse constraints from the given constraint string")
				}
				if !con.Check(v) {
					return fmt.Errorf("update version from %v to %v is not supported", myCurVersion, myNextVersion)
				}
			}
		} else {
			for _, constraints := range myCurVersion.Spec.UpdateConstraints.Allowlist.Standalone {
				con, err := semver.NewConstraint(constraints)
				if err != nil {
					return errors.Wrap(err, "unable to parse constraints from the given constraint string")
				}
				if !con.Check(v) {
					return fmt.Errorf("update version from %v to %v is not supported", myCurVersion, myNextVersion)
				}
			}
		}
	}
	// check blacklist to determine which version will be rejected for upgrading modification
	if len(myCurVersion.Spec.UpdateConstraints.Denylist.Standalone) > 0 || len(myCurVersion.Spec.UpdateConstraints.Denylist.GroupReplication) > 0 {
		v, err := semver.NewVersion(myNextVersion.Spec.Version)
		if err != nil {
			return errors.Wrap(err, "unable to parse the given version")
		}
		// check group replication/standalone constraint
		if db.Spec.Topology != nil && db.Spec.Topology.Mode != nil {
			for _, constraints := range myCurVersion.Spec.UpdateConstraints.Denylist.GroupReplication {
				con, err := semver.NewConstraint(constraints)
				if err != nil {
					return errors.Wrap(err, "unable to parse constraints from the given constraint string")
				}
				if con.Check(v) {
					return fmt.Errorf("update version from %v to %v is not supported", myCurVersion, myNextVersion)
				}
			}
		} else {
			for _, constraints := range myCurVersion.Spec.UpdateConstraints.Denylist.Standalone {
				con, err := semver.NewConstraint(constraints)
				if err != nil {
					return errors.Wrap(err, "unable to parse constraints from the given constraint string")
				}
				if con.Check(v) {
					return fmt.Errorf("update version from %v to %v is not supported", myCurVersion, myNextVersion)
				}
			}
		}
	}
	return nil
}

func (in *MySQLOpsRequestCustomWebhook) validateMySQLScalingOpsRequest(req *opsapi.MySQLOpsRequest) error {
	if req.Spec.Type == opsapi.MySQLOpsRequestTypeHorizontalScaling {
		if req.Spec.HorizontalScaling == nil {
			return errors.New("`spec.Scale.HorizontalScaling` field is nil")
		}

		if err := in.ensureMySQLGroupReplication(req); err != nil {
			return err
		}

		if !(int32(2) < *req.Spec.HorizontalScaling.Member && int32(9) > *req.Spec.HorizontalScaling.Member) {
			return errors.New("Group size can not be less than 3 or greater than 9, range: [3,9]")
		}
		return nil
	}

	if req.Spec.VerticalScaling == nil {
		return errors.New("`spec.Scale.Vertical` field is empty")
	}

	return nil
}

func (in *MySQLOpsRequestCustomWebhook) validateMySQLVolumeExpansionOpsRequest(req *opsapi.MySQLOpsRequest) error {
	if req.Spec.VolumeExpansion == nil || req.Spec.VolumeExpansion.MySQL == nil {
		return errors.New("`.Spec.VolumeExpansion` field is nil")
	}
	db := &dbapi.MySQL{}
	err := in.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: req.Spec.DatabaseRef.Name}, db)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to get mysql: %s/%s", req.Namespace, req.Spec.DatabaseRef.Name))
	}

	cur, ok := db.Spec.Storage.Resources.Requests[core.ResourceStorage]
	if !ok {
		return errors.Wrap(err, "failed to parse current storage size")
	}

	if cur.Cmp(*req.Spec.VolumeExpansion.MySQL) >= 0 {
		return fmt.Errorf("Desired storage size must be greater than current storage. Current storage: %v", cur.String())
	}

	return nil
}

func (in *MySQLOpsRequestCustomWebhook) validateMySQLReconfigurationOpsRequest(req *opsapi.MySQLOpsRequest) error {
	db := &dbapi.MySQL{}
	err := in.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: req.Spec.DatabaseRef.Name}, db)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to get mysql: %s/%s", req.Namespace, req.Spec.DatabaseRef.Name))
	}

	if req.Spec.Configuration == nil || (!req.Spec.Configuration.RemoveCustomConfig && !applyConfigExists(req.Spec.Configuration.ApplyConfig) && req.Spec.Configuration.ConfigSecret == nil) {
		return errors.New("`.Spec.Configuration` field is nil/not assigned properly")
	}

	assign := 0
	if req.Spec.Configuration.RemoveCustomConfig {
		assign++
	}
	if applyConfigExists(req.Spec.Configuration.ApplyConfig) {
		assign++
	}
	if req.Spec.Configuration.ConfigSecret != nil {
		assign++
	}
	if assign > 1 {
		return errors.New("more than 1 field have assigned to reconfigure your database but at a time you you are allowed to run one operation(`RemoveCustomConfig`, `ApplyConfig` or `ConfigSecret`) to reconfigure")
	}
	if db.Spec.ConfigSecret == nil && req.Spec.Configuration.RemoveCustomConfig {
		return errors.New("database is not custom configured. so no need to run `RemoveCustomConfig` operation.")
	}

	return nil
}

func (in *MySQLOpsRequestCustomWebhook) validateMySQLReconfigurationTLSOpsRequest(req *opsapi.MySQLOpsRequest) error {
	db := &dbapi.MySQL{}
	err := in.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: req.Spec.DatabaseRef.Name}, db)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to get mysql: %s/%s", req.Namespace, req.Spec.DatabaseRef.Name))
	}

	if req.Spec.TLS == nil || (req.Spec.TLS.Remove && req.Spec.TLS.RotateCertificates) {
		return errors.New("more than 1 field have assigned to reconfigureTLS to your database but at a time you you are allowed to run one operation(`Remove` or `RotateCertificates`)")
	}

	return nil
}

func (in *MySQLOpsRequestCustomWebhook) validateMySQLReplicationModeTransformation(req *opsapi.MySQLOpsRequest) error {
	db := &dbapi.MySQL{}
	err := in.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: req.Spec.DatabaseRef.Name}, db)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to get mysql: %s/%s", req.Namespace, req.Spec.DatabaseRef.Name))
	}

	curVersion := semver.MustParse(db.Spec.Version)
	refVersion := semver.MustParse("8.4.2")

	if curVersion.LessThan(refVersion) {
		return errors.Wrap(err, fmt.Sprintf("MySQL Replication Mode Transformation support only support for %s or upper.", refVersion))
	}

	if req.Spec.ReplicationModeTransformation != nil {
		if req.Spec.ReplicationModeTransformation.RequireSSL != nil && (req.Spec.ReplicationModeTransformation.TLSConfig.IssuerRef == nil &&
			req.Spec.ReplicationModeTransformation.TLSConfig.Certificates == nil) {
			return errors.Wrap(err, "MySQL Replication Mode Transformation requires TLS configuration to be enabled.")
		}
	}

	return nil
}

func (in *MySQLOpsRequestCustomWebhook) ensureMySQLGroupReplication(req *opsapi.MySQLOpsRequest) error {
	db := &dbapi.MySQL{}
	err := in.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: req.Spec.DatabaseRef.Name}, db)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to get mysql: %s/%s", req.Namespace, req.Spec.DatabaseRef.Name))
	}

	if db == nil {
		return errors.New("MySQL object is empty")
	}

	if db.Spec.Topology == nil || db.Spec.Topology.Mode == nil {
		return errors.New("OpsRequest haven't pointed to a Group Replication, Horizontal scaling applicable only for group Replication")
	}
	return nil
}

func (in *MySQLOpsRequestCustomWebhook) validateMySQLOpsRequest(obj, oldObj runtime.Object) error {
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
