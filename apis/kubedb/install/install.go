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

package install

import (
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"
	"kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

// Install registers the API group and adds types to a scheme
func Install(scheme *runtime.Scheme) {
	utilruntime.Must(v1alpha2.AddToScheme(scheme))
	utilruntime.Must(dbapi.AddToScheme(scheme))
	utilruntime.Must(scheme.SetVersionPriority(dbapi.SchemeGroupVersion, v1alpha2.SchemeGroupVersion))
}
