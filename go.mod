module github.com/getupcloud/undistro

go 1.13

require (
	github.com/MakeNowJust/heredoc/v2 v2.0.1
	github.com/aws/aws-sdk-go v1.34.18
	github.com/drone/envsubst v1.0.3-0.20200709223903-efdb65b94e5a
	github.com/go-logr/logr v0.1.0
	github.com/google/go-cmp v0.5.2
	github.com/google/go-github/v32 v32.1.0
	github.com/gophercloud/gophercloud v0.12.0 // indirect
	github.com/ncabatoff/go-seq v0.0.0-20180805175032-b08ef85ed833
	github.com/onsi/ginkgo v1.14.0
	github.com/onsi/gomega v1.10.1
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.7.1
	golang.org/x/oauth2 v0.0.0-20200902213428-5d25da1a8d43
	helm.sh/helm/v3 v3.3.1
	k8s.io/api v0.18.9-rc.0
	k8s.io/apiextensions-apiserver v0.18.9-rc.0
	k8s.io/apimachinery v0.18.9-rc.0
	k8s.io/cli-runtime v0.18.9-rc.0
	k8s.io/client-go v0.18.9-rc.0
	k8s.io/kubectl v0.18.9-rc.0
	k8s.io/utils v0.0.0-20200821003339-5e75c0163111
	rsc.io/letsencrypt v0.0.3 // indirect
	sigs.k8s.io/cluster-api v0.3.9
	sigs.k8s.io/cluster-api-provider-aws v0.5.5
	sigs.k8s.io/controller-runtime v0.6.2
	sigs.k8s.io/yaml v1.2.0
)

replace github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible
