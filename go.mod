module kubedb.dev/apimachinery

go 1.12

require (
	github.com/appscode/docker-registry-client v0.0.0-20180426150142-1bb02bb202b0
	github.com/appscode/go v0.0.0-20190621064509-6b292c9166e3
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7 // indirect
	github.com/evanphx/json-patch v4.5.0+incompatible
	github.com/go-openapi/spec v0.19.0
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/gorilla/mux v1.7.2 // indirect
	github.com/json-iterator/go v1.1.6
	github.com/mattn/go-colorable v0.1.2 // indirect
	github.com/orcaman/concurrent-map v0.0.0-20190314100340-2693aad1ed75
	github.com/pkg/errors v0.8.1
	github.com/robfig/cron/v3 v3.0.0
	gomodules.xyz/stow v0.2.0
	k8s.io/api v0.0.0-20190503110853-61630f889b3c
	k8s.io/apiextensions-apiserver v0.0.0-20190516231611-bf6753f2aa24
	k8s.io/apimachinery v0.0.0-20190508063446-a3da69d3723c
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/klog v0.3.1 // indirect
	k8s.io/kube-openapi v0.0.0-20190502190224-411b2483e503
	k8s.io/kubernetes v1.14.2 // indirect
	kmodules.xyz/client-go v0.0.0-20190715080709-7162a6c90b04
	kmodules.xyz/custom-resources v0.0.0-20190508103408-464e8324c3ec
	kmodules.xyz/monitoring-agent-api v0.0.0-20190513065523-186af167f817
	kmodules.xyz/objectstore-api v0.0.0-20190718002052-da668b440b0b
	kmodules.xyz/offshoot-api v0.0.0-20190715115723-36c8fce142c1
	kmodules.xyz/webhook-runtime v0.0.0-20190715115250-a84fbf77dd30
	stash.appscode.dev/stash v0.0.0-20190717091949-1c51a7c58f08
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest/autorest v0.5.0
	k8s.io/api => k8s.io/api v0.0.0-20190313235455-40a48860b5ab
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190315093550-53c4693659ed
	k8s.io/apimachinery => github.com/kmodules/apimachinery v0.0.0-20190508045248-a52a97a7a2bf
	k8s.io/apiserver => github.com/kmodules/apiserver v0.0.0-20190508082252-8397d761d4b5
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20190314001948-2899ed30580f
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20190314002645-c892ea32361a
	k8s.io/component-base => k8s.io/component-base v0.0.0-20190314000054-4a91899592f4
	k8s.io/klog => k8s.io/klog v0.3.0
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20190314000639-da8327669ac5
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20190228160746-b3a7cee44a30
	k8s.io/metrics => k8s.io/metrics v0.0.0-20190314001731-1bd6a4002213
	k8s.io/utils => k8s.io/utils v0.0.0-20190221042446-c2654d5206da
)
