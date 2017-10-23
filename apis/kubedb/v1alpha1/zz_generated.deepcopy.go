// +build !ignore_autogenerated

/*
Copyright 2017 The KubeDB Authors.

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

// This file was autogenerated by deepcopy-gen. Do not edit it manually!

package v1alpha1

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	conversion "k8s.io/apimachinery/pkg/conversion"
	runtime "k8s.io/apimachinery/pkg/runtime"
	v1 "k8s.io/client-go/pkg/api/v1"
	reflect "reflect"
)

func init() {
	SchemeBuilder.Register(RegisterDeepCopies)
}

// RegisterDeepCopies adds deep-copy functions to the given scheme. Public
// to allow building arbitrary schemes.
func RegisterDeepCopies(scheme *runtime.Scheme) error {
	return scheme.AddGeneratedDeepCopyFuncs(
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1alpha1_AzureSpec, InType: reflect.TypeOf(&AzureSpec{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1alpha1_BackupScheduleSpec, InType: reflect.TypeOf(&BackupScheduleSpec{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1alpha1_DormantDatabase, InType: reflect.TypeOf(&DormantDatabase{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1alpha1_DormantDatabaseList, InType: reflect.TypeOf(&DormantDatabaseList{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1alpha1_DormantDatabaseSpec, InType: reflect.TypeOf(&DormantDatabaseSpec{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1alpha1_DormantDatabaseStatus, InType: reflect.TypeOf(&DormantDatabaseStatus{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1alpha1_Elasticsearch, InType: reflect.TypeOf(&Elasticsearch{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1alpha1_ElasticsearchList, InType: reflect.TypeOf(&ElasticsearchList{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1alpha1_ElasticsearchNode, InType: reflect.TypeOf(&ElasticsearchNode{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1alpha1_ElasticsearchSpec, InType: reflect.TypeOf(&ElasticsearchSpec{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1alpha1_ElasticsearchStatus, InType: reflect.TypeOf(&ElasticsearchStatus{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1alpha1_ElasticsearchSummary, InType: reflect.TypeOf(&ElasticsearchSummary{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1alpha1_GCSSpec, InType: reflect.TypeOf(&GCSSpec{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1alpha1_InitSpec, InType: reflect.TypeOf(&InitSpec{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1alpha1_LocalSpec, InType: reflect.TypeOf(&LocalSpec{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1alpha1_MonitorSpec, InType: reflect.TypeOf(&MonitorSpec{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1alpha1_Origin, InType: reflect.TypeOf(&Origin{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1alpha1_OriginSpec, InType: reflect.TypeOf(&OriginSpec{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1alpha1_Postgres, InType: reflect.TypeOf(&Postgres{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1alpha1_PostgresList, InType: reflect.TypeOf(&PostgresList{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1alpha1_PostgresSchemaInfo, InType: reflect.TypeOf(&PostgresSchemaInfo{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1alpha1_PostgresSpec, InType: reflect.TypeOf(&PostgresSpec{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1alpha1_PostgresStatus, InType: reflect.TypeOf(&PostgresStatus{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1alpha1_PostgresSummary, InType: reflect.TypeOf(&PostgresSummary{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1alpha1_PostgresTableInfo, InType: reflect.TypeOf(&PostgresTableInfo{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1alpha1_PrometheusSpec, InType: reflect.TypeOf(&PrometheusSpec{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1alpha1_Report, InType: reflect.TypeOf(&Report{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1alpha1_ReportStatus, InType: reflect.TypeOf(&ReportStatus{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1alpha1_ReportSummary, InType: reflect.TypeOf(&ReportSummary{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1alpha1_S3Spec, InType: reflect.TypeOf(&S3Spec{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1alpha1_ScriptSourceSpec, InType: reflect.TypeOf(&ScriptSourceSpec{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1alpha1_Snapshot, InType: reflect.TypeOf(&Snapshot{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1alpha1_SnapshotList, InType: reflect.TypeOf(&SnapshotList{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1alpha1_SnapshotSourceSpec, InType: reflect.TypeOf(&SnapshotSourceSpec{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1alpha1_SnapshotSpec, InType: reflect.TypeOf(&SnapshotSpec{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1alpha1_SnapshotStatus, InType: reflect.TypeOf(&SnapshotStatus{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1alpha1_SnapshotStorageSpec, InType: reflect.TypeOf(&SnapshotStorageSpec{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1alpha1_SwiftSpec, InType: reflect.TypeOf(&SwiftSpec{})},
	)
}

// DeepCopy_v1alpha1_AzureSpec is an autogenerated deepcopy function.
func DeepCopy_v1alpha1_AzureSpec(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*AzureSpec)
		out := out.(*AzureSpec)
		*out = *in
		return nil
	}
}

// DeepCopy_v1alpha1_BackupScheduleSpec is an autogenerated deepcopy function.
func DeepCopy_v1alpha1_BackupScheduleSpec(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*BackupScheduleSpec)
		out := out.(*BackupScheduleSpec)
		*out = *in
		if err := DeepCopy_v1alpha1_SnapshotStorageSpec(&in.SnapshotStorageSpec, &out.SnapshotStorageSpec, c); err != nil {
			return err
		}
		if newVal, err := c.DeepCopy(&in.Resources); err != nil {
			return err
		} else {
			out.Resources = *newVal.(*v1.ResourceRequirements)
		}
		return nil
	}
}

// DeepCopy_v1alpha1_DormantDatabase is an autogenerated deepcopy function.
func DeepCopy_v1alpha1_DormantDatabase(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*DormantDatabase)
		out := out.(*DormantDatabase)
		*out = *in
		if newVal, err := c.DeepCopy(&in.ObjectMeta); err != nil {
			return err
		} else {
			out.ObjectMeta = *newVal.(*meta_v1.ObjectMeta)
		}
		if err := DeepCopy_v1alpha1_DormantDatabaseSpec(&in.Spec, &out.Spec, c); err != nil {
			return err
		}
		if err := DeepCopy_v1alpha1_DormantDatabaseStatus(&in.Status, &out.Status, c); err != nil {
			return err
		}
		return nil
	}
}

// DeepCopy_v1alpha1_DormantDatabaseList is an autogenerated deepcopy function.
func DeepCopy_v1alpha1_DormantDatabaseList(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*DormantDatabaseList)
		out := out.(*DormantDatabaseList)
		*out = *in
		if in.Items != nil {
			in, out := &in.Items, &out.Items
			*out = make([]DormantDatabase, len(*in))
			for i := range *in {
				if err := DeepCopy_v1alpha1_DormantDatabase(&(*in)[i], &(*out)[i], c); err != nil {
					return err
				}
			}
		}
		return nil
	}
}

// DeepCopy_v1alpha1_DormantDatabaseSpec is an autogenerated deepcopy function.
func DeepCopy_v1alpha1_DormantDatabaseSpec(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*DormantDatabaseSpec)
		out := out.(*DormantDatabaseSpec)
		*out = *in
		if err := DeepCopy_v1alpha1_Origin(&in.Origin, &out.Origin, c); err != nil {
			return err
		}
		return nil
	}
}

// DeepCopy_v1alpha1_DormantDatabaseStatus is an autogenerated deepcopy function.
func DeepCopy_v1alpha1_DormantDatabaseStatus(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*DormantDatabaseStatus)
		out := out.(*DormantDatabaseStatus)
		*out = *in
		if in.CreationTime != nil {
			in, out := &in.CreationTime, &out.CreationTime
			*out = new(meta_v1.Time)
			**out = (*in).DeepCopy()
		}
		if in.PausingTime != nil {
			in, out := &in.PausingTime, &out.PausingTime
			*out = new(meta_v1.Time)
			**out = (*in).DeepCopy()
		}
		if in.WipeOutTime != nil {
			in, out := &in.WipeOutTime, &out.WipeOutTime
			*out = new(meta_v1.Time)
			**out = (*in).DeepCopy()
		}
		return nil
	}
}

// DeepCopy_v1alpha1_Elasticsearch is an autogenerated deepcopy function.
func DeepCopy_v1alpha1_Elasticsearch(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*Elasticsearch)
		out := out.(*Elasticsearch)
		*out = *in
		if newVal, err := c.DeepCopy(&in.ObjectMeta); err != nil {
			return err
		} else {
			out.ObjectMeta = *newVal.(*meta_v1.ObjectMeta)
		}
		if err := DeepCopy_v1alpha1_ElasticsearchSpec(&in.Spec, &out.Spec, c); err != nil {
			return err
		}
		if err := DeepCopy_v1alpha1_ElasticsearchStatus(&in.Status, &out.Status, c); err != nil {
			return err
		}
		return nil
	}
}

// DeepCopy_v1alpha1_ElasticsearchList is an autogenerated deepcopy function.
func DeepCopy_v1alpha1_ElasticsearchList(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*ElasticsearchList)
		out := out.(*ElasticsearchList)
		*out = *in
		if in.Items != nil {
			in, out := &in.Items, &out.Items
			*out = make([]Elasticsearch, len(*in))
			for i := range *in {
				if err := DeepCopy_v1alpha1_Elasticsearch(&(*in)[i], &(*out)[i], c); err != nil {
					return err
				}
			}
		}
		return nil
	}
}

// DeepCopy_v1alpha1_ElasticsearchNode is an autogenerated deepcopy function.
func DeepCopy_v1alpha1_ElasticsearchNode(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*ElasticsearchNode)
		out := out.(*ElasticsearchNode)
		*out = *in
		if in.CombinedNodeReplicas != nil {
			in, out := &in.CombinedNodeReplicas, &out.CombinedNodeReplicas
			*out = new(int32)
			**out = **in
		}
		if in.MasterNodeReplicas != nil {
			in, out := &in.MasterNodeReplicas, &out.MasterNodeReplicas
			*out = new(int32)
			**out = **in
		}
		if in.DataNodeReplicas != nil {
			in, out := &in.DataNodeReplicas, &out.DataNodeReplicas
			*out = new(int32)
			**out = **in
		}
		if in.ClientNodeReplicas != nil {
			in, out := &in.ClientNodeReplicas, &out.ClientNodeReplicas
			*out = new(int32)
			**out = **in
		}
		return nil
	}
}

// DeepCopy_v1alpha1_ElasticsearchSpec is an autogenerated deepcopy function.
func DeepCopy_v1alpha1_ElasticsearchSpec(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*ElasticsearchSpec)
		out := out.(*ElasticsearchSpec)
		*out = *in
		if err := DeepCopy_v1alpha1_ElasticsearchNode(&in.Nodes, &out.Nodes, c); err != nil {
			return err
		}
		if in.CertificateSecret != nil {
			in, out := &in.CertificateSecret, &out.CertificateSecret
			if newVal, err := c.DeepCopy(*in); err != nil {
				return err
			} else {
				*out = newVal.(*v1.SecretVolumeSource)
			}
		}
		if in.AuthSecret != nil {
			in, out := &in.AuthSecret, &out.AuthSecret
			if newVal, err := c.DeepCopy(*in); err != nil {
				return err
			} else {
				*out = newVal.(*v1.SecretVolumeSource)
			}
		}
		if in.Storage != nil {
			in, out := &in.Storage, &out.Storage
			if newVal, err := c.DeepCopy(*in); err != nil {
				return err
			} else {
				*out = newVal.(*v1.PersistentVolumeClaimSpec)
			}
		}
		if in.NodeSelector != nil {
			in, out := &in.NodeSelector, &out.NodeSelector
			*out = make(map[string]string)
			for key, val := range *in {
				(*out)[key] = val
			}
		}
		if in.Init != nil {
			in, out := &in.Init, &out.Init
			*out = new(InitSpec)
			if err := DeepCopy_v1alpha1_InitSpec(*in, *out, c); err != nil {
				return err
			}
		}
		if in.BackupSchedule != nil {
			in, out := &in.BackupSchedule, &out.BackupSchedule
			*out = new(BackupScheduleSpec)
			if err := DeepCopy_v1alpha1_BackupScheduleSpec(*in, *out, c); err != nil {
				return err
			}
		}
		if in.Monitor != nil {
			in, out := &in.Monitor, &out.Monitor
			*out = new(MonitorSpec)
			if err := DeepCopy_v1alpha1_MonitorSpec(*in, *out, c); err != nil {
				return err
			}
		}
		if newVal, err := c.DeepCopy(&in.Resources); err != nil {
			return err
		} else {
			out.Resources = *newVal.(*v1.ResourceRequirements)
		}
		if in.Affinity != nil {
			in, out := &in.Affinity, &out.Affinity
			if newVal, err := c.DeepCopy(*in); err != nil {
				return err
			} else {
				*out = newVal.(*v1.Affinity)
			}
		}
		if in.Tolerations != nil {
			in, out := &in.Tolerations, &out.Tolerations
			*out = make([]v1.Toleration, len(*in))
			for i := range *in {
				if newVal, err := c.DeepCopy(&(*in)[i]); err != nil {
					return err
				} else {
					(*out)[i] = *newVal.(*v1.Toleration)
				}
			}
		}
		return nil
	}
}

// DeepCopy_v1alpha1_ElasticsearchStatus is an autogenerated deepcopy function.
func DeepCopy_v1alpha1_ElasticsearchStatus(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*ElasticsearchStatus)
		out := out.(*ElasticsearchStatus)
		*out = *in
		if in.CreationTime != nil {
			in, out := &in.CreationTime, &out.CreationTime
			*out = new(meta_v1.Time)
			**out = (*in).DeepCopy()
		}
		return nil
	}
}

// DeepCopy_v1alpha1_ElasticsearchSummary is an autogenerated deepcopy function.
func DeepCopy_v1alpha1_ElasticsearchSummary(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*ElasticsearchSummary)
		out := out.(*ElasticsearchSummary)
		*out = *in
		if in.IdCount != nil {
			in, out := &in.IdCount, &out.IdCount
			*out = make(map[string]int64)
			for key, val := range *in {
				(*out)[key] = val
			}
		}
		// in.Mapping is kind 'Interface'
		if in.Mapping != nil {
			if newVal, err := c.DeepCopy(&in.Mapping); err != nil {
				return err
			} else {
				out.Mapping = *newVal.(*interface{})
			}
		}
		// in.Setting is kind 'Interface'
		if in.Setting != nil {
			if newVal, err := c.DeepCopy(&in.Setting); err != nil {
				return err
			} else {
				out.Setting = *newVal.(*interface{})
			}
		}
		return nil
	}
}

// DeepCopy_v1alpha1_GCSSpec is an autogenerated deepcopy function.
func DeepCopy_v1alpha1_GCSSpec(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*GCSSpec)
		out := out.(*GCSSpec)
		*out = *in
		return nil
	}
}

// DeepCopy_v1alpha1_InitSpec is an autogenerated deepcopy function.
func DeepCopy_v1alpha1_InitSpec(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*InitSpec)
		out := out.(*InitSpec)
		*out = *in
		if in.ScriptSource != nil {
			in, out := &in.ScriptSource, &out.ScriptSource
			*out = new(ScriptSourceSpec)
			if err := DeepCopy_v1alpha1_ScriptSourceSpec(*in, *out, c); err != nil {
				return err
			}
		}
		if in.SnapshotSource != nil {
			in, out := &in.SnapshotSource, &out.SnapshotSource
			*out = new(SnapshotSourceSpec)
			**out = **in
		}
		return nil
	}
}

// DeepCopy_v1alpha1_LocalSpec is an autogenerated deepcopy function.
func DeepCopy_v1alpha1_LocalSpec(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*LocalSpec)
		out := out.(*LocalSpec)
		*out = *in
		if newVal, err := c.DeepCopy(&in.VolumeSource); err != nil {
			return err
		} else {
			out.VolumeSource = *newVal.(*v1.VolumeSource)
		}
		return nil
	}
}

// DeepCopy_v1alpha1_MonitorSpec is an autogenerated deepcopy function.
func DeepCopy_v1alpha1_MonitorSpec(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*MonitorSpec)
		out := out.(*MonitorSpec)
		*out = *in
		if in.Prometheus != nil {
			in, out := &in.Prometheus, &out.Prometheus
			*out = new(PrometheusSpec)
			if err := DeepCopy_v1alpha1_PrometheusSpec(*in, *out, c); err != nil {
				return err
			}
		}
		return nil
	}
}

// DeepCopy_v1alpha1_Origin is an autogenerated deepcopy function.
func DeepCopy_v1alpha1_Origin(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*Origin)
		out := out.(*Origin)
		*out = *in
		if newVal, err := c.DeepCopy(&in.ObjectMeta); err != nil {
			return err
		} else {
			out.ObjectMeta = *newVal.(*meta_v1.ObjectMeta)
		}
		if err := DeepCopy_v1alpha1_OriginSpec(&in.Spec, &out.Spec, c); err != nil {
			return err
		}
		return nil
	}
}

// DeepCopy_v1alpha1_OriginSpec is an autogenerated deepcopy function.
func DeepCopy_v1alpha1_OriginSpec(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*OriginSpec)
		out := out.(*OriginSpec)
		*out = *in
		if in.Elasticsearch != nil {
			in, out := &in.Elasticsearch, &out.Elasticsearch
			*out = new(ElasticsearchSpec)
			if err := DeepCopy_v1alpha1_ElasticsearchSpec(*in, *out, c); err != nil {
				return err
			}
		}
		if in.Postgres != nil {
			in, out := &in.Postgres, &out.Postgres
			*out = new(PostgresSpec)
			if err := DeepCopy_v1alpha1_PostgresSpec(*in, *out, c); err != nil {
				return err
			}
		}
		return nil
	}
}

// DeepCopy_v1alpha1_Postgres is an autogenerated deepcopy function.
func DeepCopy_v1alpha1_Postgres(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*Postgres)
		out := out.(*Postgres)
		*out = *in
		if newVal, err := c.DeepCopy(&in.ObjectMeta); err != nil {
			return err
		} else {
			out.ObjectMeta = *newVal.(*meta_v1.ObjectMeta)
		}
		if err := DeepCopy_v1alpha1_PostgresSpec(&in.Spec, &out.Spec, c); err != nil {
			return err
		}
		if err := DeepCopy_v1alpha1_PostgresStatus(&in.Status, &out.Status, c); err != nil {
			return err
		}
		return nil
	}
}

// DeepCopy_v1alpha1_PostgresList is an autogenerated deepcopy function.
func DeepCopy_v1alpha1_PostgresList(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*PostgresList)
		out := out.(*PostgresList)
		*out = *in
		if in.Items != nil {
			in, out := &in.Items, &out.Items
			*out = make([]Postgres, len(*in))
			for i := range *in {
				if err := DeepCopy_v1alpha1_Postgres(&(*in)[i], &(*out)[i], c); err != nil {
					return err
				}
			}
		}
		return nil
	}
}

// DeepCopy_v1alpha1_PostgresSchemaInfo is an autogenerated deepcopy function.
func DeepCopy_v1alpha1_PostgresSchemaInfo(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*PostgresSchemaInfo)
		out := out.(*PostgresSchemaInfo)
		*out = *in
		if in.Table != nil {
			in, out := &in.Table, &out.Table
			*out = make(map[string]*PostgresTableInfo)
			for key, val := range *in {
				if newVal, err := c.DeepCopy(&val); err != nil {
					return err
				} else {
					(*out)[key] = *newVal.(**PostgresTableInfo)
				}
			}
		}
		return nil
	}
}

// DeepCopy_v1alpha1_PostgresSpec is an autogenerated deepcopy function.
func DeepCopy_v1alpha1_PostgresSpec(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*PostgresSpec)
		out := out.(*PostgresSpec)
		*out = *in
		if in.Storage != nil {
			in, out := &in.Storage, &out.Storage
			if newVal, err := c.DeepCopy(*in); err != nil {
				return err
			} else {
				*out = newVal.(*v1.PersistentVolumeClaimSpec)
			}
		}
		if in.DatabaseSecret != nil {
			in, out := &in.DatabaseSecret, &out.DatabaseSecret
			if newVal, err := c.DeepCopy(*in); err != nil {
				return err
			} else {
				*out = newVal.(*v1.SecretVolumeSource)
			}
		}
		if in.NodeSelector != nil {
			in, out := &in.NodeSelector, &out.NodeSelector
			*out = make(map[string]string)
			for key, val := range *in {
				(*out)[key] = val
			}
		}
		if in.Init != nil {
			in, out := &in.Init, &out.Init
			*out = new(InitSpec)
			if err := DeepCopy_v1alpha1_InitSpec(*in, *out, c); err != nil {
				return err
			}
		}
		if in.BackupSchedule != nil {
			in, out := &in.BackupSchedule, &out.BackupSchedule
			*out = new(BackupScheduleSpec)
			if err := DeepCopy_v1alpha1_BackupScheduleSpec(*in, *out, c); err != nil {
				return err
			}
		}
		if in.Monitor != nil {
			in, out := &in.Monitor, &out.Monitor
			*out = new(MonitorSpec)
			if err := DeepCopy_v1alpha1_MonitorSpec(*in, *out, c); err != nil {
				return err
			}
		}
		if newVal, err := c.DeepCopy(&in.Resources); err != nil {
			return err
		} else {
			out.Resources = *newVal.(*v1.ResourceRequirements)
		}
		if in.Affinity != nil {
			in, out := &in.Affinity, &out.Affinity
			if newVal, err := c.DeepCopy(*in); err != nil {
				return err
			} else {
				*out = newVal.(*v1.Affinity)
			}
		}
		if in.Tolerations != nil {
			in, out := &in.Tolerations, &out.Tolerations
			*out = make([]v1.Toleration, len(*in))
			for i := range *in {
				if newVal, err := c.DeepCopy(&(*in)[i]); err != nil {
					return err
				} else {
					(*out)[i] = *newVal.(*v1.Toleration)
				}
			}
		}
		return nil
	}
}

// DeepCopy_v1alpha1_PostgresStatus is an autogenerated deepcopy function.
func DeepCopy_v1alpha1_PostgresStatus(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*PostgresStatus)
		out := out.(*PostgresStatus)
		*out = *in
		if in.CreationTime != nil {
			in, out := &in.CreationTime, &out.CreationTime
			*out = new(meta_v1.Time)
			**out = (*in).DeepCopy()
		}
		return nil
	}
}

// DeepCopy_v1alpha1_PostgresSummary is an autogenerated deepcopy function.
func DeepCopy_v1alpha1_PostgresSummary(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*PostgresSummary)
		out := out.(*PostgresSummary)
		*out = *in
		if in.Schema != nil {
			in, out := &in.Schema, &out.Schema
			*out = make(map[string]*PostgresSchemaInfo)
			for key, val := range *in {
				if newVal, err := c.DeepCopy(&val); err != nil {
					return err
				} else {
					(*out)[key] = *newVal.(**PostgresSchemaInfo)
				}
			}
		}
		return nil
	}
}

// DeepCopy_v1alpha1_PostgresTableInfo is an autogenerated deepcopy function.
func DeepCopy_v1alpha1_PostgresTableInfo(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*PostgresTableInfo)
		out := out.(*PostgresTableInfo)
		*out = *in
		return nil
	}
}

// DeepCopy_v1alpha1_PrometheusSpec is an autogenerated deepcopy function.
func DeepCopy_v1alpha1_PrometheusSpec(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*PrometheusSpec)
		out := out.(*PrometheusSpec)
		*out = *in
		if in.Labels != nil {
			in, out := &in.Labels, &out.Labels
			*out = make(map[string]string)
			for key, val := range *in {
				(*out)[key] = val
			}
		}
		return nil
	}
}

// DeepCopy_v1alpha1_Report is an autogenerated deepcopy function.
func DeepCopy_v1alpha1_Report(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*Report)
		out := out.(*Report)
		*out = *in
		if newVal, err := c.DeepCopy(&in.ObjectMeta); err != nil {
			return err
		} else {
			out.ObjectMeta = *newVal.(*meta_v1.ObjectMeta)
		}
		if err := DeepCopy_v1alpha1_ReportSummary(&in.Summary, &out.Summary, c); err != nil {
			return err
		}
		if err := DeepCopy_v1alpha1_ReportStatus(&in.Status, &out.Status, c); err != nil {
			return err
		}
		return nil
	}
}

// DeepCopy_v1alpha1_ReportStatus is an autogenerated deepcopy function.
func DeepCopy_v1alpha1_ReportStatus(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*ReportStatus)
		out := out.(*ReportStatus)
		*out = *in
		if in.StartTime != nil {
			in, out := &in.StartTime, &out.StartTime
			*out = new(meta_v1.Time)
			**out = (*in).DeepCopy()
		}
		if in.CompletionTime != nil {
			in, out := &in.CompletionTime, &out.CompletionTime
			*out = new(meta_v1.Time)
			**out = (*in).DeepCopy()
		}
		return nil
	}
}

// DeepCopy_v1alpha1_ReportSummary is an autogenerated deepcopy function.
func DeepCopy_v1alpha1_ReportSummary(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*ReportSummary)
		out := out.(*ReportSummary)
		*out = *in
		if in.Postgres != nil {
			in, out := &in.Postgres, &out.Postgres
			*out = make(map[string]*PostgresSummary)
			for key, val := range *in {
				if newVal, err := c.DeepCopy(&val); err != nil {
					return err
				} else {
					(*out)[key] = *newVal.(**PostgresSummary)
				}
			}
		}
		if in.Elasticsearch != nil {
			in, out := &in.Elasticsearch, &out.Elasticsearch
			*out = make(map[string]*ElasticsearchSummary)
			for key, val := range *in {
				if newVal, err := c.DeepCopy(&val); err != nil {
					return err
				} else {
					(*out)[key] = *newVal.(**ElasticsearchSummary)
				}
			}
		}
		return nil
	}
}

// DeepCopy_v1alpha1_S3Spec is an autogenerated deepcopy function.
func DeepCopy_v1alpha1_S3Spec(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*S3Spec)
		out := out.(*S3Spec)
		*out = *in
		return nil
	}
}

// DeepCopy_v1alpha1_ScriptSourceSpec is an autogenerated deepcopy function.
func DeepCopy_v1alpha1_ScriptSourceSpec(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*ScriptSourceSpec)
		out := out.(*ScriptSourceSpec)
		*out = *in
		if newVal, err := c.DeepCopy(&in.VolumeSource); err != nil {
			return err
		} else {
			out.VolumeSource = *newVal.(*v1.VolumeSource)
		}
		return nil
	}
}

// DeepCopy_v1alpha1_Snapshot is an autogenerated deepcopy function.
func DeepCopy_v1alpha1_Snapshot(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*Snapshot)
		out := out.(*Snapshot)
		*out = *in
		if newVal, err := c.DeepCopy(&in.ObjectMeta); err != nil {
			return err
		} else {
			out.ObjectMeta = *newVal.(*meta_v1.ObjectMeta)
		}
		if err := DeepCopy_v1alpha1_SnapshotSpec(&in.Spec, &out.Spec, c); err != nil {
			return err
		}
		if err := DeepCopy_v1alpha1_SnapshotStatus(&in.Status, &out.Status, c); err != nil {
			return err
		}
		return nil
	}
}

// DeepCopy_v1alpha1_SnapshotList is an autogenerated deepcopy function.
func DeepCopy_v1alpha1_SnapshotList(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*SnapshotList)
		out := out.(*SnapshotList)
		*out = *in
		if in.Items != nil {
			in, out := &in.Items, &out.Items
			*out = make([]Snapshot, len(*in))
			for i := range *in {
				if err := DeepCopy_v1alpha1_Snapshot(&(*in)[i], &(*out)[i], c); err != nil {
					return err
				}
			}
		}
		return nil
	}
}

// DeepCopy_v1alpha1_SnapshotSourceSpec is an autogenerated deepcopy function.
func DeepCopy_v1alpha1_SnapshotSourceSpec(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*SnapshotSourceSpec)
		out := out.(*SnapshotSourceSpec)
		*out = *in
		return nil
	}
}

// DeepCopy_v1alpha1_SnapshotSpec is an autogenerated deepcopy function.
func DeepCopy_v1alpha1_SnapshotSpec(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*SnapshotSpec)
		out := out.(*SnapshotSpec)
		*out = *in
		if err := DeepCopy_v1alpha1_SnapshotStorageSpec(&in.SnapshotStorageSpec, &out.SnapshotStorageSpec, c); err != nil {
			return err
		}
		if newVal, err := c.DeepCopy(&in.Resources); err != nil {
			return err
		} else {
			out.Resources = *newVal.(*v1.ResourceRequirements)
		}
		return nil
	}
}

// DeepCopy_v1alpha1_SnapshotStatus is an autogenerated deepcopy function.
func DeepCopy_v1alpha1_SnapshotStatus(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*SnapshotStatus)
		out := out.(*SnapshotStatus)
		*out = *in
		if in.StartTime != nil {
			in, out := &in.StartTime, &out.StartTime
			*out = new(meta_v1.Time)
			**out = (*in).DeepCopy()
		}
		if in.CompletionTime != nil {
			in, out := &in.CompletionTime, &out.CompletionTime
			*out = new(meta_v1.Time)
			**out = (*in).DeepCopy()
		}
		return nil
	}
}

// DeepCopy_v1alpha1_SnapshotStorageSpec is an autogenerated deepcopy function.
func DeepCopy_v1alpha1_SnapshotStorageSpec(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*SnapshotStorageSpec)
		out := out.(*SnapshotStorageSpec)
		*out = *in
		if in.Local != nil {
			in, out := &in.Local, &out.Local
			*out = new(LocalSpec)
			if err := DeepCopy_v1alpha1_LocalSpec(*in, *out, c); err != nil {
				return err
			}
		}
		if in.S3 != nil {
			in, out := &in.S3, &out.S3
			*out = new(S3Spec)
			**out = **in
		}
		if in.GCS != nil {
			in, out := &in.GCS, &out.GCS
			*out = new(GCSSpec)
			**out = **in
		}
		if in.Azure != nil {
			in, out := &in.Azure, &out.Azure
			*out = new(AzureSpec)
			**out = **in
		}
		if in.Swift != nil {
			in, out := &in.Swift, &out.Swift
			*out = new(SwiftSpec)
			**out = **in
		}
		return nil
	}
}

// DeepCopy_v1alpha1_SwiftSpec is an autogenerated deepcopy function.
func DeepCopy_v1alpha1_SwiftSpec(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*SwiftSpec)
		out := out.(*SwiftSpec)
		*out = *in
		return nil
	}
}
