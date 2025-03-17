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

// SetupRedisSentinelOpsRequestWebhookWithManager registers the webhook for RedisSentinelOpsRequest in the manager.
func SetupRedisSentinelOpsRequestWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&opsapi.RedisSentinelOpsRequest{}).
		WithValidator(&RedisSentinelOpsRequestCustomWebhook{mgr.GetClient()}).
		Complete()
}

type RedisSentinelOpsRequestCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var redissentinelLog = logf.Log.WithName("redissentinel-opsrequest")

var _ webhook.CustomValidator = &RedisSentinelOpsRequestCustomWebhook{}

func (in *RedisSentinelOpsRequestCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	req, ok := obj.(*opsapi.RedisSentinelOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an RedisSentinelOpsRequest object but got %T", obj)
	}
	return nil, in.isDatabaseRefValid(req)
}

func (in *RedisSentinelOpsRequestCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	newReq, ok := newObj.(*opsapi.RedisSentinelOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an RedisSentinelOpsRequest object but got %T", newObj)
	}
	oldReq, ok := oldObj.(*opsapi.RedisSentinelOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an RedisSentinelOpsRequest object but got %T", oldObj)
	}

	if err := validateRedisSentinelOpsRequest(newReq, oldReq); err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return nil, in.isDatabaseRefValid(newReq)
}

func (in *RedisSentinelOpsRequestCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func validateRedisSentinelOpsRequest(obj, oldObj runtime.Object) error {
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

func (in *RedisSentinelOpsRequestCustomWebhook) isDatabaseRefValid(obj *opsapi.RedisSentinelOpsRequest) error {
	rs := &v1.RedisSentinel{ObjectMeta: metav1.ObjectMeta{Name: obj.Spec.DatabaseRef.Name, Namespace: obj.Namespace}}
	return in.DefaultClient.Get(context.TODO(), client.ObjectKeyFromObject(rs), rs)
}
