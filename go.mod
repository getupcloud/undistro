module github.com/getupcloud/undistro

go 1.13

require (
	github.com/blang/semver v3.5.1+incompatible
	github.com/coredns/corefile-migration v1.0.7
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/davecgh/go-spew v1.1.1
	github.com/docker/distribution v2.7.1+incompatible
	github.com/evanphx/json-patch v4.5.0+incompatible
	github.com/go-logr/logr v0.1.0
	github.com/gogo/protobuf v1.3.1
	github.com/google/go-cmp v0.4.1
	github.com/google/gofuzz v1.1.0
	github.com/onsi/ginkgo v1.12.1
	github.com/onsi/gomega v1.10.1
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.5.1
	github.com/prometheus/client_model v0.2.0
	github.com/spf13/pflag v1.0.5
	github.com/tmc/grpc-websocket-proxy v0.0.0-20190109142713-0ad062ec5ee5 // indirect
	go.etcd.io/etcd v0.0.0-20191023171146-3cf2f69b5738
	go.uber.org/atomic v1.4.0 // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	google.golang.org/appengine v1.6.6 // indirect
	google.golang.org/grpc v1.26.0
	k8s.io/api v0.17.8
	k8s.io/apiextensions-apiserver v0.17.8
	k8s.io/apimachinery v0.17.8
	k8s.io/apiserver v0.17.8
	k8s.io/client-go v0.17.8
	k8s.io/cluster-bootstrap v0.17.8
	k8s.io/klog v1.0.0
	k8s.io/utils v0.0.0-20200619165400-6e3d28b6ed19
	sigs.k8s.io/controller-runtime v0.5.9
	sigs.k8s.io/kind v0.7.1-0.20200303021537-981bd80d3802
	sigs.k8s.io/yaml v1.2.0
)

replace github.com/evanphx/json-patch => github.com/evanphx/json-patch v0.0.0-20190815234213-e83c0a1c26c8
