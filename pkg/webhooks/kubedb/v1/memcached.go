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

	core "k8s.io/api/core/v1"
	catalogapi "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	"kubedb.dev/apimachinery/apis/kubedb"
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"
	amv "kubedb.dev/apimachinery/pkg/validator"

	"github.com/pkg/errors"
	"gomodules.xyz/pointer"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/mergepatch"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/ptr"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	ofst "kmodules.xyz/offshoot-api/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// SetupMemcachedWebhookWithManager registers the webhook for Memcached in the manager.
func SetupMemcachedWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&dbapi.Memcached{}).
		WithValidator(&MemcachedCustomWebhook{DefaultClient: mgr.GetClient()}).
		WithDefaulter(&MemcachedCustomWebhook{DefaultClient: mgr.GetClient()}).
		Complete()
}

type MemcachedCustomWebhook struct {
	DefaultClient    client.Client
	Client           kubernetes.Interface
	StrictValidation bool
}

var reservedMemcachedMountPaths = []string{
	kubedb.MemcachedDataVolumePath,
	kubedb.MemcachedConfigVolumePath,
}

var reservedMemcachedVolumes = []string{
	kubedb.MemcachedDataVolumeName,
	kubedb.MemcachedConfigVolumeName,
}

// No forbidden envs yet
var forbiddenMemcachedEnvVars = []string{}

var _ webhook.CustomDefaulter = &MemcachedCustomWebhook{}

// log is for logging in this package.
var memLog = logf.Log.WithName("memcached-resource")

func (mv *MemcachedCustomWebhook) Default(_ context.Context, obj runtime.Object) error {
	db, ok := obj.(*dbapi.Memcached)
	if !ok {
		return fmt.Errorf("expected a Memcached but got a %T", obj)
	}

	memLog.Info("defaulting", "name", db.GetName())

	if db.Spec.Version == "" {
		return errors.New(`'spec.version' is missing`)
	}

	if db.Spec.Halted {
		if db.Spec.DeletionPolicy == dbapi.DeletionPolicyDoNotTerminate {
			return errors.New(`Can't halt, since termination policy is 'DoNotTerminate'`)
		}
		db.Spec.DeletionPolicy = dbapi.DeletionPolicyHalt
	}

	if db.Spec.Replicas == nil {
		db.Spec.Replicas = pointer.Int32P(1)
	}
	var memcachedVersion catalogapi.MemcachedVersion
	err := mv.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name: db.Spec.Version,
	}, &memcachedVersion)
	if err != nil {
		return errors.Wrapf(err, "failed to get MemcachedVersion: %s", db.Spec.Version)
	}

	db.SetDefaults(&memcachedVersion)

	return nil
}

var _ webhook.CustomValidator = &MemcachedCustomWebhook{}

func (mv MemcachedCustomWebhook) validateEnvsForAllContainers(memcached *dbapi.Memcached) error {
	var err error
	for _, container := range memcached.Spec.PodTemplate.Spec.Containers {
		if errC := amv.ValidateEnvVar(container.Env, forbiddenMemcachedEnvVars, dbapi.ResourceKindMemcached); errC != nil {
			if err == nil {
				err = errC
			} else {
				err = errors.Wrap(err, errC.Error())
			}
		}
	}
	return err
}

func (mv MemcachedCustomWebhook) validateVolumeMountsForAllContainers(memcached *dbapi.Memcached) error {
	var err error
	for _, container := range memcached.Spec.PodTemplate.Spec.Containers {
		if errC := amv.ValidateMountPaths(container.VolumeMounts, reservedMemcachedMountPaths); errC != nil {
			if err == nil {
				err = errC
			} else {
				err = errors.Wrap(err, errC.Error())
			}
		}
	}
	return err
}

func (mv MemcachedCustomWebhook) validateReplica(obj, oldObj *dbapi.Memcached) error {
	if !(obj.Spec.Halted || oldObj.Spec.Halted) && (*oldObj.Spec.Replicas == 1 || ptr.Deref(obj.Spec.Replicas, 0) == 1) && *oldObj.Spec.Replicas != *obj.Spec.Replicas {
		return fmt.Errorf("can not update from %d replica to %d replica", ptr.Deref(oldObj.Spec.Replicas, 0), ptr.Deref(obj.Spec.Replicas, 0))
	}
	return nil
}

func (mv MemcachedCustomWebhook) validateUpdate(obj, oldObj *dbapi.Memcached) error {
	preconditions := meta_util.PreConditionSet{
		Set: sets.New[string](
			"spec.podTemplate.spec.nodeSelector",
			"spec.databaseSecret",
		),
	}

	_, err := meta_util.CreateStrategicPatch(oldObj, obj, preconditions.PreconditionFunc()...)
	if err != nil {
		if mergepatch.IsPreconditionFailed(err) {
			return fmt.Errorf("%v.%v", err, preconditions.Error())
		}
		return err
	}
	return mv.validateReplica(obj, oldObj)
}

func (mv MemcachedCustomWebhook) validateSpecForDB(memcached *dbapi.Memcached) error {
	container := core_util.GetContainerByName(memcached.Spec.PodTemplate.Spec.Containers, kubedb.MemcachedContainerName)
	if container == nil {
		return fmt.Errorf("memcached container %s not found", kubedb.MemcachedContainerName)
	}
	return nil
}

func (mv MemcachedCustomWebhook) validate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	memcached, ok := obj.(*dbapi.Memcached)
	if !ok {
		return nil, fmt.Errorf("expected a Memcached but got a %T", obj)
	}

	if memcached.Spec.Version == "" {
		return nil, errors.New("spec.version is missing")
	}
	var memcachedVersion catalogapi.MemcachedVersion
	err := mv.DefaultClient.Get(context.TODO(), client.ObjectKey{
		Name: memcached.Spec.Version,
	}, &memcachedVersion)
	if err != nil {
		return nil, err
	}

	if memcached.Spec.Replicas == nil || ptr.Deref(memcached.Spec.Replicas, 0) < 1 {
		return nil, fmt.Errorf(`spec.replicas "%d" invalid. Value must be greater than zero`, ptr.Deref(memcached.Spec.Replicas, 0))
	}

	// if secret name is given, check the secret exists
	// if secret name is given, check the secret exists
	if memcached.Spec.ConfigSecret != nil {
		var configSecret core.Secret
		// Get the configSecret
		err := mv.DefaultClient.Get(ctx, types.NamespacedName{
			Name:      memcached.Spec.ConfigSecret.Name,
			Namespace: memcached.Namespace,
		}, &configSecret)
		if err != nil {
			if !kerr.IsNotFound(err) {
				return nil, fmt.Errorf(`no configSecret is found named %v, in namespace %v`, memcached.Spec.ConfigSecret.Name, memcached.Namespace)
			}
			return nil, err
		}
	}
	if memcached.Spec.AuthSecret != nil {
		var authSecret core.Secret
		// Get the authSecret
		err := mv.DefaultClient.Get(ctx, types.NamespacedName{
			Name:      memcached.GetMemcachedAuthSecretName(),
			Namespace: memcached.Namespace,
		}, &authSecret)
		if err != nil {
			if !kerr.IsNotFound(err) {
				return nil, fmt.Errorf(`no authSecret is found named %v, in namespace %v`, memcached.GetMemcachedAuthSecretName(), memcached.Namespace)
			}
			return nil, err
		}
	}
	// if secret managed externally verify auth secret name is not empty
	if memcached.Spec.AuthSecret != nil &&
		memcached.Spec.AuthSecret.ExternallyManaged &&
		memcached.Spec.AuthSecret.Name == "" {
		return nil, fmt.Errorf(`for externallyManaged auth secret, user must configure "spec.authSecret.name"`)
	}

	if err := mv.validateEnvsForAllContainers(memcached); err != nil {
		return nil, err
	}

	err = amv.ValidateVolumes(ofst.ConvertVolumes(memcached.Spec.PodTemplate.Spec.Volumes), reservedMemcachedVolumes)
	if err != nil {
		return nil, err
	}
	err = mv.validateVolumeMountsForAllContainers(memcached)
	if err != nil {
		return nil, err
	}

	err = mv.validateSpecForDB(memcached)
	if err != nil {
		return nil, err
	}

	if mv.StrictValidation {

		// Check if memcachedVersion is deprecated.
		// If deprecated, return error

		if memcachedVersion.Spec.Deprecated {
			return nil, fmt.Errorf("memcached %s/%s is using deprecated version %v. Skipped processing",
				memcached.Namespace, memcached.Name, memcachedVersion.Name)
		}

		if err := memcachedVersion.ValidateSpecs(); err != nil {
			return nil, fmt.Errorf("memcached %s/%s is using invalid memcachedVersion %v. Skipped processing. reason: %v", memcached.Namespace,
				memcached.Name, memcachedVersion.Name, err)
		}
	}

	if memcached.Spec.DeletionPolicy == "" {
		return nil, fmt.Errorf(`'spec.deletionPolicy' is missing`)
	}

	monitorSpec := memcached.Spec.Monitor
	if monitorSpec != nil {
		if err := amv.ValidateMonitorSpec(monitorSpec); err != nil {
			return nil, err
		}
	}

	if err = amv.ValidateHealth(&memcached.Spec.HealthChecker); err != nil {
		return nil, err
	}

	return nil, nil
}

func (mv MemcachedCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	memcached := obj.(*dbapi.Memcached)
	memLog.Info("validating", "name", memcached.Name)
	return mv.validate(ctx, obj)
}

func (mv MemcachedCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	oldMemcached, ok := oldObj.(*dbapi.Memcached)
	if !ok {
		return nil, fmt.Errorf("expected a Memcached but got a %T", oldMemcached)
	}

	memcached := newObj.(*dbapi.Memcached)
	if !ok {
		return nil, fmt.Errorf("expected a Memcached but got a %T", memcached)
	}

	var memcachedVersion catalogapi.MemcachedVersion
	err := mv.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name: oldMemcached.Spec.Version,
	}, &memcachedVersion)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get MemcachedVersion: %s", oldMemcached.Spec.Version)
	}

	oldMemcached.SetDefaults(&memcachedVersion)

	if oldMemcached.Spec.AuthSecret == nil {
		oldMemcached.Spec.AuthSecret = memcached.Spec.AuthSecret
	}
	if err := mv.validateUpdate(memcached, oldMemcached); err != nil {
		return nil, fmt.Errorf("%v", err)
	}

	return mv.validate(ctx, newObj)
}

func (mv MemcachedCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	memcached, ok := obj.(*dbapi.Memcached)
	if !ok {
		return nil, fmt.Errorf("expected a Memcached but got a %T", obj)
	}

	var mc dbapi.Memcached
	err := mv.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      memcached.Name,
		Namespace: memcached.Namespace,
	}, &mc)
	if err != nil && !kerr.IsNotFound(err) {
		return nil, errors.Wrapf(err, "failed to get memcached %s", memcached.Name)
	} else if err == nil && mc.Spec.DeletionPolicy == dbapi.DeletionPolicyDoNotTerminate {
		return nil, fmt.Errorf("memcached %v/%v is can't terminated. To delete, change spec.deletionPolicy", mc.Namespace, mc.Name)
	}
	return nil, nil
}
