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

package v1

import (
	"context"
	"fmt"

	catalogapi "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"
	amv "kubedb.dev/apimachinery/pkg/validator"

	"github.com/pkg/errors"
	kerr "k8s.io/apimachinery/pkg/api/errors"
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

// SetupProxySQLWebhookWithManager registers the webhook for ProxySQL in the manager.
func SetupProxySQLWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&dbapi.ProxySQL{}).
		WithValidator(&ProxySQLCustomWebhook{DefaultClient: mgr.GetClient()}).
		WithDefaulter(&ProxySQLCustomWebhook{DefaultClient: mgr.GetClient()}).
		Complete()
}

var proxyLog = logf.Log.WithName("proxysql-resource")

type ProxySQLCustomWebhook struct {
	DefaultClient    client.Client
	StrictValidation bool
}

var _ webhook.CustomDefaulter = &ProxySQLCustomWebhook{}

func (w ProxySQLCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	db := obj.(*dbapi.ProxySQL)
	proxyLog.Info("defaulting", "name", db.GetName())
	if db.Spec.Version == "" {
		return errors.New(`'spec.version' is missing`)
	}

	var psVersion catalogapi.ProxySQLVersion
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name: db.Spec.Version,
	}, &psVersion)
	if err != nil {
		return errors.Wrap(err, "failed to get the proxy SQL version")
	}

	usesAcme := false
	if db.Spec.TLS != nil {
		usesAcme, err = dbapi.UsesAcmeIssuer(w.DefaultClient, db.Namespace, *db.Spec.TLS.IssuerRef)
		if err != nil {
			return err
		}
	}

	SetDefaultsWithProxyMonitorPort(db, usesAcme, &psVersion)
	return nil
}

var _ webhook.CustomValidator = &ProxySQLCustomWebhook{}

func (w ProxySQLCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (warnings admission.Warnings, err error) {
	proxysql := obj.(*dbapi.ProxySQL)
	proxyLog.Info("validating", "name", proxysql.Name)
	err = w.ValidateProxySQL(proxysql)
	return admission.Warnings{}, err
}

func (w ProxySQLCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	oldProxy, ok := oldObj.(*dbapi.ProxySQL)
	if !ok {
		return nil, fmt.Errorf("expected a Postgres but got a %T", oldProxy)
	}
	proxy, ok := newObj.(*dbapi.ProxySQL)
	if !ok {
		return nil, fmt.Errorf("expected a Postgres but got a %T", proxy)
	}

	var proxysqlVersion catalogapi.ProxySQLVersion
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name: oldProxy.Spec.Version,
	}, &proxysqlVersion)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get ProxySQLVersion : %s", oldProxy.Spec.Version)
	}
	usesAcme := false
	if oldProxy.Spec.TLS != nil {
		usesAcme, err = dbapi.UsesAcmeIssuer(w.DefaultClient, oldProxy.Namespace, *oldProxy.Spec.TLS.IssuerRef)
		if err != nil {
			return nil, err
		}
	}
	oldProxy.SetDefaults(&proxysqlVersion, usesAcme)

	if oldProxy.Spec.AuthSecret == nil {
		oldProxy.Spec.AuthSecret = proxy.Spec.AuthSecret
	}
	if err := proxyValidateUpdate(newObj, oldObj); err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	err = w.ValidateProxySQL(proxy)
	return nil, err
}

func (w ProxySQLCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	proxysql, ok := obj.(*dbapi.ProxySQL)
	if !ok {
		return nil, fmt.Errorf("expected a ProxySQL but got a %T", obj)
	}

	var ps dbapi.ProxySQL
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      proxysql.Name,
		Namespace: proxysql.Namespace,
	}, &ps)
	if err != nil && !kerr.IsNotFound(err) {
		return nil, errors.Wrapf(err, "failed to get ProxySQL: %s", proxysql.Name)
	} else if err == nil && ps.Spec.DeletionPolicy == dbapi.DeletionPolicyDoNotTerminate {
		return nil, fmt.Errorf(`proxysql "%v/%v" can't be terminated. To delete, change spec.deletionPolicy`, ps.Namespace, ps.Name)
	}
	return nil, nil
}

// ValidateProxySQL checks if the object satisfies all the requirements.
// It is not method of Interface, because it is referenced from controller package too.
func (w ProxySQLCustomWebhook) ValidateProxySQL(db *dbapi.ProxySQL) error {
	if db.Spec.Version == "" {
		return errors.New(`'spec.version' is missing`)
	}

	var proxysqlVersion catalogapi.ProxySQLVersion
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: db.Spec.Version}, &proxysqlVersion)
	if err != nil {
		return err
	}
	if db.Spec.Replicas == nil {
		return errors.New("'.spec.replicas' is missing")
	}

	if db.Spec.Backend == nil || db.Spec.Backend.Name == "" {
		return errors.New("backend specification missing")
	}

	if err := proxyValidateEnvsForAllContainers(db); err != nil {
		return err
	}

	if db.Spec.AuthSecret != nil && db.Spec.AuthSecret.ExternallyManaged && db.Spec.AuthSecret.Name == "" {
		return fmt.Errorf("for externallyManaged auth secret, user need to provide \"proxysql.Spec.AuthSecret.Name\"")
	}

	if w.StrictValidation {
		// Check if proxysql Version is deprecated.
		// If deprecated, return error
		if proxysqlVersion.Spec.Deprecated {
			return fmt.Errorf("proxysql %s/%s is using deprecated version %v. Skipped processing", db.Namespace, db.Name, proxysqlVersion.Name)
		}
	}

	monitorSpec := db.Spec.Monitor
	if monitorSpec != nil {
		if err = amv.ValidateMonitorSpec(monitorSpec); err != nil {
			return err
		}
	}

	if db.Spec.HealthChecker.PeriodSeconds != nil && *db.Spec.HealthChecker.PeriodSeconds <= 0 {
		return fmt.Errorf(`spec.healthCheck.periodSeconds: can not be less than 1`)
	}

	if db.Spec.HealthChecker.TimeoutSeconds != nil && *db.Spec.HealthChecker.TimeoutSeconds <= 0 {
		return fmt.Errorf(`spec.healthCheck.timeoutSeconds: can not be less than 1`)
	}

	if db.Spec.HealthChecker.FailureThreshold != nil && *db.Spec.HealthChecker.FailureThreshold <= 0 {
		return fmt.Errorf(`spec.healthCheck.failureThreshold: can not be less than 1`)
	}

	return nil
}

var forbiddenEnvVarsForProxySQL = []string{
	"MYSQL_ROOT_PASSWORD",
	"MYSQL_PROXY_USER",
	"MYSQL_PROXY_PASSWORD",
}

func proxyValidateEnvsForAllContainers(proxy *dbapi.ProxySQL) error {
	var err error
	for _, container := range proxy.Spec.PodTemplate.Spec.Containers {
		if errC := amv.ValidateEnvVar(container.Env, forbiddenEnvVarsForProxySQL, dbapi.ResourceKindProxySQL); errC != nil {
			if err == nil {
				err = errC
			} else {
				err = errors.Wrap(err, errC.Error())
			}
		}
	}
	return err
}

func proxyValidateUpdate(obj, oldObj runtime.Object) error {
	preconditions := meta_util.PreConditionSet{
		Set: sets.New[string](
			"spec.initConfig",
			"spec.backend",
		),
	}
	_, err := meta_util.CreateStrategicPatch(oldObj, obj, preconditions.PreconditionFunc()...)
	if err != nil {
		if mergepatch.IsPreconditionFailed(err) {
			return fmt.Errorf("%v.%v", err, preconditions.Error())
		}
		return err
	}
	return nil
}

func SetDefaultsWithProxyMonitorPort(ps *dbapi.ProxySQL, usesAcme bool, psVersion *catalogapi.ProxySQLVersion) {
	ps.SetDefaults(psVersion, usesAcme)
	if ps.Spec.Monitor != nil {
		if ps.Spec.Monitor.Prometheus != nil {
			ps.Spec.Monitor.Prometheus.Exporter.Port = 6070
		}
	}
}
