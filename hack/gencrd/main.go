package main

import (
	"github.com/appscode/go/log"
	gort "github.com/appscode/go/runtime"
	crdutils "github.com/appscode/kutil/apiextensions/v1beta1"
	"github.com/appscode/kutil/openapi"
	"github.com/go-openapi/spec"
	"github.com/golang/glog"
	"github.com/kubedb/apimachinery/apis/kubedb/install"
	"github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"io/ioutil"
	crd_api "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/kube-openapi/pkg/common"
	"os"
	"path/filepath"
)

func generateCRDDefinitions() {
	filename := gort.GOPath() + "/src/github.com/kubedb/apimachinery/apis/kubedb/v1alpha1/crds.yaml"

	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	crds := []*crd_api.CustomResourceDefinition{
		v1alpha1.Postgres{}.CustomResourceDefinition(),
		v1alpha1.Elasticsearch{}.CustomResourceDefinition(),
		v1alpha1.MySQL{}.CustomResourceDefinition(),
		v1alpha1.MongoDB{}.CustomResourceDefinition(),
		v1alpha1.Redis{}.CustomResourceDefinition(),
		v1alpha1.Memcached{}.CustomResourceDefinition(),
		v1alpha1.Snapshot{}.CustomResourceDefinition(),
		v1alpha1.DormantDatabase{}.CustomResourceDefinition(),
	}
	for _, crd := range crds {
		crdutils.MarshallCrd(f, crd, "yaml")
	}
}

func generateSwaggerJson() {
	var (
		Scheme = runtime.NewScheme()
		Codecs = serializer.NewCodecFactory(Scheme)
	)

	install.Install(Scheme)

	apispec, err := openapi.RenderOpenAPISpec(openapi.Config{
		Scheme: Scheme,
		Codecs: Codecs,
		Info: spec.InfoProps{
			Title:   "KubeDB",
			Version: "v0",
			Contact: &spec.ContactInfo{
				Name:  "AppsCode Inc.",
				URL:   "https://appscode.com",
				Email: "hello@appscode.com",
			},
			License: &spec.License{
				Name: "Apache 2.0",
				URL:  "https://www.apache.org/licenses/LICENSE-2.0.html",
			},
		},
		OpenAPIDefinitions: []common.GetOpenAPIDefinitions{
			v1alpha1.GetOpenAPIDefinitions,
		},
		Resources: []openapi.TypeInfo{
			{v1alpha1.SchemeGroupVersion, v1alpha1.ResourcePluralPostgres, v1alpha1.ResourceKindPostgres, true},
			{v1alpha1.SchemeGroupVersion, v1alpha1.ResourcePluralElasticsearch, v1alpha1.ResourceKindElasticsearch, true},
			{v1alpha1.SchemeGroupVersion, v1alpha1.ResourcePluralMongoDB, v1alpha1.ResourceKindMongoDB, true},
			{v1alpha1.SchemeGroupVersion, v1alpha1.ResourcePluralMySQL, v1alpha1.ResourceKindMySQL, true},
			{v1alpha1.SchemeGroupVersion, v1alpha1.ResourcePluralRedis, v1alpha1.ResourceKindRedis, true},
			{v1alpha1.SchemeGroupVersion, v1alpha1.ResourcePluralMemcached, v1alpha1.ResourceKindMemcached, true},
			{v1alpha1.SchemeGroupVersion, v1alpha1.ResourcePluralSnapshot, v1alpha1.ResourceKindSnapshot, true},
			{v1alpha1.SchemeGroupVersion, v1alpha1.ResourcePluralDormantDatabase, v1alpha1.ResourceKindDormantDatabase, true},
		},
	})
	if err != nil {
		glog.Fatal(err)
	}

	filename := gort.GOPath() + "/src/github.com/kubedb/apimachinery/api/openapi-spec/swagger.json"
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
	generateCRDDefinitions()
	generateSwaggerJson()
}
