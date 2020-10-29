module github.com/getupio-undistro/undistro

go 1.15

require (
	github.com/MakeNowJust/heredoc/v2 v2.0.1
	github.com/Masterminds/sprig/v3 v3.1.0
	github.com/aws/aws-sdk-go v1.35.18
	github.com/docker/docker v17.12.0-ce-rc1.0.20201019175245-6f78b438b885+incompatible // indirect
	github.com/drone/envsubst v1.0.3-0.20200709223903-efdb65b94e5a
	github.com/fatih/color v1.9.0
	github.com/go-logr/logr v0.2.1
	github.com/google/go-cmp v0.5.2
	github.com/google/go-github/v32 v32.1.0
	github.com/moby/term v0.0.0-20200915141129-7f0af18e79f2 // indirect
	github.com/ncabatoff/go-seq v0.0.0-20180805175032-b08ef85ed833
	github.com/onsi/ginkgo v1.14.2
	github.com/onsi/gomega v1.10.3
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.1.1
	github.com/spf13/viper v1.7.1
	golang.org/x/oauth2 v0.0.0-20200902213428-5d25da1a8d43
	helm.sh/helm/v3 v3.4.0
	k8s.io/api v0.19.3
	k8s.io/apiextensions-apiserver v0.19.3
	k8s.io/apimachinery v0.19.3
	k8s.io/cli-runtime v0.19.3
	k8s.io/client-go v0.19.3
	k8s.io/kubectl v0.19.3
	k8s.io/utils v0.0.0-20201027101359-01387209bb0d
	rsc.io/letsencrypt v0.0.3 // indirect
	sigs.k8s.io/cluster-api v0.3.11-0.20201016161926-0008b5ba109e
	sigs.k8s.io/cluster-api-provider-aws v0.6.1
	sigs.k8s.io/cluster-api-provider-vsphere v0.7.1
	sigs.k8s.io/controller-runtime v0.7.0-alpha.4
	sigs.k8s.io/yaml v1.2.0

)

replace github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible
