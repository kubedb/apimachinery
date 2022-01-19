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
	kdm "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/apimachinery/crds"

	"kmodules.xyz/client-go/apiextensions"
)

const (
	InitScriptName              string = "init.js"
	MongoInitScriptPath         string = "/init-scripts"
	MongoPrefix                 string = "-mongo"
	MongoDatabaseNameForEntry   string = "kubedb-system"
	MongoCollectionNameForEntry string = "databases"
)

func (_ MongoDBDatabase) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(GroupVersion.WithResource(ResourceMongoDBDatabases))
}

var _ Interface = &MongoDBDatabase{}

func (in *MongoDBDatabase) GetInit() *InitSpec {
	return in.Spec.Init
}

func (in *MongoDBDatabase) GetStatus() DatabaseStatus {
	return in.Status
}

func (db *MongoDBDatabase) GetMongoInitVolumeNameForPod() string {
	return db.GetName() + "-init-volume"
}
func (db *MongoDBDatabase) GetMongoInitJobName() string {
	return db.GetName() + "-init-job"
}
func (db *MongoDBDatabase) GetMongoInitScriptContainerName() string {
	return db.GetName() + "-init-container"
}
func (db *MongoDBDatabase) GetMongoRestoreSessionName() string {
	return db.GetName() + "-restore-session"
}

// For MongoDB Admin Role
func (db *MongoDBDatabase) GetMongoAdminRoleName() string {
	return db.GetName() + MongoPrefix + "-admin-role"
}
func (db *MongoDBDatabase) GetMongoAdminSecretAccessRequestName() string {
	return db.GetName() + MongoPrefix + "-admin-secret-access-req"
}
func (db *MongoDBDatabase) GetMongoAdminServiceAccountName() string {
	return db.GetName() + MongoPrefix + "-admin-service-account"
}

func (db *MongoDBDatabase) GetMongoSecretEngineName() string {
	return db.GetName() + MongoPrefix + "-secret-engine"
}

func (db *MongoDBDatabase) GetAuthSecretName(dbServerName string) string {
	return dbServerName + kdm.MongoDBAuthSecretSuffix
}
