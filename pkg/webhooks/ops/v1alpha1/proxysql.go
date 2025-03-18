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
	"encoding/json"
	"fmt"
	"strings"

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	"github.com/pkg/errors"
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

// SetupProxySQLOpsRequestWebhookWithManager registers the webhook for ProxySQLOpsRequest in the manager.
func SetupProxySQLOpsRequestWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&opsapi.ProxySQLOpsRequest{}).
		WithValidator(&ProxySQLOpsRequestCustomWebhook{mgr.GetClient()}).
		Complete()
}

type ProxySQLOpsRequestCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var proxyLog = logf.Log.WithName("proxysql-opsrequest")

var _ webhook.CustomValidator = &ProxySQLOpsRequestCustomWebhook{}

func validateProxySQLOpsRequest(obj, oldObj runtime.Object) error {
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

// ValidateCreate implements webhooin.Validator so a webhook will be registered for the type
func (w *ProxySQLOpsRequestCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	ops, ok := obj.(*opsapi.ProxySQLOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an ProxySQLOpsRequest object but got %T", obj)
	}
	proxyLog.Info("validate create", "name", ops.Name)
	return nil, w.validateCreateOrUpdate(ops)
}

// ValidateUpdate implements webhooin.Validator so a webhook will be registered for the type
func (w *ProxySQLOpsRequestCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	ops, ok := newObj.(*opsapi.ProxySQLOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an ProxySQLOpsRequest object but got %T", newObj)
	}
	proxyLog.Info("validate update", "name", ops.Name)

	oldOps, ok := oldObj.(*opsapi.ProxySQLOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an ProxySQLOpsRequest object but got %T", oldObj)
	}

	if err := validateProxySQLOpsRequest(ops, oldOps); err != nil {
		return nil, err
	}
	return nil, w.validateCreateOrUpdate(ops)
}

func (w *ProxySQLOpsRequestCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (w *ProxySQLOpsRequestCustomWebhook) validateCreateOrUpdate(req *opsapi.ProxySQLOpsRequest) error {
	var allErr field.ErrorList
	switch req.GetRequestType().(opsapi.ProxySQLOpsRequestType) {
	case opsapi.ProxySQLOpsRequestTypeRestart:
		if err := w.hasDatabaseRef(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("restart"),
				req.Name,
				err.Error()))
		}
	case opsapi.ProxySQLOpsRequestTypeVerticalScaling:
		if err := w.validateProxySQLScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("verticalScaling"),
				req.Name,
				err.Error()))
		}
	case opsapi.ProxySQLOpsRequestTypeHorizontalScaling:
		if err := w.validateProxySQLScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("horizontalScaling"),
				req.Name,
				err.Error()))
		}
	case opsapi.ProxySQLOpsRequestTypeReconfigure:
		if err := w.validateProxySQLReconfigurationOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("configuration"),
				req.Name,
				err.Error()))
		}
	case opsapi.ProxySQLOpsRequestTypeUpdateVersion:
		if err := w.validateProxySQLUpgradeOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("updateVersion"),
				req.Name,
				err.Error()))
		}
	case opsapi.ProxySQLOpsRequestTypeReconfigureTLS:
		if err := w.validateProxySQLReconfigurationTLSOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("tls"),
				req.Name,
				err.Error()))
		}
	default:
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("type"), req.Name,
			fmt.Sprintf("defined OpsRequestType %s is not supported, supported types for ProxySQL are %s", req.Spec.Type, strings.Join(opsapi.ProxySQLOpsRequestTypeNames(), ", "))))
	}
	if len(allErr) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "ProxySQLopsrequests.kubedb.com", Kind: "ProxySQLOpsRequest"}, req.Name, allErr)
}

func (w *ProxySQLOpsRequestCustomWebhook) hasDatabaseRef(req *opsapi.ProxySQLOpsRequest) error {
	prx := dbapi.ProxySQL{}
	if err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      req.GetDBRefName(),
		Namespace: req.GetNamespace(),
	}, &prx); err != nil {
		return errors.New(fmt.Sprintf("spec.databaseRef %s/%s, is invalid or not found", req.GetNamespace(), req.GetDBRefName()))
	}
	return nil
}

func (w *ProxySQLOpsRequestCustomWebhook) validateProxySQLUpgradeOpsRequest(req *opsapi.ProxySQLOpsRequest) error {
	if req.Spec.UpdateVersion == nil {
		return errors.New("spec.Upgrade is nil")
	}
	db := &dbapi.ProxySQL{}
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: req.GetDBRefName(), Namespace: req.GetNamespace()}, db)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to get proxysql: %s/%s", req.Namespace, req.Spec.ProxyRef.Name))
	}
	prxNextVersion := &catalog.ProxySQLVersion{}
	err = w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: req.Spec.UpdateVersion.TargetVersion}, prxNextVersion)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to get proxysqlVersion: %s", req.Spec.UpdateVersion.TargetVersion))
	}
	// check if prxNextVersion is deprecated.if deprecated, return error
	if prxNextVersion.Spec.Deprecated {
		return fmt.Errorf("proxysql target version %s/%s is deprecated. Skipped processing", db.Namespace, prxNextVersion.Name)
	}
	return nil
}

func (w *ProxySQLOpsRequestCustomWebhook) validateProxySQLScalingOpsRequest(req *opsapi.ProxySQLOpsRequest) error {
	if req.Spec.Type == opsapi.ProxySQLOpsRequestTypeHorizontalScaling {
		if req.Spec.HorizontalScaling == nil {
			return errors.New("`spec.Scale.HorizontalScaling` field is nil")
		}

		if *req.Spec.HorizontalScaling.Member <= int32(2) || *req.Spec.HorizontalScaling.Member > int32(50) {
			return errors.New("Group size can not be less than 3 or greater than 50, range: [3,50]") // todo check the max cluster size for proxysql
		}
		return nil
	}

	if req.Spec.VerticalScaling == nil {
		return errors.New("`spec.Scale.Vertical` field is empty")
	}

	return nil
}

func (w *ProxySQLOpsRequestCustomWebhook) validateProxySQLReconfigurationOpsRequest(req *opsapi.ProxySQLOpsRequest) error {
	qRulesConfig := req.Spec.Configuration.MySQLQueryRules
	if qRulesConfig != nil {
		mp, err := GetMySQLQueryRulesMapConfig(qRulesConfig.Rules)
		if err != nil {
			return err
		}
		for j := range mp {
			_, exist := mp[j]["rule_id"]
			if !exist {
				msg := fmt.Sprintf(`"rule_id" missing in spec.configuration.mysqlQueryRules.rules[%d]"`, j+1)
				return errors.New(msg)
			}
		}
	}

	return nil
}

func (w *ProxySQLOpsRequestCustomWebhook) validateProxySQLReconfigurationTLSOpsRequest(req *opsapi.ProxySQLOpsRequest) error {
	db := &dbapi.ProxySQL{}
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: req.GetDBRefName(), Namespace: req.GetNamespace()}, db)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to get proxysql: %s/%s", req.Namespace, req.Spec.ProxyRef.Name))
	}
	dbVersion := &catalog.ProxySQLVersion{}
	err = w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: db.Spec.Version}, dbVersion)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to get proxysqlversion: %s/%s", req.Namespace, req.Spec.ProxyRef.Name))
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
	if req.Spec.TLS.TLSConfig.IssuerRef != nil || req.Spec.TLS.TLSConfig.Certificates != nil {
		configCount++
	}

	if configCount == 0 {
		return errors.New("incomplete reconfiguration is provided in TLS Spec.")
	}

	if configCount > 1 {
		return errors.New("more than 1 field have assigned to reconfigureTLS to your database but at a time you you are allowed to run one operation")
	}
	return nil
}

func GetMySQLQueryRulesMapConfig(rules []*runtime.RawExtension) ([]map[string]interface{}, error) {
	var ruleArray []map[string]interface{}
	for i := range rules {
		cur := rules[i]
		data, err := json.Marshal(cur)
		if err != nil {
			return nil, err
		}
		rule := make(map[string]interface{})
		err = json.Unmarshal(data, &rule)
		if err != nil {
			return nil, err
		}
		ruleArray = append(ruleArray, rule)
	}
	return ruleArray, nil
}
