//go:build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	v1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	apiv1 "kmodules.xyz/client-go/api/v1"
	offshoot_apiapiv1 "kmodules.xyz/offshoot-api/api/v1"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Database) DeepCopyInto(out *Database) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Database.
func (in *Database) DeepCopy() *Database {
	if in == nil {
		return nil
	}
	out := new(Database)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *InitSpec) DeepCopyInto(out *InitSpec) {
	*out = *in
	in.Script.DeepCopyInto(&out.Script)
	if in.PodTemplate != nil {
		in, out := &in.PodTemplate, &out.PodTemplate
		*out = new(offshoot_apiapiv1.PodTemplateSpec)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new InitSpec.
func (in *InitSpec) DeepCopy() *InitSpec {
	if in == nil {
		return nil
	}
	out := new(InitSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MariaDBDatabase) DeepCopyInto(out *MariaDBDatabase) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MariaDBDatabase.
func (in *MariaDBDatabase) DeepCopy() *MariaDBDatabase {
	if in == nil {
		return nil
	}
	out := new(MariaDBDatabase)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MariaDBDatabase) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MariaDBDatabaseList) DeepCopyInto(out *MariaDBDatabaseList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]MariaDBDatabase, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MariaDBDatabaseList.
func (in *MariaDBDatabaseList) DeepCopy() *MariaDBDatabaseList {
	if in == nil {
		return nil
	}
	out := new(MariaDBDatabaseList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MariaDBDatabaseList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MariaDBDatabaseSpec) DeepCopyInto(out *MariaDBDatabaseSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MariaDBDatabaseSpec.
func (in *MariaDBDatabaseSpec) DeepCopy() *MariaDBDatabaseSpec {
	if in == nil {
		return nil
	}
	out := new(MariaDBDatabaseSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MariaDBDatabaseStatus) DeepCopyInto(out *MariaDBDatabaseStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MariaDBDatabaseStatus.
func (in *MariaDBDatabaseStatus) DeepCopy() *MariaDBDatabaseStatus {
	if in == nil {
		return nil
	}
	out := new(MariaDBDatabaseStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MongoDBDatabase) DeepCopyInto(out *MongoDBDatabase) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MongoDBDatabase.
func (in *MongoDBDatabase) DeepCopy() *MongoDBDatabase {
	if in == nil {
		return nil
	}
	out := new(MongoDBDatabase)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MongoDBDatabase) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MongoDBDatabaseList) DeepCopyInto(out *MongoDBDatabaseList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]MongoDBDatabase, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MongoDBDatabaseList.
func (in *MongoDBDatabaseList) DeepCopy() *MongoDBDatabaseList {
	if in == nil {
		return nil
	}
	out := new(MongoDBDatabaseList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MongoDBDatabaseList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MongoDBDatabaseSpec) DeepCopyInto(out *MongoDBDatabaseSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MongoDBDatabaseSpec.
func (in *MongoDBDatabaseSpec) DeepCopy() *MongoDBDatabaseSpec {
	if in == nil {
		return nil
	}
	out := new(MongoDBDatabaseSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MongoDBDatabaseStatus) DeepCopyInto(out *MongoDBDatabaseStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MongoDBDatabaseStatus.
func (in *MongoDBDatabaseStatus) DeepCopy() *MongoDBDatabaseStatus {
	if in == nil {
		return nil
	}
	out := new(MongoDBDatabaseStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (obj *MySQLDatabase) DeepCopyInto(out *MySQLDatabase) {
	*out = *obj
	out.TypeMeta = obj.TypeMeta
	obj.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	obj.Spec.DeepCopyInto(&out.Spec)
	obj.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MySQLDatabase.
func (obj *MySQLDatabase) DeepCopy() *MySQLDatabase {
	if obj == nil {
		return nil
	}
	out := new(MySQLDatabase)
	obj.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (obj *MySQLDatabase) DeepCopyObject() runtime.Object {
	if c := obj.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MySQLDatabaseList) DeepCopyInto(out *MySQLDatabaseList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]MySQLDatabase, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MySQLDatabaseList.
func (in *MySQLDatabaseList) DeepCopy() *MySQLDatabaseList {
	if in == nil {
		return nil
	}
	out := new(MySQLDatabaseList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MySQLDatabaseList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MySQLDatabaseSpec) DeepCopyInto(out *MySQLDatabaseSpec) {
	*out = *in
	in.DatabaseRef.DeepCopyInto(&out.DatabaseRef)
	out.VaultRef = in.VaultRef
	out.DatabaseConfig = in.DatabaseConfig
	if in.Subjects != nil {
		in, out := &in.Subjects, &out.Subjects
		*out = make([]v1.Subject, len(*in))
		copy(*out, *in)
	}
	if in.Init != nil {
		in, out := &in.Init, &out.Init
		*out = new(InitSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.Restore != nil {
		in, out := &in.Restore, &out.Restore
		*out = new(RestoreConf)
		**out = **in
	}
	if in.ValidationTimeLimit != nil {
		in, out := &in.ValidationTimeLimit, &out.ValidationTimeLimit
		*out = new(TTL)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MySQLDatabaseSpec.
func (in *MySQLDatabaseSpec) DeepCopy() *MySQLDatabaseSpec {
	if in == nil {
		return nil
	}
	out := new(MySQLDatabaseSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MySQLDatabaseStatus) DeepCopyInto(out *MySQLDatabaseStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]apiv1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.LoginCreds != nil {
		in, out := &in.LoginCreds, &out.LoginCreds
		*out = new(apiv1.ObjectReference)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MySQLDatabaseStatus.
func (in *MySQLDatabaseStatus) DeepCopy() *MySQLDatabaseStatus {
	if in == nil {
		return nil
	}
	out := new(MySQLDatabaseStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PostgresDatabase) DeepCopyInto(out *PostgresDatabase) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PostgresDatabase.
func (in *PostgresDatabase) DeepCopy() *PostgresDatabase {
	if in == nil {
		return nil
	}
	out := new(PostgresDatabase)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PostgresDatabase) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PostgresDatabaseList) DeepCopyInto(out *PostgresDatabaseList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]PostgresDatabase, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PostgresDatabaseList.
func (in *PostgresDatabaseList) DeepCopy() *PostgresDatabaseList {
	if in == nil {
		return nil
	}
	out := new(PostgresDatabaseList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PostgresDatabaseList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PostgresDatabaseSpec) DeepCopyInto(out *PostgresDatabaseSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PostgresDatabaseSpec.
func (in *PostgresDatabaseSpec) DeepCopy() *PostgresDatabaseSpec {
	if in == nil {
		return nil
	}
	out := new(PostgresDatabaseSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PostgresDatabaseStatus) DeepCopyInto(out *PostgresDatabaseStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PostgresDatabaseStatus.
func (in *PostgresDatabaseStatus) DeepCopy() *PostgresDatabaseStatus {
	if in == nil {
		return nil
	}
	out := new(PostgresDatabaseStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RedisDatabase) DeepCopyInto(out *RedisDatabase) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RedisDatabase.
func (in *RedisDatabase) DeepCopy() *RedisDatabase {
	if in == nil {
		return nil
	}
	out := new(RedisDatabase)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RedisDatabase) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RedisDatabaseList) DeepCopyInto(out *RedisDatabaseList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]RedisDatabase, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RedisDatabaseList.
func (in *RedisDatabaseList) DeepCopy() *RedisDatabaseList {
	if in == nil {
		return nil
	}
	out := new(RedisDatabaseList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RedisDatabaseList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RedisDatabaseSpec) DeepCopyInto(out *RedisDatabaseSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RedisDatabaseSpec.
func (in *RedisDatabaseSpec) DeepCopy() *RedisDatabaseSpec {
	if in == nil {
		return nil
	}
	out := new(RedisDatabaseSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RedisDatabaseStatus) DeepCopyInto(out *RedisDatabaseStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RedisDatabaseStatus.
func (in *RedisDatabaseStatus) DeepCopy() *RedisDatabaseStatus {
	if in == nil {
		return nil
	}
	out := new(RedisDatabaseStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RestoreConf) DeepCopyInto(out *RestoreConf) {
	*out = *in
	out.Repository = in.Repository
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RestoreConf.
func (in *RestoreConf) DeepCopy() *RestoreConf {
	if in == nil {
		return nil
	}
	out := new(RestoreConf)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TTL) DeepCopyInto(out *TTL) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TTL.
func (in *TTL) DeepCopy() *TTL {
	if in == nil {
		return nil
	}
	out := new(TTL)
	in.DeepCopyInto(out)
	return out
}
