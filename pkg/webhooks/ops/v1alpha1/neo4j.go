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
	"slices"
	"strconv"
	"strings"

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	"gomodules.xyz/x/arrays"
	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/mergepatch"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/klog/v2"
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
	case opsapi.Neo4jOpsRequestTypeUpdateVersion:
		if err := w.validateNeo4jUpdateVersionOpsRequest(neo4j, req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("configuration"), req.Name, err.Error()))
		}
	case opsapi.Neo4jOpsRequestTypeVolumeExpansion:
		if err := w.validateNeo4jVolumeExpansionOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("configuration"), req.Name, err.Error()))
		}
	}

	if len(allErr) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "neo4jopsrequests.kubedb.com", Kind: "Neo4jOpsRequest"}, req.Name, allErr)
}

func (w *Neo4jOpsRequestCustomWebhook) validateNeo4jVolumeExpansionOpsRequest(req *opsapi.Neo4jOpsRequest) error {
	volumeExpansionSpec := req.Spec.VolumeExpansion
	if volumeExpansionSpec == nil {
		return errors.New("spec.volumeExpansion nil not supported in VolumeExpansion type")
	}

	if volumeExpansionSpec.Server == nil {
		return errors.New("spec.volumeExpansion.Server can't be non-empty")
	}

	return nil
}

func (w *Neo4jOpsRequestCustomWebhook) validateNeo4jUpdateVersionOpsRequest(db *dbapi.Neo4j, req *opsapi.Neo4jOpsRequest) error {
	updateVersionSpec := req.Spec.UpdateVersion
	if updateVersionSpec == nil {
		return errors.New("spec.updateVersion nil not supported in UpdateVersion type")
	}

	yes, err := IsCalVerUpgradable(w.DefaultClient, catalog.ResourceKindNeo4jVersion, db.Spec.Version, updateVersionSpec.TargetVersion)
	if err != nil {
		return err
	}
	if !yes {
		return fmt.Errorf("upgrade from version %v to %v is not supported", db.Spec.Version, req.Spec.UpdateVersion.TargetVersion)
	}

	return nil
}

type CalVersionInfo struct {
	Year  int
	Month int
	Day   int
}

func IsCalVerUpgradable(kc client.Client, kind string, curr, target string) (bool, error) {
	var curVersion unstructured.Unstructured
	curVersion.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   catalog.SchemeGroupVersion.Group,
		Version: catalog.SchemeGroupVersion.Version,
		Kind:    kind,
	})

	if err := kc.Get(context.Background(), types.NamespacedName{Name: curr}, &curVersion); err != nil {
		return false, err
	}

	var cat DummyCatalog
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(curVersion.Object, &cat); err != nil {
		return false, fmt.Errorf("failed to unmarshal binding %s: %w", curVersion.GetName(), err)
	}
	klog.Infof("Checking calver upgradability of %s version %s \nIts updateConstraints = %v", kind, curr, cat.Spec.UpdateConstraints)

	var versions unstructured.UnstructuredList
	versions.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   catalog.SchemeGroupVersion.Group,
		Version: catalog.SchemeGroupVersion.Version,
		Kind:    kind,
	})

	if err := kc.List(context.Background(), &versions); err != nil {
		return false, err
	}
	list, err := getUpgradableCalVerVersions(cat.Spec.UpdateConstraints.Allowlist, cat.Spec.UpdateConstraints.Denylist, &versions)
	if err != nil {
		return false, err
	}
	return slices.Contains(list, target), nil
}

func getUpgradableCalVerVersions(allowList, denyList []string, versions *unstructured.UnstructuredList) ([]string, error) {
	allowedVersions := make([]string, 0)

	for _, v := range versions.Items {
		allowed := false
		denied := false

		version, found, err := unstructured.NestedString(v.Object, "spec", "version")
		if err != nil {
			return nil, err
		}
		if !found {
			return nil, errors.New("failed to resolve version constraints, reason: .spec.field is missing")
		}

		vc, err := parseVersion(version)
		if err != nil {
			return nil, err
		}

		for _, ac := range allowList {
			ok, err := matchesCalVerRule(vc, ac)
			if err != nil {
				return nil, err
			}
			if ok {
				allowed = true
				break
			}
		}

		for _, dc := range denyList {
			ok, err := matchesCalVerRule(vc, dc)
			if err != nil {
				return nil, err
			}
			if ok {
				denied = true
				break
			}
		}

		if len(allowList) == 0 {
			allowed = true
		}

		if allowed && !denied {
			allowedVersions = append(allowedVersions, v.GetName())
		}
	}
	return allowedVersions, nil
}

func matchesCalVerRule(version *CalVersionInfo, rule string) (bool, error) {
	rule = strings.TrimSpace(rule)
	if rule == "" {
		return false, errors.New("rule must not be empty")
	}

	parts := strings.Split(rule, ",")
	for _, part := range parts {
		operator, target, err := parseCalVerRulePart(part)
		if err != nil {
			return false, err
		}

		cmp := compareVersion(version, target)
		matched := false
		switch operator {
		case "=":
			matched = cmp == 0
		case "!=":
			matched = cmp != 0
		case ">":
			matched = cmp > 0
		case ">=":
			matched = cmp >= 0
		case "<":
			matched = cmp < 0
		case "<=":
			matched = cmp <= 0
		default:
			return false, fmt.Errorf("unsupported calver operator in rule %q", strings.TrimSpace(part))
		}

		if !matched {
			return false, nil
		}
	}

	return true, nil
}

func parseCalVerRulePart(rule string) (string, *CalVersionInfo, error) {
	rule = strings.TrimSpace(rule)
	if rule == "" {
		return "", nil, errors.New("rule segment must not be empty")
	}

	operator := "="
	value := rule
	for _, op := range []string{">=", "<=", "!=", ">", "<", "="} {
		if strings.HasPrefix(rule, op) {
			operator = op
			value = strings.TrimSpace(strings.TrimPrefix(rule, op))
			break
		}
	}

	target, err := parseVersion(value)
	if err != nil {
		return "", nil, fmt.Errorf("invalid version value %q in rule %q: %w", value, rule, err)
	}

	return operator, target, nil
}

func parseVersion(version string) (*CalVersionInfo, error) {
	version = strings.TrimSpace(version)
	version = strings.TrimPrefix(version, "v")
	if idx := strings.Index(version, "-"); idx != -1 {
		version = version[:idx]
	}

	parts := strings.Split(strings.TrimSpace(version), ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid calver %q, expected format YYYY.MM.DAY", version)
	}

	year, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid calver %q: %w", version, err)
	}
	month, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid calver %q: %w", version, err)
	}
	day, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, fmt.Errorf("invalid calver %q: %w", version, err)
	}

	return &CalVersionInfo{Year: year, Month: month, Day: day}, nil
}

func compareVersion(current, target *CalVersionInfo) int {
	if current.Year != target.Year {
		if current.Year > target.Year {
			return 1
		}
		return -1
	}
	if current.Month != target.Month {
		if current.Month > target.Month {
			return 1
		}
		return -1
	}
	if current.Day != target.Day {
		if current.Day > target.Day {
			return 1
		}
		return -1
	}
	return 0
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

	if err := validateNeo4jHorizontalScaling(current, target, req); err != nil {
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

func validateNeo4jHorizontalScaling(current, target int32, req *opsapi.Neo4jOpsRequest) error {
	if current == target && (req.Status.Phase == opsapi.OpsRequestPhasePending || req.Status.Phase == "") {
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
