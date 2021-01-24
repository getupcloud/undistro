module github.com/getupio-undistro/undistro

go 1.15

require (
	github.com/MakeNowJust/heredoc/v2 v2.0.1
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/Masterminds/sprig/v3 v3.2.0
	github.com/aws/aws-sdk-go v1.36.12
	github.com/go-logr/logr v0.3.0
	github.com/onsi/ginkgo v1.14.2
	github.com/onsi/gomega v1.10.4
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	helm.sh/helm/v3 v3.4.2
	k8s.io/api v0.20.2
	k8s.io/apiextensions-apiserver v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/cli-runtime v0.20.2
	k8s.io/client-go v0.20.2
	k8s.io/klog v1.0.0
	k8s.io/kubectl v0.20.2
	sigs.k8s.io/cluster-api v0.3.13
	sigs.k8s.io/controller-runtime v0.8.1
	sigs.k8s.io/yaml v1.2.0
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v14.2.0+incompatible
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	sigs.k8s.io/cluster-api => github.com/getupio-undistro/cluster-api v0.3.11-0.20210124214650-7dec89cfa565
)
