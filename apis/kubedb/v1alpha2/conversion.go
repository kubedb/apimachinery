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
	"strings"
	"unsafe"

	v1 "kubedb.dev/apimachinery/apis/kubedb/v1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/conversion"
	ofstv1 "kmodules.xyz/offshoot-api/api/v1"
	ofstconv "kmodules.xyz/offshoot-api/api/v1/conversion"
	ofstv2 "kmodules.xyz/offshoot-api/api/v2"
	rtconv "sigs.k8s.io/controller-runtime/pkg/conversion"
)

// Convert_v1_PodTemplateSpec_To_v2_PodTemplateSpec is an autogenerated conversion function.
func Convert_v1_PodTemplateSpec_To_v2_PodTemplateSpec(in *ofstv1.PodTemplateSpec, out *ofstv2.PodTemplateSpec, s conversion.Scope) error {
	return ofstconv.Convert_v1_PodTemplateSpec_To_v2_PodTemplateSpec(in, out, s)
}

// Convert_v2_PodTemplateSpec_To_v1_PodTemplateSpec is an autogenerated conversion function.
func Convert_v2_PodTemplateSpec_To_v1_PodTemplateSpec(in *ofstv2.PodTemplateSpec, out *ofstv1.PodTemplateSpec, s conversion.Scope) error {
	return ofstconv.Convert_v2_PodTemplateSpec_To_v1_PodTemplateSpec(in, out, s)
}

// Convert_v1_CoordinatorSpec_To_v1alpha2_CoordinatorSpec is an autogenerated conversion function.
func Convert_v1_CoordinatorSpec_To_v1alpha2_CoordinatorSpec(in *[]corev1.Container, out *CoordinatorSpec, s conversion.Scope) error {
	return autoConvert_v1_CoordinatorSpec_To_v1alpha2_CoordinatorSpec(in, out, s)
}

// Convert_v1alpha2_CoordinatorSpec_To_v1_CoordinatorSpec is an autogenerated conversion function.
func Convert_v1alpha2_CoordinatorSpec_To_v1_CoordinatorSpec(in *CoordinatorSpec, out *[]corev1.Container, s conversion.Scope) error {
	return autoConvert_v1alpha2_CoordinatorSpec_To_v1_CoordinatorSpec(in, out, s)
}

// autoConvert_v1_CoordinatorSpec_To_v1alpha2_CoordinatorSpec is an autogenerated conversion function.
func autoConvert_v1_CoordinatorSpec_To_v1alpha2_CoordinatorSpec(in *[]corev1.Container, out *CoordinatorSpec, s conversion.Scope) error {
	var container *corev1.Container
	for i := range *in {
		if strings.HasSuffix((*in)[i].Name, "coordinator") {
			container = &(*in)[i]
		}
	}
	if container == nil || !(container.Resources.Requests != nil || container.Resources.Limits != nil) {
		return nil
	}
	out.Resources = *container.Resources.DeepCopy()
	out.SecurityContext = container.SecurityContext
	return nil
}

// autoConvert_v1alpha2_CoordinatorSpec_To_v1_CoordinatorSpec is an autogenerated conversion function.
func autoConvert_v1alpha2_CoordinatorSpec_To_v1_CoordinatorSpec(in *CoordinatorSpec, out *[]corev1.Container, s conversion.Scope) error {
	var container corev1.Container
	container.SecurityContext = (*corev1.SecurityContext)(unsafe.Pointer(in.SecurityContext))
	container.Resources = in.Resources

	found := false

	for i := range *out {
		if strings.HasSuffix((*out)[i].Name, "coordinator") {
			if container.SecurityContext != nil {
				(*out)[i].SecurityContext = container.SecurityContext
			}
			if container.Resources.Limits != nil || container.Resources.Requests != nil {
				(*out)[i].Resources = *container.Resources.DeepCopy()
			}
			found = true
		}
	}
	if !found {
		*out = append(*out, container)
	}

	return nil
}

func (src *Elasticsearch) ConvertTo(dstRaw rtconv.Hub) error {
	dst := dstRaw.(*v1.Elasticsearch)
	err := Convert_v1alpha2_Elasticsearch_To_v1_Elasticsearch(src, dst, nil)
	if err != nil {
		return err
	}
	if len(dst.Spec.PodTemplate.Spec.Containers) > 0 {
		dst.Spec.PodTemplate.Spec.Containers[0].Name = "elasticsearch" // db container name used in sts
	}
	return nil
}

func (dst *Elasticsearch) ConvertFrom(srcRaw rtconv.Hub) error {
	return Convert_v1_Elasticsearch_To_v1alpha2_Elasticsearch(srcRaw.(*v1.Elasticsearch), dst, nil)
}

func (src *MariaDB) ConvertTo(dstRaw rtconv.Hub) error {
	dst := dstRaw.(*v1.MariaDB)
	err := Convert_v1alpha2_MariaDB_To_v1_MariaDB(src, dst, nil)
	if err != nil {
		return err
	}
	if len(dst.Spec.PodTemplate.Spec.Containers) > 0 {
		dst.Spec.PodTemplate.Spec.Containers[0].Name = "mariadb" // db container name used in sts
	}
	return nil
}

func (dst *MariaDB) ConvertFrom(srcRaw rtconv.Hub) error {
	return Convert_v1_MariaDB_To_v1alpha2_MariaDB(srcRaw.(*v1.MariaDB), dst, nil)
}

func (src *MongoDB) ConvertTo(dstRaw rtconv.Hub) error {
	return Convert_v1alpha2_MongoDB_To_v1_MongoDB(src, dstRaw.(*v1.MongoDB), nil)
}

func (dst *MongoDB) ConvertFrom(srcRaw rtconv.Hub) error {
	return Convert_v1_MongoDB_To_v1alpha2_MongoDB(srcRaw.(*v1.MongoDB), dst, nil)
}

func (src *MySQL) ConvertTo(dstRaw rtconv.Hub) error {
	return Convert_v1alpha2_MySQL_To_v1_MySQL(src, dstRaw.(*v1.MySQL), nil)
}

func (dst *MySQL) ConvertFrom(srcRaw rtconv.Hub) error {
	return Convert_v1_MySQL_To_v1alpha2_MySQL(srcRaw.(*v1.MySQL), dst, nil)
}

func (src *PerconaXtraDB) ConvertTo(dstRaw rtconv.Hub) error {
	return Convert_v1alpha2_PerconaXtraDB_To_v1_PerconaXtraDB(src, dstRaw.(*v1.PerconaXtraDB), nil)
}

func (dst *PerconaXtraDB) ConvertFrom(srcRaw rtconv.Hub) error {
	return Convert_v1_PerconaXtraDB_To_v1alpha2_PerconaXtraDB(srcRaw.(*v1.PerconaXtraDB), dst, nil)
}

func (src *PgBouncer) ConvertTo(dstRaw rtconv.Hub) error {
	return Convert_v1alpha2_PgBouncer_To_v1_PgBouncer(src, dstRaw.(*v1.PgBouncer), nil)
}

func (dst *PgBouncer) ConvertFrom(srcRaw rtconv.Hub) error {
	return Convert_v1_PgBouncer_To_v1alpha2_PgBouncer(srcRaw.(*v1.PgBouncer), dst, nil)
}

func (src *Postgres) ConvertTo(dstRaw rtconv.Hub) error {
	return Convert_v1alpha2_Postgres_To_v1_Postgres(src, dstRaw.(*v1.Postgres), nil)
}

func (dst *Postgres) ConvertFrom(srcRaw rtconv.Hub) error {
	return Convert_v1_Postgres_To_v1alpha2_Postgres(srcRaw.(*v1.Postgres), dst, nil)
}

func (src *ProxySQL) ConvertTo(dstRaw rtconv.Hub) error {
	return Convert_v1alpha2_ProxySQL_To_v1_ProxySQL(src, dstRaw.(*v1.ProxySQL), nil)
}

func (dst *ProxySQL) ConvertFrom(srcRaw rtconv.Hub) error {
	return Convert_v1_ProxySQL_To_v1alpha2_ProxySQL(srcRaw.(*v1.ProxySQL), dst, nil)
}

func (src *RedisSentinel) ConvertTo(dstRaw rtconv.Hub) error {
	return Convert_v1alpha2_RedisSentinel_To_v1_RedisSentinel(src, dstRaw.(*v1.RedisSentinel), nil)
}

func (dst *RedisSentinel) ConvertFrom(srcRaw rtconv.Hub) error {
	return Convert_v1_RedisSentinel_To_v1alpha2_RedisSentinel(srcRaw.(*v1.RedisSentinel), dst, nil)
}

func (src *Redis) ConvertTo(dstRaw rtconv.Hub) error {
	return Convert_v1alpha2_Redis_To_v1_Redis(src, dstRaw.(*v1.Redis), nil)
}

func (dst *Redis) ConvertFrom(srcRaw rtconv.Hub) error {
	return Convert_v1_Redis_To_v1alpha2_Redis(srcRaw.(*v1.Redis), dst, nil)
}
