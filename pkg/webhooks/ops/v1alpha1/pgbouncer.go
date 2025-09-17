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

	"kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	"github.com/pkg/errors"
	"gomodules.xyz/x/arrays"
	v1 "k8s.io/api/core/v1"
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

// SetupPgBouncerOpsRequestWebhookWithManager registers the webhook for PgBouncerOpsRequest in the manager.
func SetupPgBouncerOpsRequestWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&opsapi.PgBouncerOpsRequest{}).
		WithValidator(&PgBouncerOpsRequestCustomWebhook{mgr.GetClient()}).
		Complete()
}

type PgBouncerOpsRequestCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var pgbouncerLog = logf.Log.WithName("pgbouncer-opsrequest")

var _ webhook.CustomValidator = &PgBouncerOpsRequestCustomWebhook{}

func (w *PgBouncerOpsRequestCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	req, ok := obj.(*opsapi.PgBouncerOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an PgBouncerOpsRequest object but got %T", obj)
	}

	pgbouncerLog.Info("validate create", "name", req.Name)
	return nil, w.validateCreateOrUpdate(req)
}

func (w *PgBouncerOpsRequestCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	newReq, ok := newObj.(*opsapi.PgBouncerOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an PgBouncerOpsRequest object but got %T", newObj)
	}
	oldReq, ok := oldObj.(*opsapi.PgBouncerOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an PgBouncerOpsRequest object but got %T", oldObj)
	}

	if err := validatePgBouncerOpsRequest(newReq, oldReq); err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return nil, w.validateCreateOrUpdate(newReq)
}

func (w *PgBouncerOpsRequestCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (w *PgBouncerOpsRequestCustomWebhook) validateCreateOrUpdate(obj *opsapi.PgBouncerOpsRequest) error {
	if validType, _ := arrays.Contains(opsapi.PgBouncerOpsRequestTypeNames(), string(obj.Spec.Type)); !validType {
		return fmt.Errorf("defined OpsRequestType %s is not supported, supported types for PgBouncer are %s", obj.Spec.Type, strings.Join(opsapi.PgBouncerOpsRequestTypeNames(), ", "))
	}
	if !w.isDatabaseRefValid(obj) {
		return fmt.Errorf("target database pgbouncer %s is not valid", obj.GetDBRefName())
	}

	switch obj.Spec.Type {
	case opsapi.PgBouncerOpsRequestTypeHorizontalScaling:
		return validatePgBouncerHorizontalScalingOpsRequest(obj)
	case opsapi.PgBouncerOpsRequestTypeVerticalScaling:
		return validatePgBouncerVerticalScalingOpsRequest(obj)
	case opsapi.PgBouncerOpsRequestTypeUpdateVersion:
		return validatePgBouncerUpdateVersionOpsRequest(obj, w.DefaultClient)
	case opsapi.PgBouncerOpsRequestTypeReconfigure:
		return validatePgBouncerReconfigurationOpsRequest(obj, w.DefaultClient)
	case opsapi.PgBouncerOpsRequestTypeRestart:
		return nil
	case opsapi.PgBouncerOpsRequestTypeReconfigureTLS:
		return w.validatePgBouncerReconfigureTLSOpsRequest(obj)
	}

	return nil
}

func validatePgBouncerOpsRequest(obj, oldObj runtime.Object) error {
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

func (w *PgBouncerOpsRequestCustomWebhook) isDatabaseRefValid(obj *opsapi.PgBouncerOpsRequest) bool {
	_, err := getPgBouncer(w.DefaultClient, obj)
	return err == nil
}

func getPgBouncer(client client.Client, opsReq *opsapi.PgBouncerOpsRequest) (*dbapi.PgBouncer, error) {
	bouncer := &dbapi.PgBouncer{}
	if err := client.Get(context.TODO(), types.NamespacedName{Name: opsReq.GetDBRefName(), Namespace: opsReq.Namespace}, bouncer); err != nil {
		return nil, err
	}
	return bouncer, nil
}

func (w *PgBouncerOpsRequestCustomWebhook) validatePgBouncerReconfigureTLSOpsRequest(req *opsapi.PgBouncerOpsRequest) error {
	tls := req.Spec.TLS
	if tls == nil {
		return errors.New("`spec.tls` nil not supported in ReconfigureTLS type")
	}
	return nil
}

func validatePgBouncerHorizontalScalingOpsRequest(req *opsapi.PgBouncerOpsRequest) error {
	if req.Spec.HorizontalScaling == nil {
		return errors.New("`spec.horizontalScaling` field is nil")
	}
	if req.Spec.HorizontalScaling.Replicas == nil {
		return errors.New("`spec.horizontalScaling.replicas` field is nil")
	}
	if *req.Spec.HorizontalScaling.Replicas <= int32(0) || *req.Spec.HorizontalScaling.Replicas > int32(50) {
		return errors.New("Group size can not be less than 1 or greater than 50, range: [1,50]") // todo check the max cluster size for PgBouncer
	}
	return nil
}

func validatePgBouncerVerticalScalingOpsRequest(req *opsapi.PgBouncerOpsRequest) error {
	if req.Spec.VerticalScaling == nil {
		return errors.New("`spec.Scale.Vertical` field is empty")
	}
	return nil
}

func validatePgBouncerUpdateVersionOpsRequest(req *opsapi.PgBouncerOpsRequest, client client.Client) error {
	if req.Spec.UpdateVersion == nil {
		return errors.New("`spec.updateVersion` nil not supported in UpdateVersion type")
	}

	version, err := getPgBouncerTargetVersion(client, req)
	if err != nil {
		return err
	}

	if version.IsDeprecated() {
		return errors.New(fmt.Sprintf("pgbouncerversion %s is depricated", version.Name))
	}
	return nil
}

func getPgBouncerTargetVersion(client client.Client, opsReq *opsapi.PgBouncerOpsRequest) (*v1alpha1.PgBouncerVersion, error) {
	if opsReq.Spec.Type != opsapi.PgBouncerOpsRequestTypeUpdateVersion {
		return nil, fmt.Errorf("pgbouncer version will not be updated with this ops-request")
	}
	if opsReq.Spec.UpdateVersion.TargetVersion == "" {
		return nil, fmt.Errorf("targeted pgbouncer version name is invalid")
	}
	version := v1alpha1.PgBouncerVersion{}
	err := client.Get(context.TODO(), types.NamespacedName{Namespace: opsReq.Namespace, Name: opsReq.Spec.UpdateVersion.TargetVersion}, &version)
	if err != nil {
		return nil, err
	}
	return &version, nil
}

func validatePgBouncerReconfigurationOpsRequest(req *opsapi.PgBouncerOpsRequest, client client.Client) error {
	if req.Spec.Configuration == nil {
		return errors.New("`spec.configuration` nil not supported in Reconfigure type")
	}

	bouncerConfigReq := req.Spec.Configuration.PgBouncer

	if bouncerConfigReq.ConfigSecret == nil && bouncerConfigReq.ApplyConfig == nil {
		return errors.New("custom configuration should be specified in`spec.configuration.pgbouncer` for Reconfigure type")
	}

	if bouncerConfigReq.ConfigSecret != nil && bouncerConfigReq.ApplyConfig != nil {
		return errors.New("`spec.configuration.pgbouncer.configSecret` and `spec.configuration.pgbouncer.applyConfig` both can not be taken for Reconfigure type")
	}

	if bouncerConfigReq.ConfigSecret != nil {
		var secret v1.Secret
		if err := client.Get(context.TODO(), types.NamespacedName{Name: bouncerConfigReq.ConfigSecret.Name, Namespace: req.Namespace}, &secret); err != nil {
			return err
		}
	}

	return nil
}
