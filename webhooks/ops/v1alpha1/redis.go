package v1alpha1

import (
	"context"
	"fmt"

	v1 "kubedb.dev/apimachinery/apis/kubedb/v1"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/mergepatch"
	"k8s.io/apimachinery/pkg/util/sets"
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

func (in *RedisOpsRequestCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	req, ok := obj.(*opsapi.RedisOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an RedisOpsRequest object but got %T", obj)
	}
	return nil, in.isDatabaseRefValid(req)
}

func (in *RedisOpsRequestCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
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
	return nil, in.isDatabaseRefValid(newReq)
}

func (in *RedisOpsRequestCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
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

func (in *RedisOpsRequestCustomWebhook) isDatabaseRefValid(req *opsapi.RedisOpsRequest) error {
	redis := &v1.Redis{ObjectMeta: metav1.ObjectMeta{Name: req.Spec.DatabaseRef.Name, Namespace: req.Namespace}}
	return in.DefaultClient.Get(context.TODO(), client.ObjectKeyFromObject(redis), redis)
}
