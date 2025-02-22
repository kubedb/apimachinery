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
	"kubedb.dev/apimachinery/crds"

	"kmodules.xyz/client-go/apiextensions"
)

func GetFinalizer() string {
	return SchemeGroupVersion.Group
}

func (_ MongoDBArchiver) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourcePluralMongoDBArchiver))
}

func (_ PostgresArchiver) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourcePluralPostgresArchiver))
}

func (_ MySQLArchiver) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourcePluralMySQLArchiver))
}

func (_ MariaDBArchiver) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourcePluralMariaDBArchiver))
}

func (_ MSSQLServerArchiver) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourcePluralMSSQLServerArchiver))
}

func SetDefaultLogBackupOptions(log *LogBackupOptions) *LogBackupOptions {
	if log == nil {
		log = &LogBackupOptions{
			SuccessfulLogHistoryLimit: 5,
			FailedLogHistoryLimit:     5,
		}
	}
	return log
}
