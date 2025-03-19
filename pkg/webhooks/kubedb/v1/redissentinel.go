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
	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	amv "kubedb.dev/apimachinery/pkg/validator"

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/mergepatch"
	"k8s.io/apimachinery/pkg/util/sets"
	meta_util "kmodules.xyz/client-go/meta"
	ofstv1 "kmodules.xyz/offshoot-api/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// SetupRedisSentinelWebhookWithManager registers the webhook for RedisSentinel in the manager.
func SetupRedisSentinelWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&dbapi.RedisSentinel{}).
		WithValidator(&RedisSentinelCustomWebhook{DefaultClient: mgr.GetClient()}).
		WithDefaulter(&RedisSentinelCustomWebhook{DefaultClient: mgr.GetClient()}).
		Complete()
}

type RedisSentinelCustomWebhook struct {
	DefaultClient    client.Client
	StrictValidation bool
}

var _ webhook.CustomDefaulter = &RedisSentinelCustomWebhook{}

// log is for logging in this package.
var sentinelLog = logf.Log.WithName("redissentinel-resource")

func (w RedisSentinelCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	sentinel, ok := obj.(*dbapi.RedisSentinel)
	if !ok {
		return fmt.Errorf("expected a RedisSentinel but got a %T", obj)
	}

	sentinelLog.Info("defaulting", "name", sentinel.GetName())

	if sentinel.Spec.Version == "" {
		return errors.New(`'spec.version' is missing`)
	}

	if sentinel.Spec.Halted {
		if sentinel.Spec.DeletionPolicy == dbapi.DeletionPolicyDoNotTerminate {
			return errors.New(`Can't halt, since termination policy is 'DoNotTerminate'`)
		}
		sentinel.Spec.DeletionPolicy = dbapi.DeletionPolicyHalt
	}

	var redisVersion catalogapi.RedisVersion
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name: sentinel.Spec.Version,
	}, &redisVersion)
	if err != nil {
		return errors.Wrapf(err, "failed to get RedisVersion: %s", sentinel.Spec.Version)
	}

	return sentinel.SetDefaults(&redisVersion)
}

var _ webhook.CustomValidator = &RedisSentinelCustomWebhook{}

func (w RedisSentinelCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (warnings admission.Warnings, err error) {
	sentinel, ok := obj.(*dbapi.RedisSentinel)
	if !ok {
		return nil, fmt.Errorf("expected a RedisSentinel but got a %T", obj)
	}
	sentinelLog.Info("validating", "name", sentinel.Name)
	err = w.ValidateSentinel(sentinel)
	return admission.Warnings{}, err
}

func (w RedisSentinelCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (warnings admission.Warnings, err error) {
	oldRedisSentinel, ok := oldObj.(*dbapi.RedisSentinel)
	if !ok {
		return nil, fmt.Errorf("expected a RedisSentinel but got a %T", oldRedisSentinel)
	}
	sentinel, ok := newObj.(*dbapi.RedisSentinel)
	if !ok {
		return nil, fmt.Errorf("expected a RedisSentinel but got a %T", sentinel)
	}

	var redisVersion catalogapi.RedisVersion
	err = w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name: oldRedisSentinel.Spec.Version,
	}, &redisVersion)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get redisversion: %s", oldRedisSentinel.Spec.Version)
	}

	err = oldRedisSentinel.SetDefaults(&redisVersion)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to set defaults for RedisVersion: %s", oldRedisSentinel.Spec.Version)
	}

	if oldRedisSentinel.Spec.AuthSecret == nil {
		oldRedisSentinel.Spec.AuthSecret = sentinel.Spec.AuthSecret
	}
	if err := validateSentinelUpdate(sentinel, oldRedisSentinel); err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return nil, w.ValidateSentinel(sentinel)
}

func (w RedisSentinelCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (warnings admission.Warnings, err error) {
	sentinel, ok := obj.(*dbapi.RedisSentinel)
	if !ok {
		return nil, fmt.Errorf("expected a RedisSentinel but got a %T", obj)
	}

	var rs dbapi.RedisSentinel
	err = w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      sentinel.Name,
		Namespace: sentinel.Namespace,
	}, &rs)
	if err != nil && !kerr.IsNotFound(err) {
		return nil, errors.Wrapf(err, "failed to get RedisSentinel: %s", sentinel.Name)
	} else if err == nil && rs.Spec.DeletionPolicy == dbapi.DeletionPolicyDoNotTerminate {
		return nil, fmt.Errorf(`sentinel "%v/%v" can't be terminated. To delete, change spec.deletionPolicy`, rs.Namespace, rs.Name)
	}
	return nil, nil
}

func validateSentinelUpdate(obj, oldObj *dbapi.RedisSentinel) error {
	preconditions := meta_util.PreConditionSet{
		Set: sets.New[string](
			"spec.storageType",
			"spec.storage",
			"spec.podTemplate.spec.nodeSelector",
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

func (w RedisSentinelCustomWebhook) ValidateSentinel(sentinel *dbapi.RedisSentinel) error {
	if sentinel.Spec.Version == "" {
		return errors.New(`'spec.version' is missing`)
	}

	var redisVersion catalogapi.RedisVersion
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name: sentinel.Spec.Version,
	}, &redisVersion)
	if err != nil {
		return err
	}

	if sentinel.Spec.StorageType == "" {
		return fmt.Errorf(`'spec.storageType' is missing`)
	}
	if sentinel.Spec.StorageType != dbapi.StorageTypeDurable && sentinel.Spec.StorageType != dbapi.StorageTypeEphemeral {
		return fmt.Errorf(`'spec.storageType' %s is invalid`, sentinel.Spec.StorageType)
	}
	if err := amv.ValidateStorage(w.DefaultClient, olddbapi.StorageType(sentinel.Spec.StorageType), sentinel.Spec.Storage); err != nil {
		return err
	}

	err = amv.ValidateVolumes(ofstv1.ConvertVolumes(sentinel.Spec.PodTemplate.Spec.Volumes), redisReservedVolumes)
	if err != nil {
		return err
	}
	// err = amv.ValidateMountPaths(sentinel.Spec.PodTemplate.Spec.VolumeMounts, redisReservedMountPaths)
	err = validateSentinelVolumeMountsForAllContainers(sentinel)
	if err != nil {
		return err
	}

	err = validateSentinelVersion(redisVersion.Spec.Version)
	if err != nil {
		return err
	}
	if sentinel.Spec.DisableAuth && sentinel.Spec.AuthSecret != nil {
		return fmt.Errorf("auth Secret is not supported when disableAuth is true")
	}
	err = amv.ValidateHealth(&sentinel.Spec.HealthChecker)
	if err != nil {
		return err
	}

	if w.StrictValidation {
		// Check if redisVersion is deprecated.
		// If deprecated, return error
		if redisVersion.Spec.Deprecated {
			return fmt.Errorf("redis Sentinel %s/%s is using deprecated version %v. Skipped processing",
				sentinel.Namespace, sentinel.Name, redisVersion.Name)
		}

		if err := redisVersion.ValidateSpecs(); err != nil {
			return fmt.Errorf("redis Sentinel %s/%s is using invalid redisVersion %v. Skipped processing. reason: %v", sentinel.Namespace,
				sentinel.Name, redisVersion.Name, err)
		}
	}

	if !sentinel.Spec.DisableAuth {
		if sentinel.Spec.AuthSecret != nil &&
			sentinel.Spec.AuthSecret.ExternallyManaged &&
			sentinel.Spec.AuthSecret.Name == "" {
			return fmt.Errorf("for externallyManaged auth secret, user need to provide \"redis.Spec.AuthSecret.Name\"")
		}
	}

	if sentinel.Spec.DeletionPolicy == "" {
		return fmt.Errorf(`'spec.DeletionPolicy' is missing`)
	}

	if sentinel.Spec.StorageType == dbapi.StorageTypeEphemeral && sentinel.Spec.DeletionPolicy == dbapi.DeletionPolicyHalt {
		return fmt.Errorf(`'spec.DeletionPolicy: Halt' can not be used for 'Ephemeral' storage`)
	}

	if sentinel.Spec.TLS != nil {
		if err := checkTLSSupport(redisVersion.Spec.Version); err != nil {
			return err
		}
	}

	if err := validateSentinelEnvsForAllContainers(sentinel); err != nil {
		return err
	}

	monitorSpec := sentinel.Spec.Monitor
	if monitorSpec != nil {
		if err := amv.ValidateMonitorSpec(monitorSpec); err != nil {
			return err
		}
	}

	return nil
}

func validateSentinelEnvsForAllContainers(sentinel *dbapi.RedisSentinel) error {
	var err error
	for _, container := range sentinel.Spec.PodTemplate.Spec.Containers {
		if errC := amv.ValidateEnvVar(container.Env, forbiddenRedisEnvVars, dbapi.ResourceKindRedisSentinel); errC != nil {
			if err == nil {
				err = errC
			} else {
				err = errors.Wrap(err, errC.Error())
			}
		}
	}
	return err
}

func validateSentinelVolumeMountsForAllContainers(sentinel *dbapi.RedisSentinel) error {
	var err error
	for _, container := range sentinel.Spec.PodTemplate.Spec.Containers {
		if errC := amv.ValidateMountPaths(container.VolumeMounts, redisReservedMountPaths); errC != nil {
			if err == nil {
				err = errC
			} else {
				err = errors.Wrap(err, errC.Error())
			}
		}
	}
	return err
}

func validateSentinelVersion(version string) error {
	rdVersion, err := semver.NewVersion(version)
	if err != nil {
		return err
	}
	if rdVersion.Major() < 6 || (rdVersion.Major() == 6 && rdVersion.Minor() < 2) {
		return fmt.Errorf("minimum version needed to use kubedb managed redis sentinel is 6.2.5")
	}
	return nil
}
