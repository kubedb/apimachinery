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

package license

import (
	"context"

	catalogapi "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	configapi "kubedb.dev/apimachinery/apis/config/v1alpha1"

	"github.com/Masterminds/semver/v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func MeetsLicenseRestrictions(kc client.Client, restrictions configapi.LicenseRestrictions, dbGK schema.GroupKind, dbVersion string) (bool, error) {
	if len(restrictions) == 0 {
		return true, nil
	}
	restriction, found := restrictions[dbGK.Kind]
	if !found {
		return false, nil
	}

	var dbv unstructured.Unstructured
	dbv.SetGroupVersionKind(catalogapi.SchemeGroupVersion.WithKind(dbGK.Kind + "Version"))
	err := kc.Get(context.TODO(), client.ObjectKey{Name: dbVersion}, &dbv)
	if err != nil {
		return false, err
	}

	strVer, ok, err := unstructured.NestedString(dbv.UnstructuredContent(), "spec", "version")
	if err != nil || !ok {
		return false, err
	}
	v, err := semver.NewVersion(strVer)
	if err != nil {
		return false, err
	}

	c, err := semver.NewConstraint(restriction.VersionConstraint)
	if err != nil {
		return false, err
	}
	if !c.Check(v) {
		// write reason ?
		return false, nil
	}
	if len(restriction.Distributions) > 0 {
		strDistro, ok, err := unstructured.NestedString(dbv.UnstructuredContent(), "spec", "distribution")
		if err != nil || !ok {
			return false, err
		}
		if !contains(restriction.Distributions, strDistro) {
			return false, nil
		}
	}
	return true, nil
}

func contains(list []string, str string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}
