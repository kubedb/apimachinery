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
	"kubedb.dev/apimachinery/apis/kubedb"
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
	"k8s.io/utils/ptr"
	meta_util "kmodules.xyz/client-go/meta"
	ofstv1 "kmodules.xyz/offshoot-api/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// SetupRedisWebhookWithManager registers the webhook for Redis in the manager.
func SetupRedisWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&dbapi.Redis{}).
		WithValidator(&RedisCustomWebhook{DefaultClient: mgr.GetClient()}).
		WithDefaulter(&RedisCustomWebhook{DefaultClient: mgr.GetClient()}).
		Complete()
}

type RedisCustomWebhook struct {
	DefaultClient    client.Client
	StrictValidation bool
}

var _ webhook.CustomDefaulter = &RedisCustomWebhook{}

// log is for logging in this package.
var redisLog = logf.Log.WithName("redis-resource")

func (w RedisCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	redis, ok := obj.(*dbapi.Redis)
	if !ok {
		return fmt.Errorf("expected a Redis but got a %T", obj)
	}

	redisLog.Info("defaulting", "name", redis.GetName())

	if redis.Spec.Version == "" {
		return errors.New(`'spec.version' is missing`)
	}

	if redis.Spec.Halted {
		if redis.Spec.DeletionPolicy == dbapi.DeletionPolicyDoNotTerminate {
			return errors.New(`Can't halt, since termination policy is 'DoNotTerminate'`)
		}
		redis.Spec.DeletionPolicy = dbapi.DeletionPolicyHalt
	}

	var redisVersion catalogapi.RedisVersion
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name: redis.Spec.Version,
	}, &redisVersion)
	if err != nil {
		return errors.Wrapf(err, "failed to get RedisVersion: %s", redis.Spec.Version)
	}

	return redis.SetDefaults(&redisVersion)
}

var _ webhook.CustomValidator = &RedisCustomWebhook{}

func (w RedisCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (warnings admission.Warnings, err error) {
	redis, ok := obj.(*dbapi.Redis)
	if !ok {
		return nil, fmt.Errorf("expected a Redis but got a %T", obj)
	}
	err = w.ValidateRedis(redis)
	redisLog.Info("validating", "name", redis.Name)
	return admission.Warnings{}, err
}

func (w RedisCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (warnings admission.Warnings, err error) {
	oldRedis, ok := oldObj.(*dbapi.Redis)
	if !ok {
		return nil, fmt.Errorf("expected a Redis but got a %T", oldRedis)
	}
	redis, ok := newObj.(*dbapi.Redis)
	if !ok {
		return nil, fmt.Errorf("expected a Redis but got a %T", redis)
	}

	var redisVersion catalogapi.RedisVersion
	err = w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name: oldRedis.Spec.Version,
	}, &redisVersion)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get redisversion: %s", oldRedis.Spec.Version)
	}

	err = oldRedis.SetDefaults(&redisVersion)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to set defaults for RedisVersion: %s", oldRedis.Spec.Version)
	}

	if oldRedis.Spec.AuthSecret == nil {
		oldRedis.Spec.AuthSecret = redis.Spec.AuthSecret
	}
	if err := validateRedisUpdate(redis, oldRedis); err != nil {
		return nil, fmt.Errorf("%v", err)
	}

	err = w.ValidateRedis(redis)
	return nil, err
}

func (w RedisCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (warnings admission.Warnings, err error) {
	redis, ok := obj.(*dbapi.Redis)
	if !ok {
		return nil, fmt.Errorf("expected a Redis but got a %T", obj)
	}

	var rd dbapi.Redis
	err = w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      redis.Name,
		Namespace: redis.Namespace,
	}, &rd)
	if err != nil && !kerr.IsNotFound(err) {
		return nil, errors.Wrapf(err, "failed to get Redis: %s", redis.Name)
	} else if err == nil && rd.Spec.DeletionPolicy == dbapi.DeletionPolicyDoNotTerminate {
		return nil, fmt.Errorf(`redis "%v/%v" can't be terminated. To delete, change spec.deletionPolicy`, rd.Namespace, rd.Name)
	}
	return nil, nil
}

var forbiddenRedisEnvVars = []string{
	"REDISCLI_AUTH",
	"VALKEYCLI_AUTH",
	"SENTINEL_PASSWORD",
}

var redisReservedVolumes = []string{
	kubedb.RedisDataVolumeName,
	kubedb.RedisScriptVolumeName,
	kubedb.RedisTLSVolumeName,
	kubedb.RedisExporterTLSVolumeName,
	kubedb.RedisConfigVolumeName,
	kubedb.ValkeyConfigVolumeName,
	kubedb.GitSecretVolume,
}

var redisReservedMountPaths = []string{
	kubedb.RedisDataVolumePath,
	kubedb.RedisScriptVolumePath,
	kubedb.RedisTLSVolumePath,
	kubedb.RedisConfigVolumePath,
	kubedb.ValkeyConfigVolumePath,
	kubedb.GitSecretMountPath,
}

func validateRedisUpdate(obj, oldObj *dbapi.Redis) error {
	preconditions := meta_util.PreConditionSet{
		Set: sets.New[string](
			"spec.storageType",
		),
	}
	// Once the database has been initialized, don't let update the "spec.init" section
	if oldObj.Spec.Init != nil && oldObj.Spec.Init.Initialized {
		preconditions.Insert("spec.init")
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

func validateRedisEnvsForAllContainers(redis *dbapi.Redis) error {
	var err error
	for _, container := range redis.Spec.PodTemplate.Spec.Containers {
		if errC := amv.ValidateEnvVar(container.Env, forbiddenRedisEnvVars, dbapi.ResourceKindRedis); errC != nil {
			if err == nil {
				err = errC
			} else {
				err = errors.Wrap(err, errC.Error())
			}
		}
	}
	return err
}

func validateRedisVolumeMountsForAllContainers(redis *dbapi.Redis) error {
	var err error
	for _, container := range redis.Spec.PodTemplate.Spec.Containers {
		if errC := amv.ValidateMountPaths(container.VolumeMounts, redisReservedMountPaths); errC != nil {
			if err == nil {
				err = errC
			} else {
				err = errors.Wrap(err, errC.Error())
			}
		}
	}

	if errC := amv.ValidateGitInitRoot(redis.Spec.Init, redisReservedMountPaths); errC != nil {
		if err == nil {
			err = errC
		} else {
			err = errors.Wrap(err, errC.Error())
		}
	}
	return err
}

func checkTLSSupport(v string) error {
	rdVersion, err := semver.NewVersion(v)
	if err != nil {
		return err
	}
	if rdVersion.Major() < 6 {
		return fmt.Errorf("ssl support is available only for v6 or later versions")
	}
	return nil
}

func (w RedisCustomWebhook) ValidateRedis(redis *dbapi.Redis) error {
	if redis.Spec.Version == "" {
		return errors.New(`'spec.version' is missing`)
	}
	var redisVersion catalogapi.RedisVersion
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name: redis.Spec.Version,
	}, &redisVersion)
	if err != nil {
		return err
	}

	if redis.Spec.Mode != dbapi.RedisModeStandalone && redis.Spec.Mode != dbapi.RedisModeCluster && redis.Spec.Mode != dbapi.RedisModeSentinel {
		return fmt.Errorf(`spec.mode "%v" invalid. Value must be one of "%v", "%v" or "%v"`,
			redis.Spec.Mode, dbapi.RedisModeStandalone, dbapi.RedisModeCluster, dbapi.RedisModeSentinel)
	}

	if redis.Spec.Mode == dbapi.RedisModeStandalone && ptr.Deref(redis.Spec.Replicas, 0) != 1 {
		return fmt.Errorf(`spec.replicas "%d" invalid for standalone mode. Value must be one`, ptr.Deref(redis.Spec.Replicas, 0))
	}

	if redis.Spec.Mode == dbapi.RedisModeCluster && ptr.Deref(redis.Spec.Cluster.Shards, 0) < 3 {
		return fmt.Errorf(`spec.cluster.shards "%d" invalid. Value must be >= 3`, ptr.Deref(redis.Spec.Cluster.Shards, 0))
	}

	if redis.Spec.Mode == dbapi.RedisModeCluster && ptr.Deref(redis.Spec.Cluster.Replicas, 0) < 2 {
		return fmt.Errorf(`spec.cluster.replicas "%d" invalid. Value must be > 1`, ptr.Deref(redis.Spec.Cluster.Replicas, 0))
	}
	if redis.Spec.Mode == dbapi.RedisModeSentinel && (redis.Spec.SentinelRef == nil || redis.Spec.SentinelRef.Name == "" || redis.Spec.SentinelRef.Namespace == "") {
		return fmt.Errorf("need to provide sentinelRef Name and Namespace while redis Mode set to Sentinel for redis %s/%s", redis.Namespace, redis.Name)
	}
	// For Redis Cluster Announce
	if redis.Spec.Cluster != nil && redis.Spec.Cluster.Announce != nil {
		if redis.Spec.Mode != dbapi.RedisModeCluster {
			return fmt.Errorf("spec.cluster.announce is only valid for redis cluster mode, but got %s", redis.Spec.Mode)
		}
		if int32(len(redis.Spec.Cluster.Announce.Shards)) != ptr.Deref(redis.Spec.Cluster.Shards, 0) {
			return fmt.Errorf("spec.cluster.announce.shards length %d is not equal to spec.cluster.shards %d", len(redis.Spec.Cluster.Announce.Shards), ptr.Deref(redis.Spec.Cluster.Shards, 0))
		}
		for i, shard := range redis.Spec.Cluster.Announce.Shards {
			if int32(len(shard.Endpoints)) != ptr.Deref(redis.Spec.Cluster.Replicas, 0) {
				return fmt.Errorf("spec.cluster.announce.shards[%d].endpoints length %d is not equal to spec.cluster.replicas %d", i, len(shard.Endpoints), ptr.Deref(redis.Spec.Cluster.Replicas, 0))
			}
			for j, endpoint := range shard.Endpoints {
				if endpoint == "" {
					return fmt.Errorf("spec.cluster.announce.shards[%d].endpoints[%d] is empty", i, j)
				}
			}
		}
	}

	if redis.Spec.StorageType == "" {
		return fmt.Errorf(`'spec.storageType' is missing`)
	}
	if redis.Spec.StorageType != dbapi.StorageTypeDurable && redis.Spec.StorageType != dbapi.StorageTypeEphemeral {
		return fmt.Errorf(`'spec.storageType' %s is invalid`, redis.Spec.StorageType)
	}
	if err := amv.ValidateStorage(w.DefaultClient, olddbapi.StorageType(redis.Spec.StorageType), redis.Spec.Storage); err != nil {
		return err
	}
	err = amv.ValidateVolumes(ofstv1.ConvertVolumes(redis.Spec.PodTemplate.Spec.Volumes), redisReservedVolumes)
	if err != nil {
		return err
	}
	err = validateRedisVolumeMountsForAllContainers(redis)
	if err != nil {
		return err
	}
	err = amv.ValidateHealth(&redis.Spec.HealthChecker)
	if err != nil {
		return err
	}

	if redis.Spec.DisableAuth && redis.Spec.AuthSecret != nil {
		return fmt.Errorf("auth Secret is not supported when disableAuth is true")
	}
	if redis.Spec.Mode == dbapi.RedisModeSentinel {
		err = validateVersionForSentinelMode(&redisVersion)
		if err != nil {
			return err
		}
	}
	// if secret managed externally verify auth secret name is not empty
	if !redis.Spec.DisableAuth {
		if redis.Spec.AuthSecret != nil &&
			redis.Spec.AuthSecret.ExternallyManaged &&
			redis.Spec.AuthSecret.Name == "" {
			return fmt.Errorf("for externallyManaged auth secret, user need to provide \"redis.Spec.AuthSecret.Name\"")
		}
	}

	if w.StrictValidation {
		// Check if redisVersion is deprecated.
		// If deprecated, return error
		if redisVersion.Spec.Deprecated {
			return fmt.Errorf("redis %s/%s is using deprecated version %v. Skipped processing",
				redis.Namespace, redis.Name, redisVersion.Name)
		}

		if err := redisVersion.ValidateSpecs(); err != nil {
			return fmt.Errorf("redis %s/%s is using invalid redisVersion %v. Skipped processing. reason: %v", redis.Namespace,
				redis.Name, redisVersion.Name, err)
		}
	}

	if redis.Spec.DeletionPolicy == "" {
		return fmt.Errorf(`'spec.DeletionPolicy' is missing`)
	}

	if redis.Spec.StorageType == dbapi.StorageTypeEphemeral && redis.Spec.DeletionPolicy == dbapi.DeletionPolicyHalt {
		return fmt.Errorf(`'spec.DeletionPolicy: Halt' can not be used for 'Ephemeral' storage`)
	}

	if redis.Spec.TLS != nil {
		if err := checkTLSSupport(redisVersion.Spec.Version); err != nil {
			return err
		}
	}

	if err := validateRedisEnvsForAllContainers(redis); err != nil {
		return err
	}

	monitorSpec := redis.Spec.Monitor
	if monitorSpec != nil {
		if err := amv.ValidateMonitorSpec(monitorSpec); err != nil {
			return err
		}
	}

	return nil
}

func ValidateForSentinel(kbClient client.Client, redis *dbapi.Redis) error {
	if redis.Spec.Mode != dbapi.RedisModeSentinel {
		return nil
	}
	if redis.Spec.SentinelRef == nil || redis.Spec.SentinelRef.Name == "" || redis.Spec.SentinelRef.Namespace == "" {
		return fmt.Errorf("need to provide sentinelRef Name and Namespace while redis Mode set to Sentinel for redis %s/%s", redis.Namespace, redis.Name)
	}

	var sentinelDB dbapi.RedisSentinel
	err := kbClient.Get(context.TODO(), types.NamespacedName{
		Name:      redis.Spec.SentinelRef.Name,
		Namespace: redis.Spec.SentinelRef.Namespace,
	}, &sentinelDB)
	if err != nil {
		return err
	}

	rdVersion := catalogapi.RedisVersion{}
	if err = kbClient.Get(context.TODO(), types.NamespacedName{Name: redis.Spec.Version}, &rdVersion); err != nil {
		return err
	}
	if err = checkSentinelDistributionMatches(kbClient, &rdVersion, redis.Spec.SentinelRef); err != nil {
		return err
	}

	if redis.Spec.TLS == nil && sentinelDB.Spec.TLS != nil {
		return fmt.Errorf("can not start monitoring with TLS enabled Sentinel as Redis is not TLS enabled")
	}
	if redis.Spec.TLS != nil && sentinelDB.Spec.TLS == nil {
		return fmt.Errorf("can not start monitoring with TLS disabled Sentinel as Redis is TLS is enabled")
	}
	if redis.Spec.TLS != nil {
		if redis.Spec.TLS.IssuerRef.Kind != sentinelDB.Spec.TLS.IssuerRef.Kind ||
			redis.Spec.TLS.IssuerRef.Name != sentinelDB.Spec.TLS.IssuerRef.Name {
			return fmt.Errorf("can not use different issuer for redis and sentinel")
		}
	}
	return nil
}

func checkSentinelDistributionMatches(kbClient client.Client, rdVersion *catalogapi.RedisVersion, sentinelRef *dbapi.RedisSentinelRef) error {
	sentinelDB := dbapi.RedisSentinel{}
	err := kbClient.Get(context.TODO(), types.NamespacedName{Name: sentinelRef.Name, Namespace: sentinelRef.Namespace}, &sentinelDB)
	if err != nil {
		return err
	}
	sentinelVersion := catalogapi.RedisVersion{}
	err = kbClient.Get(context.TODO(), types.NamespacedName{Name: sentinelDB.Spec.Version}, &sentinelVersion)
	if err != nil {
		return err
	}
	if rdVersion.Spec.Distribution == sentinelVersion.Spec.Distribution {
		return nil
	}
	return fmt.Errorf("redis distribution %v is not compatible with sentinel distribution %v", rdVersion.Spec.Distribution, sentinelVersion.Spec.Distribution)
}
