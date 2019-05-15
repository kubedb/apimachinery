module github.com/kubedb/apimachinery

go 1.12

require (
	github.com/OneOfOne/xxhash v1.2.5 // indirect
	github.com/appscode/docker-registry-client v0.0.0-20180426150142-1bb02bb202b0
	github.com/appscode/go v0.0.0-20190523031839-1468ee3a76e8
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7 // indirect
	github.com/evanphx/json-patch v4.2.0+incompatible
	github.com/go-openapi/spec v0.19.0
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/gorilla/mux v1.7.2 // indirect
	github.com/graymeta/stow v0.0.0-00010101000000-000000000000
	github.com/json-iterator/go v1.1.6
	github.com/kr/pty v1.1.4 // indirect
	github.com/magiconair/properties v1.8.1 // indirect
	github.com/mattn/go-colorable v0.1.2 // indirect
	github.com/orcaman/concurrent-map v0.0.0-20190314100340-2693aad1ed75
	github.com/pkg/errors v0.8.1
	github.com/prometheus/common v0.4.1 // indirect
	google.golang.org/genproto v0.0.0-20190508193815-b515fa19cec8 // indirect
	gopkg.in/robfig/cron.v2 v2.0.0-20150107220207-be2e0b0deed5
	gopkg.in/square/go-jose.v2 v2.2.2 // indirect
	k8s.io/api v0.0.0-20190503110853-61630f889b3c
	k8s.io/apiextensions-apiserver v0.0.0-20190508184259-7784d62bc471
	k8s.io/apimachinery v0.0.0-20190508063446-a3da69d3723c
	k8s.io/apiserver v0.0.0-20190508183956-3a0abf14e58a // indirect
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/klog v0.3.1 // indirect
	k8s.io/kube-openapi v0.0.0-20190502190224-411b2483e503
	k8s.io/kubernetes v1.14.2 // indirect
	kmodules.xyz/client-go v0.0.0-20190524133821-9c8a87771aea
	kmodules.xyz/custom-resources v0.0.0-20190508103408-464e8324c3ec
	kmodules.xyz/monitoring-agent-api v0.0.0-20190508125842-489150794b9b
	kmodules.xyz/objectstore-api v0.0.0-20190506085934-94c81c8acca9
	kmodules.xyz/offshoot-api v0.0.0-20190513045534-4f3df05f40c2
	kmodules.xyz/webhook-runtime v0.0.0-20190508093950-b721b4eba5e5
	stash.appscode.dev/stash v0.0.0-20190525101310-604af0d5fc92
)

replace (
	github.com/graymeta/stow => github.com/appscode/stow v0.0.0-20190506085026-ca5baa008ea3
	gopkg.in/robfig/cron.v2 => github.com/appscode/cron v0.0.0-20170717094345-ca60c6d796d4
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
