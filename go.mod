module kubedb.dev/apimachinery

go 1.24.0

toolchain go1.24.1

require (
	github.com/Masterminds/semver/v3 v3.3.1
	github.com/cert-manager/cert-manager v1.18.2
	github.com/coreos/go-semver v0.3.1
	github.com/evanphx/json-patch v5.9.11+incompatible
	github.com/fatih/structs v1.1.0
	github.com/go-logr/logr v1.4.2
	github.com/go-logr/zapr v1.3.0
	github.com/google/go-cmp v0.7.0
	github.com/google/gofuzz v1.2.0
	github.com/google/uuid v1.6.0
	github.com/iancoleman/orderedmap v0.3.0
	github.com/json-iterator/go v1.1.12
	github.com/kubernetes-csi/external-snapshotter/client/v8 v8.2.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.81.0
	github.com/prometheus-operator/prometheus-operator/pkg/client v0.81.0
	go.bytebuilders.dev/audit v0.0.43
	go.bytebuilders.dev/license-verifier/kubernetes v0.14.9
	go.uber.org/zap v1.27.0
	go.virtual-secrets.dev/apimachinery v0.0.1
	gomodules.xyz/encoding v0.0.8
	gomodules.xyz/pointer v0.1.0
	gomodules.xyz/runtime v0.3.0
	gomodules.xyz/stow v0.2.4
	gomodules.xyz/x v0.0.17
	k8s.io/api v0.32.3
	k8s.io/apiextensions-apiserver v0.32.3
	k8s.io/apimachinery v0.32.3
	k8s.io/apiserver v0.32.3
	k8s.io/client-go v0.32.3
	k8s.io/component-base v0.32.3
	k8s.io/klog/v2 v2.130.1
	k8s.io/kube-aggregator v0.32.3
	k8s.io/kube-openapi v0.0.0-20250318190949-c8a335a9a2ff
	k8s.io/metrics v0.32.3
	k8s.io/utils v0.0.0-20241210054802-24370beab758
	kmodules.xyz/client-go v0.32.7
	kmodules.xyz/crd-schema-fuzz v0.32.0
	kmodules.xyz/custom-resources v0.32.0
	kmodules.xyz/monitoring-agent-api v0.32.1
	kmodules.xyz/objectstore-api v0.32.0
	kmodules.xyz/offshoot-api v0.32.0
	kmodules.xyz/resource-metadata v0.32.1
	kmodules.xyz/webhook-runtime v0.32.0
	kubeops.dev/csi-driver-cacerts v0.1.0
	kubeops.dev/operator-shard-manager v0.0.3
	kubeops.dev/petset v0.0.12-0.20250820054807-88384ed8aba9
	kubeops.dev/sidekick v0.0.11
	kubestash.dev/apimachinery v0.20.0
	open-cluster-management.io/api v1.0.0
	sigs.k8s.io/controller-runtime v0.20.4
	sigs.k8s.io/yaml v1.4.0
	stash.appscode.dev/apimachinery v0.41.0
)

require (
	cel.dev/expr v0.19.1 // indirect
	cloud.google.com/go/auth v0.9.4 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.4 // indirect
	cloud.google.com/go/compute/metadata v0.6.0 // indirect
	dario.cat/mergo v1.0.1 // indirect
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/Azure/azure-sdk-for-go v68.0.0+incompatible // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest v0.11.29 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.24 // indirect
	github.com/Azure/go-autorest/autorest/date v0.3.0 // indirect
	github.com/Azure/go-autorest/logger v0.2.1 // indirect
	github.com/Azure/go-autorest/tracing v0.6.0 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/sprig/v3 v3.3.0 // indirect
	github.com/NYTimes/gziphandler v1.1.1 // indirect
	github.com/PuerkitoBio/purell v1.2.1 // indirect
	github.com/antlr4-go/antlr/v4 v4.13.1 // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/aws/aws-sdk-go v1.55.6 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cloudevents/sdk-go/v2 v2.15.2 // indirect
	github.com/containerd/stargz-snapshotter/estargz v0.16.3 // indirect
	github.com/coreos/go-systemd/v22 v22.5.0 // indirect
	github.com/cyphar/filepath-securejoin v0.3.4 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dnaeon/go-vcr v1.2.0 // indirect
	github.com/docker/cli v27.5.0+incompatible // indirect
	github.com/docker/distribution v2.8.3+incompatible // indirect
	github.com/docker/docker-credential-helpers v0.8.2 // indirect
	github.com/emicklei/go-restful/v3 v3.12.1 // indirect
	github.com/evanphx/json-patch/v5 v5.9.11 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.8.0 // indirect
	github.com/fxamacker/cbor/v2 v2.7.0 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/go-sql-driver/mysql v1.9.0 // indirect
	github.com/gofrs/uuid v4.4.0+incompatible // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.2 // indirect
	github.com/golang/groupcache v0.0.0-20241129210726-2c02b8208cf8 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/google/cel-go v0.22.1 // indirect
	github.com/google/gnostic-models v0.6.9 // indirect
	github.com/google/go-containerregistry v0.20.3 // indirect
	github.com/google/s2a-go v0.1.8 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.4 // indirect
	github.com/googleapis/gax-go/v2 v2.13.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.25.1 // indirect
	github.com/huandu/xstrings v1.5.0 // indirect
	github.com/imdario/mergo v0.3.16 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jmespath/go-jmespath v0.4.1-0.20220621161143-b0104c826a24 // indirect
	github.com/jonboulle/clockwork v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.5 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/mailru/easyjson v0.9.0 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nats-io/nats.go v1.44.0 // indirect
	github.com/nats-io/nkeys v0.4.11 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/onsi/gomega v1.36.2 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0 // indirect
	github.com/prometheus/client_golang v1.20.5 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.61.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/rancher/norman v0.5.2 // indirect
	github.com/rancher/rancher/pkg/client v0.0.0-20250220153925-3abb578f42fe // indirect
	github.com/rancher/wrangler/v3 v3.2.0-rc.3 // indirect
	github.com/sergi/go-diff v1.3.1 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/spf13/cast v1.7.0 // indirect
	github.com/spf13/cobra v1.9.1 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	github.com/stoewer/go-strcase v1.3.0 // indirect
	github.com/vbatts/tar-split v0.11.6 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/yudai/gojsondiff v1.0.0 // indirect
	github.com/yudai/golcs v0.0.0-20170316035057-ecda9a501e82 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	go.bytebuilders.dev/license-proxyserver v0.0.20 // indirect
	go.bytebuilders.dev/license-verifier v0.14.9 // indirect
	go.etcd.io/etcd/api/v3 v3.5.17 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.5.17 // indirect
	go.etcd.io/etcd/client/v3 v3.5.17 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.58.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.58.0 // indirect
	go.opentelemetry.io/otel v1.33.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.33.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.33.0 // indirect
	go.opentelemetry.io/otel/metric v1.33.0 // indirect
	go.opentelemetry.io/otel/sdk v1.33.0 // indirect
	go.opentelemetry.io/otel/trace v1.33.0 // indirect
	go.opentelemetry.io/proto/otlp v1.4.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.38.0 // indirect
	golang.org/x/exp v0.0.0-20241217172543-b2144cdd0a67 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/oauth2 v0.28.0 // indirect
	golang.org/x/sync v0.14.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/term v0.32.0 // indirect
	golang.org/x/text v0.25.0 // indirect
	golang.org/x/time v0.10.0 // indirect
	gomodules.xyz/counter v0.0.1 // indirect
	gomodules.xyz/jsonpatch/v2 v2.5.0 // indirect
	gomodules.xyz/mergo v0.3.13 // indirect
	gomodules.xyz/sets v0.2.1 // indirect
	gomodules.xyz/sync v0.1.0 // indirect
	google.golang.org/api v0.198.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20241219192143-6b3ec007d9bb // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241219192143-6b3ec007d9bb // indirect
	google.golang.org/grpc v1.69.2 // indirect
	google.golang.org/protobuf v1.36.3 // indirect
	gopkg.in/evanphx/json-patch.v4 v4.12.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/kms v0.32.3 // indirect
	kmodules.xyz/apiversion v0.2.0 // indirect
	kmodules.xyz/go-containerregistry v0.0.14 // indirect
	kmodules.xyz/prober v0.32.0 // indirect
	kmodules.xyz/resource-metrics v0.30.9 // indirect
	moul.io/http2curl/v2 v2.3.1-0.20221024080105-10c404f653f7 // indirect
	sigs.k8s.io/apiserver-network-proxy/konnectivity-client v0.31.1 // indirect
	sigs.k8s.io/cli-utils v0.37.2 // indirect
	sigs.k8s.io/gateway-api v1.1.0 // indirect
	sigs.k8s.io/json v0.0.0-20241014173422-cfa47c3a1cc8 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.6.0 // indirect
	x-helm.dev/apimachinery v0.0.17 // indirect
)
