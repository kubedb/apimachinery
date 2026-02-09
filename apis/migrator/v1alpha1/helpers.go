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

func (Migrator) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourcePluralMigrator))
}

func GetDatabase(migrator *Migrator) string {
	switch {
	case migrator.Spec.Source.Postgres != nil && migrator.Spec.Target.Postgres != nil:
		return "postgres"
		// case migrator.Spec.Source.MySQL != nil && migrator.Spec.Target.MySQL != nil:
		//	return m.MigratorImages.MySQL
		// case migrator.Spec.Source.MariaDB != nil && migrator.Spec.Target.MariaDB != nil:
		//	return m.MigratorImages.MariaDB
		// case migrator.Spec.Source.MSSQLServer != nil && migrator.Spec.Target.MSSQLServer != nil:
		//	return m.MigratorImages.MSSQLServer
		// case migrator.Spec.Source.MongoDB != nil && migrator.Spec.Target.MongoDB != nil:
		//	return m.MigratorImages.MongoDB
	}
	return ""
}
