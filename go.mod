module github.com/getupio-undistro/undistro

go 1.16

require (
	github.com/MakeNowJust/heredoc/v2 v2.0.1
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/Masterminds/sprig/v3 v3.2.2
	github.com/aws/aws-sdk-go v1.38.56
	github.com/go-logr/logr v0.4.0
	github.com/gorilla/mux v1.8.0
	github.com/json-iterator/go v1.1.11
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.13.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	helm.sh/helm/v3 v3.5.4
	k8s.io/api v0.20.7
	k8s.io/apiextensions-apiserver v0.20.6
	k8s.io/apimachinery v0.20.7
	k8s.io/cli-runtime v0.20.7
	k8s.io/client-go v0.20.7
	k8s.io/klog/v2 v2.9.0
	k8s.io/kubectl v0.20.7
	sigs.k8s.io/cluster-api v0.3.17
	sigs.k8s.io/controller-runtime v0.8.3
	sigs.k8s.io/yaml v1.2.0
)

replace (
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible
	github.com/gorilla/rpc => github.com/getupio-undistro/rpc v1.2.1-0.20210520161253-a32dcc464920
	sigs.k8s.io/cluster-api => github.com/getupio-undistro/cluster-api v0.3.11-0.20210211140125-d0ad83191c76
)
