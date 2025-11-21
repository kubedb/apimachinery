package v1alpha1

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"k8s.io/klog/v2"
	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"

	"github.com/Masterminds/semver/v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DummyCatalog struct {
	Spec DummyCatalogSpec `json:"spec,omitempty" yaml:"spec,omitempty"`
}

type DummyCatalogSpec struct {
	UpdateConstraints catalog.UpdateConstraints `json:"updateConstraints,omitempty"`
}

func IsUpgradable(kc client.Client, kind string, versionName string) (bool, error) {
	klog.Infof("Checking upgradability of %s/%s version %s", kind, versionName, versionName)
	var curVersion unstructured.Unstructured
	curVersion.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   catalog.SchemeGroupVersion.Group,
		Version: catalog.SchemeGroupVersion.Version,
		Kind:    kind,
	})

	if err := kc.Get(context.Background(), types.NamespacedName{Name: versionName}, &curVersion); err != nil {
		return false, err
	}
	klog.Infof("1111111111111")

	var cat DummyCatalog
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(curVersion.Object, &cat); err != nil {
		return false, fmt.Errorf("failed to unmarshal binding %s: %w", curVersion.GetName(), err)
	}

	klog.Infof("22222222222")

	var versions unstructured.UnstructuredList
	versions.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   catalog.SchemeGroupVersion.Group,
		Version: catalog.SchemeGroupVersion.Version,
		Kind:    kind,
	})

	if err := kc.List(context.Background(), &versions); err != nil {
		return false, err
	}

	klog.Infof("3333333333333")
	list, err := getUpgradableVersions(cat.Spec.UpdateConstraints.Allowlist, cat.Spec.UpdateConstraints.Denylist, &versions)
	if err != nil {
		return false, err
	}
	klog.Infof("444444444444444")
	return slices.Contains(list, versionName), nil
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

	allowedVersion := make([]string, 0)

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
			allowedVersion = append(allowedVersion, v.GetName())
		}
	}
	return allowedVersion, nil
}
