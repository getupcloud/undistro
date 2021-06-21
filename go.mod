module github.com/getupio-undistro/undistro

go 1.16

replace sigs.k8s.io/cluster-api => github.com/getupio-undistro/cluster-api v0.3.11-0.20210608202222-4052f5f87b90

require (
	github.com/MakeNowJust/heredoc/v2 v2.0.1
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/Masterminds/sprig/v3 v3.2.2
	github.com/aws/aws-sdk-go v1.38.58
	github.com/go-logr/logr v0.4.0
	github.com/gorilla/mux v1.8.0
	github.com/json-iterator/go v1.1.11
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.13.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.8.0
	helm.sh/helm/v3 v3.6.1
	k8s.io/api v0.21.1
	k8s.io/apiextensions-apiserver v0.21.1
	k8s.io/apimachinery v0.21.1
	k8s.io/cli-runtime v0.21.1
	k8s.io/client-go v0.21.1
	k8s.io/klog/v2 v2.9.0
	k8s.io/kubectl v0.21.1
	k8s.io/utils v0.0.0-20210527160623-6fdb442a123b // indirect
	rsc.io/letsencrypt v0.0.3 // indirect
	sigs.k8s.io/cluster-api v0.0.0-00010101000000-000000000000
	sigs.k8s.io/controller-runtime v0.9.0
	sigs.k8s.io/yaml v1.2.0
)
