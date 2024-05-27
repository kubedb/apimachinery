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

package v1alpha2

import (
	"context"
	"errors"
	"fmt"

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"

	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ofst "kmodules.xyz/offshoot-api/api/v2"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var clickhouselog = logf.Log.WithName("clickhouse-resource")

var _ webhook.Defaulter = &ClickHouse{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *ClickHouse) Default() {
	if r == nil {
		return
	}
	clickhouselog.Info("default", "name", r.Name)
	r.SetDefaults()
}

var _ webhook.Validator = &ClickHouse{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *ClickHouse) ValidateCreate() (admission.Warnings, error) {
	clickhouselog.Info("validate create", "name", r.Name)
	return nil, r.ValidateCreateOrUpdate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *ClickHouse) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	clickhouselog.Info("validate update", "name", r.Name)
	return nil, r.ValidateCreateOrUpdate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *ClickHouse) ValidateDelete() (admission.Warnings, error) {
	clickhouselog.Info("validate delete", "name", r.Name)

	var allErr field.ErrorList
	if r.Spec.TerminationPolicy == TerminationPolicyDoNotTerminate {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("teminationPolicy"),
			r.Name,
			"Can not delete as terminationPolicy is set to \"DoNotTerminate\""))
		return nil, apierrors.NewInvalid(schema.GroupKind{Group: "ClickHouse.kubedb.com", Kind: "ClickHouse"}, r.Name, allErr)
	}
	return nil, nil
}

func (r *ClickHouse) ValidateCreateOrUpdate() error {
	var allErr field.ErrorList
	if r.Spec.EnableSSL {
		if r.Spec.TLS == nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("enableSSL"),
				r.Name,
				".spec.tls can't be nil, if .spec.enableSSL is true"))
		}
	} else {
		if r.Spec.TLS != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("enableSSL"),
				r.Name,
				".spec.tls must be nil, if .spec.enableSSL is disabled"))
		}
	}

	if r.Spec.ClusterTopology != nil {
		clusterName := map[string]bool{}
		clusters := r.Spec.ClusterTopology.Cluster
		for _, cluster := range clusters {
			if cluster.Shards != nil && *cluster.Shards <= 0 {
				allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("clusterTopology").Child("shards"),
					r.Name,
					"number of shards can not be 0 or less"))
			}
			if cluster.Replicas != nil && *cluster.Replicas <= 0 {
				allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("clusterTopology").Child("replicas"),
					r.Name,
					"number of replicas can't be 0 or less"))
			}
			if clusterName[cluster.Name] == true {
				allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("clusterTopology").Child(cluster.Name),
					r.Name,
					"cluster name is duplicated, use different cluster name"))
			}
			clusterName[cluster.Name] = true
		}
	} else {
		// number of replicas can not be 0 or less
		if r.Spec.Replicas != nil && *r.Spec.Replicas <= 0 {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("replicas"),
				r.Name,
				"number of replicas can't be 0 or less"))
		}
	}

	if r.Spec.Version == "" {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("version"),
			r.Name,
			"spec.version' is missing"))
	} else {
		err := r.ValidateVersion(r)
		if err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("version"),
				r.Spec.Version,
				err.Error()))
		}
	}

	err := r.validateVolumes(r)
	if err != nil {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("podTemplate").Child("spec").Child("volumes"),
			r.Name,
			err.Error()))
	}

	err = r.validateVolumesMountPaths(&r.Spec.PodTemplate)
	if err != nil {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("podTemplate").Child("spec").Child("volumeMounts"),
			r.Name,
			err.Error()))
	}

	if r.Spec.ClusterTopology != nil {
		clusters := r.Spec.ClusterTopology.Cluster
		for _, cluster := range clusters {
			allErr = r.validateClusterStorageType(cluster.StorageType, r.Spec.Storage, cluster.Name, allErr)
		}
	} else {
		allErr = r.validateStandaloneStorageType(r.Spec.StorageType, r.Spec.Storage, allErr)
	}

	//if r.Spec.ConfigSecret != nil && r.Spec.ConfigSecret.Name == "" {
	//	allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("configSecret").Child("name"),
	//		r.Name,
	//		"ConfigSecret Name can not be empty"))
	//}

	if len(allErr) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "ClickHouse.kubedb.com", Kind: "ClickHouse"}, r.Name, allErr)
}

func (c *ClickHouse) validateStandaloneStorageType(storageType StorageType, storage *core.PersistentVolumeClaimSpec, allErr field.ErrorList) field.ErrorList {
	if storageType == "" {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("storageType"),
			c.Name,
			"StorageType can not be empty"))
	} else {
		if storageType != StorageTypeDurable && c.Spec.StorageType != StorageTypeEphemeral {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("storageType"),
				c.Name,
				"StorageType should be either durable or ephemeral"))
		}
	}

	if storage == nil && c.Spec.StorageType == StorageTypeDurable {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("storage"),
			c.Name,
			"Storage can't be empty when StorageType is durable"))
	}

	return allErr
}

func (c *ClickHouse) validateClusterStorageType(storageType StorageType, storage *core.PersistentVolumeClaimSpec, cluster string, allErr field.ErrorList) field.ErrorList {
	if storageType == "" {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("clusterTopology").Child(cluster).Child("storageType"),
			c.Name,
			"StorageType can not be empty"))
	} else {
		if storageType != StorageTypeDurable && storageType != StorageTypeEphemeral {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("clusterTopology").Child(cluster).Child("storageType"),
				storageType,
				"StorageType should be either durable or ephemeral"))
		}
	}
	if storage == nil && storageType == StorageTypeDurable {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("clusterTopology").Child(cluster).Child("storage"),
			c.Name,
			"Storage can't be empty when StorageType is durable"))
	}
	return allErr
}

func (r *ClickHouse) ValidateVersion(db *ClickHouse) error {
	chVersion := catalog.ClickHouseVersion{}
	err := DefaultClient.Get(context.TODO(), types.NamespacedName{Name: db.Spec.Version}, &chVersion)
	if err != nil {
		// fmt.Sprint(db.Spec.Version, "version not supported")
		return errors.New(fmt.Sprint("version ", db.Spec.Version, " not supported"))
	}
	return nil
}

var clickhouseReservedVolumes = []string{
	ClickHouseVolumeData,
}

func (r *ClickHouse) validateVolumes(db *ClickHouse) error {
	if db.Spec.PodTemplate.Spec.Volumes == nil {
		return nil
	}
	rsv := make([]string, len(clickhouseReservedVolumes))
	copy(rsv, clickhouseReservedVolumes)
	volumes := db.Spec.PodTemplate.Spec.Volumes
	for _, rv := range rsv {
		for _, ugv := range volumes {
			if ugv.Name == rv {
				return errors.New("Cannot use a reserve volume name: " + rv)
			}
		}
	}
	return nil
}

var clickhouseReservedVolumeMountPaths = []string{
	ClickHouseDataDir,
}

func (r *ClickHouse) validateVolumesMountPaths(podTemplate *ofst.PodTemplateSpec) error {
	if podTemplate == nil {
		return nil
	}
	if podTemplate.Spec.Containers == nil {
		return nil
	}

	for _, rvmp := range clickhouseReservedVolumeMountPaths {
		containerList := podTemplate.Spec.Containers
		for i := range containerList {
			mountPathList := containerList[i].VolumeMounts
			for j := range mountPathList {
				if mountPathList[j].MountPath == rvmp {
					return errors.New("Can't use a reserve volume mount path name: " + rvmp)
				}
			}
		}
	}

	if podTemplate.Spec.InitContainers == nil {
		return nil
	}

	for _, rvmp := range clickhouseReservedVolumeMountPaths {
		containerList := podTemplate.Spec.InitContainers
		for i := range containerList {
			mountPathList := containerList[i].VolumeMounts
			for j := range mountPathList {
				if mountPathList[j].MountPath == rvmp {
					return errors.New("Can't use a reserve volume mount path name: " + rvmp)
				}
			}
		}
	}

	return nil
}
