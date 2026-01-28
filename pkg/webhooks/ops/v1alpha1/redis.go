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
	"errors"
	"fmt"
	"strings"

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	"github.com/Masterminds/semver/v3"
	"gomodules.xyz/x/arrays"
	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/mergepatch"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/ptr"
	meta_util "kmodules.xyz/client-go/meta"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// SetupRedisOpsRequestWebhookWithManager registers the webhook for RedisOpsRequest in the manager.
func SetupRedisOpsRequestWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&opsapi.RedisOpsRequest{}).
		WithValidator(&RedisOpsRequestCustomWebhook{mgr.GetClient()}).
		Complete()
}

type RedisOpsRequestCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var redisLog = logf.Log.WithName("redis-opsrequest")

var _ webhook.CustomValidator = &RedisOpsRequestCustomWebhook{}

func (w *RedisOpsRequestCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	req, ok := obj.(*opsapi.RedisOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an RedisOpsRequest object but got %T", obj)
	}

	redisLog.Info("validate create", "name", req.Name)
	err := w.isDatabaseRefValid(req)
	if err != nil {
		return nil, err
	}
	return nil, w.validateCreateOrUpdate(req)
}

func (w *RedisOpsRequestCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	newReq, ok := newObj.(*opsapi.RedisOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an RedisOpsRequest object but got %T", newObj)
	}
	oldReq, ok := oldObj.(*opsapi.RedisOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an RedisOpsRequest object but got %T", oldObj)
	}

	if err := validateRedisOpsRequest(newReq, oldReq); err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	err := w.isDatabaseRefValid(newReq)
	if err != nil {
		return nil, err
	}
	if err = w.validateCreateOrUpdate(newReq); err != nil {
		return nil, err
	}
	if isOpsReqCompleted(newReq.Status.Phase) && !isOpsReqCompleted(oldReq.Status.Phase) { // just completed
		var db dbapi.Redis
		err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: newReq.Spec.DatabaseRef.Name, Namespace: newReq.Namespace}, &db)
		if err != nil {
			return nil, err
		}
		return nil, resumeDatabase(w.DefaultClient, &db)
	}
	return nil, nil
}

func (w *RedisOpsRequestCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func validateRedisOpsRequest(obj, oldObj runtime.Object) error {
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

func (w *RedisOpsRequestCustomWebhook) isDatabaseRefValid(req *opsapi.RedisOpsRequest) error {
	redis := &dbapi.Redis{ObjectMeta: metav1.ObjectMeta{Name: req.Spec.DatabaseRef.Name, Namespace: req.Namespace}}
	return w.DefaultClient.Get(context.TODO(), client.ObjectKeyFromObject(redis), redis)
}

func (w *RedisOpsRequestCustomWebhook) validateCreateOrUpdate(req *opsapi.RedisOpsRequest) error {
	if validType, _ := arrays.Contains(opsapi.RedisOpsRequestTypeNames(), string(req.Spec.Type)); !validType {
		return field.Invalid(field.NewPath("spec").Child("type"), req.Name,
			fmt.Sprintf("defined OpsRequestType %s is not supported, supported types for Redis are %s", req.Spec.Type, strings.Join(opsapi.RedisOpsRequestTypeNames(), ", ")))
	}

	db, err := w.hasDatabaseRef(req)
	if err != nil {
		return err
	}
	var allErr field.ErrorList

	switch opsapi.RedisOpsRequestType(req.GetRequestType()) {
	case opsapi.RedisOpsRequestTypeRestart:

	case opsapi.RedisOpsRequestTypeHorizontalScaling:
		if err := w.validateRedisHorizontalScalingOpsRequest(db, req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("horizontalScaling"),
				req.Name,
				err.Error()))
		}
	case opsapi.RedisOpsRequestTypeVerticalScaling:
		if err := w.validateRedisVerticalScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("verticalScaling"),
				req.Name,
				err.Error()))
		}
	case opsapi.RedisOpsRequestTypeVolumeExpansion:
		if err := w.validateRedisVolumeExpansionOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("volumeExpansion"),
				req.Name,
				err.Error()))
		}
	case opsapi.RedisOpsRequestTypeUpdateVersion:
		if err := w.validateRedisUpdateVersionOpsRequest(db, req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("updateVersion"),
				req.Name,
				err.Error()))
		}
	case opsapi.RedisOpsRequestTypeReconfigure:
		if err := w.validateRedisReconfigureOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("configuration"),
				req.Name,
				err.Error()))
		}
	case opsapi.RedisOpsRequestTypeReconfigureTLS:
		if err := w.validateRedisReconfigureTLSOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("tls"),
				req.Name,
				err.Error()))
		}
	case opsapi.RedisOpsRequestTypeReplaceSentinel:
		if err := w.validateRedisReplaceSentinelOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("sentinel"),
				req.Name,
				err.Error()))
		}
	case opsapi.RedisOpsRequestTypeRotateAuth:
		if err := w.validateRedisRotateAuthenticationOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("authentication"),
				req.Name,
				err.Error()))
		}
	case opsapi.RedisOpsRequestTypeAnnounce:
		if err := w.validateRedisAnnounceOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("announce"),
				req.Name,
				err.Error()))
		}
	}
	if len(allErr) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "Redisopsrequests.kubedb.com", Kind: "RedisOpsRequest"}, req.Name, allErr)
}

func (w *RedisOpsRequestCustomWebhook) validateRedisUpdateVersionOpsRequest(db *dbapi.Redis, req *opsapi.RedisOpsRequest) error {
	updateVersionSpec := req.Spec.UpdateVersion
	if updateVersionSpec == nil {
		return errors.New("`spec.updateVersion` nil not supported in UpdateVersion type")
	}

	yes, err := IsUpgradable(w.DefaultClient, catalog.ResourceKindRedisVersion, db.Spec.Version, updateVersionSpec.TargetVersion)
	if err != nil {
		return err
	}
	if !yes {
		return fmt.Errorf("upgrade from version %v to %v is not supported", db.Spec.Version, req.Spec.UpdateVersion.TargetVersion)
	}

	updatedVersionName := req.Spec.UpdateVersion.TargetVersion
	redis := dbapi.Redis{}
	err = w.DefaultClient.Get(context.TODO(), types.NamespacedName{Namespace: req.Namespace, Name: req.GetDBRefName()}, &redis)
	if err != nil {
		return err
	}
	currentVersionName := redis.Spec.Version

	updatedVersion := &catalog.RedisVersion{}
	if err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: updatedVersionName}, updatedVersion); err != nil {
		return err
	}

	currentVersion := &catalog.RedisVersion{}
	if err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: currentVersionName}, currentVersion); err != nil {
		return err
	}

	if redis.Spec.Mode == dbapi.RedisModeSentinel {
		currentSentinel := &dbapi.RedisSentinel{}
		if err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: redis.Spec.SentinelRef.Name, Namespace: redis.Spec.SentinelRef.Namespace}, currentSentinel); err != nil {
			return err
		}
		if currentSentinel.Spec.Version != updatedVersion.Spec.Version {
			return fmt.Errorf("RedisOpsRequest %s/%s: can't upgrade to a different version of sentinel", req.Namespace, req.Name)
		}
	}

	if updatedVersion.Spec.Distribution != currentVersion.Spec.Distribution {
		updatedSemver, err := semver.NewVersion(updatedVersion.Spec.Version)
		if err != nil {
			return err
		}
		currentSemver, err := semver.NewVersion(currentVersion.Spec.Version)
		if err != nil {
			return err
		}
		if updatedSemver.Major() != currentSemver.Major() {
			return fmt.Errorf("RedisOpsRequest %s/%s: can't upgrade Official to Valkey with a different major version", req.Namespace, req.Name)
		}
	}
	return nil
}

func (w *RedisOpsRequestCustomWebhook) validateRedisHorizontalScalingOpsRequest(db *dbapi.Redis, req *opsapi.RedisOpsRequest) error {
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Namespace: req.Namespace, Name: req.GetDBRefName()}, db)
	if err != nil {
		return err
	}
	switch db.Spec.Mode {
	case dbapi.RedisModeStandalone:
		return fmt.Errorf("horizontal scaling is not allowed for standalone redis")
	case dbapi.RedisModeCluster:
		return w.checkHorizontalOpsReqForClusterMode(req)
	default:
		return w.checkHorizontalOpsReqForSentinelMode(req)
	}
}

func (w *RedisOpsRequestCustomWebhook) checkHorizontalOpsReqForClusterMode(req *opsapi.RedisOpsRequest) error {
	if req.Spec.HorizontalScaling.Shards != nil {
		if *req.Spec.HorizontalScaling.Shards < 3 {
			return fmt.Errorf("shards should be >= 3 in cluster mode")
		}
	}
	if req.Spec.HorizontalScaling.Replicas != nil {
		if *req.Spec.HorizontalScaling.Replicas < 2 {
			return fmt.Errorf("replica should be >= 2 in cluster mode")
		}
	}
	if req.Spec.HorizontalScaling.Shards == nil && req.Spec.HorizontalScaling.Replicas == nil {
		return fmt.Errorf("please specify shards or replica for scaling")
	}

	// verify announce if needed
	redis := dbapi.Redis{}
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Namespace: req.Namespace, Name: req.GetDBRefName()}, &redis)
	if err != nil {
		return err
	}
	oldRepCnt := *redis.Spec.Cluster.Replicas
	newRepCnt := ptr.Deref(req.Spec.HorizontalScaling.Replicas, *redis.Spec.Cluster.Replicas)
	oldShardCnt := *redis.Spec.Cluster.Shards
	newShardCnt := ptr.Deref(req.Spec.HorizontalScaling.Shards, *redis.Spec.Cluster.Shards)

	if (oldRepCnt <= newRepCnt) != (oldShardCnt <= newShardCnt) && (oldRepCnt >= newRepCnt) != (oldShardCnt >= newShardCnt) {
		return fmt.Errorf("can't scale down and up at the same time")
	}
	if redis.Spec.Cluster.Announce != nil {
		if req.Spec.HorizontalScaling.Announce == nil && (oldRepCnt < newRepCnt || oldShardCnt < newShardCnt) {
			return fmt.Errorf("spec.horizontalScaling.announce is required for announce ops request")
		}
		if oldShardCnt <= newShardCnt {
			shardIdx := 0
			for i := range newShardCnt {
				endpointsNeeded := 0
				if i < oldShardCnt && oldRepCnt < newRepCnt { // if we are looking endpoints for existing shards and need to add endpoints for new replicas
					endpointsNeeded = int(newRepCnt - oldRepCnt)
				} else if i >= oldShardCnt { // if need to add new shards
					endpointsNeeded = int(newRepCnt)
				}
				if endpointsNeeded > 0 {
					if len(req.Spec.HorizontalScaling.Announce.Shards) <= shardIdx {
						return fmt.Errorf("endpoints for shard %d should be specified", i)
					}
					if len(req.Spec.HorizontalScaling.Announce.Shards[shardIdx].Endpoints) != endpointsNeeded {
						return fmt.Errorf("number of endpoints for shard %d should be %d", i, endpointsNeeded)
					}
					shardIdx++
				}
			}
		}
	}

	return nil
}

func (w *RedisOpsRequestCustomWebhook) checkHorizontalOpsReqForSentinelMode(req *opsapi.RedisOpsRequest) error {
	if req.Spec.HorizontalScaling.Shards != nil {
		return fmt.Errorf("master count should not be specified for Sentinel mode")
	}
	if req.Spec.HorizontalScaling.Replicas != nil {
		if *req.Spec.HorizontalScaling.Replicas < 1 {
			return fmt.Errorf("replica should be >= 1")
		}
	} else {
		return fmt.Errorf("pleaase specify replica for scaling")
	}
	return nil
}

func (w *RedisOpsRequestCustomWebhook) validateRedisReconfigureOpsRequest(req *opsapi.RedisOpsRequest) error {
	reconfigureSpec := req.Spec.Configuration
	if reconfigureSpec == nil {
		return errors.New("`spec.configuration` nil not supported in Reconfigure type")
	}

	if !reconfigureSpec.RemoveCustomConfig && reconfigureSpec.ConfigSecret == nil && len(reconfigureSpec.ApplyConfig) == 0 {
		return errors.New("at least one of `RemoveCustomConfig`, `ConfigSecret`, or `ApplyConfig` must be specified")
	}

	return nil
}

func (w *RedisOpsRequestCustomWebhook) validateRedisRotateAuthenticationOpsRequest(req *opsapi.RedisOpsRequest) error {
	redis := dbapi.Redis{}
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Namespace: req.Namespace, Name: req.GetDBRefName()}, &redis)
	if err != nil {
		return err
	}
	if redis.Spec.DisableAuth {
		return fmt.Errorf("%s is running in disable auth mode. RotateAuth is not applicable", req.GetDBRefName())
	}

	authSpec := req.Spec.Authentication
	if authSpec != nil && authSpec.SecretRef != nil {
		if authSpec.SecretRef.Name == "" {
			return fmt.Errorf("spec.authentication.secretRef.name can not be empty")
		}
		err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
			Name:      authSpec.SecretRef.Name,
			Namespace: req.Namespace,
		}, &core.Secret{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return fmt.Errorf("referenced secret %s/%s not found", req.Namespace, authSpec.SecretRef.Name)
			}
			return err
		}
	}

	return nil
}

func (w *RedisOpsRequestCustomWebhook) validateRedisAnnounceOpsRequest(req *opsapi.RedisOpsRequest) error {
	redis := dbapi.Redis{}
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Namespace: req.Namespace, Name: req.GetDBRefName()}, &redis)
	if err != nil {
		return err
	}

	if redis.Spec.Mode != dbapi.RedisModeCluster {
		return fmt.Errorf("announce is only applicable for redis cluster mode")
	}

	if req.Spec.Announce == nil {
		return fmt.Errorf("spec.announce is required for announce ops request")
	}
	if int32(len(req.Spec.Announce.Shards)) != ptr.Deref(redis.Spec.Cluster.Shards, 0) {
		return fmt.Errorf("number of shards in announce (%d) does not match with redis cluster shards (%d)", len(req.Spec.Announce.Shards), ptr.Deref(redis.Spec.Cluster.Shards, 0))
	}
	for i, shard := range req.Spec.Announce.Shards {
		if int32(len(shard.Endpoints)) != ptr.Deref(redis.Spec.Cluster.Replicas, 0) {
			return fmt.Errorf("number of replicas in announce shard %d (%d) does not match with redis cluster replicas (%d)", i, len(shard.Endpoints), ptr.Deref(redis.Spec.Cluster.Replicas, 0))
		}
		for j, endpoint := range shard.Endpoints {
			if endpoint == "" {
				return fmt.Errorf("endpoint %d in announce shard %d is empty", j, i)
			}
		}
	}

	return nil
}

func (w *RedisOpsRequestCustomWebhook) hasDatabaseRef(req *opsapi.RedisOpsRequest) (*dbapi.Redis, error) {
	redis := &dbapi.Redis{}
	if err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      req.GetDBRefName(),
		Namespace: req.GetNamespace(),
	}, redis); err != nil {
		return nil, fmt.Errorf("spec.databaseRef %s/%s, is invalid or not found", req.GetNamespace(), req.GetDBRefName())
	}
	return redis, nil
}

func (w *RedisOpsRequestCustomWebhook) validateRedisVerticalScalingOpsRequest(req *opsapi.RedisOpsRequest) error {
	verticalScalingSpec := req.Spec.VerticalScaling
	if verticalScalingSpec == nil {
		return errors.New("`spec.verticalScaling` nil not supported in VerticalScaling type")
	}

	if verticalScalingSpec.Redis == nil && verticalScalingSpec.Exporter == nil && verticalScalingSpec.Coordinator == nil {
		return errors.New("at least one of `spec.verticalScaling.redis`, `spec.verticalScaling.exporter`, or `spec.verticalScaling.coordinator` must be specified")
	}
	return nil
}

func (w *RedisOpsRequestCustomWebhook) validateRedisVolumeExpansionOpsRequest(req *opsapi.RedisOpsRequest) error {
	if req.Spec.VolumeExpansion == nil || req.Spec.VolumeExpansion.Redis == nil {
		return errors.New("`spec.volumeExpansion.redis` field is required, can not be nil")
	}

	db := &dbapi.Redis{}
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: req.GetDBRefName(), Namespace: req.GetNamespace()}, db)
	if err != nil {
		return fmt.Errorf("failed to get redis: %s/%s: %v", req.Namespace, req.Spec.DatabaseRef.Name, err)
	}

	cur, ok := db.Spec.Storage.Resources.Requests[core.ResourceStorage]
	if !ok {
		return errors.New("failed to parse current storage size")
	}

	if cur.Cmp(*req.Spec.VolumeExpansion.Redis) >= 0 {
		return fmt.Errorf("desired storage size must be greater than current storage. Current storage: %v", cur.String())
	}

	return nil
}

func (w *RedisOpsRequestCustomWebhook) validateRedisReconfigureTLSOpsRequest(req *opsapi.RedisOpsRequest) error {
	tls := req.Spec.TLS
	if tls == nil {
		return errors.New("`spec.tls` nil not supported in ReconfigureTLS type")
	}

	configCount := 0
	if tls.Remove {
		configCount++
	}
	if tls.RotateCertificates {
		configCount++
	}
	if tls.IssuerRef != nil || tls.Certificates != nil {
		configCount++
	}

	if configCount == 0 {
		return errors.New("at least one of `Remove`, `RotateCertificates`, `IssuerRef`, or `Certificates` must be specified in TLS spec")
	}

	if configCount > 1 {
		return errors.New("only one TLS reconfiguration operation (`Remove`, `RotateCertificates`, or certificate update) is allowed at a time")
	}

	return nil
}

func (w *RedisOpsRequestCustomWebhook) validateRedisReplaceSentinelOpsRequest(req *opsapi.RedisOpsRequest) error {
	if req.Spec.Sentinel == nil {
		return errors.New("`spec.sentinel` is required for ReplaceSentinel ops request")
	}

	if req.Spec.Sentinel.Ref == nil {
		return errors.New("`spec.sentinel.ref` is required for ReplaceSentinel ops request")
	}

	if req.Spec.Sentinel.Ref.Name == "" {
		return errors.New("`spec.sentinel.ref.name` can not be empty")
	}

	redis := &dbapi.Redis{}
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: req.GetDBRefName(), Namespace: req.GetNamespace()}, redis)
	if err != nil {
		return fmt.Errorf("failed to get redis: %s/%s: %v", req.Namespace, req.Spec.DatabaseRef.Name, err)
	}

	if redis.Spec.Mode != dbapi.RedisModeSentinel {
		return errors.New("ReplaceSentinel is only applicable for Redis in Sentinel mode")
	}

	// Check if the new sentinel exists
	sentinelNamespace := req.Spec.Sentinel.Ref.Namespace
	if sentinelNamespace == "" {
		sentinelNamespace = req.Namespace
	}
	newSentinel := &dbapi.RedisSentinel{}
	err = w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: req.Spec.Sentinel.Ref.Name, Namespace: sentinelNamespace}, newSentinel)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return fmt.Errorf("referenced sentinel %s/%s not found", sentinelNamespace, req.Spec.Sentinel.Ref.Name)
		}
		return err
	}

	return nil
}
