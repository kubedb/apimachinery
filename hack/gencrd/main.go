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

package main

import (
	"io/ioutil"
	"os"
	"path/filepath"

	cataloginstall "kubedb.dev/apimachinery/apis/catalog/install"
	catalogv1alpha1 "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	kubedbinstall "kubedb.dev/apimachinery/apis/kubedb/install"
	kubedbv1alpha2 "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	opsinstall "kubedb.dev/apimachinery/apis/ops/install"
	opsv1alpha1 "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	"github.com/go-openapi/spec"
	"github.com/golang/glog"
	gort "gomodules.xyz/runtime"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/kube-openapi/pkg/common"
	"kmodules.xyz/client-go/openapi"
)

func generateSwaggerJson() {
	var (
		Scheme = runtime.NewScheme()
		Codecs = serializer.NewCodecFactory(Scheme)
	)

	kubedbinstall.Install(Scheme)
	cataloginstall.Install(Scheme)
	opsinstall.Install(Scheme)

	apispec, err := openapi.RenderOpenAPISpec(openapi.Config{
		Scheme: Scheme,
		Codecs: Codecs,
		Info: spec.InfoProps{
			Title:   "KubeDB",
			Version: "v0",
			Contact: &spec.ContactInfo{
				ContactInfoProps: spec.ContactInfoProps{
					Name:  "AppsCode Inc.",
					URL:   "https://appscode.com",
					Email: "hello@appscode.com",
				},
			},
			License: &spec.License{
				LicenseProps: spec.LicenseProps{
					Name: "Apache 2.0",
					URL:  "https://www.apache.org/licenses/LICENSE-2.0.html",
				},
			},
		},
		OpenAPIDefinitions: []common.GetOpenAPIDefinitions{
			kubedbv1alpha2.GetOpenAPIDefinitions,
			catalogv1alpha1.GetOpenAPIDefinitions,
			opsv1alpha1.GetOpenAPIDefinitions,
		},
		//nolint:govet
		Resources: []openapi.TypeInfo{
			{kubedbv1alpha2.SchemeGroupVersion, kubedbv1alpha2.ResourcePluralElasticsearch, kubedbv1alpha2.ResourceKindElasticsearch, true},
			{kubedbv1alpha2.SchemeGroupVersion, kubedbv1alpha2.ResourcePluralEtcd, kubedbv1alpha2.ResourceKindEtcd, true},
			{kubedbv1alpha2.SchemeGroupVersion, kubedbv1alpha2.ResourcePluralMariaDB, kubedbv1alpha2.ResourceKindMariaDB, true},
			{kubedbv1alpha2.SchemeGroupVersion, kubedbv1alpha2.ResourcePluralMemcached, kubedbv1alpha2.ResourceKindMemcached, true},
			{kubedbv1alpha2.SchemeGroupVersion, kubedbv1alpha2.ResourcePluralMongoDB, kubedbv1alpha2.ResourceKindMongoDB, true},
			{kubedbv1alpha2.SchemeGroupVersion, kubedbv1alpha2.ResourcePluralMySQL, kubedbv1alpha2.ResourceKindMySQL, true},
			{kubedbv1alpha2.SchemeGroupVersion, kubedbv1alpha2.ResourcePluralPerconaXtraDB, kubedbv1alpha2.ResourceKindPerconaXtraDB, true},
			{kubedbv1alpha2.SchemeGroupVersion, kubedbv1alpha2.ResourcePluralPgBouncer, kubedbv1alpha2.ResourceKindPgBouncer, true},
			{kubedbv1alpha2.SchemeGroupVersion, kubedbv1alpha2.ResourcePluralPostgres, kubedbv1alpha2.ResourceKindPostgres, true},
			{kubedbv1alpha2.SchemeGroupVersion, kubedbv1alpha2.ResourcePluralProxySQL, kubedbv1alpha2.ResourceKindProxySQL, true},
			{kubedbv1alpha2.SchemeGroupVersion, kubedbv1alpha2.ResourcePluralRedis, kubedbv1alpha2.ResourceKindRedis, true},

			{opsv1alpha1.SchemeGroupVersion, opsv1alpha1.ResourcePluralElasticsearchOpsRequest, opsv1alpha1.ResourceKindElasticsearchOpsRequest, true},
			{opsv1alpha1.SchemeGroupVersion, opsv1alpha1.ResourcePluralEtcdOpsRequest, opsv1alpha1.ResourceKindEtcdOpsRequest, true},
			{opsv1alpha1.SchemeGroupVersion, opsv1alpha1.ResourcePluralMemcachedOpsRequest, opsv1alpha1.ResourceKindMemcachedOpsRequest, true},
			{opsv1alpha1.SchemeGroupVersion, opsv1alpha1.ResourcePluralMongoDBOpsRequest, opsv1alpha1.ResourceKindMongoDBOpsRequest, true},
			{opsv1alpha1.SchemeGroupVersion, opsv1alpha1.ResourcePluralMySQLOpsRequest, opsv1alpha1.ResourceKindMySQLOpsRequest, true},
			{opsv1alpha1.SchemeGroupVersion, opsv1alpha1.ResourcePluralPerconaXtraDBOpsRequest, opsv1alpha1.ResourceKindPerconaXtraDBOpsRequest, true},
			{opsv1alpha1.SchemeGroupVersion, opsv1alpha1.ResourcePluralPostgresOpsRequest, opsv1alpha1.ResourceKindPostgresOpsRequest, true},
			{opsv1alpha1.SchemeGroupVersion, opsv1alpha1.ResourcePluralProxySQLOpsRequest, opsv1alpha1.ResourceKindProxySQLOpsRequest, true},
			{opsv1alpha1.SchemeGroupVersion, opsv1alpha1.ResourcePluralRedisOpsRequest, opsv1alpha1.ResourceKindRedisOpsRequest, true},

			{catalogv1alpha1.SchemeGroupVersion, catalogv1alpha1.ResourcePluralElasticsearchVersion, catalogv1alpha1.ResourceKindElasticsearchVersion, false},
			{catalogv1alpha1.SchemeGroupVersion, catalogv1alpha1.ResourcePluralEtcdVersion, catalogv1alpha1.ResourceKindEtcdVersion, false},
			{catalogv1alpha1.SchemeGroupVersion, catalogv1alpha1.ResourcePluralMemcachedVersion, catalogv1alpha1.ResourceKindMemcachedVersion, false},
			{catalogv1alpha1.SchemeGroupVersion, catalogv1alpha1.ResourcePluralMongoDBVersion, catalogv1alpha1.ResourceKindMongoDBVersion, false},
			{catalogv1alpha1.SchemeGroupVersion, catalogv1alpha1.ResourcePluralMySQLVersion, catalogv1alpha1.ResourceKindMySQLVersion, false},
			{catalogv1alpha1.SchemeGroupVersion, catalogv1alpha1.ResourcePluralPerconaXtraDBVersion, catalogv1alpha1.ResourceKindPerconaXtraDBVersion, false},
			{catalogv1alpha1.SchemeGroupVersion, catalogv1alpha1.ResourcePluralPostgresVersion, catalogv1alpha1.ResourceKindPostgresVersion, false},
			{catalogv1alpha1.SchemeGroupVersion, catalogv1alpha1.ResourcePluralProxySQLVersion, catalogv1alpha1.ResourceKindProxySQLVersion, false},
			{catalogv1alpha1.SchemeGroupVersion, catalogv1alpha1.ResourcePluralRedisVersion, catalogv1alpha1.ResourceKindRedisVersion, false},
		},
	})
	if err != nil {
		glog.Fatal(err)
	}

	filename := gort.GOPath() + "/src/kubedb.dev/apimachinery/openapi/swagger.json"
	err = os.MkdirAll(filepath.Dir(filename), 0755)
	if err != nil {
		glog.Fatal(err)
	}
	err = ioutil.WriteFile(filename, []byte(apispec), 0644)
	if err != nil {
		glog.Fatal(err)
	}
}

func main() {
	generateSwaggerJson()
}
