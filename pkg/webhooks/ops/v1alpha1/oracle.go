package v1alpha1

import (
	"context"
	"fmt"
	"strings"

	"errors"
	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	"kubedb.dev/apimachinery/apis/kubedb"
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	"gomodules.xyz/x/arrays"
	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/mergepatch"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
	metautil "kmodules.xyz/client-go/meta"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func SetupOracleOpsRequestWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&opsapi.OracleOpsRequest{}).
		WithValidator(&OracleOpsRequestCustomWebhook{mgr.GetClient()}).
		Complete()
}

type OracleOpsRequestCustomWebhook struct {
	DefaultClient client.Client
}

var oracleOpsReqLog = logf.Log.WithName("oracle-opsrequest")

var _ webhook.CustomValidator = &OracleOpsRequestCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *OracleOpsRequestCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	ops, ok := obj.(*opsapi.OracleOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an OracleOpsRequest object but got %T", obj)
	}
	oracleOpsReqLog.Info("validate create", "name", ops.Name)
	return nil, w.validateCreateOrUpdate(ops)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *OracleOpsRequestCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	ops, ok := newObj.(*opsapi.OracleOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an OracleOpsRequest object but got %T", newObj)
	}
	oracleOpsReqLog.Info("validate update", "name", ops.Name)

	oldOps, ok := oldObj.(*opsapi.OracleOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an OracleOpsRequest object but got %T", oldObj)
	}

	if err := validateOracleOpsRequest(ops, oldOps); err != nil {
		return nil, err
	}

	if err := w.validateCreateOrUpdate(ops); err != nil {
		return nil, err
	}

	if isOpsReqCompleted(ops.Status.Phase) && !isOpsReqCompleted(oldOps.Status.Phase) { // just completed
		var db dbapi.Oracle
		err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: ops.Spec.DatabaseRef.Name, Namespace: ops.Namespace}, &db)
		if err != nil {
			return nil, err
		}
		return nil, resumeDatabase(w.DefaultClient, &db)
	}
	return nil, nil
}

func (w *OracleOpsRequestCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func validateOracleOpsRequest(req *opsapi.OracleOpsRequest, oldReq *opsapi.OracleOpsRequest) error {
	preconditions := metautil.PreConditionSet{Set: sets.New[string]("spec")}
	_, err := metautil.CreateStrategicPatch(oldReq, req, preconditions.PreconditionFunc()...)
	if err != nil {
		if mergepatch.IsPreconditionFailed(err) {
			return fmt.Errorf("%v.%v", err, preconditions.Error())
		}
		return err
	}
	return nil
}

func (w *OracleOpsRequestCustomWebhook) validateCreateOrUpdate(req *opsapi.OracleOpsRequest) error {
	if validType, _ := arrays.Contains(opsapi.OracleOpsRequestTypeNames(), string(req.Spec.Type)); !validType {
		return field.Invalid(field.NewPath("spec").Child("type"), req.Name,
			fmt.Sprintf("defined OpsRequestType %s is not supported, supported types for Oracle are %s", req.Spec.Type, strings.Join(opsapi.OracleOpsRequestTypeNames(), ", ")))
	}

	db, err := w.hasDatabaseRef(req)
	if err != nil {
		return field.Invalid(field.NewPath("spec").Child("databaseRef"), req.Name, err.Error())
	}

	var allErr field.ErrorList
	switch opsapi.OracleOpsRequestType(req.GetRequestType()) {
	case opsapi.OracleOpsRequestTypeHorizontalScaling:
		if err := w.validateOracleHorizontalScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("horizontalScaling"),
				req.Name,
				err.Error()))
		}

	case opsapi.OracleOpsRequestTypeReconfigure:
		if err := w.validateOracleReconfigurationOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("configuration"),
				req.Name,
				err.Error()))
		}

	case opsapi.OracleOpsRequestTypeReconfigureTLS:
		if err := w.validateOracleReconfigureTLSOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("tls"),
				req.Name,
				err.Error()))
		}

	case opsapi.OracleOpsRequestTypeRestart:

	case opsapi.OracleOpsRequestTypeRotateAuth:
		if err := w.validateOracleRotateAuthOpsRequest(db, req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("authentication"),
				req.Name,
				err.Error()))
		}

	case opsapi.OracleOpsRequestTypeUpdateVersion:
		if err := w.validateOracleUpdateVersionOpsRequest(db, req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("updateVersion"),
				req.Name,
				err.Error()))
		}

	case opsapi.OracleOpsRequestTypeVerticalScaling:
		if err := w.validateOracleVerticalScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("verticalScaling"),
				req.Name,
				err.Error()))
		}

	case opsapi.OracleOpsRequestTypeVolumeExpansion:
		if err := w.validateOracleVolumeExpansionOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("volumeExpansion"),
				req.Name,
				err.Error()))
		}
	}

	if len(allErr) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "oracleopsrequests.kubedb.com", Kind: "OracleOpsRequest"}, req.Name, allErr)
}

func (w *OracleOpsRequestCustomWebhook) hasDatabaseRef(req *opsapi.OracleOpsRequest) (*dbapi.Oracle, error) {
	oracle := &dbapi.Oracle{}
	if err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      req.GetDBRefName(),
		Namespace: req.GetNamespace(),
	}, oracle); err != nil {
		return nil, fmt.Errorf("spec.databaseRef %s/%s, is invalid or not found", req.GetNamespace(), req.GetDBRefName())
	}

	return oracle, nil
}

func (w *OracleOpsRequestCustomWebhook) validateOracleReconfigurationOpsRequest(req *opsapi.OracleOpsRequest) error {
	reconfigureSpec := req.Spec.Configuration
	if reconfigureSpec == nil {
		return errors.New("spec configuration nil not supported in Reconfigure type")
	}

	if !reconfigureSpec.RemoveCustomConfig && reconfigureSpec.ConfigSecret == nil && len(reconfigureSpec.ApplyConfig) == 0 {
		return errors.New("at least one of `RemoveCustomConfig`, `ConfigSecret`, or `ApplyConfig` must be specified")
	}

	if reconfigureSpec.ConfigSecret != nil && reconfigureSpec.ConfigSecret.Name != "" {
		var secret core.Secret
		err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
			Name:      reconfigureSpec.ConfigSecret.Name,
			Namespace: req.Namespace,
		}, &secret)
		if err != nil {
			if apierrors.IsNotFound(err) {
				return fmt.Errorf("referenced config secret %s/%s not found", req.Namespace, reconfigureSpec.ConfigSecret.Name)
			}
			return err
		}

		if _, ok := secret.Data[kubedb.OracleConfigFileName]; !ok {
			return fmt.Errorf("config secret %s/%s does not have file named '%v'", req.Namespace, reconfigureSpec.ConfigSecret.Name, kubedb.OracleConfigFileName)
		}
	}

	// Validate ApplyConfig has the required config file if provided
	if req.Spec.Configuration.ApplyConfig != nil {
		_, ok := req.Spec.Configuration.ApplyConfig[kubedb.OracleConfigFileName]
		if !ok {
			return fmt.Errorf("`spec.configuration.applyConfig` does not have file named '%v'", kubedb.OracleConfigFileName)
		}
	}

	return nil
}

func (w *OracleOpsRequestCustomWebhook) validateOracleReconfigureTLSOpsRequest(req *opsapi.OracleOpsRequest) error {
	TLSSpec := req.Spec.TLS
	if TLSSpec == nil {
		return errors.New("spec.TLS nil not supported in ReconfigureTLS type")
	}

	configCount := 0
	if TLSSpec.Client != nil && *TLSSpec.Client {
		configCount++
	}
	if TLSSpec.P2P != nil && *TLSSpec.P2P {
		configCount++
	}
	if TLSSpec.Remove {
		configCount++
	}
	if TLSSpec.RotateCertificates {
		configCount++
	}
	if TLSSpec.IssuerRef != nil || TLSSpec.Certificates != nil {
		configCount++
	}

	if configCount == 0 {
		return errors.New("no reconfiguration is provided in TLS spec")
	}

	if configCount > 1 {
		return errors.New("more than 1 field have assigned to spec.reconfigureTLS but at a time one is allowed to run one operation")
	}

	return nil
}

func (w *OracleOpsRequestCustomWebhook) validateOracleRotateAuthOpsRequest(db *dbapi.Oracle, req *opsapi.OracleOpsRequest) error {
	if db.Spec.DisableSecurity {
		return fmt.Errorf("disableSecurity is on, RotateAuth is not applicable")
	}

	authSpec := req.Spec.Authentication
	if authSpec != nil && authSpec.SecretRef != nil {
		if authSpec.SecretRef.Name == "" {
			return errors.New("spec.authentication.secretRef.name can not be empty")
		}
		err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
			Name:      authSpec.SecretRef.Name,
			Namespace: req.Namespace,
		}, &core.Secret{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return fmt.Errorf("referenced secret %s not found", authSpec.SecretRef.Name)
			}
			return err
		}
	}
	return nil
}

func (w *OracleOpsRequestCustomWebhook) validateOracleUpdateVersionOpsRequest(db *dbapi.Oracle, req *opsapi.OracleOpsRequest) error {
	updateVersionSpec := req.Spec.UpdateVersion
	if updateVersionSpec == nil {
		return errors.New("spec.updateVersion nil not supported in UpdateVersion type")
	}

	yes, err := IsUpgradable(w.DefaultClient, catalog.ResourceKindOracleVersion, db.Spec.Version, updateVersionSpec.TargetVersion)
	if err != nil {
		return err
	}
	if !yes {
		return fmt.Errorf("upgrade from version %v to %v is not supported", db.Spec.Version, req.Spec.UpdateVersion.TargetVersion)
	}

	return nil
}

func (w *OracleOpsRequestCustomWebhook) validateOracleVerticalScalingOpsRequest(req *opsapi.OracleOpsRequest) error {
	verticalScalingSpec := req.Spec.VerticalScaling
	if verticalScalingSpec == nil {
		return errors.New("spec.verticalScaling nil not supported in VerticalScaling type")
	}

	if verticalScalingSpec.Node == nil {
		return errors.New("spec.verticalScaling.Node can't be empty")
	}

	return nil
}

func (w *OracleOpsRequestCustomWebhook) validateOracleVolumeExpansionOpsRequest(req *opsapi.OracleOpsRequest) error {
	volumeExpansionSpec := req.Spec.VolumeExpansion
	if volumeExpansionSpec == nil {
		return errors.New("spec.volumeExpansion nil not supported in VolumeExpansion type")
	}

	if volumeExpansionSpec.Node == nil {
		return errors.New("spec.volumeExpansion.Node can't be empty")
	}

	return nil
}
