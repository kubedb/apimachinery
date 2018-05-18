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
	"k8s.io/apimachinery/pkg/apimachinery/announced"
	"k8s.io/apimachinery/pkg/apimachinery/registered"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/kube-openapi/pkg/common"
	"os"
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
		groupFactoryRegistry = make(announced.APIGroupFactoryRegistry)
		registry             = registered.NewOrDie("")
		Scheme               = runtime.NewScheme()
		Codecs               = serializer.NewCodecFactory(Scheme)
	)

	install.Install(groupFactoryRegistry, registry, Scheme)

	apispec, err := openapi.RenderOpenAPISpec(openapi.Config{
		Registry: registry,
		Scheme:   Scheme,
		Codecs:   Codecs,
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
		Resources: []schema.GroupVersionResource{
			v1alpha1.SchemeGroupVersion.WithResource(v1alpha1.ResourcePluralPostgres),
			v1alpha1.SchemeGroupVersion.WithResource(v1alpha1.ResourcePluralElasticsearch),
			v1alpha1.SchemeGroupVersion.WithResource(v1alpha1.ResourcePluralMongoDB),
			v1alpha1.SchemeGroupVersion.WithResource(v1alpha1.ResourcePluralMySQL),
			v1alpha1.SchemeGroupVersion.WithResource(v1alpha1.ResourcePluralRedis),
			v1alpha1.SchemeGroupVersion.WithResource(v1alpha1.ResourcePluralMemcached),
			v1alpha1.SchemeGroupVersion.WithResource(v1alpha1.ResourcePluralSnapshot),
			v1alpha1.SchemeGroupVersion.WithResource(v1alpha1.ResourcePluralDormantDatabase),
		},
	})
	if err != nil {
		glog.Fatal(err)
	}

	filename := gort.GOPath() + "/src/github.com/kubedb/apimachinery/openapi-spec/v2/swagger.json"
	err = ioutil.WriteFile(filename, []byte(apispec), 0644)
	if err != nil {
		glog.Fatal(err)
	}
}

func main() {
	generateCRDDefinitions()
	generateSwaggerJson()
}
