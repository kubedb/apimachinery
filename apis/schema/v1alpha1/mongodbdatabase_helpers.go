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

func (in MongoDBDatabase) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(GroupVersion.WithResource(ResourceMongoDBDatabases))
}

var _ Interface = &MongoDBDatabase{}

func (in *MongoDBDatabase) GetInit() *InitSpec {
	return in.Spec.Init
}

func (in *MongoDBDatabase) GetStatus() DatabaseStatus {
	return in.Status
}

func (in *MongoDBDatabase) GetMongoInitVolumeNameForPod() string {
	return in.GetName() + "-init-volume"
}
func (in *MongoDBDatabase) GetMongoInitJobName() string {
	return in.GetName() + "-init-job"
}
func (in *MongoDBDatabase) GetMongoInitScriptContainerName() string {
	return in.GetName() + "-init-container"
}
func (in *MongoDBDatabase) GetMongoRestoreSessionName() string {
	return in.GetName() + "-restore-session"
}

// For MongoDB Admin Role
func (in *MongoDBDatabase) GetMongoAdminRoleName() string {
	return in.GetName() + MongoPrefix + "-admin-role"
}
func (in *MongoDBDatabase) GetMongoAdminSecretAccessRequestName() string {
	return in.GetName() + MongoPrefix + "-admin-secret-access-req"
}
func (in *MongoDBDatabase) GetMongoAdminServiceAccountName() string {
	return in.GetName() + MongoPrefix + "-admin-service-account"
}

func (in *MongoDBDatabase) GetMongoSecretEngineName() string {
	return in.GetName() + MongoPrefix + "-secret-engine"
}

func (in *MongoDBDatabase) GetAuthSecretName(dbServerName string) string {
	return dbServerName + kdm.MongoDBAuthSecretSuffix
}
