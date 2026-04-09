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
	meta_util "kmodules.xyz/client-go/meta"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// SetupNeo4jOpsRequestWebhookWithManager registers the webhook for Neo4jOpsRequest in the manager.
func SetupNeo4jOpsRequestWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&opsapi.Neo4jOpsRequest{}).
		WithValidator(&Neo4jOpsRequestCustomWebhook{mgr.GetClient()}).
		Complete()
}

type Neo4jOpsRequestCustomWebhook struct {
	DefaultClient client.Client
}

var neo4jLog = logf.Log.WithName("neo4j-opsrequest")

var _ webhook.CustomValidator = &Neo4jOpsRequestCustomWebhook{}

func (w *Neo4jOpsRequestCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	req, ok := obj.(*opsapi.Neo4jOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an Neo4jOpsRequest object but got %T", obj)
	}
	neo4jLog.Info("validate create", "name", req.Name)
	return nil, w.validateCreateOrUpdate(req)
}

func (w *Neo4jOpsRequestCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	req, ok := newObj.(*opsapi.Neo4jOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an Neo4jOpsRequest object but got %T", newObj)
	}
	neo4jLog.Info("validate update", "name", req.Name)

	oldReq, ok := oldObj.(*opsapi.Neo4jOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an Neo4jOpsRequest object but got %T", oldObj)
	}

	if err := validateNeo4jOpsRequest(req, oldReq); err != nil {
		return nil, err
	}
	return nil, w.validateCreateOrUpdate(req)
}

func (w *Neo4jOpsRequestCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	_, ok := obj.(*opsapi.Neo4jOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an Neo4jOpsRequest object but got %T", obj)
	}
	return nil, nil
}

func validateNeo4jOpsRequest(req *opsapi.Neo4jOpsRequest, oldReq *opsapi.Neo4jOpsRequest) error {
	preconditions := meta_util.PreConditionSet{Set: sets.New[string]("spec")}
	_, err := meta_util.CreateStrategicPatch(oldReq, req, preconditions.PreconditionFunc()...)
	if err != nil {
		if mergepatch.IsPreconditionFailed(err) {
			return fmt.Errorf("%v.%v", err, preconditions.Error())
		}
		return err
	}
	return nil
}

func (w *Neo4jOpsRequestCustomWebhook) validateCreateOrUpdate(req *opsapi.Neo4jOpsRequest) error {
	if validType, _ := arrays.Contains(opsapi.Neo4jOpsRequestTypeNames(), string(req.Spec.Type)); !validType {
		return field.Invalid(field.NewPath("spec").Child("type"), req.Name,
			fmt.Sprintf("defined OpsRequestType %s is not supported, supported types for Neo4j are %s", req.Spec.Type, strings.Join(opsapi.Neo4jOpsRequestTypeNames(), ", ")))
	}

	neo4j, err := w.hasDatabaseRef(req)
	if err != nil {
		return err
	}

	var allErr field.ErrorList
	switch opsapi.Neo4jOpsRequestType(req.GetRequestType()) {
	case opsapi.Neo4jOpsRequestTypeHorizontalScaling:
		if err := w.validateNeo4jHorizontalScalingOpsRequest(neo4j, req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("horizontalScaling"), req.Name, err.Error()))
		}
	case opsapi.Neo4jOpsRequestTypeReconfigureTLS:
		if err := w.validateNeo4jReconfigurationTLSOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("tls"), req.Name, err.Error()))
		}
	case opsapi.Neo4jOpsRequestTypeRotateAuth:
		if err := w.validateNeo4jRotateAuthenticationOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("authentication"), req.Name, err.Error()))
		}
	case opsapi.Neo4jOpsRequestTypeReconfigure:
		if err := w.validateNeo4jReconfigurationOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("configuration"), req.Name, err.Error()))
		}
	case opsapi.Neo4jOpsRequestTypeVerticalScaling:
		if err := w.validateNeo4jVerticalScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("configuration"), req.Name, err.Error()))
		}
	}

	if len(allErr) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "neo4jopsrequests.kubedb.com", Kind: "Neo4jOpsRequest"}, req.Name, allErr)
}

func (w *Neo4jOpsRequestCustomWebhook) validateNeo4jVerticalScalingOpsRequest(req *opsapi.Neo4jOpsRequest) error {
	verticalScalingSpec := req.Spec.VerticalScaling
	if verticalScalingSpec == nil {
		return errors.New("spec.verticalScaling nil not supported in VerticalScaling type")
	}
	if verticalScalingSpec.Server == nil {
		return errors.New("`spec.verticalScaling.Server`,should be present in vertical scaling ops request")
	}

	return nil
}

func (w *Neo4jOpsRequestCustomWebhook) hasDatabaseRef(req *opsapi.Neo4jOpsRequest) (*dbapi.Neo4j, error) {
	neo4j := dbapi.Neo4j{}
	if err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      req.GetDBRefName(),
		Namespace: req.GetNamespace(),
	}, &neo4j); err != nil {
		allErr := field.ErrorList{field.Invalid(field.NewPath("spec").Child("databaseRef"), req.GetDBRefName(), fmt.Sprintf("%s/%s is invalid or not found", req.GetNamespace(), req.GetDBRefName()))}
		return nil, apierrors.NewInvalid(schema.GroupKind{Group: "neo4jopsrequests.kubedb.com", Kind: "Neo4jOpsRequest"}, req.Name, allErr)
	}
	return &neo4j, nil
}

func (w *Neo4jOpsRequestCustomWebhook) validateNeo4jReconfigurationOpsRequest(req *opsapi.Neo4jOpsRequest) error {
	configurationSpec := req.Spec.Configuration
	if configurationSpec == nil {
		return errors.New("spec.configuration nil not supported in Reconfigure type")
	}

	if !configurationSpec.RemoveCustomConfig && configurationSpec.ConfigSecret == nil && len(configurationSpec.ApplyConfig) == 0 {
		return errors.New("at least one of `RemoveCustomConfig`, `ConfigSecret`, or `ApplyConfig` must be specified")
	}
	return nil
}

func (w *Neo4jOpsRequestCustomWebhook) validateNeo4jHorizontalScalingOpsRequest(neo4j *dbapi.Neo4j, req *opsapi.Neo4jOpsRequest) error {
	horizontalScalingSpec := req.Spec.HorizontalScaling
	if horizontalScalingSpec == nil {
		return errors.New("spec.horizontalScaling nil not supported in HorizontalScaling type")
	}
	if horizontalScalingSpec.Server == nil {
		return errors.New("spec.horizontalScaling.server can not be empty")
	}
	if *horizontalScalingSpec.Server <= 0 {
		return errors.New("spec.horizontalScaling.server must be positive")
	}

	current := *neo4j.Spec.Replicas
	target := *horizontalScalingSpec.Server

	if err := validateNeo4jHorizontalScaling(current, target); err != nil {
		return err
	}

	// validate reallocate strategy against scaling direction
	if horizontalScalingSpec.Reallocate != nil && horizontalScalingSpec.Reallocate.Strategy == opsapi.StrategyNone {
		if target < current {
			return errors.New("reallocate strategy 'none' is not allowed when downscaling: " +
				"removing nodes without reallocation will cause data loss. " +
				"Use 'incremental' or 'full' strategy instead")
		}
	}

	return nil
}

func validateNeo4jHorizontalScaling(current, target int32) error {
	if current == target {
		return fmt.Errorf("target replicas %d is same as current replicas %d", target, current)
	}

	if target > current {
		if current == 1 {
			return fmt.Errorf("cannot scale up from standalone (1 node) to cluster directly. migration from standalone to cluster requires a full data migration process")
		}
		return nil
	}

	if target < current {
		if target == 2 {
			return fmt.Errorf("cannot scale Neo4j cluster to 2 nodes. Neo4j requires minimum 3 voting members for system database. scale to 3 (minimum cluster)")
		}
		if target == 1 && current >= 3 {
			return fmt.Errorf("cannot scale down from cluster (%d nodes) to standalone (1 node) directly. migration from cluster to standalone requires a full data migration process", current)
		}
	}

	return nil
}

func (w *Neo4jOpsRequestCustomWebhook) validateNeo4jReconfigurationTLSOpsRequest(req *opsapi.Neo4jOpsRequest) error {
	tlsSpec := req.Spec.TLS
	if tlsSpec == nil {
		return errors.New("spec.TLS nil not supported in ReconfigureTLS type")
	}

	configCount := 0
	if tlsSpec.Remove {
		configCount++
	}
	if tlsSpec.RotateCertificates {
		configCount++
	}
	if tlsSpec.IssuerRef != nil || tlsSpec.Certificates != nil {
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

func (w *Neo4jOpsRequestCustomWebhook) validateNeo4jRotateAuthenticationOpsRequest(req *opsapi.Neo4jOpsRequest) error {
	authSpec := req.Spec.Authentication
	var newAuthsecret core.Secret
	if authSpec != nil && authSpec.SecretRef != nil {
		if authSpec.SecretRef.Name == "" {
			return errors.New("spec.authentication.secretRef.name can not be empty")
		}
		err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
			Name:      authSpec.SecretRef.Name,
			Namespace: req.Namespace,
		}, &newAuthsecret)
		if err != nil {
			if apierrors.IsNotFound(err) {
				return fmt.Errorf("referenced secret %s not found", authSpec.SecretRef.Name)
			}
			return err
		}

		if newAuthsecret.Data == nil {
			return errors.New("spec.authentication.secretRef.name is a valid secret but it does not contain any data")
		}
		if newAuthsecret.Data[core.BasicAuthUsernameKey] == nil || newAuthsecret.Data[core.BasicAuthPasswordKey] == nil {
			return errors.New("spec.authentication.secretRef.name is a valid secret but it does not contain username or password")
		}
	}

	return nil
}
