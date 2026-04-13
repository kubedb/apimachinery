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
	"time"

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	"kubedb.dev/apimachinery/apis/kubedb"
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	"github.com/Masterminds/semver/v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	cu "kmodules.xyz/client-go/client"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DummyCatalog struct {
	Spec DummyCatalogSpec `json:"spec" yaml:"spec,omitempty"`
}

type DummyCatalogSpec struct {
	UpdateConstraints catalog.UpdateConstraints `json:"updateConstraints"`
}

type VersionInfo struct {
	Major int
	Minor int
	Patch int
}

func IsUpgradable(kc client.Client, kind string, curr, target string) (bool, error) {
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
	klog.Infof("Checking upgradability of %s version %s \nIts updateConstraints = %v", kind, curr, cat.Spec.UpdateConstraints)

	var versions unstructured.UnstructuredList
	versions.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   catalog.SchemeGroupVersion.Group,
		Version: catalog.SchemeGroupVersion.Version,
		Kind:    kind,
	})

	if err := kc.List(context.Background(), &versions); err != nil {
		return false, err
	}
	list, err := getUpgradableVersions(cat.Spec.UpdateConstraints.Allowlist, cat.Spec.UpdateConstraints.Denylist, &versions)
	if err != nil {
		return false, err
	}
	return slices.Contains(list, target), nil
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

func getUpgradableVersions(allowList, denyList []string, versions *unstructured.UnstructuredList) ([]string, error) {
	allowConstraints := make([]*semver.Constraints, 0, len(allowList))
	denyConstraints := make([]*semver.Constraints, 0, len(denyList))

	for _, ac := range allowList {
		c, err := semver.NewConstraint(ac)
		if err != nil {
			return nil, err
		}
		allowConstraints = append(allowConstraints, c)
	}

	for _, dc := range denyList {
		c, err := semver.NewConstraint(dc)
		if err != nil {
			return nil, err
		}
		denyConstraints = append(denyConstraints, c)
	}

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

		vc, err := semver.NewVersion(version)
		if err != nil {
			return nil, err
		}

		for _, ac := range allowConstraints {
			if ac.Check(vc) {
				allowed = true
				break
			}
		}

		for _, dc := range denyConstraints {
			if dc.Check(vc) {
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

		vc, err := parseValidatedVersion(version)
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

func matchesCalVerRule(version *VersionInfo, rule string) (bool, error) {
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

func parseCalVerRulePart(rule string) (string, *VersionInfo, error) {
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

	target, err := parseValidatedVersion(value)
	if err != nil {
		return "", nil, fmt.Errorf("invalid version value %q in rule %q: %w", value, rule, err)
	}

	return operator, target, nil
}

func parseValidatedVersion(value string) (*VersionInfo, error) {
	v, err := parseVersion(value)
	if err != nil {
		return nil, err
	}

	currentYear := time.Now().Year()
	if v.Major < 2025 || v.Major > currentYear+1 {
		return nil, fmt.Errorf("year %d is out of reasonable range [2025, %d]", v.Major, currentYear+1)
	}
	if v.Minor < 1 || v.Minor > 12 {
		return nil, fmt.Errorf("month %d is invalid, must be between 1 and 12", v.Minor)
	}
	if v.Patch < 0 {
		return nil, fmt.Errorf("patch %d must be non-negative", v.Patch)
	}

	return v, nil
}

func parseVersion(version string) (*VersionInfo, error) {
	version = strings.TrimSpace(version)
	version = strings.TrimPrefix(version, "v")
	if idx := strings.Index(version, "-"); idx != -1 {
		version = version[:idx]
	}

	parts := strings.Split(strings.TrimSpace(version), ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid calver %q, expected format YYYY.MM.PATCH", version)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid calver %q: %w", version, err)
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid calver %q: %w", version, err)
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, fmt.Errorf("invalid calver %q: %w", version, err)
	}

	return &VersionInfo{Major: major, Minor: minor, Patch: patch}, nil
}

func compareVersion(current, target *VersionInfo) int {
	if current.Major != target.Major {
		if current.Major > target.Major {
			return 1
		}
		return -1
	}
	if current.Minor != target.Minor {
		if current.Minor > target.Minor {
			return 1
		}
		return -1
	}
	if current.Patch != target.Patch {
		if current.Patch > target.Patch {
			return 1
		}
		return -1
	}
	return 0
}

func isOpsReqCompleted(phase opsapi.OpsRequestPhase) bool {
	return phase == opsapi.OpsRequestPhaseSuccessful || phase == opsapi.OpsRequestPhaseFailed
}

func resumeDatabase(kc client.Client, db dbapi.Accessor) error {
	_, err := cu.PatchStatus(context.TODO(), kc, db, func(obj client.Object) client.Object {
		ret := obj.(dbapi.Accessor)
		ret.RemoveCondition(kubedb.DatabasePaused)
		return ret
	})
	return err
}
